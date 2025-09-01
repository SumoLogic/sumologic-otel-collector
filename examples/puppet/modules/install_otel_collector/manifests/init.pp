# @summary Installs Sumologic Otel Collector
#
# @example Basic usage
#   class { 'install_otel_collector':
#     installation_token => '...'
#     collector_tags => { 'key' => 'value' }
#   }
#
# @param installation_token
#   Sumo Logic installation token, rel: https://help.sumologic.com/docs/manage/security/installation-tokens/
# @param collector_tags
#   Collector tags, these are applied to all processed data
# @param api_url
#   Sumo Logic API url
# @param systemd_service
#   Enables creation of Systemd Service. Note that Opentelemetry Collector will not be started if this is disabled.
# @param version
#   Sumologic Otel Collector version. Defaults to latest stable.
# @param src_config_path
#   Path to a directory with config files.
#
class install_otel_collector (
  String $installation_token,
  Hash[String, String] $collector_tags = {},
  Optional[String] $api_url = undef,
  Boolean $systemd_service = true,
  Optional[String] $version = undef,
  String $src_config_path = 'puppet:///modules/install_otel_collector/conf.d',
) {
  $install_script_url = 'https://download-otel.sumologic.com/latest/download/install.sh'
  $install_script_path = '/tmp/install.sh'
  $download_timeout = 300

  # construct the install command arguments from class parameters
  $tags_command_args = $collector_tags.map |$key, $value| { "--tag ${key}=${value}" }
  if $version == undef {
    $version_command_args = []
  } else {
    $version_command_args = ["--version ${version}"]
  }
  if $api_url == undef {
    $api_command_args = []
  } else {
    $api_command_args = ["--api ${api_url}"]
  }
  if $systemd_service {
    $systemd_command_args = []
  } else {
    $systemd_command_args = ['--skip-systemd']
  }
  $install_command_args = ["--download-timeout ${download_timeout}"] + $tags_command_args + $version_command_args + $api_command_args + $systemd_command_args

  file { 'download the install script':
    source => $install_script_url,
    path   => $install_script_path,
    mode   => '0755',
  }

  $install_command_parts = ['bash', $install_script_path] + $install_command_args
  $install_command = join($install_command_parts, ' ')
  exec { 'run the installation script':
    command     => $install_command,
    path        => ['/usr/local/bin/', '/usr/bin', '/usr/sbin', '/bin'],
    user        => 'root',
    environment => ["SUMOLOGIC_INSTALLATION_TOKEN=${installation_token}"],
    require     => File[$install_script_path],
  }

  file { '/etc/otelcol-sumo/conf.d':
    ensure  => directory,
    recurse => true,
    source  => $src_config_path,
    require => Exec['run the installation script'],
  }
}
