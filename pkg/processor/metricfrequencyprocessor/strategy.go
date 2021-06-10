package metricfrequencyprocessor

import (
	"go.opentelemetry.io/collector/consumer/pdata"
	"math"
	"sort"
	"time"
)

const (
	float64EqualityThreshold = 1e-9
	minPointAccumulationTime = time.Minute * 15
	category1ReportFrequency = time.Second * 30
	category2ReportFrequency = time.Minute * 2
	category3ReportFrequency = time.Minute * 5
	safetyInterval           = time.Second * 1

	iqrAnomalyCoef            = 1.5
	variationIqrThresholdCoef = 4
)

type MetricSieve struct {
	metricCache  MetricCache
	lastReported map[string]pdata.Timestamp
}

func (fs *MetricSieve) Sift(metric pdata.Metric) bool {
	if metric.DataType() == pdata.MetricDataTypeDoubleGauge {
		return fs.siftDropGauge(metric)
	} else {
		return false
	}
}

func (fs *MetricSieve) siftDropGauge(metric pdata.Metric) bool {
	metric.DoubleGauge().DataPoints().RemoveIf(fs.siftDataPoint(metric.Name()))

	return metric.DoubleGauge().DataPoints().Len() == 0
}

func (fs *MetricSieve) siftDataPoint(name string) func(pdata.DoubleDataPoint) bool {
	return func(dataPoint pdata.DoubleDataPoint) bool {
		cachedPoints := fs.metricCache.List(name)
		fs.metricCache.Register(name, dataPoint)
		lastReported, _ := fs.lastReported[name]
		earliest := earliestTimestamp(cachedPoints)
		cachedPoints[dataPoint.Timestamp()] = dataPoint.Value()

		if metricRequiresSamples(dataPoint, earliest) {
			fs.lastReported[name] = dataPoint.Timestamp()
			return false
		}

		if pastCategoryFrequency(dataPoint, lastReported, category3ReportFrequency) {
			fs.lastReported[name] = dataPoint.Timestamp()
			return false
		}

		if isConstant(dataPoint, cachedPoints) {
			return true
		}

		if pastCategoryFrequency(dataPoint, lastReported, category2ReportFrequency) {
			fs.lastReported[name] = dataPoint.Timestamp()
			return false
		}

		if isLowInformation(cachedPoints) {
			return true
		}

		if pastCategoryFrequency(dataPoint, lastReported, category1ReportFrequency) {
			fs.lastReported[name] = dataPoint.Timestamp()
			return false
		}

		return true
	}
}

func metricRequiresSamples(point pdata.DoubleDataPoint, earliest pdata.Timestamp) bool {
	return point.Timestamp().AsTime().Before(earliest.AsTime().Add(minPointAccumulationTime))
}

func pastCategoryFrequency(point pdata.DoubleDataPoint, lastReport pdata.Timestamp, categoryFrequency time.Duration) bool {
	return point.Timestamp().AsTime().After(lastReport.AsTime().Add(categoryFrequency).Add(safetyInterval))
}

func isConstant(point pdata.DoubleDataPoint, points map[pdata.Timestamp]float64) bool {
	for _, value := range points {
		if !almostEqual(point.Value(), value) {
			return false
		}
	}

	return true
}

func isLowInformation(points map[pdata.Timestamp]float64) bool {
	q1, q3 := calculateQ1Q3(points)
	iqr := q3 - q1
	variation := calculateVariation(points)

	return withingBounds(points, q1-iqrAnomalyCoef*iqr, q3+iqrAnomalyCoef*iqr) && lowVariation(variation, iqr)
}

func calculateQ1Q3(points map[pdata.Timestamp]float64) (float64, float64) {
	values := valueSlice(points)
	sort.Float64s(values)
	q1Index := len(points) / 4
	q3Index := 3 * len(points) / 4
	return values[q1Index], values[q3Index]
}

func withingBounds(points map[pdata.Timestamp]float64, lowerBound float64, upperBound float64) bool {
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

func calculateVariation(points map[pdata.Timestamp]float64) float64 {
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

func lowVariation(variation float64, iqr float64) bool {
	return variation < variationIqrThresholdCoef*iqr
}

func earliestTimestamp(points map[pdata.Timestamp]float64) pdata.Timestamp {
	min := pdata.TimestampFromTime(time.Now())
	for k := range points {
		if k < min {
			min = k
		}
	}

	return min
}

func keySlice(mapping map[pdata.Timestamp]float64) []pdata.Timestamp {
	out := make([]pdata.Timestamp, len(mapping))
	i := 0
	for k := range mapping {
		out[i] = k
		i++
	}

	return out
}

func valueSlice(mapping map[pdata.Timestamp]float64) []float64 {
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
