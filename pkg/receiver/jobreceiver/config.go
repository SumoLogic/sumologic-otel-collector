package jobreceiver

import (
	"errors"
	"fmt"
	"time"

	"github.com/SumoLogic/sumologic-otel-collector/pkg/receiver/jobreceiver/asset"
)

// Config for monitoringjob receiver
type Config struct {
	Schedule ScheduleConfig  `mapstructure:"schedule"`
	Exec     ExecutionConfig `mapstructure:"exec"`
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
	Interval time.Duration `mapstructure:"interval,omitempty"`
}

// Validate checks the configuration is valid
func (cfg *Config) Validate() error {
	if cfg.Exec.Command == "" {
		return errors.New("command to execute must be non-empty")
	}
	if cfg.Schedule.Interval <= 0 {
		return errors.New("schedule must be set")
	}
	for i, a := range cfg.Exec.RuntimeAssets {
		if err := a.Validate(); err != nil {
			return fmt.Errorf("invalid runtime asset %d: %w", i, err)
		}
	}
	return nil
}
