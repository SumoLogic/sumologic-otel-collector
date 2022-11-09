node 'default' {
  class { 'install_otel_collector':
    install_token  => 'dummy',
    collector_tags => { 'key' => 'value' },
  }
}
