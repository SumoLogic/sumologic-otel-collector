package sumologicschemaprocessor

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/model/pdata"
	"go.uber.org/zap"
)

type sumologicSchemaProcessor struct {
	logger                  *zap.Logger
	cloudNamespaceProcessor *cloudNamespaceProcessor
}

func newSumologicSchemaProcessor(set component.ProcessorCreateSettings, config *Config) (*sumologicSchemaProcessor, error) {
	cloudNamespaceProcessor, err := newCloudNamespaceProcessor(config.AddCloudNamespace)
	if err != nil {
		return nil, err
	}

	processor := &sumologicSchemaProcessor{
		logger:                  set.Logger,
		cloudNamespaceProcessor: cloudNamespaceProcessor,
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

func (processor *sumologicSchemaProcessor) processLogs(_ context.Context, logs pdata.Logs) (pdata.Logs, error) {
	logs, err := processor.cloudNamespaceProcessor.processLogs(logs)
	if err != nil {
		return logs, err
	}

	return logs, nil
}

func (processor *sumologicSchemaProcessor) processMetrics(ctx context.Context, metrics pdata.Metrics) (pdata.Metrics, error) {
	metrics, err := processor.cloudNamespaceProcessor.processMetrics(metrics)
	if err != nil {
		return metrics, err
	}

	return metrics, nil
}

func (processor *sumologicSchemaProcessor) processTraces(ctx context.Context, traces pdata.Traces) (pdata.Traces, error) {
	traces, err := processor.cloudNamespaceProcessor.processTraces(traces)
	if err != nil {
		return traces, err
	}
	return traces, nil
}
