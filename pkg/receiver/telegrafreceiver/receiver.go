// Copyright 2021, OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package telegrafreceiver

import (
	"context"
	"errors"
	"sync"
	"time"

	backoff "github.com/cenkalti/backoff/v4"
	"github.com/influxdata/telegraf"
	telegrafagent "github.com/influxdata/telegraf/agent"
	telegrafconfig "github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/models"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/consumer/consumererror"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/receiver"
	"go.uber.org/zap"

	// Blank imports to register all the plugins
	_ "github.com/influxdata/telegraf/plugins/inputs/all"
	_ "github.com/influxdata/telegraf/plugins/parsers/all"
)

var (
	ErrAlreadyStarted = errors.New("component already started")
	ErrAlreadyStopped = errors.New("component already stopped")
)

type telegrafreceiver struct {
	sync.Mutex
	startOnce sync.Once
	stopOnce  sync.Once
	wg        sync.WaitGroup
	cancel    context.CancelFunc

	agent             *telegrafagent.Agent
	consumer          consumer.Metrics
	logger            *zap.Logger
	metricConverter   MetricConverter
	consumeRetryDelay time.Duration
	consumeMaxRetries uint64
}

// Ensure this receiver adheres to required interface.
var _ receiver.Metrics = (*telegrafreceiver)(nil)

func addOutputConfig(config *telegrafconfig.Config, ch chan telegraf.Metric) {
	outputConfig := &models.OutputConfig{
		Name:   "internal",
		Source: "channel",
		Alias:  "channel",
		ID:     "channel",
	}
	output := models.NewRunningOutput(newChannelOutput(ch), outputConfig, 1, 1)
	config.Outputs = append(config.Outputs, output)
}

// Start tells the receiver to start.
func (r *telegrafreceiver) Start(ctx context.Context, host component.Host) error {
	r.logger.Info("Starting telegraf receiver")

	r.Lock()
	defer r.Unlock()

	err := ErrAlreadyStarted
	r.startOnce.Do(func() {
		err = nil
		rctx, cancel := context.WithCancel(ctx)
		r.cancel = cancel

		ch := make(chan telegraf.Metric)
		addOutputConfig(r.agent.Config, ch)

		r.wg.Add(1)
		go func() {
			defer r.wg.Done()
			if rErr := r.agent.Run(rctx); rErr != nil {
				r.logger.Error("Problem starting receiver", zap.Error(rErr))
			}
		}()

		r.wg.Add(1)
		go func() {
			var fErr error
			defer r.wg.Done()
			// Telegraf expects its input plugins to always be able to write to this channel while running,
			// and if we stop reading from it while there's still active plugins, we'll get a deadlock.
			// As such, this loop only exits when the channel is closed by Telegraf itself.
			for m := range ch {
				if m == nil {
					r.logger.Info("got nil from channel")
					continue
				}

				var ms pmetric.Metrics
				if ms, fErr = r.metricConverter.Convert(m); fErr != nil {
					r.logger.Error(
						"Error converting telegraf.Metric to pmetric.Metrics",
						zap.Error(fErr),
					)
					continue
				}
				fErr = r.consumeWithRetry(rctx, ms)
				if fErr != nil {
					r.logger.Error("ConsumeMetrics() error",
						zap.String("error", fErr.Error()),
					)
				}
			}
		}()
	})

	return err
}

// Consume metrics and retry on recoverable errors
func (r *telegrafreceiver) consumeWithRetry(ctx context.Context, metrics pmetric.Metrics) error {
	constantBackoff := backoff.WithMaxRetries(backoff.NewConstantBackOff(r.consumeRetryDelay), r.consumeMaxRetries)

	// retry handling according to https://github.com/open-telemetry/opentelemetry-collector/blob/master/component/receiver.go#L45
	err := backoff.RetryNotify(
		func() error {
			// we need to check for context cancellation here
			select {
			case <-ctx.Done():
				return backoff.Permanent(errors.New("closing"))
			default:
			}
			err := r.consumer.ConsumeMetrics(ctx, metrics)
			if consumererror.IsPermanent(err) {
				return backoff.Permanent(err)
			} else {
				return err
			}
		},
		constantBackoff,
		func(err error, delay time.Duration) {
			r.logger.Warn("ConsumeMetrics() recoverable error, will retry",
				zap.Error(err), zap.Duration("delay", delay),
			)
		},
	)

	return err
}

// Shutdown is invoked during service shutdown.
func (r *telegrafreceiver) Shutdown(context.Context) error {
	r.Lock()
	defer r.Unlock()

	err := ErrAlreadyStopped
	r.stopOnce.Do(func() {
		r.logger.Info("Stopping telegraf receiver")
		if r.cancel != nil { // need to check because Shutdown can be called before Start
			r.cancel()
		}
		r.wg.Wait()
		err = nil
	})
	return err
}
