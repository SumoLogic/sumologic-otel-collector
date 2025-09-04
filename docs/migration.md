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
  - [Assign to an Ingest Budget](#assign-to-an-ingest-budget)
  - [Time Zone](#time-zone)
  - [Advanced](#advanced)
    - [CPU Target](#cpu-target)
    - [Collector Management](#collector-management)
- [Cloud Based Management](#cloud-based-management)
  - [Common configuration](#common-configuration)
    - [Name](#name-1)
    - [Description](#description-1)
    - [Source Category](#source-category)
    - [Fields/Metadata](#fieldsmetadata)
    - [Source Host](#source-host)
  - [Local File Source](#local-file-source)
    - [Overall example](#overall-example)
    - [Name](#name-2)
    - [Description](#description-2)
    - [File Path](#file-path)
      - [Collection should begin](#collection-should-begin)
    - [Source Host](#source-host-1)
    - [Source Category](#source-category-1)
    - [Fields](#fields-1)
    - [Advanced Options for Logs](#advanced-options-for-logs)
      - [Denylist](#denylist)
      - [Timestamp Parsing](#timestamp-parsing)
      - [Encoding](#encoding)
      - [Multiline Processing](#multiline-processing)
  - [Remote File Source](#remote-file-source)
  - [Syslog Source](#syslog-source)
    - [Overall example](#overall-example-1)
    - [Name](#name-3)
    - [Description](#description-3)
    - [Protocol and Port](#protocol-and-port)
    - [Source Category](#source-category-2)
    - [Fields](#fields-2)
    - [Advanced Options for Logs](#advanced-options-for-logs-1)
      - [Timestamp Parsing](#timestamp-parsing-1)
    - [Additional Configuration](#additional-configuration)
      - [Source Name](#source-name)
      - [Source Host](#source-host-2)
  - [Docker Logs Source](#docker-logs-source)
  - [Docker Stats Source](#docker-stats-source)
    - [Overall example](#overall-example-2)
    - [Name](#name-4)
    - [Description](#description-4)
    - [URI](#uri)
    - [Container filters](#container-filters)
    - [Source Host](#source-host-3)
    - [Source Category](#source-category-3)
    - [Fields](#fields-3)
    - [Scan interval](#scan-interval)
    - [Metrics](#metrics)
    - [Metadata](#metadata)
  - [Script Source](#script-source)
  - [Streaming Metrics Source](#streaming-metrics-source)
    - [Overall example](#overall-example-3)
    - [Name](#name-5)
    - [Description](#description-5)
    - [Protocol and Port](#protocol-and-port-1)
    - [Content Type](#content-type)
    - [Source Category](#source-category-4)
    - [Metadata](#metadata-1)
  - [Host Metrics Source](#host-metrics-source)
    - [Using Telegraf Receiver](#using-telegraf-receiver)
      - [Overall Example](#overall-example-4)
      - [Name](#name-6)
      - [Description](#description-6)
      - [Source Host](#source-host-4)
      - [Source Category](#source-category-5)
      - [Metadata](#metadata-2)
      - [Scan Interval](#scan-interval-1)
      - [Metrics](#metrics-1)
        - [CPU](#cpu)
        - [Memory](#memory)
        - [TCP](#tcp)
        - [Network](#network)
        - [Disk](#disk)
  - [Local Windows Event Log Source](#local-windows-event-log-source)
  - [Local Windows Performance Monitor Log Source](#local-windows-performance-monitor-log-source)
  - [Windows Active Directory Source](#windows-active-directory-source)
  - [Script Action](#script-action)
- [Local Configuration File](#local-configuration-file)
  - [Collector](#collector-1)
    - [user.properties](#userproperties)
  - [Common Parameters](#common-parameters)
    - [Filtering](#filtering)
      - [Include and exclude](#include-and-exclude)
      - [Masks](#masks)
      - [Data forwarding](#data-forwarding)
  - [Local File Source (LocalFile)](#local-file-source-localfile)
  - [Remote File Source (RemoteFileV2)](#remote-file-source-remotefilev2)
  - [Syslog Source (Syslog)](#syslog-source-syslog)
  - [Docker Logs Source (DockerLog)](#docker-logs-source-dockerlog)
  - [Docker Stats Source (DockerStats)](#docker-stats-source-dockerstats)
  - [Script Source (Script)](#script-source-script)
  - [Streaming Metrics Source (StreamingMetrics)](#streaming-metrics-source-streamingmetrics)
  - [Host Metrics Source (SystemStats)](#host-metrics-source-systemstats)
  - [Local Windows Event Log Source (LocalWindowsEventLog)](#local-windows-event-log-source-localwindowseventlog)
  - [Remote Windows Event Log Source (RemoteWindowsEventLog)](#remote-windows-event-log-source-remotewindowseventlog)
  - [Local Windows Performance Source (LocalWindowsPerfMon)](#local-windows-performance-source-localwindowsperfmon)
  - [Remote Windows Performance Source (RemoteWindowsPerfMon)](#remote-windows-performance-source-remotewindowsperfmon)
  - [Windows Active Directory Source (ActiveDirectory)](#windows-active-directory-source-activedirectory)

## General Configuration Concepts

Let's consider the following example:

```yaml
extensions:
  sumologic:
    installation_token: <installation_token>

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
    installation_token: <installation_token>
    collector_name: my_collector
```

### Description

To set a description, use the `collector_description` option:

```yaml
extensions:
  sumologic:
    installation_token: <installation_token>
    collector_name: my_collector
    collector_description: This is my and only my collector
```

### Host Name

Host name can be set in the [Source Processor][source-templates] configuration.
The processor will set the host name for every record sent to Sumo Logic:

```yaml
extensions:
  sumologic:
    installation_token: <installation_token>
    collector_name: my_collector
    collector_description: This is my and only my collector
processors:
  source:
    source_host: My hostname
```

### Category

To set a Collector category, use the `collector_category` option:

```yaml
extensions:
  sumologic:
    installation_token: <installation_token>
    collector_name: my_collector
    collector_description: This is my and only my collector
    collector_category: example
processors:
  source:
    source_host: My hostname
```

### Fields

Fields in the Opentelemetry Collector can be added with the `collector_fields` property of [Sumo Logic Extension][sumologicextension].

For example, to add a field with the key `author` with the value `me` to every record,
you could use the following configuration:

```yaml
extensions:
  sumologic:
    installation_token: <installation_token>
    collector_name: my_collector
    collector_description: This is my and only my collector
    collector_category: example
    collector_fields:
      author: me

processors:
  source:
    source_host: My hostname
```

### Assign to an Ingest Budget

Assignment to an Ingest Budget can be done using [ingest-budget]

### Time Zone

To set the Collector time zone, use the `time_zone` option.
For example, the following examples sets the time zone to `America/Tijuana`:

```yaml
extensions:
  sumologic:
    installation_token: <installation_token>
    collector_name: my_collector
    collector_description: This is my and only my collector
    collector_category: example
    time_zone: America/Tijuana
    collector_fields:
      author: me
processors:
  source:
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

### Common configuration

There is set of configuration common for all of the sources and this section is going to cover details of the migration for them.

#### Name

Define the name after the slash `/` in the component name.

To set `_sourceName`, use [resourceprocessor][resourceprocessor]
or set it in [sumologicexporter][sumologicexporter].

For example, the following snippet configures the name for [Filelog Receiver][filelogreceiver] instance as `my example name`:

```yaml
receivers:
  filelog/my example name:
  # ...
```

#### Description

A description can be added as a comment just above the receiver name.

For example, the following snippet configures the description for [Filelog Receiver][filelogreceiver] instance as `All my example logs`:

```yaml
receivers:
  ## All my example logs
  filelog/my example name:
  # ...
```

#### Source Category

The Source Category is set in the [Source Processor][source-templates] configuration with the `source_category` option.

For example, the following snippet configures the Source Category as `My Category`:

```yaml
processors:
  source/some name:
    source_category: My Category
```

#### Fields/Metadata

There are multiple ways to set fields/metadata in OpenTelemetry Collector

- For fields/metadata which are going to be the same for all sources, please refer to [Collector Fields](#fields)

- For fields/metadata, which should be set for specific source, you should use [Transform Processor][transformprocessor].

  For example, the following snippet configures two fields for logs, `cloud.availability_zone` and `k8s.cluster.name`:

  ```yaml
    transform/custom fields:
      log_statements:
      - context: resource
        statements:
        - set(attributes["cloud.availability_zone"], "zone-1")
        - set(attributes["k8s.cluster.name"], "my-cluster")
  ```

  The following example adds two metadata for metrics, `cloud.availability_zone` and `k8s.cluster.name`:

  ```yaml
    transform/custom fields:
      metric_statements:
      - context: resource
        statements:
        - set(attributes["cloud.availability_zone"], "zone-1")
        - set(attributes["k8s.cluster.name"], "my-cluster")
  ```

- As an alternative, you can  use [resourceprocessor][resourceprocessor] to set custom fields/metadata for source:

  For example, the following snippet configures two fields, `cloud.availability_zone` and `k8s.cluster.name`:

  ```yaml
  processors:
    resource/my example name fields:
      attributes:
      - key: cloud.availability_zone
        value: zone-1
        ## upsert will override existing cloud.availability_zone field
        action: upsert
      - key: k8s.cluster.name
        value: my-cluster
        ## insert will add cloud.availability_zone field if it doesn't exist
        action: insert
  ```

#### Source Host

A Source Host can be set in the [Source Processor][source-templates] configuration with the `source_host` option.

For example, the following snippet configures the Source Host as `my_host`:

```yaml
processors:
  source/some name:
    source_host: my_host
```

### Local File Source

Local File Source functionality is covered by OpenTelemetry [Filelog Receiver][filelogreceiver].

#### Overall example

Below is an example of an OpenTelemetry configuration for a Local File Source.

```yaml
extensions:
  sumologic:
    installation_token: <installation_token>
    ## Time Zone is a substitute of Installed Collector `Time Zone`
    ## with `Use time zone from log file. If none is detected use:` option.
    ## This is used only if log timestamp is set to 0 by transform processor.
    ## Full list of time zones is available on wikipedia:
    ## https://en.wikipedia.org/wiki/List_of_tz_database_time_zones#List
    time_zone: America/Tijuana
    ## The following configuration will add two fields to every record
    collector_fields:
      cloud.availability_zone: zone-1
      k8s.cluster.name: my-cluster
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
    - tmp/logs.log"
    ## List of local files which shouldn't be read.
    ## Installed Collector `Denylist` substitute.
    exclude:
    - /var/log/auth.log
    - /opt/app/logs/security_*.log
    ## This config can take one of two values: `beginning` or `end`.
    ## If you are looking for `Collection should begin`, please look at `Collection should begin` section in this document
    start_at: beginning
    ## encoding is substitute for Installed Collector `Encoding`.
    ## List of supported encodings:
    ## https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.134.0/receiver/filelogreceiver
    encoding: utf-8
    ## multiline is Opentelemetry Collector substitute for `Enable Multiline Processing`.
    ## As multiline detection behaves slightly different than in Installed Collector
    ## the following section in filelog documentation is recommended to read:
    ## https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.134.0/receiver/filelogreceiver#multiline-configuration
    multiline:
      ## line_start_pattern is substitute of `Boundary Regex`.
      line_start_pattern: ^\d{4}
    ## Adds file path log.file.path attribute, which can be used for timestamp parsing
    ## See operators configuration
    include_file_path: true
    ## `Operators allows to perform more advanced operations like per file timestamp parsing
    operators:
    - type: regex_parser
      ## Applies only to tmp/logs.log file
      if: 'attributes["log.file.path"] == "tmp/logs.log"'
      ## Extracts timestamp to timestamp_field using regex parser
      regex: '^(?P<timestamp_field>\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}).*'
      timestamp:
        parse_from: attributes.timestamp_field
        layout: '%Y-%m-%d %H:%M:%S'
    ## Cleanup timestamp_field
    - type: remove
      field: attributes.timestamp_field
processors:
  source:
    ## Set _sourceName
    source_name: my example name
    ## Installed Collector substitute for `Source Category`.
    source_category: example category
    ## Installed Collector substitute for `Source Host`.
    source_host: example host
  transform/logs source:
    log_statements:
      ## By default every data has timestamp (usually set to receipt time)
      ## and therefore Sumo Logic backend do not try to parse it from log body.
      ## Using this processor works like `Enable Timestamp Parsing`,
      ## where `Time Zone` is going to be taken from `extensions` section.
      ## There is no possibility to configure several time zones in one exporter.
      ## It behaves like `Timestamp Format` would be set to `Automatically detect the format`
      ## in terms of Installed Collector configuration.
      - context: log
        statements:
          - set(time_unix_nano, 0)
      ## Adds custom fields:
      ## - cloud.availability_zone=zone-1
      ## - k8s.cluster.name=my-cluster
      - context: resource
        statements:
        - set(attributes["cloud.availability_zone"], "zone-1")
        - set(attributes["k8s.cluster.name"], "my-cluster")
  ## Remove logs with timestamp before Sat Dec 31 2022 23:00:00 GMT+0000
  ## This configuration covers `Collection should begin` functionality.
  ## Please ensure that timestamps are correctly set (eg. use operators in filelog receiver)
  filter/remove older:
    logs:
      log_record:
      ## - 1672527600000000000 ns is equal to Dec 31 2022 23:00:00 GMT+0000
      ## - do not remove logs which do not have correct timestamp
      - 'time_unix_nano < 1672527600000000000 and time_unix_nano > 0'
exporters:
  sumologic:
service:
  extensions:
  - sumologic
  pipelines:
    logs/log source:
      receivers:
      - filelog/log source
      processors:
      - filter/remove older
      - transform/logs source
      - source
      exporters:
      - sumologic
```

#### Name

Please refer to [the Name section of Common configuration](#name-1).

#### Description

Please refer to [the Description section of Common configuration](#description-1).

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

The OpenTelemetry Collector substitution for this Installed Collector option requires manual timestamp parsing.
Then you can use [Filter Processor][filterprocessor] to filter out logs before a specific date.

Let's consider the following example. We want to get logs from `tmp/logs.log` which are at least from `Dec 31 2022 23:00:00`

- `tmp/logs.log`

  ```text
  2020-04-01 10:12:14 Log from 2020
  2021-01-02 12:13:54 Log from 2021
  2022-03-07 11:15:29 Log from 2022
  2023-01-02 10:37:12 Log from 2023
  ```

- `config.yaml` (only essential parts)

  ```yaml
  receivers:
    filelog/log source:
      include:
      - tmp/logs.log
      # - ...
      ## Adds file path log.file.path attribute, which will be used further in pipeline
      include_file_path: true
      ## We would like to read from beginning, as we can choose only between end and beginning
      start_at: beginning
      operators:
      - type: regex_parser
        ## Applies only to tmp/logs.log file
        if: 'attributes["log.file.path"] == "tmp/logs.log"'
        ## Extracts timestamp to timestamp_field using regex parser
        regex: '^(?P<timestamp_field>\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}).*'
        timestamp:
          parse_from: attributes.timestamp_field
          layout: '%Y-%m-%d %H:%M:%S'
      ## cleanup timestamp_field
      - type: remove
        field: attributes.timestamp_field
      # ...
  processors:
    ## Remove logs with timestamp before Sat Dec 31 2022 23:00:00 GMT+0000
    filter/remove older:
      logs:
        log_record:
        ## 1672527600000000000 ns is equal to Dec 31 2022 23:00:00 GMT+0000,
        ## but do not remove logs which do not have correct timestamp
        - 'time_unix_nano < 1672527600000000000 and time_unix_nano > 0'
  service:
    pipelines:
      logs/log source:
        receivers:
        - filelog/log source
        processors:
        # ...
        - filter/remove older
        # ...
  ```

If you want to get logs which are appended after OpenTelemetry Collector Installation,
you can simply use `start_at: end`:

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
exporters:
  sumologic:
    source_name: my example name
```

#### Source Host

Please refer to [the Source Host section of Common configuration](#source-host).

#### Source Category

Please refer to [the Source Category section of Common configuration](#source-category).

#### Fields

Please refer to [the Fields/Metadata section of Common configuration](#fieldsmetadata).

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
  transform/custom fields:
    log_statements:
    - context: resource
      statements:
      - set(attributes["cloud.availability_zone"], "zone-1")
      - set(attributes["k8s.cluster.name"], "my-cluster")
  source/some name:
    source_name: my example name
    source_host: My Host
    source_category: My Category
```

##### Timestamp Parsing

The Installed Collector option to `Extract timestamp information from log file entries` in an
OpenTelemetry configuration can be achieved with [Transform Processor][transformprocessor]:

```yaml
  transform/clear_logs_timestamp:
    log_statements:
      - context: log
        statements:
          - set(time_unix_nano, 0)
```

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
  transform/custom fields:
    log_statements:
    - context: resource
      statements:
      - set(attributes["cloud.availability_zone"], "zone-1")
      - set(attributes["k8s.cluster.name"], "my-cluster")
  source/some name:
    source_name: my example name
    source_host: My Host
    source_category: My Category
  transform/clear_logs_timestamp:
    log_statements:
      - context: log
        statements:
          - set(time_unix_nano, 0)
exporters:
  sumologic/some_name:
```

If `transform/clear_logs_timestamp` is not used, timestamp parsing should be configured
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
    ## rel: https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/v0.134.0/pkg/stanza/docs/operators/regex_parser.md
    - type: regex_parser
      regex: (?P<timestamp>^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2},\d{3} (\+|\-)\d{4})
      ## Parse timestamp from timestamp field
      ## rel: https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/v0.134.0/pkg/stanza/docs/operators/time_parser.md
      timestamp:
        parse_from: attributes.timestamp
        ## Layout are substitute for Timestamp Format configuration
        layout_type: gotime
        layout: '2006-01-02 15:04:05,000 -0700'
  # ...
processors:
  transform/custom fields:
    log_statements:
    - context: resource
      statements:
      - set(attributes["cloud.availability_zone"], "zone-1")
      - set(attributes["k8s.cluster.name"], "my-cluster")
  source/some name:
    source_name: my example name
    source_host: My Host
    source_category: My Category
exporters:
  sumologic/some_name:
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
  source/some name:
    source_name: my example name
    source_host: My Host
    source_category: My Category
exporters:
  sumologic/some_name:
```

##### Encoding

Use `encoding` to set the encoding of your data. Full list of supporter encodings can be obtained from [Filelog Receiver documentation][supported_encodings].

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
  transform/custom fields:
    log_statements:
    - context: resource
      statements:
      - set(attributes["cloud.availability_zone"], "zone-1")
      - set(attributes["k8s.cluster.name"], "my-cluster")
  source/some name:
    source_name: my example name
    source_host: My Host
    source_category: My Category
exporters:
  sumologic/some_name:
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
  transform/custom fields:
    log_statements:
    - context: resource
      statements:
      - set(attributes["cloud.availability_zone"], "zone-1")
      - set(attributes["k8s.cluster.name"], "my-cluster")
  source/some name:
    source_name: my example name
    source_host: My Host
    source_category: My Category
exporters:
  sumologic/some_name:
```

If your multiline logs have a known end pattern use the `line_end_pattern` option.

More information is available in [Filelog Receiver documentation][multiline].

### Remote File Source

Remote File Source is not supported by the OpenTelemetry Collector.

### Syslog Source

The equivalent of the Syslog Source is a combination of
[the TCP][tcplogreceiver] or [the UDP][udplogreceiver] receivers
and [the Sumo logic Syslog Processor][sumologicsyslog].

__Note: The OpenTelemetry Collector also provides the [Syslog Receiver][syslogreceiver].
See [this document](comparison.md#syslog) for details.__

__The syslog messages could also be sent to sumologic using the [syslog exporter][syslogexporter] with the
[syslog parser][syslogparser]__

#### Overall example

Below is an example of an OpenTelemetry configuration for a Syslog Source.

```yaml
extensions:
  sumologic:
    installation_token: <installation_token>
    ## Time Zone is a substitute of Installed Collector `Time Zone`
    ## with `Use time zone from log file. If none is detected use:` option.
    ## This is used only if `clear_logs_timestamp` is set to `true` in sumologic exporter.
    ## Full list of time zones is available on wikipedia:
    ## https://en.wikipedia.org/wiki/List_of_tz_database_time_zones#List
    time_zone: America/Tijuana

receivers:
  ## Use TCP receiver for TCP protocol
  tcplog/first receiver:
    ## listen address in format host:port
    ## host 0.0.0.0 mean all network interfaces
    listen_address: 0.0.0.0:514
    ## Add network attributes
    ## `net.peer.name` is going to be used as processors.source.source_host
    ## rel: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.134.0/receiver/tcplogreceiver#configuration
    add_attributes: true
  ## Use UDP receiver for UDP protocol
  udplog/first receiver:
    ## listen address in format host:port
    ## host 0.0.0.0 mean all network interfaces
    listen_address: 0.0.0.0:514
    ## Add network attributes
    ## `net.peer.name` is going to be used as processors.source.source_host
    ## rel: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.134.0/receiver/udplogreceiver#configuration
    add_attributes: true

processors:
  ## There is no substitute for `Description` in current project phase.
  ## It is recommended to use comments for that purpose, like this one.
  ## sumologic_syslog/<source group name>:
  ## <source group name> can be substitute of Installed Collector `Name`
  sumologic_syslog/syslog source:
  ## re-associates record attributes to a resource to make them available to be used by Source Processor
  groupbyattrs:
    keys:
      - net.peer.name
      - facility
  ## Leave the timestamp parsing to the Sumo Logic backend
  transform/clear_logs_timestamp:
    log_statements:
      - context: log
        statements:
          - set(time_unix_nano, 0)
  ## The following configuration will add two fields to every resource
  transform/syslog source:
    log_statements:
      - context: resource
        statements:
          - set(attributes["cloud.availability_zone"], "zone-1")
          - set(attributes["k8s.cluster.name"], "my-cluster")
  source/syslog source:
    ## Installed Collector substitute for `Source Category`.
    source_category: example category
    ## Set Source Name to be facility name
    source_name: "%{facility}"
    ## Set Source Host to `net.peer.name`
    source_host: "%{net.peer.name}"

exporters:
  sumologic/syslog:

service:
  extensions:
  - sumologic
  pipelines:
    logs/syslog source:
      receivers:
      - tcplog/first receiver
      - udplog/first receiver
      processors:
      - transform/clear_logs_timestamp
      - sumologic_syslog/syslog source
      - groupbyattrs
      - transform/syslog source
      - source/syslog source
      exporters:
      - sumologic/syslog
```

#### Name

Please refer to [the Name section of Common configuration](#name-1)

#### Description

Please refer to [the Description section of Common configuration](#description-1).

#### Protocol and Port

Protocol is defined by receiver type. For UDP use [udplogreceiver][udplogreceiver] and for TCP use [tcplogreceiver][tcplogreceiver]. Port can be set by `listen_address`, for example to listen on port `6776` on all interfaces, use `listen_address: 0.0.0.0:6776`.

You can use multiple receivers with different names nad ports like in the following example:

```yaml
receivers:
  tcplog/first receiver:
    listen_address: 0.0.0.0:514
  tcplog/second receiver:
    listen_address: 127.0.0.1:5140
  udplog/first receiver:
    listen_address: 0.0.0.0:514
  udplog/second receiver:
    listen_address: 127.0.0.1:5150

processor:
  ## All my example logs
  sumologic_syslog/my example name:
  # ...
```

#### Source Category

A Source Category can be set in the [Source Processor][source-templates] configuration with the `source_category` option.

For example, the following snippet configures the Source Category as `My Category`:

```yaml
receivers:
  tcplog/first receiver:
    listen_address: 0.0.0.0:514
  tcplog/second receiver:
    listen_address: 127.0.0.1:5140
  udplog/first receiver:
    listen_address: 0.0.0.0:514
  udplog/second receiver:
    listen_address: 127.0.0.1:5150
processor:
  ## All my example logs
  sumologic_syslog/my example name:
  # ...
  source/some name:
    source_category: My Category
```

#### Fields

Use the [resourceprocessor][resourceprocessor] to set custom fields.

For example, the following snippet configures two fields, `cloud.availability_zone` and `k8s.cluster.name`:

```yaml
receivers:
  tcplog/first receiver:
    listen_address: 0.0.0.0:514
  tcplog/second receiver:
    listen_address: 127.0.0.1:5140
  udplog/first receiver:
    listen_address: 0.0.0.0:514
  udplog/second receiver:
    listen_address: 127.0.0.1:5150
processors:
  ## All my example logs
  sumologic_syslog/my example name:
  # ...
  transform/my example name fields:
    log_statements:
      - context: resource
        statements:
          - set(attributes["cloud.availability_zone"], "zone-1")
          - set(attributes["k8s.cluster.name"], "my-cluster")
  source/some name:
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
  tcplog/first receiver:
    listen_address: 0.0.0.0:514
  tcplog/second receiver:
    listen_address: 127.0.0.1:5140
  udplog/first receiver:
    listen_address: 0.0.0.0:514
  udplog/second receiver:
    listen_address: 127.0.0.1:5150
processors:
  ## All my example logs
  sumologic_syslog/my example name:
  # ...
  transform/my example name fields:
    log_statements:
      - context: resource
        statements:
          - set(attributes["cloud.availability_zone"], "zone-1")
          - set(attributes["k8s.cluster.name"], "my-cluster")
  source/some name:
    source_category: My Category
```

If `clear_logs_timestamp` is set to `false`, timestamp parsing should be configured
manually, like in the following snippet:

```yaml
extensions:
  sumologic:
    time_zone: America/Tijuana
receivers:
  tcplog/first receiver:
    listen_address: 0.0.0.0:514
    operators:
    ## Extract timestamp into timestamp field using regex
    ## rel: https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/v0.134.0/pkg/stanza/docs/operators/regex_parser.md
    - type: regex_parser
      regex: (?P<timestamp>^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2},\d{3} (\+|\-)\d{4})
      ## Parse timestamp from timestamp field
      ## rel: https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/v0.134.0/pkg/stanza/docs/operators//time_parser.md
      timestamp:
        parse_from: attributes.timestamp
        ## Layout are substitute for Timestamp Format configuration
        layout_type: gotime
        layout: '2006-01-02 15:04:05,000 -0700'
processors:
  ## All my example logs
  sumologic_syslog/my example name:
  # ...
  transform/my example name fields:
    log_statements:
      - context: resource
        statements:
          - set(attributes["cloud.availability_zone"], "zone-1")
          - set(attributes["k8s.cluster.name"], "my-cluster")
  source/some name:
    source_category: My Category
exporters:
  sumologic/some name:
    ## Keep manually parsed timestamps
    clear_logs_timestamp: false
```

The following example snippet skips timestamp parsing so the Collector uses Receipt Time:

```yaml
extensions:
  sumologic:
    time_zone: America/Tijuana
receivers:
  tcplog/first receiver:
    listen_address: 0.0.0.0:514
processors:
  ## All my example logs
  sumologic_syslog/my example name:
  # ...
  transform/clear_logs_timestamp:
    log_statements:
      - context: log
        statements:
          - set(time_unix_nano, 0)
  transform/my example name fields:
    log_statements:
      - context: resource
        statements:
          - set(attributes["cloud.availability_zone"], "zone-1")
          - set(attributes["k8s.cluster.name"], "my-cluster")
  source/some name:
    source_category: My Category
exporters:
  sumologic/some name:
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
  tcplog/first receiver:
    listen_address: 0.0.0.0:514
processors:
  ## All my example logs
  sumologic_syslog/my example name:
  # ...
  transform/clear_logs_timestamp:
    log_statements:
      - context: log
        statements:
          - set(time_unix_nano, 0)
  transform/my example name fields:
    log_statements:
      - context: resource
        statements:
          - set(attributes["cloud.availability_zone"], "zone-1")
          - set(attributes["k8s.cluster.name"], "my-cluster")
  source/some name:
    source_category: My Category
    ## Set Source Name to facility, which is set by sumologicsyslogprocessor
    source_name: "%{facility}"
exporters:
  sumologic/some name:
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
  tcplog/first receiver:
    listen_address: 0.0.0.0:514
    add_attributes: true
processors:
  ## All my example logs
  sumologic_syslog/my example name:
  # ...
  transform/clear_logs_timestamp:
    log_statements:
      - context: log
        statements:
          - set(time_unix_nano, 0)
  transform/my example name fields:
    log_statements:
      - context: resource
        statements:
          - set(attributes["cloud.availability_zone"], "zone-1")
          - set(attributes["k8s.cluster.name"], "my-cluster")
  source/some name:
    source_category: My Category
    ## Set Source Name to facility, which is set by sumologicsyslogprocessor
    source_name: "%{facility}
    source_host: "%{net.peer.name}
exporters:
  sumologic/some name:
```

### Docker Logs Source

It is possible to scrape docker logs using [Filelog Receiver][filelogreceiver], but the following features are not supported yet:

- metadata enrichment (sending container name along with container id)
- docker events

therefore we do not provide migration process for Docker Logs Source (DockerLog).

### Docker Stats Source

Docker Stats Source can be accessed with [the Dockerstats receiver][dockerstatsreceiver].

#### Overall example

Below is an example of an OpenTelemetry configuration for a Docker Stats Source.

```yaml
extensions:
  sumologic:
    installation_token: <installation_token>
    ## Time Zone is a substitute of Installed Collector `Time Zone`
    ## with `Use time zone from log file. If none is detected use:` option.
    ## This is used only if `clear_logs_timestamp` is set to `true` in sumologic exporter.
    ## Full list of time zones is available on wikipedia:
    ## https://en.wikipedia.org/wiki/List_of_tz_database_time_zones#List
    time_zone: America/Tijuana

receivers:
  docker_stats:
    ## Docker daemon's socket address.
    ## Example for running Docker on MacOS with Colima, default is "unix:///var/run/docker.sock"
    endpoint: "unix:///Users/<user>/.colima/default/docker.sock"

    ## Default is 10s
    collection_interval: 20s

    ## A list of images for which corresponding containers won't be scraped.
    ## Strings, regexes and globs are supported, more information in the receiver's readme:
    ## https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/v0.134.0/receiver/dockerstatsreceiver#configuration
    excluded_images:
      ## Exclude particular image
      - docker.io/library/nginx:1.2
      ## Exclude by regex
      ## Note: regex must be but between / characters
      - /other-docker-registry.*nginx/
      ## Exclude by glob
      - exclude-*-this/nginx
      ## Use negation: scrape metrics only from nginx containers
      - !*nginx*
      ## Negation for regexes requires using ! before the slash character
      - !/.*nginx.*/

    ## Timeout for any Docker daemon query.
    timeout: 5s
    ## Must be 1.22 or above
    api_version: 1.22

    ## Enable or disable particular metrics.
    ## Full list of metrics with their default config is available at https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/v0.134.0/receiver/dockerstatsreceiver/documentation.md
    metrics:
      container.cpu.usage.percpu:
        enabled: true
      container.network.io.usage.tx_dropped:
        enabled: false

processors:
  filter/dockerstats:
    metrics:
      ## To filter out, use "exclude" instead
      include:
        match_type: regexp
        resource_attributes:
          - key: container.name
            value: sumo-container-.*

  sumologicschema/dockerstats:
    translate_docker_metrics: true

exporters:
  sumologic/dockerstats:

service:
  extensions:
  - sumologic
  pipelines:
    metrics/docker_stats source:
      receivers:
      - docker_stats
      processors:
      - filter/dockerstats
      - sumologicschema/dockerstats
      exporters:
      - sumologic/dockerstats
```

#### Name

Please refer to [the Name section of Common configuration](#name-1).

#### Description

Please refer to [the Description section of Common configuration](#description-1).

#### URI

To specify URI, use `endpoint` option:

```yaml
receivers:
  docker_stats:
    ## Docker daemon's socket address.
    ## Example for running Docker on MacOS with Colima, default is "unix:///var/run/docker.sock"
    endpoint: "unix:///Users/<user>/.colima/default/docker.sock"
```

`Cert Path` option is not supported in OpenTelemetry Collector.

#### Container filters

Containers cannot be filtered by their name directly in the receiver. [Filter Processor][filterprocessor] has to be used for that purpose.
To filter in containers by their name, use the following processor config:

```yaml
receivers:
  docker_stats:
    ## ...

processors:
  filter:
    metrics:
      ## To filter out, use "exclude" instead
      include:
        match_type: regexp
        resource_attributes:
          - key: container.name
            value: sumo-container-.*
```

You can also filter out containers by their image name through `excluded_images` option:

```yaml
receivers:
  docker_stats:
    ## A list of images for which corresponding containers won't be scraped.
    ## Strings, regexes and globs are supported, more information in the receiver's readme:
    ## https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/v0.134.0/receiver/dockerstatsreceiver#configuration
    excluded_images:
      ## Exclude particular image
      - docker.io/library/nginx:1.2
      ## Exclude by regex
      ## Note: regex must be but between / characters
      - /other-docker-registry.*nginx/
      ## Exclude by glob
      - exclude-*-this/nginx
      ## Use negation: scrape metrics only from nginx containers
      - !*nginx*
      ## Negation for regexes requires using ! before the slash character
      - !/.*nginx.*/
```

Valid names are strings, [regexes](https://pkg.go.dev/regexp) and [globs](https://github.com/gobwas/glob).
Negation can also be used, for example glob `!my*container` will exclude all containers for which the image name doesn't match glob `my*container`.

#### Source Host

Please refer to [the Source Host section of Common configuration](#source-host).

#### Source Category

Please refer to [the Source Category section of Common configuration](#source-category).

#### Fields

Please refer to [the Fields/Metadata section of Common configuration](#fieldsmetadata).

#### Scan interval

To indicate a scan interval, use `collection_interval` option:

```yaml
receivers:
  docker_stats:
    ## Default is 10s
    collection_interval: 20s
```

#### Metrics

In OpenTelemetry Collector, some metrics are being emitted by default, meanwhile some have to be enabled. You can enable emitting a metric by setting `metrics.<metric_name>.enabled` to `true` and disable it by setting this option to `false`, for example:

```yaml
receivers:
  dockerstats:
    metrics:
      container.cpu.usage.percpu:
        enabled: true
      container.network.io.usage.tx_dropped:
        enabled: false
```

The table below shows how some metrics commonly used by the Installed Collector map to OpenTelemetry metrics.

| Installed Collector name          | OpenTelemetry name                              | Notes               |
|-----------------------------------|-------------------------------------------------|---------------------|
| cpu_percentage                    | container.cpu.percent                           |                     |
| online_cpus                       | _not available_                                 |                     |
| system_cpu_usage                  | container.cpu.usage.system                      | disabled by default |
| cpu_usage.percpu_usage            | container.cpu.usage.percpu                      | disabled by default |
| cpu_usage.total_usage             | container.cpu.usage.total                       |                     |
| cpu_usage.usage_in_kernelmode     | container.cpu.usage.kernelmode                  |                     |
| cpu_usage.usage_in_usermode       | container.cpu.usage.usermode                    |                     |
| throttling_data.periods           | container.cpu.throttling_data.periods           | disabled by default |
| throttling_data.throttled_periods | container.cpu.throttling_data.throttled_periods | disabled by default |
| throttling_data.throttled_time    | container.cpu.throttling_data.throttled_time    | disabled by default |
|                                   |                                                 |                     |
| failcnt                           | _not available_                                 |                     |
| limit                             | container.memory.usage.limit                    |                     |
| max_usage                         | container.memory.usage.max                      | disabled by default |
| memory_percentage                 | container.memory.percent                        |                     |
| usage                             | container.memory.usage.total                    |                     |
| stats.active_anon                 | container.memory.active_anon                    | disabled by default |
| stats.active_file                 | container.memory.active_file                    | disabled by default |
| stats.cache                       | container.memory.cache                          |                     |
| stats.hierarchical_memory_limit   | container.memory.hierarchical_memory_limit      | disabled by default |
| stats.inactive_anon               | container.memory.inactive_anon                  | disabled by default |
| stats.inactive_file               | container.memory.inactive_file                  | disabled by default |
| stats.mapped_file                 | container.memory.mapped_file                    | disabled by default |
| stats.pgfault                     | container.memory.pgfault                        | disabled by default |
| stats.pgmajfault                  | container.memory.pgmajfault                     | disabled by default |
| stats.pgpgin                      | container.memory.pgpgin                         | disabled by default |
| stats.pgpgout                     | container.memory.pgpgout                        | disabled by default |
| stats.rss                         | container.memory.rss                            | disabled by default |
| stats.rss_huge                    | container.memory.rss_huge                       | disabled by default |
| stats.unevictable                 | container.memory.unevictable                    | disabled by default |
| stats.writeback                   | container.memory.writeback                      | disabled by default |
| stats.total_active_anon           | container.memory.total_active_anon              | disabled by default |
| stats.total_active_file           | container.memory.total_active_file              | disabled by default |
| stats.total_cache                 | container.memory.total_cache                    |                     |
| stats.total_inactive_anon         | container.memory.total_inactive_anon            | disabled by default |
| stats.total_mapped_file           | container.memory.total_mapped_file              | disabled by default |
| stats.total_pgfault               | container.memory.total_pgfault                  | disabled by default |
| stats.total_pgmajfault            | container.memory.total_pgmajfault               | disabled by default |
| stats.total_pgpgin                | container.memory.total_pgpgin                   | disabled by default |
| stats.total_pgpgout               | container.memory.total_pgpgout                  | disabled by default |
| stats.total_rss                   | container.memory.total_rss                      | disabled by default |
| stats.total_rss_huge              | container.memory.total_rss_huge                 | disabled by default |
| stats.total_unevictable           | container.memory.total_unevictable              | disabled by default |
| stats.total_writeback             | container.memory.total_writeback                | disabled by default |
|                                   |                                                 |                     |
| io_merged_recursive               | container.blockio.io_merged_recursive           | disabled by default |
| io_queue_recursive                | container.blockio.io_queued_recursive           | disabled by default |
| io_service_bytes_recursive        | container.blockio.io_service_bytes_recursive    |                     |
| io_service_time_recursive         | container.blockio.io_service_time_recursive     | disabled by default |
| io_serviced_recursive             | container.blockio.io_serviced_recursive         | disabled by default |
| io_time_recursive                 | container.blockio.io_time_recursive             | disabled by default |
| io_wait_time_recursive            | container.blockio.io_wait_time_recursive        | disabled by default |
| sectors_recursive                 | container.blockio.sectors_recursive             | disabled by default |
|                                   |                                                 |                     |
| current                           | _not available_                                 |                     |

Full list of metrics available in this receiver can be found in the [documentation][dockerstatsmetrics].

Unfortunately, Sumo Logic apps don't work with these metric names yet. To convieniently translate them, use [Sumo Logic Schema Processor][sumologicschemaprocessor]:

```yaml
processors:
  sumologicschema/dockerstats:
    translate_docker_metrics: true
```

#### Metadata

The metadata sent by Installed Collector correspond to the metadata sent by OpenTelemetry Collector in the following way:

- `container.FullID` corresponds to `container.id`
- `container.ID`, which is a shorter version of `container.FullID` is not being emitted - if needed, `transform` processor can be used to trim it
- `container.ImageName` corresponds to `container.image.name`
- `container.Name` corresponds to `container.name`
- `container.ImageID` and `container.ImageFullID` are not being emitted

These metadata is represented as resource attributes and can be translated by using [Sumo Logic Schema Processor][sumologicschemaprocessor]
in the same way as for translating metric names, by using the following config:

```yaml
processors:
  sumologicschema/dockerstats:
    translate_docker_metrics: true
```

In addition, there is some additional metadata sent by the OpenTelemetry Collector. Full list of it can be seen [here]
[dockerstatsmetrics].

### Script Source

Script Source is not supported by the OpenTelemetry Collector.

### Streaming Metrics Source

For the Streaming Metrics Source we are using [the Telegraf receiver][telegrafreceiver]
with [socket_listener plugin][telegraf-socket_listener].

#### Overall example

Below is an example of an OpenTelemetry configuration for a Streaming Metrics Source.

```yaml
extensions:
  sumologic:
    installation_token: <installation_token>
    ## Time Zone is a substitute of Installed Collector `Time Zone`
    ## Full list of time zones is available on wikipedia:
    ## https://en.wikipedia.org/wiki/List_of_tz_database_time_zones#List
    time_zone: America/Tijuana
receivers:
  ## There is no substitute for `Description` in current project phase.
  ## It is recommended to use comments for that purpose, like this one.
  ## telegraf/<source group name>:
  ## <source group name> can be substitute of Installed Collector `Name`.
  telegraf/metrics source:
    ## Do not add metric field separately as data point label.
    separate_field: false
    ## Telegraf configuration
    agent_config: |
      [agent]
        ## Get metrics every 15 seconds
        interval = "15s"
        ## Flush metrics every 15 seconds
        flush_interval = "15s"
      ## socket_listener listen on given protocol://hostname:port for metrics
      [[inputs.socket_listener]]
        ## listen for metrics on UDP port 2006 on localhost
        service_address = "udp://localhost:2006"
        ## Get metrics in carbon2 format
        data_format = "carbon2"
processors:
  ## The following configuration will add two metadata properties to every record
  transform/metric source:
    metric_statements:
    - context: resource
      statements:
      - set(attributes["cloud.availability_zone"], "zone-1")
      - set(attributes["k8s.cluster.name"], "my-cluster")
  source:
    ## Set _sourceName
    source_name: my example name
    ## Installed Collector substitute for `Source Category`.
    source_category: example category
    ## Installed Collector substitute for `Source Host`.
    source_host: example host
exporters:
  sumologic:
service:
  extensions:
  - sumologic
  pipelines:
    metrics/metric source:
      receivers:
      - telegraf/metrics source
      processors:
      - transform/metric source
      exporters:
      - sumologic
```

#### Name

Please refer to [the Name section of Common configuration](#name-1).

#### Description

Please refer to [the Description section of Common configuration](#description-1).

#### Protocol and Port

Protocol and Port can be configured using `service_address` in Telegraf `socket_listener` plugin configuration.

For example:

```yaml
receivers:
  ## All my example metrics
  telegraf/my example name:
    ## Telegraf configuration
    agent_config: |
      ## socket_listener listen on given protocol://hostname:port for metrics
      [[inputs.socket_listener]]
        ## listen for metrics on UDP port 2006 on localhost
        service_address = "udp://localhost:2006"
        ## Get metrics in carbon2 format
        data_format = "carbon2"
  # ...
processors:
  source:
    source_name: my example name
```

#### Content Type

Content Type can be configured using `data_format` in the Telegraf `socket_listener` plugin configuration.
Any of the [available formats][telegraf-input-formats] can be used, especially `graphite` and `carbon2`.

For example:

```yaml
receivers:
  ## All my example metrics
  telegraf/my example name:
    ## Telegraf configuration
    agent_config: |
      ## socket_listener listen on given protocol://hostname:port for metrics
      [[inputs.socket_listener]]
        ## listen for metrics on UDP port 2006 on localhost
        service_address = "udp://localhost:2006"
        ## Get metrics in carbon2 format
        data_format = "carbon2"
  # ...
processors:
  source:
    source_name: my example name
```

#### Source Category

Please refer to [the Source Category section of Common configuration](#source-category).

#### Metadata

Please refer to [the Fields/Metadata section of Common configuration](#fields).

### Host Metrics Source

It is recommended to use dedicated Sumo Logic app for Host Metrics for OpenTelemetry Collector.
In order to use old dashboards, please follow the [Using Telegraf Receiver](#using-telegraf-receiver) section.

#### Using Telegraf Receiver

The equivalent of the Host Metrics Source is [the telegraf receiver][telegrafreceiver] with appropiate plugins.

__Note: The are differences between the Installed Collector and the Openelemetry Collector host metrics.
See [this document](comparison.md#host-metrics) to learn more.__

##### Overall Example

Below is an example of an OpenTelemetry configuration for a Host Metrics Source.

```yaml
extensions:
  sumologic:
    installation_token: <installation_token>
    ## Time Zone is a substitute of Installed Collector `Time Zone`
    ## Full list of time zones is available on wikipedia:
    ## https://en.wikipedia.org/wiki/List_of_tz_database_time_zones#List
    time_zone: America/Tijuana

receivers:
  ## There is no substitute for `Description` in current project phase.
  ## It is recommended to use comments for that purpose, like this one.
  ## telegraf/<source group name>:
  ## <source group name> can be substitute of Installed Collector `Name`.
  telegraf/metrics source:
    ## Do not add metric field separately as data point label.
    separate_field: false
    ## Telegraf configuration
    agent_config: |
      [agent]
        ## Get metrics every 15 seconds
        interval = "15s"
        ## Flush metrics every 15 seconds
        flush_interval = "15s"

      ## CPU metrics
      [[inputs.cpu]]
        percpu = false
        totalcpu = true
        collect_cpu_time = false
        report_active = true
        namepass = [ "cpu" ]
        fieldpass = [ "usage_active", "usage_steal", "usage_iowait", "usage_irq", "usage_user", "usage_idle", "usage_nice", "usage_system", "usage_softirq" ]

      ## CPU metrics
      [[inputs.system]]
        namepass = [ "system" ]
        fieldpass = [ "load1", "load5", "load15" ]

      ## Memory metrics
      [[inputs.mem]]
        fieldpass = [ "total", "free", "used", "used_percent", "available", "available_percent" ]

      ## TCP metrics
      [[inputs.netstat]]
        fieldpass = [ "tcp_close", "tcp_close_wait", "tcp_closing", "tcp_established", "tcp_listen", "tcp_time_wait" ]

      ## Network metrics
      [[inputs.net]]
        interfaces = ["eth*", "en*", "lo*"]
        ignore_protocol_stats = true
        fieldpass = [ "bytes_sent", "bytes_recv", "packets_sent", "packets_recv" ]

      ## Disk metrics
      [[inputs.disk]]
        namepass = [ "disk" ]
        fieldpass = [ "used", "used_percent", "inodes_free" ]
        ignore_fs = ["tmpfs", "devtmpfs", "devfs", "iso9660", "overlay", "aufs", "squashfs"]

      ## Disk metrics
      [[inputs.diskio]]
        fieldpass = [ "reads", "read_bytes", "writes", "write_bytes" ]
processors:
  ## The following configuration will add two metadata properties to every record
  transform/metrics source:
    metric_statements:
    - context: resource
      statements:
      - set(attributes["cloud.availability_zone"], "zone-1")
      - set(attributes["k8s.cluster.name"], "my-cluster")
  source:
    ## Set _sourceName
    source_name: my example name
    ## Installed Collector substitute for `Source Category`.
    source_category: example category
    ## Installed Collector substitute for `Source Host`.
    source_host: example host
exporters:
  sumologic:
    ## Ensure compability with Installed Colllector metric name
    translate_telegraf_attributes: true
service:
  extensions:
  - sumologic
  pipelines:
    metrics/metric source:
      receivers:
      - telegraf/metrics source
      processors:
      - transform/metrics source
      - source
      exporters:
      - sumologic
```

##### Name

Please refer to [the Name section of Common configuration](#name-1).

##### Description

Please refer to [the Description section of Common configuration](#description-1).

##### Source Host

Please refer to [the Source Host section of Common configuration](#source-host).

##### Source Category

Please refer to [the Source Category section of Common configuration](#source-category).

##### Metadata

Please refer to [the Fields/Metadata section of Common configuration](#fields).

##### Scan Interval

To set Scan Interval use `interval` in Telegraf's agent configuration.

The following example shows how to set it for 1 minute:

```yaml
receivers:
  ## All my example metrics
  telegraf/my example name:
    agent_config: |
      [agent]
        interval = "1m"
        flush_interval = "1m"
  # ...
processors:
  transform/my example name fields:
    metric_statements:
    - context: resource
      statements:
      - set(attributes["cloud.availability_zone"], "zone-1")
      - set(attributes["k8s.cluster.name"], "my-cluster")
  source/some name:
    source_name: my example name
    source_host: my_host
    source_category: My Category
```

##### Metrics

Telegraf offers a set of various plugins you can use to get metrics.
In this section, we are describing only plugins that are required
for seamless migration from the Installed Collector.
If you are interested in other metrics, see [list of Telegraf input plugins][telegraf-input-plugins].

Each of the subtopics contain a table that describes how Installed Collector
metrics translate to Telegraf metrics.

To ensure all dashboards are working as before,
Telegraf metric names are translated to the Installed Collector by [sumologicexporter][sumologicexporter].
You can disable this by setting `translate_telegraf_attributes` to `false`,
but in this case you need to update your dashboards.

###### CPU

To get CPU metrics we are using the [inputs.cpu][telegraf-input-cpu]
and the [inputs.system][telegraf-input-system] Telegraf plugins.

| Metric Name       | Telegraf plugin | Telegraf metric name |
|-------------------|-----------------|----------------------|
| CPU_User          | inputs.cpu      | cpu_usage_user       |
| CPU_Sys           | inputs.cpu      | cpu_usage_System     |
| CPU_Nice          | inputs.cpu      | cpu_usage_nice       |
| CPU_Idle          | inputs.cpu      | cpu_usage_idle       |
| CPU_IOWait        | inputs.cpu      | cpu_usage_iowait     |
| CPU_Irq           | inputs.cpu      | cpu_usage_irq        |
| CPU_SoftIrq       | inputs.cpu      | cpu_usage_softirq    |
| CPU_Stolen        | inputs.cpu      | cpu_usage_steal      |
| CPU_LoadAvg_1min  | inputs.system   | system_load1         |
| CPU_LoadAvg_5min  | inputs.system   | system_load5         |
| CPU_LoadAvg_15min | inputs.system   | system_load15        |
| CPU_Total         | inputs.cpu      | cpu_usage_active     |

The following example shows the desired configuration:

```yaml
receivers:
  ## All my example metrics
  telegraf/my example name:
    agent_config: |
      [agent]
        interval = "1m"
        flush_interval = "1m"

      ## CPU metrics
      [[inputs.cpu]]
        percpu = false
        totalcpu = true
        collect_cpu_time = false
        report_active = true
        namepass = [ "cpu" ]
        fieldpass = [ "usage_active", "usage_steal", "usage_iowait", "usage_irq", "usage_user", "usage_idle", "usage_nice", "usage_system", "usage_softirq" ]

      ## CPU metrics
      [[inputs.system]]
        namepass = [ "system" ]
        fieldpass = [ "load1", "load5", "load15" ]
  # ...
processors:
  transform/my example name fields:
    metric_statements:
    - context: resource
      statements:
      - set(attributes["cloud.availability_zone"], "zone-1")
      - set(attributes["k8s.cluster.name"], "my-cluster")
  source/some name:
    source_name: my example name
    source_host: my_host
    source_category: My Category
```

###### Memory

To get CPU metrics we are using the [inputs.mem][telegraf-input-mem] Telegraf plugin.

| Metric Name     | Telegraf plugin | Telegraf metric name  |
|-----------------|-----------------|-----------------------|
| Mem_Total       | inputs.mem      | mem_total             |
| Mem_Used        | N/A             | N/A                   |
| Mem_Free        | inputs.mem      | mem_free              |
| Mem_ActualFree  | inputs.mem      | mem_available         |
| Mem_ActualUsed  | inputs.mem      | mem_used              |
| Mem_UsedPercent | inputs.mem      | mem_used_percent      |
| Mem_FreePercent | inputs.mem      | mem_available_percent |
| Mem_PhysicalRam | N/A             | N/A                   |

The following example shows the desired configuration:

```yaml
receivers:
  ## All my example metrics
  telegraf/my example name:
    agent_config: |
      [agent]
        interval = "1m"
        flush_interval = "1m"

      ## CPU metrics
      [[inputs.cpu]]
        percpu = false
        totalcpu = true
        collect_cpu_time = false
        report_active = true
        namepass = [ "cpu" ]
        fieldpass = [ "usage_active", "usage_steal", "usage_iowait", "usage_irq", "usage_user", "usage_idle", "usage_nice", "usage_system", "usage_softirq" ]

      ## CPU metrics
      [[inputs.system]]
        namepass = [ "system" ]
        fieldpass = [ "load1", "load5", "load15" ]

      ## Memory metrics
      [[inputs.mem]]
        fieldpass = [ "total", "free", "used", "used_percent", "available", "available_percent" ]
  # ...
processors:
  transform/my example name fields:
    metric_statements:
    - context: resource
      statements:
      - set(attributes["cloud.availability_zone"], "zone-1")
      - set(attributes["k8s.cluster.name"], "my-cluster")
  source/some name:
    source_name: my example name
    source_host: my_host
    source_category: My Category
```

###### TCP

To get TCP metrics we are using the [inputs.netstat][telegraf-input-netstat] Telegraf plugin.

| Metric Name       | Telegraf plugin | Telegraf metric name    |
|-------------------|-----------------|-------------------------|
| TCP_InboundTotal  | N/A             | N/A                     |
| TCP_OutboundTotal | N/A             | N/A                     |
| TCP_Established   | inputs.netstat  | netstat_tcp_established |
| TCP_Listen        | inputs.netstat  | netstat_tcp_listen      |
| TCP_Idle          | N/A             | N/A                     |
| TCP_Closing       | inputs.netstat  | netstat_tcp_closing     |
| TCP_CloseWait     | inputs.netstat  | netstat_tcp_close_wait  |
| TCP_Close         | inputs.netstat  | netstat_tcp_close       |
| TCP_TimeWait      | inputs.netstat  | netstat_tcp_time_wait   |

The following example shows the desired configuration:

```yaml
receivers:
  ## All my example metrics
  telegraf/my example name:
    agent_config: |
      [agent]
        interval = "1m"
        flush_interval = "1m"

      ## CPU metrics
      [[inputs.cpu]]
        percpu = false
        totalcpu = true
        collect_cpu_time = false
        report_active = true
        namepass = [ "cpu" ]
        fieldpass = [ "usage_active", "usage_steal", "usage_iowait", "usage_irq", "usage_user", "usage_idle", "usage_nice", "usage_system", "usage_softirq" ]

      ## CPU metrics
      [[inputs.system]]
        namepass = [ "system" ]
        fieldpass = [ "load1", "load5", "load15" ]

      ## Memory metrics
      [[inputs.mem]]
        fieldpass = [ "total", "free", "used", "used_percent", "available", "available_percent" ]

      ## TCP metrics
      [[inputs.netstat]]
        fieldpass = [ "tcp_close", "tcp_close_wait", "tcp_closing", "tcp_established", "tcp_listen", "tcp_time_wait" ]
  # ...
processors:
  transform/my example name fields:
    metric_statements:
    - context: resource
      statements:
      - set(attributes["cloud.availability_zone"], "zone-1")
      - set(attributes["k8s.cluster.name"], "my-cluster")
  source/some name:
    source_name: my example name
    source_host: my_host
    source_category: My Category
```

###### Network

To get network metrics we are using the [inputs.net][telegraf-input-net] Telegraf plugin.

| Metric Name    | Telegraf plugin | Telegraf metric name |
|----------------|-----------------|----------------------|
| Net_InPackets  | inputs.net      | net_packets_recv     |
| Net_OutPackets | inputs.net      | net_packets_sent     |
| Net_InBytes    | inputs.net      | net_bytes_recv       |
| Net_OutBytes   | inputs.net      | net_bytes_sent       |

The following example shows the desired configuration:

```yaml
receivers:
  ## All my example metrics
  telegraf/my example name:
    agent_config: |
      [agent]
        interval = "1m"
        flush_interval = "1m"

      ## CPU metrics
      [[inputs.cpu]]
        percpu = false
        totalcpu = true
        collect_cpu_time = false
        report_active = true
        namepass = [ "cpu" ]
        fieldpass = [ "usage_active", "usage_steal", "usage_iowait", "usage_irq", "usage_user", "usage_idle", "usage_nice", "usage_system", "usage_softirq" ]

      ## CPU metrics
      [[inputs.system]]
        namepass = [ "system" ]
        fieldpass = [ "load1", "load5", "load15" ]

      ## Memory metrics
      [[inputs.mem]]
        fieldpass = [ "total", "free", "used", "used_percent", "available", "available_percent" ]

      ## TCP metrics
      [[inputs.netstat]]
        fieldpass = [ "tcp_close", "tcp_close_wait", "tcp_closing", "tcp_established", "tcp_listen", "tcp_time_wait" ]

      ## Network metrics
      [[inputs.net]]
        interfaces = ["eth*", "en*", "lo*"]
        ignore_protocol_stats = true
        fieldpass = [ "bytes_sent", "bytes_recv", "packets_sent", "packets_recv" ]
  # ...
processors:
  transform/my example name fields:
    metric_statements:
    - context: resource
      statements:
      - set(attributes["cloud.availability_zone"], "zone-1")
      - set(attributes["k8s.cluster.name"], "my-cluster")
  source/some name:
    source_name: my example name
    source_host: my_host
    source_category: My Category
```

###### Disk

To get disk metrics we are using the [inputs.diskio][telegraf-input-diskio]
and the [inputs.disk][telegraf-input-disk] Telegraf plugins.

| Metric Name          | Telegraf plugin | Telegraf metric name |
|----------------------|-----------------|----------------------|
| Disk_Reads           | inputs.diskio   | diskio_reads         |
| Disk_ReadBytes       | inputs.diskio   | diskio_read_bytes    |
| Disk_Writes          | inputs.diskio   | diskio_writes        |
| Disk_WriteBytes      | inputs.diskio   | diskio_write_bytes   |
| Disk_Queue           | N/A             | N/A                  |
| Disk_InodesAvailable | inputs.disk     | disk_inodes_free     |
| Disk_Used            | inputs.disk     | disk_used            |
| Disk_UsedPercent     | inputs.disk     | disk_used_percent    |
| Disk_Available       | N/A             | N/A                  |

The following example shows the desired configuration:

```yaml
receivers:
  ## All my example metrics
  telegraf/my example name:
    agent_config: |
      [agent]
        interval = "1m"
        flush_interval = "1m"

      ## CPU metrics
      [[inputs.cpu]]
        percpu = false
        totalcpu = true
        collect_cpu_time = false
        report_active = true
        namepass = [ "cpu" ]
        fieldpass = [ "usage_active", "usage_steal", "usage_iowait", "usage_irq", "usage_user", "usage_idle", "usage_nice", "usage_system", "usage_softirq" ]

      ## CPU metrics
      [[inputs.system]]
        namepass = [ "system" ]
        fieldpass = [ "load1", "load5", "load15" ]

      ## Memory metrics
      [[inputs.mem]]
        fieldpass = [ "total", "free", "used", "used_percent", "available", "available_percent" ]

      ## TCP metrics
      [[inputs.netstat]]
        fieldpass = [ "tcp_close", "tcp_close_wait", "tcp_closing", "tcp_established", "tcp_listen", "tcp_time_wait" ]

      ## Network metrics
      [[inputs.net]]
        interfaces = ["eth*", "en*", "lo*"]
        ignore_protocol_stats = true
        fieldpass = [ "bytes_sent", "bytes_recv", "packets_sent", "packets_recv" ]

      ## Disk metrics
      [[inputs.disk]]
        namepass = [ "disk" ]
        fieldpass = [ "used", "used_percent", "inodes_free" ]
        ignore_fs = ["tmpfs", "devtmpfs", "devfs", "iso9660", "overlay", "aufs", "squashfs"]

      ## Disk metrics
      [[inputs.diskio]]
        fieldpass = [ "reads", "read_bytes", "writes", "write_bytes" ]
  # ...
processors:
  transform/my example name fields:
    metric_statements:
    - context: resource
      statements:
      - set(attributes["cloud.availability_zone"], "zone-1")
      - set(attributes["k8s.cluster.name"], "my-cluster")
  source/some name:
    source_name: my example name
    source_host: my_host
    source_category: My Category
```

### Local Windows Event Log Source

There is no migration process from Installed Collector to OpenTelemetry Collector.
In order to use OpenTelemetry Collector, dedicated Sumo Logic app needs to be
installed.

### Local Windows Performance Monitor Log Source

There is no migration process from Installed Collector to OpenTelemetry Collector.

### Windows Active Directory Source

Windows Active Directory Source is not supported by the OpenTelemetry Collector.

### Script Action

Script Action is not supported by the OpenTelemetry Collector.

## Local Configuration File

This section describes migration steps for an Installed Collector managed with a Local Configuration File.

### Collector

#### user.properties

The following table shows the equivalent [user.properties][user.properties] for OpenTelemetry.

| user.properties key                           | The OpenTelemetry Collector Key                                                           |
|-----------------------------------------------|-------------------------------------------------------------------------------------------|
| `wrapper.java.command=JRE Bin Location`       | N/A                                                                                       |
| ~~`accessid=accessId`~~                       | N/A, use [extensions.sumologic.installation_token](/docs/configuration.md#authentication) |
| ~~`accesskey=accessKey`~~                     | N/A, use [extensions.sumologic.installation_token](/docs/configuration.md#authentication) |
| `category=category`                           | [extensions.sumologic.collector_category](#category)                                      |
| `clobber=true/false`                          | [extensions.sumologic.clobber][sumologicextension]                                        |
| `description=description`                     | [extensions.sumologic.collector_description](#description)                                |
| `disableActionSource=true/false`              | N/A                                                                                       |
| `disableScriptSource=true/false`              | N/A                                                                                       |
| `disableUpgrade=true/false`                   | N/A                                                                                       |
| `enableActionSource=true/false`               | N/A                                                                                       |
| `enableScriptSource=true/false`               | N/A                                                                                       |
| `ephemeral=true/false`                        | N/A                                                                                       |
| `fields=[list of fields]`                     | [extensions.sumologic.collector_fields](#fields)                                          |
| `fipsJce=true/false`                          | N/A                                                                                       |
| `hostName=hostname`                           | [processors.source.source_host][source-templates]                                         |
| `name=name`                                   | [extensions.sumologic.collector_name](#name)                                              |
| `proxyHost=host`                              | [please see OTC documentation][proxy]                                                     |
| `proxyNtlmDomain=NTLM domain`                 | [please see OTC documentation][proxy]                                                     |
| `proxyPassword=password`                      | [please see OTC documentation][proxy]                                                     |
| `proxyPort=port`                              | [please see OTC documentation][proxy]                                                     |
| `proxyUser=username`                          | [please see OTC documentation][proxy]                                                     |
| `skipAccessKeyRemoval=true/false`             | N/A                                                                                       |
| `sources=absolute filepath or folderpath`     | [Use --config flag](/docs/configuration.md#command-line-configuration-options)            |
| `syncSources=absolute filepath or folderpath` | [Use --config flag](/docs/configuration.md#command-line-configuration-options)            |
| `targetCPU=target`                            | N/A                                                                                       |
| `timeZone=timezone`                           | [extensions.sumologic.time_zone](#time-zone)                                              |
| `token=token`                                 | [extensions.sumologic.installation_token](/docs/configuration.md#authentication)          |
| `url=collection endpoint`                     | [extensions.sumologic.api_base_url][sumologicextension]                                   |
| `wrapper.java.command=JRE Bin Location`       | N/A                                                                                       |
| `wrapper.java.command=JRE Bin Location`       | N/A                                                                                       |
| `wrapper.java.maxmemory=size`                 | N/A                                                                                       |

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
- [RemoteWindowsEventLog](#remote-windows-event-log-source-remotewindowseventlog)
- [LocalWindowsPerfMon](#local-windows-performance-source-localwindowsperfmon)
- [RemoteWindowsPerfMon](#remote-windows-performance-source-remotewindowsperfmon)
- [ActiveDirectory](#windows-active-directory-source-activedirectory)

| The Installed Collector Parameter | The OpenTelemetry Collector Key                                                                                 |
|-----------------------------------|-----------------------------------------------------------------------------------------------------------------|
| `name`                            | [processors.source.source_name](#name-2)                                                                        |
| `description`                     | A description can be added as a comment just above the receiver name. [See the linked example.](#description-2) |
| `fields`                          | Use [Transform Processor][transformprocessor] to set custom fields. [See the linked example.](#fields-1)        |
| `hostName`                        | [processors.source.source_host][source-templates]; [See the linked example.](#source-host-1)                    |
| `category`                        | [processors.source.source_category][source-templates]                                                           |
| `automaticDateParsing`            | [See Timestamp Parsing explanation](#timestamp-parsing-1)                                                       |
| `timeZone`                        | [See Timestamp Parsing explanation](#timestamp-parsing-1)                                                       |
| `forceTimeZone`                   | [See Timestamp Parsing explanation](#timestamp-parsing-1)                                                       |
| `defaultDateFormat`               | [See Timestamp Parsing explanation](#timestamp-parsing-1)                                                       |
| `defaultDateFormats`              | [See Timestamp Parsing explanation](#timestamp-parsing-1)                                                       |
| `multilineProcessingEnabled`      | [See Multiline Processing explanation](#multiline-processing)                                                   |
| `useAutolineMatching`             | [See Multiline Processing explanation](#multiline-processing)                                                   |
| `manualPrefixRegexp`              | [See Multiline Processing explanation](#multiline-processing)                                                   |
| `filters`                         | [See filtering explanation](#filtering)                                                                         |
| `cutoffTimestamp`                 | [Use Filter Processor](#collection-should-begin)                                                                |
| `cutoffRelativeTime`              | N/A                                                                                                             |

#### Filtering

##### Include and exclude

[Filter processor][filterprocessor] can be used to filter data that will be sent to Sumo Logic.
Like in case of Installed Collector's filters, you can `include` or `exclude` data.
Below is an example of how to filter data.

```yaml
processors:
  filter/bodies:
    logs:
      exclude:
        match_type: regexp

        ## Exclude logs that start with "unimportant-log-prefix"
        bodies:
          - unimportant-log-prefix.*

    metrics:
      include:
        match_type: regexp

        ## Include metrics that end with "_percentage"
        metric_names:
          - .*_percentage

        ## Include metrics that have a field "server.endpoint" with value "desired-endpoint"
        resource_attributes:
          - key: server.endpoint
            value: desired-endpoint
      exclude:
        match_type: strict

        ## Exclude metrics that have names "dummy_metric" or "another_one"
        metric_names:
          - dummy_metric
          - another_one
```

All possible options for filters can be found in the [Filter processor's documentation][filterprocessor].

##### Masks

To achieve the [Mask filter][mask-filter], you can use [Transform processor][transformprocessor].
It can be used on log bodies, attributes and fields. Below is an example config of that processor.

```yaml
processors:
  transform:
    metric_statements:
      ## Replace EVERY attribute that matches a regex.
      - context: datapoint
        statements:
          - replace_all_matches(attributes, ".*password", "***")

    trace_statements:
      ## Replace a particular field if the value matches.
      ## For patterns, you can use regex capture groups by using regexp.Expand syntax: https://pkg.go.dev/regexp#Regexp.Expand
      - context: resource
        statements:
          - replace_pattern(attributes["name"], "^kubernetes_([0-9A-Za-z]+_)", "k8s.$$1.")

      ## Replace an attribute with value matching a regex.
      - context: span
        statements:
          - replace_match(attributes["http.target"], "/user/*/list/*", "/user/{userId}/list/{listId}")

    log_statements:
      ## Replace sensitive data inside log body
      - context: log
        statements:
          - replace_pattern(body, "token - ([a-zA-Z0-9]{32})", "token - ***")
```

All possible options for masking can be seen in the [Transform processor's documentation][transformprocessor]
and [the documentation of OTTL functions][ottlfuncs] (functions `replace_match`, `replace_all_matches`, `replace_pattern`, `replace_all_patterns`).

##### Data forwarding

To forward data to some other place than `Sumo Logic` (for example, a syslog server or a generic http API),
you should define exporters that will send data to desired places and use them in the same pipeline as Sumo Logic exporter:

```yaml
receivers:
  filelog:
    ## ...

exporters:
  sumologic:
    ## ...

  syslog:
    protocol: tcp # or udp
    port: 6514 # 514 (UDP)
    endpoint: 127.0.0.1 # FQDN or IP address
    tls:
      ca_file: certs/servercert.pem
      cert_file: certs/cert.pem
      key_file: certs/key.pem
    format: rfc5424 # rfc5424 or rfc3164

  ## Note: otlphttp exporter sends data only in otlp format
  otlphttp:
    logs_endpoint: http://example.com:4318/v1/logs


service:
  pipelines:
    logs:
      receivers: [filelog]
      processors: []
      exporters: [sumologic, syslog, otlphttp]
```

Out of [methods available in Installed Collector][forward-data], only Syslog is available now in the OpenTelemetry Collector.

### Local File Source (LocalFile)

The equivalent of the Local File Source is [the filelog receiver][filelogreceiver].
More useful information can be found in [Local File Source for Cloud Based Management](#local-file-source).

| The Installed Collector Parameter | The OpenTelemetry Collector Key                         |
|-----------------------------------|---------------------------------------------------------|
| `pathExpression`                  | element of [receivers.filelog.include](#file-path) list |
| `denylist`                        | [receivers.filelog.exclude](#denylist)                  |
| `encoding`                        | [receivers.filelog.encoding](#encoding)                 |

### Remote File Source (RemoteFileV2)

Remote File Source is not supported by the OpenTelemetry Collector.

### Syslog Source (Syslog)

The equivalent of the Syslog Source is a combination of
[the TCP][tcplogreceiver] or [the UDP][udplogreceiver] receivers and [the sumologicsyslog processor][sumologicsyslog].

| The Installed Collector Parameter | The OpenTelemetry Collector Key                                                                                      |
|-----------------------------------|----------------------------------------------------------------------------------------------------------------------|
| `protocol`                        | using tcplog or udplog receiver. [See syslog explanation](#protocol-and-port)                                        |
| `port`                            | `receivers.tcplog.listen_address` or `receivers.udplog.listen_address`. [See syslog explanation](#protocol-and-port) |

### Docker Logs Source (DockerLog)

Docker Logs Source (DockerLog) is not supported by the OpenTelemetry Collector.
Refer to [Docker Logs Source](#docker-logs-source) for details.

### Docker Stats Source (DockerStats)

The equivalent of the Docker Stats Source is [the Docker Stats receiver][dockerstatsreceiver].
More useful information can be found in [Docker Stats Source for Cloud Based Management](#docker-stats-source).

| The Installed Collector Parameter | The OpenTelemetry Collector Key                                                                    |
|-----------------------------------|----------------------------------------------------------------------------------------------------|
| contentType                       | N/A                                                                                                |
| metrics                           | [receivers.docker_stats.metrics](#metrics)                                                         |
| uri                               | [receivers.docker_stats.endpoint](#uri)                                                            |
| specifiedContainers               | [processors.filter.metrics.include](#container-filters)                                            |
| allContainers                     | N/A, list of containers can be controlled by using [processors.filter.metrics](#container-filters) |
| certPath                          | N/A                                                                                                |
| pollInterval                      | [receivers.docker_stats.collection_interval](#scan-interval)                                       |

### Script Source (Script)

Script Source is not supported by the OpenTelemetry Collector.

### Streaming Metrics Source (StreamingMetrics)

The equivalent of the Streaming Metrics Source is [the telegraf receiver][telegrafreceiver] with appropiate plugins.
More useful information can be found in [Streaming Metrics Source for Cloud Based Management](#streaming-metrics-source).

| The Installed Collector Parameter | The OpenTelemetry Collector Key                                                                                 |
|-----------------------------------|-----------------------------------------------------------------------------------------------------------------|
| `name`                            | [processors.source.source_name](#name-4)                                                                        |
| `description`                     | A description can be added as a comment just above the receiver name. [See the linked example.](#description-4) |
| `category`                        | [processors.source.source_category](#source-category-3)                                                         |
| `contentType`                     | [receivers.telegraf.agent_config('inputs.socket_listener'.data_format)](#content-type)                          |
| `protocol`                        | [receivers.telegraf.agent_config('inputs.socket_listener'.service_address)](#protocol-and-port-1)               |
| `port`                            | [receivers.telegraf.agent_config('inputs.socket_listener'.service_address)](#protocol-and-port-1)               |

### Host Metrics Source (SystemStats)

The equivalent of the Host Metrics Source is [the telegraf receiver][telegrafreceiver] with appropiate plugins.
More useful information can be found in [Host Metrics Source for Cloud Based Management](#host-metrics-source).

__Note: The are differences between the Installed Collector and the Openelemetry Collector host metrics.
See [this document](comparison.md#host-metrics) to learn more.__

| The Installed Collector Parameter | The OpenTelemetry Collector Key                                                                                 |
|-----------------------------------|-----------------------------------------------------------------------------------------------------------------|
| `name`                            | [processors.source.source_name](#name-5)                                                                        |
| `description`                     | A description can be added as a comment just above the receiver name. [See the linked example.](#description-5) |
| `category`                        | [processors.source.source_category](#source-category-4)                                                         |
| `metrics`                         | [Appropiate plugins have to be configured.](#metrics-1) By default no metrics are being processed.                |
| `interval (ms)`                   | [receivers.telegraf.agent_config('agent'.interval)](#scan-interval-1)                                             |
| `hostName`                        | [processors.source.source_host](#source-host-3)                                                                 |

### Local Windows Event Log Source (LocalWindowsEventLog)

There is no migration process from Installed Collector to OpenTelemetry Collector.
In order to use OpenTelemetry Collector, dedicated Sumo Logic app needs to be
installed.

### Remote Windows Event Log Source (RemoteWindowsEventLog)

Remote Windows Event Log Source is not supported by the OpenTelemetry Collector.

### Local Windows Performance Source (LocalWindowsPerfMon)

There is no migration process from Installed Collector to OpenTelemetry Collector.

### Remote Windows Performance Source (RemoteWindowsPerfMon)

Remote Windows Performance Source is not supported by the OpenTelemetry Collector.

### Windows Active Directory Source (ActiveDirectory)

Windows Active Directory Source is not supported by the OpenTelemetry Collector.

[resourceprocessor]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.134.0/processor/resourceprocessor
[multiline]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.134.0/receiver/filelogreceiver#multiline-configuration
[supported_encodings]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.134.0/receiver/filelogreceiver#supported-encodings
[udplogreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.134.0/receiver/udplogreceiver
[tcplogreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.134.0/receiver/tcplogreceiver
[filelogreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.134.0/receiver/filelogreceiver
[syslogreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.134.0/receiver/syslogreceiver
[transformprocessor]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.134.0/processor/transformprocessor
[filterprocessor]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.134.0/processor/filterprocessor
[sumologicsyslog]: ../pkg/processor/sumologicsyslogprocessor/README.md
[network-semantic-convention]: https://github.com/open-telemetry/semantic-conventions/blob/cee22ec91448808ebcfa53df689c800c7171c9e1/docs/general/attributes.md#other-network-attributes
[sumologicextension]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.134.0/extension/sumologicextension/README.md
[sumologicexporter]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.134.0/exporter/sumologicexporter/README.md
[syslogexporter]: https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/v0.134.0/exporter/syslogexporter/README.md
[user.properties]: https://help.sumologic.com/docs/send-data/installed-collectors/collector-installation-reference/user-properties
[proxy]: https://opentelemetry.io/docs/collector/configuration/#proxy-support
[common-parameters]: https://help.sumologic.com/docs/send-data/use-json-configure-sources#common-parameters-for-log-source-types
[source-templates]: ../pkg/processor/sourceprocessor//README.md#source-templates
[syslogparser]: https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/pkg/stanza/docs/operators/syslog_parser.md
[telegrafreceiver]: ../pkg/receiver/telegrafreceiver/README.md
[telegraf-socket_listener]: https://github.com/SumoLogic/telegraf/tree/v1.24.3-sumo-4/plugins/inputs/socket_listener#socket-listener-input-plugin
[telegraf-input-formats]: https://github.com/SumoLogic/telegraf/tree/v1.24.3-sumo-4/plugins/parsers
[telegraf-input-plugins]: https://github.com/SumoLogic/telegraf/tree/v1.24.3-sumo-4/plugins/inputs
[telegraf-input-cpu]: https://github.com/SumoLogic/telegraf/tree/v1.24.3-sumo-4/plugins/inputs/cpu
[telegraf-input-system]: https://github.com/SumoLogic/telegraf/tree/v1.24.3-sumo-4/plugins/inputs/system
[telegraf-input-mem]: https://github.com/SumoLogic/telegraf/tree/v1.24.3-sumo-4/plugins/inputs/mem
[telegraf-input-net]: https://github.com/SumoLogic/telegraf/tree/v1.24.3-sumo-4/plugins/inputs/net/README.md
[telegraf-input-netstat]: https://github.com/SumoLogic/telegraf/tree/v1.24.3-sumo-4/plugins/inputs/netstat/README.md
[telegraf-input-diskio]: https://github.com/SumoLogic/telegraf/tree/v1.24.3-sumo-4/plugins/inputs/diskio
[telegraf-input-disk]: https://github.com/SumoLogic/telegraf/tree/v1.24.3-sumo-4/plugins/inputs/disk
[dockerstatsreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/v0.134.0/receiver/dockerstatsreceiver
[dockerstatsmetrics]: https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/v0.134.0/receiver/dockerstatsreceiver/documentation.md
[sumologicschemaprocessor]: ../pkg/processor/sumologicschemaprocessor/README.md
[mask-filter]: https://help.sumologic.com/docs/send-data/use-json-configure-sources/#example-mask-filter
[ottlfuncs]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.134.0/pkg/ottl/ottlfuncs#functions
[forward-data]: https://help.sumologic.com/docs/manage/data-archiving/installed-collectors/
[ingest-budget]: https://help.sumologic.com/docs/manage/ingestion-volume/ingest-budgets/
