node 'default' {
  class { 'install_otel_collector':
    installation_token => 'dummy',
    collector_tags     => { 'key' => 'value' },
  }
}
