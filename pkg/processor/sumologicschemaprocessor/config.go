package sumologicschemaprocessor

import "go.opentelemetry.io/collector/config"

type Config struct {
	config.ProcessorSettings `mapstructure:",squash"`

	AddCloudNamespace bool `mapstructure:"add_cloud_namespace"`
}

const (
	defaultAddCloudNamespace = true
)

// Ensure the Config struct satisfies the config.Processor interface.
var _ config.Processor = (*Config)(nil)

func createDefaultConfig() config.Processor {
	return &Config{
		ProcessorSettings: config.NewProcessorSettings(config.NewComponentID(typeStr)),
		AddCloudNamespace: defaultAddCloudNamespace,
	}
}
