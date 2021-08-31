module github.com/open-telemetry/opentelemetry-collector-contrib/extension/sumologicextension

go 1.14

require (
	github.com/cenkalti/backoff v2.2.1+incompatible
	github.com/cenkalti/backoff/v4 v4.1.1
	github.com/google/uuid v1.3.0
	github.com/hashicorp/go-msgpack v0.5.5 // indirect
	github.com/mattn/go-colorable v0.1.7 // indirect
	github.com/onsi/ginkgo v1.14.1 // indirect
	github.com/onsi/gomega v1.10.2 // indirect
	github.com/stretchr/testify v1.7.0
	go.opentelemetry.io/collector v0.33.0
	go.uber.org/zap v1.19.0
)

replace github.com/open-telemetry/opentelemetry-collector-contrib/exporter/sumologicexporter => ../../exporter/sumologicexporter

replace github.com/open-telemetry/opentelemetry-collector-contrib/extension/sumologicextension => ./

replace go.opentelemetry.io/collector => github.com/SumoLogic/opentelemetry-collector v0.33.0-sumo-1
