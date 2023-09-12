package logentries

import (
	"github.com/SumoLogic/sumologic-otel-collector/pkg/receiver/jobreceiver/output/consumer"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator/helper"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/split"
	"go.uber.org/zap"
)

// LogEntriesConfig handles output as if it is a stream of distinct log events
type LogEntriesConfig struct {
	// IncludeCommandName indicates to include the attribute `command.name`
	IncludeCommandName bool `mapstructure:"include_command_name,omitempty"`
	// IncludeStreamName indicates to include the attribute `command.stream.name`
	// indicating the stream that log entry was consumed from. stdout | stdin.
	IncludeStreamName bool `mapstructure:"include_stream_name,omitempty"`
	// MaxLogSize restricts the length of log output to a specified size.
	// Excess output will overflow to subsequent log entries.
	MaxLogSize helper.ByteSize `mapstructure:"max_log_size,omitempty"`
	// Encoding to expect from output
	Encoding string `mapstructure:"encoding"`
	// Multiline configures alternate log line deliniation
	Multiline split.Config `mapstructure:"multiline"`
}

func (c *LogEntriesConfig) Build(logger *zap.SugaredLogger, op consumer.WriterOp) (consumer.Interface, error) {
	return &consumer.DemoConsumer{WriterOp: op, Logger: logger}, nil
}

type LogEntriesConfigFactory struct{}

func (LogEntriesConfigFactory) CreateDefaultConfig() consumer.Builder {
	return &LogEntriesConfig{
		IncludeCommandName: true,
		IncludeStreamName:  true,
	}
}
