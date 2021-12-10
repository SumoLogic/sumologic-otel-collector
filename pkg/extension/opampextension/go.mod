module github.com/SumoLogic/sumologic-otel-collector/pkg/extension/opampextension

go 1.17

require (
	github.com/open-telemetry/opamp-go v0.0.0-20211206225820-1715144fb40f
	go.opentelemetry.io/collector v0.40.0
	go.uber.org/zap v1.19.1
)

require (
	github.com/cenkalti/backoff/v4 v4.1.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/fsnotify/fsnotify v1.4.9 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/knadh/koanf v1.3.2 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/mapstructure v1.4.2 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/spf13/cast v1.4.1 // indirect
	github.com/stretchr/testify v1.7.0 // indirect
	go.opentelemetry.io/collector/model v0.40.0 // indirect
	go.opentelemetry.io/otel v1.2.0 // indirect
	go.opentelemetry.io/otel/metric v0.25.0 // indirect
	go.opentelemetry.io/otel/trace v1.2.0 // indirect
	go.uber.org/atomic v1.9.0 // indirect
	go.uber.org/multierr v1.7.0 // indirect
	golang.org/x/sys v0.0.0-20211013075003-97ac67df715c // indirect
	google.golang.org/protobuf v1.27.1 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
)

replace github.com/open-telemetry/opamp-go => github.com/pmm-sumo/opamp-go v0.0.2

replace github.com/SumoLogic/sumologic-otel-collector/pkg/extension/opampextension => ./
