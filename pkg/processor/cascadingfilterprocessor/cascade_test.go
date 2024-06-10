// Copyright The OpenTelemetry Authors
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

package cascadingfilterprocessor

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.uber.org/zap"

	cfconfig "github.com/SumoLogic/sumologic-otel-collector/pkg/processor/cascadingfilterprocessor/config"
	"github.com/SumoLogic/sumologic-otel-collector/pkg/processor/cascadingfilterprocessor/sampling"
)

var testValue = 10 * time.Millisecond
var probabilisticFilteringRate = int32(10)
var healthCheckPattern = "health"
var statusCode = ptrace.StatusCodeError.String()

var cfg = cfconfig.Config{
	CollectorInstances:      1,
	DecisionWait:            2 * time.Second,
	NumTraces:               100,
	ExpectedNewTracesPerSec: 100,
	SpansPerSecond:          1000,
	PolicyCfgs: []cfconfig.TraceAcceptCfg{
		{
			Name:           "duration",
			SpansPerSecond: 10,
			PropertiesCfg: cfconfig.PropertiesCfg{
				MinDuration: &testValue,
			},
		},
		{
			Name:           "everything else",
			SpansPerSecond: -1,
		},
	},
	TraceRejectCfgs: []cfconfig.TraceRejectCfg{
		{
			Name:        "health-check",
			NamePattern: &healthCheckPattern,
			StatusCode:  &statusCode,
		},
	},
}

var cfgJustDropping = cfconfig.Config{
	CollectorInstances: 1,
	DecisionWait:       2 * time.Second,
	NumTraces:          100,
	TraceRejectCfgs: []cfconfig.TraceRejectCfg{
		{
			Name:        "health-check",
			NamePattern: &healthCheckPattern,
			StatusCode:  &statusCode,
		},
	},
}

var cfgAutoRate = cfconfig.Config{
	CollectorInstances:         1,
	DecisionWait:               2 * time.Second,
	ProbabilisticFilteringRate: &probabilisticFilteringRate,
	NumTraces:                  100,
	PolicyCfgs: []cfconfig.TraceAcceptCfg{
		{
			Name:           "duration",
			SpansPerSecond: 20,
			PropertiesCfg: cfconfig.PropertiesCfg{
				MinDuration: &testValue,
			},
		},
	},
	TraceRejectCfgs: []cfconfig.TraceRejectCfg{
		{
			Name:        "health-check",
			NamePattern: &healthCheckPattern,
			StatusCode:  &statusCode,
		},
	},
}

func fillSpan(span *ptrace.Span, durationMicros int64) {
	nowTs := time.Now().UnixNano()
	startTime := nowTs - durationMicros*1000

	span.Attributes().PutInt("foo", 55)
	span.SetStartTimestamp(pcommon.Timestamp(startTime))
	span.SetEndTimestamp(pcommon.Timestamp(nowTs))
	span.Status().SetCode(ptrace.StatusCodeError)
}

func createTrace(c *cascade, numSpans int, durationMicros int64) *sampling.TraceData {
	var traceBatches []ptrace.Traces

	traces := ptrace.NewTraces()
	rs := traces.ResourceSpans().AppendEmpty()
	ss := rs.ScopeSpans().AppendEmpty()

	spans := ss.Spans()
	spans.EnsureCapacity(numSpans)

	for i := 0; i < numSpans; i++ {
		span := spans.AppendEmpty()

		fillSpan(&span, durationMicros)
	}

	traceBatches = append(traceBatches, traces)

	return &sampling.TraceData{
		Mutex:           sync.Mutex{},
		Decisions:       make([]sampling.Decision, len(c.cfsp.traceAcceptRules)),
		ArrivalTime:     time.Time{},
		DecisionTime:    time.Time{},
		SpanCount:       int32(numSpans),
		ReceivedBatches: traceBatches,
	}
}

func createCascade(t *testing.T) *cascade {
	return createCascadeWithConfig(t, cfg)
}

func createCascadeWithConfig(t *testing.T, conf cfconfig.Config) *cascade {
	cascading, err := newTraceProcessor(zap.NewNop(), nil, conf, component.NewID(Type))
	assert.NoError(t, err)
	return newCascade(cascading)
}

func TestSampling(t *testing.T) {
	cascading := createCascade(t)

	decision, policy := cascading.makeProvisionalDecision(pcommon.TraceID([16]byte{0}), createTrace(cascading, 8, 1000000))
	require.NotNil(t, policy)
	require.Equal(t, sampling.Sampled, decision)

	decision, _ = cascading.makeProvisionalDecision(pcommon.TraceID([16]byte{1}), createTrace(cascading, 1000, 1000))
	require.Equal(t, sampling.SecondChance, decision)
}

func TestSecondChanceEvaluation(t *testing.T) {
	cascading := createCascade(t)

	decision, _ := cascading.makeProvisionalDecision(pcommon.TraceID([16]byte{0}), createTrace(cascading, 8, 1000))
	require.Equal(t, sampling.SecondChance, decision)

	decision, _ = cascading.makeProvisionalDecision(pcommon.TraceID([16]byte{1}), createTrace(cascading, 8, 1000))
	require.Equal(t, sampling.SecondChance, decision)

	// TODO: This could me optimized to make a decision within cascadingfilter processor, as such span would never fit anyway
	//decision, _ = cascading.makeProvisionalDecision(pcommon.TraceID([16]byte{1}), createTrace(8000, 1000), metrics)
	//require.Equal(t, sampling.NotSampled, decision)
}

func TestProbabilisticFilter(t *testing.T) {
	ratio := float32(0.5)
	cfg.ProbabilisticFilteringRatio = &ratio
	cascading := createCascade(t)

	trace1 := createTrace(cascading, 8, 1000000)
	decision, _ := cascading.makeProvisionalDecision(pcommon.TraceID([16]byte{0}), trace1)
	require.Equal(t, sampling.Sampled, decision)
	require.True(t, trace1.SelectedByProbabilisticFilter)

	trace2 := createTrace(cascading, 800, 1000000)
	decision, _ = cascading.makeProvisionalDecision(pcommon.TraceID([16]byte{1}), trace2)
	require.Equal(t, sampling.SecondChance, decision)
	require.False(t, trace2.SelectedByProbabilisticFilter)

	ratio = float32(0.0)
	cfg.ProbabilisticFilteringRatio = &ratio
}

func TestDropTraces(t *testing.T) {
	cascading := createCascade(t)

	trace1 := createTrace(cascading, 8, 1000000)
	trace2 := createTrace(cascading, 8, 1000000)
	trace2.ReceivedBatches[0].ResourceSpans().At(0).ScopeSpans().At(0).Spans().At(2).SetName("health-check")
	require.False(t, cascading.shouldBeDropped(pcommon.TraceID([16]byte{0}), trace1))
	require.True(t, cascading.shouldBeDropped(pcommon.TraceID([16]byte{0}), trace2))
}

func TestDropTracesWithDifferentStatusCode(t *testing.T) {
	cascading := createCascade(t)

	trace1 := createTrace(cascading, 1, 1000000)
	trace2 := createTrace(cascading, 1, 1000000)
	trace3 := createTrace(cascading, 1, 1000000)

	trace1.ReceivedBatches[0].ResourceSpans().At(0).ScopeSpans().At(0).Spans().At(0).SetName("health-check-trace-1")
	trace1.ReceivedBatches[0].ResourceSpans().At(0).ScopeSpans().At(0).Spans().At(0).Status().SetCode(ptrace.StatusCodeUnset)
	trace2.ReceivedBatches[0].ResourceSpans().At(0).ScopeSpans().At(0).Spans().At(0).SetName("health-check-trace-2")
	trace2.ReceivedBatches[0].ResourceSpans().At(0).ScopeSpans().At(0).Spans().At(0).Status().SetCode(ptrace.StatusCodeOk)
	trace3.ReceivedBatches[0].ResourceSpans().At(0).ScopeSpans().At(0).Spans().At(0).SetName("health-check-trace-3")

	require.False(t, cascading.shouldBeDropped(pcommon.TraceID([16]byte{0}), trace1))
	require.False(t, cascading.shouldBeDropped(pcommon.TraceID([16]byte{0}), trace2))
	require.True(t, cascading.shouldBeDropped(pcommon.TraceID([16]byte{0}), trace3))
}

func TestDropTracesAndNotLimitOthers(t *testing.T) {
	cascading := createCascadeWithConfig(t, cfgJustDropping)

	trace1 := createTrace(cascading, 1000, 1000000)
	trace2 := createTrace(cascading, 8, 1000000)
	trace2.ReceivedBatches[0].ResourceSpans().At(0).ScopeSpans().At(0).Spans().At(2).SetName("health-check")
	trace3 := createTrace(cascading, 5000, 1000000)

	decision, policy := cascading.makeProvisionalDecision(pcommon.TraceID([16]byte{0}), trace1)
	require.Nil(t, policy)
	require.Equal(t, sampling.Sampled, decision)
	require.False(t, cascading.shouldBeDropped(pcommon.TraceID([16]byte{0}), trace1))

	decision, policy = cascading.makeProvisionalDecision(pcommon.TraceID([16]byte{1}), trace2)
	require.Nil(t, policy)
	require.Equal(t, sampling.Sampled, decision)
	require.True(t, cascading.shouldBeDropped(pcommon.TraceID([16]byte{1}), trace2))

	decision, policy = cascading.makeProvisionalDecision(pcommon.TraceID([16]byte{2}), trace3)
	require.Nil(t, policy)
	require.Equal(t, sampling.Sampled, decision)
	require.False(t, cascading.shouldBeDropped(pcommon.TraceID([16]byte{2}), trace3))
}

func TestDropTracesAndAutoRateOthers(t *testing.T) {
	cascading := createCascadeWithConfig(t, cfgAutoRate)

	trace1 := createTrace(cascading, 20, 1000000)
	trace2 := createTrace(cascading, 8, 1000000)
	trace2.ReceivedBatches[0].ResourceSpans().At(0).ScopeSpans().At(0).Spans().At(2).SetName("health-check")
	trace3 := createTrace(cascading, 20, 1000000)

	decision, policy := cascading.makeProvisionalDecision(pcommon.TraceID([16]byte{0}), trace1)
	require.NotNil(t, policy)
	require.Equal(t, sampling.Sampled, decision)
	require.False(t, cascading.shouldBeDropped(pcommon.TraceID([16]byte{0}), trace1))

	decision, policy = cascading.makeProvisionalDecision(pcommon.TraceID([16]byte{1}), trace2)
	require.NotNil(t, policy)
	require.Equal(t, sampling.Sampled, decision)
	require.True(t, cascading.shouldBeDropped(pcommon.TraceID([16]byte{1}), trace2))

	decision, policy = cascading.makeProvisionalDecision(pcommon.TraceID([16]byte{2}), trace3)
	require.Nil(t, policy)
	require.Equal(t, sampling.NotSampled, decision)
	require.False(t, cascading.shouldBeDropped(pcommon.TraceID([16]byte{2}), trace3))
}

//func TestSecondChanceReevaluation(t *testing.T) {
//	cascading := createCascade()
//
//	decision, _ := cascading.makeProvisionalDecision(pcommon.TraceID([16]byte{1}), createTrace(100, 1000), metrics)
//	require.Equal(t, sampling.Sampled, decision)
//
//	// Too much
//	decision, _ = cascading.makeProvisionalDecision(pcommon.TraceID([16]byte{1}), createTrace(1000, 1000), metrics)
//	require.Equal(t, sampling.NotSampled, decision)
//
//	// Just right
//	decision, _ = cascading.makeProvisionalDecision(pcommon.TraceID([16]byte{1}), createTrace(900, 1000), metrics)
//	require.Equal(t, sampling.Sampled, decision)
//}
