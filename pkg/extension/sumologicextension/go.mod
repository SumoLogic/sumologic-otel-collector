module github.com/open-telemetry/opentelemetry-collector-contrib/extension/sumologicextension

go 1.14

require (
	github.com/cenkalti/backoff v2.2.1+incompatible
	github.com/cenkalti/backoff/v4 v4.1.1
	github.com/google/uuid v1.3.0
	github.com/stretchr/testify v1.7.0
	go.opentelemetry.io/collector v0.35.0
	go.uber.org/zap v1.19.0
)

replace github.com/open-telemetry/opentelemetry-collector-contrib/exporter/sumologicexporter => ../../exporter/sumologicexporter

replace github.com/open-telemetry/opentelemetry-collector-contrib/extension/sumologicextension => ./
