package metricfrequencyprocessor

import (
	"context"

	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/processor/processorhelper"
)

type metricsfrequencyprocessor struct {
	sieve metricSieve
}

var _ processorhelper.ProcessMetricsFunc = (*metricsfrequencyprocessor)(nil).ProcessMetrics

// ProcessMetrics applies metricSieve to incoming metrics. It mutates the argument.
func (mfp *metricsfrequencyprocessor) ProcessMetrics(_ context.Context, md pmetric.Metrics) (pmetric.Metrics, error) {
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

func metricSliceEmpty(metrics pmetric.ScopeMetrics) bool {
	return metrics.Metrics().Len() == 0
}

func ilmSliceEmpty(metrics pmetric.ResourceMetrics) bool {
	return metrics.InstrumentationLibraryMetrics().Len() == 0
}
