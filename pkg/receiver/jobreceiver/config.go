package jobreceiver

import (
	"time"

	"github.com/SumoLogic/sumologic-otel-collector/pkg/receiver/jobreceiver/asset"
	"github.com/SumoLogic/sumologic-otel-collector/pkg/receiver/jobreceiver/output"
)

const defaultTimeout = time.Second * 10

// Config for monitoringjob receiver
type Config struct {
	Exec     ExecutionConfig `mapstructure:"exec"`
	Schedule ScheduleConfig  `mapstructure:"schedule"`
	Output   output.Config   `mapstructure:"output"`
}

// ExecutionConfig defines the configuration for execution of a monitoringjob
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

// Validate checks the configuration is valid
func (cfg *Config) Validate() error {
	return nil
}

func newDefaultExecutionConfig() ExecutionConfig {
	return ExecutionConfig{
		Timeout: defaultTimeout,
	}
}

func newDefaultScheduleConfig() ScheduleConfig {
	return ScheduleConfig{}
}
