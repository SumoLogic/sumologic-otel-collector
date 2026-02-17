# enable unified mode
unified_mode true

# Sumo Logic installation token
# rel: https://www.sumologic.com/help/docs/manage/security/installation-tokens/
property :installation_token, String, required: true
# Collector tags, these are applied to all processed data
property :collector_tags, Hash, default: {}
# Sumo Logic API url
property :api_url, String
# enables creation of Systemd Service for Sumo Logic Distribution for OpenTelemetry Collector
# ** NOTE** you need to start the collector yourself if you disable this
property :systemd_service, [true, false] , default: true
# version of Sumo Logic Distribution for OpenTelemetry Collector
# rel: https://github.com/SumoLogic/sumologic-otel-collector/releases
property :version, String
# path to a directory with config files for Sumo Logic Distribution for OpenTelemetry Collector
property :src_config_path, String
# enables remote management for Sumo Logic Distribution for OpenTelemetry Collector
property :remotely_managed, [true, false] , default: false
# Sumo Logic Opamp Api url
property :opamp_api_url, String

# Linux specific constants
DOWNLOAD_TIMEOUT = 300
BINARY_PATH = '/usr/local/bin/otelcol-sumo'
BINARY_CONFIG = '/etc/otelcol-sumo/conf.d'
INSTALL_SCRIPT_PATH = "/tmp/install.sh"
INSTALL_SCRIPT_URL = "https://download-otel.sumologic.com/latest/download/install.sh"

# Windows specific constants
WINDOWS_CONFIG_DIR = 'C:\\Program Files\\Sumo Logic\\OpenTelemetry Collector\\conf.d'
WINDOWS_INSTALL_SCRIPT_URL = 'https://download-otel.sumologic.com/latest/download/install.ps1'
WINDOWS_SERVICE_NAME = 'OtelcolSumo'

action :default do
  if platform_family?('windows')
    run_action :install_windows_collector
    run_action :prepare_windows_config
    run_action :restart_windows_service
  else
    run_action :get_install_script
    install_command = get_install_script_command(new_resource)
    execute INSTALL_SCRIPT_PATH do
      command install_command
      environment ({"SUMOLOGIC_INSTALLATION_TOKEN" => new_resource.installation_token})

    end
    run_action :prepare_config
    if new_resource.systemd_service
      execute 'restart Otelcol' do
        command 'systemctl restart otelcol-sumo'
      end
    end
  end
end

action :get_install_script do
  remote_file INSTALL_SCRIPT_PATH do
    source INSTALL_SCRIPT_URL
    mode '755'
    action :create
  end
end

action :prepare_config do
  if property_is_set?(:src_config_path)
    remote_directory BINARY_CONFIG do
      source new_resource.src_config_path
      owner 'root'
      group 'root'
      mode '0755'
      action :create
    end
  end
end


def get_install_script_command(resource)
  command_parts = ["bash", INSTALL_SCRIPT_PATH, "--download-timeout 300"]
  command_parts += resource.collector_tags.map { |key, value| "--tag #{key}=#{value}" }
  if property_is_set?(:version)
     command_parts.push("--version #{resource.version}")
  end
  if property_is_set?(:api_url)
     command_parts.push("--api #{resource.api_url}")
  end
  if property_is_set?(:opamp_api_url)
     command_parts.push("--opamp-api #{resource.opamp_api_url}")
  end
  if resource.remotely_managed
     command_parts.push("--remotely-managed")
  end
  if ! resource.systemd_service
     command_parts.push("--skip-systemd")
  end
  command_parts.join(" ")
end

action :install_windows_collector do
  powershell_script 'Install Sumo Logic Otel Collector (Windows)' do
    code build_windows_install_command(new_resource)
    action :run
  end
end

action :prepare_windows_config do
  if property_is_set?(:src_config_path)
    directory WINDOWS_CONFIG_DIR do
      recursive true
      rights :full_control, 'Administrators'
      action :create
    end

    remote_directory WINDOWS_CONFIG_DIR do
      source new_resource.src_config_path
      rights :full_control, 'Administrators'
      action :create
    end
  end
end

action :restart_windows_service do
  powershell_script 'Restart Sumo Logic Otel Collector Service (Windows)' do
    code "Restart-Service -Name #{WINDOWS_SERVICE_NAME}"
    action :run
    only_if "Get-Service -Name #{WINDOWS_SERVICE_NAME} -ErrorAction SilentlyContinue"
  end
end

# Helper to build the PowerShell command to install the collector on Windows
def build_windows_install_command(resource)
  tag_lines = resource.collector_tags.map { |k, v| "  '#{k}' = '#{v}'" }
  tag_string = "@{\n#{tag_lines.join("\n")}\n}"

  install_token = resource.installation_token.gsub('"', '`"')
  api_url = resource.api_url.to_s.gsub('"', '`"')
  opamp_api_url = resource.opamp_api_url.to_s.gsub('"', '`"')
  version = resource.version.to_s.gsub('"', '`"')

  command = <<~POWERSHELL
    Set-ExecutionPolicy RemoteSigned -Scope Process -Force;
    [System.Net.ServicePointManager]::SecurityProtocol = [System.Net.ServicePointManager]::SecurityProtocol -bor 3072;
    $uri = "#{WINDOWS_INSTALL_SCRIPT_URL}";
    $path = "$env:TEMP\\install.ps1";
    (New-Object System.Net.WebClient).DownloadFile($uri, $path);
    $tags = #{tag_string};
    & $path -InstallationToken "#{install_token}" -Tags $tags
  POWERSHELL

  command.chomp!
  command << " -Api \"#{api_url}\"" if property_is_set?(:api_url)
  command << " -OpAmpApi \"#{opamp_api_url}\"" if property_is_set?(:opamp_api_url)
  command << " -Version \"#{version}\"" if property_is_set?(:version)
  command << " -RemotelyManaged \$true" if resource.remotely_managed

  command.strip
end
