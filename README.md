# Sumo Logic OpenTelemetry Distro Collector

[![Default branch build](https://github.com/SumoLogic/sumologic-otel-collector/actions/workflows/dev_builds.yml/badge.svg)](https://github.com/SumoLogic/sumologic-otel-collector/actions/workflows/dev_builds.yml)

Sumo Logic OpenTelemetry Distro Collector is a Sumo Logic-supported distribution of the [OpenTelemetry Collector][otc_link].
It is a single agent to send logs, metrics and traces to [Sumo Logic][sumologic].

**Our aim is to extend and not to replace the OpenTelemetry Collector.**

In order to learn more, pleasee see [purpose of Sumo Logic OT distro](./docs/UpstreamRelation.md#purpose-of-sumo-logic-ot-distro)

[otc_link]: https://github.com/open-telemetry/opentelemetry-collector
[sumologic]: https://www.sumologic.com

- [Installation](docs/Installation.md)
- [Configuration](docs/Configuration.md)
- [Migration from Installed Collector](docs/Migration.md)
- [Comparison between the Installed Collector and OpenTelemetry Collector](docs/Comparison.md)
- [OpenTelemetry Collector builder](./otelcolbuilder/README.md)
- [Performance](docs/Performance.md)
- [Known Issues](docs/KnownIssues.md)
- [Contributing](./CONTRIBUTING.md)
- [Changelog](./CHANGELOG.md)
- [Purpose of Sumo Logic OT Distro](./docs/UpstreamRelation.md#purpose-of-sumo-logic-ot-distro)
- [Versioning policy](./docs/UpstreamRelation.md#versioning-policy)
- [Breaking changes policy](./docs/UpstreamRelation.md#breaking-changes-policy)

## Components

This section lists the components that are included in Sumo Logic OT distro.

The `highlighted` components are delivered by Sumo Logic.

The components with an asterisk `*` are upstream OpenTelemetry components with a minor addition by Sumo Logic.

The rest of the components in the table are pure upstream OpenTelemetry components.

|                         Receivers                          |                        Processors                        |               Exporters                |                 Extensions                  |
|:----------------------------------------------------------:|:--------------------------------------------------------:|:--------------------------------------:|:-------------------------------------------:|
| [awscontainerinsightreceiver][awscontainerinsightreceiver] |            [attributes][attributesprocessor]*            |        [carbon][carbonexporter]        | [bearertokenauth][bearertokenauthextension] |
|  [awsecscontainermetrics][awsecscontainermetricsreceiver]  |                 [batch][batchprocessor]                  |          [file][fileexporter]          |    [file_storage][filestorageextension]     |
|                 [awsxray][awsxrayreceiver]                 |      [`cascading_filter`][cascadingfilterprocessor]      |         [kafka][kafkaexporter]         |    [health_check][healthcheckextension]     |
|                  [carbon][carbonreceiver]                  | [cumulativetodeltaprocessor][cumulativetodeltaprocessor] | [loadbalancing][loadbalancingexporter] |     [memory_ballast][ballastextension]      |
|                [collectd][collectdreceiver]                |                [filter][filterprocessor]*                |       [logging][loggingexporter]       |          [oidc][oidcauthextension]          |
|            [docker_stats][dockerstatsreceiver]             |          [groupbyattrs][groupbyattrsprocessor]           |          [otlp][otlpexporter]          |           [pprof][pprofextension]           |
|      [dotnet_diagnostics][dotnetdiagnosticsreceiver]       |          [groupbytrace][groupbytraceprocessor]           |      [otlphttp][otlphttpexporter]      |      [`sumologic`][sumologicextension]      |
|                 [filelog][filelogreceiver]                 |               [`k8s_tagger`][k8sprocessor]               |    [`sumologic`][sumologicexporter]    |          [zpages][zpagesextension]          |
|           [fluentforward][fluentforwardreceiver]           |         [logstransform][logstransformprocessor]          |                                        |                                             |
|      [googlecloudspanner][googlecloudspannerreceiver]      |         [memory_limiter][memorylimiterprocessor]         |                                        |                                             |
|             [hostmetrics][hostmetricsreceiver]             |      [`metric_frequency`][metricfrequencyprocessor]      |                                        |                                             |
|                  [jaeger][jaegerreceiver]                  |      [metricstransform][metricstransformprocessor]       |                                        |                                             |
|                     [jmx][jmxreceiver]                     |  [probabilistic_sampler][probabilisticsamplerprocessor]  |                                        |                                             |
|                [journald][journaldreceiver]                |              [resource][resourceprocessor]*              |                                        |                                             |
|                   [kafka][kafkareceiver]                   |     [resourcedetection][resourcedetectionprocessor]      |                                        |                                             |
|            [kafkametrics][kafkametricsreceiver]            |               [routing][routingprocessor]                |                                        |                                             |
|              [opencensus][opencensusreceiver]              |               [`source`][sourceprocessor]                |                                        |                                             |
|                    [otlp][otlpreceiver]                    |                  [span][spanprocessor]                   |                                        |                                             |
|               [podman_stats][podmanreceiver]               |           [spanmetrics][spanmetricsprocessor]            |                                        |                                             |
|              [prometheus][prometheusreceiver]              |      [`sumologic_schema`][sumologicschemaprocessor]      |                                        |                                             |
|       [prometheus_simple][simpleprometheusreceiver]        |      [`sumologic_syslog`][sumologicsyslogprocessor]      |                                        |                                             |
|          [`raw_k8_sevents`][rawk8seventsreceiver]          |          [tail_sampling][tailsamplingprocessor]          |                                        |                                             |
|            [receiver_creator][receivercreator]             |                                                          |                                        |                                             |
|                   [redis][redisreceiver]                   |                                                          |                                        |                                             |
|                    [sapm][sapmreceiver]                    |                                                          |                                        |                                             |
|                [signalfx][signalfxreceiver]                |                                                          |                                        |                                             |
|              [splunk_hec][splunkhecreceiver]               |                                                          |                                        |                                             |
|                  [statsd][statsdreceiver]                  |                                                          |                                        |                                             |
|                  [syslog][syslogreceiver]                  |                                                          |                                        |                                             |
|                  [tcplog][tcplogreceiver]                  |                                                          |                                        |                                             |
|               [`telegraf`][telegrafreceiver]               |                                                          |                                        |                                             |
|                  [udplog][udplogreceiver]                  |                                                          |                                        |                                             |
|               [wavefront][wavefrontreceiver]               |                                                          |                                        |                                             |
|     [windowsperfcounters][windowsperfcountersreceiver]     |                                                          |                                        |                                             |
|                  [zipkin][zipkinreceiver]                  |                                                          |                                        |                                             |
|               [zookeeper][zookeeperreceiver]               |                                                          |                                        |                                             |

[awscontainerinsightreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.53.0/receiver/awscontainerinsightreceiver
[awsecscontainermetricsreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.53.0/receiver/awsecscontainermetricsreceiver
[awsxrayreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.53.0/receiver/awsxrayreceiver
[carbonreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.53.0/receiver/carbonreceiver
[collectdreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.53.0/receiver/collectdreceiver
[dockerstatsreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.53.0/receiver/dockerstatsreceiver
[dotnetdiagnosticsreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.53.0/receiver/dotnetdiagnosticsreceiver
[filelogreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.53.0/receiver/filelogreceiver
[fluentforwardreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.53.0/receiver/fluentforwardreceiver
[googlecloudspannerreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.53.0/receiver/googlecloudspannerreceiver
[hostmetricsreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.53.0/receiver/hostmetricsreceiver
[jaegerreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.53.0/receiver/jaegerreceiver
[jmxreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.53.0/receiver/jmxreceiver
[journaldreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.53.0/receiver/journaldreceiver
[kafkareceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.53.0/receiver/kafkareceiver
[kafkametricsreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.53.0/receiver/kafkametricsreceiver
[opencensusreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.53.0/receiver/opencensusreceiver
[otlpreceiver]: https://github.com/open-telemetry/opentelemetry-collector/tree/v0.53.0/receiver/otlpreceiver
[podmanreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.53.0/receiver/podmanreceiver
[prometheusreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.53.0/receiver/prometheusreceiver
[rawk8seventsreceiver]: ./pkg/receiver/rawk8seventsreceiver
[receivercreator]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.53.0/receiver/receivercreator
[redisreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.53.0/receiver/redisreceiver
[sapmreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.53.0/receiver/sapmreceiver
[signalfxreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.53.0/receiver/signalfxreceiver
[simpleprometheusreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.53.0/receiver/simpleprometheusreceiver
[splunkhecreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.53.0/receiver/splunkhecreceiver
[syslogreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.53.0/receiver/syslogreceiver
[statsdreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.53.0/receiver/statsdreceiver
[tcplogreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.53.0/receiver/tcplogreceiver
[telegrafreceiver]: ./pkg/receiver/telegrafreceiver
[udplogreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.53.0/receiver/udplogreceiver
[wavefrontreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.53.0/receiver/wavefrontreceiver
[windowsperfcountersreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.53.0/receiver/windowsperfcountersreceiver
[zipkinreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.53.0/receiver/zipkinreceiver
[zookeeperreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.53.0/receiver/zookeeperreceiver

[attributesprocessor]: https://github.com/SumoLogic/opentelemetry-collector-contrib/tree/v0.53.0-filterprocessor/processor/attributesprocessor
[batchprocessor]: https://github.com/open-telemetry/opentelemetry-collector/tree/v0.53.0/processor/batchprocessor
[cascadingfilterprocessor]: ./pkg/processor/cascadingfilterprocessor
[cumulativetodeltaprocessor]:  https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.53.0/processor/cumulativetodeltaprocessor
[filterprocessor]: https://github.com/SumoLogic/opentelemetry-collector-contrib/tree/v0.53.0-filterprocessor/processor/filterprocessor
[groupbyattrsprocessor]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.53.0/processor/groupbyattrsprocessor
[groupbytraceprocessor]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.53.0/processor/groupbytraceprocessor
[k8sprocessor]: ./pkg/processor/k8sprocessor
[logstransformprocessor]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.53.0/processor/logstransformprocessor
[memorylimiterprocessor]: https://github.com/open-telemetry/opentelemetry-collector/tree/v0.53.0/processor/memorylimiterprocessor
[metricfrequencyprocessor]: ./pkg/processor/metricfrequencyprocessor
[metricstransformprocessor]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.53.0/processor/metricstransformprocessor
[probabilisticsamplerprocessor]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.53.0/processor/probabilisticsamplerprocessor
[resourcedetectionprocessor]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.53.0/processor/resourcedetectionprocessor
[resourceprocessor]: https://github.com/SumoLogic/opentelemetry-collector-contrib/tree/v0.53.0-filterprocessor/processor/resourceprocessor
[routingprocessor]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.53.0/processor/routingprocessor
[sourceprocessor]: ./pkg/processor/sourceprocessor
[spanmetricsprocessor]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.53.0/processor/spanmetricsprocessor
[spanprocessor]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.53.0/processor/spanprocessor
[sumologicschemaprocessor]: ./pkg/processor/sumologicschemaprocessor
[sumologicsyslogprocessor]: ./pkg/processor/sumologicsyslogprocessor
[tailsamplingprocessor]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.53.0/processor/tailsamplingprocessor

[carbonexporter]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.53.0/exporter/carbonexporter
[fileexporter]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.53.0/exporter/fileexporter
[kafkaexporter]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.53.0/exporter/kafkaexporter
[loadbalancingexporter]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.53.0/exporter/loadbalancingexporter
[loggingexporter]: https://github.com/open-telemetry/opentelemetry-collector/tree/v0.53.0/exporter/loggingexporter
[otlpexporter]: https://github.com/open-telemetry/opentelemetry-collector/tree/v0.53.0/exporter/otlpexporter
[otlphttpexporter]: https://github.com/open-telemetry/opentelemetry-collector/tree/v0.53.0/exporter/otlphttpexporter
[sumologicexporter]: ./pkg/exporter/sumologicexporter

[ballastextension]: https://github.com/open-telemetry/opentelemetry-collector/tree/v0.53.0/extension/ballastextension
[bearertokenauthextension]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.53.0/extension/bearertokenauthextension
[filestorageextension]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.53.0/extension/storage/filestorage
[healthcheckextension]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.53.0/extension/healthcheckextension
[oidcauthextension]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.53.0/extension/oidcauthextension
[pprofextension]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.53.0/extension/pprofextension
[sumologicextension]: ./pkg/extension/sumologicextension
[zpagesextension]: https://github.com/open-telemetry/opentelemetry-collector/tree/v0.53.0/extension/zpagesextension
