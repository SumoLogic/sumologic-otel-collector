
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

package observability

import (
	"context"
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opencensus.io/metric/metricdata"
	"go.opencensus.io/metric/metricexport"
)

type exporter struct {
	pipe chan *metricdata.Metric
}

func newExporter() *exporter {
	return &exporter{make(chan *metricdata.Metric)}
}

func (e *exporter) ReturnAfter(after int) chan []*metricdata.Metric {
	ch := make(chan []*metricdata.Metric)
	go func() {
		received := []*metricdata.Metric{}
		for m := range e.pipe {
			received = append(received, m)
			if len(received) >= after {
				break
			}
		}
		ch <- received
	}()
	return ch
}

func (e *exporter) ExportMetrics(ctx context.Context, data []*metricdata.Metric) error {
	for _, m := range data {
		e.pipe <- m
	}
	return nil
}

func metricReader(chData chan []*metricdata.Metric, fail chan struct{}, count int) {
	reader := metricexport.NewReader()
	e := newExporter()
	ch := e.ReturnAfter(count)

	// Add a manual retry mechanism in case there's a hiccup reading the
	// metrics from producers in ReadAndExport(): we can wait for the metrics
	// to come instead of failing because they didn't come right away.
	for i := 0; i < 10; i++ {
		go reader.ReadAndExport(e)

		select {
		case <-time.After(500 * time.Millisecond):

		case data := <-ch:
			chData <- data
			return
		}
	}

	fail <- struct{}{}
}

// NOTE:
// This test can only be run with -count 1 because of static
// metricproducer.GlobalManager() used in metricexport.NewReader().
func TestMetrics(t *testing.T) {
	const (
		statusCode = 200
		endpoint   = "some/uri"
	)
	type testCase struct {
		name       string
		bytes      int64
		records    int64
		recordFunc string
		duration   time.Duration
	}
	tests := []testCase{
		{
			name:       "sumologic/requests/sent",
			recordFunc: "sent",
		},
		{
			name:       "sumologic/requests/duration",
			recordFunc: "duration",
			duration:   time.Millisecond,
		},
		{
			name:       "sumologic/requests/bytes",
			recordFunc: "bytes",
			bytes:      1,
		},
		{
			name:       "sumologic/requests/records",
			recordFunc: "records",
			records:    1,
		},
	}

	var (
		fail   = make(chan struct{})
		chData = make(chan []*metricdata.Metric)
	)

	go metricReader(chData, fail, len(tests))

	for _, tt := range tests {
		switch tt.recordFunc {
		case "sent":
			RecordRequestsSent(statusCode, endpoint)
		case "duration":
			RecordRequestsDuration(tt.duration, statusCode, endpoint)
		case "bytes":
			RecordRequestsBytes(tt.bytes, statusCode, endpoint)
		case "records":
			RecordRequestsRecords(tt.records, statusCode, endpoint)
		}
	}

	var data []*metricdata.Metric
	select {
	case <-fail:
		t.Fatalf("timedout waiting for metrics to arrive")
	case data = <-chData:
	}

	sort.Slice(tests, func(i, j int) bool {
		return tests[i].name < tests[j].name
	})

	sort.Slice(data, func(i, j int) bool {
		return data[i].Descriptor.Name < data[j].Descriptor.Name
	})

	for i, tt := range tests {
		require.Len(t, data, len(tests))
		d := data[i]
		assert.Equal(t, d.Descriptor.Name, tt.name)
		require.Len(t, d.TimeSeries, 1)
		require.Len(t, d.TimeSeries[0].Points, 1)
		assert.Equal(t, d.TimeSeries[0].Points[0].Value, int64(1))

		require.Len(t, d.TimeSeries[0].LabelValues, 2)

		require.True(t, d.TimeSeries[0].LabelValues[0].Present)
		require.True(t, d.TimeSeries[0].LabelValues[1].Present)

		assert.Equal(t, d.TimeSeries[0].LabelValues[0].Value, "some/uri")
		assert.Equal(t, d.TimeSeries[0].LabelValues[1].Value, "200")
	}
}
