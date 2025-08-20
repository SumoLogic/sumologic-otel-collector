// Copyright 2024 SumoLogic, Inc.
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

package genaiprocessor

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/processor"
	"go.uber.org/zap"
)

func newProcessorCreateSettings() processor.Settings {
	return processor.Settings{
		TelemetrySettings: component.TelemetrySettings{
			Logger: zap.NewNop(),
		},
	}
}

func TestNewGenAIProcessor(t *testing.T) {
	config := &Config{
		Endpoint:      "http://localhost:4000",
		Model:         "gpt-3.5-turbo",
		SystemPrompt:  "You are a helpful assistant.",
		UserPrompt:    "Analyze: {{.body}}",
		MaxTokens:     100,
		Temperature:   0.3,
		Timeout:       30 * time.Second,
		ResponseField: "ai_analysis",
		ExtractFields: []string{"body"},
	}

	params := newProcessorCreateSettings()

	processor, err := newGenAIProcessor(params, config)
	require.NoError(t, err)
	assert.NotNil(t, processor)
	assert.Equal(t, config, processor.config)
}

func TestNewGenAIProcessorWithInvalidRegex(t *testing.T) {
	config := &Config{
		Endpoint:      "http://localhost:4000",
		Model:         "gpt-3.5-turbo",
		SystemPrompt:  "You are a helpful assistant.",
		UserPrompt:    "Analyze: {{.body}}",
		MaxTokens:     100,
		Temperature:   0.3,
		Timeout:       30 * time.Second,
		ResponseField: "ai_analysis",
		FilterRegex:   "[invalid",
		ExtractFields: []string{"body"},
	}

	params := newProcessorCreateSettings()

	_, err := newGenAIProcessor(params, config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid filter regex")
}

func TestProcessLogs(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := LiteLLMResponse{
			Choices: []Choice{
				{
					Message: Message{
						Role:    "assistant",
						Content: "This log indicates a normal operation.",
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := &Config{
		Endpoint:      server.URL,
		Model:         "gpt-3.5-turbo",
		SystemPrompt:  "You are a helpful assistant.",
		UserPrompt:    "Analyze: {{.body}}",
		MaxTokens:     100,
		Temperature:   0.3,
		Timeout:       30 * time.Second,
		ResponseField: "ai_analysis",
		ExtractFields: []string{"body"},
	}

	params := newProcessorCreateSettings()

	processor, err := newGenAIProcessor(params, config)
	require.NoError(t, err)

	// Create test logs
	logs := plog.NewLogs()
	resourceLog := logs.ResourceLogs().AppendEmpty()
	scopeLog := resourceLog.ScopeLogs().AppendEmpty()
	logRecord := scopeLog.LogRecords().AppendEmpty()
	logRecord.Body().SetStr("This is a test log message")
	logRecord.SetSeverityText("INFO")

	// Process logs
	processedLogs, err := processor.ProcessLogs(context.Background(), logs)
	require.NoError(t, err)

	// Verify the AI response was added
	processedRecord := processedLogs.ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(0)
	aiResponse, exists := processedRecord.Attributes().Get("ai_analysis")
	assert.True(t, exists)
	assert.Equal(t, "This log indicates a normal operation.", aiResponse.AsString())
}

func TestProcessLogsWithFilter(t *testing.T) {
	config := &Config{
		Endpoint:      "http://localhost:4000",
		Model:         "gpt-3.5-turbo",
		SystemPrompt:  "You are a helpful assistant.",
		UserPrompt:    "Analyze: {{.body}}",
		MaxTokens:     100,
		Temperature:   0.3,
		Timeout:       30 * time.Second,
		ResponseField: "ai_analysis",
		FilterRegex:   "ERROR",
		ExtractFields: []string{"body"},
	}

	params := newProcessorCreateSettings()

	processor, err := newGenAIProcessor(params, config)
	require.NoError(t, err)

	// Create test logs that don't match the filter
	logs := plog.NewLogs()
	resourceLog := logs.ResourceLogs().AppendEmpty()
	scopeLog := resourceLog.ScopeLogs().AppendEmpty()
	logRecord := scopeLog.LogRecords().AppendEmpty()
	logRecord.Body().SetStr("This is an INFO log message")
	logRecord.SetSeverityText("INFO")

	// Process logs
	processedLogs, err := processor.ProcessLogs(context.Background(), logs)
	require.NoError(t, err)

	// Verify the AI response was NOT added (filtered out)
	processedRecord := processedLogs.ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(0)
	_, exists := processedRecord.Attributes().Get("ai_analysis")
	assert.False(t, exists)
}

func TestExtractTemplateData(t *testing.T) {
	config := &Config{
		ExtractFields: []string{"body", "custom_field"},
	}

	params := newProcessorCreateSettings()

	processor, err := newGenAIProcessor(params, config)
	require.NoError(t, err)

	// Create a log record with attributes
	logRecord := plog.NewLogRecord()
	logRecord.Body().SetStr("Test log message")
	logRecord.SetSeverityText("ERROR")
	logRecord.Attributes().PutStr("custom_field", "custom_value")
	logRecord.Attributes().PutStr("other_field", "other_value")

	data := processor.extractTemplateData(logRecord)

	assert.Equal(t, "Test log message", data["body"])
	assert.Equal(t, "ERROR", data["severity"])
	assert.Contains(t, data, "timestamp")
	assert.Equal(t, "custom_value", data["custom_field"])
	assert.NotContains(t, data, "other_field") // Not in extract_fields
}

func TestRenderTemplate(t *testing.T) {
	config := &Config{
		UserPrompt: "Log: {{.body}}, Severity: {{.severity}}",
	}

	params := newProcessorCreateSettings()

	processor, err := newGenAIProcessor(params, config)
	require.NoError(t, err)

	data := map[string]interface{}{
		"body":     "Test message",
		"severity": "ERROR",
	}

	result, err := processor.renderTemplate(processor.userTemplate, data)
	require.NoError(t, err)
	assert.Equal(t, "Log: Test message, Severity: ERROR", result)
}
