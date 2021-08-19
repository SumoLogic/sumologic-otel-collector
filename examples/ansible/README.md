# Ansible playbook to install Sumo Logic Distro of OpenTelemetry Collector

This playbook will install Sumo Logic Distro of [OpenTelemetry Collector][otc_link].

## Settings

- `sumologic_otel_collector.version`: version of Sumo Logic Distro of OpenTelemetry Collector
- `src_config_path`: path to configuration file for Sumo Logic Distro of OpenTelemetry Collector
- `memory_high`: defines the throttling limit on memory usage for Sumo Logic Distro of OpenTelemetry Collector
- `memory_max`: defines the absolute limit on memory usage for Sumo Logic Distro of OpenTelemetry Collector
- `systemd_service`: enables creation of Systemd Service for Sumo Logic Distro of OpenTelemetry Collector

## Running playbook

- Prepare [configuration](../../docs/Configuration.md) for Sumo Logic Distro of of OpenTelemetry Collector
- Customize [inventory](inventory) file and add your host
- Adjust configuration of Sumo Logic Distro of of OpenTelemetry Collector in [vars/default.yaml](vars/default.yaml)
- Run the playbook

    ```bash
    ansible-playbook -i inventory install_sumologic_otel_collector.yaml
    ```

  *Notice*: If you need to specify a password for sudo, run `ansible-playbook` with `--ask-become-pass` (`-K` for short).

[otc_link]: https://github.com/open-telemetry/opentelemetry-collector
