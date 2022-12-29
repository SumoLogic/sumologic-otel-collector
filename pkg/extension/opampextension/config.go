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
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/config/configauth"
	"go.opentelemetry.io/collector/config/confighttp"
)

// Config has the configuration for the opamp extension.
type Config struct {
	config.ExtensionSettings      `mapstructure:",squash"` // squash ensures fields are correctly decoded in embedded struct
	confighttp.HTTPClientSettings `mapstructure:",squash"`

	// Endpoint is the OpAMP server URL. Transport based on the scheme of the URL.
	Endpoint string `mapstructure:"endpoint"`

	// InstanceUID is a ULID formatted as a 26 character string in canonical
	// representation. Auto-generated on start if missing.
	InstanceUID string `mapstructure:"instance_uid"`
}

// CreateDefaultHTTPClientSettings returns default http client settings
func CreateDefaultHTTPClientSettings() confighttp.HTTPClientSettings {
	return confighttp.HTTPClientSettings{
		Auth: &configauth.Authentication{
			AuthenticatorID: component.NewID("sumologic"),
		},
	}
}

// Validate checks if the extension configuration is valid
func (cfg *Config) Validate() error {
	_, err := ulid.ParseStrict(cfg.InstanceUID)
	if err != nil {
		return errors.New("opamp instance_uid is invalid")
	}
	return nil
}
