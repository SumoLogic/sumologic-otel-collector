// Copyright 2022 Sumo Logic, Inc.
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

package sumologicschemaprocessor

import (
	"strings"

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
)

type NestingProcessor struct {
	separator string
	enabled   bool
}

func newNestingProcessor(separator string, enabled bool) (*NestingProcessor, error) {
	proc := &NestingProcessor{
		separator: separator,
		enabled:   enabled,
	}

	return proc, nil
}

func (proc *NestingProcessor) processLogs(logs plog.Logs) error {
	if !proc.enabled {
		return nil
	}

	for i := 0; i < logs.ResourceLogs().Len(); i++ {
		rl := logs.ResourceLogs().At(i)

		if err := proc.processAttributes(rl.Resource().Attributes()); err != nil {
			return err
		}

		for j := 0; j < rl.ScopeLogs().Len(); j++ {
			logsRecord := rl.ScopeLogs().At(j).LogRecords()

			for k := 0; k < logsRecord.Len(); k++ {
				if err := proc.processAttributes(logsRecord.At(k).Attributes()); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (proc *NestingProcessor) processMetrics(metrics pmetric.Metrics) error {
	if !proc.enabled {
		return nil
	}

	for i := 0; i < metrics.ResourceMetrics().Len(); i++ {
		rm := metrics.ResourceMetrics().At(i)

		if err := proc.processAttributes(rm.Resource().Attributes()); err != nil {
			return err
		}

		for j := 0; j < rm.ScopeMetrics().Len(); j++ {
			metricsSlice := rm.ScopeMetrics().At(j).Metrics()

			for k := 0; k < metricsSlice.Len(); k++ {
				if err := processMetricLevelAttributes(proc, metricsSlice.At(k)); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (proc *NestingProcessor) processTraces(traces ptrace.Traces) error {
	if !proc.enabled {
		return nil
	}

	for i := 0; i < traces.ResourceSpans().Len(); i++ {
		rs := traces.ResourceSpans().At(i)

		if err := proc.processAttributes(rs.Resource().Attributes()); err != nil {
			return err
		}

		for j := 0; j < rs.ScopeSpans().Len(); j++ {
			spans := rs.ScopeSpans().At(j).Spans()

			for k := 0; k < spans.Len(); k++ {
				if err := proc.processAttributes(spans.At(k).Attributes()); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (proc *NestingProcessor) processAttributes(attributes pcommon.Map) error {
	newMap := pcommon.NewMap()

	attributes.Range(func(k string, v pcommon.Value) bool {
		keys := strings.Split(k, proc.separator)
		if len(keys) == 0 {
			// Split returns empty slice only if both string and separator are empty
			// set map[""] = v and return
			newVal := newMap.PutEmpty(k)
			v.CopyTo(newVal)
			return true
		}

		prevValue := pcommon.NewValueMap()
		nextMap := prevValue.Map()
		newMap.CopyTo(nextMap)

		for i := 0; i < len(keys); i++ {
			if prevValue.Type() != pcommon.ValueTypeMap {
				// If previous value was not a map, change it into a map.
				// The former value will be set under the key "".
				tempMap := pcommon.NewValueMap()
				prevValue.CopyTo(tempMap.Map().PutEmpty(""))
				tempMap.CopyTo(prevValue)
			}

			newValue, ok := prevValue.Map().Get(keys[i])
			if ok {
				prevValue = newValue
			} else if i == len(keys)-1 {
				// If we're checking the last key, insert empty value, to which v will be copied.
				prevValue = prevValue.Map().PutEmpty(keys[i])
			} else {
				// If we're not checking the last key, put a map.
				prevValue = prevValue.Map().PutEmpty(keys[i])
				prevValue.SetEmptyMap()
			}
		}

		if prevValue.Type() == pcommon.ValueTypeMap {
			// Now check the value we want to copy. If it is a map, we should merge both maps.
			// Else, just place the value under the key "".
			if v.Type() == pcommon.ValueTypeMap {
				v.Map().Range(func(k string, val pcommon.Value) bool {
					val.CopyTo(prevValue.Map().PutEmpty(k))
					return true
				})
			} else {
				v.CopyTo(prevValue.Map().PutEmpty(""))
			}
		} else {
			v.CopyTo(prevValue)
		}

		nextMap.CopyTo(newMap)
		return true
	})

	newMap.CopyTo(attributes)

	return nil
}

func (proc *NestingProcessor) isEnabled() bool {
	return proc.enabled
}

func (*NestingProcessor) ConfigPropertyName() string {
	return "nest_attributes"
}
