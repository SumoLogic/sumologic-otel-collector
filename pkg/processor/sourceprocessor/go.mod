module github.com/open-telemetry/opentelemetry-collector-contrib/processor/sourceprocessor

go 1.14

require (
	github.com/stretchr/testify v1.7.0
	go.opencensus.io v0.23.0
	go.opentelemetry.io/collector v0.31.0
	go.opentelemetry.io/collector/model v0.31.0
)

replace go.opentelemetry.io/collector => github.com/SumoLogic/opentelemetry-collector v0.31.0-sumo-1
