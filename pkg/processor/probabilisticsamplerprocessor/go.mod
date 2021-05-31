module github.com/open-telemetry/opentelemetry-collector/processor/probabilisticsamplerprocessor

go 1.14

require (
	github.com/stretchr/testify v1.7.0
	go.opentelemetry.io/collector v0.27.0
	go.uber.org/zap v1.16.0
)

replace go.opentelemetry.io/collector => go.opentelemetry.io/collector v0.27.1-0.20210520180039-2e84285efc66
