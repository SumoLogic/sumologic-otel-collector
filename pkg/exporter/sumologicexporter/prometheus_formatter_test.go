// Copyright 2020, OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package sumologicexporter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/model/pdata"
)

func TestSanitizeKey(t *testing.T) {
	f, err := newPrometheusFormatter()
	require.NoError(t, err)

	key := "&^*123-abc-ABC!?"
	expected := "___123_abc_ABC__"
	assert.Equal(t, expected, f.sanitizeKey(key))
}

func TestSanitizeValue(t *testing.T) {
	f, err := newPrometheusFormatter()
	require.NoError(t, err)

	value := `&^*123-abc-ABC!?"\\n`
	expected := `&^*123-abc-ABC!?\"\\\n`
	assert.Equal(t, expected, f.sanitizeValue(value))
}

func TestTags2StringNoLabels(t *testing.T) {
	f, err := newPrometheusFormatter()
	require.NoError(t, err)

	mp := exampleIntMetric()
	mp.attributes.Clear()
	assert.Equal(t, prometheusTags(""), f.tags2String(mp.attributes, pdata.NewStringMap()))
}

func TestTags2String(t *testing.T) {
	f, err := newPrometheusFormatter()
	require.NoError(t, err)

	mp := exampleIntMetric()
	assert.Equal(
		t,
		prometheusTags(`{test="test_value",test2="second_value"}`),
		f.tags2String(mp.attributes, pdata.NewStringMap()),
	)
}

func TestTags2StringNoAttributes(t *testing.T) {
	f, err := newPrometheusFormatter()
	require.NoError(t, err)

	mp := exampleIntMetric()
	mp.attributes.Clear()
	assert.Equal(t, prometheusTags(""), f.tags2String(pdata.NewAttributeMap(), pdata.NewStringMap()))
}

func TestPrometheusMetricDataTypeIntGauge(t *testing.T) {
	f, err := newPrometheusFormatter()
	require.NoError(t, err)
	metric := exampleIntGaugeMetric()

	result := f.metric2String(metric)
	expected := `gauge_metric_name{foo="bar",remote_name="156920",url="http://example_url"} 124 1608124661166
gauge_metric_name{foo="bar",remote_name="156955",url="http://another_url"} 245 1608124662166`
	assert.Equal(t, expected, result)
}

func TestPrometheusMetricDataTypeDoubleGauge(t *testing.T) {
	f, err := newPrometheusFormatter()
	require.NoError(t, err)
	metric := exampleDoubleGaugeMetric()

	result := f.metric2String(metric)
	expected := `gauge_metric_name_double_test{foo="bar",local_name="156720",endpoint="http://example_url"} 33.4 1608124661169
gauge_metric_name_double_test{foo="bar",local_name="156155",endpoint="http://another_url"} 56.8 1608124662186`
	assert.Equal(t, expected, result)
}

func TestPrometheusMetricDataTypeIntSum(t *testing.T) {
	f, err := newPrometheusFormatter()
	require.NoError(t, err)
	metric := exampleIntSumMetric()

	result := f.metric2String(metric)
	expected := `sum_metric_int_test{foo="bar",name="156720",address="http://example_url"} 45 1608124444169
sum_metric_int_test{foo="bar",name="156155",address="http://another_url"} 1238 1608124699186`
	assert.Equal(t, expected, result)
}

func TestPrometheusMetricDataTypeDoubleSum(t *testing.T) {
	f, err := newPrometheusFormatter()
	require.NoError(t, err)
	metric := exampleDoubleSumMetric()

	result := f.metric2String(metric)
	expected := `sum_metric_double_test{foo="bar",pod_name="lorem",namespace="default"} 45.6 1618124444169
sum_metric_double_test{foo="bar",pod_name="opsum",namespace="kube-config"} 1238.1 1608424699186`
	assert.Equal(t, expected, result)
}

func TestPrometheusMetricDataTypeSummary(t *testing.T) {
	f, err := newPrometheusFormatter()
	require.NoError(t, err)
	metric := exampleSummaryMetric()

	result := f.metric2String(metric)
	expected := `summary_metric_double_test{foo="bar",quantile="0.6",pod_name="dolor",namespace="sumologic"} 0.7 1618124444169
summary_metric_double_test{foo="bar",quantile="2.6",pod_name="dolor",namespace="sumologic"} 4 1618124444169
summary_metric_double_test_sum{foo="bar",pod_name="dolor",namespace="sumologic"} 45.6 1618124444169
summary_metric_double_test_count{foo="bar",pod_name="dolor",namespace="sumologic"} 3 1618124444169
summary_metric_double_test_sum{foo="bar",pod_name="sit",namespace="main"} 1238.1 1608424699186
summary_metric_double_test_count{foo="bar",pod_name="sit",namespace="main"} 7 1608424699186`
	assert.Equal(t, expected, result)
}

func TestPrometheusMetricDataTypeIntHistogram(t *testing.T) {
	f, err := newPrometheusFormatter()
	require.NoError(t, err)
	metric := exampleIntHistogramMetric()

	result := f.metric2String(metric)
	expected := `histogram_metric_int_test{foo="bar",le="0.1",pod_name="dolor",namespace="sumologic"} 0 1618124444169
histogram_metric_int_test{foo="bar",le="0.2",pod_name="dolor",namespace="sumologic"} 12 1618124444169
histogram_metric_int_test{foo="bar",le="0.5",pod_name="dolor",namespace="sumologic"} 19 1618124444169
histogram_metric_int_test{foo="bar",le="0.8",pod_name="dolor",namespace="sumologic"} 24 1618124444169
histogram_metric_int_test{foo="bar",le="1",pod_name="dolor",namespace="sumologic"} 32 1618124444169
histogram_metric_int_test{foo="bar",le="+Inf",pod_name="dolor",namespace="sumologic"} 45 1618124444169
histogram_metric_int_test_sum{foo="bar",pod_name="dolor",namespace="sumologic"} 45 1618124444169
histogram_metric_int_test_count{foo="bar",pod_name="dolor",namespace="sumologic"} 3 1618124444169
histogram_metric_int_test{foo="bar",le="0.1",pod_name="sit",namespace="main"} 0 1608424699186
histogram_metric_int_test{foo="bar",le="0.2",pod_name="sit",namespace="main"} 10 1608424699186
histogram_metric_int_test{foo="bar",le="0.5",pod_name="sit",namespace="main"} 11 1608424699186
histogram_metric_int_test{foo="bar",le="0.8",pod_name="sit",namespace="main"} 12 1608424699186
histogram_metric_int_test{foo="bar",le="1",pod_name="sit",namespace="main"} 16 1608424699186
histogram_metric_int_test{foo="bar",le="+Inf",pod_name="sit",namespace="main"} 22 1608424699186
histogram_metric_int_test_sum{foo="bar",pod_name="sit",namespace="main"} 54 1608424699186
histogram_metric_int_test_count{foo="bar",pod_name="sit",namespace="main"} 5 1608424699186`
	assert.Equal(t, expected, result)
}

func TestPrometheusMetricDataTypeHistogram(t *testing.T) {
	f, err := newPrometheusFormatter()
	require.NoError(t, err)
	metric := exampleHistogramMetric()

	result := f.metric2String(metric)
	expected := `histogram_metric_double_test{bar="foo",le="0.1",container="dolor",branch="sumologic"} 0 1618124444169
histogram_metric_double_test{bar="foo",le="0.2",container="dolor",branch="sumologic"} 12 1618124444169
histogram_metric_double_test{bar="foo",le="0.5",container="dolor",branch="sumologic"} 19 1618124444169
histogram_metric_double_test{bar="foo",le="0.8",container="dolor",branch="sumologic"} 24 1618124444169
histogram_metric_double_test{bar="foo",le="1",container="dolor",branch="sumologic"} 32 1618124444169
histogram_metric_double_test{bar="foo",le="+Inf",container="dolor",branch="sumologic"} 45 1618124444169
histogram_metric_double_test_sum{bar="foo",container="dolor",branch="sumologic"} 45.6 1618124444169
histogram_metric_double_test_count{bar="foo",container="dolor",branch="sumologic"} 7 1618124444169
histogram_metric_double_test{bar="foo",le="0.1",container="sit",branch="main"} 0 1608424699186
histogram_metric_double_test{bar="foo",le="0.2",container="sit",branch="main"} 10 1608424699186
histogram_metric_double_test{bar="foo",le="0.5",container="sit",branch="main"} 11 1608424699186
histogram_metric_double_test{bar="foo",le="0.8",container="sit",branch="main"} 12 1608424699186
histogram_metric_double_test{bar="foo",le="1",container="sit",branch="main"} 16 1608424699186
histogram_metric_double_test{bar="foo",le="+Inf",container="sit",branch="main"} 22 1608424699186
histogram_metric_double_test_sum{bar="foo",container="sit",branch="main"} 54.1 1608424699186
histogram_metric_double_test_count{bar="foo",container="sit",branch="main"} 98 1608424699186`
	assert.Equal(t, expected, result)
}
