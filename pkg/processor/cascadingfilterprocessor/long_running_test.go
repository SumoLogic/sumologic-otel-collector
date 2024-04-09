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
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.uber.org/zap"

	"github.com/SumoLogic/sumologic-otel-collector/pkg/processor/cascadingfilterprocessor/bigendianconverter"

	cfconfig "github.com/SumoLogic/sumologic-otel-collector/pkg/processor/cascadingfilterprocessor/config"
	"github.com/SumoLogic/sumologic-otel-collector/pkg/processor/cascadingfilterprocessor/sampling"
)

const skipLongRunning = true

func TestRandomTraceProcessing(t *testing.T) {
	if skipLongRunning {
		t.Skip("a long running test, please run manually")
	}

	stepDuration := time.Millisecond * 100
	decisionWait := time.Second * 1

	tsp := newLongRunningTraceProcessor(t, decisionWait)
	allIDs := map[pcommon.TraceID]interface{}{}

	for i := 0; i < 100; i++ {
		ids, batches := generateTraces(i*2, 100, 1, 1)
		for _, td := range batches {
			if err := tsp.ConsumeTraces(context.Background(), td); err != nil {
				t.Errorf("Failed consuming traces: %v", err)
			}
		}
		for _, id := range ids {
			allIDs[id] = true
		}
		time.Sleep(stepDuration)
	}

	time.Sleep(decisionWait * 3)

	for traceId := range allIDs {
		d, ok := tsp.idToTrace.Load(traceId)
		if ok {
			v := d.(*sampling.TraceData)
			require.Empty(t, v.ReceivedBatches)
		}
	}
}

func TestTraceProcessing(t *testing.T) {
	if skipLongRunning {
		t.Skip("a long running test, please run manually")
	}

	decisionWait := time.Second * 1

	tsp := newLongRunningTraceProcessor(t, decisionWait)

	allIDs := generateAndConsumeTraces(t, tsp)
	time.Sleep(decisionWait * 3)
	assertTracesEmpty(t, tsp, allIDs)

	allIDs = generateAndConsumeTraces(t, tsp)
	time.Sleep(decisionWait * 3)
	assertTracesEmpty(t, tsp, allIDs)
}

func generateTraces(startingTraceId int, traceCount int, spanCount int, increaseMultiplier int) ([]pcommon.TraceID, []ptrace.Traces) {
	traceIds := make([]pcommon.TraceID, traceCount)
	spanID := 0
	var tds []ptrace.Traces
	for i := 0; i < traceCount; i++ {
		traceIds[i] = bigendianconverter.UInt64ToTraceID(1, uint64(startingTraceId+(i*increaseMultiplier)))
		// Send each span in a separate batch
		for j := 0; j < spanCount; j++ {
			td := simpleTraces()
			span := td.ResourceSpans().At(0).ScopeSpans().At(0).Spans().At(0)
			span.SetTraceID(traceIds[i])

			spanID++
			span.SetSpanID(bigendianconverter.UInt64ToSpanID(uint64(spanID)))
			tds = append(tds, td)
		}
	}

	return traceIds, tds
}

func newLongRunningTraceProcessor(t *testing.T, decisionWait time.Duration) *cascadingFilterSpanProcessor {
	outputRate := int32(10)

	cfg := cfconfig.Config{
		DecisionWait:               decisionWait,
		ProbabilisticFilteringRate: &outputRate,
		NumTraces:                  100,
	}
	sp, err := newTraceProcessor(zap.NewNop(), consumertest.NewNop(), cfg, component.NewID(Type))
	require.NoError(t, err)
	return sp
}

func generateAndConsumeTraces(t *testing.T, tsp *cascadingFilterSpanProcessor) map[pcommon.TraceID]interface{} {
	allIDs := map[pcommon.TraceID]interface{}{}

	ids, batches := generateTraces(0, 1, 11, 0)
	for _, td := range batches {
		if err := tsp.ConsumeTraces(context.Background(), td); err != nil {
			t.Errorf("Failed consuming traces: %v", err)
		}
	}
	for _, id := range ids {
		allIDs[id] = true
	}

	return allIDs
}

func assertTracesEmpty(t *testing.T, tsp *cascadingFilterSpanProcessor, allIDs map[pcommon.TraceID]interface{}) {
	for traceId := range allIDs {
		d, ok := tsp.idToTrace.Load(traceId)
		if ok {
			v := d.(*sampling.TraceData)
			require.Empty(t, v.ReceivedBatches)
		}
	}
}
