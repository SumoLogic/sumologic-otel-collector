// Copyright 2021, OpenTelemetry Authors
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
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
)

func TestCarbon2TagString(t *testing.T) {
	metric, attributes := exampleIntMetric()
	data := carbon2TagString(metric, attributes)
	assert.Equal(t, "test=test_value test2=second_value metric=test.metric.data unit=bytes", data)

	metric, attributes = exampleIntGaugeMetric()
	data = carbon2TagString(metric, attributes)
	assert.Equal(t, "foo=bar metric=gauge_metric_name", data)

	metric, attributes = exampleDoubleSumMetric()
	data = carbon2TagString(metric, attributes)
	assert.Equal(t, "foo=bar metric=sum_metric_double_test", data)

	metric, attributes = exampleDoubleGaugeMetric()
	data = carbon2TagString(metric, attributes)
	assert.Equal(t, "foo=bar metric=gauge_metric_name_double_test", data)
}

func TestCarbon2InvalidCharacters(t *testing.T) {
	metric := pmetric.NewMetric()

	attributes := pcommon.NewMap()
	attributes.InsertString("= \n\r", "= \n\r")
	metric.SetName("= \n\r")

	data := carbon2TagString(metric, attributes)
	assert.Equal(t, ":___=:___ metric=:___", data)
}

func TestCarbonMetricDataTypeIntGauge(t *testing.T) {
	metric, attributes := exampleIntGaugeMetric()

	result := carbon2Metric2String(metric, attributes)
	expected := `foo=bar metric=gauge_metric_name  124 1608124661
foo=bar metric=gauge_metric_name  245 1608124662`
	assert.Equal(t, expected, result)
}

func TestCarbonMetricDataTypeDoubleGauge(t *testing.T) {
	metric, attributes := exampleDoubleGaugeMetric()

	result := carbon2Metric2String(metric, attributes)
	expected := `foo=bar metric=gauge_metric_name_double_test  33.4 1608124661
foo=bar metric=gauge_metric_name_double_test  56.8 1608124662`
	assert.Equal(t, expected, result)
}

func TestCarbonMetricDataTypeIntSum(t *testing.T) {
	metric, attributes := exampleIntSumMetric()

	result := carbon2Metric2String(metric, attributes)
	expected := `foo=bar metric=sum_metric_int_test  45 1608124444
foo=bar metric=sum_metric_int_test  1238 1608124699`
	assert.Equal(t, expected, result)
}

func TestCarbonMetricDataTypeDoubleSum(t *testing.T) {
	metric, attributes := exampleDoubleSumMetric()

	result := carbon2Metric2String(metric, attributes)
	expected := `foo=bar metric=sum_metric_double_test  45.6 1618124444
foo=bar metric=sum_metric_double_test  1238.1 1608424699`
	assert.Equal(t, expected, result)
}

func TestCarbonMetricDataTypeSummary(t *testing.T) {
	metric, attributes := exampleSummaryMetric()

	result := carbon2Metric2String(metric, attributes)
	expected := ``
	assert.Equal(t, expected, result)
}

func TestCarbonMetricDataTypeHistogram(t *testing.T) {
	metric, attributes := exampleHistogramMetric()

	result := carbon2Metric2String(metric, attributes)
	expected := ``
	assert.Equal(t, expected, result)
}

func TestCarbonMetrics(t *testing.T) {
	type testCase struct {
		name       string
		metricFunc func() (pmetric.Metric, pcommon.Map)
		expected   string
	}

	tests := []testCase{
		{
			name: "empty int gauge",
			metricFunc: func() (pmetric.Metric, pcommon.Map) {
				return buildExampleIntGaugeMetric(false)
			},
			expected: "",
		},
		{
			name: "empty double gauge",
			metricFunc: func() (pmetric.Metric, pcommon.Map) {
				return buildExampleDoubleGaugeMetric(false)
			},
			expected: "",
		},
		{
			name: "empty int sum",
			metricFunc: func() (pmetric.Metric, pcommon.Map) {
				return buildExampleIntSumMetric(false)
			},
			expected: "",
		},
		{
			name: "empty double sum",
			metricFunc: func() (pmetric.Metric, pcommon.Map) {
				return buildExampleDoubleSumMetric(false)
			},
			expected: "",
		},
		{
			name: "empty summary",
			metricFunc: func() (pmetric.Metric, pcommon.Map) {
				return buildExampleSummaryMetric(false)
			},
			expected: "",
		},
		{
			name: "empty histogram",
			metricFunc: func() (pmetric.Metric, pcommon.Map) {
				return buildExampleHistogramMetric(false)
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metric, attributes := tt.metricFunc()
			result := carbon2Metric2String(metric, attributes)
			assert.Equal(t, tt.expected, result)
		})
	}
}
