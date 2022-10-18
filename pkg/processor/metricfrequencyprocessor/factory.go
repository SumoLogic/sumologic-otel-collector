package metricfrequencyprocessor

import (
	"context"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/processor/processorhelper"
)

const (
	cfgType = "metric_frequency"

	defaultMinPointAccumulationTime       = 15 * time.Minute
	defaultConstantMetricsReportFrequency = 5 * time.Minute
	defaultLowInfoMetricsReportFrequency  = 2 * time.Minute
	defaultMaxReportFrequency             = 30 * time.Second
	defaultIqrAnomalyCoef                 = 1.5
	defaultVariationIqrThresholdCoef      = 4.0
	defaultDataPointExpirationTime        = 1 * time.Hour
	defaultDataPointCacheCleanupInterval  = 10 * time.Minute
	defaultMetricCacheCleanupInterval     = 3 * time.Hour
	stabilityLevel                        = component.StabilityLevelBeta
)

func NewFactory() component.ProcessorFactory {
	return component.NewProcessorFactory(
		cfgType,
		createDefaultConfig,
		component.WithMetricsProcessor(createMetricsProcessor, stabilityLevel),
	)
}

func createDefaultConfig() config.Processor {
	id := config.NewComponentID(cfgType)
	ps := config.NewProcessorSettings(id)

	return &Config{
		&ps,
		sieveConfig{
			MinPointAccumulationTime:       defaultMinPointAccumulationTime,
			ConstantMetricsReportFrequency: defaultConstantMetricsReportFrequency,
			LowInfoMetricsReportFrequency:  defaultLowInfoMetricsReportFrequency,
			MaxReportFrequency:             defaultMaxReportFrequency,
			IqrAnomalyCoef:                 defaultIqrAnomalyCoef,
			VariationIqrThresholdCoef:      defaultVariationIqrThresholdCoef,
		},
		cacheConfig{
			DataPointExpirationTime:       defaultDataPointExpirationTime,
			DataPointCacheCleanupInterval: defaultDataPointCacheCleanupInterval,
			MetricCacheCleanupInterval:    defaultMetricCacheCleanupInterval,
		},
	}
}

func createMetricsProcessor(
	ctx context.Context,
	params component.ProcessorCreateSettings,
	cfg config.Processor,
	nextConsumer consumer.Metrics,
) (component.MetricsProcessor, error) {
	var internalProcessor = &metricsfrequencyprocessor{
		sieve: newMetricSieve(cfg.(*Config)),
	}
	return processorhelper.NewMetricsProcessor(ctx, params, cfg, nextConsumer, internalProcessor.ProcessMetrics)
}
