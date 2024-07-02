// Copyright 2020 OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package sumologicexporter

import (
	"context"
	"fmt"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configretry"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/exporter/exporterhelper"
)

const (
	// The value of "type" key in configuration.
	typeStr        = "sumologic"
	stabilityLevel = component.StabilityLevelDeprecated
)

var Type = component.MustNewType(typeStr)

// NewFactory returns a new factory for the sumologic exporter.
func NewFactory() exporter.Factory {
	return exporter.NewFactory(
		Type,
		createDefaultConfig,
		exporter.WithLogs(createLogsExporter, stabilityLevel),
		exporter.WithMetrics(createMetricsExporter, stabilityLevel),
		exporter.WithTraces(createTracesExporter, stabilityLevel),
	)
}

func createDefaultConfig() component.Config {
	qs := exporterhelper.NewDefaultQueueSettings()
	qs.Enabled = false

	return &Config{
		MaxRequestBodySize: DefaultMaxRequestBodySize,
		LogFormat:          DefaultLogFormat,
		MetricFormat:       DefaultMetricFormat,
		Client:             DefaultClient,
		TraceFormat:        OTLPTraceFormat,

		ClientConfig:         CreateDefaultClientConfig(),
		BackOffConfig:        configretry.NewDefaultBackOffConfig(),
		QueueSettings:        qs,
		StickySessionEnabled: DefaultStickySessionEnabled,
	}
}

func createLogsExporter(
	ctx context.Context,
	params exporter.Settings,
	cfg component.Config,
) (exporter.Logs, error) {
	exp, err := newLogsExporter(ctx, params, cfg.(*Config))
	if err != nil {
		return nil, fmt.Errorf("failed to create the logs exporter: %w", err)
	}

	return exp, nil
}

func createMetricsExporter(
	ctx context.Context,
	params exporter.Settings,
	cfg component.Config,
) (exporter.Metrics, error) {
	exp, err := newMetricsExporter(ctx, params, cfg.(*Config))
	if err != nil {
		return nil, fmt.Errorf("failed to create the metrics exporter: %w", err)
	}

	return exp, nil
}

func createTracesExporter(
	ctx context.Context,
	params exporter.Settings,
	cfg component.Config,
) (exporter.Traces, error) {
	exp, err := newTracesExporter(ctx, params, cfg.(*Config))
	if err != nil {
		return nil, fmt.Errorf("failed to create the traces exporter: %w", err)
	}

	return exp, nil
}
