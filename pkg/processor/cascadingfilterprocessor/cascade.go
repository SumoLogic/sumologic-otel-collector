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
	"math"
	"sync/atomic"
	"time"

	"github.com/SumoLogic/sumologic-otel-collector/pkg/processor/cascadingfilterprocessor/idbatcher"
	"github.com/SumoLogic/sumologic-otel-collector/pkg/processor/cascadingfilterprocessor/sampling"

	"go.opencensus.io/stats"
	"go.opencensus.io/tag"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.uber.org/zap"
)

type cascade struct {
	metrics                            policyMetrics
	cfsp                               *cascadingFilterSpanProcessor
	logger                             *zap.Logger
	totalSpans                         int64
	selectedByProbabilisticFilterSpans int64
}

func newCascade(cfsp *cascadingFilterSpanProcessor) *cascade {
	return &cascade{
		metrics:                            policyMetrics{},
		cfsp:                               cfsp,
		logger:                             cfsp.logger,
		totalSpans:                         0,
		selectedByProbabilisticFilterSpans: 0,
	}
}

func (c *cascade) decideOnBatch(batch *idbatcher.Batch) {
	startTime := time.Now()
	batchLen := len(*batch)

	currSecond := time.Now().Unix()

	// There are really three steps for making a decision:
	// 1. Provisional decision - in which we also check for rate for each policy/filter/evaluator (i.e. if a given
	//    evaluator is above the limit, it will no longer make sampled decision)
	// 2. First pass - in which we check if the selected spans are within the global limit
	// 3. Second pass - in which we add anything that was tagged with "second chance" if it fits within the global limit

	for _, id := range *batch {
		d, ok := c.cfsp.idToTrace.Load(traceKey(id))
		if !ok {
			c.metrics.idNotFoundOnMapCount++
			continue
		}
		trace := d.(*sampling.TraceData)
		trace.DecisionTime = time.Now()

		var provisionalDecision sampling.Decision

		// Dropped traces are not included in probabilistic filtering calculations
		if c.shouldBeDropped(id, trace) {
			provisionalDecision = sampling.Dropped
		} else {
			c.totalSpans += int64(trace.SpanCount)
			// Iterate over evaluators and verify within rate for each of them
			provisionalDecision, _ = c.makeProvisionalDecision(id, trace)
		}

		// Select only traces that fit within the global limit
		c.firstPass(currSecond, trace, provisionalDecision)
	}

	// The second run executes the decisions and makes "SecondChance" decisions in the meantime
	for _, id := range *batch {
		d, ok := c.cfsp.idToTrace.Load(traceKey(id))
		if !ok {
			continue
		}
		trace := d.(*sampling.TraceData)

		// If there's anything left, fill-up with "second chance" traces
		c.secondPass(currSecond, trace)

		c.cfsp.decisionHistory.Add(traceKey(id), decisionHistoryInfo{
			finalDecision:       trace.FinalDecision,
			filterName:          trace.ProvisionalDecisionFilterName,
			probabilisticFilter: trace.SelectedByProbabilisticFilter})

		c.cleanup(trace)

		// Actually, we don'c need to wait since decision history is now used and we can delete the trace pretty much right away
		c.cfsp.dropTrace(traceKey(id))
	}

	//nolint:errcheck
	_ = stats.RecordWithTags(c.cfsp.ctx,
		[]tag.Mutator{tag.Insert(tagProcessorKey, c.cfsp.instanceName)},
		statOverallDecisionLatencyus.M(int64(time.Since(startTime)/time.Microsecond)),
		statDroppedTooEarlyCount.M(c.metrics.idNotFoundOnMapCount),
		statPolicyEvaluationErrorCount.M(c.metrics.evaluateErrorCount),
		statTracesOnMemoryGauge.M(int64(atomic.LoadUint64(&c.cfsp.numTracesOnMap))))

	c.cfsp.logger.Debug("Sampling policy evaluation completed",
		zap.Int("batch.len", batchLen),
		zap.Int64("sampled", c.metrics.decisionSampled),
		zap.Int64("notSampled", c.metrics.decisionNotSampled),
		zap.Int64("droppedPriorToEvaluation", c.metrics.idNotFoundOnMapCount),
		zap.Int64("policyEvaluationErrors", c.metrics.evaluateErrorCount),
	)

}

func (c *cascade) firstPass(currSecond int64, trace *sampling.TraceData, provisionalDecision sampling.Decision) {
	if provisionalDecision == sampling.Sampled {
		trace.FinalDecision = c.cfsp.decisionSpansLimitter.updateRate(currSecond, trace.SpanCount)
		if trace.FinalDecision == sampling.Sampled {
			if trace.SelectedByProbabilisticFilter {
				c.selectedByProbabilisticFilterSpans += int64(trace.SpanCount)
			}

			recordCascadingFilterDecision(c.cfsp.ctx, c.cfsp.instanceName, statusSampled)
		} else {
			recordCascadingFilterDecision(c.cfsp.ctx, c.cfsp.instanceName, statusExceededKey)
		}
	} else if provisionalDecision == sampling.SecondChance {
		trace.FinalDecision = sampling.SecondChance
	} else {
		trace.FinalDecision = provisionalDecision
		recordCascadingFilterDecision(c.cfsp.ctx, c.cfsp.instanceName, statusNotSampled)
	}
}

func (c *cascade) secondPass(currSecond int64, trace *sampling.TraceData) {
	if trace.FinalDecision == sampling.SecondChance {
		trace.FinalDecision = c.cfsp.decisionSpansLimitter.updateRate(currSecond, trace.SpanCount)
		if trace.FinalDecision == sampling.Sampled {
			recordCascadingFilterDecision(c.cfsp.ctx, c.cfsp.instanceName, statusSecondChanceSampled)
		} else {
			recordCascadingFilterDecision(c.cfsp.ctx, c.cfsp.instanceName, statusSecondChanceExceeded)
		}
	}
}

func (c *cascade) cleanup(trace *sampling.TraceData) {
	// Sampled or not, remove the batches
	trace.Lock()
	traceBatches := trace.ReceivedBatches
	trace.ReceivedBatches = nil
	trace.Unlock()

	if trace.FinalDecision == sampling.Sampled {
		c.metrics.decisionSampled++

		// Combine all individual batches into a single batch so
		// consumers may operate on the entire trace
		allSpans := ptrace.NewTraces()
		for j := 0; j < len(traceBatches); j++ {
			batch := traceBatches[j]
			batch.ResourceSpans().MoveAndAppendTo(allSpans.ResourceSpans())
		}

		if trace.SelectedByProbabilisticFilter {
			updateProbabilisticRateTag(allSpans, c.selectedByProbabilisticFilterSpans, c.totalSpans)
		} else if len(c.cfsp.traceAcceptRules) > 0 {
			// Set filtering tag only if there were actually any accept rules set otherwise
			updateFilteringTag(allSpans, trace.ProvisionalDecisionFilterName)
		}

		err := c.cfsp.nextConsumer.ConsumeTraces(c.cfsp.ctx, allSpans)
		if err != nil {
			c.cfsp.logger.Error("Sampling Policy Evaluation error on consuming traces", zap.Error(err))
		}
		recordSpanEarlyDecision(c.cfsp.ctx, c.cfsp.instanceName, statusSampled, allSpans.SpanCount())
	} else {
		recordSpanEarlyDecision(c.cfsp.ctx, c.cfsp.instanceName, statusNotSampled, int(trace.SpanCount))
		c.metrics.decisionNotSampled++
	}
}

func (c *cascade) shouldBeDropped(id pcommon.TraceID, trace *sampling.TraceData) bool {
	for _, dropRule := range c.cfsp.traceRejectRules {
		if dropRule.Evaluator.ShouldDrop(id, trace) {
			//nolint:errcheck
			_ = stats.RecordWithTags(dropRule.ctx, []tag.Mutator{tag.Insert(tagProcessorKey, c.cfsp.instanceName)}, statPolicyDecision.M(int64(1)))
			return true
		}
	}
	return false
}

func (c *cascade) makeProvisionalDecision(id pcommon.TraceID, trace *sampling.TraceData) (sampling.Decision, *TraceAcceptEvaluator) {
	// When no rules are defined, always sample
	if len(c.cfsp.traceAcceptRules) == 0 {
		return sampling.Sampled, nil
	}

	provisionalDecision := sampling.Unspecified

	for i, policy := range c.cfsp.traceAcceptRules {
		policyEvaluateStartTime := time.Now()
		decision := policy.Evaluator.Evaluate(id, trace)
		//nolint:errcheck
		_ = stats.RecordWithTags(
			policy.ctx,
			[]tag.Mutator{tag.Insert(tagProcessorKey, c.cfsp.instanceName)},
			statDecisionLatencyMicroSec.M(int64(time.Since(policyEvaluateStartTime)/time.Microsecond)))

		trace.Decisions[i] = decision

		switch decision {
		case sampling.Sampled:
			// any single policy that decides to sample will cause the decision to be sampled
			// the nextConsumer will get the context from the first matching policy
			provisionalDecision = sampling.Sampled

			if policy.probabilisticFilter {
				trace.SelectedByProbabilisticFilter = true
			} else {
				trace.ProvisionalDecisionFilterName = policy.Name
			}

			recordProvisionalDecisionMade(policy.ctx, c.cfsp.instanceName, statusSampled)

			// No need to continue
			return provisionalDecision, policy
		case sampling.NotSampled:
			if provisionalDecision == sampling.Unspecified {
				provisionalDecision = sampling.NotSampled
			}

			recordProvisionalDecisionMade(policy.ctx, c.cfsp.instanceName, statusNotSampled)
		case sampling.SecondChance:
			if provisionalDecision != sampling.Sampled {
				provisionalDecision = sampling.SecondChance
				trace.ProvisionalDecisionFilterName = policy.Name
			}

			recordProvisionalDecisionMade(policy.ctx, c.cfsp.instanceName, statusSecondChance)
		}
	}

	return provisionalDecision, nil
}

func updateProbabilisticRateTag(traces ptrace.Traces, probabilisticSpans int64, allSpans int64) {
	ratio := float64(probabilisticSpans) / float64(allSpans)

	rs := traces.ResourceSpans()

	for i := 0; i < rs.Len(); i++ {
		ss := rs.At(i).ScopeSpans()
		for j := 0; j < ss.Len(); j++ {
			spans := ss.At(j).Spans()
			for k := 0; k < spans.Len(); k++ {
				attrs := spans.At(k).Attributes()
				av, found := attrs.Get(AttributeSamplingProbability)
				if found && av.Type() == pcommon.ValueTypeDouble && !math.IsNaN(av.Double()) && av.Double() > 0.0 {
					av.SetDouble(av.Double() * ratio)
				} else {
					attrs.PutDouble(AttributeSamplingProbability, ratio)
				}

				attrs.PutStr(AttributeSamplingRule, probabilisticRuleVale)
			}
		}
	}
}

func updateFilteringTag(traces ptrace.Traces, filterName string) {
	rs := traces.ResourceSpans()

	for i := 0; i < rs.Len(); i++ {
		ss := rs.At(i).ScopeSpans()
		for j := 0; j < ss.Len(); j++ {
			spans := ss.At(j).Spans()
			for k := 0; k < spans.Len(); k++ {
				attrs := spans.At(k).Attributes()
				attrs.PutStr(AttributeSamplingRule, filteredRuleValue)
				if filterName != "" {
					attrs.PutStr(AttributeSamplingFilter, filterName)
				}
			}
		}
	}
}

func updateLateArrival(traces ptrace.Traces, filterName string, probabilistic bool) {
	rs := traces.ResourceSpans()

	for i := 0; i < rs.Len(); i++ {
		ss := rs.At(i).ScopeSpans()
		for j := 0; j < ss.Len(); j++ {
			spans := ss.At(j).Spans()
			for k := 0; k < spans.Len(); k++ {
				attrs := spans.At(k).Attributes()
				attrs.PutBool(AttributeSamplingLateArrival, true)
				if filterName != "" {
					attrs.PutStr(AttributeSamplingFilter, filterName)
				} else if probabilistic {
					attrs.PutStr(AttributeSamplingFilter, "probabilistic")
				}
			}
		}
	}
}
