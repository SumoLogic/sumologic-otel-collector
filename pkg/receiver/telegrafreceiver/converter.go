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
	"fmt"

	"github.com/influxdata/telegraf"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.uber.org/zap"
)

const (
	fieldLabel = "field"
)

type MetricConverter interface {
	Convert(telegraf.Metric) (pmetric.Metrics, error)
}

type metricConverter struct {
	separateField bool
	logger        *zap.Logger
}

func newConverter(separateField bool, logger *zap.Logger) MetricConverter {
	return metricConverter{
		separateField: separateField,
		logger:        logger,
	}
}

// Convert converts telegraf.Metric to pmetric.Metrics.
func (mc metricConverter) Convert(m telegraf.Metric) (pmetric.Metrics, error) {
	ms := pmetric.NewMetrics()
	rms := ms.ResourceMetrics()
	rm := rms.AppendEmpty()

	// Attach tags as resource attributes.
	rAttributes := rm.Resource().Attributes()
	for _, t := range m.TagList() {
		rAttributes.PutStr(t.Key, t.Value)
	}

	sm := rm.ScopeMetrics().AppendEmpty()

	scope := sm.Scope()
	scope.SetName(typeStr)
	scope.SetVersion(versionStr)

	tim := m.Time()

	metrics := sm.Metrics()

	opts := []MetricOpt{
		// Note: don't copy telegraf tags to record level attributes.
		//
		// This way we cannot use e.g. metricstransformprocessor. because
		// as of now it only allows to manipulate record level attributes
		// but we won't break existing workflows like k8sprocessor
		// relying on resource level attributes.
		//
		// WithTags(m.TagList()),

		WithTime(tim),
	}

	switch t := m.Type(); t {
	case telegraf.Gauge:
		metrics.EnsureCapacity(len(m.FieldList()))
		for _, f := range m.FieldList() {
			pm, err := mc.convertToGauge(m.Name(), f, opts...)
			if err != nil {
				mc.logger.Debug(
					"unsupported data type when handling telegraf.Gauge",
					zap.String("key", f.Key),
					zap.Any("value", f.Value),
					zap.Error(err),
				)
				continue
			}

			pm.CopyTo(metrics.AppendEmpty())
		}

	case telegraf.Untyped:
		metrics.EnsureCapacity(len(m.FieldList()))
		for _, f := range m.FieldList() {
			pm, err := mc.convertToGauge(m.Name(), f, opts...)
			if err != nil {
				mc.logger.Debug(
					"unsupported data type when handling telegraf.Untyped",
					zap.String("key", f.Key),
					zap.Any("value", f.Value),
					zap.Error(err),
				)
				continue
			}

			pm.CopyTo(metrics.AppendEmpty())
		}

	case telegraf.Counter:
		metrics.EnsureCapacity(len(m.FieldList()))
		for _, f := range m.FieldList() {
			pm, err := mc.convertToSum(m.Name(), f, opts...)
			if err != nil {
				mc.logger.Debug(
					"unsupported data type when handling telegraf.Gauge",
					zap.String("key", f.Key),
					zap.Any("value", f.Value),
					zap.Error(err),
				)
				continue
			}

			pm.CopyTo(metrics.AppendEmpty())
		}

	case telegraf.Summary:
		return pmetric.Metrics{}, fmt.Errorf("unsupported metric type: telegraf.Summary")
	case telegraf.Histogram:
		return pmetric.Metrics{}, fmt.Errorf("unsupported metric type: telegraf.Histogram")

	default:
		return pmetric.Metrics{}, fmt.Errorf("unknown metric type: %T", t)
	}

	return ms, nil
}

// convertToGauge returns a pmetric.Metric gauge converted from telegraf metric,
// based on provided metric name, field and metric options which are passed
// to metric constructors to manipulate the created metric in a functional manner.
func (mc metricConverter) convertToGauge(name string, f *telegraf.Field, opts ...MetricOpt) (pmetric.Metric, error) {
	if mc.separateField {
		opts = append(opts, WithField(f.Key))
	}
	opts = append(opts, WithName(mc.createMetricName(name, f.Key)))

	var pm pmetric.Metric
	switch v := f.Value.(type) {
	case float64:
		pm = newDoubleGauge(v, opts...)

	case int64:
		pm = newIntGauge(v, opts...)
	case uint64:
		pm = newIntGauge(int64(v), opts...)

	case bool:
		var vv int64 = 0
		if v {
			vv = 1
		}
		pm = newIntGauge(vv, opts...)

	default:
		return pm, fmt.Errorf("unsupported underlying type: %T", v)
	}

	return pm, nil
}

// convertToGauge returns a pmetric.Metric sum converted from telegraf metric,
// based on provided metric name, field and metric options which are passed
// to metric constructors to manipulate the created metric in a functional manner.
func (mc metricConverter) convertToSum(name string, f *telegraf.Field, opts ...MetricOpt) (pmetric.Metric, error) {
	if mc.separateField {
		opts = append(opts, WithField(f.Key))
	}
	opts = append(opts, WithName(mc.createMetricName(name, f.Key)))

	var pm pmetric.Metric
	switch v := f.Value.(type) {
	case float64:
		pm = newDoubleSum(v, opts...)

	case int64:
		pm = newIntSum(v, opts...)
	case uint64:
		pm = newIntSum(int64(v), opts...)

	case bool:
		var vv int64 = 0
		if v {
			vv = 1
		}
		pm = newIntSum(vv, opts...)

	default:
		return pm, fmt.Errorf("unsupported underlying type: %T", v)
	}

	return pm, nil
}

// createMetricName returns a metric name using provided metric name and key/field.
// If metric converter was configured to create metrics with separate fields then
// don't use the provided field and just use the metric name. Field name will be
// added as data point label, with "field" key name.
func (mc metricConverter) createMetricName(name string, field string) string {
	if mc.separateField {
		return name
	} else {
		return name + "_" + field
	}
}

func newDoubleSum(
	value float64,
	opts ...MetricOpt,
) pmetric.Metric {
	pm := pmetric.NewMetric()
	pm.SetEmptySum()
	// "[...] OTLP Sum is either translated into a Timeseries Counter, when
	// the sum is monotonic, or a Gauge when the sum is not monotonic."
	// https://github.com/open-telemetry/opentelemetry-specification/blob/7fc28733/specification/metrics/datamodel.md#opentelemetry-protocol-data-model
	ds := pm.Sum()
	ds.SetAggregationTemporality(pmetric.AggregationTemporalityCumulative)
	ds.SetIsMonotonic(true)
	dps := ds.DataPoints()
	dp := dps.AppendEmpty()
	dp.SetDoubleValue(value)

	for _, opt := range opts {
		opt(pm)
	}
	return pm
}

func newIntSum(
	value int64,
	opts ...MetricOpt,
) pmetric.Metric {
	pm := pmetric.NewMetric()
	pm.SetEmptySum()
	// "[...] OTLP Sum is either translated into a Timeseries Counter, when
	// the sum is monotonic, or a Gauge when the sum is not monotonic."
	// https://github.com/open-telemetry/opentelemetry-specification/blob/7fc28733/specification/metrics/datamodel.md#opentelemetry-protocol-data-model
	ds := pm.Sum()
	ds.SetAggregationTemporality(pmetric.AggregationTemporalityCumulative)
	ds.SetIsMonotonic(true)
	dps := ds.DataPoints()
	dp := dps.AppendEmpty()
	dp.SetIntValue(value)

	for _, opt := range opts {
		opt(pm)
	}
	return pm
}

func newDoubleGauge(
	value float64,
	opts ...MetricOpt,
) pmetric.Metric {
	pm := pmetric.NewMetric()
	pm.SetEmptyGauge()
	dps := pm.Gauge().DataPoints()
	dp := dps.AppendEmpty()
	dp.SetDoubleValue(value)

	for _, opt := range opts {
		opt(pm)
	}
	return pm
}

func newIntGauge(
	value int64,
	opts ...MetricOpt,
) pmetric.Metric {
	pm := pmetric.NewMetric()
	pm.SetEmptyGauge()
	dps := pm.Gauge().DataPoints()
	dp := dps.AppendEmpty()
	dp.SetIntValue(value)

	for _, opt := range opts {
		opt(pm)
	}
	return pm
}
