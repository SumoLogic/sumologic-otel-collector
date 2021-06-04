package metricfrequencyprocessor

import (
	"context"
	"go.opentelemetry.io/collector/consumer/pdata"
)

type metricsfrequencyprocessor struct {
}

func (mfp *metricsfrequencyprocessor) ProcessMetrics(_ context.Context, md pdata.Metrics) (pdata.Metrics, error) {
	rms := md.ResourceMetrics()
	out := pdata.NewMetrics()
	newRms := pdata.NewResourceMetricsSlice()
	for i := 0; i < rms.Len(); i++ {
		rm := rms.At(i)
		ilms := rm.InstrumentationLibraryMetrics()
		newRm := pdata.NewResourceMetrics()
		rm.Resource().CopyTo(newRm.Resource())
		newIlms := pdata.NewInstrumentationLibraryMetricsSlice()
		for j := 0; j < ilms.Len(); j++ {
			ilm := ilms.At(j)
			metrics := ilm.Metrics()
			newIlm := pdata.NewInstrumentationLibraryMetrics()
			ilm.InstrumentationLibrary().CopyTo(newIlm.InstrumentationLibrary())
			newMetrics := pdata.NewMetricSlice()
			for k := 0; k < metrics.Len(); k++ {
				metric := metrics.At(k)
				if mfp.filter(metric) {
					metric.CopyTo(newMetrics.AppendEmpty())
				}
			}
			if newMetrics.Len() > 0 {
				newMetrics.MoveAndAppendTo(newIlm.Metrics())
				newIlm.CopyTo(newIlms.AppendEmpty())
			}
		}
		if newIlms.Len() > 0 {
			newIlms.MoveAndAppendTo(newRm.InstrumentationLibraryMetrics())
			newRm.CopyTo(newRms.AppendEmpty())
		}
	}
	if newRms.Len() > 0 {
		newRms.MoveAndAppendTo(out.ResourceMetrics())
	}

	return out, nil
}

func (mfp *metricsfrequencyprocessor) filter(metric pdata.Metric) bool {
	return true
}
