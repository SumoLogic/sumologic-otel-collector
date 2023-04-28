# Sumo Logic Distribution for OpenTelemetry Collector

[![Default branch build](https://github.com/SumoLogic/sumologic-otel-collector/actions/workflows/dev_builds.yml/badge.svg)](https://github.com/SumoLogic/sumologic-otel-collector/actions/workflows/dev_builds.yml)

Sumo Logic Distribution for OpenTelemetry Collector is a Sumo Logic-supported distribution of the [OpenTelemetry Collector][otc_link].
It is a single agent to send logs, metrics and traces to [Sumo Logic][sumologic].

**Our aim is to extend and not to replace the OpenTelemetry Collector.**

In order to learn more, pleasee see [purpose of Sumo Logic Distribution for OpenTelemetry Collector][purpose]

[otc_link]: https://github.com/open-telemetry/opentelemetry-collector
[sumologic]: https://www.sumologic.com

- Installation
  - [Linux][linux_installation]
  - [MacOS][macos_installation]
  - [Windows][windows_installation]
  - [Container image](/docs/installation.md#container-image)
    - [Important note about local state files when using `sumologicextension`](/docs/installation.md#important-note-about-local-state-files-when-using-sumologicextension)
  - [Ansible](/docs/installation.md#ansible)
  - [Puppet](/docs/installation.md#puppet)
  - [Chef](/docs/installation.md#chef)
- [Configuration](docs/configuration.md)
- [Migration from Installed Collector](docs/migration.md)
- [Comparison between the Installed Collector and OpenTelemetry Collector](docs/comparison.md)
- [OpenTelemetry Collector builder](./otelcolbuilder/README.md)
- [Performance]
- [Known Issues][known issues]
- [Contributing](./CONTRIBUTING.md)
- [Changelog](./CHANGELOG.md)
- [Purpose of Sumo Logic Distribution for OpenTelemetry Collector][purpose]
- [Versioning policy][versioning]
- [Breaking changes policy][breaking]

[linux_installation]: https://help.sumologic.com/docs/send-data/opentelemetry-collector/install-collector-linux/
[macos_installation]: https://help.sumologic.com/docs/send-data/opentelemetry-collector/install-collector-macos/
[windows_installation]:https://help.sumologic.com/docs/send-data/opentelemetry-collector/install-collector-windows/
[performance]: https://help.sumologic.com/docs/send-data/opentelemetry-collector/#performance
[known issues]: https://help.sumologic.com/docs/send-data/opentelemetry-collector/troubleshooting-faq/#known-issues
[purpose]: https://help.sumologic.com/docs/send-data/opentelemetry-collector/sumo-logic-opentelemetry-vs-opentelemetry-upstream-relationship/
[versioning]: https://help.sumologic.com/docs/send-data/opentelemetry-collector/sumo-logic-opentelemetry-vs-opentelemetry-upstream-relationship/#versioning-policy
[breaking]: https://help.sumologic.com/docs/send-data/opentelemetry-collector/sumo-logic-opentelemetry-vs-opentelemetry-upstream-relationship/#versioning-policy

## Supported OS and architectures

| Linux                         | MacOS                         | Windows                     |
| ----------------------------- | ----------------------------- | --------------------------- |
| [amd64 (x86_64)][linux_amd64] | [amd64 (x86_64)][mac_amd64]   | [amd64 (x86_64)][win_amd64] |
| [arm64][linux_arm64]          | [arm64 (Apple M1)][mac_arm64] |                             |

[linux_amd64]: ./docs/installation.md#linux-on-amd64-x86-64
[linux_arm64]: ./docs/installation.md#linux-on-arm64
[mac_amd64]: ./docs/installation.md#macos-on-amd64-x86-64
[mac_arm64]: ./docs/installation.md#macos-on-arm64-apple-m1-x86-64
[win_amd64]: ./docs/installation.md#windows

## Components

This section lists the components that are included in Sumo Logic Distribution for OpenTelemetry Collector.

The `highlighted` components are delivered by Sumo Logic.

The components with an asterisk `*` are upstream OpenTelemetry components with a minor addition by Sumo Logic.

The rest of the components in the table are pure upstream OpenTelemetry components.

|                         Receivers                          |                          Processors                          |                Exporters                 |                    Extensions                    |              Connectors               |
| :--------------------------------------------------------: | :----------------------------------------------------------: | :--------------------------------------: | :----------------------------------------------: | :-----------------------------------: |
|      [active_directory_ds][activedirectorydsreceiver]      |              [attributes][attributesprocessor]               |         [carbon][carbonexporter]         |         [asapclient][asapauthextension]          |      [forward][forwardconnector]      |
|               [aerospike][aerospikereceiver]               |                   [batch][batchprocessor]                    |           [file][fileexporter]           |               [awsproxy][awsproxy]               |        [count][countconnector]        |
|                  [apache][apachereceiver]                  |        [`cascading_filter`][cascadingfilterprocessor]        |          [kafka][kafkaexporter]          |         [basicauth][basicauthextension]          | [servicegraph][servicegraphconnector] |
|       [awscloudwatchreceiver][awscloudwatchreceiver]       |       [cumulativetodelta][cumulativetodeltaprocessor]        |  [loadbalancing][loadbalancingexporter]  |   [bearertokenauth][bearertokenauthextension]    |  [spanmetrics][spanmetricsconnector]  |
| [awscontainerinsightreceiver][awscontainerinsightreceiver] |                 [datadog][datadogprocessor]                  |        [logging][loggingexporter]        |             [db_storage][dbstorage]              |                                       |
|  [awsecscontainermetrics][awsecscontainermetricsreceiver]  |             [deltatorate][deltatorateprocessor]              |           [otlp][otlpexporter]           |        [docker_observer][dockerobserver]         |                                       |
|             [awsfirehose][awsfirehosereceiver]             | [experimental_metricsgeneration][metricsgenerationprocessor] |       [otlphttp][otlphttpexporter]       |           [ecs_observer][ecsobserver]            |                                       |
|                 [awsxray][awsxrayreceiver]                 |                  [filter][filterprocessor]                   | [prometheusexporter][prometheusexporter] |       [ecs_task_observer][ecstaskobserver]       |                                       |
|       [azureeventhubreceiver][azureeventhubreceiver]       |            [groupbyattrs][groupbyattrsprocessor]             |     [`sumologic`][sumologicexporter]     |           [file_storage][filestorage]            |                                       |
|                   [bigip][bigipreceiver]                   |            [groupbytrace][groupbytraceprocessor]             |    [`syslogexporter`][syslogexporter]    | [headerssetterextension][headerssetterextension] |                                       |
|                  [carbon][carbonreceiver]                  |                 [`k8s_tagger`][k8sprocessor]                 |                                          |       [health_check][healthcheckextension]       |                                       |
|              [cloudflare][cloudflarereceiver]              |           [k8sattributes][k8sattributesprocessor]            |                                          |          [host_observer][hostobserver]           |                                       |
|            [cloudfoundry][cloudfoundryreceiver]            |           [logstransform][logstransformprocessor]            |                                          |         [http_forwarder][httpforwarder]          |                                       |
|                [collectd][collectdreceiver]                |           [memory_limiter][memorylimiterprocessor]           |                                          |   [jaegerremotesampling][jaegerremotesampling]   |                                       |
|                 [couchdb][couchdbreceiver]                 |        [`metric_frequency`][metricfrequencyprocessor]        |                                          |           [k8s_observer][k8sobserver]            |                                       |
|                 [datadog][datadogreceiver]                 |        [metricstransform][metricstransformprocessor]         |                                          |        [memory_ballast][ballastextension]        |                                       |
|            [docker_stats][dockerstatsreceiver]             |    [probabilistic_sampler][probabilisticsamplerprocessor]    |                                          |    [oauth2client][oauth2clientauthextension]     |                                       |
|           [elasticsearch][elasticsearchreceiver]           |               [redaction][redactionprocessor]                |                                          |            [oidc][oidcauthextension]             |                                       |
|                  [expvar][expvarreceiver]                  |                [resource][resourceprocessor]                 |                                          |             [pprof][pprofextension]              |                                       |
|                 [filelog][filelogreceiver]                 |       [resourcedetection][resourcedetectionprocessor]        |                                          |         [sigv4auth][sigv4authextension]          |                                       |
|            [flinkmetrics][flinkmetricsreceiver]            |                 [routing][routingprocessor]                  |                                          |        [`sumologic`][sumologicextension]         |                                       |
|           [fluentforward][fluentforwardreceiver]           |                  [schema][schemaprocessor]                   |                                          |            [zpages][zpagesextension]             |                                       |
|       [googlecloudpubsub][googlecloudpubsubreceiver]       |            [servicegraph][servicegraphprocessor]             |                                          |                                                  |                                       |
|      [googlecloudspanner][googlecloudspannerreceiver]      |                 [`source`][sourceprocessor]                  |                                          |                                                  |                                       |
|                 [haproxy][haproxyreceiver]                 |                    [span][spanprocessor]                     |                                          |                                                  |                                       |
|             [hostmetrics][hostmetricsreceiver]             |             [spanmetrics][spanmetricsprocessor]              |                                          |                                                  |                                       |
|               [httpcheck][httpcheckreceiver]               |        [`sumologic_schema`][sumologicschemaprocessor]        |                                          |                                                  |                                       |
|                     [iis][iisreceiver]                     |        [`sumologic_syslog`][sumologicsyslogprocessor]        |                                          |                                                  |                                       |
|                [influxdb][influxdbreceiver]                |            [tail_sampling][tailsamplingprocessor]            |                                          |                                                  |                                       |
|                  [jaeger][jaegerreceiver]                  |               [transform][transformprocessor]                |                                          |                                                  |                                       |
|                     [jmx][jmxreceiver]                     |                                                              |                                          |                                                  |                                       |
|                [journald][journaldreceiver]                |                                                              |                                          |                                                  |                                       |
|             [k8s_cluster][k8sclusterreceiver]              |                                                              |                                          |                                                  |                                       |
|              [k8s_events][k8seventsreceiver]               |                                                              |                                          |                                                  |                                       |
|              [k8sobjects][k8sobjectsreceiver]              |                                                              |                                          |                                                  |                                       |
|                   [kafka][kafkareceiver]                   |                                                              |                                          |                                                  |                                       |
|            [kafkametrics][kafkametricsreceiver]            |                                                              |                                          |                                                  |                                       |
|            [kubeletstats][kubeletstatsreceiver]            |                                                              |                                          |                                                  |                                       |
|                    [loki][lokireceiver]                    |                                                              |                                          |                                                  |                                       |
|               [memcached][memcachedreceiver]               |                                                              |                                          |                                                  |                                       |
|                 [mongodb][mongodbreceiver]                 |                                                              |                                          |                                                  |                                       |
|            [mongodbatlas][mongodbatlasreceiver]            |                                                              |                                          |                                                  |                                       |
|                   [mysql][mysqlreceiver]                   |                                                              |                                          |                                                  |                                       |
|                   [nginx][nginxreceiver]                   |                                                              |                                          |                                                  |                                       |
|                    [nsxt][nsxtreceiver]                    |                                                              |                                          |                                                  |                                       |
|              [opencensus][opencensusreceiver]              |                                                              |                                          |                                                  |                                       |
|                    [otlp][otlpreceiver]                    |                                                              |                                          |                                                  |                                       |
|            [otlpjsonfile][otlpjsonfilereceiver]            |                                                              |                                          |                                                  |                                       |
|               [podman_stats][podmanreceiver]               |                                                              |                                          |                                                  |                                       |
|              [postgresql][postgresqlreceiver]              |                                                              |                                          |                                                  |                                       |
|       [prometheus_simple][simpleprometheusreceiver]        |                                                              |                                          |                                                  |                                       |
|              [prometheus][prometheusreceiver]              |                                                              |                                          |                                                  |                                       |
|                  [purefa][purefareceiver]                  |                                                              |                                          |                                                  |                                       |
|                  [purefb][purefbreceiver]                  |                                                              |                                          |                                                  |                                       |
|                [rabbitmq][rabbitmqreceiver]                |                                                              |                                          |                                                  |                                       |
|          [`raw_k8s_events`][rawk8seventsreceiver]          |                                                              |                                          |                                                  |                                       |
|            [receiver_creator][receivercreator]             |                                                              |                                          |                                                  |                                       |
|                   [redis][redisreceiver]                   |                                                              |                                          |                                                  |                                       |
|                    [riak][riakreceiver]                    |                                                              |                                          |                                                  |                                       |
|                 [saphana][saphanareceiver]                 |                                                              |                                          |                                                  |                                       |
|                    [sapm][sapmreceiver]                    |                                                              |                                          |                                                  |                                       |
|                [signalfx][signalfxreceiver]                |                                                              |                                          |                                                  |                                       |
|              [skywalking][skywalkingreceiver]              |                                                              |                                          |                                                  |                                       |
|                    [snmp][snmpreceiver]                    |                                                              |                                          |                                                  |                                       |
|                  [solace][solacereceiver]                  |                                                              |                                          |                                                  |                                       |
|              [splunk_hec][splunkhecreceiver]               |                                                              |                                          |                                                  |                                       |
|                [sqlquery][sqlqueryreceiver]                |                                                              |                                          |                                                  |                                       |
|               [sqlserver][sqlserverreceiver]               |                                                              |                                          |                                                  |                                       |
|                [sshcheck][sshcheckreceiver]                |                                                              |                                          |                                                  |                                       |
|                  [statsd][statsdreceiver]                  |                                                              |                                          |                                                  |                                       |
|                  [syslog][syslogreceiver]                  |                                                              |                                          |                                                  |                                       |
|                  [tcplog][tcplogreceiver]                  |                                                              |                                          |                                                  |                                       |
|               [`telegraf`][telegrafreceiver]               |                                                              |                                          |                                                  |                                       |
|                  [udplog][udplogreceiver]                  |                                                              |                                          |                                                  |                                       |
|                 [vcenter][vcenterreceiver]                 |                                                              |                                          |                                                  |                                       |
|               [wavefront][wavefrontreceiver]               |                                                              |                                          |                                                  |                                       |
|         [windowseventlog][windowseventlogreceiver]         |                                                              |                                          |                                                  |                                       |
|     [windowsperfcounters][windowsperfcountersreceiver]     |                                                              |                                          |                                                  |                                       |
|                  [zipkin][zipkinreceiver]                  |                                                              |                                          |                                                  |                                       |
|               [zookeeper][zookeeperreceiver]               |                                                              |                                          |                                                  |                                       |

[activedirectorydsreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/receiver/activedirectorydsreceiver
[aerospikereceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/receiver/aerospikereceiver
[apachereceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/receiver/apachereceiver
[awscloudwatchreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/receiver/awscloudwatchreceiver
[awscontainerinsightreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/receiver/awscontainerinsightreceiver
[awsecscontainermetricsreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/receiver/awsecscontainermetricsreceiver
[awsfirehosereceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/receiver/awsfirehosereceiver
[awsxrayreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/receiver/awsxrayreceiver
[azureeventhubreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/receiver/azureeventhubreceiver
[bigipreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/receiver/bigipreceiver
[carbonreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/receiver/carbonreceiver
[cloudfoundryreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/receiver/cloudfoundryreceiver
[collectdreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/receiver/collectdreceiver
[couchdbreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/receiver/couchdbreceiver
[datadogreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/receiver/datadogreceiver
[dockerstatsreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/receiver/dockerstatsreceiver
[dotnetdiagnosticsreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/receiver/dotnetdiagnosticsreceiver
[elasticsearchreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/receiver/elasticsearchreceiver
[expvarreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/receiver/expvarreceiver
[filelogreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/receiver/filelogreceiver
[flinkmetricsreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/receiver/flinkmetricsreceiver
[fluentforwardreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/receiver/fluentforwardreceiver
[googlecloudpubsubreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/receiver/googlecloudpubsubreceiver
[googlecloudspannerreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/receiver/googlecloudspannerreceiver
[haproxyreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/receiver/haproxyreceiver
[hostmetricsreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/receiver/hostmetricsreceiver
[httpcheckreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/receiver/httpcheckreceiver
[iisreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/receiver/iisreceiver
[influxdbreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/receiver/influxdbreceiver
[jaegerreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/receiver/jaegerreceiver
[jmxreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/receiver/jmxreceiver
[journaldreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/receiver/journaldreceiver
[k8sclusterreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/receiver/k8sclusterreceiver
[k8seventsreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/receiver/k8seventsreceiver
[k8sobjectsreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/receiver/k8sobjectsreceiver
[kafkareceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/receiver/kafkareceiver
[kafkametricsreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/receiver/kafkametricsreceiver
[kubeletstatsreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/receiver/kubeletstatsreceiver
[memcachedreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/receiver/memcachedreceiver
[mongodbreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/receiver/mongodbreceiver
[mongodbatlasreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/receiver/mongodbatlasreceiver
[mysqlreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/receiver/mysqlreceiver
[nginxreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/receiver/nginxreceiver
[nsxtreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/receiver/nsxtreceiver
[opencensusreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/receiver/opencensusreceiver
[otlpreceiver]: https://github.com/open-telemetry/opentelemetry-collector/tree/v0.76.0/receiver/otlpreceiver
[otlpjsonfilereceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/receiver/otlpjsonfilereceiver
[podmanreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/receiver/podmanreceiver
[postgresqlreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/receiver/postgresqlreceiver
[simpleprometheusreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/receiver/simpleprometheusreceiver
[prometheusreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/receiver/prometheusreceiver
[purefareceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/receiver/purefareceiver
[purefbreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/receiver/purefbreceiver
[rabbitmqreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/receiver/rabbitmqreceiver
[rawk8seventsreceiver]: ./pkg/receiver/rawk8seventsreceiver
[receivercreator]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/receiver/receivercreator
[redisreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/receiver/redisreceiver
[riakreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/receiver/riakreceiver
[saphanareceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/receiver/saphanareceiver
[sapmreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/receiver/sapmreceiver
[signalfxreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/receiver/signalfxreceiver
[skywalkingreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/receiver/skywalkingreceiver
[snmpreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/receiver/snmpreceiver
[solacereceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/receiver/solacereceiver
[splunkhecreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/receiver/splunkhecreceiver
[sqlqueryreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/receiver/sqlqueryreceiver
[sqlserverreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/receiver/sqlserverreceiver
[sshcheckreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/receiver/sshcheckreceiver
[statsdreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/receiver/statsdreceiver
[syslogreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/receiver/syslogreceiver
[tcplogreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/receiver/tcplogreceiver
[telegrafreceiver]: ./pkg/receiver/telegrafreceiver
[udplogreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/receiver/udplogreceiver
[vcenterreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/receiver/vcenterreceiver
[wavefrontreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/receiver/wavefrontreceiver
[windowseventlogreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/receiver/windowseventlogreceiver
[windowsperfcountersreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/receiver/windowsperfcountersreceiver
[zipkinreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/receiver/zipkinreceiver
[zookeeperreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/receiver/zookeeperreceiver

[attributesprocessor]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/processor/attributesprocessor
[batchprocessor]: https://github.com/open-telemetry/opentelemetry-collector/tree/v0.76.0/processor/batchprocessor
[cascadingfilterprocessor]: ./pkg/processor/cascadingfilterprocessor
[cumulativetodeltaprocessor]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/processor/cumulativetodeltaprocessor
[datadogprocessor]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/processor/datadogprocessor
[deltatorateprocessor]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/processor/deltatorateprocessor
[metricsgenerationprocessor]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/processor/metricsgenerationprocessor
[filterprocessor]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/processor/filterprocessor
[groupbyattrsprocessor]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/processor/groupbyattrsprocessor
[groupbytraceprocessor]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/processor/groupbytraceprocessor
[k8sprocessor]: ./pkg/processor/k8sprocessor
[k8sattributesprocessor]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/processor/k8sattributesprocessor
[logstransformprocessor]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/processor/logstransformprocessor
[memorylimiterprocessor]: https://github.com/open-telemetry/opentelemetry-collector/tree/v0.76.0/processor/memorylimiterprocessor
[metricfrequencyprocessor]: ./pkg/processor/metricfrequencyprocessor
[metricstransformprocessor]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/processor/metricstransformprocessor
[probabilisticsamplerprocessor]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/processor/probabilisticsamplerprocessor
[redactionprocessor]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/processor/redactionprocessor
[resourceprocessor]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/processor/resourceprocessor
[resourcedetectionprocessor]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/processor/resourcedetectionprocessor
[routingprocessor]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/processor/routingprocessor
[schemaprocessor]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/processor/schemaprocessor
[servicegraphprocessor]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/processor/servicegraphprocessor
[sourceprocessor]: ./pkg/processor/sourceprocessor
[spanprocessor]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/processor/spanprocessor
[spanmetricsprocessor]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/processor/spanmetricsprocessor
[sumologicschemaprocessor]: ./pkg/processor/sumologicschemaprocessor
[sumologicsyslogprocessor]: ./pkg/processor/sumologicsyslogprocessor
[tailsamplingprocessor]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/processor/tailsamplingprocessor
[transformprocessor]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/processor/transformprocessor

[carbonexporter]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/exporter/carbonexporter
[fileexporter]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/exporter/fileexporter
[kafkaexporter]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/exporter/kafkaexporter
[loadbalancingexporter]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/exporter/loadbalancingexporter
[loggingexporter]: https://github.com/open-telemetry/opentelemetry-collector/tree/v0.76.0/exporter/loggingexporter
[otlpexporter]: https://github.com/open-telemetry/opentelemetry-collector/tree/v0.76.0/exporter/otlpexporter
[otlphttpexporter]: https://github.com/open-telemetry/opentelemetry-collector/tree/v0.76.0/exporter/otlphttpexporter
[prometheusexporter]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/exporter/prometheusexporter
[sumologicexporter]: ./pkg/exporter/sumologicexporter
[syslogexporter]: ./pkg/exporter/syslogexporter

[asapauthextension]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/extension/asapauthextension
[awsproxy]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/extension/awsproxy
[basicauthextension]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/extension/basicauthextension
[bearertokenauthextension]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/extension/bearertokenauthextension
[dbstorage]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/extension/storage/dbstorage
[dockerobserver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/extension/observer/dockerobserver
[ecsobserver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/extension/observer/ecsobserver
[ecstaskobserver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/extension/observer/ecstaskobserver
[filestorage]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/extension/storage/filestorage
[headerssetterextension]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/extension/headerssetterextension
[healthcheckextension]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/extension/healthcheckextension
[hostobserver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/extension/observer/hostobserver
[httpforwarder]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/extension/httpforwarder
[jaegerremotesampling]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/extension/jaegerremotesampling
[k8sobserver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/extension/observer/k8sobserver
[ballastextension]: https://github.com/open-telemetry/opentelemetry-collector/tree/v0.76.0/extension/ballastextension
[oauth2clientauthextension]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/extension/oauth2clientauthextension
[oidcauthextension]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/extension/oidcauthextension
[pprofextension]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/extension/pprofextension
[sigv4authextension]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/extension/sigv4authextension
[sumologicextension]: ./pkg/extension/sumologicextension
[zpagesextension]: https://github.com/open-telemetry/opentelemetry-collector/tree/v0.76.0/extension/zpagesextension

[forwardconnector]: https://github.com/open-telemetry/opentelemetry-collector/tree/v0.76.0/connector/forwardconnector
[countconnector]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.76.3/connector/countconnector
