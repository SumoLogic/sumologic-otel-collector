package luaprocessor

import "go.opentelemetry.io/collector/config"

type Config struct {
	*config.ProcessorSettings `mapstructure:"-"`
	Script                    string `mapstructure:"script"`
}
