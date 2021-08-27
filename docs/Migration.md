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
  - [Collector](#collector-1)
    - [user.properties](#user.properties)
  - [Common Parameters](#common-parameters)
  - [Local File Source (LocalFile)](#local-file-source-localfile)
  - [Remote File Source (RemoteFileV2)](#remote-file-source-remotefilev2)
  - [Syslog Source (Syslog)](#syslog-source-syslog)
  - [Docker Logs Source (DockerLog)](#docker-logs-source-dockerlog)
  - [Docker Stats Source (DockerStats)](#docker-stats-source-dockerstats)
  - [Script Source (Script)](#script-source-script)
  - [Streaming Metrics Source (StreamingMetrics)](#streaming-metrics-source-streamingmetrics)
  - [Host Metrics Source (SystemStats)](#host-metrics-source-systemstats)
  - [Local Windows Event Log Source (LocalWindowsEventLog)](#local-windows-event-log-source-localwindowseventlog)
  - [Remote Windows Event Log Source (RemoteWindowsEventLog)](#local-windows-event-log-source-remotewindowseventlog)
  - [Local Windows Performance Source (LocalWindowsPerfMon)](#local-windows-performance-monitor-log-source-localwindowsperfmon)
  - [Remote Windows Performance Source (RemoteWindowsPerfMon)](#local-windows-performance-monitor-log-source-remotewindowsperfmon)
  - [Windows Active Directory Source (ActiveDirectory)](#windows-active-directory-source-activedirectory)

## General Configuration Concepts

Let's consider the following example:

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

To use those configured modules, they need to be mentioned in the `service` section.
`service` consists of `extensions` (they are global across collector) and `pipelines`.
`Pipelines` can be `logs`, `metrics`, and `traces` and each of them can have
`receivers`, `processors` and `exporters`. Multiple pipelines of one type can be configured using aliases,
such as `example pipeline` for `logs` in the example above.

## Collector

Collector registration and configuration is handled by the [sumologicextension][sumologicextension].

### Name

Collector name can be specified by setting the `collector_name` option:

```yaml
extensions:
  sumologic:
    access_id: <access_id>
    access_key: <access_key>
    collector_name: my_collector
```

### Description

To set a description, use the `collector_description` option:

```yaml
extensions:
  sumologic:
    access_id: <access_id>
    access_key: <access_key>
    collector_name: my_collector
    collector_description: This is my and only my collector
```

### Host Name

Host name can be set in the sumologic exporter configuration.
The exporter will set the host name for every record sent to Sumo Logic:

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

To set a Collector category, use the `collector_category` option:

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

Fields in the Opentelemetry Collector can be added with the [resourceprocessor][resourceprocessor].
For example, to add a field with the key `author` with the value `me` to every record,
you could use the following configuration:

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

### Assign to an Ingest Budget

Assignment to an Ingest Budget is not supported by Opentelemetry Collector.

### Time Zone

To set the Collector time zone, use the `time_zone` option.
For example, the following examples sets the time zone to `America/Tijuana`:

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

CPU Target is not supported by the Opentelemetry Collector.

#### Collector Management

Currently, the Opentelemetry Collector can only be managed with Local Configuration File Management.
Depending on your setup, follow the steps in [Cloud Based Management](#cloud-based-management)
or [Local Configuration File](#local-configuration-file) for migration details.

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
    ## https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.33.0/receiver/filelogreceiver
    encoding: utf-8
    ## multiline is Opentelemetry Collector substitute for `Enable Multiline Processing`.
    ## As multiline detection behaves slightly different than in Installed Collector
    ## the following section in filelog documentation is recommended to read:
    ## https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.33.0/receiver/filelogreceiver#multiline-configuration
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

Remote File Source is not supported by the Opentelemetry Collector.

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
    ## rel: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.33.0/receiver/tcplogreceiver#configuration
    add_attributes: true
  ## Use udpreceiver for UDP protocol
  udpreceiver/first receiver:
    ## listen address in format host:port
    ## host 0.0.0.0 mean all network interfaces
    listen_address: 0.0.0.0:514
    ## Add network attributes
    ## `net.peer.name` is going to be used as exporters.sumologic.source_host
    ## rel: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.33.0/receiver/udplogreceiver#configuration
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

The OpenTelemetry Collector requires the Source Name to be set manually.
In the exporter configuration, use the [Sumologicsyslogprocessor][sumologicsyslog]
to set the `facility` attribute.

For example:

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

The OpenTelemetry Collector requires the Source Host to be set manually.
Set `add_attributes` to `true` for [tcplogreceiver][tcplogreceiver]/[udplogreceiver][udplogreceiver].
This adds [connection related attributes][network-semantic-convention],
especially `net.peer.name` which should be set as the Source Host.

For example:

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

Docker Logs Source is not supported by the OpenTelemetry Collector.

### Docker Stats Source

Docker Stats Source is not supported by the OpenTelemetry Collector.

### Script Source

Script Source is not supported by the OpenTelemetry Collector.

### Streaming Metrics Source

Streaming Metrics Source is not supported by the OpenTelemetry Collector.

### Host Metrics Source

Host Metrics Source is not supported by the OpenTelemetry Collector.

### Local Windows Event Log Source

Local Windows Event Log Source is not supported by the OpenTelemetry Collector.

### Local Windows Performance Monitor Log Source

Local Windows Performance Monitor Log Source is not supported by the OpenTelemetry Collector.

### Windows Active Directory Source

Windows Active Directory Source is not supported by the OpenTelemetry Collector.

### Script Action

Script Action is not supported by the OpenTelemetry Collector.

## Local Configuration File

This section describes migration steps for an Installed Collector managed with a Local Configuration File.

### Collector

#### user.properties

The following table shows the equivalent [user.properties][user.properties] for OpenTelemetry.

| user.properties key                           | The OpenTelemetry Collector Key                           |
|-----------------------------------------------|------------------------------------------------------------|
| `wrapper.java.command=JRE Bin Location`       | N/A                                                        |
| `accessid=accessId`                           | `extensions.sumologic.access_id`                           |
| `accesskey=accessKey`                         | `extensions.sumologic.access_key`                          |
| `category=category`                           | [extensions.sumologic.collector_category](#category)       |
| `clobber=true/false`                          | `extensions.sumologic.clobber`                             |
| `description=description`                     | [extensions.sumologic.collector_description](#description) |
| `disableActionSource=true/false`              | N/A                                                        |
| `disableScriptSource=true/false`              | N/A                                                        |
| `disableUpgrade=true/false`                   | N/A                                                        |
| `enableActionSource=true/false`               | N/A                                                        |
| `enableScriptSource=true/false`               | N/A                                                        |
| `ephemeral=true/false`                        | N/A                                                        |
| `fields=[list of fields]`                     | [processors.resource](#fields)                             |
| `fipsJce=true/false`                          | N/A                                                        |
| `hostName=hostname`                           | `exporters.sumologic.source_host`                          |
| `name=name`                                   | [extensions.sumologic.collector_name](#name)               |
| `proxyHost=host`                              | [plese see OTC documentation][proxy]                       |
| `proxyNtlmDomain=NTLM domain`                 | [plese see OTC documentation][proxy]                       |
| `proxyPassword=password`                      | [plese see OTC documentation][proxy]                       |
| `proxyPort=port`                              | [plese see OTC documentation][proxy]                       |
| `proxyUser=username`                          | [plese see OTC documentation][proxy]                       |
| `skipAccessKeyRemoval=true/false`             | N/A                                                        |
| `sources=absolute filepath or folderpath`     | N/A                                                        |
| `syncSources=absolute filepath or folderpath` | N/A                                                        |
| `targetCPU=target`                            | N/A                                                        |
| `timeZone=timezone`                           | [extensions.sumologic.time_zone](#time-zone)               |
| `token=token`                                 | N/A                                                        |
| `url=collection endpoint`                     | `extensions.sumologic.api.base.url`                        |
| `wrapper.java.command=JRE Bin Location`       | N/A                                                        |
| `wrapper.java.command=JRE Bin Location`       | N/A                                                        |
| `wrapper.java.maxmemory=size`                 | N/A                                                        |

### Common Parameters

This section describes migration steps for [common parameters][common-parameters].

`sourceType` migration:

- [LocalFile](#local-file-source-localfile)
- [RemoteFileV2](#remote-file-source-remotefilev2)
- [Syslog](#syslog-source-syslog)
- [DockerLog](#docker-logs-source-dockerlog)
- [DockerStats](#docker-stats-source-dockerstats)
- [Script](#script-source-script)
- [StreamingMetrics](#streaming-metrics-source-streamingmetrics)
- [SystemStats](#host-metrics-source-systemstats)
- [LocalWindowsEventLog](#local-windows-event-log-source-localwindowseventlog)
- [RemoteWindowsEventLog](#local-windows-event-log-source-remotewindowseventlog)
- [LocalWindowsPerfMon](#local-windows-performance-monitor-log-source-localwindowsperfmon)
- [RemoteWindowsPerfMon](#local-windows-performance-monitor-log-source-remotewindowsperfmon)
- [ActiveDirectory](#windows-active-directory-source-activedirectory)

| The Installed Collector Parameter | The OpenTelemetry Collector Key                                                                                |
|-----------------------------------|-----------------------------------------------------------------------------------------------------------------|
| `name`                            | Define the name after the slash `/` in the receiver name. [See the linked example.](#name-1)                    |
| `description`                     | A description can be added as a comment just above the receiver name. [See the linked example.](#description-1) |
| `fields`                          | Use the [resourceprocessor][resourceprocessor] to set custom fields. [See the linked example.](#fields-1)       |
| `hostName`                        | [exporters.sumologic.source_host][source-templates]; [See the linked example.](#host-name-1)                     |
| `category`                        | [exporters.sumologic.source_category][source-templates]                                                         |
| `automaticDateParsing`            | [See Timestamp Parsing explanation](#timestamp-parsing-1)                                                       |
| `timeZone`                        | [See Timestamp Parsing explanation](#timestamp-parsing-1)                                                       |
| `forceTimeZone`                   | [See Timestamp Parsing explanation](#timestamp-parsing-1)                                                       |
| `defaultDateFormat`               | [See Timestamp Parsing explanation](#timestamp-parsing-1)                                                       |
| `defaultDateFormats`              | [See Timestamp Parsing explanation](#timestamp-parsing-1)                                                       |
| `multilineProcessingEnabled`      | [See Multiline Processing explanation](#multiline-processing)                                                   |
| `useAutolineMatching`             | [See Multiline Processing explanation](#multiline-processing)                                                   |
| `manualPrefixRegexp`              | [See Multiline Processing explanation](#multiline-processing)                                                   |
| `filters`                         | N/A                                                                                                             |
| `cutoffTimestamp`                 | N/A                                                                                                             |
| `cutoffRelativeTime`              | N/A                                                                                                             |

### Local File Source (LocalFile)

Local File Source is not supported by the OpenTelemetry Collector.

### Remote File Source (RemoteFileV2)

Remote File Source is not supported by the OpenTelemetry Collector.

### Syslog Source (Syslog)

Remote File Source is not supported by the OpenTelemetry Collector.

### Docker Logs Source (DockerLog)

Docker Logs Source is not supported by the OpenTelemetry Collector.

### Docker Stats Source (DockerStats)

Docker Stats Source is not supported by the OpenTelemetry Collector.

### Script Source (Script)

Script Source is not supported by the OpenTelemetry Collector.

### Streaming Metrics Source (StreamingMetrics)

Streaming Metrics Source is not supported by the OpenTelemetry Collector.

### Host Metrics Source (SystemStats)

Host Metrics Source is not supported by the OpenTelemetry Collector.

### Local Windows Event Log Source (LocalWindowsEventLog)

Local Windows Event Log Source is not supported by the OpenTelemetry Collector.

### Remote Windows Event Log Source (RemoteWindowsEventLog)

Remote Windows Event Log Source is not supported by the OpenTelemetry Collector.

### Local Windows Performance Source (LocalWindowsPerfMon)

Local Windows Performance Source is not supported by the OpenTelemetry Collector.

### Remote Windows Performance Source (RemoteWindowsPerfMon)

Remote Windows Performance Source is not supported by the OpenTelemetry Collector.

### Windows Active Directory Source (ActiveDirectory)

Windows Active Directory Source is not supported by the OpenTelemetry Collector.

[resourceprocessor]: https://github.com/open-telemetry/opentelemetry-collector/tree/v0.33.0/processor/resourceprocessor
[multiline]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.33.0/receiver/filelogreceiver#multiline-configuration
[supported_encodings]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.33.0/receiver/filelogreceiver#supported-encodings
[udplogreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.33.0/receiver/udplogreceiver
[tcplogreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.33.0/receiver/tcplogreceiver
[sumologicsyslog]: ../pkg/processor/sumologicsyslogprocessor/README.md
[network-semantic-convention]: https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/trace/semantic_conventions/span-general.md#general-network-connection-attributes
[sumologicextension]: ../pkg/extension/sumologicextension/README.md
[user.properties]: https://help.sumologic.com/03Send-Data/Installed-Collectors/05Reference-Information-for-Collector-Installation/06user.properties
[proxy]: https://opentelemetry.io/docs/collector/configuration/#proxy-support
[common-parameters]: https://help.sumologic.com/03Send-Data/Sources/03Use-JSON-to-Configure-Sources#common-parameters-for-log-source-types
[source-templates]: ../pkg/exporter/sumologicexporter/README.md#source-templates
