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
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/extension/extensionhelper"
)

var defaultEnsureOpAMPVal = true

const (
	// The value of extension "type" in configuration.
	typeStr               = "opamp"
	DefaultServerEndpoint = "ws://localhost:4320/v1/opamp"
	DefaultConfigPath     = "config-opamp.yaml"
)

// NewFactory creates a factory for Sumo Logic extension.
func NewFactory() component.ExtensionFactory {
	return extensionhelper.NewFactory(
		typeStr,
		createDefaultConfig,
		createExtension,
	)
}

func createDefaultConfig() config.Extension {
	return &Config{
		ExtensionSettings: config.NewExtensionSettings(config.NewComponentID(typeStr)),
		ServerEndpoint:    DefaultServerEndpoint,
		ConfigPath:        DefaultConfigPath,
		EnsureOpAMP:       &defaultEnsureOpAMPVal,
	}
}

func createExtension(_ context.Context, params component.ExtensionCreateSettings, cfg config.Extension) (component.Extension, error) {
	config := cfg.(*Config)

	return newOpAMPExtension(config, params.Logger, params.BuildInfo.Description, params.BuildInfo.Version)
}
