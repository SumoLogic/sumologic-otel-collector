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

package sampling

import (
	"sync"
	"time"

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/ptrace"
)

// TraceData stores the sampling related trace data.
type TraceData struct {
	sync.Mutex
	// Decisions gives the current status of the sampling decision for each policy.
	Decisions []Decision
	// FinalDecision describes the ultimate fate of the trace
	FinalDecision Decision
	// SelectedByProbabilisticFilter determines if this trace was selected by probabilistic filter
	SelectedByProbabilisticFilter bool
	// ProvisionalDecisionFilter includes the name of the filter which has selected the trace
	ProvisionalDecisionFilterName string
	// Arrival time the first span for the trace was received.
	ArrivalTime time.Time
	// Decisiontime time when sampling decision was taken.
	DecisionTime time.Time
	// SpanCount track the number of spans on the trace.
	SpanCount int32
	// ReceivedBatches stores all the batches received for the trace.
	ReceivedBatches []ptrace.Traces
}

// Decision gives the status of sampling decision.
type Decision int32

const (
	// Unspecified indicates that the status of the decision was not set yet.
	Unspecified Decision = iota
	// Pending indicates that the policy was not evaluated yet.
	Pending
	// Sampled is used to indicate that the decision was already taken
	// to sample the data.
	Sampled
	// SecondChance is a special category that allows to make a final decision
	// after all batches are processed. It should be converted to Sampled or NotSampled
	SecondChance
	// NotSampled is used to indicate that the decision was already taken
	// to not sample the data.
	NotSampled
	// Dropped is used when data needs to be purged before the sampling policy
	// had a chance to evaluate it.
	Dropped
)

// PolicyEvaluator implements a cascading policy evaluator,
// which makes a sampling decision for a given trace when requested.
type PolicyEvaluator interface {
	// Evaluate looks at the trace data and returns a corresponding SamplingDecision.
	Evaluate(traceID pcommon.TraceID, trace *TraceData) Decision
}

// DropTraceEvaluator implements a cascading policy evaluator,
// which checks if trace should be dropped completely before making any other operations
type DropTraceEvaluator interface {
	// ShouldDrop checks if trace should be dropped
	ShouldDrop(traceID pcommon.TraceID, trace *TraceData) bool
}
