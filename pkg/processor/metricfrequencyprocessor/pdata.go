package metricfrequencyprocessor

import (
	"sort"

	"go.opentelemetry.io/collector/model/pdata"
)

type pdataTimestampByValue []pdata.Timestamp

func (pta pdataTimestampByValue) Len() int {
	return len(pta)
}

func (pta pdataTimestampByValue) Less(i, j int) bool {
	return pta[i] < pta[j]
}

func (pta pdataTimestampByValue) Swap(i, j int) {
	pta[i], pta[j] = pta[j], pta[i]
}

func sortTimestampArray(timestamps []pdata.Timestamp) {
	sort.Sort(pdataTimestampByValue(timestamps))
}
