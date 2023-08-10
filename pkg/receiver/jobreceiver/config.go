package jobreceiver

import (
	"time"

	"github.com/SumoLogic/sumologic-otel-collector/pkg/receiver/jobreceiver/asset"
)

// Config for monitoringjob receiver
type Config struct {
	Exec     ExecutionConfig `mapstructure:"exec"`
	Schedule ScheduleConfig  `mapstructure:",squash"`
	Output   OutputConfig    `mapstructure:",squash"`
}

// ExecutionConfig defines the configuration for execution of a monitorinjob
// process
type ExecutionConfig struct {
	// Command is the name of the binary to be executed
	Command string `mapstructure:"command"`
	// Arguments to pass to the command
	Arguments []string `mapstructure:"arguments"`
	// RuntimeAssets made available to the execution context
	RuntimeAssets []asset.Spec `mapstructure:"runtime_assets"`
	// Timeout is the time to wait for the process to exit before attempting
	// to make it stop
	Timeout time.Duration `mapstructure:"timeout,omitempty"`
}

// ScheduleConfig defines configuration for the scheduling of the monitoringjob
// receiver
type ScheduleConfig struct {
	// Interval to schedule monitoring job at
	Interval time.Duration `mapstructure:"interval,omitempty"`
}

type OutputConfig struct {
	// Attributes to include with log events
	Attributes map[string]string `mapstructure:"attributes,omitempty"`
	// Resource attributes to include with log events
	Resource map[string]string `mapstructure:"resource,omitempty"`
	// Encoding expected in output streams
	Encoding string `mapstructure:"encoding,omitempty"`
	// Multiline configuration for augmenting how log events are delimeted
	Multiline MultilineConfig `mapstructure:"multiline,omitempty"`
}

func (cfg *OutputConfig) Validate() error {
	return nil
}

type MultilineConfig struct {
	// LineStartPattern regex pattern for detecting the start of a log line
	LineStartPattern string `mapstructure:"line_start_pattern"`
	// LineEndPattern regex pattern for detecting the end of a log line
	LineEndPattern string `mapstructure:"line_end_pattern"`
}

func (cfg *MultilineConfig) Validate() error {
	return nil
}

// Validate checks the configuration is valid
func (cfg *Config) Validate() error {
	return nil
}
