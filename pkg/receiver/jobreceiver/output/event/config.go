package event

import (
	"github.com/SumoLogic/sumologic-otel-collector/pkg/receiver/jobreceiver/output/consumer"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator/helper"
	"go.uber.org/zap"
)

const (
	commandNameLabel     = "command.name"
	commandStatusLabel   = "command.status"
	commandDurationLabel = "command.duration"
)

// EventConfig handles output as if it is a monitoring job.
// Should emit a single event per command execution summarizing the execution.
type EventConfig struct {
	// IncludeCommandName indicates to include the attribute `command.name`
	IncludeCommandName bool `mapstructure:"include_command_name,omitempty"`
	// IncludeCommandStatus indicates to include the attribute `command.status`
	// indicating the command's exit code
	IncludeCommandStatus bool `mapstructure:"include_command_status,omitempty"`
	// IncludeDuration indicates to include the attribute `command.duration`
	IncludeDuration bool `mapstructure:"include_command_duration,omitempty"`
	// MaxBodySize restricts the length of command output to a specified size.
	// Excess output will be truncated.
	MaxBodySize helper.ByteSize `mapstructure:"max_body_size,omitempty"`
}

func (c *EventConfig) Build(logger *zap.SugaredLogger, op consumer.WriterOp) (consumer.Interface, error) {
	return &handler{writer: op, logger: logger, config: *c}, nil
}

type EventConfigFactory struct{}

func (EventConfigFactory) CreateDefaultConfig() consumer.Builder {
	return &EventConfig{
		IncludeCommandName:   true,
		IncludeCommandStatus: true,
		IncludeDuration:      true,
	}
}
