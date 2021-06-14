package metricfrequencyprocessor

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/processor/processorhelper"
)

const cfgType = "metric_frequency"

func NewFactory() component.ProcessorFactory {
	return processorhelper.NewFactory(
		cfgType,
		nil,
		processorhelper.WithMetrics(createMetricsProcessor),
	)
}

func createMetricsProcessor(
	_ context.Context,
	_ component.ProcessorCreateParams,
	cfg config.Processor,
	nextConsumer consumer.Metrics,
) (component.MetricsProcessor, error) {
	return processorhelper.NewMetricsProcessor(cfg, nextConsumer, &metricsfrequencyprocessor{})
}
