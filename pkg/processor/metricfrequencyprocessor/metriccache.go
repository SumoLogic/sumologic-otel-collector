package metricfrequencyprocessor

import (
	"github.com/patrickmn/go-cache"
	"go.opentelemetry.io/collector/consumer/pdata"
	"time"
)

type DataPoint struct {
	Timestamp pdata.Timestamp
	Value     float64
}

type metriccache struct {
	internalCaches map[string]*cache.Cache
}

func (mc *metriccache) Register(metric pdata.Metric) {
	if metric.DataType() != pdata.MetricDataTypeDoubleGauge {
		return
	}
	gauge := metric.DoubleGauge().DataPoints()

	internalCache, found := mc.internalCaches[metric.Name()]
	if !found {
		newCache := mc.newCache()
		mc.internalCaches[metric.Name()] = newCache
		internalCache = newCache
	}

	for i := 0; i < gauge.Len(); i++ {
		key := gauge.At(i).Timestamp().String()
		value := &DataPoint{Timestamp: gauge.At(i).Timestamp(), Value: gauge.At(i).Value()}
		internalCache.Set(key, value, cache.DefaultExpiration)
	}
}

func (mc *metriccache) List(metricName string) map[pdata.Timestamp]float64 {
	out := make(map[pdata.Timestamp]float64)
	internalCache, found := mc.internalCaches[metricName]
	if found {
		for _, item := range internalCache.Items() {
			dataPoint := item.Object.(*DataPoint)
			out[dataPoint.Timestamp] = dataPoint.Value
		}
	}

	return out
}

func (mc *metriccache) Cleanup() {
	for key, internalCache := range mc.internalCaches {
		if internalCache.ItemCount() == 0 {
			delete(mc.internalCaches, key)
		}
	}
}

func (mc *metriccache) newCache() *cache.Cache {
	return cache.New(time.Hour*1, time.Minute*10)
}
