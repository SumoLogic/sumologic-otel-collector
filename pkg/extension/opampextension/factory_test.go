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
	"testing"

	"go.opentelemetry.io/collector/config/configoptional"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/config/configauth"
	"go.opentelemetry.io/collector/config/confighttp"
	"go.opentelemetry.io/collector/extension/extensiontest"

	"github.com/open-telemetry/opentelemetry-collector-contrib/extension/sumologicextension"
)

func TestFactory_CreateDefaultConfig(t *testing.T) {
	f := NewFactory()
	cfg := createDefaultConfig()

	assert.Equal(t, cfg, &Config{
		ClientConfig: confighttp.ClientConfig{
			Auth: configoptional.Some[configauth.Config](configauth.Config{
				AuthenticatorID: component.NewID(sumologicextension.NewFactory().Type()),
			}),
		},
		AcceptsRemoteConfiguration: true,
	})

	assert.NoError(t, componenttest.CheckConfigStruct(cfg))
	ext, err := createExtension(context.Background(), extensiontest.NewNopSettings(f.Type()), cfg)
	require.NoError(t, err)
	require.NotNil(t, ext)
}

func TestFactory_CreateExtension(t *testing.T) {
	f := NewFactory()
	cfg := createDefaultConfig().(*Config)
	ext, err := createExtension(context.Background(), extensiontest.NewNopSettings(f.Type()), cfg)
	require.NoError(t, err)
	require.NotNil(t, ext)
}
