// Copyright OpenTelemetry Authors
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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/consumer/consumererror"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/receiver/receivertest"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func createTestConfig() *Config {
	config := createDefaultConfig().(*Config)
	config.AgentConfig = `
[agent]
	interval = "2s"
	flush_interval = "3s"
[[inputs.mem]]
	`
	return config
}

type countingErrorConsumer struct {
	err       error
	CallCount int
}

func (er *countingErrorConsumer) ConsumeMetrics(context.Context, pmetric.Metrics) error {
	er.CallCount++
	return er.err
}

func (er *countingErrorConsumer) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{}
}

// newCountingErrorConsumer returns a Consumer that drops all received data and returns the specified error to Consume* callers
// It also counts the number of time Consume* was called
func newCountingErrorConsumer(err error) *countingErrorConsumer {
	return &countingErrorConsumer{err: err, CallCount: 0}
}

func TestStartShutdown(t *testing.T) {
	f := NewFactory()
	ctx := context.Background()
	cfg := createTestConfig()
	receiver, err := createMetricsReceiver(ctx, receivertest.NewNopSettings(f.Type()), cfg, consumertest.NewNop())
	require.NoError(t, err)
	require.NoError(t, receiver.Start(ctx, componenttest.NewNopHost()))
	require.NoError(t, receiver.Shutdown(ctx))
}

func TestShutdownBeforeStart(t *testing.T) {
	f := NewFactory()
	ctx := context.Background()
	cfg := createTestConfig()
	receiver, err := createMetricsReceiver(ctx, receivertest.NewNopSettings(f.Type()), cfg, consumertest.NewNop())
	require.NoError(t, err)
	require.NoError(t, receiver.Shutdown(ctx))
}

func TestConsumeRetryOnRecoverableError(t *testing.T) {
	ctx := context.Background()
	maxRetries := 1
	core, observedLogs := observer.New(zapcore.InfoLevel)
	logger := zap.New(core)
	consumerError := errors.New("recoverable error")
	consumer := newCountingErrorConsumer(consumerError)
	receiver := &telegrafreceiver{
		consumer:          consumer,
		logger:            logger,
		metricConverter:   newConverter(true, logger),
		consumeRetryDelay: time.Nanosecond,
		consumeMaxRetries: uint64(maxRetries),
	}

	metrics := pmetric.Metrics{}
	err := receiver.consumeWithRetry(ctx, metrics)
	assert.NotEqual(t, nil, err)
	assert.Equal(t, maxRetries+1, consumer.CallCount)
	assert.Equal(t, maxRetries, observedLogs.Len())
	assert.Equal(t, maxRetries, observedLogs.FilterMessage("ConsumeMetrics() recoverable error, will retry").Len())
}

func TestConsumeNoRetryOnPermanentError(t *testing.T) {
	ctx := context.Background()
	core, observedLogs := observer.New(zapcore.InfoLevel)
	logger := zap.New(core)
	consumerError := consumererror.NewPermanent(errors.New("recoverable error"))
	consumer := newCountingErrorConsumer(consumerError)
	receiver := &telegrafreceiver{
		consumer:          consumer,
		logger:            logger,
		metricConverter:   newConverter(true, logger),
		consumeRetryDelay: time.Nanosecond,
		consumeMaxRetries: 10,
	}

	metrics := pmetric.Metrics{}
	err := receiver.consumeWithRetry(ctx, metrics)
	assert.Error(t, err)
	assert.Equal(t, 1, consumer.CallCount)
	assert.Equal(t, 0, observedLogs.Len())
}
