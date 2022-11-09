# Mapping OpenTelemetry concepts to Sumo Logic

## Data model considerations

OpenTelemetry has a [rich data model], which is internally constructed out of several layers. For all signals,
these can be broken down into following:

- **Resource** - includes Attributes describing the resource from which given set of data comes from.
  Should follow [resource semantic conventions].
- **Instrumentation Scope** - additional information about the scope of data (e.g. instrumentation library name).
- **Record** - specific record (Log, Span, Metric) that also includes its own set of Attributes. May follow given
  signal type semantic conventions (e.g. [trace], [metrics], [logs]), though they may also contain key/values that are
  specific to the context of the record. Additionally, Logs can also include attributes in the Body.

As can be observed, while attributes can be present at both Resource and Record level currently, they are not created
equal and should be interpreted separately. Resource-level attributes have a much wider scope and are used to identify
where the data comes from while Record-level attributes context is much narrower, related just to the single record,
frequently with much high cardinality of both keys and values.

At Sumo Logic, there is a concept of [Fields](https://help.sumologic.com/docs/manage/fields) for log data. Fields offer
a powerful capability to associate indexable metadata with logs, though only limited number of them can be used
at a given time. Also, they need to be defined first.

Looking from the OpenTelemetry standpoint, Fields are a good match for Resource-level attributes,
while Log Record-level attributes are good fit for [structured representation of the log via JSON], which
is automatically supported by Sumo Logic Search.

All resource-level attributes are stored as fields. If a matching field is not defined, it will be skipped (the list
of ignored fields can be checked via [dropped fields view]).
When log contains record-level attributes, they are stored as JSON representation. Body, if present, is then
stored under `log` key.

### Examples

#### Log with both Resource-level and Record-level attributes

Consider following input log:

```
  Resource:
    Attributes:
      "indexed-field": "some value"
  Log:
    Body: "sample body"
    Attributes:
      "log-level-attribute": 42
```

Such log will be stored as following set of data at Sumo Logic:

```
  Fields:
    "indexed-field": "some value"

  _raw (JSON): {
    "log": "sample body",
    "log-level-attribute": "42"
  }
```

#### Log with Resource-level attributes only

In case of no log-level attributes, body is stored inline. I.e. for following input:

```
  Resource:
    Attributes:
      "indexed-field": "some value"
  Log:
    Body: "sample body"
```

The output is stored as:

```
  Fields:
    "indexed-field": "some value"

  _raw: "sample body"
```

[rich data model]: https://github.com/open-telemetry/opentelemetry-proto/tree/main/opentelemetry/proto
[resource semantic conventions]: https://github.com/open-telemetry/opentelemetry-specification/tree/main/specification/resource/semantic_conventions
[trace]: https://github.com/open-telemetry/opentelemetry-specification/tree/main/specification/trace/semantic_conventions
[metrics]: https://github.com/open-telemetry/opentelemetry-specification/tree/main/specification/metrics/semantic_conventions
[logs]: https://github.com/open-telemetry/opentelemetry-specification/tree/main/specification/logs/semantic_conventions
[structured representation of the log via JSON]: https://help.sumologic.com/docs/search/get-started-with-search/search-basics/view-search-results-json-logs
[dropped fields view]: https://help.sumologic.com/docs/manage/fields/#view-dropped-fields
