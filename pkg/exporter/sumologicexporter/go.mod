module github.com/open-telemetry/opentelemetry-collector-contrib/exporter/sumologicexporter

go 1.15

require (
	github.com/klauspost/compress v1.13.1
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/sumologicextension v0.0.0-00010101000000-000000000000
	github.com/stretchr/testify v1.7.0
	go.opentelemetry.io/collector v0.36.0
	go.opentelemetry.io/collector/model v0.36.0
	go.uber.org/zap v1.19.1
)

replace github.com/open-telemetry/opentelemetry-collector-contrib/extension/sumologicextension => ./../../extension/sumologicextension
