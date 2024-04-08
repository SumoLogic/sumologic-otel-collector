package jobreceiver

import (
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/adapter"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/receiver"

	"github.com/SumoLogic/sumologic-otel-collector/pkg/receiver/jobreceiver/builder"
	"github.com/SumoLogic/sumologic-otel-collector/pkg/receiver/jobreceiver/output"
)

var Type = component.MustNewType("monitoringjob")

// NewFactory uses the stanza adapter factory to facilitate the initialization
// and running of the stanza output pipeline.
func NewFactory() receiver.Factory {
	return adapter.NewFactory(stanzaAdapter{}, component.StabilityLevelDevelopment)
}

// stazaAdapter shim
type stanzaAdapter struct{}

func (stanzaAdapter) InputConfig(cfg component.Config) operator.Config {
	jobCfg := cfg.(*Config)
	return operator.NewConfig(builder.NewOperatorBuilder(
		jobCfg.Output,
		jobCfg,
	))
}

func (stanzaAdapter) CreateDefaultConfig() component.Config {
	return &Config{
		Exec:     newDefaultExecutionConfig(),
		Schedule: newDefaultScheduleConfig(),
		Output:   output.NewDefaultConfig(),
	}
}

func (stanzaAdapter) Type() component.Type {
	return Type
}

func (stanzaAdapter) BaseConfig(cfg component.Config) adapter.BaseConfig {
	return cfg.(*Config).Output.BaseConfig
}
