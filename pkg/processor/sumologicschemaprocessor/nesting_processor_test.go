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
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/pcommon"
)

func TestNestingAttributes(t *testing.T) {
	testCases := []struct {
		name     string
		input    map[string]pcommon.Value
		expected map[string]pcommon.Value
	}{
		{
			name: "sample nesting",
			input: map[string]pcommon.Value{
				"kubernetes.container_name": pcommon.NewValueStr("xyz"),
				"kubernetes.host.name":      pcommon.NewValueStr("the host"),
				"kubernetes.host.address":   pcommon.NewValueStr("127.0.0.1"),
				"kubernetes.namespace_name": pcommon.NewValueStr("sumologic"),
				"another_attr":              pcommon.NewValueStr("42"),
			},
			expected: map[string]pcommon.Value{
				"kubernetes": mapToPcommonValue(map[string]pcommon.Value{
					"container_name": pcommon.NewValueStr("xyz"),
					"namespace_name": pcommon.NewValueStr("sumologic"),
					"host": mapToPcommonValue(map[string]pcommon.Value{
						"name":    pcommon.NewValueStr("the host"),
						"address": pcommon.NewValueStr("127.0.0.1"),
					}),
				}),
				"another_attr": pcommon.NewValueStr("42"),
			},
		},
		{
			name: "single values",
			input: map[string]pcommon.Value{
				"a": mapToPcommonValue(map[string]pcommon.Value{
					"b": mapToPcommonValue(map[string]pcommon.Value{
						"c": pcommon.NewValueStr("d"),
					}),
				}),
				"a.b.c": pcommon.NewValueStr("d"),
				"d.g.e": pcommon.NewValueStr("l"),
				"b.g.c": pcommon.NewValueStr("bonus"),
			},
			expected: map[string]pcommon.Value{
				"a": mapToPcommonValue(map[string]pcommon.Value{
					"b": mapToPcommonValue(map[string]pcommon.Value{
						"c": pcommon.NewValueStr("d"),
					}),
				}),
				"d": mapToPcommonValue(map[string]pcommon.Value{
					"g": mapToPcommonValue(map[string]pcommon.Value{
						"e": pcommon.NewValueStr("l"),
					}),
				}),
				"b": mapToPcommonValue(map[string]pcommon.Value{
					"g": mapToPcommonValue(map[string]pcommon.Value{
						"c": pcommon.NewValueStr("bonus"),
					}),
				}),
			},
		},
		{
			name: "overwrite map with simple value",
			input: map[string]pcommon.Value{
				"sumo.logic": pcommon.NewValueBool(true),
				"sumo":       pcommon.NewValueBool(false),
			},
			expected: map[string]pcommon.Value{
				"sumo": mapToPcommonValue(map[string]pcommon.Value{
					"":      pcommon.NewValueBool(false),
					"logic": pcommon.NewValueBool(true),
				}),
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			proc, err := newNestingProcessor(&NestingProcessorConfig{Separator: ".", Enabled: true})
			require.NoError(t, err)

			attrs := mapToPcommonMap(testCase.input)
			err = proc.processAttributes(attrs)
			require.NoError(t, err)

			expected := mapToPcommonMap(testCase.expected)

			require.Equal(t, expected.AsRaw(), attrs.AsRaw())
		})
	}
}
