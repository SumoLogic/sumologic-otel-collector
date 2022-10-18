package metricfrequencyprocessor

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
)

func TestEmptyRead(t *testing.T) {
	cache := newCache()

	result := cache.List("a")

	assert.Equal(t, emptyResult, result)
}

func TestSingleRegister(t *testing.T) {
	cache := newCache()
	cache.Register("a", newDataPoint(timestamp1, 0.0))

	result := cache.List("a")

	assert.Equal(t, map[pcommon.Timestamp]float64{timestamp1: 0.0}, result)
}

func TestTwoRegistersOfSingleMetric(t *testing.T) {
	cache := newCache()
	cache.Register("a", newDataPoint(timestamp1, 0.0))
	cache.Register("a", newDataPoint(timestamp2, 1.0))

	result := cache.List("a")

	assert.Equal(t, map[pcommon.Timestamp]float64{timestamp1: 0.0, timestamp2: 1.0}, result)
}

func TestTwoRegistersOnTwoMetrics(t *testing.T) {
	cache := newCache()
	cache.Register("a", newDataPoint(timestamp1, 0.0))
	cache.Register("b", newDataPoint(timestamp2, 1.0))

	result1 := cache.List("a")
	result2 := cache.List("b")

	assert.Equal(t, map[pcommon.Timestamp]float64{timestamp1: 0.0}, result1)
	assert.Equal(t, map[pcommon.Timestamp]float64{timestamp2: 1.0}, result2)
}

var emptyResult = make(map[pcommon.Timestamp]float64)
var timestamp1 = pcommon.NewTimestampFromTime(time.Unix(0, 0))
var timestamp2 = pcommon.NewTimestampFromTime(time.Unix(1, 0))

func newCache() *metricCache {
	return newMetricCache(createDefaultConfig().(*Config).cacheConfig)
}

func newDataPoint(timestamp pcommon.Timestamp, value float64) pmetric.NumberDataPoint {
	result := pmetric.NewNumberDataPoint()
	result.SetTimestamp(timestamp)
	result.SetDoubleValue(value)
	return result
}
