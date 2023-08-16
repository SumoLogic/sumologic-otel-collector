package output

import (
	"fmt"
	"io"
	"time"

	"go.opentelemetry.io/collector/confmap"
	"go.opentelemetry.io/collector/consumer"
	"go.uber.org/zap"
)

var factories map[string]builderConfigFactory = map[string]builderConfigFactory{
	"event":       eventConfigFactory{},
	"log_entries": logEntriesConfigFactory{},
}

// Config for the output stage
// Dynamic configuration uses key 'type' to establish the configuration type
type Config struct {
	Format string `mapstructure:"format"`
	Builder
}

// NewDefaultConfig builds the default output configuration
func NewDefaultConfig() Config {
	return Config{
		Format:  "event",
		Builder: eventConfigFactory{}.CreateDefaultConfig(),
	}
}

// Unmarshal dynamic Builder underlying Config
func (c *Config) Unmarshal(component *confmap.Conf) error {
	if component == nil {
		return nil
	}

	// Load non-dynamic parts like normal
	if err := component.Unmarshal(c); err != nil {
		return err
	}

	if !component.IsSet("format") {
		return fmt.Errorf("missing required field 'format'")
	}

	formatInterface := component.Get("format")
	format, ok := formatInterface.(string)
	if !ok {
		return fmt.Errorf("invalid type %T for field 'format'", formatInterface)
	}

	factory, ok := getOutputFormatFactory(format)
	if !ok {
		return fmt.Errorf("invalid value %s for field 'format'", format)
	}

	formatComponent, err := component.Sub(format)
	if err != nil {
		return err
	}

	builder := factory.CreateDefaultConfig()
	if err := formatComponent.Unmarshal(builder, confmap.WithErrorUnused()); err != nil {
		return err
	}
	c.Builder = builder
	return nil
}

// Builder
type Builder interface {
	Build(*zap.Logger, consumer.Logs) Consumer
}

// Consumer consumes command output and emits telemetry data
type Consumer interface {
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

func (c *EventConfig) Build(logger *zap.Logger, next consumer.Logs) Consumer {
	return nopEmitter{}
}

// LogEntriesConfig handles output as if it is a stream of distinct log events
type LogEntriesConfig struct {
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

func (c *LogEntriesConfig) Build(logger *zap.Logger, next consumer.Logs) Consumer {
	return nopEmitter{}
}

func getOutputFormatFactory(format string) (builderConfigFactory, bool) {
	factory, ok := factories[format]
	return factory, ok
}

type builderConfigFactory interface {
	CreateDefaultConfig() Builder
}

type eventConfigFactory struct{}

func (eventConfigFactory) CreateDefaultConfig() Builder {
	return &EventConfig{
		IncludeCommandName:   true,
		IncludeCommandStatus: true,
		IncludeDuration:      true,
	}
}

type logEntriesConfigFactory struct{}

func (logEntriesConfigFactory) CreateDefaultConfig() Builder {
	return &LogEntriesConfig{
		IncludeCommandName: true,
		IncludeStreamName:  true,
	}
}
