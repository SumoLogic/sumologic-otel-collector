module github.com/SumoLogic/sumologic-otel-collector/pkg/configprovider/opampprovider

go 1.21.0

toolchain go1.22.6

require (
	github.com/SumoLogic/sumologic-otel-collector/pkg/configprovider/globprovider v0.0.0-00010101000000-000000000000
	github.com/google/go-cmp v0.5.9
	go.opentelemetry.io/collector/confmap v0.106.1
	go.opentelemetry.io/collector/confmap/provider/fileprovider v0.106.1
	gopkg.in/yaml.v2 v2.4.0
)

require (
	github.com/go-viper/mapstructure/v2 v2.0.0 // indirect
	github.com/hashicorp/go-version v1.7.0 // indirect
	github.com/knadh/koanf/maps v0.1.1 // indirect
	github.com/knadh/koanf/providers/confmap v0.1.0 // indirect
	github.com/knadh/koanf/v2 v2.1.1 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	go.opentelemetry.io/collector/featuregate v1.12.0 // indirect
	go.opentelemetry.io/collector/internal/globalgates v0.106.1 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/SumoLogic/sumologic-otel-collector/pkg/configprovider/globprovider => ../globprovider
