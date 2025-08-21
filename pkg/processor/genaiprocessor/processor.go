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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"text/template"

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/processor"
	"go.uber.org/zap"
)

// LiteLLMRequest represents the request structure for LiteLLM API
type LiteLLMRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	MaxTokens   int       `json:"max_tokens"`
	Temperature float64   `json:"temperature"`
}

// Message represents a chat message
type Message struct {
	Role        string        `json:"role"`
	Content     string        `json:"content"`
	Annotations []interface{} `json:"annotations,omitempty"`
}

// LiteLLMResponse represents the response structure from LiteLLM API
type LiteLLMResponse struct {
	ID                string    `json:"id"`
	Created           int64     `json:"created"`
	Model             string    `json:"model"`
	Object            string    `json:"object"`
	SystemFingerprint string    `json:"system_fingerprint"`
	Choices           []Choice  `json:"choices"`
	Usage             *Usage    `json:"usage,omitempty"`
	ServiceTier       string    `json:"service_tier,omitempty"`
	Error             *APIError `json:"error,omitempty"`
}

// Usage represents token usage information
type Usage struct {
	CompletionTokens int `json:"completion_tokens"`
	PromptTokens     int `json:"prompt_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// Choice represents a response choice
type Choice struct {
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
	Index        int     `json:"index"`
}

// APIError represents an API error
type APIError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code"`
}

// genaiProcessor processes logs using GenAI models via LiteLLM
type genaiProcessor struct {
	config         *Config
	logger         *zap.Logger
	httpClient     *http.Client
	filterRegex    *regexp.Regexp
	userTemplate   *template.Template
	systemTemplate *template.Template
}

// newGenAIProcessor creates a new GenAI processor
func newGenAIProcessor(params processor.Settings, config *Config) (*genaiProcessor, error) {
	httpClient := &http.Client{
		Timeout: config.Timeout,
	}

	var filterRegex *regexp.Regexp
	var err error
	if config.FilterRegex != "" {
		filterRegex, err = regexp.Compile(config.FilterRegex)
		if err != nil {
			return nil, fmt.Errorf("invalid filter regex: %w", err)
		}
	}

	// Parse user prompt template
	userTemplate, err := template.New("user").Parse(config.UserPrompt)
	if err != nil {
		return nil, fmt.Errorf("invalid user prompt template: %w", err)
	}

	// Parse system prompt template
	systemTemplate, err := template.New("system").Parse(config.SystemPrompt)
	if err != nil {
		return nil, fmt.Errorf("invalid system prompt template: %w", err)
	}

	return &genaiProcessor{
		config:         config,
		logger:         params.Logger,
		httpClient:     httpClient,
		filterRegex:    filterRegex,
		userTemplate:   userTemplate,
		systemTemplate: systemTemplate,
	}, nil
}

// ProcessLogs processes log records through GenAI models
func (p *genaiProcessor) ProcessLogs(ctx context.Context, logs plog.Logs) (plog.Logs, error) {
	resourceLogs := logs.ResourceLogs()

	// Build new collection with processed records
	for i := 0; i < resourceLogs.Len(); i++ {
		resourceLog := resourceLogs.At(i)
		scopeLogs := resourceLog.ScopeLogs()

		for j := 0; j < scopeLogs.Len(); j++ {
			scopeLog := scopeLogs.At(j)
			logRecords := scopeLog.LogRecords()

			// Collect all records to keep (non-array) and expanded records (from arrays)
			var finalRecords []plog.LogRecord

			for k := 0; k < logRecords.Len(); k++ {
				record := logRecords.At(k)

				// Check if this record should be processed (contains JSON array)
				if p.shouldProcessRecord(record) {
					expandedRecords, err := p.expandJsonArrayRecord(ctx, record)
					if err != nil {
						p.logger.Error("Failed to expand JSON array record", zap.Error(err))
						// Keep original record on error
						finalRecords = append(finalRecords, record)
						continue
					}

					if len(expandedRecords) > 0 {
						// Add expanded records instead of original
						finalRecords = append(finalRecords, expandedRecords...)
					} else {
						// Keep original if no expansion occurred
						finalRecords = append(finalRecords, record)
					}
				} else {
					// Keep non-array records as-is
					finalRecords = append(finalRecords, record)
				}
			}

			// Clear existing records and add final records
			logRecords.RemoveIf(func(lr plog.LogRecord) bool {
				return true // Remove all
			})

			// Add all final records
			for _, finalRecord := range finalRecords {
				targetRecord := logRecords.AppendEmpty()
				finalRecord.CopyTo(targetRecord)
			}
		}
	}
	return logs, nil
}

// extractTemplateData extracts data from log record for template rendering
func (p *genaiProcessor) extractTemplateData(record plog.LogRecord) map[string]interface{} {
	data := make(map[string]interface{})

	// Always include basic log data
	data["body"] = record.Body().AsString()
	data["timestamp"] = record.Timestamp().String()
	data["severity"] = record.SeverityText()

	// Extract specified fields from attributes
	record.Attributes().Range(func(key string, value pcommon.Value) bool {
		for _, field := range p.config.ExtractFields {
			if key == field {
				data[key] = value.AsString()
				break
			}
		}
		return true
	})

	return data
}

// renderTemplate renders a template with the given data
func (p *genaiProcessor) renderTemplate(tmpl *template.Template, data map[string]interface{}) (string, error) {
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// callLiteLLM makes a request to the LiteLLM API
func (p *genaiProcessor) callLiteLLM(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	// Prepare the request payload
	request := LiteLLMRequest{
		Model: p.config.Model,
		Messages: []Message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		MaxTokens:   p.config.MaxTokens,
		Temperature: p.config.Temperature,
	}

	// Marshal the request to JSON
	requestBody, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.config.Endpoint+"/chat/completions", bytes.NewBuffer(requestBody))
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	if p.config.APIKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+p.config.APIKey)
	}

	// Make the request
	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to make HTTP request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	// Check for HTTP errors
	if resp.StatusCode != http.StatusOK {
		responsePreview := string(responseBody)
		if len(responsePreview) > 1000 {
			responsePreview = responsePreview[:1000] + "..."
		}
		p.logger.Error("LiteLLM API request failed",
			zap.Int("status_code", resp.StatusCode),
			zap.String("response", responsePreview))
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(responseBody))
	}

	// Always log basic response info for debugging
	p.logger.Info("Received response from LiteLLM API",
		zap.Int("status_code", resp.StatusCode),
		zap.Int("response_length", len(responseBody)))

	// Log response body for debugging if it's not JSON
	if !json.Valid(responseBody) {
		p.logger.Error("Received non-JSON response from API",
			zap.String("full_response", string(responseBody)))
		return "", fmt.Errorf("received non-JSON response from API (likely HTML redirect)")
	}

	// Parse the response
	var llmResponse LiteLLMResponse
	if err := json.Unmarshal(responseBody, &llmResponse); err != nil {
		responsePreview := string(responseBody)
		if len(responsePreview) > 200 {
			responsePreview = responsePreview[:200] + "..."
		}
		p.logger.Error("Failed to unmarshal response", zap.String("response", responsePreview))
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Check for API errors
	if llmResponse.Error != nil {
		return "", fmt.Errorf("API error: %s", llmResponse.Error.Message)
	}

	// Extract the response content
	if len(llmResponse.Choices) == 0 {
		return "", fmt.Errorf("no choices in API response")
	}

	responseContent := strings.TrimSpace(llmResponse.Choices[0].Message.Content)

	// Log successful parsing
	p.logger.Info("Successfully parsed LiteLLM response",
		zap.String("model", llmResponse.Model),
		zap.Int("total_tokens", func() int {
			if llmResponse.Usage != nil {
				return llmResponse.Usage.TotalTokens
			}
			return 0
		}()),
		zap.Int("response_length", len(responseContent)))

	return responseContent, nil
}

// shouldProcessRecord checks if a log record contains a JSON array that should be split
func (p *genaiProcessor) shouldProcessRecord(record plog.LogRecord) bool {
	// Check if the log record should be processed based on filter regex
	if p.filterRegex != nil {
		logBody := record.Body().AsString()
		if !p.filterRegex.MatchString(logBody) {
			return false // Skip this record
		}
	}

	// Check if the body looks like a JSON array
	logBody := strings.TrimSpace(record.Body().AsString())
	return strings.HasPrefix(logBody, "[") && strings.HasSuffix(logBody, "]")
}

// expandJsonArrayRecord processes a single record containing a JSON array and returns multiple individual records
func (p *genaiProcessor) expandJsonArrayRecord(ctx context.Context, originalRecord plog.LogRecord) ([]plog.LogRecord, error) {
	// Extract data from log record for template
	templateData := p.extractTemplateData(originalRecord)

	// Generate prompts using templates
	systemPrompt, err := p.renderTemplate(p.systemTemplate, templateData)
	if err != nil {
		return nil, fmt.Errorf("failed to render system prompt: %w", err)
	}

	userPrompt, err := p.renderTemplate(p.userTemplate, templateData)
	if err != nil {
		return nil, fmt.Errorf("failed to render user prompt: %w", err)
	}

	// Call LiteLLM API
	response, err := p.callLiteLLM(ctx, systemPrompt, userPrompt)
	if err != nil {
		// Log the error but don't fail the entire pipeline
		p.logger.Warn("Skipping GenAI processing due to API error", zap.Error(err))
		return nil, nil // Return empty slice, not an error
	}

	// Split the response into individual JSON lines
	lines := strings.Split(strings.TrimSpace(response), "\n")
	var expandedRecords []plog.LogRecord

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Validate that the line is valid JSON
		if !json.Valid([]byte(line)) {
			p.logger.Warn("Skipping invalid JSON line from AI response", zap.String("line", line))
			continue
		}

		// Create a new log record for each JSON object
		newRecord := plog.NewLogRecord()
		originalRecord.CopyTo(newRecord)

		// Set the body to the individual JSON object
		newRecord.Body().SetStr(line)

		// Add metadata about the processing
		newRecord.Attributes().PutStr(p.config.ResponseField+"_status", "split_from_array")
		newRecord.Attributes().PutStr("processing_type", "json_array_split")

		expandedRecords = append(expandedRecords, newRecord)
	}

	p.logger.Info("Successfully split JSON array into individual records",
		zap.Int("original_records", 1),
		zap.Int("expanded_records", len(expandedRecords)))

	return expandedRecords, nil
}
