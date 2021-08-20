# Migration from Installed Collector

The Installed Collector can gather data from several different types of Sources.
You should manually migrate your Sources to an OpenTelemetry Configuration.

- [General Configuration Concepts](#general-configuration-concepts)
- [Collector](#collector)
  - [Name](#name)
  - [Description](#description)
  - [Host Name](#host-name)
  - [Category](#category)
  - [Fields](#fields)
  - [Assign to a budget](#assign-to-a-budget)
  - [Time Zone](#time-zone)
  - [Advanced](#advanced)
    - [CPU Target](#cpu-target)
    - [Collector Management](#collector-management)
- [Cloud Based Management](#cloud-based-management)
  - [Local File Source](#local-file-source)
    - [Overall example](#overall-example)
    - [Name](#name-1)
    - [Description](#description-1)
    - [File Path](#file-path)
      - [Collection should begin](#collection-should-begin)
    - [Source Host](#source-host)
    - [Source Category](#source-category)
    - [Fields](#fields-1)
    - [Advanced Options for Logs](#advanced-options-for-logs)
      - [Denylist](#denylist)
      - [Timestamp Parsing](#timestamp-parsing)
      - [Encoding](#encoding)
      - [Multiline Processing](#multiline-processing)
  - [Remote File Source](#remote-file-source)
  - [Syslog Source](#syslog-source)
    - [Overall example](#overall-example-1)
    - [Name](#name-2)
    - [Description](#description-2)
    - [Protocol](#protocol-and-port)
    - [Source Category](#source-category-1)
    - [Fields](#fields-2)
    - [Advanced Options for Logs](#advanced-options-for-logs-1)
      - [Timestamp Parsing](#timestamp-parsing-1)
    - [Additional Configuration](#additional-configuration)
      - [Source Name](#source-name)
      - [Source Host](#source-host-1)
  - [Docker Logs Source](#docker-logs-source)
  - [Docker Stats Source](#docker-stats-source)
  - [Script Source](#script-source)
  - [Streaming Metrics Source](#streaming-metrics-source)
  - [Host Metrics Source](#host-metrics-source)
  - [Local Windows Event Log Source](#local-windows-event-log-source)
  - [Local Windows Performance Monitor Log Source](#local-windows-performance-monitor-log-source)
  - [Windows Active Directory Source](#windows-active-directory-source)
  - [Script Action](#script-action)
- [Local Configuration File](#local-configuration-file)
  - [Local File Source](#local-file-source-1)
  - [Remote File Source](#remote-file-source-1)
  - [Syslog Source](#syslog-source-1)
  - [Docker Logs Source](#docker-logs-source-1)
  - [Docker Stats Source](#docker-stats-source-1)
  - [Script Source](#script-source-1)
  - [Streaming Metrics Source](#streaming-metrics-source-1)
  - [Host Metrics Source](#host-metrics-source-1)
  - [Local Windows Event Log Source](#local-windows-event-log-source-1)
  - [Local Windows Performance Monitor Log Source](#local-windows-performance-monitor-log-source-1)
  - [Windows Active Directory Source](#windows-active-directory-source-1)
  - [Script Action](#script-action-1)

## General Configuration Concepts

Lets consider the following example:

```yaml
extensions:
  sumologic:
    access_id: <access_id>
    access_key: <access_key>

receivers:
  filelog:
    include:
      - /var/log/syslog
  tcplog:
    listen_address: 0.0.0.0:514

processors:
  memory_limiter:
    check_interval: 1s
    limit_mib: 4000
    spike_limit_mib: 800
  resource:
    attributes:
    - key: author
      value: me
      action: insert

exporters:
  sumologic:

service:
  extensions:
  - sumologic
  pipelines:
  logs/example pipeline:
    receivers:
    - filelog
    - tcplog
    processors:
    - memory_limiter
    - resource
    exporters:
    - sumologic
```

We can differentiate four types of modules:

- extensions - unrelated to data processing, but responsible for additional actions,
  like collector registration (eg. [sumologic extension][sumologicextension])
- receivers - responsible for receiving data and pushing it to processors
- processors - responsible for data modification, like adding fields, limiting memory and so on
- exporters - responsible for sending data, received by receivers and processed by processors

To use those configured modules, they need to be mentioned in `service` section.
`service` consists of `extensions` (they are global across collector) and `pipelines`.
`Pipelines` can be `logs`, `metrics` and `traces` and every of them can have
`receivers`, `processors` and `exporters`. Multiple pipelines of one type can be configured using aliases (`example pipeline` for `logs` in above example).

## Collector

Collector registration and configuration is handle by [sumologicextension][sumologicextension].

### Name

Collector name can specified by setting `collector_name` option:

```yaml
extensions:
  sumologic:
    access_id: <access_id>
    access_key: <access_key>
    collector_name: my_collector
```

### Description

To set a description, use `collector_description` option:

```yaml
extensions:
  sumologic:
    access_id: <access_id>
    access_key: <access_key>
    collector_name: my_collector
    collector_description: This is my and only my collector
```

### Host Name

Host name can be set in sumologic exporter configuration.
Exporter will set host name for every record being sended to Sumo:

```yaml
extensions:
  sumologic:
    access_id: <access_id>
    access_key: <access_key>
    collector_name: my_collector
    collector_description: This is my and only my collector
exporters:
  sumologic:
    source_host: My hostname
```

### Category

To set Collector category, use `collector_category` configuration:

```yaml
extensions:
  sumologic:
    access_id: <access_id>
    access_key: <access_key>
    collector_name: my_collector
    collector_description: This is my and only my collector
    collector_category: example
exporters:
  sumologic:
    source_host: My hostname
```

### Fields

Fields in Opentelemetry Collector can be added using [resourceprocessor][resourceprocessor].
For example to add a field with key `author` and with value `me` to every record,
please consider the following configuration:

```yaml
extensions:
  sumologic:
    access_id: <access_id>
    access_key: <access_key>
    collector_name: my_collector
    collector_description: This is my and only my collector
    collector_category: example
processors:
  resource:
    attributes:
    - key: author
      value: me
      action: insert
exporters:
  sumologic:
    source_host: My hostname
```

### Assign to a budget

Assign to a budget is not supported by Opentelemetry Collector.

### Time Zone

To set Collector time zone, `time_zone` should be used.
Example usage has been shown on the following snippet:

```yaml
extensions:
  sumologic:
    access_id: <access_id>
    access_key: <access_key>
    collector_name: my_collector
    collector_description: This is my and only my collector
    collector_category: example
    time_zone: America/Tijuana
processors:
  resource:
    attributes:
    - key: author
      value: me
      action: insert
exporters:
  sumologic:
    source_host: My hostname
```

### Advanced

#### CPU Target

CPU Target is not supported by Opentelemetry Collector.

#### Collector Management

Currently Opentelemetry Collectot is local file managed.
Depending on your setup please follow [Cloud Based Management](#cloud-based-management)
or [Local Configuration File](#local-configuration-file) migration details.

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
    ## the following section in filelog documentation is recommended to read:
    ## https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.31.0/receiver/filelogreceiver#multiline-configuration
    multiline:
      ## line_start_pattern is substitute of `Boundary Regex`.
      line_start_pattern: ^\d{4}
processors:
  ## The following configuration will add two fields to every record
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
        from: $$body.log
        to: $$body
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

### Remote File Source

Remote File Source is not supported by Opentelemetry Collector.

### Syslog Source

#### Overall example

Below is an example of an OpenTelemetry configuration for a Syslog Source.

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
  ## Use tcpreceiver for TCP protocol
  tcpreceiver/first receiver:
    ## listen address in format host:port
    ## host 0.0.0.0 mean all network interfaces
    listen_address: 0.0.0.0:514
    ## Add network attributes
    ## `net.peer.name` is going to be used as exporters.sumologic.source_host
    ## rel: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.31.0/receiver/tcplogreceiver#configuration
    add_attributes: true
  ## Use udpreceiver for UDP protocol
  udpreceiver/first receiver:
    ## listen address in format host:port
    ## host 0.0.0.0 mean all network interfaces
    listen_address: 0.0.0.0:514
    ## Add network attributes
    ## `net.peer.name` is going to be used as exporters.sumologic.source_host
    ## rel: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.31.0/receiver/udplogreceiver#configuration
    add_attributes: true

processors:
  ## There is no substitute for `Description` in current project phase.
  ## It is recommended to use comments for that purpose, like this one.
  ## sumologicsyslog/<source group name>:
  ## <source group name> can be substitute of Installed Collector `Name`.
  sumologicsyslog/syslog source:

  ## The following configuration will add two fields to every record
  resource/syslog source:
    attributes:
    - key: cloud.availability_zone
      value: zone-1
      action: insert
    - key: k8s.cluster.name
      value: my-cluster
      action: insert
exporters:
  sumologic/syslog:
    ## Installed Collector substitute for `Source Category`.
    source_category: example category
    ## clear_logs_timestamp is by default set to True.
    ## If it's set to true, it works like `Enable Timestamp Parsing`,
    ## and `Time Zone` is going to be taken from `extensions` section.
    ## There is no possibility to configure several time zones in one exporter.
    ## clear_logs_timestamp sets to true also behaves like
    ## `Timestamp Format` would be set to `Automatically detect the format`
    ## in terms of Installed Collector configuration.
    clear_logs_timestamp: true
    ## Set Source Name to be facility name
    source_name: "%{facility}"
    ## Set Source Host to `net.peer.name`
    source_host: "%{net.peer.name}
service:
  extensions:
  - sumologic
  pipelines:
    logs/syslog source:
      receivers:
      - filelog/syslog source
      processors:
      - resource/syslog source
      exporters:
      - sumologic/syslog
```

#### Name

Define the name after the slash `/` in the processor name.

For example, the following snippet configures the name as `my example name`:

```yaml
processor:
  sumologicsyslog/my example name:
  # ...
```

#### Description

A description can be added as a comment just above the processor name.

For example, the following snippet configures the description as `All my example logs`:

```yaml
processor:
  ## All my example logs
  sumologicsyslog/my example name:
  # ...
```

#### Protocol and Port

Protocol is defined by receiver type. For UDP use [udplogreceiver][udplogreceiver] and for TCP use [tcplogreceiver][tcplogreceiver]. Port can be set by `listen_address`, for example to listen on port `6776` on all interfaces, use `listen_address: 0.0.0.0:6776`.

You can use multiple receivers with different names nad ports like in the following example:

```yaml
receivers:
  tcpreceiver/first receiver:
    listen_address: 0.0.0.0:514
  tcpreceiver/second receiver:
    listen_address: 127.0.0.1:5140
  udpreceiver/first receiver:
    listen_address: 0.0.0.0:514
  udpreceiver/second receiver:
    listen_address: 127.0.0.1:5150
processor:
  ## All my example logs
  sumologicsyslog/my example name:
  # ...
```

#### Source Category

A Source Category can set in the exporter configuration with the `source_category` option.

For example, the following snippet configures the Source Category as `My Category`:

```yaml
receivers:
  tcpreceiver/first receiver:
    listen_address: 0.0.0.0:514
  tcpreceiver/second receiver:
    listen_address: 127.0.0.1:5140
  udpreceiver/first receiver:
    listen_address: 0.0.0.0:514
  udpreceiver/second receiver:
    listen_address: 127.0.0.1:5150
processor:
  ## All my example logs
  sumologicsyslog/my example name:
  # ...
exporters:
  sumologic/some name:
    source_category: My Category
```

#### Fields

Use the [resourceprocessor][resourceprocessor] to set custom fields.

For example, the following snippet configures two fields, `cloud.availability_zone` and `k8s.cluster.name`:

```yaml
receivers:
  tcpreceiver/first receiver:
    listen_address: 0.0.0.0:514
  tcpreceiver/second receiver:
    listen_address: 127.0.0.1:5140
  udpreceiver/first receiver:
    listen_address: 0.0.0.0:514
  udpreceiver/second receiver:
    listen_address: 127.0.0.1:5150
processors:
  ## All my example logs
  sumologicsyslog/my example name:
  # ...
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
    source_category: My Category
```

#### Advanced Options for Logs

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
  tcpreceiver/first receiver:
    listen_address: 0.0.0.0:514
  tcpreceiver/second receiver:
    listen_address: 127.0.0.1:5140
  udpreceiver/first receiver:
    listen_address: 0.0.0.0:514
  udpreceiver/second receiver:
    listen_address: 127.0.0.1:5150
processors:
  ## All my example logs
  sumologicsyslog/my example name:
  # ...
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
    source_category: My Category
```

If `clear_logs_timestamp` is set to `false`, timestamp parsing should be configured
manually, like in the following snippet:

```yaml
extensions:
  sumologic:
    time_zone: America/Tijuana
receivers:
  tcpreceiver/first receiver:
    listen_address: 0.0.0.0:514
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
        from: $$body.log
        to: $$body
processors:
  ## All my example logs
  sumologicsyslog/my example name:
  # ...
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
  tcpreceiver/first receiver:
    listen_address: 0.0.0.0:514
processors:
  ## All my example logs
  sumologicsyslog/my example name:
  # ...
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
    source_category: My Category
    ## Keep manually parsed timestamps
    clear_logs_timestamp: true
```

#### Additional Configuration

##### Source Name

In Opentelemetry Collector Source Name has to be set manually.
[Sumologicsyslogprocessor][sumologicsyslog] sets `facility` attribute,
which should be set as source name in exporter configuration.

See following example for better understanding:

```yaml
extensions:
  sumologic:
    time_zone: America/Tijuana
receivers:
  tcpreceiver/first receiver:
    listen_address: 0.0.0.0:514
processors:
  ## All my example logs
  sumologicsyslog/my example name:
  # ...
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
    source_category: My Category
    ## Keep manually parsed timestamps
    clear_logs_timestamp: true
    ## Set Source Name to facility, which is set by sumologicsyslogprocessor
    source_name: "%{facility}
```

##### Source Host

In Opentelemetry Collector Source Name has to be set manually.
To do it `add_attributes` should be set to `true`
for [tcplogreceiver][tcplogreceiver]/[udplogreceiver][udplogreceiver].
This is going to add [connection related attributes][network-semantic-convention],
especially `net.peer.name` which should be set as Source Host.

Please see the following configuration for better understanding:

```yaml
extensions:
  sumologic:
    time_zone: America/Tijuana
receivers:
  tcpreceiver/first receiver:
    listen_address: 0.0.0.0:514
    add_attributes: true
processors:
  ## All my example logs
  sumologicsyslog/my example name:
  # ...
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
    source_category: My Category
    ## Keep manually parsed timestamps
    clear_logs_timestamp: true
    ## Set Source Name to facility, which is set by sumologicsyslogprocessor
    source_name: "%{facility}
    source_host: "%{net.peer.name}
```

### Docker Logs Source

Docker Logs Source is not supported by Opentelemetry Collector.

### Docker Stats Source

Docker Stats Source is not supported by Opentelemetry Collector.

### Script Source

Script Source is not supported by Opentelemetry Collector.

### Streaming Metrics Source

Streaming Metrics Source is not supported by Opentelemetry Collector.

### Host Metrics Source

Host Metrics Source is not supported by Opentelemetry Collector.

### Local Windows Event Log Source

Local Windows Event Log Source is not supported by Opentelemetry Collector.

### Local Windows Performance Monitor Log Source

Local Windows Performance Monitor Log Source is not supported by Opentelemetry Collector.

### Windows Active Directory Source

Windows Active Directory Source is not supported by Opentelemetry Collector.

### Script Action

Script Action is not supported by Opentelemetry Collector.

## Local Configuration File

This section describes migration steps for Sources managed locally.

### Local File Source

Local File Source is not supported by Opentelemetry Collector.

### Remote File Source

Remote File Source is not supported by Opentelemetry Collector.

### Syslog Source

Remote File Source is not supported by Opentelemetry Collector.

### Docker Logs Source

Docker Logs Source is not supported by Opentelemetry Collector.

### Docker Stats Source

Docker Stats Source is not supported by Opentelemetry Collector.

### Script Source

Script Source is not supported by Opentelemetry Collector.

### Streaming Metrics Source

Streaming Metrics Source is not supported by Opentelemetry Collector.

### Host Metrics Source

Host Metrics Source is not supported by Opentelemetry Collector.

### Local Windows Event Log Source

Local Windows Event Log Source is not supported by Opentelemetry Collector.

### Local Windows Performance Monitor Log Source

Local Windows Performance Monitor Log Source is not supported by Opentelemetry Collector.

### Windows Active Directory Source

Windows Active Directory Source is not supported by Opentelemetry Collector.

### Script Action

Script Action is not supported by Opentelemetry Collector.

[resourceprocessor]: https://github.com/open-telemetry/opentelemetry-collector/tree/v0.31.0/processor/resourceprocessor
[multiline]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.31.0/receiver/filelogreceiver#multiline-configuration
[supported_encodings]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.31.0/receiver/filelogreceiver#supported-encodings
[udplogreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.31.0/receiver/udplogreceiver
[tcplogreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.31.0/receiver/tcplogreceiver
[sumologicsyslog]: https://github.com/SumoLogic/sumologic-otel-collector/tree/v0.0.19-beta.0/pkg/processor/sumologicsyslogprocessor
[network-semantic-convention]: https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/trace/semantic_conventions/span-general.md#general-network-connection-attributes
[sumologicextension]: https://github.com/SumoLogic/sumologic-otel-collector/tree/v0.0.19-beta.0/pkg/extension/sumologicextension