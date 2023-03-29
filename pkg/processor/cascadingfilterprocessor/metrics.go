// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cascadingfilterprocessor

import (
	"context"

	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
	"go.opentelemetry.io/collector/config/configtelemetry"
)

// Variables related to metrics specific to Cascading Filter.
var (
	statusSampled              = "Sampled"
	statusNotSampled           = "NotSampled"
	statusExceededKey          = "RateExceeded"
	statusSecondChance         = "SecondChance"
	statusSecondChanceSampled  = "SecondChanceSampled"
	statusSecondChanceExceeded = "SecondChanceRateExceeded"
	statusDropped              = "Dropped"

	tagPolicyKey, _                  = tag.NewKey("policy")                    // nolint:errcheck
	tagCascadingFilterDecisionKey, _ = tag.NewKey("cascading_filter_decision") // nolint:errcheck
	tagPolicyDecisionKey, _          = tag.NewKey("policy_decision")           // nolint:errcheck
	tagProcessorKey, _               = tag.NewKey("processor")                 // nolint:errcheck

	statDecisionLatencyMicroSec  = stats.Int64("policy_decision_latency", "Latency (in microseconds) of a given filtering policy", "µs")
	statOverallDecisionLatencyus = stats.Int64("cascading_filtering_batch_processing_latency", "Latency (in microseconds) of each run of the cascading filter timer", "µs")

	statPolicyEvaluationErrorCount = stats.Int64("cascading_policy_evaluation_error", "Count of cascading policy evaluation errors", stats.UnitDimensionless)

	statCascadingFilterDecision = stats.Int64("count_final_decision", "Count of traces that were filtered or not", stats.UnitDimensionless)
	statPolicyDecision          = stats.Int64("count_policy_decision", "Count of provisional (policy) decisions if traces were filtered or not", stats.UnitDimensionless)

	statCascadingFilterDecidedSpans = stats.Int64("count_decided_spans", "Count of spans that were handled on decision time", stats.UnitDimensionless)
	statCascadingFilterLateSpans    = stats.Int64("count_late_spans", "Count of spans that were handled in batches after the one where decision was made", stats.UnitDimensionless)

	statDroppedTooEarlyCount    = stats.Int64("casdading_trace_dropped_too_early", "Count of traces that needed to be dropped the configured wait time", stats.UnitDimensionless)
	statNewTraceIDReceivedCount = stats.Int64("cascading_new_trace_id_received", "Counts the arrival of new traces", stats.UnitDimensionless)
	statTracesOnMemoryGauge     = stats.Int64("cascading_traces_on_memory", "Tracks the number of traces current on memory", stats.UnitDimensionless)
)

func recordProvisionalDecisionMade(ctx context.Context, instanceName string, decisionKey string) {
	//nolint:errcheck
	_ = stats.RecordWithTags(
		ctx,
		[]tag.Mutator{
			tag.Insert(tagProcessorKey, instanceName),
			tag.Insert(tagPolicyDecisionKey, decisionKey),
		},
		statPolicyDecision.M(int64(1)))
}

func recordCascadingFilterDecision(ctx context.Context, instanceName string, decisionKey string) {
	//nolint:errcheck
	_ = stats.RecordWithTags(
		ctx,
		[]tag.Mutator{
			tag.Insert(tagProcessorKey, instanceName),
			tag.Insert(tagPolicyDecisionKey, decisionKey),
		},
		statCascadingFilterDecision.M(int64(1)))
}

func recordSpanLateDecision(ctx context.Context, instanceName string, decision string, count int) {
	//nolint:errcheck
	_ = stats.RecordWithTags(
		ctx,
		[]tag.Mutator{tag.Insert(tagProcessorKey, instanceName), tag.Insert(tagCascadingFilterDecisionKey, decision)},
		statCascadingFilterLateSpans.M(int64(count)),
	)
}

func recordSpanEarlyDecision(ctx context.Context, instanceName string, decision string, count int) {
	//nolint:errcheck
	_ = stats.RecordWithTags(
		ctx,
		[]tag.Mutator{tag.Insert(tagProcessorKey, instanceName), tag.Insert(tagCascadingFilterDecisionKey, decision)},
		statCascadingFilterDecidedSpans.M(int64(count)),
	)
}

// CascadingFilterMetricViews return the metrics views according to given telemetry level.
func CascadingFilterMetricViews(level configtelemetry.Level) []*view.View {
	if level == configtelemetry.LevelNone {
		return nil
	}

	latencyDistributionAggregation := view.Distribution(1, 2, 5, 10, 25, 50, 75, 100, 150, 200, 300, 400, 500, 750, 1000, 2000, 3000, 4000, 5000, 10000, 20000, 30000, 50000)

	overallDecisionLatencyView := &view.View{
		Name:        statOverallDecisionLatencyus.Name(),
		Measure:     statOverallDecisionLatencyus,
		Description: statOverallDecisionLatencyus.Description(),
		TagKeys:     []tag.Key{tagProcessorKey},
		Aggregation: latencyDistributionAggregation,
	}

	countPolicyEvaluationErrorView := &view.View{
		Name:        statPolicyEvaluationErrorCount.Name(),
		Measure:     statPolicyEvaluationErrorCount,
		Description: statPolicyEvaluationErrorCount.Description(),
		TagKeys:     []tag.Key{tagProcessorKey},
		Aggregation: view.Sum(),
	}

	countFinalDecisionView := &view.View{
		Name:        statCascadingFilterDecision.Name(),
		Measure:     statCascadingFilterDecision,
		Description: statCascadingFilterDecision.Description(),
		TagKeys:     []tag.Key{tagProcessorKey, tagPolicyKey, tagCascadingFilterDecisionKey},
		Aggregation: view.Sum(),
	}

	countPolicyDecisionsView := &view.View{
		Name:        statPolicyDecision.Name(),
		Measure:     statPolicyDecision,
		Description: statPolicyDecision.Description(),
		TagKeys:     []tag.Key{tagProcessorKey, tagPolicyKey, tagPolicyDecisionKey},
		Aggregation: view.Sum(),
	}

	policyLatencyView := &view.View{
		Name:        statDecisionLatencyMicroSec.Name(),
		Measure:     statDecisionLatencyMicroSec,
		Description: statDecisionLatencyMicroSec.Description(),
		TagKeys:     []tag.Key{tagProcessorKey, tagPolicyKey},
		Aggregation: view.Sum(),
	}

	countTraceDroppedTooEarlyView := &view.View{
		Name:        statDroppedTooEarlyCount.Name(),
		Measure:     statDroppedTooEarlyCount,
		Description: statDroppedTooEarlyCount.Description(),
		TagKeys:     []tag.Key{tagProcessorKey},
		Aggregation: view.Sum(),
	}
	countTraceIDArrivalView := &view.View{
		Name:        statNewTraceIDReceivedCount.Name(),
		Measure:     statNewTraceIDReceivedCount,
		Description: statNewTraceIDReceivedCount.Description(),
		TagKeys:     []tag.Key{tagProcessorKey},
		Aggregation: view.Sum(),
	}
	trackTracesOnMemorylView := &view.View{
		Name:        statTracesOnMemoryGauge.Name(),
		Measure:     statTracesOnMemoryGauge,
		Description: statTracesOnMemoryGauge.Description(),
		TagKeys:     []tag.Key{tagProcessorKey},
		Aggregation: view.LastValue(),
	}

	countEarlySpans := &view.View{
		Name:        statCascadingFilterDecidedSpans.Name(),
		Measure:     statCascadingFilterDecidedSpans,
		Description: statCascadingFilterDecidedSpans.Description(),
		TagKeys:     []tag.Key{tagProcessorKey, tagCascadingFilterDecisionKey},
		Aggregation: view.Sum(),
	}

	countLateSpans := &view.View{
		Name:        statCascadingFilterLateSpans.Name(),
		Measure:     statCascadingFilterLateSpans,
		Description: statCascadingFilterLateSpans.Description(),
		TagKeys:     []tag.Key{tagProcessorKey, tagCascadingFilterDecisionKey},
		Aggregation: view.Sum(),
	}

	legacyViews := []*view.View{
		overallDecisionLatencyView,

		countPolicyDecisionsView,
		policyLatencyView,
		countFinalDecisionView,

		countEarlySpans,
		countLateSpans,

		countPolicyEvaluationErrorView,
		countTraceDroppedTooEarlyView,
		countTraceIDArrivalView,
		trackTracesOnMemorylView,
	}

	// return obsreport.ProcessorMetricViews(typeStr, legacyViews)
	return legacyViews
}
