package metricfrequencyprocessor

import "go.opentelemetry.io/collector/config"

// Config defines configuration for Source processor.
type Config struct {
	*config.ProcessorSettings `mapstructure:"-"`

	MinPointAccumulationSeconds           int     `mapstructure:"min_point_accumulation_seconds"`
	ConstantMetricsReportFrequencySeconds int64   `mapstructure:"constant_metrics_report_frequency_seconds"`
	LowInfoMetricsReportFrequencySeconds  int     `mapstructure:"low_info_metrics_report_frequency_seconds"`
	MaxReportFrequencySeconds             int     `mapstructure:"max_report_frequency_seconds"`
	IqrAnomalyCoef                        float64 `mapstructure:"iqr_anomaly_coefficient"`
	VariationIqrThresholdCoef             float64 `mapstructure:"variation_iqr_threshold_coefficient"`
	DataPointExpirationMinutes            int     `mapstructure:"data_point_expiration_minutes"`
	DataPointCacheCleanupIntervalMinutes  int     `mapstructure:"data_point_cache_cleanup_interval_minutes"`
	MetricCacheCleanupIntervalMinutes     int     `mapstructure:"metric_cache_cleanup_interval_minutes"`
}
