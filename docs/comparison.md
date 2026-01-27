# Comparison between the Installed Collector and OpenTelemetry Collector

## When to use OpenTelemetry

**Ideal use cases** include when you:

- leverage the **Supported Sources** and **Supported Platforms** listed below
- are looking for a single agent as opposed to managing multiple agents
- are having scale issues with FluentD on Kubernetes Collection
- want to use an open-source, vendor-neutral observability framework with broad community support
- need unified collection of logs, metrics, and traces in a single agent

**Avoid use cases** that:

- are using an **Unsupported Source**
- require CPU target

## Support Matrix

<table>
  <tr>
   <td><strong>Supported Platforms</strong>
   </td>
   <td><strong>Unsupported Platforms</strong>
   </td>
  </tr>
  <tr>
   <td>
<ul>
  <li>Linux
  <li>MacOS
  <li>Kubernetes
  <li>Windows
  </li>
</ul>
   </td>
   <td>
  </td>
  </tr>
  <tr>
   <td><strong>Supported Sources</strong>
   </td>
   <td><strong>Unsupported Sources</strong>
   </td>
  </tr>
  <tr>
  <td>
<ul>

<li>Local File
<li>Syslog
<li>Host/Process Metrics
<li>Streaming Metrics
<li>Transaction Tracing
<li>All Telegraf Input Plugins
<li>Docker Stats / Logs
<li>Script Sources
<li>Windows Log Event Receiver
<li>Windows Performance Counters Receiver
</li>
</ul>
   </td>
   <td>
<ul>
<li>Script Actions
<li>Remote File
<li>Windows Active Directory Source
<li>Remote Windows Event Log Source
</li>
</ul>
   </td>
  </tr>
</table>

## Source specific considerations

- [Syslog](#syslog)
  - [Syslog Receiver](#syslog-receiver)
  - [TCPlog/UDPlog Receiver and Sumo Logic Syslog Processor](#tcplogudplog-receiver-and-sumo-logic-syslog-processor)
- [Host Metrics](#host-metrics)

### Syslog

The OpenTelemetry Collector offers two approaches for syslog processing:

- Syslog Receiver
- TCPlog/UDPlog Receiver and Sumo Logic Syslog Processor

Read this section to learn about the differences.

#### Syslog Receiver

Syslog Receiver is a perfect solution if you are sending logs using a certain RFC protocol.
There are two supported formats: `rfc3164` and `rfc5424`.
Parsing is strict, if you send `rfc5424` logs to the `rfc3164` endpoint,
it will fail with an error and the log (and timestamp as well) won't be parsed.

For example, with the following configuration:

```yaml
receivers:
  syslog:
    tcp:
      listen_address: "0.0.0.0:54526"
    protocol: rfc5424
exporters:
  debug:
    verbosity: detailed
service:
  pipelines:
    logs:
      receivers: [syslog]
      exporters: [debug]
```

and the following example logs:

```text
<14>Apr  19 09:50:00 mymachineICIP su: RFC3164
<34>1 2021-04-09T07:54:14.001Z mymachine.example.com su - - - RFC5424 | TCP
```

it will produce the following logs:

```text
...
2021-08-24T12:55:43.323+0200    error   Failed to process entry {"kind": "receiver", "name": "syslog", "operator_id": "$..syslog_parser", "operator_type": "syslog_parser", "error": "expecting a version value in the range 1-999 [col 4]", "action": "send", "entry": {"timestamp":"2021-08-24T12:55:43.323699582+02:00","body":"<14>Apr  19 09:50:00 mymachineICIP su: RFC3164","severity":0}}
2021-08-24T12:55:43.374+0200    INFO    loggingexporter/logging_exporter.go:71  LogsExporter    {"#logs": 1}
2021-08-24T12:55:43.374+0200    DEBUG   loggingexporter/logging_exporter.go:81  ResourceLog #0
InstrumentationLibraryLogs #0
InstrumentationLibrary
LogRecord #0
Timestamp: 2021-08-24 10:55:43.323699582 +0000 UTC
Severity: Undefined
ShortName:
Body: <14>Apr  19 09:50:00 mymachineICIP su: RFC3164

2021-08-24T12:55:55.173+0200    INFO    loggingexporter/logging_exporter.go:71  LogsExporter    {"#logs": 1}
2021-08-24T12:55:55.174+0200    DEBUG   loggingexporter/logging_exporter.go:81  ResourceLog #0
InstrumentationLibraryLogs #0
InstrumentationLibrary
LogRecord #0
Timestamp: 2021-04-09 07:54:14.001 +0000 UTC
Severity: Error2
ShortName:
Body: {
     -> appname: STRING(su)
     -> facility: INT(4)
     -> hostname: STRING(mymachine.example.com)
     -> message: STRING(RFC5424 | TCP)
     -> priority: INT(34)
     -> version: INT(1)
}
...
```

#### TCPlog/UDPlog Receiver and Sumo Logic Syslog Processor

This second approach is compatible with the current Installed Collector behavior.
It doesn't parse out the fields on the collector side,
but extracts the facility name and sends it as the `Source Name`.
In addition, it doesn't verify the protocol of incoming logs,
so every format is treated the same.

For example, with the following configuration:

```yaml
extensions:
  sumologic:
    installation_token: <token>
receivers:
  tcplog:
    listen_address: "0.0.0.0:54526"
    add_attributes: true

processors:
  sumologic_syslog: {}
  groupbyattrs:
    keys:
    - net.peer.name
    - facility

exporters:
  sumologic:
    ## Set Source Name to facility name
    source_name: "%{facility}"
    ## Set Source Host to client hostname
    source_host: "%{net.peer.name}"
  debug:
    verbosity: detailed
service:
  extensions: [sumologic]
  pipelines:
    logs:
      receivers: [tcplog]
      processors: [sumologic_syslog, groupbyattrs]
      exporters: [debug, sumologic]
```

and the following example logs:

```text
<14>Apr  19 09:50:00 mymachineICIP su: RFC3164
<34>1 2021-04-09T07:54:14.001Z mymachine.example.com su - - - RFC5424 | TCP
```

it will produce the following logs:

```text
2021-08-24T13:18:41.464+0200    INFO    loggingexporter/logging_exporter.go:71  LogsExporter    {"#logs": 1}
2021-08-24T13:18:41.464+0200    DEBUG   loggingexporter/logging_exporter.go:81  ResourceLog #0
Resource labels:
     -> net.peer.name: STRING(localhost)
     -> facility: STRING(user-level messages)
InstrumentationLibraryLogs #0
InstrumentationLibrary
LogRecord #0
Timestamp: 2021-08-24 11:18:41.394337919 +0000 UTC
Severity: Undefined
ShortName:
Body: <14>Apr  19 09:50:00 mymachineICIP su: RFC3164
Attributes:
     -> net.transport: STRING(IP.TCP)
     -> net.peer.ip: STRING(127.0.0.1)
     -> net.peer.port: STRING(56790)
     -> net.host.name: STRING(localhost)
     -> net.host.ip: STRING(127.0.0.1)
     -> net.host.port: STRING(54526)

2021-08-24T13:18:42.854+0200    INFO    loggingexporter/logging_exporter.go:71  LogsExporter    {"#logs": 1}
2021-08-24T13:18:42.854+0200    DEBUG   loggingexporter/logging_exporter.go:81  ResourceLog #0
Resource labels:
     -> net.peer.name: STRING(localhost)
     -> facility: STRING(security/authorization messages)
InstrumentationLibraryLogs #0
InstrumentationLibrary
LogRecord #0
Timestamp: 2021-08-24 11:18:41.653010477 +0000 UTC
Severity: Undefined
ShortName:
Body: <34>1 2021-04-09T07:54:14.001Z mymachine.example.com su - - - RFC5424 | TCP
Attributes:
     -> net.transport: STRING(IP.TCP)
     -> net.peer.ip: STRING(127.0.0.1)
     -> net.peer.port: STRING(56790)
     -> net.host.name: STRING(localhost)
     -> net.host.ip: STRING(127.0.0.1)
     -> net.host.port: STRING(54526)
```

### Host Metrics

The Installed Collector and OpenTelemetry Collector have different
codebases that cause some host metrics to have different names.

Use the `translate_telegraf_attributes` option in the
[sumologic processor](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.144.0/processor/sumologicprocessor)
to keep metric names compatible with Sumo Logic Apps and consistent with the Installed Collector.

There may be differences between metric values calculated by the Installed Collector
and OpenTelemetry Collector since they use different calculation formulas.

The following list has the metrics gathered by the Installed Collector
that don't have equivalents in the OpenTelemetry Collector:

- `Mem_Used`
- `Mem_PhysicalRam`
- `TCP_InboundTotal`
- `TCP_OutboundTotal`
- `TCP_Idle`
- `Disk_Queue`
- `Disk_Available`

### Windows collection

The OpenTelemetry collector currently supports the following kinds of Windows data.

#### Windows Log Event Receiver

This receiver tails and parses logs from the windows event log API.

##### Configuration Fields

| Field           | Default  | Description                                                                                                                                                                                                                                    |
|-----------------|----------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `channel`       | required | The windows event log channel to monitor                                                                                                                                                                                                       |
| `max_reads`     | 100      | The maximum number of records read into memory, before beginning a new batch                                                                                                                                                                   |
| `start_at`      | `end`    | On first startup, where to start reading logs from the API. Options are `beginning` or `end`                                                                                                                                                   |
| `poll_interval` | 1s       | The interval at which the channel is checked for new log entries. This check begins again after all new bodies have been read.                                                                                                                 |
| `attributes`    | {}       | A map of `key: value` pairs to add to the entry's attributes.                                                                                                                                                                                  |
| `resource`      | {}       | A map of `key: value` pairs to add to the entry's resource.                                                                                                                                                                                    |
| `operators`     | []       | An array of [operators](https://github.com/open-telemetry/opentelemetry-log-collection/blob/main/docs/operators/README.md#what-operators-are-available). See below for more details                                                            |
| `raw`           | false    | If true, the windows events are not processed and sent as XML.                                                                                                                                                                                 |
| `storage`       | none     | The ID of a storage extension to be used to store bookmarks. Bookmarks allow the receiver to pick up where it left off in the case of a collector restart. If no storage extension is used, the receiver will manage bookmarks in memory only. |

For more information, please refer to the [windows event log receiver][windowseventlogreceiver] readme.

#### Windows Performance Counters Receiver

This receiver, for Windows only, captures the configured system, application, or custom performance counter data from the Windows registry using the [PDH interface](https://learn.microsoft.com/en-us/windows/win32/perfctrs/using-the-pdh-functions-to-consume-counter-data). It is based on the [Telegraf Windows Performance Counters Input Plugin](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/win_perf_counters)

- `Memory\Committed Bytes`
- `Processor\% Processor Time`, with a datapoint for each `Instance` label = (`_Total`, `1`, `2`, `3`, ... )

If one of the specified performance counters cannot be loaded on startup, a
warning will be printed, but the application will not fail fast. It is expected
that some performance counters may not exist on some systems due to different OS
configuration.

##### Configuration

The collection interval and the list of performance counters to be scraped can
be configured:

```yaml
windowsperfcounters:
  collection_interval: <duration> # default = "1m"
  metrics:
    <metric name>:
      description: <description>
      unit: <unit type>
      gauge:
    <metric name>:
      description: <description>
      unit: <unit type>
      sum:
        aggregation: <cumulative or delta>
        monotonic: <true or false>
  perfcounters:
    - object: <object name>
      instances: [<instance name>]*
      counters:
        - name: <counter name>
          metric: <metric name>
          attributes:
            <key>: <value>
```

For more information, please refer to the [windows performance counter receiver][windowsperfcountersreceiver] readme.

[windowsperfcountersreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/windowsperfcountersreceiver/README.md
[windowseventlogreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/windowseventlogreceiver/README.md
