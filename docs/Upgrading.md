# Upgrading

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
