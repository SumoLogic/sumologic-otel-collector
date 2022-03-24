package sumologicschemaprocessor

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/processor/processorhelper"
)

const (
	// The value of "type" key in configuration.
	typeStr = "sumologic_schema"
)

var processorCapabilities = consumer.Capabilities{MutatesData: true}

// NewFactory returns a new factory for the processor.
func NewFactory() component.ProcessorFactory {
	return component.NewProcessorFactory(
		typeStr,
		createDefaultConfig,
		component.WithTracesProcessor(createTracesProcessor),
		component.WithMetricsProcessor(createMetricsProcessor),
		component.WithLogsProcessor(createLogsProcessor))
}

func createLogsProcessor(
	_ context.Context,
	set component.ProcessorCreateSettings,
	cfg config.Processor,
	nextConsumer consumer.Logs,
) (component.LogsProcessor, error) {
	processor, err := newSumologicSchemaProcessor(set, cfg.(*Config))
	if err != nil {
		return nil, err
	}
	return processorhelper.NewLogsProcessor(
		cfg,
		nextConsumer,
		processor.processLogs,
		processorhelper.WithCapabilities(processorCapabilities),
		processorhelper.WithStart(processor.start),
		processorhelper.WithShutdown(processor.shutdown))
}

func createMetricsProcessor(
	_ context.Context,
	set component.ProcessorCreateSettings,
	cfg config.Processor,
	nextConsumer consumer.Metrics,
) (component.MetricsProcessor, error) {
	processor, err := newSumologicSchemaProcessor(set, cfg.(*Config))
	if err != nil {
		return nil, err
	}
	return processorhelper.NewMetricsProcessor(
		cfg,
		nextConsumer,
		processor.processMetrics,
		processorhelper.WithCapabilities(processorCapabilities),
		processorhelper.WithStart(processor.start),
		processorhelper.WithShutdown(processor.shutdown))
}

func createTracesProcessor(
	_ context.Context,
	set component.ProcessorCreateSettings,
	cfg config.Processor,
	nextConsumer consumer.Traces,
) (component.TracesProcessor, error) {
	processor, err := newSumologicSchemaProcessor(set, cfg.(*Config))
	if err != nil {
		return nil, err
	}
	return processorhelper.NewTracesProcessor(
		cfg,
		nextConsumer,
		processor.processTraces,
		processorhelper.WithCapabilities(processorCapabilities),
		processorhelper.WithStart(processor.start),
		processorhelper.WithShutdown(processor.shutdown))
}
