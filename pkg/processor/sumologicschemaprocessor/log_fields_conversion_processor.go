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
	"encoding/hex"

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
)

const (
	SeverityNumberAttributeName = "loglevel"
	SeverityTextAttributeName   = "severitytext"
	SpanIdAttributeName         = "spanid"
	TraceIdAttributeName        = "traceid"
)

type logFieldAttribute struct {
	Enabled       bool   `mapstructure:"enabled"`
	AttributeName string `mapstructure:"attribute_name"`
}

// SpanIDToHexOrEmptyString returns a hex string from SpanID.
// An empty string is returned, if SpanID is empty.
func SpanIDToHexOrEmptyString(id pcommon.SpanID) string {
	if id.IsEmpty() {
		return ""
	}
	return hex.EncodeToString(id[:])
}

// TraceIDToHexOrEmptyString returns a hex string from TraceID.
// An empty string is returned, if TraceID is empty.
func TraceIDToHexOrEmptyString(id pcommon.TraceID) string {
	if id.IsEmpty() {
		return ""
	}
	return hex.EncodeToString(id[:])
}

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
	severityNumberEnabled *logFieldAttribute
	severityTextEnabled   *logFieldAttribute
	spanIdEnabled         *logFieldAttribute
	traceIdEnabled        *logFieldAttribute
}

func newLogFieldConversionProcessor(severityNumberEnabled *logFieldAttribute,
	severityTextEnabled *logFieldAttribute,
	spanIdEnabled *logFieldAttribute,
	traceIdEnabled *logFieldAttribute) (*logFieldsConversionProcessor, error) {
	return &logFieldsConversionProcessor{
		severityNumberEnabled: severityNumberEnabled,
		severityTextEnabled:   severityTextEnabled,
		spanIdEnabled:         spanIdEnabled,
		traceIdEnabled:        traceIdEnabled,
	}, nil
}

func (proc *logFieldsConversionProcessor) addAttributes(log plog.LogRecord) {
	if log.SeverityNumber() != plog.SeverityNumberUnspecified {
		if _, found := log.Attributes().Get(SeverityNumberAttributeName); !found && proc.severityNumberEnabled.Enabled {
			level := severityNumberToLevel[log.SeverityNumber().String()]
			log.Attributes().PutStr(proc.severityNumberEnabled.AttributeName, level)
		}
	}
	if _, found := log.Attributes().Get(SeverityTextAttributeName); !found && proc.severityTextEnabled.Enabled {
		log.Attributes().PutStr(proc.severityTextEnabled.AttributeName, log.SeverityText())
	}
	if _, found := log.Attributes().Get(SpanIdAttributeName); !found && proc.spanIdEnabled.Enabled {
		log.Attributes().PutStr(proc.spanIdEnabled.AttributeName, SpanIDToHexOrEmptyString(log.SpanID()))
	}
	if _, found := log.Attributes().Get(TraceIdAttributeName); !found && proc.traceIdEnabled.Enabled {
		log.Attributes().PutStr(proc.traceIdEnabled.AttributeName, TraceIDToHexOrEmptyString(log.TraceID()))
	}
}

func (proc *logFieldsConversionProcessor) processLogs(logs plog.Logs) error {
	if proc.isEnabled() {
		rls := logs.ResourceLogs()
		for i := 0; i < rls.Len(); i++ {
			ills := rls.At(i).ScopeLogs()

			for j := 0; j < ills.Len(); j++ {
				logs := ills.At(j).LogRecords()
				for k := 0; k < logs.Len(); k++ {
					proc.addAttributes(logs.At(k))
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
	return proc.severityNumberEnabled.Enabled ||
		proc.severityTextEnabled.Enabled ||
		proc.spanIdEnabled.Enabled ||
		proc.traceIdEnabled.Enabled
}

func (*logFieldsConversionProcessor) ConfigPropertyName() string {
	return "add_severity_level_attribute"
}
