# Configuration

- [Basic configuration](#basic-configuration)
  - [Alternative: OpAmp remote-managed configuration](#alternative-opamp-remote-managed-configuration)
  - [Basic configuration for logs](#basic-configuration-for-logs)
  - [Basic configuration for metrics](#basic-configuration-for-metrics)
  - [Basic configuration for traces](#basic-configuration-for-traces)
  - [Putting it all together](#putting-it-all-together)
- [Authentication](#authentication)
  - [Running the collector as systemd service](#running-the-collector-as-systemd-service)
  - [Using multiple Sumo Logic extensions](#using-multiple-sumo-logic-extensions)
- [Persistence](#persistence)
- [Collecting logs from files](#collecting-logs-from-files)
  - [Keeping track of position in files](#keeping-track-of-position-in-files)
  - [Parsing JSON logs](#parsing-json-logs)
- [Setting source category](#setting-source-category)
  - [Setting source category on logs from files](#setting-source-category-on-logs-from-files)
  - [Setting source category on Prometheus metrics](#setting-source-category-on-prometheus-metrics)
  - [Setting source category with the Resource processor](#setting-source-category-with-the-resource-processor)
- [Setting source host](#setting-source-host)
- [Command-line configuration options](#command-line-configuration-options)
- [Proxy Support](#proxy-support)
- [Keeping Prometheus format using OTLP exporter](#keeping-prometheus-format-using-otlp-exporter)

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

To create the `installation_token` please follow [this documenation][sumologic_docs_installation_token].

Here's a starting point for the configuration file that you will want to use:

```yaml
exporters:
  sumologic:

extensions:
  sumologic:
    installation_token: ${SUMOLOGIC_INSTALLATION_TOKEN}

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

> **IMPORTANT NOTE**:
> It is recommended to limit access to the configuration file as it contains sensitive information.
> You can change access permissions to the configuration file using:
>
> ```bash
> chmod 640 config.yaml
> ```

[sumologic_docs_installation_token]: https://help.sumologic.com/docs/manage/security/installation-tokens

### Alternative: OpAmp remote-managed configuration

If you would like to leverage OpAmp for remote configuration management, then
you can use the `--remote-config` flag instead of the `--config` flag.

```shell
otelcol-sumo --remote-config opamp:remote-config.yaml
```

The `--remote-config` and `--config` flags are mutually exclusive and cannot be
used together. The remote configuration details are loaded from `remote-config.yaml`
and used to connect to the configured OpAmp server.

The `remote-config.yaml` also specifies the local directory that will be used
by the OpAmp client to keep configuration synchronized with the server's
prescribed configuration for the collector.

```yaml
extensions:
  sumologic:
    installation_token: ${SUMOLOGIC_INSTALLATION_TOKEN}
  opamp:
    endpoint: "wss://example.com/v1/opamp"
    remote_configuration_directory: /etc/opamp-config
```

### Basic configuration for logs

To send logs from local files, use the [Filelog Receiver][filelogreceiver_readme].

Example configuration:

```yaml
exporters:
  sumologic:
    sending_queue:
      enabled: true
      storage: file_storage

extensions:
  file_storage:
    directory: .
  sumologic:
    installation_token: ${SUMOLOGIC_INSTALLATION_TOKEN}

receivers:
  filelog:
    include:
    - /var/log/myservice/*.log
    - /other/path/**/*.txt
    include_file_name: false
    include_file_path_resolved: true
    operators:
    - type: move
      from: attributes["log.file.path_resolved"]
      to: resource["_sourceName"]
    start_at: beginning

service:
  extensions: [file_storage, sumologic]
  pipelines:
    logs:
      receivers: [filelog]
      exporters: [sumologic]
```

Adding the [File Storage extension][filestorageextension_docs] allows the Filelog receiver
to persist the position in the files it reads between restarts.

See section below on [Collecting logs from files](#collecting-logs-from-files) for details on configuring the Filelog receiver.

[filestorageextension_docs]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.134.0/extension/storage/filestorage

### Basic configuration for metrics

Sumo Logic Distribution for OpenTelemetry Collector uses the [Host Metrics Receiver][hostmetricsreceiver_docs] to ingest metrics.

Here's a minimal `config.yaml` file that sends the host's memory metrics to Sumo Logic:

```yaml
exporters:
  sumologic:

extensions:
  sumologic:
    installation_token: ${SUMOLOGIC_INSTALLATION_TOKEN}

receivers:
  hostmetrics:
    collection_interval: 30s
    scrapers:
      cpu:
      memory:

service:
  extensions: [sumologic]
  pipelines:
    metrics:
      receivers: [hostmetrics]
      exporters: [sumologic]
```

[hostmetricsreceiver_docs]: https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/v0.134.0/receiver/hostmetricsreceiver/README.md

### Basic configuration for traces

Use the [OTLP Receiver][otlpreceiver_readme] to ingest traces.

Example configuration:

```yaml
exporters:
  sumologic:

extensions:
  sumologic:
    installation_token: ${SUMOLOGIC_INSTALLATION_TOKEN}

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

[otlpreceiver_readme]: https://github.com/open-telemetry/opentelemetry-collector/tree/v0.134.0/receiver/otlpreceiver

### Putting it all together

Here's an example configuration file that collects all the signals - logs, metrics and traces.

```yaml
exporters:
  sumologic:
    sending_queue:
      enabled: true
      storage: file_storage

extensions:
  file_storage:
    directory: .
  sumologic:
    installation_token: ${SUMOLOGIC_INSTALLATION_TOKEN}

receivers:
  filelog:
    include:
    - /var/log/myservice/*.log
    - /other/path/**/*.txt
    include_file_name: false
    include_file_path_resolved: true
    operators:
    - type: move
      from: attributes["log.file.path_resolved"]
      to: resource["_sourceName"]
    start_at: beginning
  hostmetrics:
    collection_interval: 30s
    scrapers:
      cpu:
      memory:
  otlp:
    protocols:
      grpc:

service:
  extensions: [file_storage, sumologic]
  pipelines:
    logs:
      receivers: [filelog]
      exporters: [sumologic]
    metrics:
      receivers: [hostmetrics]
      exporters: [sumologic]
    traces:
      receivers: [otlp]
      exporters: [sumologic]
```

## Authentication

To send data to [Sumo Logic][sumologic_webpage] you need to configure
the [sumologicextension][sumologicextension] with credentials and define it
(the extension) in the same service as the [sumologicexporter][sumologicexporter]
is defined so that it's used as an auth extension.

The following configuration is a basic example to collect CPU load metrics using
the [Host Metrics Receiver][hostmetricsreceiver] and send them to Sumo Logic:

```yaml
extensions:
  sumologic:
    installation_token: ${SUMOLOGIC_INSTALLATION_TOKEN}
    collector_name: ${MY_COLLECTOR_NAME}

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

### Running the collector as systemd service

The credentials are stored on local filesystem (by default in `$HOME/.sumologic-otel-collector`) to be reused when collector gets restarted.
However, systemd services are often run as users without a home directory,
so keep in mind that to store credentials either the user needs a home directory or the store location should be explicitely changed.

More information about this feature can be found in the [extension's documentation][sumologicextension_store_credentials].

[sumologic_webpage]: https://www.sumologic.com/
[sumologicextension]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.134.0/extension/sumologicextension
[sumologicexporter]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.134.0/exporter/sumologicexporter
[hostmetricsreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.134.0/receiver/hostmetricsreceiver
[sumologicextension_configuration]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.134.0/extension/sumologicextension#configuration
[sumologicextension_store_credentials]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.134.0/extension/sumologicextension#storing-credentials

### Using multiple Sumo Logic extensions

If you want to register multiple collectors and/or send data to
mutiple Sumo Logic accounts, mutiple `sumologicextension`s can be defined within the
pipeline and used in exporter definitions.

In this case, you need to specify a custom authenticator name that points to
the correct extension ID.

Example:

```yaml
extensions:
  sumologic/custom_auth1:
    installation_token: ${SUMOLOGIC_INSTALLATION_TOKEN_1}
    collector_name: <COLLECTOR_NAME_1>

  sumologic/custom_auth2:
    installation_token: ${SUMOLOGIC_INSTALLATION_TOKEN_2}
    collector_name: <COLLECTOR_NAME_2>

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

## Persistence

When using the [Sumo Logic exporter][sumologicexporter_docs], it is recommended to store its state in a persistent storage
to prevent loss of data buffered in the exporter between restarts.

To do that, add the [File Storage extension][filestorageextension_docs] to the configuration
and configure the exporter to use persistent queue
with the `sending_queue.enabled: true`
and `sending_queue.storage: <storage_name>` flags.

Here's an example configuration:

```yaml
exporters:
  sumologic:
    sending_queue:
      enabled: true
      storage: file_storage

extensions:
  file_storage:
    directory: .
  sumologic:
    installation_token: ${SUMOLOGIC_INSTALLATION_TOKEN}

receivers:
  hostmetrics:
    collection_interval: 3s
    scrapers:
      load:

service:
  extensions:
  - file_storage
  - sumologic
  pipelines:
    metrics:
      exporters:
      - sumologic
      receivers:
      - hostmetrics
```

[sumologicexporter_docs]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.134.0/exporter/sumologicexporter/README.md

## Collecting logs from files

The Filelog receiver tails and parses logs from files on local file system.

The following is a basic configuration for the Filelog receiver:

```yaml
receivers:
  filelog:
    include:
    - /var/log/myservice/*.log
    - /other/path/**/*.txt
    include_file_name: false
    include_file_path_resolved: true
    operators:
    - type: move
      from: attributes["log.file.path_resolved"]
      to: resource["_sourceName"]
```

The `include_file_name: false` prevents the receiver from adding `log.file.name` attribute to the logs.
Instead, we are using `include_file_path_resolved: true`,
which adds a `log.file.path_resolved` attribute to the logs
that contains the whole path of the file, as opposed to just the name of the file.
The configuration then uses the `move` operator to move the record-level attribute `log.file.path_resolved`
to resource-level attribute `_sourceName`, so that it is displayed as source name in Sumo Logic.

### Keeping track of position in files

By default, the Filelog receiver watches files starting at their end
(`start_at: end` is the [default][filelogreceiver_readme]),
so nothing will be read after the otelcol process starts until new data is added to the files.
To change this, add `start_at: beginning` to the receiver's configuration.
To prevent the receiver from reading the same data over and over again on each otelcol restart,
also add the [File Storage extension][filestorageextension_docs] that will allow Filelog receiver to persist the current
position in watched files between otelcol restarts. Here's an example of such configuration:

```yaml
extensions:
  file_storage:
    directory: .

receivers:
  filelog:
    include:
    - /var/log/myservice/*.log
    - /other/path/**/*.txt
    include_file_name: false
    include_file_path_resolved: true
    operators:
    - type: move
      from: attributes["log.file.path_resolved"]
      to: resource["_sourceName"]
    start_at: beginning
```

For more details, see the [Filelog Receiver documentation][filelogreceiver_readme].

### Parsing JSON logs

Filelog Receiver with [json_parser][json_parser] operator can be used for parsing JSON logs.
The [json_parser][json_parser] operator parses the string-type field selected by `parse_from` as JSON
(by default `parse_from` is set to `$body` which indicates the whole log record).

For example when logs has following form in the file:

```json
{"message": "{\"key\": \"val\"}"}
```

then configuration to extract JSON which is represented as string (`{\"key\": \"val\"}`) has following form:

```yaml
receivers:
  filelog:
    include:
    - /log/path/**/*.log
    operators:
      - type: json_parser   # this parses log line as JSON
      - type: json_parser   # this parses string under 'message' key as JSON
        parse_from: message
```

and the parsed log entry can be observed in [Debug exporter][debugexporter_docs]'s output as:

```console
2022-02-24T10:23:37.809Z        INFO    loggingexporter/logging_exporter.go:69  LogsExporter    {"#logs": 1}
2022-02-24T10:23:37.809Z        DEBUG   loggingexporter/logging_exporter.go:79  ResourceLog #0
Resource SchemaURL:
InstrumentationLibraryLogs #0
InstrumentationLibraryMetrics SchemaURL:
InstrumentationLibrary
LogRecord #0
Timestamp: 2022-02-24 10:23:37.714661255 +0000 UTC
Severity:
ShortName:
Body: {
     -> key: STRING(val)
}
Attributes:
     -> log.file.name: STRING(example.log)
Trace ID:
Span ID:
Flags: 0
```

Example configuration with example log can be found in [/examples/otelcolconfigs/logs_json/](/examples/otelcolconfigs/logs_json/) directory.

[json_parser]: https://github.com/open-telemetry/opentelemetry-log-collection/blob/main/docs/operators/json_parser.md
[filelogreceiver_readme]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.134.0/receiver/filelogreceiver
[debugexporter_docs]: https://github.com/open-telemetry/opentelemetry-collector/tree/v0.134.0/exporter/debugexporter

## Setting source category

For many Sumo Logic customers the [source category][source_category_docs] is a crucial piece of metadata
that allows to differentiate data coming from different sources.
This section describes how to configure the Sumo Logic Distribution of OpenTelemetry
to decorate data with this attribute.

To decorate all the data from the collector with the same source category,
set the `source_category` property on the [source processor][source_proc] like this:

```yaml
processors:
  source:
    source_category: my/source/category
```

You can also use the [source templates][source_proc_templates]
to use other attributes' values to compose source category, like this:

```yaml
processors:
  source:
    source_category: "%{k8s.namespace.name}/%{k8s.pod.pod_name}"
```

However, please bear in mind that the source processor is going to be removed in the future.

If you want data from different sources to have different source categories,
you'll need to set a resource attribute named `_sourceCategory` earlier in the pipeline.
See below for examples on how to do this in various scenarios.

[source_category_docs]: https://help.sumologic.com/docs/send-data/reference-information/metadata-naming-conventions#source-categories
[source_proc]: ../pkg/processor/sourceprocessor
[source_proc_templates]: ../pkg/processor/sourceprocessor/README.md#source-templates

### Setting source category on logs from files

The [Filelog receiver][filelogreceiver_readme] allows to set attributes on received data.
Here's an example on how to attach different source categories to logs from different files
with the Filelog receiver:

```yaml
exporters:
  sumologic:

receivers:

  filelog/apache:
    include:
    - /logs/from/apache.txt
    resource:
      _sourceCategory: apache-logs

  filelog/nginx:
    include:
    - /logs/from/nginx.txt
    resource:
      _sourceCategory: nginx-logs

service:
  pipelines:
    logs:
      exporters:
      - sumologic
      receivers:
      - filelog/apache
      - filelog/nginx
```

### Setting source category on Prometheus metrics

The [Prometheus receiver][prometheusreceiver_docs] uses the same config
as the full-blown [Prometheus][prometheus_website], so you can use its [relabel config][prometheus_relabel_config]
to create the `_sourceCategory` label.

```yaml
exporters:
  sumologic:

receivers:
  prometheus:
    config:
      scrape_configs:

      - job_name: mysql-metrics
        relabel_configs:
        - replacement: db-metrics
          target_label: _sourceCategory
        static_configs:
        - targets:
          - "0.0.0.0:9104"

      - job_name: otelcol-metrics
        relabel_configs:
        - source_labels:
          - job
          target_label: _sourceCategory
        static_configs:
        - targets:
          - "0.0.0.0:8888"

service:
  pipelines:
    metrics:
      exporters:
      - sumologic
      receivers:
      - prometheus
```

The first example creates a `_sourceCategory` label with a hardcoded value of `db-metrics`.

The second example creates a `_sourceCategory` label by copying to it the value of Prometheus' `job` label,
which contains the name of the job - in this case, `otelcol-metrics`.

[prometheusreceiver_docs]: https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/v0.134.0/receiver/prometheusreceiver/README.md
[prometheus_website]: https://prometheus.io/
[prometheus_relabel_config]: https://prometheus.io/docs/prometheus/latest/configuration/configuration/#relabel_config

### Setting source category with the Resource processor

If the receiver does not allow to set attributes on the received data,
use the [Resource processor][resourceprocessor_docs] to add the `_sourceCategory` attribute
later in the pipeline.
Here's an example on how to get statsd metrics from two different apps to have different source categories.

```yaml
exporters:
  sumologic:

processors:

  resource/one-source-category:
    attributes:
    - action: insert
      key: _sourceCategory
      value: one-app

  resource/another-source-category:
    attributes:
    - action: insert
      key: _sourceCategory
      value: another-app

receivers:

  statsd/one-app:
    endpoint: "localhost:8125"

  statsd/another-app:
    endpoint: "localhost:8126"

service:
  pipelines:
    metrics/one-app:
      exporters:
      - sumologic
      processors:
      - resource/one-app
      receivers:
      - statsd/one-app

    metrics/another-app:
      exporters:
      - sumologic
      processors:
      - resource/another-app
      receivers:
      - statsd/another-app
```

[resourceprocessor_docs]: https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/v0.134.0/processor/resourceprocessor/README.md

## Setting source host

You can use the source processor's `source_host` property
to set the [Sumo Logic source host][sumologic_source_host_docs] attribute
to a static value like this:

```yaml
processors:
  source:
    source_host: my-host-name
```

But this is most likely not what you want.
You'd rather have the collector retrieve the name of the host from the operating system,
instead of needing to manually hardcode it in the config.

This is what the [Resource Detection processor][resourcedetectionprocessor_docs] does.
Use its built-in [`system` detector][resourcedetectionprocessor_system_detector]
to set the OpenTelemetry standard `host.name` resource attribute
to the name of the host that the collector is running on.
After that is set, you need to add the `_sourceHost` attribute
with the value from the `host.name` attribute.

```yaml
exporters:
  sumologic:

processors:
  resource/add-source-host:
    attributes:
    - action: insert
      key: _sourceHost
      from_attribute: host.name
    # Optionally, delete the original attributes created by the Resource Detection processor.
    - action: delete
      key: host.name
    - action: delete
      key: os.type

  resourcedetection/detect-host-name:
    detectors:
    - system
    system:
      hostname_sources:
      - os

receivers:
  hostmetrics:
    scrapers:
      memory:

service:
  pipelines:
    metrics:
      exporters:
      - sumologic
      processors:
      - resourcedetection/detect-host-name
      - resource/add-source-host
      receivers:
      - hostmetrics
```

Make sure to put the Resource Detection processor *before* the Resource processor in the pipeline
so that the `host.name` attribute is already set in the `resource` processor.

Only the first Resource processor's action is required to correctly set the `_sourceHost` attribute.
The other two actions perform an optional metadata cleanup - they delete the unneeded attributes.

[sumologic_source_host_docs]: https://help.sumologic.com/docs/send-data/reference-information/metadata-naming-conventions#source-host
[resourcedetectionprocessor_docs]: https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/v0.134.0/processor/resourcedetectionprocessor/README.md
[resourcedetectionprocessor_system_detector]: https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/v0.134.0/processor/resourcedetectionprocessor/README.md#system-metadata

## Command-line configuration options

```console
Usage:
  otelcol-sumo [flags]

Flags:
      --config -config=file:/path/to/first --config=file:path/to/second   Locations to the config file(s), note that only a single pattern can be set per flag entry e.g. -config=file:/path/to/single_file --config="glob:path/with/glob/*". (default [])
      --feature-gates Flag                                                Comma-delimited list of feature gate identifiers. Prefix with '-' to disable the feature.  '+' or no prefix will enable the feature.
  -h, --help                                                              help for otelcol-sumo
      --set stringArray                                                   Set arbitrary component config property. The component has to be defined in the config file and the flag has a higher precedence. Array config properties are overridden and maps are joined, note that only a single (first) array property can be set e.g. -set=processors.attributes.actions.key=some_key. Example --set=processors.batch.timeout=2s (default [])
  -v, --version                                                           version for otelcol-sumo
```

## Proxy Support

Exporters leverage the HTTP communication and respect the following proxy environment variables:

- `HTTP_PROXY`
- `HTTPS_PROXY`
- `NO_PROXY`

You may either export proxy environment variables locally e.g.

```bash
export FTP_PROXY=<PROXY-ADDRESS>:<PROXY-PORT>
export HTTP_PROXY=<PROXY-ADDRESS>:<PROXY-PORT>
export HTTPS_PROXY=<PROXY-ADDRESS>:<PROXY-PORT>
```

or make them available globally for all users, e.g.

```bash
tee -a /etc/profile << END
export FTP_PROXY=<PROXY-ADDRESS>:<PROXY-PORT>
export HTTP_PROXY=<PROXY-ADDRESS>:<PROXY-PORT>
export HTTPS_PROXY=<PROXY-ADDRESS>:<PROXY-PORT>
END
```

## Keeping Prometheus format using OTLP exporter

In order to keep the [Prometheus compatible metric names][prometheus_data_model] using OTLP exporting format,
[the Metrics Transform Processor][metricstransformprocessor] can be used.

Please see the following example of replacing last period char with underscore:

```yaml
processors:
  metricstransform:
    transforms:
      ## Replace last period char in metric name with underscore
      - include: ^(.*)\.([^\.]*)$$
        match_type: regexp
        action: update
        new_name: $${1}_$${2}
# ...
service:
  pipelines:
    metrics:
      receivers:
        - telegraf
      processors:
        - metricstransform
      exporters:
        - sumologic
# ...
```

[metricstransformprocessor]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.134.0/processor/metricstransformprocessor
[prometheus_data_model]: https://prometheus.io/docs/concepts/data_model/#metric-names-and-labels
