package jobreceiver

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/receiver"
	"go.uber.org/zap"
)

type monitoringJobReceiver struct {
	logger *zap.Logger
}

var _ receiver.Logs = (*monitoringJobReceiver)(nil)

// Start monitoringJobReceiver
// TODO(ck)
func (r *monitoringJobReceiver) Start(ctx context.Context, host component.Host) error {
	r.logger.Warn("starting monitoringjob receiver. not yet implemented.")
	return nil
}

// Shutdown monitoringJobReceiver
// TODO(ck)
func (r *monitoringJobReceiver) Shutdown(ctx context.Context) error {
	return nil
}
