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
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.uber.org/zap"
)

func newStringAttributeFilter() *policyEvaluator {
	return &policyEvaluator{
		logger: zap.NewNop(),
		stringAttr: &stringAttributeFilter{
			key:    "example",
			values: map[string]struct{}{"value": {}},
		},
		maxSpansPerSecond: math.MaxInt32,
	}
}

func newStringAttributeRegexFilter() *policyEvaluator {
	return &policyEvaluator{
		logger: zap.NewNop(),
		stringAttr: &stringAttributeFilter{
			key:      "example",
			patterns: []*regexp.Regexp{regexp.MustCompile("val.*")},
			values:   map[string]struct{}{},
		},
		maxSpansPerSecond: math.MaxInt32,
	}
}

func TestStringTagFilter(t *testing.T) {
	var empty = map[string]interface{}{}
	filter := newStringAttributeFilter()
	regexFilter := newStringAttributeRegexFilter()

	cases := []struct {
		Desc     string
		Trace    *TraceData
		Decision Decision
	}{
		{
			Desc:     "nonmatching node attribute key",
			Trace:    newTraceStringAttrs(map[string]interface{}{"non_matching": "value"}, "", ""),
			Decision: NotSampled,
		},
		{
			Desc:     "nonmatching node attribute value",
			Trace:    newTraceStringAttrs(map[string]interface{}{"example": "non_matching"}, "", ""),
			Decision: NotSampled,
		},
		{
			Desc:     "matching node attribute",
			Trace:    newTraceStringAttrs(map[string]interface{}{"example": "value"}, "", ""),
			Decision: Sampled,
		},
		{
			Desc:     "nonmatching span attribute key",
			Trace:    newTraceStringAttrs(empty, "nonmatching", "value"),
			Decision: NotSampled,
		},
		{
			Desc:     "nonmatching span attribute value",
			Trace:    newTraceStringAttrs(empty, "example", "nonmatching"),
			Decision: NotSampled,
		},
		{
			Desc:     "matching span attribute",
			Trace:    newTraceStringAttrs(empty, "example", "value"),
			Decision: Sampled,
		},
	}

	for _, c := range cases {
		t.Run(c.Desc, func(t *testing.T) {
			decisionPlain := filter.Evaluate(pcommon.TraceID([16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}), c.Trace)
			decisionRegex := regexFilter.Evaluate(pcommon.TraceID([16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}), c.Trace)
			assert.Equal(t, decisionPlain, c.Decision)
			assert.Equal(t, decisionRegex, c.Decision)
		})
	}
}

func newTraceStringAttrs(nodeAttrs map[string]interface{}, spanAttrKey string, spanAttrValue string) *TraceData {
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
	span.Attributes().PutStr(spanAttrKey, spanAttrValue)
	traceBatches = append(traceBatches, traces)
	return &TraceData{
		ReceivedBatches: traceBatches,
	}
}
