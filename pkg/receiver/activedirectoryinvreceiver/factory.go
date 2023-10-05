package activedirectoryinvreceiver

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver"
)

const (
	// The value of "type" key in configuration.
	typeStr = "activedirectoryinv"
)

// NewFactory creates a factory for Active Directory Inventory receiver
func NewFactory() receiver.Factory {
	return receiver.NewFactory(
		typeStr,
		CreateDefaultConfig,
		receiver.WithLogs(createLogsReceiver, component.StabilityLevelAlpha),
	)
}

// CreateDefaultConfig creates the default configuration for the receiver
func CreateDefaultConfig() component.Config {
	return &ADConfig{
		CN:           "test user",
		OU:           "test",
		Password:     "test",
		DC:           "exampledomain.com",
		Host:         "hostname.exampledomain.com",
		PollInterval: 60,
	}
}

func createLogsReceiver(
	_ context.Context,
	params receiver.CreateSettings,
	rConf component.Config,
	consumer consumer.Logs,
) (receiver.Logs, error) {
	cfg := rConf.(*ADConfig)
	rcvr := newLogsReceiver(cfg, params.Logger, consumer)
	return rcvr, nil
}
