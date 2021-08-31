module github.com/open-telemetry/opentelemetry-collector-contrib/processor/sumologicsyslogprocessor

go 1.14

require (
	github.com/stretchr/testify v1.7.0
	go.opentelemetry.io/collector v0.33.0
	go.opentelemetry.io/collector/model v0.33.0
	go.uber.org/zap v1.19.0
)

replace go.opentelemetry.io/collector => github.com/SumoLogic/opentelemetry-collector v0.33.0-sumo-1
