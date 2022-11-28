# Upgrading

- [Unreleased](#unreleased)
  - [`filelog` receiver: has been removed from sub-parsers](#filelog-receiver-has-been-removed-from-sub-parsers)
  - [`sending_queue`: require explicit storage set](#sending_queue-require-explicit-storage-set)
  - [`apache` receiver: turn on feature gates for resource attributes](#apache-receiver-turn-on-feature-gates-for-resource-attributes)
  - [`elasticsearch` receiver: turn on feature gates for resource attributes](#elasticsearch-receiver-turn-on-feature-gates-for-resource-attributes)
- [Upgrading to v0.57.2-sumo-0](#upgrading-to-v0572-sumo-0)
  - [`sumologic` exporter: drop support for source templates](#sumologic-exporter-drop-support-for-source-headers)
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

## Unreleased

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
    install_token: ${SUMOLOGIC_INSTALL_TOKEN}
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

### `elasticsearch` receiver: turn on feature gates for resource attributes

The metrics `elasticsearch.index.operation.count`, `elasticsearch.index.operation.count` and `elasticsearch.cluster.shards` emit more data points now.

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

Since the exporter should not modify the data, setting the source headers is no longer supported in the `sumologic` exporter.
This deprecation includes deprecating the source templates.

If you need to set the `_source*` headers in the collector, use the `transform` processor, for example:

```yaml
processors:
  # ...
  transform:
    logs:
      - set(attributes["_sourceHost"], "source_host")
```

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

[sumologicschema_processor]: ../pkg/processor/sumologicschemaprocessor/
[sumologicschema_processor_readme]: ../pkg/processor/sumologicschemaprocessor/README.md

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
[filter]: https://github.com/SumoLogic/opentelemetry-collector-contrib/tree/v0.54.0-filterprocessor/processor/filterprocessor

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

### Several changes to receivers using [opentelemetry-log-collection][opentelemetry-log-collection]

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

#### Removing unnecessary metadata using the [resourceprocessor][resourceprocessor]

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
