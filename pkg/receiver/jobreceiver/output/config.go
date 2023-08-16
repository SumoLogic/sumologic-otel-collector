package output

import (
	"fmt"
	"io"
	"time"

	"go.opentelemetry.io/collector/confmap"
	"go.opentelemetry.io/collector/consumer"
	"go.uber.org/zap"
)

// Config for the output stage
// Dynamic configuration uses key 'type' to establish the configuration type
type Config struct {
	Builder
}

// NewDefaultConfig builds the default output configuration
func NewDefaultConfig() Config {
	return Config{
		Builder: newDefaultEventConfig(),
	}
}

func newDefaultEventConfig() *EventConfig {
	return &EventConfig{
		Type:                 "event",
		IncludeCommandName:   true,
		IncludeCommandStatus: true,
		IncludeDuration:      true,
	}
}

func newDefaultLogEntriesConfig() *LogEntriesConfig {
	return &LogEntriesConfig{
		Type:               "log_entries",
		IncludeCommandName: true,
		IncludeStreamName:  true,
	}
}

// Unmarshal dynamic Builder underlying Config
func (c *Config) Unmarshal(component *confmap.Conf) error {
	if !component.IsSet("type") {
		return fmt.Errorf("missing required field 'type'")
	}

	typeInterface := component.Get("type")
	typeName, ok := typeInterface.(string)
	if !ok {
		return fmt.Errorf("invalid type %T for field 'type'", typeInterface)
	}

	switch typeName {
	case "event":
		builder := newDefaultEventConfig()
		if err := component.Unmarshal(builder, confmap.WithErrorUnused()); err != nil {
			return err
		}
		c.Builder = builder
	case "log_entries":
		builder := newDefaultLogEntriesConfig()
		if err := component.Unmarshal(builder, confmap.WithErrorUnused()); err != nil {
			return err
		}
		c.Builder = builder
	default:
		return fmt.Errorf("unsupported value %s for field 'type'", typeName)
	}
	return nil
}

// Builder
type Builder interface {
	Build(*zap.Logger, consumer.Logs) Emitter
}

// Emitter consumes command output and emits telemetry data
type Emitter interface {
	Consume(stdin, stderr io.Reader) CloseFunc
}

// CloseFunc
type CloseFunc func(ExecutionSummary)

type nopEmitter struct{}

func (nopEmitter) Consume(_, _ io.Reader) CloseFunc {
	return func(_ ExecutionSummary) {}
}

// ExecutionSummary describes a command execution
type ExecutionSummary struct {
	Command     string
	ExitCode    int
	RunDuration time.Duration
}

// EventConfig handles output as if it is a monitoring job.
// Should emit a single event per command execution summarizing the execution.
type EventConfig struct {
	Type string `mapstructure:"type"`
	// IncludeCommandName indicates to include the attribute `command.name`
	IncludeCommandName bool `mapstructure:"include_command_name,omitempty"`
	// IncludeCommandStatus indicates to include the attribute `command.status`
	// indicating the command's exit code
	IncludeCommandStatus bool `mapstructure:"include_command_status,omitempty"`
	// IncludeDuration indicates to include the attribute `command.duration`
	IncludeDuration bool `mapstructure:"include_command_duration,omitempty"`
	// MaxBodySize restricts the length of command output to a specified size.
	// Excess output will be truncated.
	MaxBodySize ByteSize `mapstructure:"max_body_size,omitempty"`
}

func (c *EventConfig) Build(logger *zap.Logger, next consumer.Logs) Emitter {
	return nopEmitter{}
}

// LogEntriesConfig handles output as if it is a stream of distinct log events
type LogEntriesConfig struct {
	Type string `mapstructure:"type"`
	// IncludeCommandName indicates to include the attribute `command.name`
	IncludeCommandName bool `mapstructure:"include_command_name,omitempty"`
	// IncludeStreamName indicates to include the attribute `command.stream.name`
	// indicating the stream that log entry was consumed from. stdout | stdin.
	IncludeStreamName bool `mapstructure:"include_stream_name,omitempty"`
	// MaxLogSize restricts the length of log output to a specified size.
	// Excess output will overflow to subsequent log entries.
	MaxLogSize ByteSize `mapstructure:"max_log_size,omitempty"`
	// Encoding to expect from output
	Encoding string `mapstructure:"encoding"`
	// Multiline configures alternate log line deliniation
	Multiline MultilineConfig `mapstructure:"multiline"`
}

func (c *LogEntriesConfig) Build(logger *zap.Logger, next consumer.Logs) Emitter {
	return nopEmitter{}
}

// MultilineConfig configures how log entries should be delimited
type MultilineConfig struct {
	LineStartPattern string `mapstructure:"line_start_pattern"`
	LineEndPattern   string `mapstructure:"line_end_pattern"`
}
