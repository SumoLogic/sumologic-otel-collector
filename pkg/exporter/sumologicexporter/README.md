# Sumo Logic Exporter

This exporter supports sending logs and metrics data to [Sumo Logic](https://www.sumologic.com/).

We are strongly recommend to use this exporter with [sumologicextension](../../extension/sumologicextension/README.md).

Configuration is specified via the yaml in the following structure:

```yaml
exporters:
  # ...
  sumologic:
    # unique URL generated for your HTTP Source, this is the address to send data to
    # deprecated, please use sumologicextension to manage your endpoints
    # if sumologicextension is not being used, the endpoint is required
    endpoint: <HTTP_Source_URL>
    # Compression encoding format, empty string means no compression, default = gzip
    compress_encoding: {gzip, deflate, ""}
    # max HTTP request body size in bytes before compression (if applied),
    # default = 1_048_576 (1MB)
    max_request_body_size: <max_request_body_size>

    # format to use when sending logs to Sumo, default = otlp,
    # NOTE: only `otlp` is supported when used with sumologicextension
    log_format: {json, text, otlp}

    # format to use when sending metrics to Sumo, default = otlp,
    # NOTE: only `otlp` is supported when used with sumologicextension
    metric_format: {carbon2, graphite, otlp, prometheus}

    # format to use when sending traces to Sumo,
    # currently only otlp is supported
    trace_format: {otlp}

    # timeout is the timeout for every attempt to send data to the backend,
    # maximum connection timeout is 55s, default = 5s
    timeout: <timeout>

    # defines if timestamp for logs should be set to 0,
    # it indicates that backend will extract timestamp from logs,
    # this option affects OTLP format only
    # default = true
    clear_logs_timestamp: {true, false}

    # For below described source and graphite template related configuration,
    # please refer to "Source templates" documentation chapter from this document.

    # desired source category, useful if you want to override the source category
    # configured for the source.
    source_category: <source_category>
    # desired source name, useful if you want to override the source name
    # configured for the source.
    source_name: <source_name>
    # desired host name, useful if you want to override the source host
    # configured for the source.
    source_host: <source_host>
    # template for Graphite format, applied only if metric_format is set to graphite;
    # source templating is going to be applied,
    # default = `%{_metric_}`
    graphite_template: <graphite_template>

    # translate_attributes specifies whether attributes should be translated
    # from OpenTelemetry to Sumo conventions;
    # see "Attribute translation" documentation chapter from this document,
    # default = true
    translate_attributes: {true, false}

    # Specifies whether telegraf metric names should be translated to match
    # Sumo conventions expected in Sumo host related apps (for example
    # `procstat_num_threads` => `Proc_Threads` or `cpu_usage_irq` => `CPU_Irq`).
    # See `translate_metrics.go` for full list of translations.
    # default = true
    translate_telegraf_attributes: {true, false}

    # list of regexes for attributes which should be sent as metadata,
    # use OpenTelemetry attribute names, see "Attribute translation" documentation
    # chapter from this document.
    metadata_attributes:
      - <regex1>
      - <regex2>

    # instructs sumologicextension to use automatically generated sumologic endpoint;
    # to use direct endpoint, set it to null
    # see sumologicextension documentation for details
    # default = sumologic
    auth:
      authenticator: {sumologic, null}

    # for below described queueing and retry related configuration please refer to:
    # https://github.com/open-telemetry/opentelemetry-collector/blob/main/exporter/exporterhelper/README.md#configuration

    retry_on_failure:
      # default = true
      enabled: {true, false}
      # time to wait after the first failure before retrying;
      # ignored if enabled is false, default = 5s
      initial_interval: <initial_interval>
      # is the upper bound on backoff; ignored if enabled is false, default = 30s
      max_interval: <max_interval>
      # is the maximum amount of time spent trying to send a batch;
      # ignored if enabled is false, default = 120s
      max_elapsed_time: <max_elapsed_time>

    sending_queue:
      # default = false
      enabled: {true, false}
      # number of consumers that dequeue batches; ignored if enabled is false,
      # default = 10
      num_consumers: <num_consumers>
      # maximum number of batches kept in memory before data;
      # ignored if enabled is false, default = 5000
      #
      # user should calculate this as num_seconds * requests_per_second where:
      # num_seconds is the number of seconds to buffer in case of a backend outage,
      # requests_per_second is the average number of requests per seconds.
      queue_size: <queue_size>
```

[sumologicextension]: ./../../extension/sumologicextension

## Attribute translation

Attribute translation changes some of the attribute keys from OpenTelemetry convention to Sumo convention.
For example, OpenTelemetry convention for the attribute containing Kubernetes pod name is `k8s.pod.name`,
but Sumo expects it to be in attribute named `pod`.

If attribute with target name eg. `pod` already exists,
translation is not being done for corresponding attribute (`k8s.pod.name` in this example).

This feature is turned on by default.
To turn it off, set the `translate_attributes` configuration option to `false`.
Note that this may cause some of Sumo apps, built-in dashboards to not work correctly.

Below is a list of all attribute keys that are being translated.

| OTC key name              | Sumo key name      |
|---------------------------|--------------------|
| `cloud.account.id`        | `AccountId`        |
| `cloud.availability_zone` | `AvailabilityZone` |
| `cloud.platform`          | `aws_service`      |
| `cloud.region`            | `Region`           |
| `host.id`                 | `InstanceId`       |
| `host.name`               | `host`             |
| `host.type`               | `InstanceType`     |
| `k8s.cluster.name`        | `Cluster`          |
| `k8s.container.name`      | `container`        |
| `k8s.daemonset.name`      | `daemonset`        |
| `k8s.deployment.name`     | `Deployment`       |
| `k8s.namespace.name`      | `Namespace`        |
| `k8s.node.name`           | `node`             |
| `k8s.pod.hostname`        | `host`             |
| `k8s.pod.name`            | `pod`              |
| `k8s.pod.uid`             | `pod_id`           |
| `k8s.replicaset.name`     | `replicaset`       |
| `k8s.statefulset.name`    | `statefulset`      |
| `service.name`            | `service`          |
| `file.path.resolved`      | `_sourceName`      |

## Source Templates

You can specify a template with an attribute for `source_category`, `source_name`,
`source_host` or `graphite_template` using `%{attr_name}`.

For example, when there is an attribute `my_attr`: `my_value`, `metrics/%{my_attr}`
would be expanded to `metrics/my_value`.
Use OpenTelemetry attribute names, even when [attribute translation](#attribute-translation)
is turned on.

For `graphite_template`, in addition to above, `%{_metric_}` is going to be replaced
with metric name.

If an attribute is not found, it is replaced with `undefined`.
For example, `%{existing_attr}/%{nonexistent_attr}` becomes `value-of-existing-attr/undefined`.

## Example Configuration

### Example with sumologicextension

```yaml
extensions:
  sumologic:
    access_id: aaa
    access_key: bbbbbbbbbbbbbbbbbbbbbb
    collector_name: my_collector

receivers:
  hostmetrics:
    collection_interval: 30s
    scrapers:
      load:

exporters:
  sumologic:
    source_category: "custom category"
    source_name: "custom name"
    source_host: "custom host"
    metadata_attributes:
      - k8s.*

service:
  extensions: [sumologic]
  pipelines:
    metrics:
      receivers: [hostmetrics]
      exporters: [sumologic]
```

### Example without sumologicextension

```yaml
exporters:
  sumologic:
    endpoint: http://localhost:3000
    compress_encoding: "gzip"
    max_request_body_size: "1_048_576"  # 1MB
    log_format: "text"
    metric_format: "prometheus"
    source_category: "custom category"
    source_name: "custom name"
    source_host: "custom host"
    metadata_attributes:
      - k8s.*
```
