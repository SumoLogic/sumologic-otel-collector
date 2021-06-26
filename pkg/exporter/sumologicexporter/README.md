# Sumo Logic Exporter

This exporter supports sending logs and metrics data to [Sumo Logic](https://www.sumologic.com/).

The following configuration options are supported:

- `endpoint` (required, unless configured with [`sumologicextension`][sumologicextension]): Unique URL generated for your HTTP Source.
  This is the address to send metrics to.
- `compress_encoding` (optional): Compression encoding format. Empty string means no compression
  - default: `gzip`
  - allowed values: `gzip`, `deflate`, `""`
- `max_request_body_size` (optional): Max HTTP request body size in bytes before compression (if applied).
  - default: `1_048_576` (1MB)
- `translate_metadata` (optional): Specifies whether metadata attributes should be translated
  from OpenTelemetry to Sumo conventions.
  See [Metadata translation](#metadata-translation).
  Default is `true`.
- `metadata_attributes` (optional): List of regexes for attributes which should be send as metadata
- `metadata_attributes` (optional): List of regexes for attributes which should be sent as metadata.
  Use OpenTelemetry attribute names (see [Metadata translation](#metadata-translation)).
- `log_format` (optional) (logs only): Format to use when sending logs to Sumo.
  - default: `json`
  - allowed values: `json`, `text`, `otlp`
  - **NOTE**: only `otlp` is supported when used with [`sumologicextension`][sumologicextension].
- `metric_format` (optional) (metrics only): Format of the metrics to be sent
  - default: `prometheus`
  - allowed values: `carbon2`, `graphite`, `otlp`, `prometheus`.
  - **NOTE**: only `otlp` is supported when used with [`sumologicextension`][sumologicextension].
- `graphite_template` (default=`%{_metric_}`) (optional) (metrics only): Template for Graphite format.
  [Source templates](#source-templates) are going to be applied.
  Applied only if `metric_format` is set to `graphite`.
- `source_category` (optional): Desired source category. Useful if you want to override the source category configured for the source.
  [Source templates](#source-templates) are going to be applied.
- `source_name` (optional): Desired source name. Useful if you want to override the source name configured for the source.
  [Source templates](#source-templates) are going to be applied.
- `source_host` (optional): Desired host name. Useful if you want to override the source host configured for the source.
  [Source templates](#source-templates) are going to be applied.
- `timeout`: Is the timeout for every attempt to send data to the backend.
  Maximum connection timeout is 55s.
  - default: `5s`
- `retry_on_failure`
  - `enabled` (default = `true`)
  - `initial_interval` (default = `5s`): Time to wait after the first failure before retrying; ignored if `enabled` is `false`
  - `max_interval` (default = `30s`): Is the upper bound on backoff; ignored if `enabled` is `false`
  - `max_elapsed_time` (default = `120s`): Is the maximum amount of time spent trying to send a batch; ignored if `enabled` is `false`
- `sending_queue`
  - `enabled` (default = `false`)
  - `num_consumers` (default = `10`): Number of consumers that dequeue batches; ignored if `enabled` is `false`
  - `queue_size` (default = `5000`): Maximum number of batches kept in memory before data; ignored if `enabled` is `false`;
  User should calculate this as `num_seconds * requests_per_second` where:
    - `num_seconds` is the number of seconds to buffer in case of a backend outage
    - `requests_per_second` is the average number of requests per seconds.

[sumologicextension]: ./../../extension/sumologicextension

## Metadata translation

Metadata translation changes some of the attribute keys from OpenTelemetry convention to Sumo convention.
For example, OpenTelemetry convention for the attribute containing Kubernetes pod name is `k8s.pod.name`,
but Sumo expects it to be in attribute named `pod`.

This feature is turned on by default.
To turn it off, set the `translate_metadata` configuration option to `false`.
Note that this may cause some of Sumo apps, built-in dashboards to not work correctly.

Below is a list of all metadata keys that are being translated.

| OTC key name            | Sumo key name    |
|-------------------------|------------------|
| cloud.account.id        | accountId        |
| cloud.availability_zone | availabilityZone |
| cloud.platform          | aws_service      |
| cloud.region            | region           |
| host.id                 | instanceId       |
| host.name               | host             |
| host.type               | instanceType     |
| k8s.cluster.name        | cluster          |
| k8s.container.name      | container        |
| k8s.daemonset.name      | daemonset        |
| k8s.deployment.name     | deployment       |
| k8s.namespace.name      | namespace        |
| k8s.node.name           | node             |
| k8s.pod.hostname        | host             |
| k8s.pod.name            | pod              |
| k8s.pod.uid             | pod_id           |
| k8s.replicaset.name     | replicaset       |
| k8s.statefulset.name    | statefulset      |
| service.name            | service          |

## Source Templates

You can specify a template with an attribute for `source_category`, `source_name`, `source_host` or `graphite_template` using `%{attr_name}`.

For example, when there is an attribute `my_attr`: `my_value`, `metrics/%{my_attr}` would be expanded to `metrics/my_value`.
Use OpenTelemetry attribute names, even when [metadata translation](#metadata-translation) is turned on.

For `graphite_template`, in addition to above, `%{_metric_}` is going to be replaced with metric name.

## Example Configuration

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
