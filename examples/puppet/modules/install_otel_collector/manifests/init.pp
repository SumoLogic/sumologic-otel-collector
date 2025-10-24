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
# @param opamp_api_url
#   Optional OpAmp API URL
#

class install_otel_collector (
  String $installation_token,
  Hash[String, String] $collector_tags = {},
  Optional[String] $api_url = undef,
  Boolean $systemd_service = true,
  Optional[String] $version = undef,
  String $src_config_path = 'puppet:///modules/install_otel_collector/conf.d',
  Optional[String] $opamp_api_url = undef,
  Boolean $remotely_managed = false,
) {
  if $facts['os']['family'] == 'windows' {

    $install_script_url  = 'https://download-otel.sumologic.com/latest/download/install.ps1'
    $install_script_path = 'C:/Windows/Temp/install.ps1'
    $exe_path            = 'C:/Program Files/Sumo Logic/OpenTelemetry Collector/bin/otelcol-sumo.exe'

    # Escape single quotes for PowerShell
    function escape_single_quotes(String $str) {
      $str.gsub("'", "''")
    }

    # Build the escaped hashtable
    $tags_ps_lines = $collector_tags.map |$k, $v| {
      "    '${escape_single_quotes($k)}' = '${escape_single_quotes($v)}';"
    }
    $tags_ps_block = "@{\n${tags_ps_lines.join("\n")}\n}"

    # Optional arguments as strings (empty if undef)
    $api_arg    = $api_url ? { undef => '', default => "-Api '${api_url}'" }
    $opamp_arg  = $opamp_api_url ? { undef => '', default => "-OpAmpApi '${opamp_api_url}'" }
    $remote_arg =  "-RemotelyManaged \$${remotely_managed}"
    # Download install.ps1 script
    exec { 'Download Otel Collector Script':
      command   => "C:/Windows/System32/WindowsPowerShell/v1.0/powershell.exe -ExecutionPolicy Bypass -Command \"Invoke-WebRequest -Uri '${install_script_url}' -OutFile '${install_script_path}'\"",
      creates   => $install_script_path,
      logoutput => true,
    }

    # Run install script with tags defined inside PowerShell command
    exec { 'Install Otel Collector':
      command => "C:/Windows/System32/WindowsPowerShell/v1.0/powershell.exe -ExecutionPolicy RemoteSigned -Command \"\
        Set-ExecutionPolicy RemoteSigned -Scope Process -Force; \
        [System.Net.ServicePointManager]::SecurityProtocol = [System.Net.ServicePointManager]::SecurityProtocol -bor 3072; \
        \$tags = ${tags_ps_block}; \
        & '${install_script_path}' -InstallationToken '${installation_token}' -Tags \$tags ${remote_arg} ${api_arg} ${opamp_arg}\"",
      unless  => "C:/Windows/System32/WindowsPowerShell/v1.0/powershell.exe -Command \"if (Test-Path '${exe_path}') { exit 0 } else { exit 1 }\"",
      require => Exec['Download Otel Collector Script'],
      logoutput => true,
    }

    # Ensure directories exist
    file { 'C:/Program Files/Sumo Logic/OpenTelemetry Collector/conf.d':
      ensure  => directory,
      require => File['C:/Program Files/Sumo Logic/OpenTelemetry Collector'],
    }
  } else {

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
    if $remotely_managed {
      $remotely_managed_command_args = ['--remotely-managed']
    } else {
      $remotely_managed_command_args = []
    }
    if $opamp_api_url == undef {
      $opamp_command_args = []
    } else {
      $opamp_command_args = ["--opamp-api ${opamp_api_url}"]
    }
    $install_command_args = ["--download-timeout ${download_timeout}"] + $tags_command_args + $version_command_args + $api_command_args + $systemd_command_args + $remotely_managed_command_args + $opamp_command_args

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
}
