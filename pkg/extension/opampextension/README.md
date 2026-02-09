# OpAMP Agent Extension

**Stability level**: Alpha

This extension implements an [`OpAMP agent`][opamp_spec] for remote collector
configuration management. This extension needs to be used in conjuction with the
[`sumologicextension`][sumologicextension] in order to authenticate with the
[Sumo Logic][sumologic] OpAMP server.

It manages:

- authentication (using `sumologicextension` to retreive credentials)
- registration (sends an initial OpAMP agent-to-server message)
- reporting (responds to OpAMP server requests with an agent status, e.g. the
  collector's effective configuration and health status)
- health monitoring (tracks and reports component health status to the OpAMP server)
- local configuration (writes to a local OpenTelemetry YAML configuration file
  for a provider (i.e. glob) to read)
- collector configuration reloads (SIGHUP reloads on local configuration changes)

[opamp_spec]: https://github.com/open-telemetry/opamp-spec/blob/main/specification.md#opamp-open-agent-management-protocol
[sumologicextension]: https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/v0.127.0/extension/sumologicextension
[sumologic]: https://www.sumologic.com/

## Configuration

- `instance_uid`: a ULID formatted as a 26 character string in canonical
  representation. Auto-generated on start if missing. Setting this ensures the
  instance UID remains constant across process restarts.
- `endpoint`: (required) the OpAMP server secure websocket URL.
- `remote_configuration_directory`: (required) the directory used to store
  configuration received from the OpAMP server. This directory must coincide
  with a configuration provider (e.g. glob) for the configuration to be loaded
  by the collector.
- `disable_tag_replacement`: (optional) Boolean flag to disable new config merge
to disable tag replacement which replaces tags instead of appending to them for remotely managed collectors
  feature for remotely managed collectors. Default value - false
- `reports_health`: (optional) Boolean flag to enable/disable health reporting.
  When enabled, the agent tracks component health status and reports it to the
  OpAMP server. Default value - true

## Example Config

```yaml
extensions:
  sumologic:
    installation_token: <token>
    api_base_url: <api_endpoint_url>
  opamp:
    endpoint: <wss_endpoint_url>
    remote_configuration_directory: /etc/otelcol-sumo/opamp.d
    reports_health: true  # Optional: enable health reporting (default: true)
```

## API URLs

When integrating the extension with a different Sumo Logic deployment than the
default one (i.e. `https://open-collectors.sumologic.com`), one needs to specify
the Sumo Logic extension (`sumologic`) base API URL (`api_base_url`) and the
OpAMP extension (`opamp`) secure websocket endpoint (`endpoint`).

Here is a list of valid values for the Sumo Logic `api_base_url` configuration
option:

|  Deployment   | API base URL                                |
| :-----------: | ------------------------------------------- |
| default/`US1` | `https://open-collectors.sumologic.com`     |
|     `US2`     | `https://open-collectors.us2.sumologic.com` |
|     `AU`      | `https://open-collectors.au.sumologic.com`  |
|     `DE`      | `https://open-collectors.de.sumologic.com`  |
|     `EU`      | `https://open-collectors.eu.sumologic.com`  |
|     `JP`      | `https://open-collectors.jp.sumologic.com`  |
|     `CA`      | `https://open-collectors.ca.sumologic.com`  |
|     `KR`      | `https://open-collectors.kr.sumologic.com`  |

Here is a list of valid values for the OpAMP `endpoint** configuration option:

**Note:** These endpoints are not yet available.

|  Deployment   | API base URL                                 |
| :-----------: | -------------------------------------------- |
| default/`US1` | `https://opamp-collectors.sumologic.com`     |
|     `US2`     | `https://opamp-collectors.us2.sumologic.com` |
|     `AU`      | `https://opamp-collectors.au.sumologic.com`  |
|     `DE`      | `https://opamp-collectors.de.sumologic.com`  |
|     `EU`      | `https://opamp-collectors.eu.sumologic.com`  |
|     `JP`      | `https://opamp-collectors.jp.sumologic.com`  |
|     `CA`      | `https://opamp-collectors.ca.sumologic.com`  |
|     `KR`      | `https://opamp-collectors.kr.sumologic.com`  |

## Storing local configuration

When the OpAMP extension receives a remote configuration from the OpAMP server,
it persists each received YAML configuration to a local file in the
`remote_configuration_directory`. The existing contents of the
`remote_configuration_directory` are removed before doing so. A configuration
provider must be used in order to load the stored configuration, for example:
`--config "glob:/etc/otelcol-sumo/opamp.d/*"`.

## Components

This section lists the components that are included in the sumologic opamp extension for OpenTelemetry Collector.

|       Receivers        |      Processors      |   Exporters   |      Extensions      | Connectors |
| :--------------------: | :------------------: | :-----------: | :------------------: | :--------: |
|          nop           |      attributes      |  awskinesis   |       awsproxy       |            |
|         apache         |        batch         |     awss3     |     filestorage      |            |
|        filelog         |    memorylimiter     |    carbon     |     healthcheck      |            |
|      hostmetrics       |  resourcedetection   |     debug     |        opamp         |            |
|          otlp          |       resource       |     file      |        pprof         |            |
|    windowseventlog     |        filter        |     kafka     |      sumologic       |            |
|         nginx          |      transform       | loadbalancing |       asapauth       |            |
|         redis          |  cumulativetodelta   |      nop      |      basicauth       |            |
|         kafka          |     deltatorate      |     otlp      |   bearertokenauth    |            |
|      kafkametrics      |  metricsgeneration   |   sumologic   |      dbstorage       |            |
|      dockerstats       |     groupbyattrs     |    syslog     |    dockerobserver    |            |
|        rabbitmq        |     groupbytrace     |  prometheus   |    headerssetter     |            |
|  windowsperfcounters   |    k8sattributes     |   otlphttp    |     hostobserver     |            |
|         syslog         |       logdedup       |               |    httpforwarder     |            |
|         mysql          |    logstransform     |               | jaegerremotesampling |            |
|     elasticsearch      |   metricstransform   |               |     k8sobserver      |            |
|       postgresql       | probabilisticsampler |               |   oauth2clientauth   |            |
|     awscloudwatch      |        geoip         |               |       oidcauth       |            |
|  awscontainerinsight   |      redaction       |               |        pprof         |            |
| awsecscontainermetrics |      remotetap       |               |      sigv4auth       |            |
|      awsfirehose       |                      |               |        zpages        |            |
|        awsxray         |        schema        |               |  asapauthextension   |            |
|        collectd        |         span         |               |       ecstask        |            |
|        couchdb         |     tailsampling     |               |                      |            |
|        datadog         |        unroll        |               |                      |            |
|         expvar         |                      |               |                      |            |
|       filestats        |                      |               |                      |            |
|      flinkmetrics      |                      |               |                      |            |
|     fluentforward      |                      |               |                      |            |
|   googlecloudpubsub    |                      |               |                      |            |
|   googlecloudspanner   |                      |               |                      |            |
|        haproxy         |                      |               |                      |            |
|   activedirectoryds    |                      |               |                      |            |
|       aerospike        |                      |               |                      |            |
|     azureeventhub      |                      |               |                      |            |
|         bigip          |                      |               |                      |            |
|        carbonr         |                      |               |                      |            |
|         chrony         |                      |               |                      |            |
|       cloudflare       |                      |               |                      |            |
|      cloudfoundry      |                      |               |                      |            |
|       httpcheck        |                      |               |                      |            |
|          iis           |                      |               |                      |            |
|        influxdb        |                      |               |                      |            |
|         jaeger         |                      |               |                      |            |
|          jmx           |                      |               |                      |            |
|        journald        |                      |               |                      |            |
|       k8scluster       |                      |               |                      |            |
|       k8sevents        |                      |               |                      |            |
|       k8sobjects       |                      |               |                      |            |
|      kubeletstats      |                      |               |                      |            |
|          loki          |                      |               |                      |            |
|       memcached        |                      |               |                      |            |
|        mongodb         |                      |               |                      |            |
|      mongodbatlas      |                      |               |                      |            |
|          nsxt          |                      |               |                      |            |
|       opencensus       |                      |               |                      |            |
|        oracledb        |                      |               |                      |            |
|      otlpjsonfile      |                      |               |                      |            |
|         podman         |                      |               |                      |            |
|    simpleprometheus    |                      |               |                      |            |
|       prometheus       |                      |               |                      |            |
|         pulsar         |                      |               |                      |            |
|         purefa         |                      |               |                      |            |
|         purefb         |                      |               |                      |            |
|        receive         |                      |               |                      |            |
|          riak          |                      |               |                      |            |
|        saphana         |                      |               |                      |            |
|          sapm          |                      |               |                      |            |
|        signalfx        |                      |               |                      |            |
|       skywalking       |                      |               |                      |            |
|       snowflake        |                      |               |                      |            |
|          snmp          |                      |               |                      |            |
|         solace         |                      |               |                      |            |
|       splunkhec        |                      |               |                      |            |
|        sqlquery        |                      |               |                      |            |
|       sqlserver        |                      |               |                      |            |
|        sshcheck        |                      |               |                      |            |
|         statsd         |                      |               |                      |            |
|         tcplog         |                      |               |                      |            |
|         udplog         |                      |               |                      |            |
|       wavefront        |                      |               |                      |            |
|         zipkin         |                      |               |                      |            |
|       zookeeper        |                      |               |                      |            |
