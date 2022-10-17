class install_otel_collector {

   $otel_collector_version = "0.50.0-sumo-0" # version of Sumo Logic Distribution for OpenTelemetry Collector
   $systemd_service = false                  # enables creation of Systemd Service for Sumo Logic Distribution for OpenTelemetry Collector

   $arch = $facts['os']['architecture'] ? {
      'aarch64' => 'arm64',
      'arm64'   => 'arm64',
      default   => 'amd64',
   }

   $os_family = $facts['os']['family'] ? {
      'Darwin' => 'darwin',
      default   => 'linux',
   }

   exec {"download the release binary":
      cwd     => "/usr/local/bin/",
      command => "curl -sLo otelcol-sumo https://github.com/SumoLogic/sumologic-otel-collector/releases/download/v${otel_collector_version}/otelcol-sumo-${otel_collector_version}-${os_family}_${arch}",
      path    => ['/usr/local/bin/', '/usr/bin', '/usr/sbin', '/bin'],
   }

   exec {"make otelcol-sumo executable":
      cwd     => "/usr/local/bin/",
      command => "chmod +x otelcol-sumo",
      path    => ['/usr/local/bin/', '/usr/bin', '/usr/sbin', '/bin'],
   }

   file {"/etc/otelcol-sumo":
     ensure => 'directory',
   }

   file {"/etc/otelcol-sumo/config.yaml":
     source => "puppet:///modules/install_otel_collector/config.yaml",
     mode => "640",
   }

   group {"otelcol-sumo":
      ensure  => "present",
   }

   user {"otelcol-sumo":
      ensure  => "present",
      groups  => ["otelcol-sumo"],
      managehome => true,
   }

   if $systemd_service {
      file {"/etc/systemd/system/otelcol-sumo.service":
         source => "puppet:///modules/install_otel_collector/systemd_service",
         mode => "644",
      }

      service {"otelcol-sumo":
         ensure => "running",
         enable => true,
      }
   } else {
      exec { 'run otelcol-sumo in background':
         command => 'sudo -u otelcol-sumo /usr/local/bin/otelcol-sumo --config /etc/otelcol-sumo/config.yaml > /var/log/otelcol.log 2>&1 &',
         path    => ['/usr/local/bin/', '/usr/bin', '/usr/sbin', '/bin'],
         logoutput => true,
         provider => shell,
      }
   }
}
