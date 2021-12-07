package luaprocessor

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/processor/processorhelper"
)

const (
	// The value of "type" key in configuration.
	typeStr = "lua"
)

var processorCapabilities = consumer.Capabilities{MutatesData: true}

// NewFactory returns a new factory for the Lua processor.
func NewFactory() component.ProcessorFactory {
	return processorhelper.NewFactory(
		typeStr,
		createDefaultConfig,
		processorhelper.WithTraces(createTracesProcessor),
		processorhelper.WithMetrics(createMetricsProcessor),
		processorhelper.WithLogs(createLogsProcessor),
	)
}

func createDefaultConfig() config.Processor {
	ps := config.NewProcessorSettings(config.NewComponentID(typeStr))
	return &Config{
		ProcessorSettings: &ps,
	}
}

// createMetricsProcessor creates a metrics processor based on this config
func createMetricsProcessor(
	_ context.Context,
	params component.ProcessorCreateSettings,
	cfg config.Processor,
	next consumer.Metrics,
) (component.MetricsProcessor, error) {
	oCfg := cfg.(*Config)

	sp := newLuaProcessor(oCfg)
	return processorhelper.NewMetricsProcessor(
		cfg,
		next,
		sp.ProcessMetrics,
		processorhelper.WithCapabilities(processorCapabilities),
	)
}

// createTracesProcessor creates a traces processor based on this config
func createTracesProcessor(
	ctx context.Context,
	params component.ProcessorCreateSettings,
	cfg config.Processor,
	next consumer.Traces,
) (component.TracesProcessor, error) {
	oCfg := cfg.(*Config)

	sp := newLuaProcessor(oCfg)
	return processorhelper.NewTracesProcessor(
		cfg,
		next,
		sp.ProcessTraces,
		processorhelper.WithCapabilities(processorCapabilities),
	)
}

// createLogsProcessor creates a logs processor based on this config
func createLogsProcessor(
	_ context.Context,
	params component.ProcessorCreateSettings,
	cfg config.Processor,
	next consumer.Logs,
) (component.LogsProcessor, error) {
	oCfg := cfg.(*Config)

	sp := newLuaProcessor(oCfg)
	return processorhelper.NewLogsProcessor(
		cfg,
		next,
		sp.ProcessLogs,
		processorhelper.WithCapabilities(processorCapabilities),
	)
}
