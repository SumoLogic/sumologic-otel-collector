# Ansible playbook to install Sumo Logic Distribution for OpenTelemetry Collector

This playbook will install Sumo Logic Distro of [OpenTelemetry Collector][otc_link].

## Running playbook

- Get an [installation token][installation_token] from Sumo Logic, see
- Prepare [configuration](../../docs/configuration.md) file for Sumo Logic Distribution for OpenTelemetry Collector and put the file in a directory of your choice. You can put multiple configuration files in this directory, and all of them will be used.

  **NOTE**: The playbook will prepare a [base configuration][base_configuration] for you, and configure the [extension][sumologicextension] as well.
- Customize [inventory](inventory) file and add your host
- Run the playbook, passing the prepared values via the command line:

    ```bash
    ansible-playbook -i inventory install_sumologic_otel_collector.yaml \
      -e '{"installation_token": "<your_token>", "collector_tags": {"tag_name": "tag_value"}, "src_config_path": "<your_config_path>"}'
    ```

  *Notice*: If you need to specify a password for sudo, run `ansible-playbook` with `--ask-become-pass` (`-K` for short).

## Playbook variables

- `installation_token`: Sumo Logic [installation token][installation_token]
- `collector_tags`: Collector tags, these are applied to all processed data
- `api_url`: Sumo Logic API url. You shouldn't need to set this in most normal circumstances.
- `version`: version of Sumo Logic Distribution for OpenTelemetry Collector. The default is the latest stable version.
- `systemd_service`: enables creation of Systemd Service for Sumo Logic Distribution for OpenTelemetry Collector. Enabled by default. Note that this playbook will not start the collector if you disable this.
- `src_config_path`: path to configuration directory for Sumo Logic Distribution for OpenTelemetry Collector

[otc_link]: https://github.com/open-telemetry/opentelemetry-collector
[installation_token]: https://help.sumologic.com/docs/manage/security/installation-tokens/
[base_configuration]: ../sumologic.yaml
[sumologicextension]: ../../pkg/extension/sumologicextension/
