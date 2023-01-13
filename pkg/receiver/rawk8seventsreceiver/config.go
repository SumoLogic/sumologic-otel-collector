// Copyright 2021, OpenTelemetry Authors
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

package rawk8seventsreceiver

import (
	"time"
)

// Config defines configuration for the receiver.
type Config struct {
	APIConfig `mapstructure:",squash"`
	// List of ‘namespaces’ to collect events from.
	// Empty list means all namespaces
	Namespaces []string `mapstructure:"namespaces"`

	// Maximum age of event relative to receiver start time
	// Events older than StartTime - MaxEventAge will not be collected
	MaxEventAge time.Duration `mapstructure:"max_event_age"`

	// ConsumeRetryDelay is the retry delay for recoverable pipeline errors
	// one frequent source of these kinds of errors is the memory_limiter processor
	ConsumeRetryDelay time.Duration `mapstructure:"consume_retry_delay"`

	// ConsumeMaxRetries is the maximum number of retries for recoverable pipeline errors
	ConsumeMaxRetries uint64 `mapstructure:"consume_max_retries"`
}

// Validate checks if the receiver configuration is valid
func (cfg *Config) Validate() error {
	return cfg.APIConfig.Validate()
}
