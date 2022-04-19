package metricfrequencyprocessor

import "go.opentelemetry.io/collector/pdata/pmetric"

type siftAllSieve struct{}

func (s *siftAllSieve) Sift(metric pmetric.Metric) bool {
	return true
}

type keepAllSieve struct{}

func (s *keepAllSieve) Sift(metric pmetric.Metric) bool {
	return false
}

type singleMetricSieve struct {
	name string
}

func (s *singleMetricSieve) Sift(metric pmetric.Metric) bool {
	return metric.Name() == s.name
}
