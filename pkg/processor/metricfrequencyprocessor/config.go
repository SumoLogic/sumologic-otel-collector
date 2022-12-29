package metricfrequencyprocessor

import (
	"time"
)

// Config defines configuration for metricfrequencyprocessor.
type Config struct {
	sieveConfig `mapstructure:",squash"`
	cacheConfig `mapstructure:",squash"`
}

type sieveConfig struct {
	// MinPointAccumulationTime defines warm up time for the processor.
	// Processor will not sift data points for metrics
	// which do not have data points in processor's cache older than MinPointAccumulationTime.
	MinPointAccumulationTime time.Duration `mapstructure:"min_point_accumulation_time"`

	// ConstantMetricsReportFrequency defines minimum time between reports of a constant metric.
	ConstantMetricsReportFrequency time.Duration `mapstructure:"constant_metrics_report_frequency"`

	// LowInfoMetricsReportFrequency defines minimum time between reports of a low info metric.
	LowInfoMetricsReportFrequency time.Duration `mapstructure:"low_info_metrics_report_frequency"`

	// MaxReportFrequency defines minimum time between reports of any metric.
	MaxReportFrequency time.Duration `mapstructure:"max_report_frequency"`

	// IqrAnomalyCoef defines relative deviation from interquartile range which constitutes an anomaly.
	// I.e. value v such that v > Iqr * IqrAnomalyCoef is considered an anomaly.
	IqrAnomalyCoef float64 `mapstructure:"iqr_anomaly_coefficient"`

	// VariationIqrThresholdCoef variation to iqr quotient above which a metric is no longer considered low info.
	// Variation means sum of absolute values of differences between consecutive data points.
	// I.e. if current variation v of a metric satisfies v / Iqr > VariationIqrThresholdCoef
	// then the metric is not considered low info.
	VariationIqrThresholdCoef float64 `mapstructure:"variation_iqr_threshold_coefficient"`
}

type cacheConfig struct {
	// DataPointExpirationTime defines how long a data point should be used for determining metric's category.
	DataPointExpirationTime time.Duration `mapstructure:"data_point_expiration_time"`

	// DataPointCacheCleanupInterval defines how often expired data points are removed from memory.
	DataPointCacheCleanupInterval time.Duration `mapstructure:"data_point_cache_cleanup_interval"`

	// MetricCacheCleanupInterval defines how often no longer seen metrics are removed from memory.
	MetricCacheCleanupInterval time.Duration `mapstructure:"metric_cache_cleanup_interval"`
}
