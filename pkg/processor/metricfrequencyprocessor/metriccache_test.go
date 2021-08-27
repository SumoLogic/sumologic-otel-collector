package metricfrequencyprocessor

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/model/pdata"
)

func TestEmptyRead(t *testing.T) {
	cache := newMetricCache(createDefaultConfig().(*Config))

	result := cache.List("a")

	assert.Equal(t, result, emptyResult)
}

func TestSingleRegister(t *testing.T) {
	cache := newMetricCache(createDefaultConfig().(*Config))
	cache.Register("a", newDataPoint(timestamp1, 0.0))

	result := cache.List("a")

	assert.Equal(t, result, map[pdata.Timestamp]float64{timestamp1: 0.0})
}

func TestTwoRegistersOfSingleMetric(t *testing.T) {
	cache := newMetricCache(createDefaultConfig().(*Config))
	cache.Register("a", newDataPoint(timestamp1, 0.0))
	cache.Register("a", newDataPoint(timestamp2, 1.0))

	result := cache.List("a")

	assert.Equal(t, result, map[pdata.Timestamp]float64{timestamp1: 0.0, timestamp2: 1.0})
}

func TestTwoRegistersOnTwoMetrics(t *testing.T) {
	cache := newMetricCache(createDefaultConfig().(*Config))
	cache.Register("a", newDataPoint(timestamp1, 0.0))
	cache.Register("b", newDataPoint(timestamp2, 1.0))

	result1 := cache.List("a")
	result2 := cache.List("b")

	assert.Equal(t, result1, map[pdata.Timestamp]float64{timestamp1: 0.0})
	assert.Equal(t, result2, map[pdata.Timestamp]float64{timestamp2: 1.0})
}

var emptyResult = make(map[pdata.Timestamp]float64)
var timestamp1 = pdata.TimestampFromTime(time.Unix(0, 0))
var timestamp2 = pdata.TimestampFromTime(time.Unix(1, 0))

func newDataPoint(timestamp pdata.Timestamp, value float64) pdata.NumberDataPoint {
	result := pdata.NewNumberDataPoint()
	result.SetTimestamp(timestamp)
	result.SetDoubleVal(value)
	return result
}
