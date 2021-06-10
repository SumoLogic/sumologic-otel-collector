package metricfrequencyprocessor

import (
	"go.opentelemetry.io/collector/consumer/pdata"
	"sort"
)

type PdataTimestampArr []pdata.Timestamp

func (pta PdataTimestampArr) Len() int {
	return len(pta)
}

func (pta PdataTimestampArr) Less(i, j int) bool {
	return pta[i] < pta[j]
}

func (pta PdataTimestampArr) Swap(i, j int) {
	placeholder := pta[i]
	pta[i] = pta[j]
	pta[j] = placeholder
}

func sortTimestampArray(timestamps []pdata.Timestamp) {
	sort.Sort(PdataTimestampArr(timestamps))
}
