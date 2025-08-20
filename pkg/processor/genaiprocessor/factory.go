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
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/processor"
	"go.opentelemetry.io/collector/processor/processorhelper"
)

const (
	// The value of "type" key in configuration.
	typeStr = "genai"

	// Default configuration values
	defaultEndpoint      = "http://localhost:4000"
	defaultModel         = "gpt-3.5-turbo"
	defaultSystemPrompt  = "You are a helpful assistant that analyzes log data."
	defaultUserPrompt    = "Analyze this log entry and provide insights: {{.body}}"
	defaultMaxTokens     = 150
	defaultTemperature   = 0.3
	defaultTimeout       = 30 * time.Second
	defaultResponseField = "ai_analysis"

	stabilityLevel = component.StabilityLevelDevelopment
)

var Type = component.MustNewType(typeStr)

var processorCapabilities = consumer.Capabilities{MutatesData: true}

// NewFactory returns a new factory for the GenAI processor.
func NewFactory() processor.Factory {
	return processor.NewFactory(
		Type,
		createDefaultConfig,
		processor.WithLogs(createLogsProcessor, stabilityLevel),
	)
}

// createDefaultConfig creates the default configuration for processor.
func createDefaultConfig() component.Config {
	return &Config{
		Endpoint:      defaultEndpoint,
		Model:         defaultModel,
		SystemPrompt:  defaultSystemPrompt,
		UserPrompt:    defaultUserPrompt,
		MaxTokens:     defaultMaxTokens,
		Temperature:   defaultTemperature,
		Timeout:       defaultTimeout,
		ResponseField: defaultResponseField,
		ExtractFields: []string{"body"},
	}
}

// createLogsProcessor creates a logs processor based on this config
func createLogsProcessor(
	ctx context.Context,
	params processor.Settings,
	cfg component.Config,
	next consumer.Logs,
) (processor.Logs, error) {
	oCfg := cfg.(*Config)

	processor, err := newGenAIProcessor(params, oCfg)
	if err != nil {
		return nil, err
	}

	return processorhelper.NewLogs(
		ctx,
		params,
		cfg,
		next,
		processor.ProcessLogs,
		processorhelper.WithCapabilities(processorCapabilities),
	)
}
