# Sumo Logic installation token
# rel: https://help.sumologic.com/docs/manage/security/installation-tokens/
# Can be provided via: Chef Vault, Encrypted Data Bag, or directly as attribute
default['sumologic_otel_collector']['installation_token'] = nil

# Chef Vault configuration (most secure option, requires Chef Server)
default['sumologic_otel_collector']['use_vault'] = false
default['sumologic_otel_collector']['vault']['name'] = 'sumologic'
default['sumologic_otel_collector']['vault']['item'] = 'tokens'

# Encrypted Data Bag configuration (secure option, requires Chef Server)
default['sumologic_otel_collector']['use_data_bag'] = false
default['sumologic_otel_collector']['credentials']['bag_name'] = 'sumologic'
default['sumologic_otel_collector']['credentials']['item_name'] = 'tokens'
default['sumologic_otel_collector']['credentials']['secret_file'] = nil  # Optional path to secret key file

# Collector tags, these are applied to all processed data
default['sumologic_otel_collector']['collector_tags'] = {}

# Sumo Logic API url (optional)
# You shouldn't need to set this in most normal circumstances
default['sumologic_otel_collector']['api_url'] = nil

# Version of Sumo Logic Distribution for OpenTelemetry Collector (optional)
# rel: https://github.com/SumoLogic/sumologic-otel-collector/releases
# If not specified, installs the latest version
default['sumologic_otel_collector']['version'] = nil

# Path to a directory with config files for Sumo Logic Distribution for OpenTelemetry Collector (optional)
default['sumologic_otel_collector']['src_config_path'] = nil

# Enables creation of Systemd Service for Sumo Logic Distribution for OpenTelemetry Collector
# ** NOTE** you need to start the collector yourself if you disable this
default['sumologic_otel_collector']['systemd_service'] = true

# Enables remote management for Sumo Logic Distribution for OpenTelemetry Collector
default['sumologic_otel_collector']['remotely_managed'] = false

# Sumo Logic Opamp Api url (optional)
# You shouldn't need to set this in most normal circumstances
default['sumologic_otel_collector']['opamp_api_url'] = nil
