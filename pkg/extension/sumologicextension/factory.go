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

package sumologicextension

import (
	"context"

	"github.com/cenkalti/backoff/v4"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config"

	"github.com/SumoLogic/sumologic-otel-collector/pkg/extension/sumologicextension/credentials"
)

const (
	// The value of extension "type" in configuration.
	typeStr           = "sumologic"
	DefaultApiBaseUrl = "https://open-collectors.sumologic.com"
)

// NewFactory creates a factory for Sumo Logic extension.
func NewFactory() component.ExtensionFactory {
	return component.NewExtensionFactory(
		typeStr,
		createDefaultConfig,
		createExtension,
		component.StabilityLevelBeta,
	)
}

func createDefaultConfig() config.Extension {
	defaultCredsPath, err := credentials.GetDefaultCollectorCredentialsDirectory()
	if err != nil {
		return nil
	}

	return &Config{
		ExtensionSettings:             config.NewExtensionSettings(config.NewComponentID(typeStr)),
		ApiBaseUrl:                    DefaultApiBaseUrl,
		HeartBeatInterval:             DefaultHeartbeatInterval,
		CollectorCredentialsDirectory: defaultCredsPath,
		Clobber:                       false,
		ForceRegistration:             false,
		Ephemeral:                     false,
		TimeZone:                      "",
		BackOff: backOffConfig{
			InitialInterval: backoff.DefaultInitialInterval,
			MaxInterval:     backoff.DefaultMaxInterval,
			MaxElapsedTime:  backoff.DefaultMaxElapsedTime,
		},
	}
}

func createExtension(_ context.Context, params component.ExtensionCreateSettings, cfg config.Extension) (component.Extension, error) {
	config := cfg.(*Config)
	return newSumologicExtension(config, params.Logger)
}
