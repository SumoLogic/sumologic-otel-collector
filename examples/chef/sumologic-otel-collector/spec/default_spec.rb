require 'chefspec'
require 'chefspec/berkshelf'

describe 'sumologic-otel-collector::default' do
  let(:chef_run) { ChefSpec::SoloRunner.new(platform: platform, version: version).converge(described_recipe) }
  let(:installation_token) { 'test-token' }
  let(:collector_tags) { { 'env' => 'test', 'role' => 'collector' } }
  let(:api_url) { 'https://api.sumologic.com' }
  let(:src_config_path) { '/path/to/config' }

  before do
    chef_run.node.override['sumologic']['installation_token'] = installation_token
    chef_run.node.override['sumologic']['collector_tags'] = collector_tags
    chef_run.node.override['sumologic']['api_url'] = api_url
    chef_run.node.override['sumologic']['src_config_path'] = src_config_path
  end

  context 'on Linux' do
    let(:platform) { 'ubuntu' }
    let(:version) { '20.04' }

    it 'downloads the install script' do
      expect(chef_run).to create_remote_file('/tmp/install.sh').with(
        source: 'https://github.com/SumoLogic/sumologic-otel-collector-packaging/releases/latest/download/install.sh',
        mode: '755'
      )
    end

    it 'executes the install script with correct command' do
      expect(chef_run).to run_execute('/tmp/install.sh').with(
        command: "bash /tmp/install.sh --download-timeout 300 --tag env=test --tag role=collector --api https://api.sumologic.com",
        environment: { 'SUMOLOGIC_INSTALLATION_TOKEN' => installation_token }
      )
    end

    it 'creates the configuration directory if src_config_path is set' do
      expect(chef_run).to create_remote_directory('/etc/otelcol-sumo/conf.d').with(
        source: src_config_path,
        owner: 'root',
        group: 'root',
        mode: '0755'
      )
    end

    it 'restarts the otelcol service if systemd_service is enabled' do
      chef_run.node.override['sumologic']['systemd_service'] = true
      chef_run.converge(described_recipe)
      expect(chef_run).to run_execute('restart Otelcol').with(command: 'systemctl restart otelcol-sumo')
    end
  end
end