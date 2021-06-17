package metricfrequencyprocessor

import (
	"context"

	"go.opentelemetry.io/collector/consumer/pdata"
	"go.opentelemetry.io/collector/processor/processorhelper"
)

type metricsfrequencyprocessor struct {
	sieve metricSieve
}

var _ processorhelper.MProcessor = (*metricsfrequencyprocessor)(nil)

// ProcessMetrics applies metricSieve to incoming metrics. It mutates the argument.
func (mfp *metricsfrequencyprocessor) ProcessMetrics(_ context.Context, md pdata.Metrics) (pdata.Metrics, error) {
	rms := md.ResourceMetrics()
	for i := 0; i < rms.Len(); i++ {
		rm := rms.At(i)
		ilms := rm.InstrumentationLibraryMetrics()
		for j := 0; j < ilms.Len(); j++ {
			ilm := ilms.At(j)
			metrics := ilm.Metrics()
			metrics.RemoveIf(mfp.sieve.Sift)
		}
		ilms.RemoveIf(metricSliceEmpty)
	}
	rms.RemoveIf(ilmSliceEmpty)

	return md, nil
}

func metricSliceEmpty(metrics pdata.InstrumentationLibraryMetrics) bool {
	return metrics.Metrics().Len() == 0
}

func ilmSliceEmpty(metrics pdata.ResourceMetrics) bool {
	return metrics.InstrumentationLibraryMetrics().Len() == 0
}
