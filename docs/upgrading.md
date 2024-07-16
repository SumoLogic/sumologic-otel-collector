# Upgrading

- [Upgrading to v0.104.0-sumo-0](#upgrading-to-v01040-sumo-0)
  - [`sumologic` exporter: remove `compress_encoding`](#sumologic-exporter-remove-compress_encoding)
- [Upgrading to v0.103.0-sumo-0](#upgrading-to-v01030-sumo-0)
  - [`sumologic` configuration: modified the `configuration files` merge behaviour](#sumologic-configuration-modified-the-configuration-files-merge-behaviour)
- [Upgrading to v0.96.0-sumo-0](#upgrading-to-v0960-sumo-0)
  - [`sumologic` exporter: remove `json_logs`](#sumologic-exporter-remove-json_logs)
  - [`sumologic` exporter: remove `clear_logs_timestamp`](#sumologic-exporter-remove-clear_logs_timestamp)
- [Upgrading to v0.94.0-sumo-0](#upgrading-to-v0940-sumo-0)
  - [`servicegraph` processor: removed in favor of `servicegraph` connector](#servicegraph-processor-removed-in-favor-of-servicegraph-connector)
- [Upgrading to v0.92.0-sumo-0](#upgrading-to-v0920-sumo-0)
  - [Exporters: changed retry logic when using persistent queue](#exporters-changed-retry-logic-when-using-persistent-queue)
- [Upgrading to v0.91.0-sumo-0](#upgrading-to-v0910-sumo-0)
  - [Sumo Logic Schema processor replaced with Sumo Logic processor](#sumo-logic-schema-processor-replaced-with-sumo-logic-processor)
  - [`k8s_tagger` processor: default name of podID attribute has changed](#k8s_tagger-processor-default-name-of-podid-attribute-has-changed)
- [Upgrading to v0.90.1-sumo-0](#upgrading-to-v0901-sumo-0)
  - [Change configuration for `syslogexporter`](#change-configuration-for-syslogexporter)
  - [`sumologic` exporter: deprecate `clear_logs_timestamp`](#sumologic-exporter-deprecate-clear_logs_timestamp)
  - [`sumologic` exporter: remove `routing_attributes_to_drop`](#sumologic-exporter-remove-routing_attributes_to_drop)
  - [`sumologic` exporter: deprecate `json_logs`](#sumologic-exporter-deprecate-json_logs)
    - [Migration example for `add_timestamp` and `timestamp_key`](#migration-example-for-add_timestamp-and-timestamp_key)
    - [Migration example for `flatten_body`](#migration-example-for-flatten_body)
    - [Migration example for `log_key`](#migration-example-for-log_key)
- [Upgrading to v0.89.0-sumo-0](#upgrading-to-v0890-sumo-0)
  - [`remoteobserver` processor: renamed to `remotetap` processor](#remoteobserver-processor-renamed-to-remotetap-processor)
  - [`sumologic` exporter: changed default `timeout` from `5s` to `30s`](#sumologic-exporter-changed-default-timeout-from-5s-to-30s)
  - [`sumologic` extension: changed default `discover_collector_tags` from `false` to `true`](#sumologic-extension-changed-default-discover_collector_tags-from-false-to-true)
- [Upgrading to v0.84.0-sumo-0](#upgrading-to-v0840-sumo-0)
  - [`sumologic` extension: removed `install_token` in favor of `installation_token`](#sumologic-extension-removed-install_token-in-favor-of-installation_token)
- [Upgrading to v0.77.0-sumo-0](#upgrading-to-v0770-sumo-0)
  - [Full Prometheus metric name normalization is now disabled by default](#full-prometheus-metric-name-normalization-is-now-disabled-by-default)
- [Upgrading to v0.73.0-sumo-1](#upgrading-to-v0730-sumo-1)
  - [The default collector name for sumologic extension is now the host FQDN](#the-default-collector-name-for-sumologic-extension-is-now-the-host-fqdn)
- [Upgrading to v0.66.0-sumo-0](#upgrading-to-v0660-sumo-0)
  - [`filelog` receiver: has been removed from sub-parsers](#filelog-receiver-has-been-removed-from-sub-parsers)
  - [`sending_queue`: require explicit storage set](#sending_queue-require-explicit-storage-set)
  - [`apache` receiver: turn on feature gates for resource attributes](#apache-receiver-turn-on-feature-gates-for-resource-attributes)
  - [`elasticsearch` receiver: turn on more datapoints](#elasticsearch-receiver-turn-on-more-datapoints)
- [Upgrading to v0.57.2-sumo-0](#upgrading-to-v0572-sumo-0)
  - [`sumologic` exporter: drop support for source headers](#sumologic-exporter-drop-support-for-source-headers)
- [Upgrading to v0.56.0-sumo-0](#upgrading-to-v0560-sumo-0)
  - [`sumologic` exporter: drop support for translating attributes](#sumologic-exporter-drop-support-for-translating-attributes)
  - [`sumologic` exporter: drop support for translating Telegraf metric names](#sumologic-exporter-drop-support-for-translating-telegraf-metric-names)
- [Upgrading to v0.55.0-sumo-0](#upgrading-to-v0550-sumo-0)
  - [`filter` processor: drop support for `expr` language](#filter-processor-drop-support-for-expr-language)
- [Upgrading to v0.52.0-sumo-0](#upgrading-to-v0520-sumo-0)
  - [`sumologic` exporter: Removed `carbon2` and `graphite` metric formats](#sumologic-exporter-removed-carbon2-and-graphite-metric-formats)
- [Upgrading to v0.51.0-sumo-0](#upgrading-to-v0510-sumo-0)
  - [`k8s_tagger` processor: removed `clusterName` metadata extraction option](#k8s_tagger-processor-removed-clustername-metadata-extraction-option)
    - [How to upgrade](#how-to-upgrade)
  - [`sumologic` exporter: metadata translation: changed the attribute that is translated to `_sourceName` from `file.path.resolved` to `log.file.path_resolved`](#sumologic-exporter-metadata-translation-changed-the-attribute-that-is-translated-to-_sourcename-from-filepathresolved-to-logfilepath_resolved)
    - [How to upgrade](#how-to-upgrade-1)
- [Upgrading to 0.49.0-sumo-0](#upgrading-to-0490-sumo-0)
  - [Several changes to receivers using opentelemetry-log-collection](#several-changes-to-receivers-using-opentelemetry-log-collection)
  - [Sumo Logic exporter metadata handling](#sumo-logic-exporter-metadata-handling)
    - [Removing unnecessary metadata using the resourceprocessor](#removing-unnecessary-metadata-using-the-resourceprocessor)
    - [Moving record-level attributes used for metadata to the resource level](#moving-record-level-attributes-used-for-metadata-to-the-resource-level)

## Upgrading to v0.104.0-sumo-0

### `sumologic` exporter: remove `compress_encoding`

`compress_encoding` has been removed in favor of using `compression` from client config.

To migrate do the following step:

- use `compression` in replace of `compress_encoding`:

Change this:

  ```yaml
  exporters:
    sumologic:
      compress_encoding: ""
  ```

to this:

  ```yaml
  exporters:
    sumologic:
      compression: ""
  ```

## Upgrading to v0.103.0-sumo-0

### `sumologic` configuration: modified the `configuration files` merge behaviour

Modified the configuration merge behaviour to perform overwrite instead of update for `collector_fields` key of [Sumo Logic Extension](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/extension/sumologicextension#configuration).

For example, if two configuration files(say `conf1.yaml` and `conf2.yaml`) define the `collector_fields`,

In previous flow, the values of `collector_fields` from `conf2.yaml` will be added to `conf1.yaml` and the effective configuration will have vaules from both the configurations.

Starting from `v0.103.0-sumo-0`, the values of `collector_fields` tag from `conf1.yaml` will be overwritten by values from `conf2.yaml` and the effective configuration will have `collector_fields` value of `conf2.yaml` only.

For example:

`conf1.yaml`:

```
extensions:
  sumologic:
    collector_description: "My OpenTelemetry Collector"
    collector_fields:
      cluster: "cluster-1"
    some_list:
      - element 1
      - element 2
```

`conf2.yaml`:

```
extensions:
  sumologic:
    collector_fields:
      zone: "eu"
    some_list:
      - element 3
      - element 4
```

effective configuration (`old behaviour`)

```
extensions:
  sumologic:
    collector_description: "My OpenTelemetry Collector"
    collector_fields:
      cluster: "cluster-1"
      zone: "eu"
    some_list:
      - element 3
      - element 4
```

effective configuration (`new behaviour`)

```
extensions:
  sumologic:
    collector_description: "My OpenTelemetry Collector"
    collector_fields:
      zone: "eu"
    some_list:
      - element 3
      - element 4
```

If you have multiple config files with `collector_fields` key specified, only the value from the last file(**alphabetically** sorted order) will be present in effective configuration.
Due to above, avoid maintaining `collector_fields` in multiple configuration files and move them to a single file.

Note: This applies only to `collector_fields` key, all other key behaviour will remain the same.

For more details regarding configuration structure and merge behaviour, see https://help.sumologic.com/docs/send-data/opentelemetry-collector/data-source-configurations/overview/#configuration-structure

## Upgrading to v0.96.0-sumo-0

### `sumologic` exporter: remove `json_logs`

`json_logs` has been removed in favor of `transform` processor.

Please follow [the migration process](#sumologic-exporter-deprecate-json_logs)

### `sumologic` exporter: remove `clear_logs_timestamp`

`clear_logs_timestamp` has been removed in favor of `transform` processor.

Please follow [the migration process](#sumologic-exporter-deprecate-clear_logs_timestamp)

## Upgrading to v0.94.0-sumo-0

### `servicegraph` processor: removed in favor of `servicegraph` connector

The deprecated [Service Graph processor][servicegraphprocessor] has been removed.
Use the [Service Graph connector][servicegraphconnector] instead.

[servicegraphprocessor]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.92.0/processor/servicegraphprocessor
[servicegraphconnector]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.94.0/connector/servicegraphconnector

## Upgrading to v0.92.0-sumo-0

### Exporters: changed retry logic when using persistent queue

In previous versions, when an exporter (e.g. Sumo Logic exporter) was configured to use retries and persistent queue, the data would be retried indefinitely as long as the queue wasn't full.
This was because after reaching the retry limit configured in `retry_on_failure.max_elapsed_time`, the data would be put back in the sending queue, if the queue wasn't full.
Starting in `v0.92.0-sumo-0`, this behavior is changed. Now the data is only retried for `retry_on_failure.max_elapsed_time`
(which currently defaults to five minutes) and dropped after that.
To prevent the exporter form ever dropping data that was successfully queued, set `retry_on_failure.max_elapsed_time` to `0`.

For example, change this:

```yaml
exporters:
  sumologic:
    endpoint: ...
    retry_on_failure:
      enabled: true
    sending_queue:
      enabled: true
      storage: file_storage
```

to this:

```yaml
exporters:
  sumologic:
    endpoint: ...
    retry_on_failure:
      enabled: true
      max_elapsed_time: 0
    sending_queue:
      enabled: true
      storage: file_storage
```

See the change for details: https://github.com/open-telemetry/opentelemetry-collector/pull/9090.

## Upgrading to v0.91.0-sumo-0

### Sumo Logic Schema processor replaced with Sumo Logic processor

The [Sumo Logic Schema processor][sumologicschema] has been deprecated in favor of the [Sumo Logic processor][sumologicprocessor].
To ensure you are using the latest version change `sumologic_schema` to `sumologic` in your configuration.

For example, change this:

```yaml
processors:
  sumologic_schema:
    # ...
service:
  pipelines:
    logs:
      processors:
        - sumologic_schema
```

to this:

```yaml
processors:
  sumologic:
    # ...
service:
  pipelines:
    logs:
      processors:
        - sumologic
```

[sumologicschema]: https://github.com/SumoLogic/sumologic-otel-collector/tree/main/pkg/processor/sumologicschemaprocessor
[sumologicprocessor]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/sumologicprocessor

### `k8s_tagger` processor: default name of podID attribute has changed

By a mistake, in the [k8s_tagger][k8staggerprocessor], the default name for podID was set to `k8s.pod.id`. It has been changed to `k8s.pod.uid`.
If you want to still use the old name, add the following option to the config of the `k8s_tagger`:

```yaml
processors:
  k8s_tagger:
    tags:
      podID: k8s.pod.id
```

[k8staggerprocessor]: https://github.com/SumoLogic/sumologic-otel-collector/tree/main/pkg/processor/k8sprocessor

## Upgrading to v0.90.1-sumo-0

### Change configuration for `syslogexporter`

To migrate, rename the following keys in configuration for `syslogexporter`:

- rename `protocol` property to `network`
- rename `format` property to `protocol`

For example, given the following configuration:

```yaml
  syslog:
    protocol: tcp
    port: 514
    endpoint: 127.0.0.1
    format: rfc5424
    tls:
      ca_file: ca.pem
      cert_file: cert.pem
      key_file: key.pem
```

change it to:

```yaml
  syslog:
    network: tcp
    port: 514
    endpoint: 127.0.0.1
    protocol: rfc5424
    tls:
      ca_file: ca.pem
      cert_file: cert.pem
      key_file:  key.pem
```

### `sumologic` exporter: deprecate `clear_logs_timestamp`

`clear_logs_timestamp` has been deprecated in favor of `transform` processor. It is going to be removed in `v0.95.0-sumo-0`. To migrate:

- set `clear_logs_timestamp` to `false`
- add the following processor:

  ```yaml
  processors:
    transform/clear_logs_timestamp:
      log_statements:
        - context: log
          statements:
            - set(time_unix_nano, 0)
  ```

For example, given the following configuration:

```yaml
exporters:
  sumologic:
```

change it to:

```yaml
exporters:
  sumologic:
    clear_logs_timestamp: false
processors:
  transform/clear_logs_timestamp:
    log_statements:
      - context: log
        statements:
          - set(time_unix_nano, 0)
service:
  pipelines:
    logs:
      processors:
        # - ...
        - transform/clear_logs_timestamp
```

### `sumologic` exporter: remove `routing_attributes_to_drop`

`routing_attributes_to_drop` has been removed from `sumologic` exporter in favor of `routing` processor's `drop_resource_routing_attribute`.

To migrate, perform the following steps:

- remove `routing_attributes_to_drop` from `sumologic` exporter
- add `drop_resource_routing_attribute` to `routing` processor

For example, given the following configuration:

```yaml
processors:
  routing:
    from_attribute: X-Tenant
    default_exporters:
    - jaeger
    table:
    - value: acme
      exporters: [jaeger/acme]
exporters:
  sumologic:
    routing_attributes_to_drop: X-Tenant
```

change it to:

```yaml
processors:
  routing:
    drop_resource_routing_attribute: true
    from_attribute: X-Tenant
    default_exporters:
    - jaeger
    table:
    - value: acme
      exporters: [jaeger/acme]
exporters:
  sumologic:
```

### `sumologic` exporter: deprecate `json_logs`

`json_logs` has been deprecated in favor of `transform` processor. It is going to be removed in `v0.95.0-sumo-0`.

To migrate perform the following steps:

- use `transform` processor in replace of `json_logs.add_timestamp` and `json_logs.timestamp_key`:

  ```yaml
  processors:
    transform/add_timestamp:
      log_statements:
        - context: log
          statements:
            - set(time, Now()) where time_unix_nano == 0
            - set(attributes["timestamp_key"], Int(time_unix_nano / 1000000))
  ```

- use `transform` processor in replace of `json_logs.flatten_body`:

  ```yaml
  processors:
    transform/flatten:
      error_mode: ignore
      log_statements:
        - context: log
          statements:
            - merge_maps(attributes, body, "insert") where IsMap(body)
            - set(body, "") where IsMap(body)

  ```

- use `transform` processor in replace of `json_logs.log_key`:

  ```yaml
  processors:
    transform/set_log_key:
      log_statements:
        - context: log
          statements:
            - set(attributes["log"], body)
            - set(body, "")
  ```

#### Migration example for `add_timestamp` and `timestamp_key`

Given the following configuration:

```yaml
exporters:
  sumologic:
    log_format: json
    json_logs:
      timestamp_key: timestamp_key
      add_timestamp: true
```

change it to:

```yaml
exporters:
  sumologic:
    log_format: json
    json_logs:
      add_timestamp: false
processors:
  transform/add_timestamp:
    log_statements:
      - context: log
        statements:
          - set(time, Now()) where time_unix_nano == 0
          - set(attributes["timestamp_key"], Int(time_unix_nano / 1000000))
service:
  pipelines:
    logs:
      processors:
        # ...
        - transform/add_timestamp
```

#### Migration example for `flatten_body`

Given the following configuration:

```yaml
exporters:
  sumologic:
    log_format: json
    json_logs:
      flatten_body: true
```

change it to:

```yaml
exporters:
  sumologic:
    log_format: json
    json_logs:
      flatten_body: false
processors:
  transform/flatten:
    error_mode: ignore
    log_statements:
      - context: log
        statements:
          - merge_maps(attributes, body, "insert") where IsMap(body)
          - set(body, "") where IsMap(body)
service:
  pipelines:
    logs:
      processors:
        # ...
        - transform/flatten
```

#### Migration example for `log_key`

Given the following configuration:

```yaml
exporters:
  sumologic:
    log_format: json
    json_logs:
      log_key: my_log
```

change it to:

```yaml
exporters:
  sumologic:
    log_format: json
    json_logs:
processors:
  transform/set_log_key:
    log_statements:
      - context: log
        statements:
          - set(attributes["my_log"], body)
          - set(body, "")
service:
  pipelines:
    logs:
      processors:
        # ...
        - transform/set_log_key
```

## Upgrading to v0.89.0-sumo-0

### `remoteobserver` processor: renamed to `remotetap` processor

To migrate, change the processor name `remoteobserver` to `remotetap` in your configuration files.

For example, given the following configuration:

```yaml
processors:
  remoteobserver:
    port: 1234
  remoteobserver/another-one:
    limit: 2

pipelines:
  logs:
    exporters: ["..."]
    processors:
    - remoteobserver
    receivers: ["..."]
  metrics:
    exporters: ["..."]
    processors:
    - remoteobserver/another-one
    receivers: ["..."]
```

change it to:

```yaml
processors:
  remotetap:
    port: 1234
  remotetap/another-one:
    limit: 2

pipelines:
  logs:
    exporters: ["..."]
    processors:
    - remotetap
    receivers: ["..."]
  metrics:
    exporters: ["..."]
    processors:
    - remotetap/another-one
    receivers: ["..."]
```

### `sumologic` exporter: changed default `timeout` from `5s` to `30s`

We believe 30 seconds is a better default timeout for the Sumo Logic exporter.
The bigger the payload is, the longer it takes for the Sumo Logic backend to process it.

If you want to revert to the previous behavior, set the `timeout` property to `5s` or another value. Example:

```yaml
exporters:
  sumologic:
    timeout: 5s
```

### `sumologic` extension: changed default `discover_collector_tags` from `false` to `true`

If you want to revert to the previous behavior, set the `discover_collector_tags` property to `false`. Example:

```yaml
extensions:
  sumologic:
    discover_collector_tags: false
```

## Upgrading to v0.84.0-sumo-0

### `sumologic` extension: removed `install_token` in favor of `installation_token`

The `install_token` configuration property was deprecated in `v0.72.0-sumo-0` in February 2023 in favor of `installation_token`.
It is now being removed in `v0.84.0-sumo-0`. To upgrade, replace `install_token` property with `installation_token` in your Sumo Logic extension configuration.

For example, change this:

```yaml
extensions:
  sumologic:
    install_token: xyz
```

to this:

```yaml
extensions:
  sumologic:
    installation_token: xyz
```

## Upgrading to v0.77.0-sumo-0

### Full Prometheus metric name normalization is now disabled by default

Prometheus and otel metric naming conventions aren't entirely compatible. Prometheus metric name normalization is a feature intended
to convert metrics between the two conventions. See the [upstream documentation][prometheus_naming] for details.

In the 0.76.1 release, the feature flag for this normalization was [enabled by default][feature_pr].

The feature primarily affects the `prometheus` and `prometheusremotewrite` exporters. However, it also modifies some metrics collected
by the `prometheus` receiver. More specifically, it trims certain suffixes from metric names. Unfortunately, this affects a lot of
widely used metrics. For example, the standard container CPU usage metric:

`container_cpu_usage_seconds_total` -> `container_cpu_usage_seconds`

This change breaks a lot of content built using existing metric names and prevents the Prometheus receiver from being used as a drop-in
replacement for Prometheus. Therefore, we've decided to default to having this flag disabled.

The current behaviour can be re-enabled by passing the `--feature-gates=+pkg.translator.prometheus.NormalizeName` flag to the collector at startup.

There is an ongoing discussion about making this behaviour configurable at runtime. Please follow [this issue][runtime_issue] if you'd like to learn more.

[prometheus_naming]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/pkg/translator/prometheus#metric-name
[feature_pr]: https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/20519
[runtime_issue]: https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/21743

## Upgrading to v0.73.0-sumo-1

### The default collector name for sumologic extension is now the host FQDN

In an effort to make defaults consistent between [resource detection] processor and the Sumo extension,
we've changed the default collector name and the host name it reports in its metadata to be the host
FQDN instead of the hostname reported by the OS. This makes the value consistent with the value of the
`host.name` attribute [resource detection] processor sets by default.

This will only affect newly registered collectors. If you have local credentials on your host, the
existing collector will be used, but if those credentials are cleared, a new collector will be created
with a different name. If you'd like to keep using the old name, set `CollectorName` explicitly in the
extension settings.

[resource detection]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/resourcedetectionprocessor

## Upgrading to v0.66.0-sumo-0

### `filelog` receiver: has been removed from sub-parsers

`preserve_to` has been removed from sub-parsers ([#9331]).

Now sub-parsers behaves like `preserve_to` would be equal to `parse_from`.

### `sending_queue`: require explicit storage set

`persistent_storage_enabled` configuration option is no longer available ([#5784]).
It is replaced by `storage` which takes component name as value:

```yaml
exporters:
  sumologic:
    sending_queue:
      enabled: true
      storage: file_storage

extensions:
  file_storage:
    directory: .
  sumologic:
    collector_name: sumologic-demo
    installation_token: ${SUMOLOGIC_INSTALLATION_TOKEN}
```

### `apache` receiver: turn on feature gates for resource attributes

The metrics in this receiver are now being sent with two resource attributes: `apache.server.name` and `apache.server.port`.
Additionally, `apache.server.name` replaces `server_name` metric-level attribute.

Both features are hidden behind feature gates, but because they are important for Sumo Logic apps,
they have been enabled by default ahead of the normal deprecation timeline.

To disable the new features, disable the feature gates in otelcol's arguments:

```bash
otelcol-sumo --config=file:config.yaml --feature-gates=-receiver.apache.emitServerNameAsResourceAttribute,-receiver.apache.emitPortAsResourceAttribute
```

More information about the feature gates can be found [here][apache-feature-gates].
The target release for the removal of feature gates is `v0.68`.

### `elasticsearch` receiver: turn on more datapoints

The metrics `elasticsearch.index.operation.count`, `elasticsearch.index.operation.time` and `elasticsearch.cluster.shards` emit more data points now.

These features are hidden behind feature gates, but because they are important for Sumo Logic apps,
they have been enabled by default ahead of the normal deprecation timeline.

To disable the new features, disable the feature gates in otelcol's arguments:

```bash
otelcol-sumo --config=file:config.yaml --feature-gates=-receiver.elasticsearch.emitClusterHealthDetailedShardMetrics,-receiver.elasticsearch.emitAllIndexOperationMetrics
```

More information about the feature gates can be found [here][elasticsearch-feature-gates].
The target release for the removal of feature gates is `v0.71`.

[#5784]: https://github.com/open-telemetry/opentelemetry-collector/pull/5784
[#9331]: https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/9331
[apache-feature-gates]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.64.0/receiver/apachereceiver#feature-gate-configurations
[elasticsearch-feature-gates]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/elasticsearchreceiver#feature-gate-configurations

## Upgrading to v0.57.2-sumo-0

### `sumologic` exporter: drop support for source headers

The following properties of the [Sumo Logic exporter][sumologicexporter_docs] are deprecated in `v0.57.2-sumo-0`:

- `source_category`
- `source_host`
- `source_name`

To upgrade, move these properties from the Sumo Logic exporter configuration
to the [Source processor][sourceprocessor_docs].

For example, the following configuration:

```yaml
exporters:
  sumologic:
    source_category: my-source-category
    source_host: my-source-host
    source_name: my-source-name
```

should be changed to the following:

```yaml
processors:
  source:
    source_category: my-source-category
    source_category_prefix: ""
    source_category_replace_dash: "-"
    source_host: my-source-host
    source_name: my-source-name
```

The reason for the additional `source_category_prefix` and `source_category_replace_dash`
is that the Source processor has more features and these properties must be set
to make it behave like the Sumo Logic exporter.

See the [Source processor documentation][sourceprocessor_docs] for more details.

[sumologicexporter_docs]: https://github.com/sumologic/sumologic-otel-collector/tree/v0.57.2-sumo-0/pkg/exporter/sumologicexporter/README.md
[sourceprocessor_docs]: https://github.com/sumologic/sumologic-otel-collector/tree/v0.57.2-sumo-0/pkg/processor/sourceprocessor/README.md

## Upgrading to v0.56.0-sumo-0

### `sumologic` exporter: drop support for translating attributes

Translating the attributes is harmless, but the exporters should not modify the data.
This functionality has been moved to the [sumologicschema processor][sumologicschema_processor].
Due to that, this functionality is now deprecated in the exporter and will be removed soon.

However, if the attributes are not translated, some Sumo Logic apps might not work correctly.
To migrate, add a `sumologicschema` processor to your pipelines that use the `sumologic` exporter and disable that functionality in the exporter:

```yaml
processors:
  # ...
  sumologic_schema:
    translate_attributes: true

exporters:
  sumologic:
    # ...
    translate_attributes: false

# ...

service:
  pipelines:
    logs:
      processors:
        # - ...
        - sumologic_schema
```

**Note**: by default, the `sumologicschema` processor also performs other actions, like adding `cloud.namespace` attribute to the data.
If you don't want this to happen, you should explicitly disable these functionalities:

```yaml
processors:
  sumologic_schema:
    add_cloud_namespace: false
    # ...
```

Full list of configuration settings can be found in the [readme][sumologicschema_processor_readme] of the processor.

### `sumologic` exporter: drop support for translating Telegraf metric names

Similar as above, the translation should not happen in the exporter and has been moved to the [sumologicschema processor][sumologicschema_processor].
The functionality is now deprecated and will be removed soon.

However, if the attributes are not translated, some Sumo Logic apps might not work correctly.
To migrate, add a `sumologicschema` processor to your pipelines that use the `sumologic` exporter and disable that functionality in the exporter:

```yaml
processors:
  # ...
  sumologic_schema:
    translate_telegraf_attributes: true

exporters:
  sumologic:
    # ...
    translate_telegraf_attributes: false

# ...

service:
  pipelines:
    logs:
      processors:
        # - ...
        - sumologic_schema
```

**Note**: By default, the `sumologicschema` processor also performs other actions. Please see a corresponding warning in paragraph [`sumologic` exporter: drop support for translating attributes](#sumologic-exporter-drop-support-for-translating-attributes) for more information.

[sumologicschema_processor]: https://github.com/sumologic/sumologic-otel-collector/tree/v0.56.0-sumo-0/pkg/processor/sumologicschemaprocessor/
[sumologicschema_processor_readme]: https://github.com/sumologic/sumologic-otel-collector/tree/v0.56.0-sumo-0/pkg/processor/sumologicschemaprocessor/README.md

## Upgrading to v0.55.0-sumo-0

### `filter` processor: drop support for `expr` language

Expr language is supported by [logstransform] processor, so there is no need to have this functionality in [filter] processor.

The following configuration of `filter` processor:

```yaml
processors:
  filter:
    logs:
      include:
        match_type: expr
        expressions:
          - Body matches "log to include"
      exclude:
        match_type: expr
        expressions:
          - Body matches "log to exclude"
```

is equivalent of the following configuration of `logstransform` processor:

```yaml
processors:
  logstransform:
    operators:
      - type: filter
        expr: 'body matches "log to include"'
      - type: filter
        expr: 'not(body matches "log to exclude")'
```

[logstransform]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.55.0/processor/logstransformprocessor
[filter]: https://github.com/SumoLogic/opentelemetry-collector-contrib/tree/main/processor/filterprocessor

## Upgrading to v0.52.0-sumo-0

### `sumologic` exporter: Removed `carbon2` and `graphite` metric formats

These metric formats don't offer any advantages over Prometheus or OTLP formats. To migrate, simply switch
the format to `prometheus`.

```yaml
exporters:
  sumologic:
    metric_format: prometheus
```

## Upgrading to v0.51.0-sumo-0

### `k8s_tagger` processor: removed `clusterName` metadata extraction option

Before `v0.51.0-sumo-0`, you could specify `clusterName` as one of the options for metadata extraction:

```yaml
processors:
  k8s_tagger:
    extract:
      metadata:
      - clusterName
```

Starting with `v0.51.0-sumo-0`, the `clusterName` option is removed.
This is a result of an upstream change that removes the `k8s.cluster.name` metadata ([link](https://github.com/open-telemetry/opentelemetry-collector-contrib/pull/9885)),
which in turn is a result of Kubernetes libraries removing support for this deprecated and non-functional field ([link](https://github.com/kubernetes/apimachinery/commit/430b920312ca0fa10eca95967764ff08f34083a3)).

#### How to upgrade

Check if you are using the `extract.metadata` option in your `k8s_tagger` processor configuration
and if yes, check if it includes the `clusterName` entry. If it does, remove it.

For example, change the following configuration:

```yaml
processors:
  k8s_tagger:
    extract:
      metadata:
      - clusterName
      - deploymentName
      - podName
      - serviceName
      - statefulSetName
```

to the following, removing the `clusterName` entry from the list:

```yaml
processors:
  k8s_tagger:
    extract:
      metadata:
      - deploymentName
      - podName
      - serviceName
      - statefulSetName
```

### `sumologic` exporter: metadata translation: changed the attribute that is translated to `_sourceName` from `file.path.resolved` to `log.file.path_resolved`

This change should have landed already in previous version `v0.49.0-sumo-0`.
It is a result of a change in [Filelog receiver][filelog_receiver_v0_49_0], which starting with `v0.49.0`,
changed the names of the attributes it creates when one of the  configuration properties is set to `true`:

- `include_file_name`: changed attribute name from `file.name` to `log.file.name`
- `include_file_name_resolved`: changed attribute name from `file.name.resolved` to `log.file.name_resolved`
- `include_file_path`: changed attribute name from `file.path` to `log.file.path`
- `include_file_path_resolved`: changed attribute name from `file.path.resolved` to `log.file.path_resolved`

See [documentation][ot_logs_collection_v0_29_0] that describes this change.

#### How to upgrade

This change is technically a breaking change, but it should be transparent and require no changes in your configuration.

If you have a pipeline that includes Filelog receiver (configured with `include_file_path_resolved: true`) and Sumo Logic exporter,
the Filelog receiver will create the `log.file.path_resolved` attribute
and Sumo Logic exporter will translate this attribute to `_sourceName`.

```yaml
exporters:
  sumologic:
    endpoint: ...

receivers:
  filelog:
    include:
    - ...
    include_file_path_resolved: true

service:
  pipelines:
    logs:
      exporters:
      - sumologic
      receivers:
      - filelog
```

In the unlikely scenario that you have a component other than Filelog receiver that creates the `file.path.resolved` attribute
that you relied on Sumo Logic exporter to be translated to `_sourceName` attribute,
you can perform the translation with Resource processor like the following:

```yaml
processors:
  resource:
    attributes:
    - action: insert
      from_attribute: file.path.resolved
      key: _sourceName
    - action: delete
      key: file.path.resolved
```

[filelog_receiver_v0_49_0]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.49.0/receiver/filelogreceiver
[ot_logs_collection_v0_29_0]: https://github.com/open-telemetry/opentelemetry-log-collection/releases/tag/v0.29.0

## Upgrading to 0.49.0-sumo-0

### Several changes to receivers using opentelemetry-log-collection

The affected receivers are: `filelog`, `syslog`, `tcplog`, `udplog`, and `journald`.

[opentelemetry-log-collection] is a shared library used by the aforementioned receivers. This release contains several breaking
changes to its configuration syntax.
Please refer to the [official upgrade guide][opentelemetry-log-collection-upgrade-guide] for more information.

[opentelemetry-log-collection]: https://github.com/open-telemetry/opentelemetry-log-collection
[opentelemetry-log-collection-upgrade-guide]: https://github.com/open-telemetry/opentelemetry-log-collection/releases/tag/v0.29.0

### Sumo Logic exporter metadata handling

The [OpenTelemetry data format][ot-data-format] makes a distinction between record-level attributes and
resource-level attributes. The `metadata_attributes` configuration option in the [`sumologicexporter`][sumologicexporter]
allowed setting metadata for records sent to the Sumo Logic backend based on both record and resource-level
attributes. Only attributes matching the supplied regular expressions were sent.

However, this is conceptually incompatible with OpenTelemetry. Our intent with the exporter is to use OpenTelemetry
conventions as much as we can, to the point where it should eventually be possible to export data to Sumo using the
upstream OTLP exporter. This is why we are changing the behaviour. From now on:

1. `metadata_attributes` no longer exists.
1. Metadata for sent records is based on resource-level attributes.

In order to retain current behaviour, processors should be used to transform the data before it is exported. This
potentially involves two transformations:

#### Removing unnecessary metadata using the resourceprocessor

`metadata_attributes` allowed filtering based on regular expressions. An equivalent processor doesn't yet
exist, but resource-level attributed can be dropped using the [resourceprocessor][resourceprocessor]. For example:

```yaml
processors:
  resource:
    attributes:
      - pattern: ^k8s\.pod\..*
        action: delete
```

will delete all attributes starting with `k8s.pod.`.

**NOTE**: The ability to delete attributes based on a regular expression is currently unique to our fork of the
[resourceprocessor][resourceprocessor], and isn't available in upstream.

#### Moving record-level attributes used for metadata to the resource level

This can be done using the [Group by Attributes processor][groupbyattrsprocessor]. If you were using the Sumo Logic
exporter to export data with a `host` record-level attribute:

```yaml
exporters:
  sumologicexporter:
    ...
    metadata_attributes:
      - host
```

You can achieve the same effect with the following processor configuration:

```yaml
processors:
  groupbyattrsprocessor:
    keys:
      - host
```

Keep in mind that your attribute may already be resource-level, in which case no changes are necessary.

[ot-data-format]: https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/common/README.md
[groupbyattrsprocessor]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/groupbyattrsprocessor
[resourceprocessor]: https://github.com/SumoLogic/opentelemetry-collector-contrib/tree/2ae9e24dc7efd940e1aa2f6efb288504b591af9b/processor/resourceprocessor
[sumologicexporter]: https://github.com/SumoLogic/sumologic-otel-collector/tree/v0.48.0-sumo-0/pkg/exporter/sumologicexporter
