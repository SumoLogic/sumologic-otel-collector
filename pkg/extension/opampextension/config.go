// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package opampextension

import (
	"errors"

	"github.com/oklog/ulid/v2"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configauth"
	"go.opentelemetry.io/collector/config/confighttp"

	"github.com/open-telemetry/opentelemetry-collector-contrib/extension/sumologicextension"
)

// Config has the configuration for the opamp extension.
type Config struct {
	confighttp.ClientConfig `mapstructure:",squash"`

	// InstanceUID is a ULID formatted as a 26 character string in canonical
	// representation. Auto-generated on start if missing.
	InstanceUID string `mapstructure:"instance_uid"`

	// RemoteConfigurationDirectory is where received OpAMP remote configuration
	// is stored.
	RemoteConfigurationDirectory string `mapstructure:"remote_configuration_directory"`

	// AcceptsRemoteConfiguration indicates if the OpAMP agent will accept remote configuration.
	AcceptsRemoteConfiguration bool `mapstructure:"accepts_remote_configuration"`

	// Flag to toggle new config merge flow introduced for collector tag edit feature
	NewConfigMergeFlowDisabled bool `mapstructure:"new_configmergeflow_disabled"`
}

// CreateDefaultClientConfig returns default http client settings
func CreateDefaultClientConfig() confighttp.ClientConfig {
	return confighttp.ClientConfig{
		Auth: &configauth.Authentication{
			AuthenticatorID: component.NewID(sumologicextension.NewFactory().Type()),
		},
	}
}

// Validate checks if the extension configuration is valid
func (cfg *Config) Validate() error {
	if cfg.InstanceUID != "" {
		_, err := ulid.ParseStrict(cfg.InstanceUID)
		if err != nil {
			return errors.New("opamp instance_uid is invalid")
		}
	}

	if cfg.RemoteConfigurationDirectory == "" {
		return errors.New("opamp remote_configuration_directory must be provided")
	}

	return nil
}
