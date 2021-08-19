# Migration from Installed Collector

Installed Collector is able to gather data from several different type of source.
Every source should be manually migrated to Opentelemetry Configuration.

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

This section describes migration steps for sources managed from cloud.

### Local File Source

#### Overall example

Below is an example of an Open Telemetry configuration for local file source

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

In Opentelemetry Collector it can be added after slash `/` to the receiver name.

For `Name` `my example name` it will look like in following snippet:

```yaml
receivers:
  filelog/my example name:
  # ...
```

#### Description

In Opentelemetry Collector configuration it can be added as a comment just above the receiver name.

For `Description` `All my example logs` it will look like in following snippet:

```yaml
receivers:
  ## All my example logs
  filelog/my example name:
  # ...
```

#### File Path

As like Installed Collector, Opentelemetry Collector supports regular expression for paths.
In addition, multiple different paths/expressions can be specified.
They should be added as elements of `include` configuration option.

For all `.log` files from `/var/log/` and `/opt/my_app/` following snippet will apply:

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

Opentelemetry Collector doesn't have a substitute for this Installed Collector option.
Supported are only two modes, start at file beginning or file end.
Starting at `beginning` will read all file every time it is started,
starting at `end` will take only logs appended to file after Opentelemetry Collector is started.
This is configurable by `start_at` option.

To read logs only appended logs, following snippet can be used:

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

Source Host can be set in exporter configuration in `source_host` option.

For example for `My Host` following snippet can be used:

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

Source Category can be set in exporter configuration in `source_category` option.

For example for `My Category` following snippet can be used:

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

To set custom fields [resourceprocessor][resourceprocessor] can be used.

For example to add two more fields `cloud.availability_zone` and `k8s.cluster.name`,
following config can be used

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

Substitute for Denylist in Opentelemetry is `exclude` list in filelog receiver.

For example to exclude `/var/log/sensitive.log` following config can be used.

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

Substitute for `Extract timestamp information from log file entries` is `clear_logs_timestamp`
in exporter configuration. This is by default set to `true`.

This works like `Extract timestamp information from log file entries` combined with
`Ignore time zone from log file and instead use:` set to `Use Collector Default`.
To set time_zone for collector, `extensions.sumologic.time_zone` has to be set like in following example.

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

If `clear_logs_timestamp` is set to `false`, timestamp parsing should be configured manually.
To do that, following code snippet can be a good template.

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

To get receipt time, parsing can be skipped like in following example.

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
    ## Keep manually parsed timestamps (get receipt time by default)
    clear_logs_timestamp: true
```

##### Encoding

Substitute for `Encoding` is `encoding`. Full list of supporter encodings can be obtained from [filelogreceiver documentation][supported_encodings].

See following snippet for example usage:

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

Multiline in Opentelemetry Collector has to be always specified manually. There is no automatic boundary detection.

For boundary regex `^\d{4}-\d{2}-\d{2}` (matches for example `2021-06-06`) following config can be used.

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

Opentelemetry Collector supports also `line_end_pattern`. It can be used for multiline logs, where log has known end pattern.

More information is available in [filelogreceiver documentation][multiline]

## Local Configuration File

This section describes migration steps for sources managed locally.

[resourceprocessor]: https://github.com/open-telemetry/opentelemetry-collector/tree/v0.31.0/processor/resourceprocessor
[multiline]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.31.0/receiver/filelogreceiver#multiline-configuration
[supported_encodings]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.31.0/receiver/filelogreceiver#supported-encodings
