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

type MetricCache struct {
	internalCaches map[string]*cache.Cache
}

func (mc *MetricCache) Register(name string, dataPoint pdata.DoubleDataPoint) {

	internalCache, exists := mc.internalCaches[name]
	if !exists {
		newCache := newCache()
		mc.internalCaches[name] = newCache
		internalCache = newCache
	}

	key := dataPoint.Timestamp().String()
	value := &DataPoint{Timestamp: dataPoint.Timestamp(), Value: dataPoint.Value()}
	internalCache.Set(key, value, cache.DefaultExpiration)
}

func (mc *MetricCache) List(metricName string) map[pdata.Timestamp]float64 {
	out := make(map[pdata.Timestamp]float64)
	internalCache, found := mc.internalCaches[metricName]
	if found {
		for _, item := range internalCache.Items() {
			dataPoint, ok := item.Object.(*DataPoint)
			if !ok {
				panic(fmt.Sprintf("item.Object is not a DataPoint but a %T", item.Object))
			}
			out[dataPoint.Timestamp] = dataPoint.Value
		}
	}

	return out
}

func (mc *MetricCache) Cleanup() {
	for key, internalCache := range mc.internalCaches {
		if internalCache.ItemCount() == 0 {
			delete(mc.internalCaches, key)
		}
	}
}

func newCache() *cache.Cache {
	return cache.New(time.Hour*1, time.Minute*10)
}
