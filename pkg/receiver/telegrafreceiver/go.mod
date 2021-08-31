module github.com/open-telemetry/opentelemetry-collector-contrib/receiver/telegrafreceiver

go 1.14

require (
	cloud.google.com/go/monitoring v0.1.0 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.1.1 // indirect
	github.com/influxdata/telegraf v1.19.0
	github.com/stretchr/testify v1.7.0
	go.opentelemetry.io/collector v0.33.0
	go.opentelemetry.io/collector/model v0.33.0
	go.uber.org/zap v1.19.0
)

replace github.com/influxdata/telegraf => github.com/sumologic/telegraf v1.19.0-sumo-3

replace go.opentelemetry.io/collector => github.com/SumoLogic/opentelemetry-collector v0.33.0-sumo-1
