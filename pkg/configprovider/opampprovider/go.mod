module github.com/SumoLogic/sumologic-otel-collector/pkg/configprovider/opampprovider

go 1.24

toolchain go1.25.1

require (
	github.com/SumoLogic/sumologic-otel-collector/pkg/configprovider/globprovider v0.130.1-sumo-0
	github.com/SumoLogic/sumologic-otel-collector/pkg/configprovider/providerutil v0.130.1-sumo-0
	github.com/google/go-cmp v0.7.0
	go.opentelemetry.io/collector/confmap v1.40.0
	go.opentelemetry.io/collector/confmap/provider/fileprovider v1.40.0
	gopkg.in/yaml.v2 v2.4.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/go-viper/mapstructure/v2 v2.4.0 // indirect
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/hashicorp/go-version v1.7.0 // indirect
	github.com/knadh/koanf/maps v0.1.2 // indirect
	github.com/knadh/koanf/providers/confmap v1.0.0 // indirect
	github.com/knadh/koanf/v2 v2.2.2 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	go.opentelemetry.io/collector/featuregate v1.40.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
)

replace github.com/SumoLogic/sumologic-otel-collector/pkg/configprovider/globprovider => ../globprovider

replace github.com/SumoLogic/sumologic-otel-collector/pkg/configprovider/providerutil => ../providerutil
