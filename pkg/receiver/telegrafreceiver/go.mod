module github.com/open-telemetry/opentelemetry-collector-contrib/receiver/telegrafreceiver

go 1.14

require (
	github.com/influxdata/telegraf v1.19.0
	github.com/stretchr/testify v1.7.0
	go.opentelemetry.io/collector v0.29.0
	go.uber.org/zap v1.17.0
)

replace github.com/influxdata/telegraf => github.com/sumologic/telegraf v1.19.0-sumo-1

replace go.opentelemetry.io/collector => github.com/SumoLogic/opentelemetry-collector v0.29.0-sumo-1

// Needed due to https://github.com/golang/go/issues/46645 present in go1.17-beta1
// and in go1.17rc1.
//
// TODO: remove when this is fixed in newer go versions.
replace golang.org/x/sys => golang.org/x/sys v0.0.0-20210630005230-0f9fa26af87c
