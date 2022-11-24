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
	"math"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.uber.org/zap"
)

func newNumericAttributeFilter(minValue int64, maxValue int64) *policyEvaluator {
	return &policyEvaluator{
		logger: zap.NewNop(),
		numericAttr: &numericAttributeFilter{
			key:      "example",
			minValue: minValue,
			maxValue: maxValue,
		},
		maxSpansPerSecond: math.MaxInt32,
	}
}

func TestNumericTagFilter(t *testing.T) {
	var empty = map[string]interface{}{}
	filter := newNumericAttributeFilter(math.MinInt32, math.MaxInt32)

	resAttr := map[string]interface{}{}
	resAttr["example"] = pcommon.NewValueInt(8)

	cases := []struct {
		Desc     string
		Trace    *TraceData
		Decision Decision
	}{
		{
			Desc:     "nonmatching span attribute",
			Trace:    newTraceIntAttrs(empty, "non_matching", math.MinInt32),
			Decision: NotSampled,
		},
		{
			Desc:     "span attribute with lower limit",
			Trace:    newTraceIntAttrs(empty, "example", math.MinInt32),
			Decision: Sampled,
		},
		{
			Desc:     "span attribute with upper limit",
			Trace:    newTraceIntAttrs(empty, "example", math.MaxInt32),
			Decision: Sampled,
		},
		{
			Desc:     "span attribute below min limit",
			Trace:    newTraceIntAttrs(empty, "example", math.MinInt32-1),
			Decision: NotSampled,
		},
		{
			Desc:     "span attribute above max limit",
			Trace:    newTraceIntAttrs(empty, "example", math.MaxInt32+1),
			Decision: NotSampled,
		},
	}

	for _, c := range cases {
		t.Run(c.Desc, func(t *testing.T) {
			u, err := uuid.NewRandom()
			require.NoError(t, err)
			decision := filter.Evaluate(pcommon.TraceID(u), c.Trace)
			assert.Equal(t, decision, c.Decision)
		})
	}
}

func newTraceIntAttrs(nodeAttrs map[string]interface{}, spanAttrKey string, spanAttrValue int64) *TraceData {
	var traceBatches []ptrace.Traces
	traces := ptrace.NewTraces()
	rs := traces.ResourceSpans().AppendEmpty()
	m := pcommon.NewMap()
	err := m.FromRaw(nodeAttrs)
	if err != nil {
		return &TraceData{}
	}
	m.CopyTo(rs.Resource().Attributes())
	ss := rs.ScopeSpans().AppendEmpty()
	span := ss.Spans().AppendEmpty()
	span.SetTraceID(pcommon.TraceID([16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}))
	span.SetSpanID(pcommon.SpanID([8]byte{1, 2, 3, 4, 5, 6, 7, 8}))
	span.Attributes().PutInt(spanAttrKey, spanAttrValue)
	traceBatches = append(traceBatches, traces)
	return &TraceData{
		ReceivedBatches: traceBatches,
	}
}
