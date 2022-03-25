// Copyright 2022 Sumo Logic, Inc.
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
	"fmt"
	"time"

	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
)

// TODO: re-think if processor should register it's own telemetry views or if some other
// mechanism should be used by the collector to discover views from all components

func init() {
	err := view.Register(
		viewRequestsSent,
		viewRequestsDuration,
	)
	if err != nil {
		fmt.Printf("Failed to register sumologic exporter's views: %v\n", err)
	}
}

var (
	mRequestsSent     = stats.Int64("sumologic/requests/sent", "Number of requests per status_code", "1")
	mRequestsDuration = stats.Int64("sumologic/requests/duration", "Duration of HTTP requests (in milliseconds)", "0")
	statusKey, _      = tag.NewKey("status_code")
)

var viewRequestsSent = &view.View{
	Name:        mRequestsSent.Name(),
	Description: mRequestsSent.Description(),
	Measure:     mRequestsSent,
	TagKeys:     []tag.Key{statusKey},
	Aggregation: view.Count(),
}

var viewRequestsDuration = &view.View{
	Name:        mRequestsDuration.Name(),
	Description: mRequestsDuration.Description(),
	Measure:     mRequestsDuration,
	TagKeys:     []tag.Key{statusKey},
	Aggregation: view.Sum(),
}

// RecordPodUpdated increments the metric that records pod update events received.
func RecordRequestsSent(status_code int) {
	stats.RecordWithTags(
		context.Background(),
		[]tag.Mutator{tag.Insert(statusKey, fmt.Sprintf("%d", status_code))},
		mRequestsSent.M(int64(1)),
	)
}

// RecordPodAdded increments the metric that records pod add events receiver.
func RecordRequestsDuration(duration time.Duration, status_code int) {
	stats.RecordWithTags(
		context.Background(),
		[]tag.Mutator{tag.Insert(statusKey, fmt.Sprintf("%d", status_code))},
		mRequestsDuration.M(duration.Milliseconds()),
	)
}
