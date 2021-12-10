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

package servicemapreceiver

import (
	"context"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver/receiverhelper"
)

const (
	typeStr    = "servicemap"
	versionStr = "v0.1"
)

// NewFactory creates a factory for telegraf receiver.
func NewFactory() component.ReceiverFactory {
	return receiverhelper.NewFactory(
		typeStr,
		createDefaultConfig,
		receiverhelper.WithMetrics(createMetricsReceiver),
		receiverhelper.WithTraces(createTracesReceiver),
	)
}

func createDefaultConfig() config.Receiver {
	rs := config.NewReceiverSettings(config.NewComponentID(typeStr))
	return &Config{
		&rs,
	}
}

// createMetricsReceiver creates a metrics receiver based on provided config.
func createMetricsReceiver(
	ctx context.Context,
	params component.ReceiverCreateSettings,
	cfg config.Receiver,
	nextConsumer consumer.Metrics,
) (component.MetricsReceiver, error) {
	return newServiceMapReceiver(params.Logger), nil
}

// createMetricsReceiver creates a metrics receiver based on provided config.
func createTracesReceiver(
	ctx context.Context,
	params component.ReceiverCreateSettings,
	cfg config.Receiver,
	nextConsumer consumer.Traces,
) (component.TracesReceiver, error) {
	rcv := newServiceMapReceiver(params.Logger)
	rcv.tracesConsumer = nextConsumer
	return rcv, nil
}
