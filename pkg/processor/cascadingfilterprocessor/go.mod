module github.com/open-telemetry/opentelemetry-collector-contrib/processor/cascadingfilterprocessor

go 1.14

require (
	github.com/google/uuid v1.2.0
	github.com/stretchr/testify v1.7.0
	go.opencensus.io v0.23.0
	go.opentelemetry.io/collector v0.19.0
	go.uber.org/zap v1.16.0
)

// replace go.opentelemetry.io/collector => github.com/SumoLogic/opentelemetry-collector v0.26.0-sumo-1-rc.0
replace go.opentelemetry.io/collector => github.com/SumoLogic/opentelemetry-collector v0.2.7-0.20210524103057-96a028d589eb
