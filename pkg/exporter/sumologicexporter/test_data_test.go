// Copyright 2020 OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package sumologicexporter

import (
	"go.opentelemetry.io/collector/model/pdata"
)

func exampleIntMetric() metricPair {
	return buildExampleIntMetric(true)
}

func buildExampleIntMetric(fillData bool) metricPair {
	metric := pdata.NewMetric()
	metric.SetName("test.metric.data")
	metric.SetUnit("bytes")
	metric.SetDataType(pdata.MetricDataTypeSum)

	if fillData {
		dp := metric.Sum().DataPoints().AppendEmpty()
		dp.SetTimestamp(1605534165 * 1e9)
		dp.SetIntVal(14500)
	}

	attributes := pdata.NewAttributeMap()
	attributes.InsertString("test", "test_value")
	attributes.InsertString("test2", "second_value")

	return metricPair{
		metric:     metric,
		attributes: attributes,
	}
}

func exampleIntGaugeMetric() metricPair {
	return buildExampleIntGaugeMetric(true)
}

func buildExampleIntGaugeMetric(fillData bool) metricPair {
	metric := metricPair{
		attributes: pdata.NewAttributeMap(),
		metric:     pdata.NewMetric(),
	}

	metric.metric.SetDataType(pdata.MetricDataTypeGauge)
	metric.metric.SetName("gauge_metric_name")

	metric.attributes.InsertString("foo", "bar")

	if fillData {
		dp := metric.metric.Gauge().DataPoints().AppendEmpty()
		dp.Attributes().InsertString("remote_name", "156920")
		dp.Attributes().InsertString("url", "http://example_url")
		dp.SetIntVal(124)
		dp.SetTimestamp(1608124661.166 * 1e9)

		dp = metric.metric.Gauge().DataPoints().AppendEmpty()
		dp.Attributes().InsertString("remote_name", "156955")
		dp.Attributes().InsertString("url", "http://another_url")
		dp.SetIntVal(245)
		dp.SetTimestamp(1608124662.166 * 1e9)
	}

	return metric
}

func exampleDoubleGaugeMetric() metricPair {
	return buildExampleDoubleGaugeMetric(true)
}

func buildExampleDoubleGaugeMetric(fillData bool) metricPair {
	metric := metricPair{
		attributes: pdata.NewAttributeMap(),
		metric:     pdata.NewMetric(),
	}

	metric.metric.SetDataType(pdata.MetricDataTypeGauge)
	metric.metric.SetName("gauge_metric_name_double_test")

	metric.attributes.InsertString("foo", "bar")

	if fillData {
		dp := metric.metric.Gauge().DataPoints().AppendEmpty()
		dp.Attributes().InsertString("local_name", "156720")
		dp.Attributes().InsertString("endpoint", "http://example_url")
		dp.SetDoubleVal(33.4)
		dp.SetTimestamp(1608124661.169 * 1e9)

		dp = metric.metric.Gauge().DataPoints().AppendEmpty()
		dp.Attributes().InsertString("local_name", "156155")
		dp.Attributes().InsertString("endpoint", "http://another_url")
		dp.SetDoubleVal(56.8)
		dp.SetTimestamp(1608124662.186 * 1e9)
	}

	return metric
}

func exampleIntSumMetric() metricPair {
	return buildExampleIntSumMetric(true)
}

func buildExampleIntSumMetric(fillData bool) metricPair {
	metric := metricPair{
		attributes: pdata.NewAttributeMap(),
		metric:     pdata.NewMetric(),
	}

	metric.metric.SetDataType(pdata.MetricDataTypeSum)
	metric.metric.SetName("sum_metric_int_test")

	metric.attributes.InsertString("foo", "bar")

	if fillData {
		dp := metric.metric.Sum().DataPoints().AppendEmpty()
		dp.Attributes().InsertString("name", "156720")
		dp.Attributes().InsertString("address", "http://example_url")
		dp.SetIntVal(45)
		dp.SetTimestamp(1608124444.169 * 1e9)

		dp = metric.metric.Sum().DataPoints().AppendEmpty()
		dp.Attributes().InsertString("name", "156155")
		dp.Attributes().InsertString("address", "http://another_url")
		dp.SetIntVal(1238)
		dp.SetTimestamp(1608124699.186 * 1e9)
	}

	return metric
}

func exampleDoubleSumMetric() metricPair {
	return buildExampleDoubleSumMetric(true)
}

func buildExampleDoubleSumMetric(fillData bool) metricPair {
	metric := metricPair{
		attributes: pdata.NewAttributeMap(),
		metric:     pdata.NewMetric(),
	}

	metric.metric.SetDataType(pdata.MetricDataTypeSum)
	metric.metric.SetName("sum_metric_double_test")

	metric.attributes.InsertString("foo", "bar")

	if fillData {
		dp := metric.metric.Sum().DataPoints().AppendEmpty()
		dp.Attributes().InsertString("pod_name", "lorem")
		dp.Attributes().InsertString("namespace", "default")
		dp.SetDoubleVal(45.6)
		dp.SetTimestamp(1618124444.169 * 1e9)

		dp = metric.metric.Sum().DataPoints().AppendEmpty()
		dp.Attributes().InsertString("pod_name", "opsum")
		dp.Attributes().InsertString("namespace", "kube-config")
		dp.SetDoubleVal(1238.1)
		dp.SetTimestamp(1608424699.186 * 1e9)
	}

	return metric
}

func exampleSummaryMetric() metricPair {
	return buildExampleSummaryMetric(true)
}

func buildExampleSummaryMetric(fillData bool) metricPair {
	metric := metricPair{
		attributes: pdata.NewAttributeMap(),
		metric:     pdata.NewMetric(),
	}

	metric.metric.SetDataType(pdata.MetricDataTypeSummary)
	metric.metric.SetName("summary_metric_double_test")

	metric.attributes.InsertString("foo", "bar")

	if fillData {
		dp := metric.metric.Summary().DataPoints().AppendEmpty()
		dp.Attributes().InsertString("pod_name", "dolor")
		dp.Attributes().InsertString("namespace", "sumologic")
		dp.SetSum(45.6)
		dp.SetCount(3)
		dp.SetTimestamp(1618124444.169 * 1e9)

		quantile := dp.QuantileValues().AppendEmpty()
		quantile.SetQuantile(0.6)
		quantile.SetValue(0.7)

		quantile = dp.QuantileValues().AppendEmpty()
		quantile.SetQuantile(2.6)
		quantile.SetValue(4)

		dp = metric.metric.Summary().DataPoints().AppendEmpty()
		dp.Attributes().InsertString("pod_name", "sit")
		dp.Attributes().InsertString("namespace", "main")
		dp.SetSum(1238.1)
		dp.SetCount(7)
		dp.SetTimestamp(1608424699.186 * 1e9)
	}

	return metric
}

func exampleHistogramMetric() metricPair {
	return buildExampleHistogramMetric(true)
}

func buildExampleHistogramMetric(fillData bool) metricPair {
	metric := metricPair{
		attributes: pdata.NewAttributeMap(),
		metric:     pdata.NewMetric(),
	}

	metric.metric.SetDataType(pdata.MetricDataTypeHistogram)
	metric.metric.SetName("histogram_metric_double_test")

	metric.attributes.InsertString("bar", "foo")

	if fillData {
		dp := metric.metric.Histogram().DataPoints().AppendEmpty()
		dp.Attributes().InsertString("container", "dolor")
		dp.Attributes().InsertString("branch", "sumologic")
		dp.SetBucketCounts([]uint64{0, 12, 7, 5, 8, 13})
		dp.SetExplicitBounds([]float64{0.1, 0.2, 0.5, 0.8, 1})
		dp.SetTimestamp(1618124444.169 * 1e9)
		dp.SetSum(45.6)
		dp.SetCount(7)

		dp = metric.metric.Histogram().DataPoints().AppendEmpty()
		dp.Attributes().InsertString("container", "sit")
		dp.Attributes().InsertString("branch", "main")
		dp.SetBucketCounts([]uint64{0, 10, 1, 1, 4, 6})
		dp.SetExplicitBounds([]float64{0.1, 0.2, 0.5, 0.8, 1})
		dp.SetTimestamp(1608424699.186 * 1e9)
		dp.SetSum(54.1)
		dp.SetCount(98)
	} else {
		dp := metric.metric.Histogram().DataPoints().AppendEmpty()
		dp.SetCount(0)
	}

	return metric
}

func metricPairToMetrics(mp []metricPair) pdata.Metrics {
	metrics := pdata.NewMetrics()
	metrics.ResourceMetrics().EnsureCapacity(len(mp))
	for _, record := range mp {
		rms := metrics.ResourceMetrics().AppendEmpty()
		record.attributes.CopyTo(rms.Resource().Attributes())
		// TODO: Change metricPair to have an init metric func.
		record.metric.CopyTo(rms.InstrumentationLibraryMetrics().AppendEmpty().Metrics().AppendEmpty())
	}

	return metrics
}

func fieldsFromMap(s map[string]string) fields {
	attrMap := pdata.NewAttributeMap()
	for k, v := range s {
		attrMap.InsertString(k, v)
	}
	return newFields(attrMap)
}

func exampleTrace() pdata.Traces {
	td := pdata.NewTraces()
	rs := td.ResourceSpans().AppendEmpty()
	rs.Resource().Attributes().UpsertString("hostname", "testHost")
	rs.Resource().Attributes().UpsertString("_sourceHost", "source_host")
	rs.Resource().Attributes().UpsertString("_sourceName", "source_name")
	rs.Resource().Attributes().UpsertString("_sourceCategory", "source_category")
	span := rs.InstrumentationLibrarySpans().AppendEmpty().Spans().AppendEmpty()
	span.SetTraceID(pdata.NewTraceID([16]byte{0x5B, 0x8E, 0xFF, 0xF7, 0x98, 0x3, 0x81, 0x3, 0xD2, 0x69, 0xB6, 0x33, 0x81, 0x3F, 0xC6, 0xC}))
	span.SetSpanID(pdata.NewSpanID([8]byte{0xEE, 0xE1, 0x9B, 0x7E, 0xC3, 0xC1, 0xB1, 0x73}))
	span.SetName("testSpan")
	span.SetStartTimestamp(1544712660000000000)
	span.SetEndTimestamp(1544712661000000000)
	span.Attributes().UpsertInt("attr1", 55)
	return td
}
