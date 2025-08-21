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
	"time"
)

// Config defines configuration for GenAI processor.
type Config struct {
	// LiteLLM endpoint URL
	Endpoint string `mapstructure:"endpoint"`

	// API key for authentication (optional)
	APIKey string `mapstructure:"api_key"`

	// Model to use for processing
	Model string `mapstructure:"model"`

	// System prompt template
	SystemPrompt string `mapstructure:"system_prompt"`

	// User prompt template (can use placeholders for log data)
	UserPrompt string `mapstructure:"user_prompt"`

	// Maximum tokens for the response
	MaxTokens int `mapstructure:"max_tokens"`

	// Temperature for response generation
	Temperature float64 `mapstructure:"temperature"`

	// Timeout for API requests
	Timeout time.Duration `mapstructure:"timeout"`

	// Field to store the AI response in the log record
	ResponseField string `mapstructure:"response_field"`

	// Only process logs that match this regex (optional)
	FilterRegex string `mapstructure:"filter_regex"`

	// Fields to extract from log records for processing
	ExtractFields []string `mapstructure:"extract_fields"`
}
