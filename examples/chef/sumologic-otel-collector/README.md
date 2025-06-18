# Chef cookbook to install Sumo Logic Distribution for OpenTelemetry Collector

This cookbook will install Sumo Logic Distro of [OpenTelemetry Collector][otc_link].

## Using the cookbook

- Get an [installation token][installation_token] from Sumo Logic
- Prepare [configuration](../../../docs/configuration.md) file for Sumo Logic Distribution for OpenTelemetry Collector and put the file in a directory of your choice. You can put multiple configuration files in this directory, and all of them will be used.

  **NOTE**: The playbook will prepare a [base configuration][base_configuration] for you, and configure the [extension][sumologicextension] as well.
- Prepare Chef Recipe and save it in the [recipes/default.rb](recipes/default.rb) file

    ```ruby
    sumologic_otel_collector 'sumologic-otel-collector' do
      installation_token '<your_token>'
      tags ({'abc' => 'def'})
      src_config_path '<your_config_path>'
    end
    ```

- Apply the changes to your environment. If using `chef-solo`, run the following:

    ```bash
    sudo chef-solo --config-option cookbook_path=$(pwd) -o sumologic-otel-collector
    ```

## Properties

- `installation_token`: Sumo Logic [installation token][installation_token]
- `collector_tags`: Collector tags, these are applied to all processed data
- `api_url`: Sumo Logic API url. You shouldn't need to set this in most normal circumstances.
- `version`: version of Sumo Logic Distribution for OpenTelemetry Collector. The default is the latest stable version.
- `systemd_service`: enables creation of Systemd Service for Sumo Logic Distribution for OpenTelemetry Collector. Enabled by default. Note that this playbook will not start the collector if you disable this.
- `src_config_path`: path to configuration directory for Sumo Logic Distribution for OpenTelemetry Collector

## Test on Vagrant

Chef-solo is installed in Vagrant environment to simplify testing and modifying chef cookbook.

[Chef playground](.) is mounted as `/sumologic/examples/chef`.
The following steps describe procedure of testing changes:

- Prepare configuration as outlined in [Using the cookbook](#using-the-cookbook)
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
  sudo journalctl -u otelcol-sumo
  ```

[otc_link]: https://github.com/open-telemetry/opentelemetry-collector
[installation_token]: https://help.sumologic.com/docs/manage/security/installation-tokens/
[base_configuration]: ../../sumologic.yaml
[sumologicextension]: ../../../pkg/extension/sumologicextension/
