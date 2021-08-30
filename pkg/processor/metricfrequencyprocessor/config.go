package metricfrequencyprocessor

import (
	"time"

	"go.opentelemetry.io/collector/config"
)

// Config defines configuration for Source processor.
type Config struct {
	*config.ProcessorSettings `mapstructure:"-"`

	MinPointAccumulationTime       time.Duration `mapstructure:"min_point_accumulation_time"`
	ConstantMetricsReportFrequency time.Duration `mapstructure:"constant_metrics_report_frequency"`
	LowInfoMetricsReportFrequency  time.Duration `mapstructure:"low_info_metrics_report_frequency"`
	MaxReportFrequency             time.Duration `mapstructure:"max_report_frequency_seconds"`
	IqrAnomalyCoef                 float64       `mapstructure:"iqr_anomaly_coefficient"`
	VariationIqrThresholdCoef      float64       `mapstructure:"variation_iqr_threshold_coefficient"`
	DataPointExpirationTime        time.Duration `mapstructure:"data_point_expiration_time"`
	DataPointCacheCleanupInterval  time.Duration `mapstructure:"data_point_cache_cleanup_interval"`
	MetricCacheCleanupInterval     time.Duration `mapstructure:"metric_cache_cleanup_interval"`
}
