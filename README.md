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
- [Comparison between the Installed Collector and OpenTelemetry Collector](docs/Comparison.md)
- [OpenTelemetry Collector builder](./otelcolbuilder/README.md)
- [Performance](docs/Performance.md)
- [Known Issues](docs/KnownIssues.md)
- [Contributing](./CONTRIBUTING.md)

## Components

This section lists the components that are included in Sumo Logic OT distro.

The `highlighted` components are delivered by Sumo Logic.
The rest of the components in the table are upstream OpenTelemetry components.

|                         Receivers                          |                       Processors                       |               Exporters                |                 Extensions                  |
|:----------------------------------------------------------:|:------------------------------------------------------:|:--------------------------------------:|:-------------------------------------------:|
| [awscontainerinsightreceiver][awscontainerinsightreceiver] |           [attributes][attributesprocessor]            |        [carbon][carbonexporter]        |     [memory_ballast][ballastextension]      |
|  [awsecscontainermetrics][awsecscontainermetricsreceiver]  |     [`cascading_filter`][cascadingfilterprocessor]     |          [file][fileexporter]          | [bearertokenauth][bearertokenauthextension] |
|                 [awsxray][awsxrayreceiver]                 |               [filter][filterprocessor]                |         [kafka][kafkaexporter]         |      [db_storage][dbstorageextension]       |
|                  [carbon][carbonreceiver]                  |         [groupbyattrs][groupbyattrsprocessor]          | [loadbalancing][loadbalancingexporter] |    [file_storage][filestorageextension]     |
|                [collectd][collectdreceiver]                |         [groupbytrace][groupbytraceprocessor]          |       [logging][loggingexporter]       |    [health_check][healthcheckextension]     |
|            [docker_stats][dockerstatsreceiver]             |              [`k8s_tagger`][k8sprocessor]              |          [otlp][otlpexporter]          |          [oidc][oidcauthextension]          |
|      [dotnet_diagnostics][dotnetdiagnosticsreceiver]       |     [`metric_frequency`][metricfrequencyprocessor]     |      [otlphttp][otlphttpexporter]      |           [pprof][pprofextension]           |
|                 [filelog][filelogreceiver]                 |     [metricstransform][metricstransformprocessor]      |    [`sumologic`][sumologicexporter]    |      [`sumologic`][sumologicextension]      |
|           [fluentforward][fluentforwardreceiver]           | [probabilistic_sampler][probabilisticsamplerprocessor] |                                        |          [zpages][zpagesextension]          |
|      [googlecloudspanner][googlecloudspannerreceiver]      |    [resourcedetection][resourcedetectionprocessor]     |                                        |                                             |
|             [hostmetrics][hostmetricsreceiver]             |             [resource][resourceprocessor]              |                                        |                                             |
|                  [jaeger][jaegerreceiver]                  |              [routing][routingprocessor]               |                                        |                                             |
|                     [jmx][jmxreceiver]                     |              [`source`][sourceprocessor]               |                                        |                                             |
|                [journald][journaldreceiver]                |          [spanmetrics][spanmetricsprocessor]           |                                        |                                             |
|                   [kafka][kafkareceiver]                   |                 [span][spanprocessor]                  |                                        |                                             |
|            [kafkametrics][kafkametricsreceiver]            |     [`sumologic_syslog`][sumologicsyslogprocessor]     |                                        |                                             |
|              [opencensus][opencensusreceiver]              |         [tails_ampling][tailsamplingprocessor]         |                                        |                                             |
|               [podman_stats][podmanreceiver]               |                                                        |                                        |                                             |
|              [prometheus][prometheusreceiver]              |                                                        |                                        |                                             |
|       [prometheus_simple][simpleprometheusreceiver]        |                                                        |                                        |                                             |
|            [receiver_creator][receivercreator]             |                                                        |                                        |                                             |
|                   [redis][redisreceiver]                   |                                                        |                                        |                                             |
|                    [sapm][sapmreceiver]                    |                                                        |                                        |                                             |
|                [signalfx][signalfxreceiver]                |                                                        |                                        |                                             |
|              [splunk_hec][splunkhecreceiver]               |                                                        |                                        |                                             |
|                  [syslog][syslogreceiver]                  |                                                        |                                        |                                             |
|                  [statsd][statsdreceiver]                  |                                                        |                                        |                                             |
|                  [tcplog][tcplogreceiver]                  |                                                        |                                        |                                             |
|               [`telegraf`][telegrafreceiver]               |                                                        |                                        |                                             |
|                  [udplog][udplogreceiver]                  |                                                        |                                        |                                             |
|               [wavefront][wavefrontreceiver]               |                                                        |                                        |                                             |
|     [windowsperfcounters][windowsperfcountersreceiver]     |                                                        |                                        |                                             |
|                  [zipkin][zipkinreceiver]                  |                                                        |                                        |                                             |
|               [zookeeper][zookeeperreceiver]               |                                                        |                                        |                                             |

[awscontainerinsightreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.46.0/receiver/awscontainerinsightreceiver
[awsecscontainermetricsreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.46.0/receiver/awsecscontainermetricsreceiver
[awsxrayreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.46.0/receiver/awsxrayreceiver
[carbonreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.46.0/receiver/carbonreceiver
[collectdreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.46.0/receiver/collectdreceiver
[dockerstatsreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.46.0/receiver/dockerstatsreceiver
[dotnetdiagnosticsreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.46.0/receiver/dotnetdiagnosticsreceiver
[filelogreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.46.0/receiver/filelogreceiver
[fluentforwardreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.46.0/receiver/fluentforwardreceiver
[googlecloudspannerreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.46.0/receiver/googlecloudspannerreceiver
[hostmetricsreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.46.0/receiver/hostmetricsreceiver
[jaegerreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.46.0/receiver/jaegerreceiver
[jmxreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.46.0/receiver/jmxreceiver
[journaldreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.46.0/receiver/journaldreceiver
[kafkareceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.46.0/receiver/kafkareceiver
[kafkametricsreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.46.0/receiver/kafkametricsreceiver
[opencensusreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.46.0/receiver/opencensusreceiver
[podmanreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.46.0/receiver/podmanreceiver
[prometheusreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.46.0/receiver/prometheusreceiver
[receivercreator]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.46.0/receiver/receivercreator
[redisreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.46.0/receiver/redisreceiver
[sapmreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.46.0/receiver/sapmreceiver
[signalfxreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.46.0/receiver/signalfxreceiver
[simpleprometheusreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.46.0/receiver/simpleprometheusreceiver
[splunkhecreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.46.0/receiver/splunkhecreceiver
[syslogreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.46.0/receiver/syslogreceiver
[statsdreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.46.0/receiver/statsdreceiver
[tcplogreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.46.0/receiver/tcplogreceiver
[telegrafreceiver]: ./pkg/receiver/telegrafreceiver
[udplogreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.46.0/receiver/udplogreceiver
[wavefrontreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.46.0/receiver/wavefrontreceiver
[windowsperfcountersreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.46.0/receiver/windowsperfcountersreceiver
[zipkinreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.46.0/receiver/zipkinreceiver
[zookeeperreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.46.0/receiver/zookeeperreceiver

[attributesprocessor]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.46.0/processor/attributesprocessor
[cascadingfilterprocessor]: ./pkg/processor/cascadingfilterprocessor
[filterprocessor]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.46.0/processor/filterprocessor
[groupbyattrsprocessor]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.46.0/processor/groupbyattrsprocessor
[groupbytraceprocessor]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.46.0/processor/groupbytraceprocessor
[k8sprocessor]: ./pkg/processor/k8sprocessor
[metricfrequencyprocessor]: ./pkg/processor/metricfrequencyprocessor
[metricstransformprocessor]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.46.0/processor/metricstransformprocessor
[probabilisticsamplerprocessor]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.46.0/processor/probabilisticsamplerprocessor
[resourcedetectionprocessor]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.46.0/processor/resourcedetectionprocessor
[resourceprocessor]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.46.0/processor/resourceprocessor
[routingprocessor]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.46.0/processor/routingprocessor
[sourceprocessor]: ./pkg/processor/sourceprocessor
[spanmetricsprocessor]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.46.0/processor/spanmetricsprocessor
[spanprocessor]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.46.0/processor/spanprocessor
[sumologicsyslogprocessor]: ./pkg/processor/sumologicsyslogprocessor
[tailsamplingprocessor]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.46.0/processor/tailsamplingprocessor

[carbonexporter]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.46.0/exporter/carbonexporter
[fileexporter]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.46.0/exporter/fileexporter
[kafkaexporter]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.46.0/exporter/kafkaexporter
[loadbalancingexporter]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.46.0/exporter/loadbalancingexporter
[loggingexporter]: https://github.com/open-telemetry/opentelemetry-collector/tree/v0.46.0/exporter/loggingexporter
[otlpexporter]: https://github.com/open-telemetry/opentelemetry-collector/tree/v0.46.0/exporter/otlpexporter
[otlphttpexporter]: https://github.com/open-telemetry/opentelemetry-collector/tree/v0.46.0/exporter/otlphttpexporter
[sumologicexporter]: ./pkg/exporter/sumologicexporter

[ballastextension]: https://github.com/open-telemetry/opentelemetry-collector/tree/v0.46.0/extension/ballastextension
[bearertokenauthextension]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.46.0/extension/bearertokenauthextension
[dbstorageextension]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.46.0/extension/storage/dbstorage
[filestorageextension]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.46.0/extension/storage/filestorage
[healthcheckextension]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.46.0/extension/healthcheckextension
[oidcauthextension]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.46.0/extension/oidcauthextension
[pprofextension]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.46.0/extension/pprofextension
[sumologicextension]: ./pkg/extension/sumologicextension
[zpagesextension]: https://github.com/open-telemetry/opentelemetry-collector/tree/v0.46.0/extension/zpagesextension
