# Installation token can be provided via multiple methods (in order of precedence):
# 1. Chef Vault (most secure, requires Chef Server)
# 2. Encrypted Data Bag (secure, requires Chef Server)
# 3. Node attributes (flexible, works with chef-solo)

installation_token = nil

# Method 1: Try Chef Vault first
if node['sumologic_otel_collector']['use_vault']
  begin
    require 'chef-vault'
    vault_name = node['sumologic_otel_collector']['vault']['name']
    vault_item = node['sumologic_otel_collector']['vault']['item']

    sumo_creds = ChefVault::Item.load(vault_name, vault_item)
    installation_token = sumo_creds['installation_token']
    Chef::Log.info('Loaded Sumo Logic credentials from Chef Vault')
  rescue LoadError
    Chef::Log.warn('chef-vault gem not available, falling back to data bags or attributes')
  rescue ChefVault::Exceptions::KeysNotFound
    Chef::Log.warn('Chef Vault item not found, falling back to data bags or attributes')
  end
end

# Method 2: Try encrypted data bag if vault didn't work
if installation_token.nil? && node['sumologic_otel_collector']['use_data_bag']
  begin
    bag_name = node['sumologic_otel_collector']['credentials']['bag_name']
    item_name = node['sumologic_otel_collector']['credentials']['item_name']
    secret_file = node['sumologic_otel_collector']['credentials']['secret_file']

    if secret_file && ::File.exist?(secret_file)
      secret = Chef::EncryptedDataBagItem.load_secret(secret_file)
      sumo_creds = Chef::EncryptedDataBagItem.load(bag_name, item_name, secret)
    else
      sumo_creds = Chef::EncryptedDataBagItem.load(bag_name, item_name)
    end

    installation_token = sumo_creds['installation_token']
    Chef::Log.info('Loaded Sumo Logic credentials from encrypted data bag')
  rescue Net::HTTPClientException, Chef::Exceptions::InvalidDataBagPath
    Chef::Log.warn('Data bag not found, falling back to attributes')
  end
end

# Method 3: Fall back to node attributes (works with chef-solo)
if installation_token.nil?
  installation_token = node['sumologic_otel_collector']['installation_token']
  Chef::Log.info('Using Sumo Logic credentials from node attributes')
end

# Fail if no token found from any method
if installation_token.nil? || installation_token.empty?
  raise <<-ERROR
    Sumo Logic installation token not provided!

    Please provide the token using one of these methods:

    1. Chef Vault (recommended for Chef Server):
       knife vault create sumologic tokens '{"installation_token":"YOUR_TOKEN"}' --search "role:yourRole"
       Set node['sumologic_otel_collector']['use_vault'] = true

    2. Encrypted Data Bag (for Chef Server):
       knife data bag create sumologic tokens --secret-file ~/.chef/secret
       Set node['sumologic_otel_collector']['use_data_bag'] = true

    3. Node Attributes (works with chef-solo):
       Set node['sumologic_otel_collector']['installation_token'] = 'YOUR_TOKEN'
       Or use JSON config: {"sumologic_otel_collector":{"installation_token":"YOUR_TOKEN"}}

    See README.md for detailed instructions.
  ERROR
end

# Install and configure collector
sumologic_otel_collector 'sumologic-otel-collector' do
  installation_token installation_token
  collector_tags node['sumologic_otel_collector']['collector_tags']
  api_url node['sumologic_otel_collector']['api_url'] if node['sumologic_otel_collector']['api_url']
  version node['sumologic_otel_collector']['version'] if node['sumologic_otel_collector']['version']
  systemd_service node['sumologic_otel_collector']['systemd_service']
  src_config_path node['sumologic_otel_collector']['src_config_path'] if node['sumologic_otel_collector']['src_config_path']
  remotely_managed node['sumologic_otel_collector']['remotely_managed']
  opamp_api_url node['sumologic_otel_collector']['opamp_api_url'] if node['sumologic_otel_collector']['opamp_api_url']
end

# Ensure service is running
service_name = platform_family?('windows') ? 'OtelcolSumo' : 'otelcol-sumo'

service service_name do
  action [:enable, :start]
  only_if { node['sumologic_otel_collector']['systemd_service'] }
end
