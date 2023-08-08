package jobreceiver

import (
	"context"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver"
)

const (
	typeStr   = "monitoringjob"
	stability = component.StabilityLevelDevelopment

	defaultInterval = 5 * time.Minute
)

// NewFactory creates a factory for the prometheusexec receiver
func NewFactory() receiver.Factory {
	return receiver.NewFactory(
		typeStr,
		createDefaultConfig,
		receiver.WithLogs(createLogsReceiver, stability))
}

func createDefaultConfig() component.Config {
	return &Config{
		Schedule: ScheduleConfig{
			Interval: defaultInterval,
		},
	}
}

func createLogsReceiver(ctx context.Context,
	params receiver.CreateSettings,
	cfg component.Config,
	next consumer.Logs,
) (receiver.Logs, error) {
	return &monitoringJobReceiver{
		logger: params.Logger,
	}, nil
}
