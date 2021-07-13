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
	"go.opentelemetry.io/collector/consumer/pdata"
)

// MetricOpt is an option func that takes in a pdata.Metric and manipulates it.
type MetricOpt func(m pdata.Metric)

// WithName returns a MetricOpt which will set the returned metric name.
func WithName(name string) MetricOpt {
	return func(m pdata.Metric) {
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
	return func(m pdata.Metric) {
		opts := options{
			timeopt: dataPointTimeOpt(t),
		}

		switch m.DataType() {
		case pdata.MetricDataTypeIntGauge:
			handleIntDataPoints(
				m.IntGauge().DataPoints(),
				opts,
			)
		case pdata.MetricDataTypeDoubleGauge:
			handleDoubleDataPoints(
				m.DoubleGauge().DataPoints(),
				opts,
			)

		case pdata.MetricDataTypeIntSum:
			handleIntDataPoints(
				m.IntSum().DataPoints(),
				opts,
			)
		case pdata.MetricDataTypeDoubleSum:
			handleDoubleDataPoints(
				m.DoubleSum().DataPoints(),
				opts,
			)
		}
	}
}

// WithField returns a MetricOpt which will set the returned metric's
// field tag to the specified one.
func WithField(field string) MetricOpt {
	f := WithTag(&telegraf.Tag{Key: fieldLabel, Value: field})
	return func(m pdata.Metric) {
		f(m)
	}
}

type StringMapOpt func(pdata.StringMap)
type TimeOpt func() time.Time

type options struct {
	stringMapOpts []StringMapOpt
	timeopt       TimeOpt
}

func handleIntDataPoints(dps pdata.IntDataPointSlice, opts options) {
	for i := 0; i < dps.Len(); i++ {
		dp := dps.At(i)
		for _, opt := range opts.stringMapOpts {
			opt(dp.LabelsMap())
		}

		if opts.timeopt != nil {
			dp.SetTimestamp(pdata.Timestamp(opts.timeopt().UnixNano()))
		}
	}
}
func handleDoubleDataPoints(dps pdata.DoubleDataPointSlice, opts options) {
	for i := 0; i < dps.Len(); i++ {
		dp := dps.At(i)
		for _, opt := range opts.stringMapOpts {
			opt(dp.LabelsMap())
		}

		if opts.timeopt != nil {
			dp.SetTimestamp(pdata.Timestamp(opts.timeopt().UnixNano()))
		}
	}
}

func insertTagToPdataStringMapOpt(tag *telegraf.Tag) func(pdata.StringMap) {
	return func(sm pdata.StringMap) {
		sm.Insert(tag.Key, tag.Value)
	}
}

// WithTag returns a MetricOpt which will insert a specified telegraf tag into
// all underlying data points' label maps.
func WithTag(tag *telegraf.Tag) MetricOpt {
	return func(m pdata.Metric) {
		opts := options{
			stringMapOpts: []StringMapOpt{
				insertTagToPdataStringMapOpt(tag),
			},
		}

		switch m.DataType() {
		case pdata.MetricDataTypeIntGauge:
			handleIntDataPoints(
				m.IntGauge().DataPoints(),
				opts,
			)
		case pdata.MetricDataTypeDoubleGauge:
			handleDoubleDataPoints(
				m.DoubleGauge().DataPoints(),
				opts,
			)

		case pdata.MetricDataTypeIntSum:
			handleIntDataPoints(
				m.IntSum().DataPoints(),
				opts,
			)
		case pdata.MetricDataTypeDoubleSum:
			handleDoubleDataPoints(
				m.DoubleSum().DataPoints(),
				opts,
			)
		}
	}
}

func insertTagsToPdataStringMapOpt(tags []*telegraf.Tag) func(pdata.StringMap) {
	return func(sm pdata.StringMap) {
		for _, tag := range tags {
			sm.Insert(tag.Key, tag.Value)
		}
	}
}

// WithTags returns a MetricOpt which will insert a list of telegraf tags into
// all underlying data points' label maps.
func WithTags(tags []*telegraf.Tag) MetricOpt {
	return func(m pdata.Metric) {
		opts := options{
			stringMapOpts: []StringMapOpt{
				insertTagsToPdataStringMapOpt(tags),
			},
		}

		switch m.DataType() {
		case pdata.MetricDataTypeIntGauge:
			handleIntDataPoints(
				m.IntGauge().DataPoints(),
				opts,
			)
		case pdata.MetricDataTypeDoubleGauge:
			handleDoubleDataPoints(
				m.DoubleGauge().DataPoints(),
				opts,
			)

		case pdata.MetricDataTypeIntSum:
			handleIntDataPoints(
				m.IntSum().DataPoints(),
				opts,
			)
		case pdata.MetricDataTypeDoubleSum:
			handleDoubleDataPoints(
				m.DoubleSum().DataPoints(),
				opts,
			)
		}
	}
}
