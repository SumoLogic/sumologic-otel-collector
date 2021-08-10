module github.com/open-telemetry/opentelemetry-collector-contrib/processor/cascadingfilterprocessor

go 1.14

require (
	github.com/google/uuid v1.3.0
	github.com/stretchr/testify v1.7.0
	go.opencensus.io v0.23.0
	go.opentelemetry.io/collector v0.31.0
	go.opentelemetry.io/collector/model v0.31.0
	go.uber.org/zap v1.18.1
)

replace go.opentelemetry.io/collector => github.com/SumoLogic/opentelemetry-collector v0.31.0-sumo-1
