module github.com/open-telemetry/opentelemetry-collector-contrib/processor/metricfrequencyprocessor

go 1.16

require (
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/stretchr/testify v1.7.0
	go.opentelemetry.io/collector v0.31.0
	go.opentelemetry.io/collector/model v0.31.0
)

replace go.opentelemetry.io/collector => github.com/SumoLogic/opentelemetry-collector v0.31.0-sumo-1
