# sumologic-otel-collector

[![Default branch build](https://github.com/SumoLogic/sumologic-otel-collector/actions/workflows/dev_builds.yml/badge.svg)](https://github.com/SumoLogic/sumologic-otel-collector/actions/workflows/dev_builds.yml)

Sumo Logic Distro of [OpenTelemetry Collector][otc_link] built with
[opentelemetry-collector-builder][otc_builder_link].

[otc_link]: https://github.com/open-telemetry/opentelemetry-collector
[otc_builder_link]: https://github.com/open-telemetry/opentelemetry-collector-builder

**This software is currently in beta and is not recommended for production environments.**
**If you wish to participate in this beta, please contact your Sumo Logic account team or Sumo Logic Support.**

- [Usage](#usage)
- [Built-in Components](#built-in-components)
  - [Receivers](#receivers)
  - [Processors](#processors)
  - [Exporters](#exporters)
  - [Extensions](#extensions)
- [Contributing](#contributing)

## Usage

See the [documentation](docs/README.md).

## Built-in Components

This sections represents the supported components that are included in Sumo Logic
OT distro.

<!-- markdownlint-disable MD013 -->

### Receivers

#### Sumo Logic supported receivers

| Name                                                           | Source                                                                                        |
|----------------------------------------------------------------|-----------------------------------------------------------------------------------------------|
| `telegrafreceiver` [configuration help][telegrafreceiver_help] | https://github.com/SumoLogic/sumologic-otel-collector/tree/main/pkg/receiver/telegrafreceiver |

[telegrafreceiver_help]: https://github.com/SumoLogic/sumologic-otel-collector/blob/main/docs/Configuration.md#telegraf-receiver

#### Upstream receivers

| Name                                                                     | Source                                                                                                        |
|--------------------------------------------------------------------------|---------------------------------------------------------------------------------------------------------------|
| `filelogreceiver` [configuration help][filelogreceiver_help]             | https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.35.0/receiver/filelogreceiver       |
| `fluentforwardreceiver` [configuration help][fluentforwardreceiver_help] | https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.35.0/receiver/fluentforwardreceiver |
| `hostmetricsreceiver` [configuration help][hostmetricsreceiver_help]     | https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.35.0/receiver/hostmetricsreceiver   |
| `syslogreceiver` [configuration help][syslogreceiver_help]               | https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.35.0/receiver/syslogreceiver        |
| `statsdreceiver` [configuration help][statsdreceiver_help]               | https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.35.0/receiver/statsdreceiver        |
| `tcplogreceiver` [configuration help][tcplogreceiver_help]               | https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.35.0/receiver/tcplogreceiver        |
| `udplogreceiver` [configuration help][udplogreceiver_help]               | https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.35.0/receiver/udplogreceiver        |
| `zipkinreceiver` [configuration help][zipkinreceiver_help]               | https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.35.0/receiver/zipkinreceiver        |

[filelogreceiver_help]: https://github.com/SumoLogic/sumologic-otel-collector/blob/main/docs/Configuration.md#filelog-receiver
[fluentforwardreceiver_help]: https://github.com/SumoLogic/sumologic-otel-collector/blob/main/docs/Configuration.md#fluent-forward-receiver
[hostmetricsreceiver_help]: https://github.com/SumoLogic/sumologic-otel-collector/blob/main/docs/Configuration.md#host-metrics-receiver
[statsdreceiver_help]: https://github.com/SumoLogic/sumologic-otel-collector/blob/main/docs/Configuration.md#statsd-receiver
[syslogreceiver_help]: https://github.com/SumoLogic/sumologic-otel-collector/blob/main/docs/Configuration.md#syslog-receiver
[tcplogreceiver_help]: https://github.com/SumoLogic/sumologic-otel-collector/blob/main/docs/Configuration.md#tcplog-receiver
[udplogreceiver_help]: https://github.com/SumoLogic/sumologic-otel-collector/blob/main/docs/Configuration.md#udplog-receiver
[zipkinreceiver_help]: https://github.com/SumoLogic/sumologic-otel-collector/blob/main/docs/Configuration.md#zipkin-receiver

### Processors

#### Sumo Logic supported processors

| Name                                                                           | Source                                                                                                 |
|--------------------------------------------------------------------------------|--------------------------------------------------------------------------------------------------------|
| `cascadingfilterprocessor` [configuration help][cascadingfilterprocessor_help] | https://github.com/SumoLogic/sumologic-otel-collector/tree/main/pkg/processor/cascadingfilterprocessor |
| `k8sprocessor` [configuration help][k8sprocessor_help]                         | https://github.com/SumoLogic/sumologic-otel-collector/tree/main/pkg/processor/k8sprocessor             |
| `sourceprocessor` [configuration help][sourceprocessor_help]                   | https://github.com/SumoLogic/sumologic-otel-collector/tree/main/pkg/processor/sourceprocessor          |
| `sumologicsyslogprocessor` [configuration help][sumologicsyslogprocessor_help] | https://github.com/SumoLogic/sumologic-otel-collector/tree/main/pkg/processor/sumologicsyslogprocessor |

[cascadingfilterprocessor_help]: https://github.com/SumoLogic/sumologic-otel-collector/blob/main/docs/Configuration.md#cascading-filter-processor
[k8sprocessor_help]: https://github.com/SumoLogic/sumologic-otel-collector/blob/main/docs/Configuration.md#kubernetes-processor
[sourceprocessor_help]: https://github.com/SumoLogic/sumologic-otel-collector/blob/main/docs/Configuration.md#source-processor
[sumologicsyslogprocessor_help]: https://github.com/SumoLogic/sumologic-otel-collector/blob/main/docs/Configuration.md#sumo-logic-syslog-processor

#### Upstream processors

| Name                                                                               | Source                                                                                                              |
|------------------------------------------------------------------------------------|---------------------------------------------------------------------------------------------------------------------|
| `attributesprocessor` [configuration help][attributesprocessor_help]               | https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.35.0/processor/attributesprocessor          |
| `groupbyattrsprocessor` [configuration help][groupbyattrsprocessor_help]           | https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.35.0/processor/groupbyattrsprocessor      |
| `groupbytraceprocessor` [configuration help][groupbytraceprocessor_help]           | https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.35.0/processor/groupbytraceprocessor      |
| `metricstransformprocessor` [configuration help][metricstransformprocessor_help]   | https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.35.0/processor/metricstransformprocessor  |
| `resourcedetectionprocessor` [configuration help][resourcedetectionprocessor_help] | https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.35.0/processor/resourcedetectionprocessor |
| `resourceprocessor` [configuration help][resourceprocessor_help]                   | https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.35.0/processor/resourceprocessor          |
| `routingprocessor` [configuration help][routingprocessor_help]                     | https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.35.0/processor/routingprocessor           |
| `spanmetricsprocessor` [configuration help][spanmetricsprocessor_help]             | https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.35.0/processor/spanmetricsprocessor       |
| `tailsamplingprocessor` [configuration help][tailsamplingprocessor_help]           | https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.35.0/processor/tailsamplingprocessor      |
| `filterprocessor` [configuration help][filterprocessor_help]                       | https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.35.0/processor/filterprocessor            |

[attributesprocessor_help]: https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/processor/attributesprocessor
[groupbyattrsprocessor_help]: https://github.com/SumoLogic/sumologic-otel-collector/blob/main/docs/Configuration.md#group-by-attributes-processor
[groupbytraceprocessor_help]: https://github.com/SumoLogic/sumologic-otel-collector/blob/main/docs/Configuration.md#group-by-trace-processor
[metricstransformprocessor_help]: https://github.com/SumoLogic/sumologic-otel-collector/blob/main/docs/Configuration.md#metrics-transform-processor
[resourcedetectionprocessor_help]: https://github.com/SumoLogic/sumologic-otel-collector/blob/main/docs/Configuration.md#resource-detection-processor
[resourceprocessor_help]: https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/processor/resourceprocessor
[routingprocessor_help]: https://github.com/SumoLogic/sumologic-otel-collector/blob/main/docs/Configuration.md#routing-processor-processor
[spanmetricsprocessor_help]: https://github.com/SumoLogic/sumologic-otel-collector/blob/main/docs/Configuration.md#span-metrics-processor
[tailsamplingprocessor_help]: https://github.com/SumoLogic/sumologic-otel-collector/blob/main/docs/Configuration.md#tail-sampling-processor
[filterprocessor_help]: https://github.com/SumoLogic/sumologic-otel-collector/blob/main/docs/Configuration.md#tail-sampling-processor

### Exporters

#### Sumo Logic supported exporters

| Name                                                             | Source                                                                                         |
|------------------------------------------------------------------|------------------------------------------------------------------------------------------------|
| `sumologicexporter` [configuration help][sumologicexporter_help] | https://github.com/SumoLogic/sumologic-otel-collector/tree/main/pkg/exporter/sumologicexporter |

[sumologicexporter_help]: https://github.com/SumoLogic/sumologic-otel-collector/blob/main/docs/Configuration.md#sumo-logic-exporter

#### Upstream exporters

| Name                                                                     | Source                                                                                                        |
|--------------------------------------------------------------------------|---------------------------------------------------------------------------------------------------------------|
| `loadbalancingexporter` [configuration help][loadbalancingexporter_help] | https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.35.0/exporter/loadbalancingexporter |
| `loggingexporter` [configuration help][loggingexporter_help]             | https://github.com/open-telemetry/opentelemetry-collector/tree/v0.35.0/exporter/loggingexporter               |

[loadbalancingexporter_help]: https://github.com/SumoLogic/sumologic-otel-collector/blob/main/docs/Configuration.md#load-balancing-exporter
[loggingexporter_help]: https://github.com/SumoLogic/sumologic-otel-collector/blob/main/docs/Configuration.md#logging-exporter

### Extensions

#### Sumo Logic supported extensions

| Name                                                               | Source                                                                                           |
|--------------------------------------------------------------------|--------------------------------------------------------------------------------------------------|
| `sumologicextension` [configuration help][sumologicextension_help] | https://github.com/SumoLogic/sumologic-otel-collector/tree/main/pkg/extension/sumologicextension |

[sumologicextension_help]: https://github.com/SumoLogic/sumologic-otel-collector/blob/main/docs/Configuration.md#sumo-logic-extension

#### Upstream extensions

| Name                       | Source                                                                                                            |
|----------------------------|-------------------------------------------------------------------------------------------------------------------|
| `ballastextension`         | https://github.com/open-telemetry/opentelemetry-collector/tree/v0.35.0/extension/ballastextension                 |
| `bearertokenauthextension` | https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.35.0/extension/bearertokenauthextension |
| `healthcheckextension`     | https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.35.0/extension/healthcheckextension     |
| `oidcauthextension`        | https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.35.0/extension/oidcauthextension        |
| `pprofextension`           | https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.35.0/extension/pprofextension           |
| `zpagesextension`          | https://github.com/open-telemetry/opentelemetry-collector/tree/v0.35.0/extension/zpagesextension                  |

<!-- markdownlint-enable MD013 -->

## Contributing

For contributing guidelines, see [CONTRIBUTING](./CONTRIBUTING.md).
