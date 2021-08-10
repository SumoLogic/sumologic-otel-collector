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

package telegrafreceiver

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/model/pdata"
	"go.uber.org/zap"
)

func TestConverter(t *testing.T) {
	tim := time.Now()

	tests := []struct {
		name          string
		metricsFn     func() telegraf.Metric
		separateField bool
		expectedErr   bool
		expectedFn    func() pdata.MetricSlice
	}{
		{
			name:          "gauge_int_with_one_field",
			separateField: false,
			metricsFn: func() telegraf.Metric {
				fields := map[string]interface{}{
					"available": uint64(39097651200),
				}

				return metric.New("mem", nil, fields, tim, telegraf.Gauge)
			},
			expectedFn: func() pdata.MetricSlice {
				metrics := pdata.NewMetricSlice()

				newIntGauge(39097651200,
					WithName("mem_available"),
					WithTime(tim),
				).CopyTo(metrics.AppendEmpty())
				return metrics
			},
		},
		{
			name:          "gauge_int_separate_field_with_one_field",
			separateField: true,
			metricsFn: func() telegraf.Metric {
				fields := map[string]interface{}{
					"available": uint64(39097651200),
				}

				return metric.New("mem", nil, fields, tim, telegraf.Gauge)
			},
			expectedFn: func() pdata.MetricSlice {
				metrics := pdata.NewMetricSlice()
				newIntGauge(39097651200,
					WithName("mem"),
					WithField("available"),
					WithTime(tim),
				).CopyTo(metrics.AppendEmpty())
				return metrics
			},
		},
		// Don't expect telegraf tags to be added to data point labels.
		// We only add the telegraf tags to resource level attributes.
		// {
		// 	name:          "gauge_int_with_one_field_and_tag",
		// 	separateField: false,
		// 	metricsFn: func() telegraf.Metric {
		// 		fields := map[string]interface{}{
		// 			"available": uint64(39097651200),
		// 		}
		// 		tags := map[string]string{
		// 			"host": "localhost",
		// 		}
		// 		return metric.New("mem", tags, fields, tim, telegraf.Gauge)
		// 	},
		// 	expectedFn: func() pdata.MetricSlice {
		// 		metrics := pdata.NewMetricSlice()
		// 		metrics.Append(
		// 			newIntGauge(39097651200,
		// 				WithName("mem_available"),
		// 				WithTime(tim),
		// 				WithTag(&telegraf.Tag{Key: "host", Value: "localhost"}),
		// 			),
		// 		)
		// 		return metrics
		// 	},
		// },
		{
			name:          "gauge_double_with_one_field",
			separateField: false,
			metricsFn: func() telegraf.Metric {
				fields := map[string]interface{}{
					"available_percent": 54.505050,
				}

				return metric.New("mem", nil, fields, tim, telegraf.Gauge)
			},
			expectedFn: func() pdata.MetricSlice {
				metrics := pdata.NewMetricSlice()
				newDoubleGauge(54.505050,
					WithName("mem_available_percent"),
					WithTime(tim),
				).CopyTo(metrics.AppendEmpty())
				return metrics
			},
		},
		{
			name:          "gauge_double_separate_field_with_one_field",
			separateField: true,
			metricsFn: func() telegraf.Metric {
				fields := map[string]interface{}{
					"available_percent": 54.505050,
				}

				return metric.New("mem", nil, fields, tim, telegraf.Gauge)
			},
			expectedFn: func() pdata.MetricSlice {
				metrics := pdata.NewMetricSlice()
				newDoubleGauge(54.505050,
					WithName("mem"),
					WithField("available_percent"),
					WithTime(tim),
				).CopyTo(metrics.AppendEmpty())
				return metrics
			},
		},
		// Don't expect telegraf tags to be added to data point labels.
		// We only add the telegraf tags to resource level attributes.
		// {
		// 	name:          "gauge_double_with_one_field_and_one_tag",
		// 	separateField: false,
		// 	metricsFn: func() telegraf.Metric {
		// 		fields := map[string]interface{}{
		// 			"available_percent": 54.505050,
		// 		}
		// 		tags := map[string]string{
		// 			"host": "localhost",
		// 		}
		// 		return metric.New("mem", tags, fields, tim, telegraf.Gauge)
		// 	},
		// 	expectedFn: func() pdata.MetricSlice {
		// 		metrics := pdata.NewMetricSlice()
		// 		metrics.Append(
		// 			newDoubleGauge(54.505050,
		// 				WithName("mem_available_percent"),
		// 				WithTime(tim),
		// 				WithTag(&telegraf.Tag{Key: "host", Value: "localhost"}),
		// 			),
		// 		)
		// 		return metrics
		// 	},
		// },
		{
			name:          "gauge_double_with_one_field_and_one_tag_doesnt_get_copied_to_record_attributes",
			separateField: false,
			metricsFn: func() telegraf.Metric {
				fields := map[string]interface{}{
					"available_percent": 54.505050,
				}
				tags := map[string]string{
					"host": "localhost",
				}
				return metric.New("mem", tags, fields, tim, telegraf.Gauge)
			},
			expectedFn: func() pdata.MetricSlice {
				metrics := pdata.NewMetricSlice()
				newDoubleGauge(54.505050,
					WithName("mem_available_percent"),
					WithTime(tim),
				).CopyTo(metrics.AppendEmpty())
				return metrics
			},
		},
		{
			name:          "gauge_int_with_multiple_fields",
			separateField: false,
			metricsFn: func() telegraf.Metric {
				fields := map[string]interface{}{
					"available":    uint64(39097651200),
					"free":         uint64(24322170880),
					"total":        uint64(68719476736),
					"used":         uint64(29621825536),
					"used_percent": 43.10542941093445,
				}

				return metric.New("mem", nil, fields, tim, telegraf.Gauge)
			},
			expectedFn: func() pdata.MetricSlice {
				metrics := pdata.NewMetricSlice()
				newIntGauge(39097651200,
					WithName("mem_available"),
					WithTime(tim),
				).CopyTo(metrics.AppendEmpty())
				newIntGauge(24322170880,
					WithName("mem_free"),
					WithTime(tim),
				).CopyTo(metrics.AppendEmpty())
				newIntGauge(68719476736,
					WithName("mem_total"),
					WithTime(tim),
				).CopyTo(metrics.AppendEmpty())
				newIntGauge(29621825536,
					WithName("mem_used"),
					WithTime(tim),
				).CopyTo(metrics.AppendEmpty())
				newDoubleGauge(43.10542941093445,
					WithName("mem_used_percent"),
					WithTime(tim),
				).CopyTo(metrics.AppendEmpty())
				return metrics
			},
		},
		{
			name:          "gauge_int_separate_field_with_multiple_fields",
			separateField: true,
			metricsFn: func() telegraf.Metric {
				fields := map[string]interface{}{
					"available": uint64(39097651200),
					"free":      uint64(24322170880),
				}

				return metric.New("mem", nil, fields, tim, telegraf.Gauge)
			},
			expectedFn: func() pdata.MetricSlice {
				metrics := pdata.NewMetricSlice()
				newIntGauge(39097651200,
					WithName("mem"),
					WithField("available"),
					WithTime(tim),
				).CopyTo(metrics.AppendEmpty())
				newIntGauge(24322170880,
					WithName("mem"),
					WithField("free"),
					WithTime(tim),
				).CopyTo(metrics.AppendEmpty())
				return metrics
			},
		},
		{
			name:          "sum_int_with_one_field",
			separateField: false,
			metricsFn: func() telegraf.Metric {
				fields := map[string]interface{}{
					"available": uint64(39097651200),
				}

				return metric.New("mem", nil, fields, tim, telegraf.Counter)
			},
			expectedFn: func() pdata.MetricSlice {
				metrics := pdata.NewMetricSlice()
				newIntSum(39097651200,
					WithName("mem_available"),
					WithTime(tim),
				).CopyTo(metrics.AppendEmpty())
				return metrics
			},
		},
		{
			name:          "sum_int_separate_field_with_one_field",
			separateField: true,
			metricsFn: func() telegraf.Metric {
				fields := map[string]interface{}{
					"available": uint64(39097651200),
				}

				return metric.New("mem", nil, fields, tim, telegraf.Counter)
			},
			expectedFn: func() pdata.MetricSlice {
				metrics := pdata.NewMetricSlice()
				newIntSum(39097651200,
					WithName("mem"),
					WithField("available"),
					WithTime(tim),
				).CopyTo(metrics.AppendEmpty())
				return metrics
			},
		},
		// Don't expect telegraf tags to be added to data point labels.
		// We only add the telegraf tags to resource level attributes.
		// {
		// 	name:          "sum_int_with_one_field_and_one_tag",
		// 	separateField: false,
		// 	metricsFn: func() telegraf.Metric {
		// 		fields := map[string]interface{}{
		// 			"available": uint64(39097651200),
		// 		}
		// 		tags := map[string]string{
		// 			"host": "localhost",
		// 		}
		// 		return metric.New("mem", tags, fields, tim, telegraf.Counter)
		// 	},
		// 	expectedFn: func() pdata.MetricSlice {
		// 		metrics := pdata.NewMetricSlice()
		// 		metrics.Append(
		// 			newIntSum(39097651200,
		// 				WithName("mem_available"),
		// 				WithTime(tim),
		// 				WithTag(&telegraf.Tag{Key: "host", Value: "localhost"}),
		// 			),
		// 		)
		// 		return metrics
		// 	},
		// },
		// Don't expect telegraf tags to be added to data point labels.
		// We only add the telegraf tags to resource level attributes.
		// {
		// 	name:          "sum_int_with_one_field_and_three_tags",
		// 	separateField: false,
		// 	metricsFn: func() telegraf.Metric {
		// 		fields := map[string]interface{}{
		// 			"available": uint64(39097651200),
		// 		}
		// 		tags := map[string]string{
		// 			"host":    "localhost",
		// 			"cluster": "my-cluster",
		// 			"blade":   "blade0",
		// 		}
		// 		return metric.New("mem", tags, fields, tim, telegraf.Counter)
		// 	},
		// 	expectedFn: func() pdata.MetricSlice {
		// 		metrics := pdata.NewMetricSlice()
		// 		metrics.Append(
		// 			newIntSum(39097651200,
		// 				WithName("mem_available"),
		// 				WithTime(tim),
		// 				WithTag(&telegraf.Tag{Key: "blade", Value: "blade0"}),
		// 				WithTag(&telegraf.Tag{Key: "host", Value: "localhost"}),
		// 				WithTag(&telegraf.Tag{Key: "cluster", Value: "my-cluster"}),
		// 			),
		// 		)
		// 		return metrics
		// 	},
		// },
		{
			name:          "sum_double_with_one_field",
			separateField: false,
			metricsFn: func() telegraf.Metric {
				fields := map[string]interface{}{
					"available": float64(39097651200.123),
				}

				return metric.New("mem", nil, fields, tim, telegraf.Counter)
			},
			expectedFn: func() pdata.MetricSlice {
				metrics := pdata.NewMetricSlice()
				newDoubleSum(39097651200.123,
					WithName("mem_available"),
					WithTime(tim),
				).CopyTo(metrics.AppendEmpty())
				return metrics
			},
		},
		{
			name:          "sum_double_separate_field_with_one_field",
			separateField: true,
			metricsFn: func() telegraf.Metric {
				fields := map[string]interface{}{
					"available": float64(39097651200.123),
				}

				return metric.New("mem", nil, fields, tim, telegraf.Counter)
			},
			expectedFn: func() pdata.MetricSlice {
				metrics := pdata.NewMetricSlice()
				newDoubleSum(39097651200.123,
					WithName("mem"),
					WithField("available"),
					WithTime(tim),
				).CopyTo(metrics.AppendEmpty())
				return metrics
			},
		},
		// Don't expect telegraf tags to be added to data point labels.
		// We only add the telegraf tags to resource level attributes.
		// {
		// 	name:          "sum_double_separate_field_with_one_field_and_one_tag",
		// 	separateField: true,
		// 	metricsFn: func() telegraf.Metric {
		// 		fields := map[string]interface{}{
		// 			"available": float64(39097651200.123),
		// 		}
		// 		tags := map[string]string{
		// 			"host": "localhost",
		// 		}
		// 		return metric.New("mem", tags, fields, tim, telegraf.Counter)
		// 	},
		// 	expectedFn: func() pdata.MetricSlice {
		// 		metrics := pdata.NewMetricSlice()
		// 		metrics.Append(
		// 			newDoubleSum(39097651200.123,
		// 				WithName("mem"),
		// 				WithField("available"),
		// 				WithTime(tim),
		// 				WithTag(&telegraf.Tag{Key: "host", Value: "localhost"}),
		// 			),
		// 		)
		// 		return metrics
		// 	},
		// },
		{
			name:          "sum_int_with_multiple_fields",
			separateField: false,
			metricsFn: func() telegraf.Metric {
				fields := map[string]interface{}{
					"available": uint64(39097651200),
					"free":      uint64(24322170880),
					"total":     uint64(68719476736),
					"used":      uint64(29621825536),
				}

				return metric.New("mem", nil, fields, tim, telegraf.Counter)
			},
			expectedFn: func() pdata.MetricSlice {
				metrics := pdata.NewMetricSlice()
				newIntSum(39097651200,
					WithName("mem_available"),
					WithTime(tim),
				).CopyTo(metrics.AppendEmpty())
				newIntSum(24322170880,
					WithName("mem_free"),
					WithTime(tim),
				).CopyTo(metrics.AppendEmpty())
				newIntSum(68719476736,
					WithName("mem_total"),
					WithTime(tim),
				).CopyTo(metrics.AppendEmpty())
				newIntSum(29621825536,
					WithName("mem_used"),
					WithTime(tim),
				).CopyTo(metrics.AppendEmpty())
				return metrics
			},
		},
		{
			name:          "sum_int_separate_field_with_multiple_fields",
			separateField: true,
			metricsFn: func() telegraf.Metric {
				fields := map[string]interface{}{
					"available": uint64(39097651200),
					"free":      uint64(24322170880),
				}

				return metric.New("mem", nil, fields, tim, telegraf.Counter)
			},
			expectedFn: func() pdata.MetricSlice {
				metrics := pdata.NewMetricSlice()
				newIntSum(39097651200,
					WithName("mem"),
					WithField("available"),
					WithTime(tim),
				).CopyTo(metrics.AppendEmpty())
				newIntSum(24322170880,
					WithName("mem"),
					WithField("free"),
					WithTime(tim),
				).CopyTo(metrics.AppendEmpty())
				return metrics
			},
		},
		{
			name:          "untyped_int_with_one_field",
			separateField: false,
			metricsFn: func() telegraf.Metric {
				fields := map[string]interface{}{
					"available": uint64(39097651200),
				}

				return metric.New("mem", nil, fields, tim, telegraf.Untyped)
			},
			expectedFn: func() pdata.MetricSlice {
				metrics := pdata.NewMetricSlice()
				newIntGauge(39097651200,
					WithName("mem_available"),
					WithTime(tim),
				).CopyTo(metrics.AppendEmpty())
				return metrics
			},
		},
		{
			name:          "untyped_double_with_one_field",
			separateField: false,
			metricsFn: func() telegraf.Metric {
				fields := map[string]interface{}{
					"used_percent": 43.10542941093445,
				}

				return metric.New("mem", nil, fields, tim, telegraf.Untyped)
			},
			expectedFn: func() pdata.MetricSlice {
				metrics := pdata.NewMetricSlice()
				newDoubleGauge(43.10542941093445,
					WithName("mem_used_percent"),
					WithTime(tim),
				).CopyTo(metrics.AppendEmpty())
				return metrics
			},
		},
		{
			name:          "untyped_bool_with_one_field_false",
			separateField: false,
			metricsFn: func() telegraf.Metric {
				fields := map[string]interface{}{
					"throttling_supported": false,
				}

				return metric.New("cpu", nil, fields, tim, telegraf.Untyped)
			},
			expectedFn: func() pdata.MetricSlice {
				metrics := pdata.NewMetricSlice()
				newIntGauge(0,
					WithName("cpu_throttling_supported"),
					WithTime(tim),
				).CopyTo(metrics.AppendEmpty())
				return metrics
			},
		},
		{
			name:          "untyped_bool_with_one_field_true",
			separateField: false,
			metricsFn: func() telegraf.Metric {
				fields := map[string]interface{}{
					"throttling_supported": true,
				}

				return metric.New("cpu", nil, fields, tim, telegraf.Untyped)
			},
			expectedFn: func() pdata.MetricSlice {
				metrics := pdata.NewMetricSlice()
				newIntGauge(1,
					WithName("cpu_throttling_supported"),
					WithTime(tim),
				).CopyTo(metrics.AppendEmpty())
				return metrics
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.metricsFn()

			mc := newConverter(tt.separateField, zap.NewNop())
			out, err := mc.Convert(m)

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)

				resourceMetrics := out.ResourceMetrics().At(0)
				assertResourceAttributes(t, m.TagList(), resourceMetrics.Resource())

				actual := resourceMetrics.InstrumentationLibraryMetrics().At(0).Metrics()

				expected := tt.expectedFn()
				require.Equal(t, expected.Len(), actual.Len())
				if tt.separateField {
					pdataMetricSlicesWithFieldsAreEqual(t, expected, actual)
				} else {
					pdataMetricSlicesAreEqual(t, expected, actual)
				}
			}
		})
	}
}

type DataPointWithLabelsMap interface {
	LabelsMap() pdata.StringMap
}

func assertResourceAttributes(t *testing.T, tags []*telegraf.Tag, resource pdata.Resource) {
	resource.Attributes().Range(func(k string, v pdata.AttributeValue) bool {
		var found bool
		for _, tag := range tags {
			if k != tag.Key {
				continue
			}
			if assert.Equal(t, pdata.AttributeValueTypeString, v.Type()) {
				if assert.Equal(t, tag.Value, v.StringVal()) {
					found = true
				}
				break
			}
		}
		assert.True(t, found, "attribute: %v not found", k)
		return true
	})
}

func pdataMetricSlicesAreEqual(t *testing.T, expected, actual pdata.MetricSlice) {
	for i := 0; i < expected.Len(); i++ {
		em := expected.At(i)
		eName := em.Name()

		var pass bool
		for j := 0; j < actual.Len(); j++ {
			am := actual.At(j)
			aName := am.Name()
			if eName == aName {
				// Note: cannot compare with assert/require because
				// each DataPoints() LabelsMap() is a map without
				// order.
				// assert.EqualValues(t, em, am)
				assert.Equal(t, em.Description(), am.Description())
				assert.Equal(t, em.Unit(), am.Unit())
				if assert.Equal(t, em.DataType(), am.DataType()) {
					assertEqualDataPointsWithLabels(t, em, am)
				}
				pass = true
				break
			}
		}
		assert.True(t, pass, "%q metric not found", eName)
	}
}

// assertEqualDataPointsWithLabels checks that provided metrics have the same
// data points with the same set of labels.
func assertEqualDataPointsWithLabels(t *testing.T, em pdata.Metric, am pdata.Metric) {
	switch em.DataType() {
	case pdata.MetricDataTypeIntGauge:
		edps := em.IntGauge().DataPoints()
		adps := am.IntGauge().DataPoints()
		assert.Equal(t, edps.Len(), adps.Len())
		for i := 0; i < edps.Len(); i++ {
			expected := edps.At(i)
			actual := adps.At(i)
			assert.Equal(t, expected.Value(), actual.Value())
			assertEqualDataPoints(t, am.Name(), expected, actual)
		}
	case pdata.MetricDataTypeGauge:
		edps := em.Gauge().DataPoints()
		adps := am.Gauge().DataPoints()
		assert.Equal(t, edps.Len(), adps.Len())
		for i := 0; i < edps.Len(); i++ {
			expected := edps.At(i)
			actual := adps.At(i)
			assert.Equal(t, expected.Value(), actual.Value())
			assertEqualDataPoints(t, am.Name(), expected, actual)
		}
	case pdata.MetricDataTypeIntSum:
		edps := em.IntSum().DataPoints()
		adps := am.IntSum().DataPoints()
		assert.Equal(t, edps.Len(), adps.Len())
		for i := 0; i < edps.Len(); i++ {
			expected := edps.At(i)
			actual := adps.At(i)
			assert.Equal(t, expected.Value(), actual.Value())
			assertEqualDataPoints(t, am.Name(), expected, actual)
		}
	case pdata.MetricDataTypeSum:
		edps := em.Sum().DataPoints()
		adps := am.Sum().DataPoints()
		assert.Equal(t, edps.Len(), adps.Len())
		for i := 0; i < edps.Len(); i++ {
			expected := edps.At(i)
			actual := adps.At(i)
			assert.Equal(t, expected.Value(), actual.Value())
			assertEqualDataPoints(t, am.Name(), expected, actual)
		}
	}
}

type DataPoint interface {
	Timestamp() pdata.Timestamp
	StartTimestamp() pdata.Timestamp
	LabelsMap() pdata.StringMap
}

func assertEqualDataPoints(t *testing.T, metricName string, expected, actual DataPoint) {
	// NOTE: cannot compare values due to different return types of Value()
	// func for different metric types.
	// assert.Equal(t, edp.Value(), adp.Value())
	assert.Equal(t, expected.Timestamp(), actual.Timestamp())
	assert.Equal(t, expected.StartTimestamp(), actual.StartTimestamp())

	// Expect that there are no labels added to data points because we don't
	// copy over the telegraf tags to record level attributes.
	assert.Equal(t, expected.LabelsMap().Len(), actual.LabelsMap().Len(),
		"The amount of actual data point labels on %q metric is not as expected",
		metricName,
	)

	// Don't expect telegraf tags to be added to data point labels.
	// We only add the telegraf tags to resource level attributes.
	// assert.Equal(t, edp.LabelsMap().Sort(), adp.LabelsMap().Sort())
}

func pdataMetricSlicesWithFieldsAreEqual(t *testing.T, expected, actual pdata.MetricSlice) {
	for i := 0; i < expected.Len(); i++ {
		em := expected.At(i)
		eName := em.Name()
		eFields := getFieldsFromMetric(em)

		// assert the fields
		for ef := range eFields {
			am, ok := metricSliceContainsMetricWithField(actual, eName, ef)
			if assert.True(t, ok, "pdata.MetricSlice doesn't contain %s", eName) {

				t.Logf("expected field name %s", ef)
				adp, ok := fieldFromMetric(am, ef)
				if assert.True(t, ok, "%q field not present for %q metric", ef, am.Name()) {
					edp, _ := fieldFromMetric(em, ef)
					assert.EqualValues(t, edp, adp)
				}
			}
		}
	}
}

// metricSliceContainsMetricWithField searches through metrics in pdata.MetricSlice
// and return the pdata.Metric that contains the requested field and a flag
// whether such a metric was found.
func metricSliceContainsMetricWithField(ms pdata.MetricSlice, name string, field string) (pdata.Metric, bool) {
	for i := 0; i < ms.Len(); i++ {
		m := ms.At(i)
		if m.Name() == name {
			switch m.DataType() {
			case pdata.MetricDataTypeIntGauge:
				mg := m.IntGauge()
				dps := mg.DataPoints()
				for i := 0; i < dps.Len(); i++ {
					dp := dps.At(i)
					l, ok := dp.LabelsMap().Get("field")
					if !ok {
						continue
					}

					if l == field {
						return m, true
					}
				}

			case pdata.MetricDataTypeGauge:
				mg := m.Gauge()
				dps := mg.DataPoints()
				for i := 0; i < dps.Len(); i++ {
					dp := dps.At(i)
					l, ok := dp.LabelsMap().Get("field")
					if !ok {
						continue
					}

					if l == field {
						return m, true
					}
				}
			}
		}
	}

	return pdata.Metric{}, false
}

// getFieldsFromMetric returns a map of fields in a metric gathered from all
// data points' label maps.
func getFieldsFromMetric(m pdata.Metric) map[string]struct{} {
	switch m.DataType() {
	case pdata.MetricDataTypeIntGauge:
		ret := make(map[string]struct{})
		for i := 0; i < m.IntGauge().DataPoints().Len(); i++ {
			dp := m.IntGauge().DataPoints().At(i)
			l, ok := dp.LabelsMap().Get("field")
			if !ok {
				continue
			}
			ret[l] = struct{}{}
		}
		return ret

	case pdata.MetricDataTypeGauge:
		ret := make(map[string]struct{})
		for i := 0; i < m.Gauge().DataPoints().Len(); i++ {
			dp := m.Gauge().DataPoints().At(i)
			l, ok := dp.LabelsMap().Get("field")
			if !ok {
				continue
			}
			ret[l] = struct{}{}
		}
		return ret

	default:
		return nil
	}
}

// fieldFromMetric searches through pdata.Metric's data points to find
// a particular field.
func fieldFromMetric(m pdata.Metric, field string) (DataPointWithLabelsMap, bool) {
	switch m.DataType() {
	case pdata.MetricDataTypeIntGauge:
		dps := m.IntGauge().DataPoints()
		for i := 0; i < dps.Len(); i++ {
			dp := dps.At(i)
			l, ok := dp.LabelsMap().Get("field")
			if !ok {
				continue
			}

			if l == field {
				return dp, true
			}
		}

	case pdata.MetricDataTypeGauge:
		dps := m.Gauge().DataPoints()
		for i := 0; i < dps.Len(); i++ {
			dp := dps.At(i)
			l, ok := dp.LabelsMap().Get("field")
			if !ok {
				continue
			}

			if l == field {
				return dp, true
			}
		}

	default:
		return nil, false
	}

	return nil, false
}
