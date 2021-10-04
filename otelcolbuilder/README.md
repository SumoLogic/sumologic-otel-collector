# Sumo Logic Distribution of OpenTelemetry with `opentelemetry-collector-builder`

## Tests

In order to run tests run the following command:

```
make test
```

This whill:

- ensure that you have [`opentelemetry-collector-builder`][otcbuilder] installed
- generate boilerplate according to the [.otelcol-builder.yaml][otconfig] config file
- run `go test` checking (golden set of) test configuration files against
  the produced binary

[otcbuilder]: https://github.com/open-telemetry/opentelemetry-collector-builder
[otconfig]: ./.otelcol-builder.yaml

Exemplar output:

```
Installing github.com/open-telemetry/opentelemetry-collector-builder@0.36.0...
go install github.com/open-telemetry/opentelemetry-collector-builder@v0.36.0
CGO_ENABLED=1 opentelemetry-collector-builder \
                --go go \
                --version "v0.0.30-beta.0-9-g6f287a4371" \
                --config .otelcol-builder.yaml \
                --output-path ./cmd \
                --skip-compilation=true \
                --name otelcol-sumo
2021-10-04T16:28:14.733+0200    INFO    cmd/root.go:99  OpenTelemetry Collector distribution builder    {"version": "0.36.0", "date": "2021-10-01T17:43:48Z"}
2021-10-04T16:28:14.734+0200    INFO    cmd/root.go:115 Using config file       {"path": ".otelcol-builder.yaml"}
2021-10-04T16:28:14.853+0200    INFO    builder/config.go:102   Using go        {"Go executable": "go"}
2021-10-04T16:28:14.856+0200    INFO    builder/main.go:87      Sources created {"path": "./cmd"}
2021-10-04T16:28:15.667+0200    INFO    builder/main.go:119     Getting go modules
2021-10-04T16:28:15.864+0200    INFO    builder/main.go:94      Generating source codes only, the distribution will not be compiled.
go test -count 1 -parallel 1 ./...
ok      github.com/SumoLogic/opentelemetry-collector-builder    9.998s

```