# Migration from Installed Collector

The Installed Collector can gather data from several different types of Sources.
You should manually migrate your Sources to an OpenTelemetry Configuration.

- [Cloud Based Management](#cloud-based-management)
  - [Local File Source](#local-file-source)
    - [Overall example](#overall-example)
    - [Name](#name)
    - [Description](#description)
    - [File Path](#file-path)
      - [Collection should begin](#collection-should-begin)
    - [Source Host](#source-host)
    - [Source Category](#source-category)
    - [Fields](#fields)
    - [Advanced Options for Logs](#advanced-options-for-logs)
      - [Denylist](#denylist)
    - [Timestamp Parsing](#timestamp-parsing)
    - [Encoding](#encoding)
    - [Multiline Processing](#multiline-processing)
- [Local Configuration File](#local-configuration-file)

## Cloud Based Management

This section describes migration steps for Sources managed from the Cloud.

### Local File Source

#### Overall example

Below is an example of an OpenTelemetry configuration for a Local File Source.

```yaml
extensions:
  sumologic:
    access_id: <access_id>
    access_key: <access_key>
    ## Time Zone is a substitute of Installed Collector `Time Zone`
    ## with `Use time zone from log file. If none is detected use:` option.
    ## This is used only if `clear_logs_timestamp` is set to `true` in sumologic exporter.
    ## Full list of time zones is available on wikipedia:
    ## https://en.wikipedia.org/wiki/List_of_tz_database_time_zones#List
    time_zone: America/Tijuana
receivers:
  ## There is no substitute for `Description` in current project phase.
  ## It is recommended to use comments for that purpose, like this one.
  ## filelog/<source group name>:
  ## <source group name> can be substitute of Installed Collector `Name`.
  filelog/log source:
    ## List of local files which should be read.
    ## Installed Collector `File path` substitute.
    include:
    - /var/log/*.log
     - /opt/app/logs/*.log
    ## List of local files which shouldn't be read.
    ## Installed Collector `Denylist` substitute.
    exclude:
    - /var/log/auth.log
    - /opt/app/logs/security_*.log
    ## There is no substitute of Installed Collector `Collection should begin`.
    ## This is nearest config and can take one of two values: `beginning` or `end`.
    start_at: beginning
    ## encoding is substitute for Installed Collector `Encoding`.
    ## List of supported encodings:
    ## https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.31.0/receiver/filelogreceiver
    encoding: utf-8
    ## multiline is Opentelemetry Collector substitute for `Enable Multiline Processing`.
    ## As multiline detection behaves slightly different than in Installed Collector
    ## following section in filelog documentation is recommended to read:
    ## https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.31.0/receiver/filelogreceiver#multiline-configuration
    multiline:
      ## line_start_pattern is substitute of `Boundary Regex`.
      line_start_pattern: ^\d{4}
processors:
  ## Following configuration will add two fields to every record
  resource/log source:
    attributes:
    - key: cloud.availability_zone
      value: zone-1
      action: insert
    - key: k8s.cluster.name
      value: my-cluster
      action: insert
exporters:
  sumologic:
    ## Installed Collector substitute for `Source Category`.
    source_category: example category
    ## Installed Collector substitute for `Source Host`.
    source_host: example host
    ## clear_logs_timestamp is by default set to True.
    ## If it's set to true, it works like `Enable Timestamp Parsing`,
    ## and `Time Zone` is going to be taken from `extensions` section.
    ## There is no possibility to configure several time zones in one exporter.
    ## clear_logs_timestamp sets to true also behaves like
    ## `Timestamp Format` would be set to `Automatically detect the format`
    ## in terms of Installed Collector configuration.
    clear_logs_timestamp: true
service:
  extensions:
  - sumologic
  pipelines:
    logs/log source:
      receivers:
      - filelog/log source
      processors:
      - resource/log source
      exporters:
      - sumologic
```

#### Name

Define the name after the slash `/` in the receiver name.

For example, the following snippet configures the name as `my example name`:

```yaml
receivers:
  filelog/my example name:
  # ...
```

#### Description

A description can be added as a comment just above the receiver name.

For example, the following snippet configures the description as `All my example logs`:

```yaml
receivers:
  ## All my example logs
  filelog/my example name:
  # ...
```

#### File Path

Like the Installed Collector, the OpenTelemetry Collector supports regular expression for paths.
In addition, you can specify multiple different path expressions.
Add them as elements of the `include` configuration option.

For example, the following snippet configures the path to all `.log` files from `/var/log/` and `/opt/my_app/`:

```yaml
receivers:
  ## All my example logs
  filelog/my example name:
    include:
    - /var/log/*.log
    - /opt/my_app/*.log
  # ...
```

##### Collection should begin

The OpenTelemetry Collector doesn't have a substitute for this Installed Collector option.
It supports two options, starting at the beginning or end of a file.
Starting at the `beginning` will read the entire file every time it's started.
Starting at the `end` will read only logs appended to file after it's started.
This is configurable with the `start_at` option.

For example, the following snippet configures the Collector to only read appended logs:

```yaml
receivers:
  ## All my example logs
  filelog/my example name:
    include:
    - /var/log/*.log
    - /opt/my_app/*.log
    start_at: end
  # ...
```

#### Source Host

The Source Host is set in the exporter configuration with the `source_host` option.

For example, the following snippet configures the Source Host as `My Host`:

```yaml
receivers:
  ## All my example logs
  filelog/my example name:
    include:
    - /var/log/*.log
    - /opt/my_app/*.log
    start_at: end
  # ...
exporters:
  sumologic/some name:
    source_host: My Host
```

#### Source Category

The Source Category is set in the exporter configuration with the `source_category` option.

For example, the following snippet configures the Source Category as `My Category`:

```yaml
receivers:
  ## All my example logs
  filelog/my example name:
    include:
    - /var/log/*.log
    - /opt/my_app/*.log
    start_at: end
  # ...
exporters:
  sumologic/some name:
    source_host: My Host
    source_category: My Category
```

#### Fields

Use the [resourceprocessor][resourceprocessor] to set custom fields.

For example, the following snippet configures two fields, `cloud.availability_zone` and `k8s.cluster.name`:

```yaml
receivers:
  ## All my example logs
  filelog/my example name:
    include:
    - /var/log/*.log
    - /opt/my_app/*.log
    start_at: end
  # ...
processors:
  resource/my example name fields:
    attributes:
    - key: cloud.availability_zone
      value: zone-1
      action: upsert
    - key: k8s.cluster.name
      value: my-cluster
      action: insert
exporters:
  sumologic/some name:
    source_host: My Host
    source_category: My Category
```

#### Advanced Options for Logs

##### Denylist

Use the `exclude` option in the filelog receiver to specify files you don't want collected.

For example, the following snippet excludes `/var/log/sensitive.log` from collection:

```yaml
receivers:
  ## All my example logs
  filelog/my example name:
    include:
    - /var/log/*.log
    - /opt/my_app/*.log
    exclude:
    - /var/log/sensitive.log
    start_at: end
  # ...
processors:
  resource/my example name fields:
    attributes:
    - key: cloud.availability_zone
      value: zone-1
      action: upsert
    - key: k8s.cluster.name
      value: my-cluster
      action: insert
exporters:
  sumologic/some name:
    source_host: My Host
    source_category: My Category
```

##### Timestamp Parsing

The Installed Collector option to `Extract timestamp information from log file entries` in an
OpenTelemtry configuration is `clear_logs_timestamp`. This is set to `true` by default.

This works like `Extract timestamp information from log file entries` combined with
`Ignore time zone from log file and instead use:` set to `Use Collector Default`.

For example, the following configuration sets the time_zone for a Collector with `extensions.sumologic.time_zone`:

```yaml
extensions:
  sumologic:
    time_zone: America/Tijuana
receivers:
  ## All my example logs
  filelog/my example name:
    include:
    - /var/log/*.log
    - /opt/my_app/*.log
    exclude:
    - /var/log/sensitive.log
    start_at: end
  # ...
processors:
  resource/my example name fields:
    attributes:
    - key: cloud.availability_zone
      value: zone-1
      action: upsert
    - key: k8s.cluster.name
      value: my-cluster
      action: insert
exporters:
  sumologic/some name:
    source_host: My Host
    source_category: My Category
```

If `clear_logs_timestamp` is set to `false`, timestamp parsing should be configured
manually, like in the following snippet:

```yaml
extensions:
  sumologic:
    time_zone: America/Tijuana
receivers:
  ## All my example logs
  filelog/my example name:
    include:
    - /var/log/*.log
    - /opt/my_app/*.log
    exclude:
    - /var/log/sensitive.log
    start_at: end
    operators:
    ## Extract timestamp into timestamp field using regex
    ## rel: https://github.com/sumo-drosiek/opentelemetry-log-collection/blob/b506aadf913d6c1691cef10a534d495338c87dee/docs/operators/regex_parser.md
    - type: regex_parser
      regex: (?P<timestamp>^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2},\d{3} (\+|\-)\d{4})
      ## Keep original record in log field
      preserve_to: $$body.log
      ## Parse timestamp from timestamp field
      ## rel: https://github.com/sumo-drosiek/opentelemetry-log-collection/blob/b506aadf913d6c1691cef10a534d495338c87dee/docs/operators/time_parser.md
      timestamp:
        parse_from: $$body.timestamp
        ## Layout are substitute for Timestamp Format configuration
        layout_type: gotime
        layout: '2006-01-02 15:04:05,000 -0700'
    ## Restore record from log field
    ## rel: https://github.com/sumo-drosiek/opentelemetry-log-collection/blob/b506aadf913d6c1691cef10a534d495338c87dee/docs/operators/restructure.md#move
    - type: restructure
      ops:
      - move:
        from: $$body.empty_timestamp
        to: $$timestamp
  # ...
processors:
  resource/my example name fields:
    attributes:
    - key: cloud.availability_zone
      value: zone-1
      action: upsert
    - key: k8s.cluster.name
      value: my-cluster
      action: insert
exporters:
  sumologic/some name:
    source_host: My Host
    source_category: My Category
    ## Keep manually parsed timestamps
    clear_logs_timestamp: true
```

The following example snippet skips timestamp parsing so the Collector uses Receipt Time:

```yaml
extensions:
  sumologic:
    time_zone: America/Tijuana
receivers:
  ## All my example logs
  filelog/my example name:
    include:
    - /var/log/*.log
    - /opt/my_app/*.log
    exclude:
    - /var/log/sensitive.log
    start_at: end
  # ...
processors:
  resource/my example name fields:
    attributes:
    - key: cloud.availability_zone
      value: zone-1
      action: upsert
    - key: k8s.cluster.name
      value: my-cluster
      action: insert
exporters:
  sumologic/some name:
    source_host: My Host
    source_category: My Category
    ## Keep manually parsed timestamps (use Receipt Time by default)
    clear_logs_timestamp: true
```

##### Encoding

Use `encoding` to set the encoding of your data. Full list of supporter encodings can be obtained from [filelogreceiver documentation][supported_encodings].

The following snippet sets the encoding to UTF-8:

```yaml
extensions:
  sumologic:
    time_zone: America/Tijuana
receivers:
  ## All my example logs
  filelog/my example name:
    include:
    - /var/log/*.log
    - /opt/my_app/*.log
    exclude:
    - /var/log/sensitive.log
    start_at: end
    encoding: utf-8
  # ...
processors:
  resource/my example name fields:
    attributes:
    - key: cloud.availability_zone
      value: zone-1
      action: upsert
    - key: k8s.cluster.name
      value: my-cluster
      action: insert
exporters:
  sumologic/some name:
    source_host: My Host
    source_category: My Category
```

##### Multiline Processing

Multiline processing in the Opentelemetry Collector is set manually. There is no automatic boundary detection.

The following snippet sets the boundary regex as `^\d{4}-\d{2}-\d{2}` to match, for example, `2021-06-06`):

```yaml
extensions:
  sumologic:
    time_zone: America/Tijuana
receivers:
  ## All my example logs
  filelog/my example name:
    include:
    - /var/log/*.log
    - /opt/my_app/*.log
    exclude:
    - /var/log/sensitive.log
    start_at: end
    encoding: utf-8
    multiline:
      line_start_pattern: ^\d{4}-\d{2}-\d{2}
  # ...
processors:
  resource/my example name fields:
    attributes:
    - key: cloud.availability_zone
      value: zone-1
      action: upsert
    - key: k8s.cluster.name
      value: my-cluster
      action: insert
exporters:
  sumologic/some name:
    source_host: My Host
    source_category: My Category
```

If your multiline logs have a known end pattern use the `line_end_pattern` option.

More information is available in [filelogreceiver documentation][multiline].

## Local Configuration File

This section describes migration steps for Sources managed locally.

[resourceprocessor]: https://github.com/open-telemetry/opentelemetry-collector/tree/v0.31.0/processor/resourceprocessor
[multiline]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.31.0/receiver/filelogreceiver#multiline-configuration
[supported_encodings]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.31.0/receiver/filelogreceiver#supported-encodings
