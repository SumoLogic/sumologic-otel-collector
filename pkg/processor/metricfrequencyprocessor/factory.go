package metricfrequencyprocessor

import (
	"context"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/processor"
	"go.opentelemetry.io/collector/processor/processorhelper"
)

const (
	typeStr = "metric_frequency"

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

var Type = component.MustNewType(typeStr)

func NewFactory() processor.Factory {
	return processor.NewFactory(
		Type,
		createDefaultConfig,
		processor.WithMetrics(createMetricsProcessor, stabilityLevel),
	)
}

func createDefaultConfig() component.Config {
	return &Config{
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
	params processor.Settings,
	cfg component.Config,
	nextConsumer consumer.Metrics,
) (processor.Metrics, error) {
	var internalProcessor = &metricsfrequencyprocessor{
		sieve: newMetricSieve(cfg.(*Config)),
	}
	return processorhelper.NewMetrics(ctx, params, cfg, nextConsumer, internalProcessor.ProcessMetrics)
}
