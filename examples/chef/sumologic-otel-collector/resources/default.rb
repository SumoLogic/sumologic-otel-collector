# enable unified mode
unified_mode true

# Sumo Logic install token
# rel: https://help.sumologic.com/docs/manage/security/installation-tokens/
property :install_token, String, required: true
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

BINARY_PATH = '/usr/local/bin/otelcol-sumo'
BINARY_CONFIG = '/etc/otelcol-sumo/conf.d'
INSTALL_SCRIPT_PATH = "/tmp/install.sh"
INSTALL_SCRIPT_URL = "https://raw.githubusercontent.com/SumoLogic/sumologic-otel-collector/main/scripts/install.sh"

action :default do
  run_action :get_install_script
  install_command = get_install_script_command(new_resource)
  execute INSTALL_SCRIPT_PATH do
    command install_command
  end
  run_action :prepare_config
  if new_resource.systemd_service
    execute 'restart Otelcol' do
      command 'systemctl restart otelcol-sumo'
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
  command_parts = ["bash", INSTALL_SCRIPT_PATH, "--installation-token #{resource.install_token}"]
  command_parts += resource.collector_tags.map { |key, value| "--tag #{key}=#{value}" }
  if property_is_set?(:version)
     command_parts.push("--version #{resource.version}")
  end
  if property_is_set?(:api_url)
     command_parts.push("--api #{resource.api_url}")
  end
  if ! resource.systemd_service
     command_parts.push("--skip-systemd")
  end
  command_parts.join(" ")
end
