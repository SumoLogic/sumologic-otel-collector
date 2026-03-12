# Chef cookbook to install Sumo Logic Distribution for OpenTelemetry Collector(Linux/Windows)

This cookbook will install Sumo Logic Distro of [OpenTelemetry Collector][otc_link].

## Windows Support

- To install the Sumo Logic OpenTelemetry Collector on Windows:
  Ensure PowerShell execution policy allows scripts:

   ```powershell
   Set-ExecutionPolicy RemoteSigned -Scope Process -Force
   ```

## Using the cookbook

- Get an [installation token][installation_token] from Sumo Logic
- Prepare [configuration](../../../docs/configuration.md) file for Sumo Logic Distribution for OpenTelemetry Collector and put the file in a directory of your choice. You can put multiple configuration files in this directory, and all of them will be used.

  **NOTE**: The cookbook will prepare a [base configuration][base_configuration] for you, and configure the [extension][sumologicextension] as well.

### Option 1: Using Attributes (Recommended)

Set attributes in your wrapper cookbook, role, or environment:

```ruby
# In a wrapper cookbook's attributes/default.rb
default['sumologic_otel_collector']['installation_token'] = 'your_token'
default['sumologic_otel_collector']['collector_tags'] = {
  'environment' => 'production',
  'team' => 'platform'
}
default['sumologic_otel_collector']['src_config_path'] = '/etc/otelcol-sumo/conf.d'

# In your recipe
include_recipe 'sumologic-otel-collector::default'
```

### Option 2: Using JSON File

Create a JSON file with your configuration:

```json
{
  "sumologic_otel_collector": {
    "installation_token": "your_token",
    "collector_tags": {
      "environment": "production"
    }
  },
  "run_list": ["recipe[sumologic-otel-collector]"]
}
```

Apply with chef-solo:

```bash
sudo chef-solo -j node.json --config-option cookbook_path=$(pwd)
```

### Option 3: Direct Resource Usage

Use the resource directly in your own recipe:

```ruby
sumologic_otel_collector 'sumologic-otel-collector' do
  installation_token 'your_token'
  collector_tags ({ 'environment' => 'production' })
  src_config_path '/etc/otelcol-sumo/conf.d'
end
```

## Attributes

All attributes are under the `node['sumologic_otel_collector']` namespace.

| Attribute | Type | Description | Default |
|-----------|------|-------------|---------|
| `installation_token` | String | Sumo Logic [installation token][installation_token] | `nil` (required) |
| `collector_tags` | Hash | Collector tags, these are applied to all processed data | `{}` |
| `api_url` | String | Sumo Logic API url. You shouldn't need to set this in most normal circumstances. | `nil` |
| `version` | String | Version of Sumo Logic Distribution for OpenTelemetry Collector. If not specified, installs the latest stable version. | `nil` |
| `systemd_service` | Boolean | Enables creation of Systemd Service for Sumo Logic Distribution for OpenTelemetry Collector. Note: the collector will not start if you disable this. | `true` |
| `src_config_path` | String | Path to configuration directory for Sumo Logic Distribution for OpenTelemetry Collector | `nil` |
| `remotely_managed` | Boolean | Enables creation of remotely managed Sumo Logic Distribution for OpenTelemetry Collector | `false` |
| `opamp_api_url` | String | Sumo Logic Opamp API url. You shouldn't need to set this in most normal circumstances. | `nil` |

## Resource Properties

The `sumologic_otel_collector` resource accepts the following properties (same as attributes above):

- `installation_token`: Sumo Logic [installation token][installation_token] (required)
- `collector_tags`: Collector tags, these are applied to all processed data
- `api_url`: Sumo Logic API url
- `version`: version of Sumo Logic Distribution for OpenTelemetry Collector
- `systemd_service`: enables creation of Systemd Service
- `src_config_path`: path to configuration directory
- `remotely_managed`: enables remote management
- `opamp_api_url`: Sumo Logic Opamp API url

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

- In Windows for verifying the service:

  ```powershell
  Get-Service -Name OtelcolSumo
  ```

[otc_link]: https://github.com/open-telemetry/opentelemetry-collector
[installation_token]: https://www.sumologic.com/help/docs/manage/security/installation-tokens/
[base_configuration]: ../../sumologic.yaml
[sumologicextension]: https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/v0.127.0/extension/sumologicextension
