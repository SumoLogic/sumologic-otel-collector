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
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
)

const (
	levelAttributeName = "loglevel"
)

var severityNumberToLevel = map[string]string{
	plog.SeverityNumberUnspecified.String(): "UNSPECIFIED",
	plog.SeverityNumberTrace.String():       "TRACE",
	plog.SeverityNumberTrace2.String():      "TRACE2",
	plog.SeverityNumberTrace3.String():      "TRACE3",
	plog.SeverityNumberTrace4.String():      "TRACE4",
	plog.SeverityNumberDebug.String():       "DEBUG",
	plog.SeverityNumberDebug2.String():      "DEBUG2",
	plog.SeverityNumberDebug3.String():      "DEBUG3",
	plog.SeverityNumberDebug4.String():      "DEBUG4",
	plog.SeverityNumberInfo.String():        "INFO",
	plog.SeverityNumberInfo2.String():       "INFO2",
	plog.SeverityNumberInfo3.String():       "INFO3",
	plog.SeverityNumberInfo4.String():       "INFO4",
	plog.SeverityNumberWarn.String():        "WARN",
	plog.SeverityNumberWarn2.String():       "WARN2",
	plog.SeverityNumberWarn3.String():       "WARN3",
	plog.SeverityNumberWarn4.String():       "WARN4",
	plog.SeverityNumberError.String():       "ERROR",
	plog.SeverityNumberError2.String():      "ERROR2",
	plog.SeverityNumberError3.String():      "ERROR3",
	plog.SeverityNumberError4.String():      "ERROR4",
	plog.SeverityNumberFatal.String():       "FATAL",
	plog.SeverityNumberFatal2.String():      "FATAL2",
	plog.SeverityNumberFatal3.String():      "FATAL3",
	plog.SeverityNumberFatal4.String():      "FATAL4",
}

// logFieldsConversionProcessor converts specific log entries to attributes which leads to presenting them as fields
// in the backend
type logFieldsConversionProcessor struct {
	shouldConvert bool
}

func newLogFieldConversionProcessor(shouldConvert bool) (*logFieldsConversionProcessor, error) {
	return &logFieldsConversionProcessor{
		shouldConvert: shouldConvert,
	}, nil
}

func addLogLevelAttribute(log plog.LogRecord) {
	if log.SeverityNumber() == plog.SeverityNumberUnspecified {
		return
	}
	if _, found := log.Attributes().Get(levelAttributeName); !found {
		level := severityNumberToLevel[log.SeverityNumber().String()]
		log.Attributes().PutStr(levelAttributeName, level)
	}
}

func (proc *logFieldsConversionProcessor) processLogs(logs plog.Logs) error {
	if proc.shouldConvert {
		rls := logs.ResourceLogs()
		for i := 0; i < rls.Len(); i++ {
			ills := rls.At(i).ScopeLogs()

			for j := 0; j < ills.Len(); j++ {
				logs := ills.At(j).LogRecords()
				for k := 0; k < logs.Len(); k++ {
					// adds level attribute from log.severityNumber
					addLogLevelAttribute(logs.At(k))
				}
			}
		}
	}
	return nil
}

func (proc *logFieldsConversionProcessor) processMetrics(metrics pmetric.Metrics) error {
	// No-op. Metrics should not be translated.
	return nil
}

func (proc *logFieldsConversionProcessor) processTraces(traces ptrace.Traces) error {
	// No-op. Traces should not be translated.
	return nil
}

func (proc *logFieldsConversionProcessor) isEnabled() bool {
	return proc.shouldConvert
}

func (*logFieldsConversionProcessor) ConfigPropertyName() string {
	return "add_severity_level_attribute"
}
