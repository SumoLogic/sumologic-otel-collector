package output

import (
	"fmt"

	"github.com/SumoLogic/sumologic-otel-collector/pkg/receiver/jobreceiver/output/consumer"
	"github.com/SumoLogic/sumologic-otel-collector/pkg/receiver/jobreceiver/output/event"
	"github.com/SumoLogic/sumologic-otel-collector/pkg/receiver/jobreceiver/output/logentries"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/adapter"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator/helper"
	"go.opentelemetry.io/collector/confmap"
)

const (
	inputTypeEvent      = "event"
	inputTypeLogEntries = "log_entries"
)

var factories map[string]builderConfigFactory = map[string]builderConfigFactory{
	inputTypeEvent:      event.EventConfigFactory{},
	inputTypeLogEntries: logentries.LogEntriesConfigFactory{},
}

func getOutputFormatFactory(format string) (builderConfigFactory, bool) {
	factory, ok := factories[format]
	return factory, ok
}

type builderConfigFactory interface {
	CreateDefaultConfig() consumer.Builder
}

// Config for the output stage
// Dynamic configuration uses key 'type' to establish the configuration type
type Config struct {
	BaseConfig  adapter.BaseConfig `mapstructure:",squash"`
	InputConfig helper.InputConfig `mapstructure:",squash"`
	Builder     consumer.Builder
}

// NewDefaultConfig builds the default output configuration
func NewDefaultConfig() Config {
	return Config{
		InputConfig: helper.NewInputConfig("", inputTypeEvent),
		Builder:     event.EventConfigFactory{}.CreateDefaultConfig(),
	}
}

// Unmarshal dynamic Builder underlying Config
func (c *Config) Unmarshal(component *confmap.Conf) error {
	if component == nil {
		return nil
	}
	// Load non-dynamic parts like normal
	if err := component.Unmarshal(c, confmap.WithIgnoreUnused()); err != nil {
		return err
	}

	if !component.IsSet("type") {
		return fmt.Errorf("missing required field 'type'")
	}

	formatInterface := component.Get("type")
	format, ok := formatInterface.(string)
	if !ok {
		return fmt.Errorf("invalid type %T for field 'type'", formatInterface)
	}

	factory, ok := getOutputFormatFactory(format)
	if !ok {
		return fmt.Errorf("invalid value %s for field 'type'", format)
	}

	subComponent, err := component.Sub(format)
	if err != nil {
		return err
	}

	c.Builder = factory.CreateDefaultConfig()
	if err := subComponent.Unmarshal(c.Builder); err != nil {
		return err
	}
	return nil
}
