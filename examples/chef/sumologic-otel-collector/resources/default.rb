# enable unified mode
unified_mode true

# version of Sumo Logic Distribution for OpenTelemetry Collector
# rel: https://github.com/SumoLogic/sumologic-otel-collector/releases
property :version, String, default: '0.50.0-sumo-0'
# path to configuration file for Sumo Logic Distribution for OpenTelemetry Collector
property :src_config_path, String, default: '/sumologic/examples/config_logging.yaml'
# defines the throttling limit on memory usage for Sumo Logic Distribution for OpenTelemetry Collector
property :memory_high, String, default: '200M'
# defines the absolute limit on memory usage for Sumo Logic Distribution for OpenTelemetry Collector
property :memory_max, String, default: '300M'
# enables creation of Systemd Service for Sumo Logic Distribution for OpenTelemetry Collector
property :systemd_service, [true, false] , default: true
# os architecture
property :os_arch, ['amd64', 'arm_64'], default: 'amd64'
# os family
property :os_family, ['linux', 'darwin'], default: 'linux'

USER = 'otelcol-sumo'
GROUP = 'otelcol-sumo'
BINARY_PATH = '/usr/local/bin/otelcol-sumo'
BINARY_CONFIG = '/etc/otelcol-sumo/config.yaml'

action :default do
  run_action :get_url
  run_action :prepare_config

  if new_resource.systemd_service
    run_action :systemd_service
  else
    run_action :run_in_background
  end
end

action :get_url do
  run_action :create_group
  run_action :create_user

  remote_file BINARY_PATH do
    source "https://github.com/SumoLogic/sumologic-otel-collector/releases/download/v#{new_resource.version}/otelcol-sumo-#{new_resource.version}-#{new_resource.os_family}_#{new_resource.os_arch}"
    owner USER
    group GROUP
    mode '755'
    action :create
  end
end

action :create_group do
  group GROUP do
  end
end

action :create_user do
  user USER do
    gid GROUP
    system true
  end
end

action :systemd_service do
  if node['platform_family'] == 'windows'
    return
  end

  systemd_unit 'otelcol-sumo.service' do
    content({ Unit: {
      Description: 'Sumologic OpenTelemetry Collector',
    },
             Service: {
      ExecStart: "#{BINARY_PATH} --config #{BINARY_CONFIG}",
      User: USER,
      Group: GROUP,
      MemoryHigh: new_resource.memory_high,
      MemoryMax: new_resource.memory_max,
    },
             Install: {
      WantedBy: 'multi-user.target',
    } })

    action [:create, :enable, :start]
  end
end

action :run_in_background do
  execute 'otelcol-sumo' do
    command "sudo -u #{USER} #{BINARY_PATH} --config #{BINARY_CONFIG} > /var/log/otelcol.log 2>&1 &"
  end
end

action :prepare_config do
  run_action :create_group
  run_action :create_user

  directory '/etc/otelcol-sumo' do
    owner USER
    group GROUP
    mode '0750'
    action :create
  end

  file BINARY_CONFIG do
    owner USER
    group GROUP
    content ::File.open(new_resource.src_config_path).read
    action :create
    mode '0640'
  end
end
