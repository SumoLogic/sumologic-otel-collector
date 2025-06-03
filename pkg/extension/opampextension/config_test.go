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
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configauth"
	"go.opentelemetry.io/collector/config/confighttp"
	"go.opentelemetry.io/collector/confmap"
	"go.opentelemetry.io/collector/confmap/confmaptest"

	"github.com/open-telemetry/opentelemetry-collector-contrib/extension/sumologicextension"
)

const (
	errOpampDirMust = "opamp remote_configuration_directory must be provided"
	errOpampInvalidInstanceUuid = "opamp instance_uid is invalid"
)

func TestUnmarshalDefaultConfig(t *testing.T) {
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig()
	assert.NoError(t, confmap.New().Unmarshal(&cfg))
	assert.Equal(t, factory.CreateDefaultConfig(), cfg)
}

func TestUnmarshalConfig(t *testing.T) {
	cm, err := confmaptest.LoadConf(filepath.Join("testdata", "config.yaml"))
	require.NoError(t, err)
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig()
	assert.NoError(t, cm.Unmarshal(&cfg))
	assert.Equal(t,
		&Config{
			ClientConfig: confighttp.ClientConfig{
				Endpoint: "wss://127.0.0.1:4320/v1/opamp",
				Auth: &configauth.Authentication{
					AuthenticatorID: component.NewID(sumologicextension.NewFactory().Type()),
				},
			},
			InstanceUID:                  "01BX5ZZKBKACTAV9WEVGEMMVRZ",
			RemoteConfigurationDirectory: "/tmp/opamp.d",
			AcceptsRemoteConfiguration:   true,
		}, cfg)
}

func TestConfigValidate(t *testing.T) {
	cfg := &Config{}
	err := cfg.Validate()
	require.Error(t, err)
	assert.Equal(t, errOpampDirMust, err.Error())

	d, err := os.MkdirTemp("", "opamp.d")
	assert.NoError(t, err)
	defer os.RemoveAll(d)

	cfg.RemoteConfigurationDirectory = d
	err = cfg.Validate()
	assert.NoError(t, err)

	cfg.InstanceUID = "01BX5ZZKBKACTAV9WEVGEMMVRZFAIL"
	err = cfg.Validate()
	require.Error(t, err)
	assert.Equal(t, errOpampInvalidInstanceUuid, err.Error())
}
