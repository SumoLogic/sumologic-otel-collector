# Examples

- [config_cpu_load_metrics.yaml](./config_cpu_load_metrics.yaml) - example configuration to collect CPU load metrics using [Host Metrics Receiver][hostmetricsreceiver]
- [config_logging.yaml](./config_logging.yaml) - example configuration to collect CPU load metrics using
  [Host Metrics Receiver][hostmetricsreceiver] without sending them to Sumo Logic.
- [logs_json](./logs_json) - example configuration to parse logs using [json_parser][json_parser] with example log
- [tracing agent configuration](./agent_configuration_template.yaml) - example configuration to collect traces.
- [tracing gateway configuration](./gateway_configuration_template.yaml) - example configuration to load balance traces using [loadbalancingexporter][loadbalancingexporter].
- [tracing sampler configuration](./sampler_configuration_template.yaml) - example configuration to filter traces using [cascadingfilterprocessor][cascadingfilterprocessor].
- [multiple tracing sampler configuration](./sampler_configuration_multiple_instances_template.yaml) - example configuration to filter traces using [cascadingfilterprocessor][cascadingfilterprocessor]
  in case of multiple instances.

[hostmetricsreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.52.0/receiver/hostmetricsreceiver
[json_parser]: https://github.com/open-telemetry/opentelemetry-log-collection/blob/main/docs/operators/json_parser.md
[loadbalancingexporter]: https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/exporter/loadbalancingexporter
[cascadingfilterprocessor]: https://github.com/SumoLogic/sumologic-otel-collector/tree/main/pkg/processor/cascadingfilterprocessor
