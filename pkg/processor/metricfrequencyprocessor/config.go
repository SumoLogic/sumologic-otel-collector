package metricfrequencyprocessor

import (
	"time"

	"go.opentelemetry.io/collector/config"
)

// Config defines configuration for ProcessMetrics.
type Config struct {
	*config.ProcessorSettings `mapstructure:"-"`

	// MinPointAccumulationTime defines warm up time for the processor.
	// Processor will not sift data points for metrics
	// which do not have data points in processor's cache older than MinPointAccumulationTime.
	MinPointAccumulationTime time.Duration `mapstructure:"min_point_accumulation_time"`

	// ConstantMetricsReportFrequency defines minimum time between reports of a constant metric.
	ConstantMetricsReportFrequency time.Duration `mapstructure:"constant_metrics_report_frequency"`

	// LowInfoMetricsReportFrequency defines minimum time between reports of a low info metric.
	LowInfoMetricsReportFrequency time.Duration `mapstructure:"low_info_metrics_report_frequency"`

	// MaxReportFrequency defines minimum time between reports of any metric.
	MaxReportFrequency time.Duration `mapstructure:"max_report_frequency_seconds"`

	// IqrAnomalyCoef defines relative deviation from interquartile range which constitutes an anomaly.
	// I.e. value v such that v > Iqr * IqrAnomalyCoef is considered an anomaly.
	IqrAnomalyCoef float64 `mapstructure:"iqr_anomaly_coefficient"`

	// VariationIqrThresholdCoef variation to iqr quotient above which a metric is no longer considered low info.
	// Variation means sum of absolute values of differences between consecutive data points.
	// I.e. if current variation v of a metric satisfies v / Iqr > VariationIqrThresholdCoef
	// then the metric is not considered low info.
	VariationIqrThresholdCoef float64 `mapstructure:"variation_iqr_threshold_coefficient"`

	// DataPointExpirationTime defines how long a data point should be used for determining metric's category.
	DataPointExpirationTime time.Duration `mapstructure:"data_point_expiration_time"`

	// DataPointCacheCleanupInterval defines how often expired data points are removed from memory.
	DataPointCacheCleanupInterval time.Duration `mapstructure:"data_point_cache_cleanup_interval"`

	// MetricCacheCleanupInterval defines how often no longer seen metrics are removed from memory.
	MetricCacheCleanupInterval time.Duration `mapstructure:"metric_cache_cleanup_interval"`
}

type sieveConfig struct {
	MinPointAccumulationTime       time.Duration
	ConstantMetricsReportFrequency time.Duration
	LowInfoMetricsReportFrequency  time.Duration
	MaxReportFrequency             time.Duration
	IqrAnomalyCoef                 float64
	VariationIqrThresholdCoef      float64
}

func toSieveConfig(config *Config) *sieveConfig {
	return &sieveConfig{
		MinPointAccumulationTime:       config.MinPointAccumulationTime,
		ConstantMetricsReportFrequency: config.ConstantMetricsReportFrequency,
		LowInfoMetricsReportFrequency:  config.LowInfoMetricsReportFrequency,
		MaxReportFrequency:             config.MaxReportFrequency,
		IqrAnomalyCoef:                 config.IqrAnomalyCoef,
		VariationIqrThresholdCoef:      config.VariationIqrThresholdCoef,
	}
}

type cacheConfig struct {
	DataPointExpirationTime       time.Duration
	DataPointCacheCleanupInterval time.Duration
	MetricCacheCleanupInterval    time.Duration
}

func toCacheConfig(config *Config) *cacheConfig {
	return &cacheConfig{
		DataPointExpirationTime:       config.DataPointExpirationTime,
		DataPointCacheCleanupInterval: config.DataPointCacheCleanupInterval,
		MetricCacheCleanupInterval:    config.MetricCacheCleanupInterval,
	}
}
