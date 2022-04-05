sumologic_otel_collector 'sumologic-otel-collector' do
  version '0.47.0-sumo-0'
  src_config_path '/sumologic/examples/config_logging.yaml'
  memory_high '200M'
  memory_max '300M'
  systemd_service true
  os_family 'linux'
  os_arch 'amd64'

  action :default
end
