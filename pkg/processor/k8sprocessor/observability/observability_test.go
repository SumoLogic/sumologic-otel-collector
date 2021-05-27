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

// NOTE:
// This test can only be run with -count 1 because of static
// metricproducer.GlobalManager() used in metricexport.NewReader().
func TestMetrics(t *testing.T) {
	type testCase struct {
		name       string
		recordFunc func()
	}
	tests := []testCase{
		{
			"otelsvc/k8s/pod_added",
			RecordPodAdded,
		},
		{
			"otelsvc/k8s/pod_deleted",
			RecordPodDeleted,
		},
		{
			"otelsvc/k8s/pod_updated",
			RecordPodUpdated,
		},
		{
			"otelsvc/k8s/ip_lookup_miss",
			RecordIPLookupMiss,
		},
	}

	var (
		fail   = make(chan struct{})
		chData = make(chan []*metricdata.Metric)
	)

	go func() {
		reader := metricexport.NewReader()
		e := newExporter()
		ch := e.ReturnAfter(len(tests))

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
	}()

	for _, tt := range tests {
		tt.recordFunc()
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
	}
}
