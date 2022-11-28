# Sumo Logic Distribution for OpenTelemetry Collector

[![Default branch build](https://github.com/SumoLogic/sumologic-otel-collector/actions/workflows/dev_builds.yml/badge.svg)](https://github.com/SumoLogic/sumologic-otel-collector/actions/workflows/dev_builds.yml)

Sumo Logic Distribution for OpenTelemetry Collector is a Sumo Logic-supported distribution of the [OpenTelemetry Collector][otc_link].
It is a single agent to send logs, metrics and traces to [Sumo Logic][sumologic].

**Our aim is to extend and not to replace the OpenTelemetry Collector.**

In order to learn more, pleasee see [purpose of Sumo Logic Distribution for OpenTelemetry Collector](./docs/upstream-relation.md#purpose-of-sumo-logic-distribution-for-opentelemetry-collector)

[otc_link]: https://github.com/open-telemetry/opentelemetry-collector
[sumologic]: https://www.sumologic.com

- [Installation](docs/installation.md)
- [Configuration](docs/configuration.md)
- [Migration from Installed Collector](docs/migration.md)
- [Comparison between the Installed Collector and OpenTelemetry Collector](docs/comparison.md)
- [OpenTelemetry Collector builder](./otelcolbuilder/README.md)
- [Performance](docs/performance.md)
- [Known Issues](docs/known-issues.md)
- [Contributing](./CONTRIBUTING.md)
- [Changelog](./CHANGELOG.md)
- [Purpose of Sumo Logic Distribution for OpenTelemetry Collector](./docs/upstream-relation.md#purpose-of-sumo-logic-distribution-for-opentelemetry-collector)
- [Versioning policy](./docs/upstream-relation.md#versioning-policy)
- [Breaking changes policy](./docs/upstream-relation.md#breaking-changes-policy)

## Supported OS and architectures

| Linux                         | MacOS                         |
|-------------------------------|-------------------------------|
| [amd64 (x86_64)][linux_amd64] | [amd64 (x86_64)][mac_amd64]   |
| [arm64][linux_arm64]          | [arm64 (Apple M1)][mac_arm64] |

[linux_amd64]: ./docs/installation.md#linux-on-amd64-x86-64
[linux_arm64]: ./docs/installation.md#linux-on-arm64
[mac_amd64]: ./docs/installation.md#macos-on-amd64-x86-64
[mac_arm64]: ./docs/installation.md#macos-on-arm64-apple-m1-x86-64

## Components

This section lists the components that are included in Sumo Logic Distribution for OpenTelemetry Collector.

The `highlighted` components are delivered by Sumo Logic.

The components with an asterisk `*` are upstream OpenTelemetry components with a minor addition by Sumo Logic.

The rest of the components in the table are pure upstream OpenTelemetry components.

|                         Receivers                          |                          Processors                          |                Exporters                 |                    Extensions                    |
|:----------------------------------------------------------:|:------------------------------------------------------------:|:----------------------------------------:|:------------------------------------------------:|
|      [active_directory_ds][activedirectorydsreceiver]      |              [attributes][attributesprocessor]               |         [carbon][carbonexporter]         |         [asapclient][asapauthextension]          |
|               [aerospike][aerospikereceiver]               |                   [batch][batchprocessor]                    |           [file][fileexporter]           |               [awsproxy][awsproxy]               |
|                  [apache][apachereceiver]                  |        [`cascading_filter`][cascadingfilterprocessor]        |          [kafka][kafkaexporter]          |         [basicauth][basicauthextension]          |
| [awscontainerinsightreceiver][awscontainerinsightreceiver] |       [cumulativetodelta][cumulativetodeltaprocessor]        |  [loadbalancing][loadbalancingexporter]  |   [bearertokenauth][bearertokenauthextension]    |
|  [awsecscontainermetrics][awsecscontainermetricsreceiver]  |             [deltatorate][deltatorateprocessor]              |        [logging][loggingexporter]        |             [db_storage][dbstorage]              |
|             [awsfirehose][awsfirehosereceiver]             | [experimental_metricsgeneration][metricsgenerationprocessor] |           [otlp][otlpexporter]           |        [docker_observer][dockerobserver]         |
|                 [awsxray][awsxrayreceiver]                 |                  [filter][filterprocessor]                   |       [otlphttp][otlphttpexporter]       |           [ecs_observer][ecsobserver]            |
|                   [bigip][bigipreceiver]                   |            [groupbyattrs][groupbyattrsprocessor]             | [prometheusexporter][prometheusexporter] |       [ecs_task_observer][ecstaskobserver]       |
|                  [carbon][carbonreceiver]                  |            [groupbytrace][groupbytraceprocessor]             |     [`sumologic`][sumologicexporter]     |           [file_storage][filestorage]            |
|            [cloudfoundry][cloudfoundryreceiver]            |                 [`k8s_tagger`][k8sprocessor]                 |                                          | [headerssetterextension][headerssetterextension] |
|                [collectd][collectdreceiver]                |           [k8sattributes][k8sattributesprocessor]            |                                          |       [health_check][healthcheckextension]       |
|                 [couchdb][couchdbreceiver]                 |           [logstransform][logstransformprocessor]            |                                          |          [host_observer][hostobserver]           |
|            [docker_stats][dockerstatsreceiver]             |           [memory_limiter][memorylimiterprocessor]           |                                          |         [http_forwarder][httpforwarder]          |
|      [dotnet_diagnostics][dotnetdiagnosticsreceiver]       |        [`metric_frequency`][metricfrequencyprocessor]        |                                          |   [jaegerremotesampling][jaegerremotesampling]   |
|           [elasticsearch][elasticsearchreceiver]           |        [metricstransform][metricstransformprocessor]         |                                          |           [k8s_observer][k8sobserver]            |
|                  [expvar][expvarreceiver]                  |    [probabilistic_sampler][probabilisticsamplerprocessor]    |                                          |        [memory_ballast][ballastextension]        |
|                 [filelog][filelogreceiver]                 |               [redaction][redactionprocessor]                |                                          |    [oauth2client][oauth2clientauthextension]     |
|            [flinkmetrics][flinkmetricsreceiver]            |                [resource][resourceprocessor]                 |                                          |            [oidc][oidcauthextension]             |
|           [fluentforward][fluentforwardreceiver]           |       [resourcedetection][resourcedetectionprocessor]        |                                          |             [pprof][pprofextension]              |
|       [googlecloudpubsub][googlecloudpubsubreceiver]       |                 [routing][routingprocessor]                  |                                          |         [sigv4auth][sigv4authextension]          |
|      [googlecloudspanner][googlecloudspannerreceiver]      |                  [schema][schemaprocessor]                   |                                          |        [`sumologic`][sumologicextension]         |
|             [hostmetrics][hostmetricsreceiver]             |                 [`source`][sourceprocessor]                  |                                          |            [zpages][zpagesextension]             |
|                     [iis][iisreceiver]                     |                    [span][spanprocessor]                     |                                          |                                                  |
|               [httpcheck][httpcheckreceiver]               |             [spanmetrics][spanmetricsprocessor]              |                                          |                                                  |
|                [influxdb][influxdbreceiver]                |        [`sumologic_schema`][sumologicschemaprocessor]        |                                          |                                                  |
|                  [jaeger][jaegerreceiver]                  |        [`sumologic_syslog`][sumologicsyslogprocessor]        |                                          |                                                  |
|                     [jmx][jmxreceiver]                     |            [tail_sampling][tailsamplingprocessor]            |                                          |                                                  |
|                [journald][journaldreceiver]                |               [transform][transformprocessor]                |                                          |                                                  |
|             [k8s_cluster][k8sclusterreceiver]              |                                                              |                                          |                                                  |
|              [k8s_events][k8seventsreceiver]               |                                                              |                                          |                                                  |
|                   [kafka][kafkareceiver]                   |                                                              |                                          |                                                  |
|            [kafkametrics][kafkametricsreceiver]            |                                                              |                                          |                                                  |
|            [kubeletstats][kubeletstatsreceiver]            |                                                              |                                          |                                                  |
|               [memcached][memcachedreceiver]               |                                                              |                                          |                                                  |
|                 [mongodb][mongodbreceiver]                 |                                                              |                                          |                                                  |
|            [mongodbatlas][mongodbatlasreceiver]            |                                                              |                                          |                                                  |
|                   [mysql][mysqlreceiver]                   |                                                              |                                          |                                                  |
|                   [nginx][nginxreceiver]                   |                                                              |                                          |                                                  |
|                    [nsxt][nsxtreceiver]                    |                                                              |                                          |                                                  |
|              [opencensus][opencensusreceiver]              |                                                              |                                          |                                                  |
|                    [otlp][otlpreceiver]                    |                                                              |                                          |                                                  |
|               [podman_stats][podmanreceiver]               |                                                              |                                          |                                                  |
|              [postgresql][postgresqlreceiver]              |                                                              |                                          |                                                  |
|       [prometheus_simple][simpleprometheusreceiver]        |                                                              |                                          |                                                  |
|              [prometheus][prometheusreceiver]              |                                                              |                                          |                                                  |
|                [rabbitmq][rabbitmqreceiver]                |                                                              |                                          |                                                  |
|          [`raw_k8s_events`][rawk8seventsreceiver]          |                                                              |                                          |                                                  |
|            [receiver_creator][receivercreator]             |                                                              |                                          |                                                  |
|                   [redis][redisreceiver]                   |                                                              |                                          |                                                  |
|                    [riak][riakreceiver]                    |                                                              |                                          |                                                  |
|                 [saphana][saphanareceiver]                 |                                                              |                                          |                                                  |
|                    [sapm][sapmreceiver]                    |                                                              |                                          |                                                  |
|                [signalfx][signalfxreceiver]                |                                                              |                                          |                                                  |
|              [skywalking][skywalkingreceiver]              |                                                              |                                          |                                                  |
|              [splunk_hec][splunkhecreceiver]               |                                                              |                                          |                                                  |
|                [sqlquery][sqlqueryreceiver]                |                                                              |                                          |                                                  |
|               [sqlserver][sqlserverreceiver]               |                                                              |                                          |                                                  |
|                  [statsd][statsdreceiver]                  |                                                              |                                          |                                                  |
|                  [syslog][syslogreceiver]                  |                                                              |                                          |                                                  |
|                  [tcplog][tcplogreceiver]                  |                                                              |                                          |                                                  |
|               [`telegraf`][telegrafreceiver]               |                                                              |                                          |                                                  |
|                  [udplog][udplogreceiver]                  |                                                              |                                          |                                                  |
|                 [vcenter][vcenterreceiver]                 |                                                              |                                          |                                                  |
|               [wavefront][wavefrontreceiver]               |                                                              |                                          |                                                  |
|         [windowseventlog][windowseventlogreceiver]         |                                                              |                                          |                                                  |
|     [windowsperfcounters][windowsperfcountersreceiver]     |                                                              |                                          |                                                  |
|                  [zipkin][zipkinreceiver]                  |                                                              |                                          |                                                  |
|               [zookeeper][zookeeperreceiver]               |                                                              |                                          |                                                  |

[activedirectorydsreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.66.0/receiver/activedirectorydsreceiver
[aerospikereceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.66.0/receiver/aerospikereceiver
[apachereceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.66.0/receiver/apachereceiver
[awscontainerinsightreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.66.0/receiver/awscontainerinsightreceiver
[awsecscontainermetricsreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.66.0/receiver/awsecscontainermetricsreceiver
[awsfirehosereceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.66.0/receiver/awsfirehosereceiver
[awsxrayreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.66.0/receiver/awsxrayreceiver
[bigipreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.66.0/receiver/bigipreceiver
[carbonreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.66.0/receiver/carbonreceiver
[cloudfoundryreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.66.0/receiver/cloudfoundryreceiver
[collectdreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.66.0/receiver/collectdreceiver
[couchdbreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.66.0/receiver/couchdbreceiver
[dockerstatsreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.66.0/receiver/dockerstatsreceiver
[dotnetdiagnosticsreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.66.0/receiver/dotnetdiagnosticsreceiver
[elasticsearchreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.66.0/receiver/elasticsearchreceiver
[expvarreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.66.0/receiver/expvarreceiver
[filelogreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.66.0/receiver/filelogreceiver
[flinkmetricsreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.66.0/receiver/flinkmetricsreceiver
[fluentforwardreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.66.0/receiver/fluentforwardreceiver
[googlecloudpubsubreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.66.0/receiver/googlecloudpubsubreceiver
[googlecloudspannerreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.66.0/receiver/googlecloudspannerreceiver
[hostmetricsreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.66.0/receiver/hostmetricsreceiver
[httpcheckreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.66.0/receiver/httpcheckreceiver
[iisreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.66.0/receiver/iisreceiver
[influxdbreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.66.0/receiver/influxdbreceiver
[jaegerreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.66.0/receiver/jaegerreceiver
[jmxreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.66.0/receiver/jmxreceiver
[journaldreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.66.0/receiver/journaldreceiver
[k8sclusterreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.66.0/receiver/k8sclusterreceiver
[k8seventsreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.66.0/receiver/k8seventsreceiver
[kafkareceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.66.0/receiver/kafkareceiver
[kafkametricsreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.66.0/receiver/kafkametricsreceiver
[kubeletstatsreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.66.0/receiver/kubeletstatsreceiver
[memcachedreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.66.0/receiver/memcachedreceiver
[mongodbreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.66.0/receiver/mongodbreceiver
[mongodbatlasreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.66.0/receiver/mongodbatlasreceiver
[mysqlreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.66.0/receiver/mysqlreceiver
[nginxreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.66.0/receiver/nginxreceiver
[nsxtreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.66.0/receiver/nsxtreceiver
[opencensusreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.66.0/receiver/opencensusreceiver
[otlpreceiver]: https://github.com/open-telemetry/opentelemetry-collector/tree/v0.66.0/receiver/otlpreceiver
[podmanreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.66.0/receiver/podmanreceiver
[postgresqlreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.66.0/receiver/postgresqlreceiver
[simpleprometheusreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.66.0/receiver/simpleprometheusreceiver
[prometheusreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.66.0/receiver/prometheusreceiver
[rabbitmqreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.66.0/receiver/rabbitmqreceiver
[rawk8seventsreceiver]: ./pkg/receiver/rawk8seventsreceiver
[receivercreator]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.66.0/receiver/receivercreator
[redisreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.66.0/receiver/redisreceiver
[riakreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.66.0/receiver/riakreceiver
[saphanareceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.66.0/receiver/saphanareceiver
[sapmreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.66.0/receiver/sapmreceiver
[signalfxreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.66.0/receiver/signalfxreceiver
[skywalkingreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.66.0/receiver/skywalkingreceiver
[splunkhecreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.66.0/receiver/splunkhecreceiver
[sqlqueryreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.66.0/receiver/sqlqueryreceiver
[sqlserverreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.66.0/receiver/sqlserverreceiver
[statsdreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.66.0/receiver/statsdreceiver
[syslogreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.66.0/receiver/syslogreceiver
[tcplogreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.66.0/receiver/tcplogreceiver
[telegrafreceiver]: ./pkg/receiver/telegrafreceiver
[udplogreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.66.0/receiver/udplogreceiver
[vcenterreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.66.0/receiver/vcenterreceiver
[wavefrontreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.66.0/receiver/wavefrontreceiver
[windowseventlogreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.66.0/receiver/windowseventlogreceiver
[windowsperfcountersreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.66.0/receiver/windowsperfcountersreceiver
[zipkinreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.66.0/receiver/zipkinreceiver
[zookeeperreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.66.0/receiver/zookeeperreceiver

[attributesprocessor]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.66.0/processor/attributesprocessor
[batchprocessor]: https://github.com/open-telemetry/opentelemetry-collector/tree/v0.66.0/processor/batchprocessor
[cascadingfilterprocessor]: ./pkg/processor/cascadingfilterprocessor
[cumulativetodeltaprocessor]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.66.0/processor/cumulativetodeltaprocessor
[deltatorateprocessor]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.66.0/processor/deltatorateprocessor
[metricsgenerationprocessor]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.66.0/processor/metricsgenerationprocessor

[filterprocessor]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.66.0/processor/filterprocessor
[groupbyattrsprocessor]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.66.0/processor/groupbyattrsprocessor
[groupbytraceprocessor]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.66.0/processor/groupbytraceprocessor
[k8sprocessor]: ./pkg/processor/k8sprocessor
[k8sattributesprocessor]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.66.0/processor/k8sattributesprocessor
[logstransformprocessor]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.66.0/processor/logstransformprocessor
[memorylimiterprocessor]: https://github.com/open-telemetry/opentelemetry-collector/tree/v0.66.0/processor/memorylimiterprocessor
[metricfrequencyprocessor]: ./pkg/processor/metricfrequencyprocessor
[metricstransformprocessor]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.66.0/processor/metricstransformprocessor
[probabilisticsamplerprocessor]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.66.0/processor/probabilisticsamplerprocessor
[redactionprocessor]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.66.0/processor/redactionprocessor
[resourceprocessor]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.66.0/processor/resourceprocessor
[resourcedetectionprocessor]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.66.0/processor/resourcedetectionprocessor
[routingprocessor]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.66.0/processor/routingprocessor
[schemaprocessor]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.66.0/processor/schemaprocessor
[sourceprocessor]: ./pkg/processor/sourceprocessor
[spanprocessor]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.66.0/processor/spanprocessor
[spanmetricsprocessor]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.66.0/processor/spanmetricsprocessor
[sumologicschemaprocessor]: ./pkg/processor/sumologicschemaprocessor
[sumologicsyslogprocessor]: ./pkg/processor/sumologicsyslogprocessor
[tailsamplingprocessor]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.66.0/processor/tailsamplingprocessor
[transformprocessor]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.66.0/processor/transformprocessor

[carbonexporter]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.66.0/exporter/carbonexporter
[fileexporter]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.66.0/exporter/fileexporter
[kafkaexporter]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.66.0/exporter/kafkaexporter
[loadbalancingexporter]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.66.0/exporter/loadbalancingexporter
[loggingexporter]: https://github.com/open-telemetry/opentelemetry-collector/tree/v0.66.0/exporter/loggingexporter
[otlpexporter]: https://github.com/open-telemetry/opentelemetry-collector/tree/v0.66.0/exporter/otlpexporter
[otlphttpexporter]: https://github.com/open-telemetry/opentelemetry-collector/tree/v0.66.0/exporter/otlphttpexporter
[prometheusexporter]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.66.0/exporter/prometheusexporter
[sumologicexporter]: ./pkg/exporter/sumologicexporter

[asapauthextension]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.66.0/extension/asapauthextension
[awsproxy]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.66.0/extension/awsproxy
[basicauthextension]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.66.0/extension/basicauthextension
[bearertokenauthextension]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.66.0/extension/bearertokenauthextension
[dbstorage]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.66.0/extension/storage/dbstorage
[dockerobserver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.66.0/extension/observer/dockerobserver
[ecsobserver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.66.0/extension/observer/ecsobserver
[ecstaskobserver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.66.0/extension/observer/ecstaskobserver
[filestorage]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.66.0/extension/storage/filestorage
[headerssetterextension]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.66.0/extension/headerssetterextension
[healthcheckextension]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.66.0/extension/healthcheckextension
[hostobserver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.66.0/extension/observer/hostobserver
[httpforwarder]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.66.0/extension/httpforwarder
[jaegerremotesampling]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.66.0/extension/jaegerremotesampling
[k8sobserver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.66.0/extension/observer/k8sobserver
[ballastextension]: https://github.com/open-telemetry/opentelemetry-collector/tree/v0.66.0/extension/ballastextension
[oauth2clientauthextension]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.66.0/extension/oauth2clientauthextension
[oidcauthextension]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.66.0/extension/oidcauthextension
[pprofextension]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.66.0/extension/pprofextension
[sigv4authextension]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.66.0/extension/sigv4authextension
[sumologicextension]: ./pkg/extension/sumologicextension
[zpagesextension]: https://github.com/open-telemetry/opentelemetry-collector/tree/v0.66.0/extension/zpagesextension
