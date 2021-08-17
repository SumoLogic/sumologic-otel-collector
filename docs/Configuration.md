# Configuration

- [Basic configuration](#basic-configuration)
  - [Basic configuration for logs](#basic-configuration-for-logs)
  - [Basic configuration for metrics](#basic-configuration-for-metrics)
  - [Basic configuration for traces](#basic-configuration-for-traces)
- [Putting it all together](#putting-it-all-together)
- [Extensions](#extensions)
  - [Sumo Logic Extension](#sumo-logic-extension)
    - [Using multiple Sumo Logic extensions](#using-multiple-sumo-logic-extensions)
- [Receivers](#receivers)
  - [Sumo Logic Custom Receivers](#sumo-logic-custom-receivers)
    - [Telegraf Receiver](#telegraf-receiver)
  - [Open Telemetry Upstream Receivers](#open-telemetry-upstream-receivers)
    - [Filelog Receiver](#filelog-receiver)
    - [Fluent Forward Receiver](#fluent-forward-receiver)
    - [Syslog Receiver](#syslog-receiver)
    - [Statsd Receiver](#statsd-receiver)
    - [OTLP Receiver](#otlp-receiver)
    - [TCPlog Receiver](#tcplog-receiver)
    - [UDPlog Receiver](#udplog-receiver)
    - [Receivers from OpenTelemetry Collector](#receivers-from-opentelemetry-collector)
- [Processors](#processors)
  - [Sumo Logic Custom Processors](#sumo-logic-custom-processors)
    - [Cascading Filter Processor](#cascading-filter-processor)
    - [Kubernetes Processor](#kubernetes-processor)
    - [Source Processor](#source-processor)
    - [Sumo Logic Syslog Processor](#sumo-logic-syslog-processor)
  - [Open Telemetry Upstream Processors](#open-telemetry-upstream-processors)
    - [Group by Attributes Processor](#group-by-attributes-processor)
    - [Group by Trace Processor](#group-by-trace-processor)
    - [Metrics Transform Processor](#metrics-transform-processor)
    - [Resource Detection Processor](#resource-detection-processor)
    - [Routing Processor](#routing-processor)
    - [Span Metrics Processor](#span-metrics-processor)
    - [Tail Sampling Processor](#tail-sampling-processor)
- [Exporters](#exporters)
  - [Sumo Logic Custom Exporters](#sumo-logic-custom-exporters)
    - [Sumo Logic Exporter](#sumo-logic-exporter)
  - [Open Telemetry Upstream Exporters](#open-telemetry-upstream-exporters)
    - [Load Balancing Exporter](#load-balancing-exporter)
- [Command-line configuration options](#command-line-configuration-options)

---

## Basic configuration

The only required option to run the collector is the `--config` option that points to the configuration file.

```shell
otelcol-sumo --config config.yaml
```

For all the available command line options, see [Command-line configuration options](#command-line-configuration-options).

The file `config.yaml` is a regular OpenTelemetry Collector configuration file
that contains a pipeline with some receivers, processors and exporters.
If you are new to OpenTelemetry Collector,
you can familiarize yourself with the terms reading the [upstream documentation](https://opentelemetry.io/docs/collector/configuration/).

The primary components that make it easy to send data to Sumo Logic are
the [Sumo Logic Exporter][sumologicexporter_docs]
and the [Sumo Logic Extension][sumologicextension_configuration].

Here's a starting point for the configuration file that you will want to use:

```yaml
exporters:
  sumologic:

extensions:
  sumologic:
    access_id: <my_access_id>
    access_key: <my_access_key>

receivers:
  ... # fill in receiver configurations here

service:
  extensions: [sumologic]
  pipelines:
    logs:
      receivers: [...] # fill in logs receiver names here
      exporters: [sumologic]
    metrics:
      receivers: [...] # fill in metrics receiver names here
      exporters: [sumologic]
    traces:
      receivers: [...] # fill in trace receiver names here
      exporters: [sumologic]
```

The Sumo Logic exporter automatically detects the Sumo Logic extension
if it's added in the `service.extensions` property
and uses it as the authentication provider to connect and send data to the Sumo Logic backend.

You add the receivers for the data you want to be collected
and put them together in one pipeline.
You can of course also add other components according to your needs -
extensions, processors, other exporters etc.

Let's look at some examples for configuring logs, metrics and traces to be sent to Sumo,
and after that let's put that all together.

### Basic configuration for logs

To send logs from local files, use the [Filelog Receiver][filelogreceiver_readme].

Example configuration:

```yaml
exporters:
  sumologic:

extensions:
  sumologic:
    access_id: <my_access_id>
    access_key: <my_access_key>

receivers:
  filelog:
    include:
    - /var/log/myservice/*.log
    - /other/path/**/*.txt

service:
  extensions: [sumologic]
  pipelines:
    logs:
      receivers: [filelog]
      exporters: [sumologic]
```

See [Receivers](#receivers) section for sending data from other sources including Fluentd/Fluent Bit, syslog and others.

### Basic configuration for metrics

Sumo Logic OT Distro uses the Telegraf Receiver to ingest metrics.

Here's a minimal `config.yaml` file that sends the host's memory metrics to Sumo Logic:

```yaml
exporters:
  sumologic:

extensions:
  sumologic:
    access_id: <my_access_id>
    access_key: <my_access_key>

receivers:
  telegraf:
    separate_field: false
    agent_config: |
      [agent]
        interval = "3s"
        flush_interval = "3s"
      [[inputs.mem]]

service:
  extensions: [sumologic]
  pipelines:
    metrics:
      receivers: [telegraf]
      exporters: [sumologic]
```

### Basic configuration for traces

Use the [OTLP Receiver][otlpreceiver_readme] to send traces to Sumo Logic.

Example configuration:

```yaml
exporters:
  sumologic:

extensions:
  sumologic:
    access_id: <my_access_id>
    access_key: <my_access_key>

receivers:
  otlp:
    protocols:
      grpc:

service:
  extensions: [sumologic]
  pipelines:
    traces:
      receivers: [otlp]
      exporters: [sumologic]
```

## Putting it all together

Here's an example configuration file that collects all the signals - logs, metrics and traces.

```yaml
exporters:
  sumologic:

extensions:
  sumologic:
    access_id: <my_access_id>
    access_key: <my_access_key>


receivers:
  filelog:
    include:
    - /var/log/myservice/*.log
    - /other/path/**/*.txt
  telegraf:
    separate_field: false
    agent_config: |
      [agent]
        interval = "3s"
        flush_interval = "3s"
      [[inputs.mem]]
  otlp:
    protocols:
      grpc:

service:
  extensions: [sumologic]
  pipelines:
    logs:
      receivers: [filelog]
      exporters: [sumologic]
    metrics:
      receivers: [telegraf]
      exporters: [sumologic]
    traces:
      receivers: [otlp]
      exporters: [sumologic]
```

See below for details on configuring all the components available in the Sumo Logic OT Distro -
extensions, receivers, processors, exporters.

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

### Sumo Logic Custom Receivers

The following receivers have been developed by Sumo Logic.

#### Telegraf Receiver

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

For details, see the [Telegraf Receiver documentation][telegrafreceiver_readme].

[input_plugins]: https://github.com/influxdata/telegraf/tree/master/plugins/inputs
[telegrafreceiver_readme]: ../pkg/receiver/telegrafreceiver

### Open Telemetry Upstream Receivers

The following receivers have been developed by the Open Telemetry community
and are incorporated into the Sumo Logic Open Telemetry distro without any changes.

If you are already familiar with Open Telemetry, you may know how the upstream components work
and you can expect no changes in their behaviour.

#### Filelog Receiver

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
[filelogreceiver_readme]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.31.0/receiver/filelogreceiver

#### Fluent Forward Receiver

The Fluent Forward Receiver runs a TCP server that accepts events via the [Fluent Forward
protocol][fluent_forward_protocol].

The basic configuration for Fluent Forward Receiver has following format:

```yaml
receivers:
  fluentforward:
    endpoint: 0.0.0.0:8006
```

For details, see the [Fluent Forward Receiver documentation][fluentforwardreceiver_readme].

[fluent_forward_protocol]: https://github.com/fluent/fluentd/wiki/Forward-Protocol-Specification-v1
[fluentforwardreceiver_readme]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/release/v0.27.x/receiver/fluentforwardreceiver

#### Syslog Receiver

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

For details, see the [Syslog Receiver documentation][syslogreceiver_readme].

[syslogreceiver_readme]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.31.0/receiver/syslogreceiver

#### Statsd Receiver

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

For details, see the [StatsD Receiver documentation][statsdreceiver_readme].

[statsd_messages]: https://github.com/statsd/statsd/blob/master/docs/metric_types.md
[statsdreceiver_readme]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/release/v0.27.x/receiver/statsdreceiver

#### OTLP Receiver

The OTLP Receiver receives data via gRPC or HTTP using [OTLP][otlp] format.

The following is a basic configuration for the OTLP Receiver:

```yaml
receivers:
  otlp:
    protocols:
      grpc:
      http:
```

For details, see the [OTLP Receiver documentation][otlpreceiver_readme].

[otlp]: https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/protocol/otlp.md
[otlpreceiver_readme]: https://github.com/open-telemetry/opentelemetry-collector/tree/v0.31.0/receiver/otlpreceiver

#### TCPlog Receiver

The TCPlog Receiver receives logs data via TCP using text format.

The following is a basic configuration for the TCPlog Receiver:

```yaml
receivers:
  tcplog:
    listen_address: "0.0.0.0:54525"
```

For details, see the [TCPlog Receiver documentation][tcplogreceiver_readme].

[tcplogreceiver_readme]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.31.0/receiver/tcplogreceiver

#### UDPlog Receiver

The UDPlog Receiver receives logs data via UDP using text format.

The following is a basic configuration for the UDPlog Receiver:

```yaml
receivers:
  udplog:
    listen_address: "0.0.0.0:54525"
```

For details, see the [UDPlog Receiver documentation][udplogreceiver_readme].

[udplogreceiver_readme]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.31.0/receiver/udplogreceiver

#### Receivers from OpenTelemetry Collector

The Sumo Logic OT Distro has built-in receivers from the [OpenTelemetry Collector](https://github.com/SumoLogic/opentelemetry-collector) and are allowed in the configuration for this distribution.

The following is an example configuration to collect CPU load metrics using the [Host Metrics Receiver][hostmetricsreceiver]:

```yaml
receivers:
  hostmetrics:
    collection_interval: 30s
    scrapers:
      load:
```

For details, see the [receiver documentation][opentelemetry-collector-receivers].

[hostmetricsreceiver]: https://github.com/SumoLogic/opentelemetry-collector/tree/release-0.27/receiver/hostmetricsreceiver
[opentelemetry-collector]: https://github.com/SumoLogic/opentelemetry-collector/tree/release-0.27
[opentelemetry-collector-receivers]: https://github.com/SumoLogic/opentelemetry-collector/tree/release-0.27/receiver

---

## Processors

### Sumo Logic Custom Processors

The following processors have been developed by Sumo Logic
either from scratch (like the [Sumo Logic Syslog Processor](#sumo-logic-syslog-processor))
or as a customized version of upstream processor (like the [Kubernetes Processor](#kubernetes-processor)).

#### Cascading Filter Processor

The Cascading Filter Processor is a trace sampling processor
that allows to define smart cascading filtering rules with preset limits.

Example configuration:

```yaml
processors:
  cascading_filter:
    decision_wait: 10s
    num_traces: 100
    expected_new_traces_per_sec: 10
    spans_per_second: 1000
    probabilistic_filtering_ratio: 0.1
    policies:
      [
        {
          name: test-policy-1,
          spans_per_second: 35,
        },
        {
          name: test-policy-2,
          spans_per_second: 50,
          properties: { min_duration: 9s }
        }
      ]
```

For details, see the [Cascading Filter Processor documentation][cascadingfilterprocessor_docs].

[cascadingfilterprocessor_docs]: https://github.com/SumoLogic/opentelemetry-collector-contrib/blob/main/processor/cascadingfilterprocessor/README.md

#### Kubernetes Processor

The Kubernetes Processor adds Kubernetes-specific metadata to traces, metrics and logs
by querying the Kubernetes cluster's API server.

This is a Sumo Logic fork of the [upstream k8sprocessor][upstream_k8sprocessor].

Example configuration:

```yaml
processors:
  k8s_tagger:
    extract:
      metadata:
        - hostName
        - startTime
      tags:
        hostName: hostname
```

For details, see the [Kubernetes Processor documentation][k8sprocessor_docs].

[upstream_k8sprocessor]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/k8sprocessor
[k8sprocessor_docs]: https://github.com/SumoLogic/opentelemetry-collector-contrib/blob/main/processor/k8sprocessor/README.md

#### Source Processor

The Source Processor adds Sumo Logic-specific source metadata like `_source`, `_sourceCategory` etc.
to traces, metrics and logs.

Example configuration:

```yaml
processors:
  source:
    collector: "mycollector"
    source_name: "%{namespace}.%{pod}.%{container}"
    source_category: "%{namespace}/%{pod_name}"
    source_category_prefix: "kubernetes/"
    source_category_replace_dash: "/"
    exclude_namespace_regex: "kube-system"
```

For details, see the [Source Processor documentation][sourceprocessor_docs].

[sourceprocessor_docs]: https://github.com/SumoLogic/opentelemetry-collector-contrib/blob/main/processor/sourceprocessor/README.md

#### Sumo Logic Syslog Processor

The Sumo Logic Syslog Processor tries to extract facility code from syslog logs
and adds the facility's name as a metadata attribute.

Example configuration:

```yaml
processors:
  sumologic_syslog:
    facility_attr: syslog.facility.name
```

For details, see the [Sumo Logic Syslog Processor documentation][sumologicsyslogprocessor_docs].

[sumologicsyslogprocessor_docs]: https://github.com/SumoLogic/opentelemetry-collector-contrib/blob/main/processor/sumologicsyslogprocessor/README.md

### Open Telemetry Upstream Processors

The following processors have been developed by the Open Telemetry community
and are incorporated into the Sumo Logic Open Telemetry distro without any changes.

If you are already familiar with Open Telemetry, you may know how the upstream components work
and you can expect no changes in their behaviour.

#### Group by Attributes Processor

The Group by Attributes Processor groups records by provided attributes, extracting them from the record to resource level.

Example configuration:

```yaml
processors:
  groupbyattrs:
    keys:
      - host.name
```

For details, see the [Group by Attributes Processor documentation][groupbyattrsprocessor_docs].

[groupbyattrsprocessor_docs]: https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/processor/groupbyattrsprocessor/README.md

#### Group by Trace Processor

The Group by Trace Processor tries to collect all spans in a trace
before releasing that trace for further processing in the collector pipeline.

Example configuration:

```yaml
processors:
  groupbytrace:
    wait_duration: 10s
    num_traces: 1000
```

For details, see the [Group by Trace Processor documentation][groupbytraceprocessor_docs].

[groupbytraceprocessor_docs]: https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/processor/groupbytraceprocessor/README.md

#### Metrics Transform Processor

The Metrics Transform Processor can be used to rename metrics, and add, rename or delete label keys and values.

Example configuration:

```yaml
processors:
  metricstransform:
    transforms:
      # rename system.cpu.usage to system.cpu.usage_time
      - include: system.cpu.usage
        action: update
        new_name: system.cpu.usage_time
```

For details, see the [Metrics Transform Processor documentation][metrictransformprocessor_docs].

[metrictransformprocessor_docs]: https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/processor/groupbytraceprocessor/README.md

#### Resource Detection Processor

The Resource Detection Processor detects resource information from runtime environment
and adds metadata with this information to the traces, metrics and logs.

Example configuration:

```yaml
processors:
  resourcedetection:
    detectors: ["eks", "ecs", "ec2"]
```

For details, see the [Resource Detection Processor documentation][resourcedetectionprocessor_docs].

[resourcedetectionprocessor_docs]: https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/processor/resourcedetectionprocessor/README.md

#### Routing Processor

The Routing Processor does not alter the records in any way by itself.
It routes records to specific exporters based based on an attribute's value.

Note that this should be the last processor in the pipeline.

Example configuration:

```yaml
processors:
  routing:
    from_attribute: X-Tenant
    default_exporters: jaeger
    table:
    - value: acme
      exporters: [jaeger/acme]
exporters:
  jaeger:
    endpoint: localhost:14250
  jaeger/acme:
    endpoint: localhost:24250
```

For details, see the [Routing Processor documentation][routingprocessor_docs].

[routingprocessor_docs]: https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/processor/routingprocessor/README.md

#### Span Metrics Processor

The Span Metrics Processor aggregates request, error and duration (R.E.D) metrics from span data.

Example configuration:

```yaml
receivers:
  jaeger:
    protocols:
      thrift_http:
        endpoint: "0.0.0.0:14278"

  # Dummy receiver that's never used, because a pipeline is required to have one.
  otlp/dummy:
    protocols:
      grpc:
        endpoint: "localhost:12345"

exporters:
  prometheus:
    endpoint: "0.0.0.0:8889"
    namespace: promexample

  jaeger:
    endpoint: "localhost:14250"
    insecure: true

processors:
  batch:
  spanmetrics:
    metrics_exporter: prometheus
    latency_histogram_buckets: [100us, 1ms, 2ms, 6ms, 10ms, 100ms, 250ms]
    dimensions:
      - name: http.method
        default: GET
      - name: http.status_code

service:
  pipelines:
    traces:
      receivers: [jaeger]
      # spanmetrics will pass on span data untouched to next processor
      # while also accumulating metrics to be sent to the configured 'prometheus' exporter.
      processors: [spanmetrics, batch]
      exporters: [jaeger]

    metrics:
      # This receiver is just a dummy and never used.
      # Added to pass validation requiring at least one receiver in a pipeline.
      receivers: [otlp/dummy]
      # The metrics_exporter must be present in this list.
      exporters: [prometheus]
```

For details, see the [Span Metrics Processor documentation][spanmetricsprocessor_docs].

[spanmetricsprocessor_docs]: https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/processor/spanmetricsprocessor/README.md

#### Tail Sampling Processor

The Tail Sampling Processor samples traces based on a set of defined policies.

Example configuration:

```yaml
processors:
  tail_sampling:
    decision_wait: 10s
    num_traces: 100
    expected_new_traces_per_sec: 10
    policies:
      [
          {
            name: test-policy-1,
            type: always_sample
          },
          {
            name: test-policy-2,
            type: rate_limiting,
            rate_limiting: {spans_per_second: 35}
         }
      ]
```

For details, see the [Tail Sampling Processor documentation][tailsamplingprocessor_docs].

[tailsamplingprocessor_docs]: https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/processor/tailsamplingprocessor/README.md

## Exporters

### Sumo Logic Custom Exporters

The following exporters have been developed by Sumo Logic.

#### Sumo Logic Exporter

The Sumo Logic Exporter supports sending data to [Sumo Logic](https://www.sumologic.com/).

Example configuration with using the [Sumo Logic Extension](#sumo-logic-extension) for authentication
and setting a custom source category:

```yaml
exporters:
  sumologic:
    auth:
      authenticator: sumologic
    source_category: my-category

extensions:
  sumologic:
    access_id: <my_access_id>
    access_key: <my_access_key>
```

For details, see the [Sumo Logic Exporter documentation][sumologicexporter_docs].

[sumologicexporter_docs]: ../pkg/exporter/sumologicexporter/README.md

### Open Telemetry Upstream Exporters

The following exporters have been developed by the Open Telemetry community
and are incorporated into the Sumo Logic Open Telemetry distro without any changes.

If you are already familiar with Open Telemetry, you may know how the upstream components work
and you can expect no changes in their behaviour.

#### Load Balancing Exporter

The Load Balancing Exporter consistently exports spans and logs belonging to the same trace to the same backend.

Example configuration:

```yaml
exporters:
  loadbalancing:
    protocol:
      otlp:
        # all options from the OTLP exporter are supported
        # except the endpoint
        timeout: 1s
    resolver:
      static:
        hostnames:
        - backend-1:4317
        - backend-2:4317
        - backend-3:4317
        - backend-4:4317
```

For details, see the [Load Balancing Exporter documentation][loadbalancingexporter_docs].

[loadbalancingexporter_docs]: https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/exporter/loadbalancingexporter/README.md

## Command-line configuration options

```bash
Usage:
  otelcol-sumo [flags]

Flags:
      --add-instance-id             Flag to control the addition of 'service.instance.id' to the collector metrics. (default true)
      --config string               Path to the config file
  -h, --help                        help for otelcol-sumo
      --log-format string           Format of logs to use (json, console) (default "console")
      --log-level Level             Output level of logs (DEBUG, INFO, WARN, ERROR, DPANIC, PANIC, FATAL) (default info)
      --log-profile string          Logging profile to use (dev, prod) (default "prod")
      --mem-ballast-size-mib uint   Flag to specify size of memory (MiB) ballast to set. Ballast is not used when this is not specified. default settings: 0
      --metrics-addr string         [address]:port for exposing collector telemetry. (default ":8888")
      --metrics-level Level         Output level of telemetry metrics (none, basic, normal, detailed) (default basic)
      --metrics-prefix string       Prefix to the metrics generated by the collector. (default "otelcol")
      --set stringArray             Set arbitrary component config property. The component has to be defined in the config file and the flag has a higher precedence. Array config properties are overridden and maps are joined, note that only a single (first) array property can be set e.g. -set=processors.attributes.actions.key=some_key. Example --set=processors.batch.timeout=2s (default [])
  -v, --version                     version for otelcol-sumo
```
