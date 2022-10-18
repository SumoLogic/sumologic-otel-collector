package metricfrequencyprocessor

import (
	"math"
	"sort"
	"time"

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
)

const (
	float64EqualityThreshold = 1e-9
	safetyInterval           = time.Second * 1
)

type metricSieve interface {
	Sift(metric pmetric.Metric) bool
}

// defaultMetricSieve removes data points from MetricSlices that would be reported more often than preset
// frequency for a given category.
// For metric sieve, there are three categories of metrics:
// 1) Constant metrics
// 2) Low info metrics - i.e. no anomaly in terms of iqr and low variation
// 3) All other metrics
type defaultMetricSieve struct {
	config sieveConfig

	metricCache  *metricCache
	lastReported map[string]pcommon.Timestamp
}

var _ metricSieve = (*defaultMetricSieve)(nil)

func newMetricSieve(config *Config) *defaultMetricSieve {
	return &defaultMetricSieve{
		metricCache:  newMetricCache(config.cacheConfig),
		lastReported: make(map[string]pcommon.Timestamp),
		config:       config.sieveConfig,
	}
}

// Sift removes data points from MetricSlices of the metric argument according to specified strategy.
// It returns true if the metric should be removed.
func (ms *defaultMetricSieve) Sift(metric pmetric.Metric) bool {
	switch metric.Type() {
	case pmetric.MetricTypeGauge:
		return ms.siftDropGauge(metric)
	default:
		return false
	}
}

func (ms *defaultMetricSieve) siftDropGauge(metric pmetric.Metric) bool {
	metric.Gauge().DataPoints().RemoveIf(ms.siftDataPoint(metric.Name()))

	return metric.Gauge().DataPoints().Len() == 0
}

func (ms *defaultMetricSieve) siftDataPoint(name string) func(pmetric.NumberDataPoint) bool {
	return func(dataPoint pmetric.NumberDataPoint) bool {
		if math.IsNaN(getVal(dataPoint)) {
			return false
		}

		cachedPoints := ms.metricCache.List(name)
		ms.metricCache.Register(name, dataPoint)
		lastReported, exists := ms.lastReported[name]
		if !exists {
			ms.lastReported[name] = dataPoint.Timestamp()
			return false
		}
		earliest := earliestTimestamp(cachedPoints)
		cachedPoints[dataPoint.Timestamp()] = getVal(dataPoint)

		if ms.metricRequiresSamples(dataPoint, earliest) {
			ms.lastReported[name] = dataPoint.Timestamp()
			return false
		}

		if pastCategoryFrequency(dataPoint, lastReported, ms.config.ConstantMetricsReportFrequency) {
			ms.lastReported[name] = dataPoint.Timestamp()
			return false
		}

		if isConstant(dataPoint, cachedPoints) {
			return true
		}

		if pastCategoryFrequency(dataPoint, lastReported, ms.config.LowInfoMetricsReportFrequency) {
			ms.lastReported[name] = dataPoint.Timestamp()
			return false
		}

		if ms.isLowInformation(cachedPoints) {
			return true
		}

		if pastCategoryFrequency(dataPoint, lastReported, ms.config.MaxReportFrequency) {
			ms.lastReported[name] = dataPoint.Timestamp()
			return false
		}

		return true
	}
}

func (ms *defaultMetricSieve) metricRequiresSamples(point pmetric.NumberDataPoint, earliest pcommon.Timestamp) bool {
	return point.Timestamp().AsTime().Before(earliest.AsTime().Add(ms.config.MinPointAccumulationTime))
}

func pastCategoryFrequency(point pmetric.NumberDataPoint, lastReport pcommon.Timestamp, categoryFrequency time.Duration) bool {
	return point.Timestamp().AsTime().Add(safetyInterval).After(lastReport.AsTime().Add(categoryFrequency))
}

func isConstant(point pmetric.NumberDataPoint, points map[pcommon.Timestamp]float64) bool {
	for _, value := range points {
		if !almostEqual(getVal(point), value) {
			return false
		}
	}

	return true
}

// isLowInformation is a heuristic attempt at defining uninteresting metrics. Requirements:
// 1) no big changes - defined by no iqr anomalies
// 2) little oscillations - defined by low variation
func (ms *defaultMetricSieve) isLowInformation(points map[pcommon.Timestamp]float64) bool {
	q1, q3 := calculateQ1Q3(points)
	iqr := q3 - q1
	variation := calculateVariation(points)

	noAnomaly := withinBounds(points, q1-ms.config.IqrAnomalyCoef*iqr, q3+ms.config.IqrAnomalyCoef*iqr)
	return noAnomaly && ms.lowVariation(variation, iqr)
}

// calculateQ1Q3 returns specific quantiles - it refers to quantiles .25 and .75 respectively
func calculateQ1Q3(points map[pcommon.Timestamp]float64) (float64, float64) {
	values := valueSlice(points)
	sort.Float64s(values)
	q1Index := len(points) / 4
	q3Index := 3 * len(points) / 4
	return values[q1Index], values[q3Index]
}

func withinBounds(points map[pcommon.Timestamp]float64, lowerBound float64, upperBound float64) bool {
	for _, v := range points {
		if v < lowerBound {
			return false
		}
		if v > upperBound {
			return false
		}
	}

	return true
}

// calculateVariation returns a sum of absolute values of differences of subsequent data points.
func calculateVariation(points map[pcommon.Timestamp]float64) float64 {
	keys := keySlice(points)
	sortTimestampArray(keys)

	variation := 0.0
	previous := keys[0]
	for i := 1; i < len(keys); i++ {
		current := keys[i]
		variation += math.Abs(points[current] - points[previous])
		previous = current
	}

	return variation
}

// lowVariation returns a heuristic check indicating that data points display little oscillations
func (ms *defaultMetricSieve) lowVariation(variation float64, iqr float64) bool {
	return variation < ms.config.VariationIqrThresholdCoef*iqr
}

func earliestTimestamp(points map[pcommon.Timestamp]float64) pcommon.Timestamp {
	min := pcommon.NewTimestampFromTime(time.Now())
	for k := range points {
		if k < min {
			min = k
		}
	}

	return min
}

func keySlice(mapping map[pcommon.Timestamp]float64) []pcommon.Timestamp {
	out := make([]pcommon.Timestamp, len(mapping))
	i := 0
	for k := range mapping {
		out[i] = k
		i++
	}

	return out
}

func valueSlice(mapping map[pcommon.Timestamp]float64) []float64 {
	out := make([]float64, len(mapping))
	i := 0
	for _, v := range mapping {
		out[i] = v
		i++
	}

	return out
}

func almostEqual(a, b float64) bool {
	return math.Abs(a-b) <= float64EqualityThreshold
}
