package metricfrequencyprocessor

import (
	"go.opentelemetry.io/collector/model/pdata"
)

type siftAllSieve struct{}

func (s *siftAllSieve) Sift(metric pdata.Metric) bool {
	return true
}

type keepAllSieve struct{}

func (s *keepAllSieve) Sift(metric pdata.Metric) bool {
	return false
}

type singleMetricSieve struct {
	name string
}

func (s *singleMetricSieve) Sift(metric pdata.Metric) bool {
	return metric.Name() == s.name
}
