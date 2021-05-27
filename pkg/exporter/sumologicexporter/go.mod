module github.com/open-telemetry/opentelemetry-collector-contrib/exporter/sumologicexporter

go 1.15

require (
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/sumologicextension v0.0.0-00010101000000-000000000000
	github.com/stretchr/testify v1.7.0
	go.opentelemetry.io/collector v0.26.1-0.20210513162346-453d1d0dd603
)

replace go.opentelemetry.io/collector => go.opentelemetry.io/collector v0.27.1-0.20210520180039-2e84285efc66

replace github.com/open-telemetry/opentelemetry-collector-contrib/extension/sumologicextension => ./../../extension/sumologicextension
