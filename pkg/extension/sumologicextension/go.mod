module github.com/open-telemetry/opentelemetry-collector-contrib/extension/sumologicextension

go 1.14

require (
	github.com/armon/go-metrics v0.3.3 // indirect
	github.com/gogo/googleapis v1.3.0 // indirect
	github.com/hashicorp/go-immutable-radix v1.2.0 // indirect
	github.com/hashicorp/go-msgpack v0.5.5 // indirect
	github.com/mattn/go-colorable v0.1.7 // indirect
	github.com/onsi/ginkgo v1.14.1 // indirect
	github.com/onsi/gomega v1.10.2 // indirect
	github.com/pelletier/go-toml v1.8.0 // indirect
	github.com/stretchr/testify v1.7.0
	go.opentelemetry.io/collector v0.26.1-0.20210513162346-453d1d0dd603
	go.uber.org/zap v1.16.0
	gopkg.in/ini.v1 v1.57.0 // indirect
)

replace go.opentelemetry.io/collector => go.opentelemetry.io/collector v0.27.1-0.20210520180039-2e84285efc66

replace github.com/open-telemetry/opentelemetry-collector-contrib/exporter/sumologicexporter => ../../exporter/sumologicexporter

replace github.com/open-telemetry/opentelemetry-collector-contrib/extension/sumologicextension => ./
