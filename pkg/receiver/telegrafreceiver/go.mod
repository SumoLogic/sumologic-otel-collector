module github.com/open-telemetry/opentelemetry-collector-contrib/receiver/telegrafreceiver

go 1.14

require (
	cloud.google.com/go/kms v0.1.0 // indirect
	cloud.google.com/go/monitoring v0.1.0 // indirect
	github.com/influxdata/telegraf v1.19.0
	github.com/stretchr/testify v1.7.0
	go.opentelemetry.io/collector v0.36.0
	go.opentelemetry.io/collector/model v0.36.0
	go.uber.org/zap v1.19.1
)

replace github.com/influxdata/telegraf => github.com/sumologic/telegraf v1.19.3-sumo-0
