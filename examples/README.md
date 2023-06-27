# Examples

- [ansible](ansible) - example Ansible playbook to install Sumo Logic Distribution for OpenTelemetry Collector
- [puppet](puppet) - example manifest and module to install Sumo Logic Distribution for OpenTelemetry Collector with Puppet
- [chef](chef) - example cookbook to install Sumo Logic Distribution for OpenTelemetry Collector with Chef
- [otelcolconfigs](otelcolconfigs) - a directory that contains sample opentelemetry collector configurations, such as:
  - [config_cpu_load_metrics.yaml](otelcolconfigs/config_cpu_load_metrics.yaml) - example configuration to collect CPU load metrics using [Host Metrics Receiver][hostmetricsreceiver]
  - [config_logging.yaml](otelcolconfigs/config_logging.yaml) - example configuration to collect CPU load metrics using
  [Host Metrics Receiver][hostmetricsreceiver] without sending them to Sumo Logic.
  - [logs_json](otelcolconfigs/logs_json) - example configuration to parse logs using [json_parser][json_parser] with example log

[hostmetricsreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.52.0/receiver/hostmetricsreceiver
[json_parser]: https://github.com/open-telemetry/opentelemetry-log-collection/blob/main/docs/operators/json_parser.md
