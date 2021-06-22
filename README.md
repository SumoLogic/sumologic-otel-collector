# sumologic-otel-collector

[![Default branch build](https://github.com/SumoLogic/sumologic-otel-collector/actions/workflows/dev_builds.yml/badge.svg)](https://github.com/SumoLogic/sumologic-otel-collector/actions/workflows/dev_builds.yml)

SumoLogic Distro of [OpenTelemetry Collector][otc_link] built with
[opentelemetry-collector-builder][otc_builder_link].

[otc_link]: https://github.com/open-telemetry/opentelemetry-collector
[otc_builder_link]: https://github.com/open-telemetry/opentelemetry-collector-builder

## Configuration

In order to send data to [Sumo Logic][sumologic_webpage] one needs to configure
the [sumologicextension][sumologicextension] with credentials and define it (the extension) in the same service
as [sumologicexporter][sumologicexporter] is defined so that it's used as auth
extension.

```yaml
extensions:
  sumologic:
    access_id: <my_access_id>
    access_key: <my_access_key>
    collector_name: <my_collector_name>

# Define your receivers...
receivers:
  hostmetrics:
    collection_interval: 30s
    scrapers:
      load:

# Define your processors...
processors:

exporters:
  sumologic:
    auth:
      authenticator: sumologic

service:
  extensions: [sumologic]
  pipelines:
    metrics:
      receivers: [hostmetrics]
      processors: []
      exporters: [sumologic]
```

For a list of all configuration options for sumologicextension please refer to
[the documentation][sumologicextension_configuration].

[sumologic_webpage]: https://www.sumologic.com/
[sumologicexporter]: ./pkg/exporter/sumologicexporter/
[sumologicextension]: ./pkg/extension/sumologicextension/
[sumologicextension_configuration]: ./pkg/extension/sumologicextension#configuration

## How to build

```
$ cd otelcolbuilder && make build
opentelemetry-collector-builder \
                --config .otelcol-builder.yaml \
                --output-path ./cmd \
                --name otelcol-sumo
2021-05-24T16:29:03.494+0200    INFO    cmd/root.go:99  OpenTelemetry Collector distribution builder    {"version": "dev", "date": "unknown"}
2021-05-24T16:29:03.498+0200    INFO    builder/main.go:90      Sources created {"path": "./cmd"}
2021-05-24T16:29:03.612+0200    INFO    builder/main.go:126     Getting go modules
2021-05-24T16:29:03.957+0200    INFO    builder/main.go:107     Compiling
2021-05-24T16:29:09.770+0200    INFO    builder/main.go:113     Compiled        {"binary": "./cmd/otelcol-sumo"}
```

In order to build for a different platform one can use `otelcol-sumo-${platform}_${arch}`
make targets e.g.:

```
$ cd otelcolbuilder && make otelcol-sumo-linux_arm64
GOOS=linux   GOARCH=arm64 /Library/Developer/CommandLineTools/usr/bin/make build BINARY_NAME=otelcol-sumo-linux_arm64
opentelemetry-collector-builder \
                --config .otelcol-builder.yaml \
                --output-path ./cmd \
                --name otelcol-sumo-linux_arm64
2021-05-24T16:32:11.963+0200    INFO    cmd/root.go:99  OpenTelemetry Collector distribution builder    {"version": "dev", "date": "unknown"}
2021-05-24T16:32:11.965+0200    INFO    builder/main.go:90      Sources created {"path": "./cmd"}
2021-05-24T16:32:12.066+0200    INFO    builder/main.go:126     Getting go modules
2021-05-24T16:32:12.376+0200    INFO    builder/main.go:107     Compiling
2021-05-24T16:32:37.326+0200    INFO    builder/main.go:113     Compiled        {"binary": "./cmd/otelcol-sumo-linux_arm64"}
```
