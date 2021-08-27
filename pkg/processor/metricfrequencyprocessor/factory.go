package metricfrequencyprocessor

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/processor/processorhelper"
)

const (
	cfgType = "metric_frequency"

	defaultMinPointAccumulationSeconds           = 15 * 60
	defaultConstantMetricsReportFrequencySeconds = 5 * 60
	defaultLowInfoMetricsReportFrequencySeconds  = 2 * 60
	defaultMaxReportFrequencySeconds             = 30
	defaultIqrAnomalyCoef                        = 1.5
	defaultVariationIqrThresholdCoef             = 4.0
	defaultDataPointExpirationMinutes            = 60
	defaultDataPointCacheCleanupIntervalMinutes  = 10
	defaultMetricCacheCleanupIntervalMinutes     = 180
)

func NewFactory() component.ProcessorFactory {
	return processorhelper.NewFactory(
		cfgType,
		createDefaultConfig,
		processorhelper.WithMetrics(createMetricsProcessor),
	)
}

func createDefaultConfig() config.Processor {
	id := config.NewID(cfgType)
	ps := config.NewProcessorSettings(id)

	return &Config{
		ProcessorSettings:                     &ps,
		MinPointAccumulationSeconds:           defaultMinPointAccumulationSeconds,
		ConstantMetricsReportFrequencySeconds: defaultConstantMetricsReportFrequencySeconds,
		LowInfoMetricsReportFrequencySeconds:  defaultLowInfoMetricsReportFrequencySeconds,
		MaxReportFrequencySeconds:             defaultMaxReportFrequencySeconds,
		IqrAnomalyCoef:                        defaultIqrAnomalyCoef,
		VariationIqrThresholdCoef:             defaultVariationIqrThresholdCoef,
		DataPointExpirationMinutes:            defaultDataPointExpirationMinutes,
		DataPointCacheCleanupIntervalMinutes:  defaultDataPointCacheCleanupIntervalMinutes,
		MetricCacheCleanupIntervalMinutes:     defaultMetricCacheCleanupIntervalMinutes,
	}
}

func createMetricsProcessor(
	_ context.Context,
	_ component.ProcessorCreateSettings,
	cfg config.Processor,
	nextConsumer consumer.Metrics,
) (component.MetricsProcessor, error) {
	var internalProcessor = &metricsfrequencyprocessor{
		sieve: newMetricSieve(cfg.(*Config)),
	}
	return processorhelper.NewMetricsProcessor(cfg, nextConsumer, internalProcessor.ProcessMetrics)
}
