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
	"go.opentelemetry.io/collector/extension"
)

const (
	// The value of extension "type" in configuration.
	typeStr = "opamp"
)

var Type = component.MustNewType(typeStr)

func NewFactory() extension.Factory {
	return extension.NewFactory(Type, createDefaultConfig, createExtension, component.StabilityLevelAlpha)
}

func createDefaultConfig() component.Config {
	return &Config{
		ClientConfig:               CreateDefaultClientConfig(),
		AcceptsRemoteConfiguration: true,
	}
}

func createExtension(_ context.Context, set extension.Settings, cfg component.Config) (extension.Extension, error) {
	return newOpampAgent(cfg.(*Config), set.Logger, set.BuildInfo, set.Resource)
}
