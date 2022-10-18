package metricfrequencyprocessor

import (
	"context"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
)

func TestSieveAllFromEmpty(t *testing.T) {
	sieve := &siftAllSieve{}
	processor := &metricsfrequencyprocessor{sieve: sieve}

	input := createMetrics()
	result, err := processor.ProcessMetrics(context.Background(), input)

	require.NoError(t, err)
	assert.Equal(t, result.ResourceMetrics().Len(), 0)
}

func TestSieveAllFromNonempty(t *testing.T) {
	sieve := &siftAllSieve{}
	processor := &metricsfrequencyprocessor{sieve: sieve}

	resource1metrics := map[string][]string{
		"lib-1": {"m1", "m2"},
		"lib-2": {"m1", "m3"},
	}
	resource2metrics := map[string][]string{
		"lib-1": {"m1", "m2"},
		"lib-3": {"m1", "m3"},
		"lib-4": {"m2", "m3"},
	}
	input := createMetrics(resource1metrics, resource2metrics)
	result, err := processor.ProcessMetrics(context.Background(), input)

	require.NoError(t, err)
	assert.Equal(t, result.ResourceMetrics().Len(), 0)
}

func TestKeepAllFromEmpty(t *testing.T) {
	sieve := &keepAllSieve{}
	processor := &metricsfrequencyprocessor{sieve: sieve}

	input := createMetrics()
	result, err := processor.ProcessMetrics(context.Background(), input)

	require.NoError(t, err)
	assert.Equal(t, result.ResourceMetrics().Len(), 0)
}

func TestKeepAllFromNonEmpty(t *testing.T) {
	sieve := &keepAllSieve{}
	processor := &metricsfrequencyprocessor{sieve: sieve}

	resource1metrics := map[string][]string{
		"lib-1": {"m1", "m2"},
		"lib-2": {"m1", "m2", "m3"},
	}
	resource2metrics := map[string][]string{
		"lib-1": {"m1"},
		"lib-3": {"m1", "m3"},
		"lib-4": {"m2", "m3", "m4"},
	}
	input := createMetrics(resource1metrics, resource2metrics)
	result, err := processor.ProcessMetrics(context.Background(), input)

	require.NoError(t, err)
	require.Equal(t, 2, result.ResourceMetrics().Len())
	assert.Equal(t, 2, result.ResourceMetrics().At(0).ScopeMetrics().Len())
	assert.Equal(t, 2, result.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().Len())
	assert.Equal(t, 3, result.ResourceMetrics().At(0).ScopeMetrics().At(1).Metrics().Len())
	require.Equal(t, 3, result.ResourceMetrics().At(1).ScopeMetrics().Len())
	assert.Equal(t, 1, result.ResourceMetrics().At(1).ScopeMetrics().At(0).Metrics().Len())
	assert.Equal(t, 2, result.ResourceMetrics().At(1).ScopeMetrics().At(1).Metrics().Len())
	assert.Equal(t, 3, result.ResourceMetrics().At(1).ScopeMetrics().At(2).Metrics().Len())
}

func TestSelectionWithSingleMetric(t *testing.T) {
	sieve := &singleMetricSieve{name: "m1"}
	processor := &metricsfrequencyprocessor{sieve: sieve}

	resourceMetrics := map[string][]string{
		"lib-1": {"m1"},
	}

	input := createMetrics(resourceMetrics)
	result, err := processor.ProcessMetrics(context.Background(), input)

	require.NoError(t, err)
	assert.Equal(t, 0, result.ResourceMetrics().Len())
}

func TestSelectionWithTwoResources(t *testing.T) {
	sieve := &singleMetricSieve{name: "m1"}
	processor := &metricsfrequencyprocessor{sieve: sieve}

	resource1metrics := map[string][]string{
		"lib-1": {"m1"},
	}
	resource2metrics := map[string][]string{
		"lib-1": {"m2"},
	}

	input := createMetrics(resource1metrics, resource2metrics)
	result, err := processor.ProcessMetrics(context.Background(), input)

	require.NoError(t, err)
	require.Equal(t, 1, result.ResourceMetrics().Len())
	assert.Equal(t, "m2", result.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().At(0).Name())
}

func TestSelectionWithTwoLibraries(t *testing.T) {
	sieve := &singleMetricSieve{name: "m1"}
	processor := &metricsfrequencyprocessor{sieve: sieve}

	resourceMetrics := map[string][]string{
		"lib-1": {"m1"},
		"lib-2": {"m1", "m2"},
	}

	input := createMetrics(resourceMetrics)
	result, err := processor.ProcessMetrics(context.Background(), input)

	require.NoError(t, err)
	require.Equal(t, 1, result.ResourceMetrics().Len())
	require.Equal(t, 1, result.ResourceMetrics().At(0).ScopeMetrics().Len())
	assert.Equal(t, "lib-2", result.ResourceMetrics().At(0).ScopeMetrics().At(0).Scope().Name())
	require.Equal(t, 1, result.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().Len())
	assert.Equal(t, "m2", result.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().At(0).Name())
}

func createGauge() pmetric.Gauge {
	dpSlice := pmetric.NewNumberDataPointSlice()
	pmetric.NewNumberDataPoint().CopyTo(dpSlice.AppendEmpty())

	gauge := pmetric.NewGauge()
	dpSlice.CopyTo(gauge.DataPoints())

	return gauge
}

func createMetric(name string) pmetric.Metric {
	metric := pmetric.NewMetric()
	metric.SetName(name)
	metric.SetEmptyGauge()
	createGauge().CopyTo(metric.Gauge())

	return metric
}

func createInstrumentedLibrary(name string) pcommon.InstrumentationScope {
	library := pcommon.NewInstrumentationScope()
	library.SetName(name)
	library.SetVersion("-")

	return library
}

func createScopeMetrics(name string, metricNames []string) pmetric.ScopeMetrics {
	sm := pmetric.NewScopeMetrics()
	createInstrumentedLibrary(name).CopyTo(sm.Scope())
	for _, metricName := range metricNames {
		createMetric(metricName).CopyTo(sm.Metrics().AppendEmpty())
	}

	return sm
}

func createRm(metricsPerLibrary map[string][]string) pmetric.ResourceMetrics {
	rm := pmetric.NewResourceMetrics()
	keys := getStringKeySlice(metricsPerLibrary)
	sort.Strings(keys)
	pcommon.NewResource().CopyTo(rm.Resource())
	for _, key := range keys {
		library := key
		metrics := metricsPerLibrary[library]
		createScopeMetrics(library, metrics).CopyTo(rm.ScopeMetrics().AppendEmpty())
	}

	return rm
}

func createMetrics(metricsPerLibraryArgs ...map[string][]string) pmetric.Metrics {
	metrics := pmetric.NewMetrics()
	for _, metricsPerLibrary := range metricsPerLibraryArgs {
		createRm(metricsPerLibrary).CopyTo(metrics.ResourceMetrics().AppendEmpty())
	}

	return metrics
}

func getStringKeySlice(mapping map[string][]string) []string {
	out := make([]string, len(mapping))
	i := 0
	for k := range mapping {
		out[i] = k
		i++
	}

	return out
}
