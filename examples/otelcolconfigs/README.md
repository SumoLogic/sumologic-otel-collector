# Examples

- [config_cpu_load_metrics.yaml](config_cpu_load_metrics.yaml) - example configuration to collect CPU load metrics using [Host Metrics Receiver][hostmetricsreceiver]
- [config_logging.yaml](config_logging.yaml) - example configuration to collect CPU load metrics using
  [Host Metrics Receiver][hostmetricsreceiver] without sending them to Sumo Logic.
- [logs_json](logs_json) - example configuration to parse logs using [json_parser][json_parser] with example log

[hostmetricsreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.52.0/receiver/hostmetricsreceiver
[json_parser]: https://github.com/open-telemetry/opentelemetry-log-collection/blob/main/docs/operators/json_parser.md
