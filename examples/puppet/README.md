# Installation of Sumo Logic Distro of OpenTelemetry Collector with Puppet

This [Puppet][puppet] [manifest](manifest/../manifests/install_otel_collector.pp) along with
[module](modules/install_otel_collector/) will install Sumo Logic Distro of [OpenTelemetry Collector][otc_link].

## Configuration

- Prepare [configuration](../../docs/Configuration.md) for Sumo Logic Distro of OpenTelemetry Collector and
  save it in [files](modules/install_otel_collector/files/) directory for `instal_otel_collector` module as `config.yaml`.
- If needed modify variables in [modules/install_otel_collector/manifests/init.pp](modules/install_otel_collector/manifests/init.pp):

  ```ruby
  class install_otel_collector {
     $otel_collector_version = "0.47.0-sumo-0" # version of Sumo Logic Distro of OpenTelemetry Collector
     $systemd_service = false                  # enables creation of Systemd Service for Sumo Logic Distro of OpenTelemetry Collector

  ...
  }
  ```

- Adjust settings for Systemd Service in [system_service](modules/install_otel_collector/files/systemd_service) when it needs to be created.

## Test on Vagrant

Puppet server and Puppet agent are installed on single host in Vagrant environment.

Example Puppet manifest and module are mounted to Vagrant virtual machine:

- [modules/](modules/)  is mounted to `/etc/puppetlabs/code/environments/production/modules/`
- [manifests/](manifests/) is mounted to `/etc/puppetlabs/code/environments/production/manifests/`

To install Sumo Logic Distro of OpenTelemetry Collector with Puppet on Vagrant virtual machine:

- Prepare configuration using steps described in [Configuration](#configuration)
- From main directory of this repository start virtual machine:

  ```bash
  vagrant up
  ```

- Connect to virtual machine:

  ```bash
  vagrant ssh
  ```

- Pull configuration for Puppet agent:

  ```bash
  sudo puppet agent --test --waitforcert 60
  ```

- In another terminal window for Vagrant virtual machine, sign the certificate:

  ```bash
  sudo puppetserver ca sign --certname agent
  ```

- See that Puppet agent pulls configuration from Puppet server.
- Verify installation:

  ```bash
  sudo ps aux | grep otelcol-sumo
  ```

- Verify logs:

  ```bash
  cat /var/log/otelcol.log
  ```

[otc_link]: https://github.com/open-telemetry/opentelemetry-collector
[puppet]: https://puppet.com/
