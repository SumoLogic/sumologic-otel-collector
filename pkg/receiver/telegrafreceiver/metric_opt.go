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
	"time"

	"github.com/influxdata/telegraf"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
)

// MetricOpt is an option func that takes in a pmetric.Metric and manipulates it.
type MetricOpt func(m pmetric.Metric)

// WithName returns a MetricOpt which will set the returned metric name.
func WithName(name string) MetricOpt {
	return func(m pmetric.Metric) {
		m.SetName(name)
	}
}

func dataPointTimeOpt(t time.Time) TimeOpt {
	return func() time.Time {
		return t
	}
}

// WithTime returns a MetricOpt which will set the returned metric's timestamp.
func WithTime(t time.Time) MetricOpt {
	return func(m pmetric.Metric) {
		opts := options{
			timeopt: dataPointTimeOpt(t),
		}

		switch m.Type() {
		case pmetric.MetricTypeGauge:
			handleDataPoints(
				m.Gauge().DataPoints(),
				opts,
			)
		case pmetric.MetricTypeSum:
			handleDataPoints(
				m.Sum().DataPoints(),
				opts,
			)
		}
	}
}

// WithField returns a MetricOpt which will set the returned metric's
// field tag to the specified one.
func WithField(field string) MetricOpt {
	f := WithTag(&telegraf.Tag{Key: fieldLabel, Value: field})
	return func(m pmetric.Metric) {
		f(m)
	}
}

type AttributeMapOpt func(attributeMap pcommon.Map)
type TimeOpt func() time.Time

type options struct {
	stringMapOpts []AttributeMapOpt
	timeopt       TimeOpt
}

func handleDataPoints(dps pmetric.NumberDataPointSlice, opts options) {
	for i := 0; i < dps.Len(); i++ {
		dp := dps.At(i)
		for _, opt := range opts.stringMapOpts {
			opt(dp.Attributes())
		}

		if opts.timeopt != nil {
			dp.SetTimestamp(pcommon.Timestamp(opts.timeopt().UnixNano()))
		}
	}
}

func insertTagToPdataStringMapOpt(tag *telegraf.Tag) func(attributeMap pcommon.Map) {
	return func(sm pcommon.Map) {
		sm.PutStr(tag.Key, tag.Value)
	}
}

// WithTag returns a MetricOpt which will insert a specified telegraf tag into
// all underlying data points' label maps.
func WithTag(tag *telegraf.Tag) MetricOpt {
	return func(m pmetric.Metric) {
		opts := options{
			stringMapOpts: []AttributeMapOpt{
				insertTagToPdataStringMapOpt(tag),
			},
		}

		switch m.Type() {
		case pmetric.MetricTypeGauge:
			handleDataPoints(
				m.Gauge().DataPoints(),
				opts,
			)

		case pmetric.MetricTypeSum:
			handleDataPoints(
				m.Sum().DataPoints(),
				opts,
			)
		}
	}
}

func insertTagsToPdataStringMapOpt(tags []*telegraf.Tag) func(attributeMap pcommon.Map) {
	return func(sm pcommon.Map) {
		for _, tag := range tags {
			sm.PutStr(tag.Key, tag.Value)
		}
	}
}

// WithTags returns a MetricOpt which will insert a list of telegraf tags into
// all underlying data points' label maps.
func WithTags(tags []*telegraf.Tag) MetricOpt {
	return func(m pmetric.Metric) {
		opts := options{
			stringMapOpts: []AttributeMapOpt{
				insertTagsToPdataStringMapOpt(tags),
			},
		}

		switch m.Type() {
		case pmetric.MetricTypeGauge:
			handleDataPoints(
				m.Gauge().DataPoints(),
				opts,
			)
		case pmetric.MetricTypeSum:
			handleDataPoints(
				m.Sum().DataPoints(),
				opts,
			)
		}
	}
}
