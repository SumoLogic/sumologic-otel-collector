# Configuration

- [Extensions](#extensions)
  - [Sumo Logic Extension](#sumo-logic-extension)
    - [Using multiple Sumo Logic extensions](#using-multiple-sumo-logic-extensions)
- [Receivers](#receivers)
  - [Filelog Receiver](#filelog-receiver)
  - [Fluent Forward Receiver](#fluent-forward-receiver)
  - [Syslog Receiver](#syslog-receiver)
  - [Statsd Receiver](#statsd-receiver)
  - [Telegraf Receiver](#telegraf-receiver)
  - [OTLP Receiver](#otlp-receiver)
  - [Receivers from OpenTelemetry Collector](#receivers-from-opentelemetry-collector)
- [Processors](#processors)
- [Exporters](#exporters)

---

## Extensions

### Sumo Logic Extension

To send data to [Sumo Logic][sumologic_webpage] you need to configure
the [sumologicextension][sumologicextension] with credentials and define it
(the extension) in the same service as the [sumologicexporter][sumologicexporter]
is defined so that it's used as an auth extension.

The following configuration is a basic example to collect CPU load metrics using 
the [Host Metrics Receiver][hostmetricsreceiver] and send them to Sumo Logic:

```yaml
extensions:
  sumologic:
    access_id: <my_access_id>
    access_key: <my_access_key>
    collector_name: <my_collector_name>

receivers:
  hostmetrics:
    collection_interval: 30s
    scrapers:
      load:

exporters:
  sumologic:

service:
  extensions: [sumologic]
  pipelines:
    metrics:
      receivers: [hostmetrics]
      exporters: [sumologic]
```

For a list of all the configuration options for sumologicextension refer to
[this documentation][sumologicextension_configuration].

[sumologic_webpage]: https://www.sumologic.com/
[sumologicextension]: ../pkg/extension/sumologicextension/
[sumologicexporter]: ../pkg/exporter/sumologicexporter/
[hostmetricsreceiver]: https://github.com/SumoLogic/opentelemetry-collector/tree/release-0.27/receiver/hostmetricsreceiver
[sumologicextension_configuration]: ../pkg/extension/sumologicextension#configuration

#### Using multiple Sumo Logic extensions

If you want to register multiple collectors and/or send data to
mutiple Sumo Logic accounts, mutiple `sumologicextension`s can be defined within the
pipeline and used in exporter definitions.

In this case, you need to specify a custom authenticator name that points to
the correct extension ID.

Example:

```yaml
extensions:
  sumologic/custom_auth1:
    access_id: <my_access_id1>
    access_key: <my_access_key1>
    collector_name: <my_collector_name1>

  sumologic/custom_auth2:
    access_id: <my_access_id2>
    access_key: <my_access_key2>
    collector_name: <my_collector_name2>

receivers:
  hostmetrics:
    collection_interval: 30s
    scrapers:
      load:
  filelog:
    include: [ "**.log" ]

exporters:
  sumologic/custom1:
    auth:
      authenticator: sumologic/custom_auth1
  sumologic/custom2:
    auth:
      authenticator: sumologic/custom_auth2

service:
  extensions: [sumologic/custom_auth1, sumologic/custom_auth2]
  pipelines:
    metrics/1:
      receivers: [hostmetrics]
      exporters: [sumologic/custom1]
    logs/1:
      receivers: [filelog]
      exporters: [sumologic/custom2]
```

---

## Receivers

### Filelog Receiver

The Filelog Receiver tails and parses logs from files using the [opentelemetry-log-collection][opentelemetry-log-collection] library.

The following is a basic configuration for the Filelog Receiver:

```yaml
receivers:
  filelog:
    include: [ /var/log/myservice/*.json ]
    operators:
      - type: json_parser
        timestamp:
          parse_from: time
          layout: '%Y-%m-%d %H:%M:%S'
```

For details, see the [Filelog Receiver documentation][filelogreceiver_readme].

[opentelemetry-log-collection]: https://github.com/open-telemetry/opentelemetry-log-collection
[filelogreceiver_readme]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.27.0/receiver/filelogreceiver

### Fluent Forward Receiver

The Fluent Forward Receiver runs a TCP server that accepts events via the [Fluent Forward
protocol][fluent_forward_protocol].

The basic configuration for Fluent Forward Receiver has following format:

```yaml
receivers:
  fluentforward:
    endpoint: 0.0.0.0:8006
```

For details see the [Fluent Forward Receiver documentation][fluentforwardreceiver_readme].

[fluent_forward_protocol]: https://github.com/fluent/fluentd/wiki/Forward-Protocol-Specification-v1
[fluentforwardreceiver_readme]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/release/v0.27.x/receiver/fluentforwardreceiver

### Syslog Receiver

The Syslog Receiver parses Syslogs from tcp/udp using
the [opentelemetry-log-collection](https://github.com/open-telemetry/opentelemetry-log-collection) library.

The following is a basic example for the Syslog Receiver with a TCP configuration:

```yaml
receivers:
  syslog:
    tcp:
      listen_address: "0.0.0.0:54526"
    protocol: rfc5424
```

The following is a basic example for the Syslog Receiver with a UDP Configuration:

```yaml
receivers:
  syslog:
    udp:
      listen_address: "0.0.0.0:54526"
    protocol: rfc3164
    location: UTC
```

For details see the [Syslog Receiver documentation][syslogreceiver_readme].

[syslogreceiver_readme]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.27.0/receiver/syslogreceiver

### Statsd Receiver

The StatsD Receiver ingests [StatsD messages][statsd_messages] into the OpenTelemetry Collector.

The following is a basic configuration for the StatsD Receiver:

```yaml
receivers:
  statsd:
  statsd/2:
    endpoint: "localhost:8127"
    aggregation_interval: 70s
    enable_metric_type: true
    timer_histogram_mapping:
      - statsd_type: "histogram"
        observer_type: "gauge"
      - statsd_type: "timing"
        observer_type: "gauge"
```

For details see the [StatsD Receiver documentation][statsdreceiver_readme].

[statsd_messages]: https://github.com/statsd/statsd/blob/master/docs/metric_types.md
[statsdreceiver_readme]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/release/v0.27.x/receiver/statsdreceiver

### Telegraf Receiver

The Telegraf Receiver ingests metrics from various [input plugins][input_plugins]
into the OTC pipeline.

The following is a basic configuration for the Telegraf Receiver:

```yaml
receivers:
  telegraf:
    separate_field: false
    agent_config: |
      [agent]
        interval = "3s"
        flush_interval = "3s"
      [[inputs.mem]]
```

For details see the [Telegraf Receiver documentation][telegrafreceiver_readme].

[input_plugins]: https://github.com/influxdata/telegraf/tree/master/plugins/inputs
[telegrafreceiver_readme]: ../pkg/receiver/telegrafreceiver

### OTLP Receiver

The OTLP Receiver receives data via gRPC or HTTP using [OTLP][otlp] format.

The following is a basic configuration for the OTLP Receiver:

```yaml
receivers:
  otlp:
    protocols:
      grpc:
      http:
```

For details see the [OTLP Receiver documentation][otlpreceiver_readme]

[otlp]: https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/protocol/otlp.md
[otlpreceiver_readme]: https://github.com/open-telemetry/opentelemetry-collector/tree/v0.27.0/receiver/otlpreceiver

### Receivers from OpenTelemetry Collector

The Sumo Logic OT Distro has built-in receivers from the [OpenTelemetry Collector](https://github.com/SumoLogic/opentelemetry-collector) and are allowed in the configuration for this distribution.

The following is an example configuration to collect CPU load metrics using the [Host Metrics Receiver][hostmetricsreceiver]:

```yaml
receivers:
  hostmetrics:
    collection_interval: 30s
    scrapers:
      load:
```

For details see the [receiver documentation][opentelemetry-collector-receivers].

[hostmetricsreceiver]: https://github.com/SumoLogic/opentelemetry-collector/tree/release-0.27/receiver/hostmetricsreceiver
[opentelemetry-collector]: https://github.com/SumoLogic/opentelemetry-collector/tree/release-0.27
[opentelemetry-collector-receivers]: https://github.com/SumoLogic/opentelemetry-collector/tree/release-0.27/receiver

---

## Processors

---

## Exporters
