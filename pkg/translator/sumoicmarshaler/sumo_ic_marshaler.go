// Copyright 2022, OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package sumoicmarshaler

import (
	"encoding/json"

	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/ptrace"
)

const (
	SourceCategoryKey = "_sourceCategory"
	SourceHostKey     = "_sourceHost"
	SourceNameKey     = "log.file.path_resolved"
)

type SumoICMarshaler struct {
	format string
}

type fieldSet struct {
	attributes map[string]interface{}
}

type icRecord struct {
	Data           string   `json:"data"`
	SourceName     string   `json:"sourceName"`
	SourceHost     string   `json:"sourceHost"`
	SourceCategory string   `json:"sourceCategory"`
	Fields         fieldSet `json:"fields"`
	Message        string   `json:"message"`
}

func (fields fieldSet) MarshalJSON() ([]byte, error) {
	return json.Marshal(fields.attributes)
}

func transformLogsToIc(ld plog.Logs) ([]icRecord, error) {
	icRecords := []icRecord{}
	rls := ld.ResourceLogs()

	for i := 0; i < rls.Len(); i++ {
		rl := rls.At(i)
		sourceCategoryVal := ""
		sourceCategory, exists := rl.Resource().Attributes().Get(SourceCategoryKey)
		if exists {
			sourceCategoryVal = sourceCategory.AsString()
		}
		sourceHostVal := ""
		sourceHost, exists := rl.Resource().Attributes().Get(SourceHostKey)
		if exists {
			sourceHostVal = sourceHost.AsString()
		}
		fields := fieldSet{attributes: make(map[string]interface{})}
		attrs := rl.Resource().Attributes().AsRaw()

		ills := rl.ScopeLogs()
		for j := 0; j < ills.Len(); j++ {
			ils := ills.At(j)
			logs := ils.LogRecords()
			for k := 0; k < logs.Len(); k++ {
				lr := logs.At(k)
				sourceNameVal := ""
				sourceName, exists := lr.Attributes().Get(SourceNameKey)
				if exists {
					sourceNameVal = sourceName.AsString()
				}

				for ak, av := range attrs {
					fields.attributes[ak] = av
				}

				icRecords = append(icRecords, icRecord{lr.ObservedTimestamp().String(),
					sourceNameVal, sourceHostVal,
					sourceCategoryVal, fields, lr.Body().AsString()})
			}
		}
	}
	return icRecords, nil
}

func (m *SumoICMarshaler) MarshalLogs(ld plog.Logs) ([]byte, error) {
	buffer := []byte{}
	icRecords, err := transformLogsToIc(ld)
	if err != nil {
		return nil, err
	}
	for _, record := range icRecords {
		marshalledRecord, err := json.Marshal(record)
		if err != nil {
			return nil, err
		}

		buffer = append(buffer, marshalledRecord...)
	}
	return buffer, nil
}

func (m *SumoICMarshaler) MarshalTraces(traces ptrace.Traces) ([]byte, error) {
	return nil, nil
}
