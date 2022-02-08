# Sumo Logic Distribution of OpenTelemetry

[![Default branch build](https://github.com/SumoLogic/sumologic-otel-collector/actions/workflows/dev_builds.yml/badge.svg)](https://github.com/SumoLogic/sumologic-otel-collector/actions/workflows/dev_builds.yml)

Sumo Logic Distro of [OpenTelemetry Collector][otc_link] built with
[opentelemetry-collector-builder][otc_builder_link].

[otc_link]: https://github.com/open-telemetry/opentelemetry-collector
[otc_builder_link]: https://github.com/open-telemetry/opentelemetry-collector-builder

**This software is currently in beta and is not recommended for production environments.**
**If you wish to participate in this beta, please contact your Sumo Logic account team or Sumo Logic Support.**

- [Installation](docs/Installation.md)
- [Configuration](docs/Configuration.md)
- [Migration from Installed Collector](docs/Migration.md)
- [Differences between Installed Collector and Opentelemetry Collector](docs/Comparison.md)
- [Open Telemetry collector builder](./otelcolbuilder/README.md)
- [Built-in Components](#built-in-components)
  - [Receivers](#receivers)
  - [Processors](#processors)
  - [Exporters](#exporters)
  - [Extensions](#extensions)
- [Performance](docs/Performance.md)
- [Known Issues](docs/KnownIssues.md)
- [Contributing](#contributing-guide)

## Built-in Components

This sections represents the supported components that are included in Sumo Logic
OT distro.

<!-- markdownlint-disable MD013 -->
<!-- markdownlint-disable MD034 -->

### Receivers

#### Sumo Logic custom receivers

| Name                                                           | Source                                                                                        |
|----------------------------------------------------------------|-----------------------------------------------------------------------------------------------|
| `telegrafreceiver` [configuration help][telegrafreceiver_help] | https://github.com/SumoLogic/sumologic-otel-collector/tree/main/pkg/receiver/telegrafreceiver |

[telegrafreceiver_help]: ./docs/Configuration.md#telegraf-receiver

#### Upstream receivers

| Name                                                                     | Source                                                                                                                 |
|--------------------------------------------------------------------------|------------------------------------------------------------------------------------------------------------------------|
| `awscontainerinsightreceiver`                                            | https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.42.0/receiver/awscontainerinsightreceiver    |
| `awsecscontainermetricsreceiver`                                         | https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.42.0/receiver/awsecscontainermetricsreceiver |
| `awsxrayreceiver`                                                        | https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.42.0/receiver/awsxrayreceiver                |
| `carbonreceiver`                                                         | https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.42.0/receiver/carbonreceiver                 |
| `collectdreceiver`                                                       | https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.42.0/receiver/collectdreceiver               |
| `dockerstatsreceiver`                                                    | https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.42.0/receiver/dockerstatsreceiver            |
| `dotnetdiagnosticsreceiver`                                              | https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.42.0/receiver/dotnetdiagnosticsreceiver      |
| `filelogreceiver` [configuration help][filelogreceiver_help]             | https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.42.0/receiver/filelogreceiver                |
| `fluentforwardreceiver` [configuration help][fluentforwardreceiver_help] | https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.42.0/receiver/fluentforwardreceiver          |
| `googlecloudspannerreceiver`                                             | https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.42.0/receiver/googlecloudspannerreceiver     |
| `hostmetricsreceiver` [configuration help][hostmetricsreceiver_help]     | https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.42.0/receiver/hostmetricsreceiver            |
| `jaegerreceiver` [configuration help][jaegerreceiver_help]               | https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.42.0/receiver/jaegerreceiver                 |
| `jmxreceiver`                                                            | https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.42.0/receiver/jmxreceiver                    |
| `journaldreceiver`                                                       | https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.42.0/receiver/journaldreceiver               |
| `kafkametricsreceiver`                                                   | https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.42.0/receiver/kafkametricsreceiver           |
| `kafkareceiver`                                                          | https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.42.0/receiver/kafkareceiver                  |
| `opencensusreceiver` [configuration help][opencensusreceiver_help]       | https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.42.0/receiver/opencensusreceiver             |
| `podmanreceiver`                                                         | https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.42.0/receiver/podmanreceiver                 |
| `receivercreator`                                                        | https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.42.0/receiver/receivercreator                |
| `redisreceiver`                                                          | https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.42.0/receiver/redisreceiver                  |
| `sapmreceiver`                                                           | https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.42.0/receiver/sapmreceiver                   |
| `signalfxreceiver`                                                       | https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.42.0/receiver/signalfxreceiver               |
| `splunkhecreceiver`                                                      | https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.42.0/receiver/splunkhecreceiver              |
| `syslogreceiver` [configuration help][syslogreceiver_help]               | https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.42.0/receiver/syslogreceiver                 |
| `statsdreceiver` [configuration help][statsdreceiver_help]               | https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.42.0/receiver/statsdreceiver                 |
| `tcplogreceiver` [configuration help][tcplogreceiver_help]               | https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.42.0/receiver/tcplogreceiver                 |
| `udplogreceiver` [configuration help][udplogreceiver_help]               | https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.42.0/receiver/udplogreceiver                 |
| `wavefrontreceiver`                                                      | https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.42.0/receiver/wavefrontreceiver              |
| `windowsperfcountersreceiver`                                            | https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.42.0/receiver/windowsperfcountersreceiver    |
| `zipkinreceiver` [configuration help][zipkinreceiver_help]               | https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.42.0/receiver/zipkinreceiver                 |
| `zookeeperreceiver`                                                      | https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.42.0/receiver/zookeeperreceiver              |

[filelogreceiver_help]: ./docs/Configuration.md#filelog-receiver
[fluentforwardreceiver_help]: ./docs/Configuration.md#fluent-forward-receiver
[hostmetricsreceiver_help]: ./docs/Configuration.md#host-metrics-receiver
[jaegerreceiver_help]: ./docs/Configuration.md#jaeger-receiver
[opencensusreceiver_help]: ./docs/Configuration.md#opencensus-receiver
[statsdreceiver_help]: ./docs/Configuration.md#statsd-receiver
[syslogreceiver_help]: ./docs/Configuration.md#syslog-receiver
[tcplogreceiver_help]: ./docs/Configuration.md#tcplog-receiver
[udplogreceiver_help]: ./docs/Configuration.md#udplog-receiver
[zipkinreceiver_help]: ./docs/Configuration.md#zipkin-receiver

### Processors

#### Sumo Logic custom processors

| Name                                                                           | Source                                                                                                 |
|--------------------------------------------------------------------------------|--------------------------------------------------------------------------------------------------------|
| `cascadingfilterprocessor` [configuration help][cascadingfilterprocessor_help] | https://github.com/SumoLogic/sumologic-otel-collector/tree/main/pkg/processor/cascadingfilterprocessor |
| `k8sprocessor` [configuration help][k8sprocessor_help]                         | https://github.com/SumoLogic/sumologic-otel-collector/tree/main/pkg/processor/k8sprocessor             |
| `metricfrequencyprocessor` [configuration_help][metricfrequencyprocessor_help] | https://github.com/SumoLogic/sumologic-otel-collector/tree/main/pkg/processor/metricfrequencyprocessor |
| `sourceprocessor` [configuration help][sourceprocessor_help]                   | https://github.com/SumoLogic/sumologic-otel-collector/tree/main/pkg/processor/sourceprocessor          |
| `sumologicsyslogprocessor` [configuration help][sumologicsyslogprocessor_help] | https://github.com/SumoLogic/sumologic-otel-collector/tree/main/pkg/processor/sumologicsyslogprocessor |

[cascadingfilterprocessor_help]: ./docs/Configuration.md#cascading-filter-processor
[k8sprocessor_help]: ./docs/Configuration.md#kubernetes-processor
[metricfrequencyprocessor_help]: ./docs/Configuration.md#metric-frequency-processor
[sourceprocessor_help]: ./docs/Configuration.md#source-processor
[sumologicsyslogprocessor_help]: ./docs/Configuration.md#sumo-logic-syslog-processor

#### Upstream processors

| Name                                                                               | Source                                                                                                                 |
|------------------------------------------------------------------------------------|------------------------------------------------------------------------------------------------------------------------|
| `attributesprocessor` [configuration help][attributesprocessor_help]               | https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.42.0/processor/attributesprocessor           |
| `filterprocessor` [configuration help][filterprocessor_help]                       | https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.42.0/processor/filterprocessor               |
| `groupbyattrsprocessor` [configuration help][groupbyattrsprocessor_help]           | https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.42.0/processor/groupbyattrsprocessor         |
| `groupbytraceprocessor` [configuration help][groupbytraceprocessor_help]           | https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.42.0/processor/groupbytraceprocessor         |
| `metricstransformprocessor` [configuration help][metricstransformprocessor_help]   | https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.42.0/processor/metricstransformprocessor     |
| `probabilisticsamplerprocessor`                                                    | https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.42.0/processor/probabilisticsamplerprocessor |
| `resourcedetectionprocessor` [configuration help][resourcedetectionprocessor_help] | https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.42.0/processor/resourcedetectionprocessor    |
| `resourceprocessor` [configuration help][resourceprocessor_help]                   | https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.42.0/processor/resourceprocessor             |
| `routingprocessor` [configuration help][routingprocessor_help]                     | https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.42.0/processor/routingprocessor              |
| `spanmetricsprocessor` [configuration help][spanmetricsprocessor_help]             | https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.42.0/processor/spanmetricsprocessor          |
| `spanprocessor`                                                                    | https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.42.0/processor/spanprocessor                 |
| `tailsamplingprocessor` [configuration help][tailsamplingprocessor_help]           | https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.42.0/processor/tailsamplingprocessor         |

[attributesprocessor_help]: https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/v0.42.0/processor/attributesprocessor
[groupbyattrsprocessor_help]: ./docs/Configuration.md#group-by-attributes-processor
[groupbytraceprocessor_help]: ./docs/Configuration.md#group-by-trace-processor
[metricstransformprocessor_help]: ./docs/Configuration.md#metrics-transform-processor
[resourcedetectionprocessor_help]: ./docs/Configuration.md#resource-detection-processor
[resourceprocessor_help]: https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/v0.42.0/processor/resourceprocessor
[routingprocessor_help]: ./docs/Configuration.md#routing-processor
[spanmetricsprocessor_help]: ./docs/Configuration.md#span-metrics-processor
[tailsamplingprocessor_help]: ./docs/Configuration.md#tail-sampling-processor
[filterprocessor_help]: ./docs/Configuration.md#filter-processor

### Exporters

#### Sumo Logic custom exporters

| Name                                                             | Source                                                                                         |
|------------------------------------------------------------------|------------------------------------------------------------------------------------------------|
| `sumologicexporter` [configuration help][sumologicexporter_help] | https://github.com/SumoLogic/sumologic-otel-collector/tree/main/pkg/exporter/sumologicexporter |

[sumologicexporter_help]: ./docs/Configuration.md#sumo-logic-exporter

#### Upstream exporters

| Name                                                                     | Source                                                                                                        |
|--------------------------------------------------------------------------|---------------------------------------------------------------------------------------------------------------|
| `carbonexporter` [configuration help][carbonexporter_help]               | https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.42.0/exporter/carbonexporter        |
| `fileexporter` [configuration help][fileexporter_help]                   | https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.42.0/exporter/fileexporter          |
| `kafkaexporter` [configuration help][kafkaexporter_help]                 | https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.42.0/exporter/kafkaexporter         |
| `loadbalancingexporter` [configuration help][loadbalancingexporter_help] | https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.42.0/exporter/loadbalancingexporter |
| `loggingexporter` [configuration help][loggingexporter_help]             | https://github.com/open-telemetry/opentelemetry-collector/tree/v0.42.0/exporter/loggingexporter               |
| `otlpexporter`                                                           | https://github.com/open-telemetry/opentelemetry-collector/tree/v0.42.0/exporter/otlpexporter                  |
| `otlphttpexporter`                                                       | https://github.com/open-telemetry/opentelemetry-collector/tree/v0.42.0/exporter/otlphttpexporter              |

[carbonexporter_help]: ./docs/Configuration.md#carbon-exporter
[fileexporter_help]: ./docs/Configuration.md#file-exporter
[kafkaexporter_help]: ./docs/Configuration.md#kafka-exporter
[loadbalancingexporter_help]: ./docs/Configuration.md#load-balancing-exporter
[loggingexporter_help]: ./docs/Configuration.md#logging-exporter

### Extensions

#### Sumo Logic custom extensions

| Name                                                               | Source                                                                                           |
|--------------------------------------------------------------------|--------------------------------------------------------------------------------------------------|
| `sumologicextension` [configuration help][sumologicextension_help] | https://github.com/SumoLogic/sumologic-otel-collector/tree/main/pkg/extension/sumologicextension |

[sumologicextension_help]: ./docs/Configuration.md#sumo-logic-extension

#### Upstream extensions

| Name                                             | Source                                                                                                            |
|--------------------------------------------------|-------------------------------------------------------------------------------------------------------------------|
| `ballastextension`                               | https://github.com/open-telemetry/opentelemetry-collector/tree/v0.42.0/extension/ballastextension                 |
| `bearertokenauthextension`                       | https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.42.0/extension/bearertokenauthextension |
| `storage` [configuration help][filestorage_help] | https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.42.0/extension/storage                  |
| `healthcheckextension`                           | https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.42.0/extension/healthcheckextension     |
| `oidcauthextension`                              | https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.42.0/extension/oidcauthextension        |
| `pprofextension`                                 | https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.42.0/extension/pprofextension           |
| `zpagesextension`                                | https://github.com/open-telemetry/opentelemetry-collector/tree/v0.42.0/extension/zpagesextension                  |

[filestorage_help]: ./docs/Configuration.md#file-storage-extension

<!-- markdownlint-enable MD013 -->
<!-- markdownlint-enable MD034 -->

## Contributing Guide

- [How to build](#how-to-build)
- [Running Tests](#running-tests)

---

To contribute you will need to ensure you have the following setup:

- working Go environment
- installed `opentelemetry-collector-builder`

  `opentelemetry-collector-builder` can be installed using following command:

  ```bash
  cd otelcolbuilder && \
  sudo make install-builder BUILDER_BIN_PATH=/usr/local/bin/opentelemetry-collector-builder && \
  cd ..
  ```

### How to build

```bash
$ cd otelcolbuilder && make build
opentelemetry-collector-builder \
                --config .otelcol-builder.yaml \
                --output-path ./cmd \
                --name otelcol-sumo
2021-05-24T16:29:03.494+0200    INFO    cmd/root.go:99  OpenTelemetry Collector distribution builder    {"version": "dev", "date": "unknown"}
2021-05-24T16:29:03.498+0200    INFO    builder/main.go:90      Sources created {"path": "./cmd"}
2021-05-24T16:29:03.612+0200    INFO    builder/main.go:126     Getting go modules
2021-05-24T16:29:03.957+0200    INFO    builder/main.go:107     Compiling
2021-05-24T16:29:09.770+0200    INFO    builder/main.go:113     Compiled        {"binary": "./cmd/otelcol-sumo"}
```

In order to build for a different platform one can use `otelcol-sumo-${platform}_${arch}`
make targets e.g.:

```bash
$ cd otelcolbuilder && make otelcol-sumo-linux_arm64
GOOS=linux   GOARCH=arm64 /Library/Developer/CommandLineTools/usr/bin/make build BINARY_NAME=otelcol-sumo-linux_arm64
opentelemetry-collector-builder \
                --config .otelcol-builder.yaml \
                --output-path ./cmd \
                --name otelcol-sumo-linux_arm64
2021-05-24T16:32:11.963+0200    INFO    cmd/root.go:99  OpenTelemetry Collector distribution builder    {"version": "dev", "date": "unknown"}
2021-05-24T16:32:11.965+0200    INFO    builder/main.go:90      Sources created {"path": "./cmd"}
2021-05-24T16:32:12.066+0200    INFO    builder/main.go:126     Getting go modules
2021-05-24T16:32:12.376+0200    INFO    builder/main.go:107     Compiling
2021-05-24T16:32:37.326+0200    INFO    builder/main.go:113     Compiled        {"binary": "./cmd/otelcol-sumo-linux_arm64"}
```

### Running Tests

In order to run tests run `make gotest` in root directory of this repository.
This will run tests in every module from this repo by running `make test` in its
directory.
