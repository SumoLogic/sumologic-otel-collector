package metricfrequencyprocessor

import (
	"fmt"
	"time"

	"github.com/patrickmn/go-cache"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
)

type DataPoint struct {
	Timestamp pcommon.Timestamp
	Value     float64
}

// metricCache caches data points into two level mapping structure.
// To easily list all data points of a given metric it keeps a separate cache for each incoming metric.
type metricCache struct {
	config cacheConfig

	internalCaches map[string]*cache.Cache
}

func newMetricCache(config cacheConfig) *metricCache {
	c := &metricCache{
		config:         config,
		internalCaches: make(map[string]*cache.Cache),
	}

	go func(c *metricCache) {
		t := time.NewTicker(c.config.MetricCacheCleanupInterval)
		defer t.Stop()
		for range t.C {
			c.Cleanup()
		}
	}(c)

	return c
}

func (mc *metricCache) Register(name string, dataPoint pmetric.NumberDataPoint) {

	internalCache, exists := mc.internalCaches[name]
	if !exists {
		newCache := mc.newCache()
		mc.internalCaches[name] = newCache
		internalCache = newCache
	}

	key := dataPoint.Timestamp().String()
	value := &DataPoint{Timestamp: dataPoint.Timestamp(), Value: getVal(dataPoint)}
	internalCache.Set(key, value, cache.DefaultExpiration)
}

func (mc *metricCache) List(metricName string) map[pcommon.Timestamp]float64 {
	out := make(map[pcommon.Timestamp]float64)
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

func (mc *metricCache) Cleanup() {
	for key, internalCache := range mc.internalCaches {
		if internalCache.ItemCount() == 0 {
			delete(mc.internalCaches, key)
		}
	}
}

func (mc *metricCache) newCache() *cache.Cache {
	return cache.New(mc.config.DataPointExpirationTime, mc.config.DataPointCacheCleanupInterval)
}
