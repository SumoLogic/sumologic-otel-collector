# sumologic-otel-collector

[![Default branch build](https://github.com/SumoLogic/sumologic-otel-collector/actions/workflows/dev_builds.yml/badge.svg)](https://github.com/SumoLogic/sumologic-otel-collector/actions/workflows/dev_builds.yml)

Sumo Logic Distro of [OpenTelemetry Collector][otc_link] built with
[opentelemetry-collector-builder][otc_builder_link].

[otc_link]: https://github.com/open-telemetry/opentelemetry-collector
[otc_builder_link]: https://github.com/open-telemetry/opentelemetry-collector-builder

---

- [Configuration](#configuration)
- [How to run](#how-to-run)
  - [Binary](#binary)
  - [Docker container](#docker-container)
- [How to build](#how-to-build)

---

## Configuration

In order to send data to [Sumo Logic][sumologic_webpage] one needs to configure
the [sumologicextension][sumologicextension] with credentials and define it
(the extension) in the same service as [sumologicexporter][sumologicexporter]
is defined so that it's used as auth extension.

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

## How to run

- [Run as binary](./docs/HowToRun.md#binary)
- [Run as Docker container](./docs/HowToRun.md#docker-container)

## How to build

- [Build binary](./CONTRIBUTING.md#how-to-build)

## Contributing

For contributing guidelines please refer to the [CONTRIBUTING.md](./CONTRIBUTING.md)