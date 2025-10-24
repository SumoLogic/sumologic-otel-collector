# Installation of Sumo Logic Distribution for OpenTelemetry Collector with Puppet(Linux/Windows)

This [Puppet][puppet] [manifest](manifest/../manifests/install_otel_collector.pp) along with
[module](modules/install_otel_collector/) will install Sumo Logic Distro of [OpenTelemetry Collector][otc_link].

## Windows Support

- To install the Sumo Logic OpenTelemetry Collector on Windows:
  Ensure PowerShell execution policy allows scripts:

   ```powershell
   Set-ExecutionPolicy RemoteSigned -Scope Process -Force
   ```

## Using the module

- Get an [installation token][installation_token] from Sumo Logic
- Prepare [configuration](../../docs/configuration.md) file for Sumo Logic Distribution for OpenTelemetry Collector and put the file in a directory of your choice. You can put multiple configuration files in this directory, and all of them will be used.

  **NOTE**: The playbook will prepare a [base configuration][base_configuration] for you, and configure the [extension][sumologicextension] as well.
- Modify properties in [manifests/install_otel_collector.pp](manifests/install_otel_collector.pp):

  ```ruby
  class install_otel_collector {
     $installation_token => "<your_token>"
     $collector_tags => {"key" => "value"}
     src_config_path => <your_config_path>
  }
  ```

- Apply the changes to your environment. In local mode, run:

  ```bash
  puppet apply
  ```

### Properties

- `installation_token`: Sumo Logic installation token, rel: [installation_token]
- `collector_tags`: Collector tags, these are applied to all processed data
- `api_url`: Sumo Logic API url. You shouldn't need to set this in most normal circumstances.
- `version`: version of Sumo Logic Distribution for OpenTelemetry Collector
- `systemd_service`: enables creation of Systemd Service for Sumo Logic Distribution for OpenTelemetry Collector. Enabled by default. Note that this recipe will not start the collector if you disable this.
- `src_config_path`: path to configuration directory for Sumo Logic Distribution for OpenTelemetry Collector
- `remotely_managed`: enables creation of remotely managed Sumo Logic Distribution for OpenTelemetry Collector. Disabled by default.
- `opamp_api_url`: Sumo Logic Opamp API url. You shouldn't need to set this in most normal circumstances.

## Test on Vagrant

Puppet server and Puppet agent are installed on single host in Vagrant environment.

Example Puppet manifest and module are mounted to Vagrant virtual machine:

- [modules/](modules/)  is mounted to `/etc/puppetlabs/code/environments/production/modules/`
- [manifests/](manifests/) is mounted to `/etc/puppetlabs/code/environments/production/manifests/`

To install Sumo Logic Distribution for OpenTelemetry Collector with Puppet on Vagrant virtual machine:

- Prepare configuration as outlined in [Using the module](#using-the-module)
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
  sudo journalctl -u otelcol-sumo
  ```

- In Windows for verifying the service:

  ```bash
  Get-Service -Name OtelcolSumo
  ```

[puppet]: https://puppet.com/
[otc_link]: https://github.com/open-telemetry/opentelemetry-collector
[installation_token]: https://help.sumologic.com/docs/manage/security/installation-tokens/
[base_configuration]: ../sumologic.yaml
[sumologicextension]: https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/v0.127.0/extension/sumologicextension
