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

	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
)

func init() {
	// To get rid of this nolint directive, we want to get rid of OpenCensus dependency and report metrics in Otel native way
	// See https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/29867
	//nolint:errcheck
	_ = view.Register(
		viewPodsUpdated,
		viewPodsAdded,
		viewPodsDeleted,
		viewOtherUpdated,
		viewOtherAdded,
		viewOtherDeleted,
		viewIPLookupMiss,
		viewPodTableSize,
	)
}

var (
	mPodsUpdated  = stats.Int64("otelsvc/k8s/pod_updated", "Number of pod update events received", "1")
	mPodsAdded    = stats.Int64("otelsvc/k8s/pod_added", "Number of pod add events received", "1")
	mPodsDeleted  = stats.Int64("otelsvc/k8s/pod_deleted", "Number of pod delete events received", "1")
	mPodTableSize = stats.Int64("otelsvc/k8s/pod_table_size", "Size of table containing pod info", "1")

	mOtherUpdated = stats.Int64("otelsvc/k8s/other_updated", "Number of other update events received", "1")
	mOtherAdded   = stats.Int64("otelsvc/k8s/other_added", "Number of other add events received", "1")
	mOtherDeleted = stats.Int64("otelsvc/k8s/other_deleted", "Number of other delete events received", "1")

	mIPLookupMiss = stats.Int64("otelsvc/k8s/ip_lookup_miss", "Number of times pod by IP lookup failed.", "1")

	resourceKind, _ = tag.NewKey("kind") // nolint:errcheck
)

var viewPodsUpdated = &view.View{
	Name:        mPodsUpdated.Name(),
	Description: mPodsUpdated.Description(),
	Measure:     mPodsUpdated,
	Aggregation: view.Sum(),
}

var viewPodsAdded = &view.View{
	Name:        mPodsAdded.Name(),
	Description: mPodsAdded.Description(),
	Measure:     mPodsAdded,
	Aggregation: view.Sum(),
}

var viewPodsDeleted = &view.View{
	Name:        mPodsDeleted.Name(),
	Description: mPodsDeleted.Description(),
	Measure:     mPodsDeleted,
	Aggregation: view.Sum(),
}

var viewOtherUpdated = &view.View{
	Name:        mOtherUpdated.Name(),
	Description: mOtherUpdated.Description(),
	Measure:     mOtherUpdated,
	TagKeys:     []tag.Key{resourceKind},
	Aggregation: view.Sum(),
}

var viewOtherAdded = &view.View{
	Name:        mOtherAdded.Name(),
	Description: mOtherAdded.Description(),
	Measure:     mOtherAdded,
	TagKeys:     []tag.Key{resourceKind},
	Aggregation: view.Sum(),
}

var viewOtherDeleted = &view.View{
	Name:        mOtherDeleted.Name(),
	Description: mOtherDeleted.Description(),
	Measure:     mOtherDeleted,
	TagKeys:     []tag.Key{resourceKind},
	Aggregation: view.Sum(),
}

var viewIPLookupMiss = &view.View{
	Name:        mIPLookupMiss.Name(),
	Description: mIPLookupMiss.Description(),
	Measure:     mIPLookupMiss,
	Aggregation: view.Sum(),
}
var viewPodTableSize = &view.View{
	Name:        mPodTableSize.Name(),
	Description: mPodTableSize.Description(),
	Measure:     mPodTableSize,
	Aggregation: view.LastValue(),
}

// RecordPodUpdated increments the metric that records pod update events received.
func RecordPodUpdated() {
	stats.Record(context.Background(), mPodsUpdated.M(int64(1)))
}

// RecordPodAdded increments the metric that records pod add events receiver.
func RecordPodAdded() {
	stats.Record(context.Background(), mPodsAdded.M(int64(1)))
}

// RecordPodDeleted increments the metric that records pod events deleted.
func RecordPodDeleted() {
	stats.Record(context.Background(), mPodsDeleted.M(int64(1)))
}

// RecordOtherUpdated increments the metric that records other update events received.
func RecordOtherUpdated(kind string) {
	stats.RecordWithTags( // nolint:errcheck
		context.Background(),
		[]tag.Mutator{
			tag.Insert(resourceKind, kind),
		},
		mOtherUpdated.M(int64(1)))
}

// RecordOtherAdded increments the metric that records other add events receiver.
func RecordOtherAdded(kind string) {
	stats.RecordWithTags( // nolint:errcheck
		context.Background(),
		[]tag.Mutator{
			tag.Insert(resourceKind, kind),
		},
		mOtherAdded.M(int64(1)))
}

// RecordOtherDeleted increments the metric that records other events deleted.
func RecordOtherDeleted(kind string) {
	stats.RecordWithTags( // nolint:errcheck
		context.Background(),
		[]tag.Mutator{
			tag.Insert(resourceKind, kind),
		},
		mOtherAdded.M(int64(1)))
}

// RecordIPLookupMiss increments the metric that records Pod lookup by IP misses.
func RecordIPLookupMiss() {
	stats.Record(context.Background(), mIPLookupMiss.M(int64(1)))
}

// RecordPodTableSize store size of pod table field in WatchClient
func RecordPodTableSize(podTableSize int64) {
	stats.Record(context.Background(), mPodTableSize.M(podTableSize))
}
