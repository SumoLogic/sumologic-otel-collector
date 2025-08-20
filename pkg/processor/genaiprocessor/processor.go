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
	Role    string `json:"role"`
	Content string `json:"content"`
}

// LiteLLMResponse represents the response structure from LiteLLM API
type LiteLLMResponse struct {
	Choices []Choice `json:"choices"`
	Error   *APIError `json:"error,omitempty"`
}

// Choice represents a response choice
type Choice struct {
	Message Message `json:"message"`
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
	for i := 0; i < resourceLogs.Len(); i++ {
		resource := resourceLogs.At(i)
		scopeLogs := resource.ScopeLogs()
		for j := 0; j < scopeLogs.Len(); j++ {
			scope := scopeLogs.At(j)
			logRecords := scope.LogRecords()
			for k := 0; k < logRecords.Len(); k++ {
				record := logRecords.At(k)
				if err := p.processLogRecord(ctx, record); err != nil {
					p.logger.Error("Failed to process log record with GenAI", zap.Error(err))
					// Continue processing other records even if one fails
				}
			}
		}
	}
	return logs, nil
}

// processLogRecord processes a single log record
func (p *genaiProcessor) processLogRecord(ctx context.Context, record plog.LogRecord) error {
	// Check if the log record should be processed based on filter regex
	if p.filterRegex != nil {
		logBody := record.Body().AsString()
		if !p.filterRegex.MatchString(logBody) {
			return nil // Skip this record
		}
	}

	// Extract data from log record for template
	templateData := p.extractTemplateData(record)

	// Generate prompts using templates
	systemPrompt, err := p.renderTemplate(p.systemTemplate, templateData)
	if err != nil {
		return fmt.Errorf("failed to render system prompt: %w", err)
	}

	userPrompt, err := p.renderTemplate(p.userTemplate, templateData)
	if err != nil {
		return fmt.Errorf("failed to render user prompt: %w", err)
	}

	// Call LiteLLM API
	response, err := p.callLiteLLM(ctx, systemPrompt, userPrompt)
	if err != nil {
		return fmt.Errorf("failed to call LiteLLM API: %w", err)
	}

	// Add the AI response to the log record attributes
	record.Attributes().PutStr(p.config.ResponseField, response)

	return nil
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
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(responseBody))
	}

	// Parse the response
	var llmResponse LiteLLMResponse
	if err := json.Unmarshal(responseBody, &llmResponse); err != nil {
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

	return strings.TrimSpace(llmResponse.Choices[0].Message.Content), nil
}
