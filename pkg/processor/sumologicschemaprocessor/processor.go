// Copyright 2022 Sumo Logic, Inc.
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

package sumologicschemaprocessor

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.uber.org/zap"
)

type sumologicSchemaProcessor struct {
	logger                       *zap.Logger
	cloudNamespaceProcessor      *cloudNamespaceProcessor
	translateAttributesProcessor *translateAttributesProcessor
}

func newSumologicSchemaProcessor(set component.ProcessorCreateSettings, config *Config) (*sumologicSchemaProcessor, error) {
	cloudNamespaceProcessor, err := newCloudNamespaceProcessor(config.AddCloudNamespace)
	if err != nil {
		return nil, err
	}

	translateAttributesProcessor, err := newTranslateAttributesProcessor(config.TranslateAttributes)
	if err != nil {
		return nil, err
	}

	processor := &sumologicSchemaProcessor{
		logger:                       set.Logger,
		cloudNamespaceProcessor:      cloudNamespaceProcessor,
		translateAttributesProcessor: translateAttributesProcessor,
	}

	return processor, nil
}

func (processor *sumologicSchemaProcessor) start(_ context.Context, host component.Host) error {
	processor.logger.Info(
		"Processor sumologic_schema has started.",
		zap.Bool("add_cloud_namespace", processor.cloudNamespaceProcessor.addCloudNamespace),
	)
	return nil
}

func (processor *sumologicSchemaProcessor) shutdown(_ context.Context) error {
	processor.logger.Info("Processor sumologic_schema has shut down.")
	return nil
}

func (processor *sumologicSchemaProcessor) processLogs(_ context.Context, logs plog.Logs) (plog.Logs, error) {
	logs1, err := processor.cloudNamespaceProcessor.processLogs(logs)
	if err != nil {
		return logs1, err
	}

	logs2, err := processor.translateAttributesProcessor.processLogs(logs1)
	if err != nil {
		return logs2, err
	}

	return logs2, nil
}

func (processor *sumologicSchemaProcessor) processMetrics(ctx context.Context, metrics pmetric.Metrics) (pmetric.Metrics, error) {
	metrics1, err := processor.cloudNamespaceProcessor.processMetrics(metrics)
	if err != nil {
		return metrics1, err
	}

	metrics2, err := processor.translateAttributesProcessor.processMetrics(metrics1)
	if err != nil {
		return metrics2, err
	}

	return metrics2, nil
}

func (processor *sumologicSchemaProcessor) processTraces(ctx context.Context, traces ptrace.Traces) (ptrace.Traces, error) {
	traces1, err := processor.cloudNamespaceProcessor.processTraces(traces)
	if err != nil {
		return traces1, err
	}

	traces2, err := processor.translateAttributesProcessor.processTraces(traces)
	if err != nil {
		return traces2, err
	}

	return traces2, nil
}
