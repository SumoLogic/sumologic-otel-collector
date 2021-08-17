# Ansible playbook to install Sumo Logic Distro of OpenTelemetry Collector

This playbook will install Sumo Logic Distro of [OpenTelemetry Collector][otc_link].

## Settings

- `sumologic_otel_collector.version`: version of Sumo Logic Distro of of OpenTelemetry Collector

## Running playbook

- Customize [inventory](inventory) file and add your host
- Set desired version of Sumo Logic Distro of of OpenTelemetry Collector in [vars/default.yaml](vars/default.yaml)
- Run the playbook

    ```bash
    ansible-playbook -i inventory install_sumologic_otel_collector.yaml
    ```

  *Notice*: If you need to specify a password for sudo, run `ansible-playbook` with `--ask-become-pass` (`-K` for short).

[otc_link]: https://github.com/open-telemetry/opentelemetry-collector