package metricfrequencyprocessor

import (
	"math"
	"sort"

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
)

type pdataTimestampByValue []pcommon.Timestamp

func (pta pdataTimestampByValue) Len() int {
	return len(pta)
}

func (pta pdataTimestampByValue) Less(i, j int) bool {
	return pta[i] < pta[j]
}

func (pta pdataTimestampByValue) Swap(i, j int) {
	pta[i], pta[j] = pta[j], pta[i]
}

func sortTimestampArray(timestamps []pcommon.Timestamp) {
	sort.Sort(pdataTimestampByValue(timestamps))
}

func getVal(point pmetric.NumberDataPoint) float64 {
	switch point.ValueType() {
	case pmetric.NumberDataPointValueTypeDouble:
		return point.DoubleValue()
	case pmetric.NumberDataPointValueTypeInt:
		return float64(point.IntValue())
	default:
		return math.NaN()
	}
}
