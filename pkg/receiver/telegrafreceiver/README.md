# Telegraf Receiver

**Stability level**: Beta

> **Deprecated**
> The `telegrafreceiver` is deprecated and will be removed in a future release. Please migrate to direct otel receivers instead of telegraf plugins. If you still want to use telegraf plugin, you can setup a standalone telegraf and forward metrics to sumo otel collector using otlp/http output plugin in telegraf.
>
> Example standalone Telegraf configuration with OTLP output:
>
> ```toml
> [agent]
>   interval = "30s"
>   flush_interval = "30s"
>
> [[inputs.cpu]]
>   percpu = true
>   totalcpu = true
>
> [[inputs.mem]]
>
> [[outputs.opentelemetry]]
>   service_address = "localhost:4317"
> ```
>
> Then configure the OTel Collector with an `otlp` receiver:
>
> ```yaml
> receivers:
>   otlp:
>     protocols:
>       grpc:
>         endpoint: 0.0.0.0:4317
>
> service:
>   pipelines:
>     metrics:
>       receivers: [otlp]
>       processors: [batch]
>       exporters: [sumologic]
> ```
>
> ## Migrating from Telegraf `http_listener_v2` to `prometheus_remote_write` receiver
>
> If you were using the Telegraf receiver with the `http_listener_v2` input plugin
> to accept Prometheus remote write metrics, you can migrate to the native
> `prometheus_remote_write` receiver.
>
> **Before (telegrafreceiver with http_listener_v2):**
>
> ```yaml
> receivers:
>   telegraf:
>     agent_config: |
>       [agent]
>         interval = "30s"
>         flush_interval = "30s"
>       [[inputs.http_listener_v2]]
>         service_address = ":9090"
>         paths = ["/api/v1/write"]
>         data_format = "prometheusremotewrite"
>
> service:
>   pipelines:
>     metrics:
>       receivers: [telegraf]
>       processors: [batch]
>       exporters: [sumologic]
> ```
>
> **After (prometheus_remote_write receiver):**
>
> ```yaml
> receivers:
>   prometheus_remote_write:
>     endpoint: 0.0.0.0:9090
>
> service:
>   pipelines:
>     metrics:
>       receivers: [prometheus_remote_write]
>       processors: [batch]
>       exporters: [sumologic]
> ```
>
> The `prometheus_remote_write` receiver listens on `/api/v1/write` by default.
> Configure Prometheus to send to the collector:
>
> ```yaml
> remote_write:
>   - url: "http://<collector-host>:9090/api/v1/write"
> ```
>
> **Note:** The `prometheus_remote_write` receiver is currently in **Alpha** stability and only supports metrics.
> Refer to the [upstream documentation][prometheusremotewritereceiver] for full configuration options and known limitations.
>
> [prometheusremotewritereceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.150.0/receiver/prometheusremotewritereceiver

Telegraf receiver for ingesting metrics from various [input plugins][input_plugins]
into otc pipeline.

Supported pipeline types: metrics

Use case: user configures telegraf input plugins in config for ingestion and otc
processors and exporters for data processing and export.

[input_plugins]: https://github.com/SumoLogic/telegraf/tree/v1.22.0-sumo-4/plugins/inputs

## Configuration

The following settings are required:

- `agent_config`: Telegraf config. For now it allows to provide agent and input
  plugins configuration. One can refer to
  [telegraf configuration docs][telegraf_config_docs] for full list of
  configuration options.

The Following settings are optional:

- `separate_field` (default value is `false`): Specify whether metric field
  should be added separately as data point label.
- `consume_retry_delay` (default value is `500ms`): The retry delay for recoverable
  errors from the rest of the pipeline. Don't change this or the related setting below
  unless you know what you're doing.
- `consume_max_retries` (default value is `10`): The maximum number of retries for recoverable
  errors from the rest of the pipeline.

Example:

```yaml
receivers:
  telegraf:
    separate_field: false
    agent_config: |
      [agent]
        interval = "2s"
        flush_interval = "3s"
      [[inputs.mem]]
```

The full list of settings exposed for this receiver are documented in
[config.go](./config.go).

[telegraf_config_docs]: https://github.com/SumoLogic/telegraf/blob/v1.22.0-sumo-4/docs/CONFIGURATION.md

## Limitations

With its current implementation Telegraf receiver has the following limitations:

- only input plugins can be configured in telegraf agent confugration section
  (apart from agent's configuration itself). That means that metrics go straight
  from input plugin to the receiver for translation (into otc data model) without
  any processing
- only the following Telegraf metric data types are supported:
  - `telegraf.Gauge` that is translated to `pmetric.MetricDataTypeGauge`,
  - `telegraf.Counter` that is translated to `pmetric.MetricDataTypeSum`.

## Migration from Telegraf

### Data model

Internal OTC metric format differs from the Telegraf one and `separate_field` controls the conversion:

- If `separate_field` is `false`, the Open Telemetry metric name is going to be concatenated from the Telegraf metric name
  and the Telegraf field with `_` as separator.

  The following telegraf structure:

  ```json
  {
    "fields": {
      "HeapMemoryUsage.committed": 1007157248
      "HeapMemoryUsage.init": 1007157248
    },
    "name": "tomcat_jmx_jvm_memory",
    "tags": {
      "component": "webserver",
      "environment": "dev",
      "host": "32fafdb10522",
      "jolokia_agent_url": "http://tomcat:8080/jolokia",
      "webserver_system": "tomcat"
    },
    "timestamp": 1646904912
  }
  ```

  is going to be converted to the following OpenTelemetry structure:

  ```console
  2022-03-10T07:16:34.117Z  DEBUG loggingexporter/logging_exporter.go:64
  ResourceMetrics #0
  Resource SchemaURL:
  Resource labels:
      -> component: STRING(webserver)
      -> environment: STRING(dev)
      -> host: STRING(32fafdb10522)
      -> jolokia_agent_url: STRING(http://tomcat:8080/jolokia)
      -> webserver_system: STRING(tomcat)
  InstrumentationLibraryMetrics #0
  InstrumentationLibraryMetrics SchemaURL:
  InstrumentationLibrary telegraf v0.1
  Metric #0
  Descriptor:
      -> Name: tomcat_jmx_jvm_memory_HeapMemoryUsage.committed
      -> Description:
      -> Unit:
      -> DataType: Gauge
  NumberDataPoints #0
  StartTimestamp: 1970-01-01 00:00:00 +0000 UTC
  Timestamp: 2022-03-10 09:35:12 +0000 UTC
  Value: 1007157248.000000
  Metric #1
  Descriptor:
      -> Name: tomcat_jmx_jvm_memory_HeapMemoryUsage.init
      -> Description:
      -> Unit:
      -> DataType: Gauge
  NumberDataPoints #0
  StartTimestamp: 1970-01-01 00:00:00 +0000 UTC
  Timestamp: 2022-03-10 09:35:12 +0000 UTC
  Value: 1007157248.000000
  ```

- If `separate_fields` is `true`, the Open Telemetry metric name is going to be the same as the Telegraf one,
  and the Telegraf `field` is going to be converted to the Open Telemetry data point attribute.

  The following telegraf structure:

  ```json
  {
    "fields": {
      "HeapMemoryUsage.committed": 1007157248
      "HeapMemoryUsage.init": 1007157248
    },
    "name": "tomcat_jmx_jvm_memory",
    "tags": {
      "component": "webserver",
      "environment": "dev",
      "host": "32fafdb10522",
      "jolokia_agent_url": "http://tomcat:8080/jolokia",
      "webserver_system": "tomcat"
    },
    "timestamp": 1646904912
  }
  ```

  is going to be converted to the following OpenTelemetry structure:

  ```console
  2022-03-10T11:28:30.333Z  DEBUG loggingexporter/logging_exporter.go:64
  ResourceMetrics #0
  Resource SchemaURL:
  Resource labels:
      -> component: STRING(webserver)
      -> environment: STRING(dev)
      -> host: STRING(32fafdb10522)
      -> jolokia_agent_url: STRING(http://tomcat:8080/jolokia)
      -> webserver_system: STRING(tomcat)
  InstrumentationLibraryMetrics #0
  InstrumentationLibraryMetrics SchemaURL:
  InstrumentationLibrary
  Metric #0
  Descriptor:
      -> Name: tomcat_jmx_jvm_memory
      -> Description:
      -> Unit:
      -> DataType: Gauge
  NumberDataPoints #0
  Data point attributes:
      -> field: STRING(HeapMemoryUsage.committed)
  StartTimestamp: 1970-01-01 00:00:00 +0000 UTC
  Timestamp: 2022-03-10 09:35:12 +0000 UTC
  Value: 1007157248.000000
  Metric #1
  Descriptor:
      -> Name: tomcat_jmx_jvm_memory
      -> Description:
      -> Unit:
      -> DataType: Gauge
  NumberDataPoints #0
  Data point attributes:
      -> field: STRING(HeapMemoryUsage.init)
  StartTimestamp: 1970-01-01 00:00:00 +0000 UTC
  Timestamp: 2022-03-10 09:35:12 +0000 UTC
  Value: 1007157248.000000
  ```

  </details>

### Keep compatibility while sending metrics to Sumo Logic

In Telegraf, metrics can be sent to Sumo Logic using [Sumologic Output Plugin][sumologic_output_plugin].
It supports three formats (`prometheus`, `carbon2`, `graphite`),
where each of them has some limitations (e.g. [only specific set of chars can be used for metric name for prometheus][prometheus_data_model]).

OTLP doesn't have most of those limitations, so in order to keep the same metric names, some transformations should be done by processors.

Let's consider the following example.
`metric.with.dots` is going to be sent as `metric_with_dots` by prometheus or `metric.with.dots` by OTLP.
To unify it, you can use the [Metrics Transform Processor][metricstransformprocessor]:

```yaml
processors:
  metricstransform:
    transforms:
      ## Replace metric.with.dots metric to metric_with_dots
      - include: metric.with.dots
        match_type: strict
        action: update
        new_name: metric_with_dots
# ...
service:
  pipelines:
    metrics:
      receivers:
        - telegraf
      processors:
        - metricstransform
      exporters:
        - sumologic
# ...
```

With [Metrics Transform Processor][metricstransformprocessor] and regular expressions you can also handle more complex scenarios,
like in the following snippet:

```yaml
processors:
  metricstransform:
    transforms:
      ## Change <part1>.<part2> metrics to to <part1>_<part2>
      - include: ^([^\.]*)\.([^\.]*)$$
        match_type: strict
        action: update
        new_name: $${1}.$${2}
      ## Change <part1>.<part2>.<part3> metrics to to <part1>_<part2>_<part3>
      - include: ^([^\.]*)\.([^\.]*)\.([^\.]*)$$
        match_type: strict
        action: update
        new_name: $${1}.$${2}.${3}
# ...
service:
  pipelines:
    metrics:
      receivers:
        - telegraf
      processors:
        - metricstransform
      exporters:
        - sumologic
# ...
```

[prometheus_data_model]: https://prometheus.io/docs/concepts/data_model/#metric-names-and-labels
[sumologic_output_plugin]: https://github.com/influxdata/telegraf/tree/master/plugins/outputs/sumologic
[metricstransformprocessor]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.150.0/processor/metricstransformprocessor
