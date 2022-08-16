# Sumo Logic Distribution for OpenTelemetry Collector

[![Default branch build](https://github.com/SumoLogic/sumologic-otel-collector/actions/workflows/dev_builds.yml/badge.svg)](https://github.com/SumoLogic/sumologic-otel-collector/actions/workflows/dev_builds.yml)

Sumo Logic Distribution for OpenTelemetry Collector is a Sumo Logic-supported distribution of the [OpenTelemetry Collector][otc_link].
It is a single agent to send logs, metrics and traces to [Sumo Logic][sumologic].

**Our aim is to extend and not to replace the OpenTelemetry Collector.**

In order to learn more, pleasee see [purpose of Sumo Logic Distribution for OpenTelemetry Collector](./docs/UpstreamRelation.md#purpose-of-sumo-logic-distribution-for-opentelemetry-collector)

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
- [Purpose of Sumo Logic Distribution for OpenTelemetry Collector](./docs/UpstreamRelation.md#purpose-of-sumo-logic-distribution-for-opentelemetry-collector)
- [Versioning policy](./docs/UpstreamRelation.md#versioning-policy)
- [Breaking changes policy](./docs/UpstreamRelation.md#breaking-changes-policy)

## Supported OS and architectures

| Linux                         | MacOS                         |
|-------------------------------|-------------------------------|
| [amd64 (x86_64)][linux_amd64] | [amd64 (x86_64)][mac_amd64]   |
| [arm64][linux_arm64]          | [arm64 (Apple M1)][mac_arm64] |

[linux_amd64]: ./docs/Installation.md#linux-on-amd64-x86-64
[linux_arm64]: ./docs/Installation.md#linux-on-arm64
[mac_amd64]: ./docs/Installation.md#macos-on-amd64-x86-64
[mac_arm64]: ./docs/Installation.md#macos-on-arm64-apple-m1-x86-64

## Components

This section lists the components that are included in Sumo Logic Distribution for OpenTelemetry Collector.

The `highlighted` components are delivered by Sumo Logic.

The components with an asterisk `*` are upstream OpenTelemetry components with a minor addition by Sumo Logic.

The rest of the components in the table are pure upstream OpenTelemetry components.

|                         Receivers                          |                          Processors                          |               Exporters                |                  Extensions                  |
|:----------------------------------------------------------:|:------------------------------------------------------------:|:--------------------------------------:|:--------------------------------------------:|
|      [active_directory_ds][activedirectorydsreceiver]      |              [attributes][attributesprocessor]*              |        [carbon][carbonexporter]        |       [asapclient][asapauthextension]        |
|               [aerospike][aerospikereceiver]               |                   [batch][batchprocessor]                    |          [file][fileexporter]          |             [awsproxy][awsproxy]             |
|                  [apache][apachereceiver]                  |        [`cascading_filter`][cascadingfilterprocessor]        |         [kafka][kafkaexporter]         |       [basicauth][basicauthextension]        |
| [awscontainerinsightreceiver][awscontainerinsightreceiver] |       [cumulativetodelta][cumulativetodeltaprocessor]        | [loadbalancing][loadbalancingexporter] | [bearertokenauth][bearertokenauthextension]  |
|  [awsecscontainermetrics][awsecscontainermetricsreceiver]  |             [deltatorate][deltatorateprocessor]              |       [logging][loggingexporter]       |           [db_storage][dbstorage]            |
|             [awsfirehose][awsfirehosereceiver]             | [experimental_metricsgeneration][metricsgenerationprocessor] |          [otlp][otlpexporter]          |      [docker_observer][dockerobserver]       |
|                 [awsxray][awsxrayreceiver]                 |                  [filter][filterprocessor]                   |      [otlphttp][otlphttpexporter]      |         [ecs_observer][ecsobserver]          |
|                   [bigip][bigipreceiver]                   |            [groupbyattrs][groupbyattrsprocessor]             |    [`sumologic`][sumologicexporter]    |     [ecs_task_observer][ecstaskobserver]     |
|                  [carbon][carbonreceiver]                  |            [groupbytrace][groupbytraceprocessor]             |                                        |         [file_storage][filestorage]          |
|            [cloudfoundry][cloudfoundryreceiver]            |                 [`k8s_tagger`][k8sprocessor]                 |                                        |     [health_check][healthcheckextension]     |
|                [collectd][collectdreceiver]                |           [k8sattributes][k8sattributesprocessor]            |                                        |        [host_observer][hostobserver]         |
|                 [couchdb][couchdbreceiver]                 |           [logstransform][logstransformprocessor]            |                                        |       [http_forwarder][httpforwarder]        |
|            [docker_stats][dockerstatsreceiver]             |           [memory_limiter][memorylimiterprocessor]           |                                        | [jaegerremotesampling][jaegerremotesampling] |
|      [dotnet_diagnostics][dotnetdiagnosticsreceiver]       |        [`metric_frequency`][metricfrequencyprocessor]        |                                        |         [k8s_observer][k8sobserver]          |
|           [elasticsearch][elasticsearchreceiver]           |        [metricstransform][metricstransformprocessor]         |                                        |      [memory_ballast][ballastextension]      |
|                  [expvar][expvarreceiver]                  |    [probabilistic_sampler][probabilisticsamplerprocessor]    |                                        |  [oauth2client][oauth2clientauthextension]   |
|                 [filelog][filelogreceiver]                 |               [redaction][redactionprocessor]                |                                        |          [oidc][oidcauthextension]           |
|            [flinkmetrics][flinkmetricsreceiver]            |                [resource][resourceprocessor]*                |                                        |           [pprof][pprofextension]            |
|           [fluentforward][fluentforwardreceiver]           |       [resourcedetection][resourcedetectionprocessor]        |                                        |       [sigv4auth][sigv4authextension]        |
|       [googlecloudpubsub][googlecloudpubsubreceiver]       |                 [routing][routingprocessor]                  |                                        |      [`sumologic`][sumologicextension]       |
|      [googlecloudspanner][googlecloudspannerreceiver]      |                  [schema][schemaprocessor]                   |                                        |          [zpages][zpagesextension]           |
|             [hostmetrics][hostmetricsreceiver]             |                 [`source`][sourceprocessor]                  |                                        |                                              |
|                     [iis][iisreceiver]                     |                    [span][spanprocessor]                     |                                        |                                              |
|                [influxdb][influxdbreceiver]                |             [spanmetrics][spanmetricsprocessor]              |                                        |                                              |
|                  [jaeger][jaegerreceiver]                  |        [`sumologic_schema`][sumologicschemaprocessor]        |                                        |                                              |
|                     [jmx][jmxreceiver]                     |        [`sumologic_syslog`][sumologicsyslogprocessor]        |                                        |                                              |
|                [journald][journaldreceiver]                |            [tail_sampling][tailsamplingprocessor]            |                                        |                                              |
|             [k8s_cluster][k8sclusterreceiver]              |               [transform][transformprocessor]                |                                        |                                              |
|              [k8s_events][k8seventsreceiver]               |                                                              |                                        |                                              |
|                   [kafka][kafkareceiver]                   |                                                              |                                        |                                              |
|            [kafkametrics][kafkametricsreceiver]            |                                                              |                                        |                                              |
|            [kubeletstats][kubeletstatsreceiver]            |                                                              |                                        |                                              |
|               [memcached][memcachedreceiver]               |                                                              |                                        |                                              |
|                 [mongodb][mongodbreceiver]                 |                                                              |                                        |                                              |
|            [mongodbatlas][mongodbatlasreceiver]            |                                                              |                                        |                                              |
|                   [mysql][mysqlreceiver]                   |                                                              |                                        |                                              |
|                   [nginx][nginxreceiver]                   |                                                              |                                        |                                              |
|                    [nsxt][nsxtreceiver]                    |                                                              |                                        |                                              |
|              [opencensus][opencensusreceiver]              |                                                              |                                        |                                              |
|                    [otlp][otlpreceiver]                    |                                                              |                                        |                                              |
|               [podman_stats][podmanreceiver]               |                                                              |                                        |                                              |
|              [postgresql][postgresqlreceiver]              |                                                              |                                        |                                              |
|       [prometheus_simple][simpleprometheusreceiver]        |                                                              |                                        |                                              |
|              [prometheus][prometheusreceiver]              |                                                              |                                        |                                              |
|                [rabbitmq][rabbitmqreceiver]                |                                                              |                                        |                                              |
|          [`raw_k8s_events`][rawk8seventsreceiver]          |                                                              |                                        |                                              |
|            [receiver_creator][receivercreator]             |                                                              |                                        |                                              |
|                   [redis][redisreceiver]                   |                                                              |                                        |                                              |
|                    [riak][riakreceiver]                    |                                                              |                                        |                                              |
|                 [saphana][saphanareceiver]                 |                                                              |                                        |                                              |
|                    [sapm][sapmreceiver]                    |                                                              |                                        |                                              |
|                [signalfx][signalfxreceiver]                |                                                              |                                        |                                              |
|              [skywalking][skywalkingreceiver]              |                                                              |                                        |                                              |
|              [splunk_hec][splunkhecreceiver]               |                                                              |                                        |                                              |
|                [sqlquery][sqlqueryreceiver]                |                                                              |                                        |                                              |
|               [sqlserver][sqlserverreceiver]               |                                                              |                                        |                                              |
|                  [statsd][statsdreceiver]                  |                                                              |                                        |                                              |
|                  [syslog][syslogreceiver]                  |                                                              |                                        |                                              |
|                  [tcplog][tcplogreceiver]                  |                                                              |                                        |                                              |
|               [`telegraf`][telegrafreceiver]               |                                                              |                                        |                                              |
|                  [udplog][udplogreceiver]                  |                                                              |                                        |                                              |
|                 [vcenter][vcenterreceiver]                 |                                                              |                                        |                                              |
|               [wavefront][wavefrontreceiver]               |                                                              |                                        |                                              |
|         [windowseventlog][windowseventlogreceiver]         |                                                              |                                        |                                              |
|     [windowsperfcounters][windowsperfcountersreceiver]     |                                                              |                                        |                                              |
|                  [zipkin][zipkinreceiver]                  |                                                              |                                        |                                              |
|               [zookeeper][zookeeperreceiver]               |                                                              |                                        |                                              |

[activedirectorydsreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.57.2/receiver/activedirectorydsreceiver
[aerospikereceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.57.2/receiver/aerospikereceiver
[apachereceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.57.2/receiver/apachereceiver
[awscontainerinsightreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.57.2/receiver/awscontainerinsightreceiver
[awsecscontainermetricsreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.57.2/receiver/awsecscontainermetricsreceiver
[awsfirehosereceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.57.2/receiver/awsfirehosereceiver
[awsxrayreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.57.2/receiver/awsxrayreceiver
[bigipreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.57.2/receiver/bigipreceiver
[carbonreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.57.2/receiver/carbonreceiver
[cloudfoundryreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.57.2/receiver/cloudfoundryreceiver
[collectdreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.57.2/receiver/collectdreceiver
[couchdbreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.57.2/receiver/couchdbreceiver
[dockerstatsreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.57.2/receiver/dockerstatsreceiver
[dotnetdiagnosticsreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.57.2/receiver/dotnetdiagnosticsreceiver
[elasticsearchreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.57.2/receiver/elasticsearchreceiver
[expvarreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.57.2/receiver/expvarreceiver
[filelogreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.57.2/receiver/filelogreceiver
[flinkmetricsreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.57.2/receiver/flinkmetricsreceiver
[fluentforwardreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.57.2/receiver/fluentforwardreceiver
[googlecloudpubsubreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.57.2/receiver/googlecloudpubsubreceiver
[googlecloudspannerreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.57.2/receiver/googlecloudspannerreceiver
[hostmetricsreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.57.2/receiver/hostmetricsreceiver
[iisreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.57.2/receiver/iisreceiver
[influxdbreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.57.2/receiver/influxdbreceiver
[jaegerreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.57.2/receiver/jaegerreceiver
[jmxreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.57.2/receiver/jmxreceiver
[journaldreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.57.2/receiver/journaldreceiver
[k8sclusterreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.57.2/receiver/k8sclusterreceiver
[k8seventsreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.57.2/receiver/k8seventsreceiver
[kafkareceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.57.2/receiver/kafkareceiver
[kafkametricsreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.57.2/receiver/kafkametricsreceiver
[kubeletstatsreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.57.2/receiver/kubeletstatsreceiver
[memcachedreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.57.2/receiver/memcachedreceiver
[mongodbreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.57.2/receiver/mongodbreceiver
[mongodbatlasreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.57.2/receiver/mongodbatlasreceiver
[mysqlreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.57.2/receiver/mysqlreceiver
[nginxreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.57.2/receiver/nginxreceiver
[nsxtreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.57.2/receiver/nsxtreceiver
[opencensusreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.57.2/receiver/opencensusreceiver
[otlpreceiver]: https://github.com/open-telemetry/opentelemetry-collector/tree/v0.57.2/receiver/otlpreceiver
[podmanreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.57.2/receiver/podmanreceiver
[postgresqlreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.57.2/receiver/postgresqlreceiver
[simpleprometheusreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.57.2/receiver/simpleprometheusreceiver
[prometheusreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.57.2/receiver/prometheusreceiver
[rabbitmqreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.57.2/receiver/rabbitmqreceiver
[rawk8seventsreceiver]: ./pkg/receiver/rawk8seventsreceiver
[receivercreator]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.57.2/receiver/receivercreator
[redisreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.57.2/receiver/redisreceiver
[riakreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.57.2/receiver/riakreceiver
[saphanareceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.57.2/receiver/saphanareceiver
[sapmreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.57.2/receiver/sapmreceiver
[signalfxreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.57.2/receiver/signalfxreceiver
[skywalkingreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.57.2/receiver/skywalkingreceiver
[splunkhecreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.57.2/receiver/splunkhecreceiver
[sqlqueryreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.57.2/receiver/sqlqueryreceiver
[sqlserverreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.57.2/receiver/sqlserverreceiver
[statsdreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.57.2/receiver/statsdreceiver
[syslogreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.57.2/receiver/syslogreceiver
[tcplogreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.57.2/receiver/tcplogreceiver
[telegrafreceiver]: ./pkg/receiver/telegrafreceiver
[udplogreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.57.2/receiver/udplogreceiver
[vcenterreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.57.2/receiver/vcenterreceiver
[wavefrontreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.57.2/receiver/wavefrontreceiver
[windowseventlogreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.57.2/receiver/windowseventlogreceiver
[windowsperfcountersreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.57.2/receiver/windowsperfcountersreceiver
[zipkinreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.57.2/receiver/zipkinreceiver
[zookeeperreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.57.2/receiver/zookeeperreceiver

[attributesprocessor]: https://github.com/SumoLogic/opentelemetry-collector-contrib/tree/v0.57.2-filterprocessor/processor/attributesprocessor
[batchprocessor]: https://github.com/open-telemetry/opentelemetry-collector/tree/v0.57.2/processor/batchprocessor
[cascadingfilterprocessor]: ./pkg/processor/cascadingfilterprocessor
[cumulativetodeltaprocessor]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.57.2/processor/cumulativetodeltaprocessor
[deltatorateprocessor]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.57.2/processor/deltatorateprocessor
[metricsgenerationprocessor]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.57.2/processor/metricsgenerationprocessor

[filterprocessor]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.57.2/processor/filterprocessor
[groupbyattrsprocessor]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.57.2/processor/groupbyattrsprocessor
[groupbytraceprocessor]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.57.2/processor/groupbytraceprocessor
[k8sprocessor]: ./pkg/processor/k8sprocessor
[k8sattributesprocessor]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.57.2/processor/k8sattributesprocessor
[logstransformprocessor]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.57.2/processor/logstransformprocessor
[memorylimiterprocessor]: https://github.com/open-telemetry/opentelemetry-collector/tree/v0.57.2/processor/memorylimiterprocessor
[metricfrequencyprocessor]: ./pkg/processor/metricfrequencyprocessor
[metricstransformprocessor]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.57.2/processor/metricstransformprocessor
[probabilisticsamplerprocessor]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.57.2/processor/probabilisticsamplerprocessor
[redactionprocessor]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.57.2/processor/redactionprocessor
[resourceprocessor]: https://github.com/SumoLogic/opentelemetry-collector-contrib/tree/v0.57.2-filterprocessor/processor/resourceprocessor
[resourcedetectionprocessor]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.57.2/processor/resourcedetectionprocessor
[routingprocessor]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.57.2/processor/routingprocessor
[schemaprocessor]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.57.2/processor/schemaprocessor
[sourceprocessor]: ./pkg/processor/sourceprocessor
[spanprocessor]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.57.2/processor/spanprocessor
[spanmetricsprocessor]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.57.2/processor/spanmetricsprocessor
[sumologicschemaprocessor]: ./pkg/processor/sumologicschemaprocessor
[sumologicsyslogprocessor]: ./pkg/processor/sumologicsyslogprocessor
[tailsamplingprocessor]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.57.2/processor/tailsamplingprocessor
[transformprocessor]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.57.2/processor/transformprocessor

[carbonexporter]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.57.2/exporter/carbonexporter
[fileexporter]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.57.2/exporter/fileexporter
[kafkaexporter]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.57.2/exporter/kafkaexporter
[loadbalancingexporter]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.57.2/exporter/loadbalancingexporter
[loggingexporter]: https://github.com/open-telemetry/opentelemetry-collector/tree/v0.57.2/exporter/loggingexporter
[otlpexporter]: https://github.com/open-telemetry/opentelemetry-collector/tree/v0.57.2/exporter/otlpexporter
[otlphttpexporter]: https://github.com/open-telemetry/opentelemetry-collector/tree/v0.57.2/exporter/otlphttpexporter
[sumologicexporter]: ./pkg/exporter/sumologicexporter

[asapauthextension]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.57.2/extension/asapauthextension
[awsproxy]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.57.2/extension/awsproxy
[basicauthextension]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.57.2/extension/basicauthextension
[bearertokenauthextension]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.57.2/extension/bearertokenauthextension
[dbstorage]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.57.2/extension/storage/dbstorage
[dockerobserver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.57.2/extension/observer/dockerobserver
[ecsobserver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.57.2/extension/observer/ecsobserver
[ecstaskobserver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.57.2/extension/observer/ecstaskobserver
[filestorage]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.57.2/extension/storage/filestorage
[healthcheckextension]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.57.2/extension/healthcheckextension
[hostobserver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.57.2/extension/observer/hostobserver
[httpforwarder]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.57.2/extension/httpforwarder
[jaegerremotesampling]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.57.2/extension/jaegerremotesampling
[k8sobserver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.57.2/extension/observer/k8sobserver
[ballastextension]: https://github.com/open-telemetry/opentelemetry-collector/tree/v0.57.2/extension/ballastextension
[oauth2clientauthextension]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.57.2/extension/oauth2clientauthextension
[oidcauthextension]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.57.2/extension/oidcauthextension
[pprofextension]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.57.2/extension/pprofextension
[sigv4authextension]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.57.2/extension/sigv4authextension
[sumologicextension]: ./pkg/extension/sumologicextension
[zpagesextension]: https://github.com/open-telemetry/opentelemetry-collector/tree/v0.57.2/extension/zpagesextension
