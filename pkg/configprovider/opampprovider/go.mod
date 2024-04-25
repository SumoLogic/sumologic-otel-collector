module github.com/SumoLogic/sumologic-otel-collector/pkg/configprovider/opampprovider

go 1.21

toolchain go1.21.4

require (
	github.com/SumoLogic/sumologic-otel-collector/pkg/configprovider/globprovider v0.0.0-00010101000000-000000000000
	github.com/google/go-cmp v0.5.9
	go.opentelemetry.io/collector/confmap v0.99.0
	go.opentelemetry.io/collector/confmap/provider/fileprovider v0.99.0
	gopkg.in/yaml.v2 v2.4.0
)

require (
	github.com/go-viper/mapstructure/v2 v2.0.0-alpha.1 // indirect
	github.com/knadh/koanf/maps v0.1.1 // indirect
	github.com/knadh/koanf/providers/confmap v0.1.0 // indirect
	github.com/knadh/koanf/v2 v2.1.1 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/SumoLogic/sumologic-otel-collector/pkg/configprovider/globprovider => ../globprovider
