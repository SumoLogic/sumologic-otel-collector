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

package sampling

import (
	"time"

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/ptrace"
)

func tsToMicros(ts pcommon.Timestamp) int64 {
	return int64(ts / 1000)
}

func checkIfAttrsMatched(resAttrs pcommon.Map, spanAttrs pcommon.Map, filters []attributeFilter) bool {
	for _, filter := range filters {
		var resAttrMatched bool
		spanAttrMatched, spanAttrFound := checkAttributeFilterMatchedAndFound(spanAttrs, filter)
		if !spanAttrFound {
			resAttrMatched, _ = checkAttributeFilterMatchedAndFound(resAttrs, filter)
		}

		if !resAttrMatched && !spanAttrMatched {
			return false
		}
	}
	return true
}

func checkAttributeFilterMatchedAndFound(attrs pcommon.Map, filter attributeFilter) (bool, bool) {
	if v, ok := attrs.Get(filter.key); ok {
		// String patterns vs values is exclusive
		if len(filter.patterns) > 0 {
			// Pattern matching
			truncableStr := v.Str()
			for _, re := range filter.patterns {
				if re.MatchString(truncableStr) {
					return true, true
				}
			}
		} else if len(filter.values) > 0 {
			// Exact matching
			truncableStr := v.Str()
			if len(truncableStr) > 0 {
				if _, ok := filter.values[truncableStr]; ok {
					return true, true
				}
			}
		}

		if len(filter.ranges) > 0 {
			if v.Type() == pcommon.ValueTypeDouble {
				value := v.Double()
				for _, r := range filter.ranges {
					if value >= float64(r.minValue) && value <= float64(r.maxValue) {
						return true, true
					}
				}
			} else if v.Type() == pcommon.ValueTypeInt {
				value := v.Int()
				for _, r := range filter.ranges {
					if value >= r.minValue && value <= r.maxValue {
						return true, true
					}
				}
			}
		}

		// This is special condition which just checks if any filters were defined or not; For latter, pass if key found
		if len(filter.ranges) == 0 && len(filter.values) == 0 && len(filter.patterns) == 0 {
			return true, true
		}

		return false, true
	}

	// Not found and not matched
	return false, false
}

func checkIfNumericAttrFound(attrs pcommon.Map, filter *numericAttributeFilter) bool {
	if v, ok := attrs.Get(filter.key); ok {
		value := v.Int()
		if value >= filter.minValue && value <= filter.maxValue {
			return true
		}
	}
	return false
}

func checkIfStringAttrFound(attrs pcommon.Map, filter *stringAttributeFilter) bool {
	if v, ok := attrs.Get(filter.key); ok {
		truncableStr := v.Str()
		if filter.patterns != nil {
			// Pattern matching
			for _, re := range filter.patterns {
				if re.MatchString(truncableStr) {
					return true
				}
			}
		} else {
			// Exact matching
			if len(truncableStr) > 0 {
				if _, ok := filter.values[truncableStr]; ok {
					return true
				}
			}
		}
	}
	return false
}

// evaluateRules goes through the defined properties and checks if they are matched
func (pe *policyEvaluator) evaluateRules(_ pcommon.TraceID, trace *TraceData) Decision {
	trace.Lock()
	batches := trace.ReceivedBatches
	trace.Unlock()

	matchingOperationFound := false
	matchingStringAttrFound := false
	matchingNumericAttrFound := false
	matchingAttrsFound := false

	spanCount := 0
	errorCount := 0
	minStartTime := int64(0)
	maxEndTime := int64(0)

	for _, batch := range batches {
		rs := batch.ResourceSpans()

		for i := 0; i < rs.Len(); i++ {
			res := rs.At(i).Resource()

			if !matchingStringAttrFound && pe.stringAttr != nil {
				matchingStringAttrFound = checkIfStringAttrFound(res.Attributes(), pe.stringAttr)
			}

			if !matchingNumericAttrFound && pe.numericAttr != nil {
				matchingNumericAttrFound = checkIfNumericAttrFound(res.Attributes(), pe.numericAttr)
			}

			ss := rs.At(i).ScopeSpans()
			for j := 0; j < ss.Len(); j++ {
				spans := ss.At(j).Spans()
				spanCount += spans.Len()
				for k := 0; k < spans.Len(); k++ {
					span := spans.At(k)

					if !matchingAttrsFound && len(pe.attrs) > 0 {
						matchingAttrsFound = checkIfAttrsMatched(res.Attributes(), span.Attributes(), pe.attrs)
					}

					if !matchingStringAttrFound && pe.stringAttr != nil {
						matchingStringAttrFound = checkIfStringAttrFound(span.Attributes(), pe.stringAttr)
					}

					if !matchingNumericAttrFound && pe.numericAttr != nil {
						matchingNumericAttrFound = checkIfNumericAttrFound(span.Attributes(), pe.numericAttr)
					}

					if pe.operationRe != nil && !matchingOperationFound {
						if pe.operationRe.MatchString(span.Name()) {
							matchingOperationFound = true
						}
					}

					if pe.minDuration != nil {
						startTs := tsToMicros(span.StartTimestamp())
						endTs := tsToMicros(span.EndTimestamp())

						if minStartTime == 0 {
							minStartTime = startTs
							maxEndTime = endTs
						} else {
							if startTs < minStartTime {
								minStartTime = startTs
							}
							if endTs > maxEndTime {
								maxEndTime = endTs
							}
						}
					}

					if span.Status().Code() == ptrace.StatusCodeError {
						errorCount++
					}
				}
			}
		}
	}

	conditionMet := struct {
		operationName, minDuration, minSpanCount, stringAttr, numericAttr, attrs, minErrorCount bool
	}{
		operationName: true,
		minDuration:   true,
		minSpanCount:  true,
		stringAttr:    true,
		numericAttr:   true,
		attrs:         true,
		minErrorCount: true,
	}

	if pe.operationRe != nil {
		conditionMet.operationName = matchingOperationFound
	}
	if pe.minNumberOfSpans != nil {
		conditionMet.minSpanCount = spanCount >= *pe.minNumberOfSpans
	}
	if pe.minDuration != nil {
		conditionMet.minDuration = maxEndTime > minStartTime && maxEndTime-minStartTime >= pe.minDuration.Microseconds()
	}
	if pe.numericAttr != nil {
		conditionMet.numericAttr = matchingNumericAttrFound
	}
	if pe.stringAttr != nil {
		conditionMet.stringAttr = matchingStringAttrFound
	}
	if len(pe.attrs) > 0 {
		conditionMet.attrs = matchingAttrsFound
	}
	if pe.minNumberOfErrors != nil {
		conditionMet.minErrorCount = errorCount >= *pe.minNumberOfErrors
	}

	if conditionMet.minSpanCount &&
		conditionMet.minDuration &&
		conditionMet.operationName &&
		conditionMet.numericAttr &&
		conditionMet.stringAttr &&
		conditionMet.attrs &&
		conditionMet.minErrorCount {
		if pe.invertMatch {
			return NotSampled
		}
		return Sampled
	}

	if pe.invertMatch {
		return Sampled
	}
	return NotSampled
}

func (pe *policyEvaluator) shouldConsider(currSecond int64, trace *TraceData) bool {
	if pe.maxSpansPerSecond < 0 {
		// This emits "second chance" traces
		return true
	} else if trace.SpanCount > pe.maxSpansPerSecond {
		// This trace will never fit, there are more spans than max limit
		return false
	} else if pe.currentSecond == currSecond && trace.SpanCount > pe.maxSpansPerSecond-pe.spansInCurrentSecond {
		// This trace will not fit in this second, no way
		return false
	} else {
		// This has some chances
		return true
	}
}

func (pe *policyEvaluator) emitsSecondChance() bool {
	return pe.maxSpansPerSecond < 0
}

func (pe *policyEvaluator) updateRate(currSecond int64, numSpans int32) Decision {
	if pe.currentSecond != currSecond {
		pe.currentSecond = currSecond
		pe.spansInCurrentSecond = 0
	}

	spansInSecondIfSampled := pe.spansInCurrentSecond + numSpans
	if spansInSecondIfSampled <= pe.maxSpansPerSecond {
		pe.spansInCurrentSecond = spansInSecondIfSampled
		return Sampled
	}

	return NotSampled
}

// Evaluate looks at the trace data and returns a corresponding SamplingDecision. Also takes into account
// the usage of sampling rate budget
func (pe *policyEvaluator) Evaluate(traceID pcommon.TraceID, trace *TraceData) Decision {
	currSecond := time.Now().Unix()

	if !pe.shouldConsider(currSecond, trace) {
		return NotSampled
	}

	decision := pe.evaluateRules(traceID, trace)
	if decision != Sampled {
		return decision
	}

	if pe.emitsSecondChance() {
		return SecondChance
	}

	return pe.updateRate(currSecond, trace.SpanCount)
}
