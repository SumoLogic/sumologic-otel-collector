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
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/exporter/exporterhelper"
)

const (
	// The value of "type" key in configuration.
	typeStr = "sumologic"
)

// NewFactory returns a new factory for the sumologic exporter.
func NewFactory() component.ExporterFactory {
	return component.NewExporterFactory(
		typeStr,
		createDefaultConfig,
		component.WithLogsExporter(createLogsExporter),
		component.WithMetricsExporter(createMetricsExporter),
		component.WithTracesExporter(createTracesExporter),
	)
}

func createDefaultConfig() config.Exporter {
	qs := exporterhelper.NewDefaultQueueSettings()
	qs.Enabled = false

	return &Config{
		ExporterSettings: config.NewExporterSettings(config.NewComponentID(typeStr)),

		TranslateAttributes:      DefaultTranslateAttributes,
		TranslateTelegrafMetrics: DefaultTranslateTelegrafMetrics,
		CompressEncoding:         DefaultCompressEncoding,
		MaxRequestBodySize:       DefaultMaxRequestBodySize,
		LogFormat:                DefaultLogFormat,
		MetricFormat:             DefaultMetricFormat,
		SourceCategory:           DefaultSourceCategory,
		SourceName:               DefaultSourceName,
		SourceHost:               DefaultSourceHost,
		Client:                   DefaultClient,
		ClearLogsTimestamp:       DefaultClearLogsTimestamp,
		JSONLogs: JSONLogs{
			LogKey:       DefaultLogKey,
			AddTimestamp: DefaultAddTimestamp,
			TimestampKey: DefaultTimestampKey,
			FlattenBody:  DefaultFlattenBody,
		},
		TraceFormat: OTLPTraceFormat,

		HTTPClientSettings:   CreateDefaultHTTPClientSettings(),
		RetrySettings:        exporterhelper.NewDefaultRetrySettings(),
		QueueSettings:        qs,
		DropRoutingAttribute: DefaultDropRoutingAttribute,
	}
}

func createLogsExporter(
	_ context.Context,
	params component.ExporterCreateSettings,
	cfg config.Exporter,
) (component.LogsExporter, error) {
	exp, err := newLogsExporter(cfg.(*Config), params)
	if err != nil {
		return nil, fmt.Errorf("failed to create the logs exporter: %w", err)
	}

	return exp, nil
}

func createMetricsExporter(
	_ context.Context,
	params component.ExporterCreateSettings,
	cfg config.Exporter,
) (component.MetricsExporter, error) {
	exp, err := newMetricsExporter(cfg.(*Config), params)
	if err != nil {
		return nil, fmt.Errorf("failed to create the metrics exporter: %w", err)
	}

	return exp, nil
}

func createTracesExporter(
	_ context.Context,
	params component.ExporterCreateSettings,
	cfg config.Exporter,
) (component.TracesExporter, error) {
	exp, err := newTracesExporter(cfg.(*Config), params)
	if err != nil {
		return nil, fmt.Errorf("failed to create the traces exporter: %w", err)
	}

	return exp, nil
}
