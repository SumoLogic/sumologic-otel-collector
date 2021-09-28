# Chef cookbook to install Sumo Logic Distro of OpenTelemetry Collector

This cookbook will install Sumo Logic Distro of [OpenTelemetry Collector][otc_link].

## Properties

- `version`: version of Sumo Logic Distro of OpenTelemetry Collector
- `src_config_path`: path to configuration file for Sumo Logic Distro of OpenTelemetry Collector
- `memory_high`: defines the throttling limit on memory usage for Sumo Logic Distro of OpenTelemetry Collector
- `memory_max`: defines the absolute limit on memory usage for Sumo Logic Distro of OpenTelemetry Collector
- `systemd_service`: enables creation of Systemd Service for Sumo Logic Distro of OpenTelemetry Collector
- `os_family`: OS family, can by either `linux` or `darwin`
- `os_arch`: OS architecture, can be either `amd64` or `arm64`. `arm64` is supported for `linux` `os_family` only

## Test on Vagrant

Chef-solo is installed in Vagrant environment to simplify testing and modifying chef cookbook.

[Chef playground](.) is mounted as `/sumologic/examples/chef`.
The following steps describe procedure of testing changes:

- Prepare configuration for Sumo Logic Distro of OpenTelemetry Collector
  using steps described in [Configuration](../../docs/Configuration.md)
- Adjust [recipe](sumologic-otel-collector/recipes/default.rb) to your needs
- From main directory of this repository start virtual machine:

  ```bash
  vagrant up
  ```

- Connect to virtual machine:

  ```bash
  vagrant ssh
  ```

- Run the cookbook with the default recipe:

  ```bash
  sudo chef-solo -c /sumologic/examples/chef/config.rb -o sumologic-otel-collector
  ```

- Verify installation:
  
  ```bash
  sudo ps aux | grep otelcol-sumo
  ```

- Verify logs:
  
  ```bash
  cat /var/log/otelcol.log 
  ```

[otc_link]: https://github.com/open-telemetry/opentelemetry-collector
