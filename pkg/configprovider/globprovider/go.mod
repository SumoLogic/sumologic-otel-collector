module github.com/SumoLogic/sumologic-otel-collector/pkg/configprovider/globprovider

go 1.22.0

toolchain go1.23.0

require (
	github.com/SumoLogic/sumologic-otel-collector/pkg/configprovider/providerutil v0.0.0-00010101000000-000000000000
	github.com/stretchr/testify v1.9.0
	go.opentelemetry.io/collector/confmap v0.107.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/go-viper/mapstructure/v2 v2.3.0 // indirect
	github.com/hashicorp/go-version v1.7.0 // indirect
	github.com/knadh/koanf/maps v0.1.1 // indirect
	github.com/knadh/koanf/providers/confmap v0.1.0 // indirect
	github.com/knadh/koanf/v2 v2.1.1 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	go.opentelemetry.io/collector/featuregate v1.13.0 // indirect
	go.opentelemetry.io/collector/internal/globalgates v0.107.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
)

replace github.com/SumoLogic/sumologic-otel-collector/pkg/configprovider/providerutil => ../providerutil
