module github.com/open-telemetry/opentelemetry-collector-contrib/processor/metricfrequencyprocessor

go 1.16

require (
	github.com/patrickmn/go-cache v2.1.0+incompatible // indirect
	go.opentelemetry.io/collector v0.28.0
)

replace go.opentelemetry.io/collector => go.opentelemetry.io/collector v0.27.1-0.20210520180039-2e84285efc66
