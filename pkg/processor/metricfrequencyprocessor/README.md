# Metric Frequency Processor

**Stability level**: Beta

The `metricfrequencyprocessor` is a metrics processor that helps reduce DPM by automatic tuning of metrics reporting
frequency which adjusts for metric's information volume.

For `metricfrequencyprocessor`, there are three categories of metrics - constant, low information and regular. Constant
ones are self-explanatory. Low information ones are defined by heuristic - having no anomalous data points and having
small relative variation. All metrics not falling to above categories are regular and do not have their report frequency
tuned beyond `maxReportFrequency`.

Metrics are categorised by their recent data points, so a category for a metric can change in time.

`metricfrequencyprocessor` works by sifting out data points that would be reported earlier than according to their
category's frequency.

## Config

- `min_point_accumulation_time` - warm up time for processor. Processor won't sift any data point from a metric with no
  earlier data point older than this value.
- `constant_metrics_report_frequency` - minimum time between reports of a constant metric.
- `low_info_metrics_report_frequency` - minimum time between reports of a low info metric.
- `max_report_frequency` - minimum time between reports of any metric.

### Low info definition

- `iqr_anomaly_coefficient` - relative deviation from interquartile range which constitutes an anomaly.
- `variation_iqr_threshold_coefficient` - variation to iqr quotient under which a metric is considered low info.

### Data point caching

- `data_point_expiration_time` - how long a data point should be used for determining metrics category.
- `data_point_cache_cleanup_interval` - how often expired data points are removed from memory.
- `metric_cache_cleanup_interval` - how often no longer seen metrics are removed from memory.

## Example config

```yaml
processors:
  metric_frequency:
    min_point_accumulation_time: 15m
    constant_metrics_report_frequency: 5m
    low_info_metrics_report_frequency: 2m
    max_report_frequency: 30s
    data_point_expiration_time: 1h
```
