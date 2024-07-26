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
	// "context"
	"os"
	"path/filepath"
	"testing"

	"github.com/oklog/ulid/v2"
	"github.com/open-telemetry/opamp-go/protobufs"
	"github.com/stretchr/testify/assert"

	"go.opentelemetry.io/collector/component"
	// "go.opentelemetry.io/collector/component/componenttest"
	// "go.opentelemetry.io/collector/config/confighttp"
	"go.opentelemetry.io/collector/extension/extensiontest"
	"go.opentelemetry.io/collector/extension"

	semconv "go.opentelemetry.io/collector/semconv/v1.18.0"
)

func remoteConfig(t *testing.T, path string) *protobufs.AgentRemoteConfig {
	rb, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	rc := &protobufs.AgentRemoteConfig{
		Config: &protobufs.AgentConfigMap{
			ConfigMap: map[string]*protobufs.AgentConfigFile{
				"default": {
					Body: rb,
				},
			},
		},
		ConfigHash: []byte("b2b1e3e7f45d564db1c0b621bbf67008"),
	}
	return rc
}

func TestOpampAgent(t *testing.T) {
	testCases := []struct {
		name      string
		setup     func() (*Config, extension.Settings)
		validate  func(*opampAgent, *testing.T)
	}{
		{
			name: "NewOpampAgent",
			setup: func() (*Config, extension.Settings) {
				cfg := createDefaultConfig()
				set := extensiontest.NewNopSettings()
				set.BuildInfo = component.BuildInfo{Version: "test version", Command: "otelcoltest"}
				return cfg.(*Config), set
			},
			validate: func(o *opampAgent, t *testing.T) {
				assert.Equal(t, "otelcoltest", o.agentType)
				assert.Equal(t, "test version", o.agentVersion)
				assert.NotEmpty(t, o.instanceId.String())
				assert.Empty(t, o.effectiveConfig)
				assert.Nil(t, o.agentDescription)
			},
		},
		{
			name: "NewOpampAgentAttributes",
			setup: func() (*Config, extension.Settings) {
				cfg := createDefaultConfig()
				set := extensiontest.NewNopSettings()
				set.BuildInfo = component.BuildInfo{Version: "test version", Command: "otelcoltest"}
				set.Resource.Attributes().PutStr(semconv.AttributeServiceName, "otelcol-sumo")
				set.Resource.Attributes().PutStr(semconv.AttributeServiceVersion, "sumo.0")
				set.Resource.Attributes().PutStr(semconv.AttributeServiceInstanceID, "f8999bc1-4c9b-4619-9bae-7f009d2411ec")
				return cfg.(*Config), set
			},
			validate: func(o *opampAgent, t *testing.T) {
				assert.Equal(t, "otelcol-sumo", o.agentType)
				assert.Equal(t, "sumo.0", o.agentVersion)
				assert.Equal(t, "7RK6DW2K4V8RCSQBKZ02EJ84FC", o.instanceId.String())
			},
		},
		{
			name: "GetAgentCapabilities",
			setup: func() (*Config, extension.Settings) {
				cfg := createDefaultConfig().(*Config)
				set := extensiontest.NewNopSettings()
				return cfg, set
			},
			validate: func(o *opampAgent, t *testing.T) {
				assert.Equal(t, o.getAgentCapabilities(), protobufs.AgentCapabilities(4102))
				
				o.cfg.AcceptsRemoteConfiguration = false
				assert.Equal(t, o.getAgentCapabilities(), protobufs.AgentCapabilities(4))
			},
		},
		{
			name: "CreateAgentDescription",
			setup: func() (*Config, extension.Settings) {
				cfg := createDefaultConfig()
				set := extensiontest.NewNopSettings()
				return cfg.(*Config), set
			},
			validate: func(o *opampAgent, t *testing.T) {
				assert.Nil(t, o.agentDescription)
				assert.NoError(t, o.createAgentDescription())
				assert.NotNil(t, o.agentDescription)
			},
		},
		{
			name: "LoadEffectiveConfig",
			setup: func() (*Config, extension.Settings) {
				cfg := createDefaultConfig()
				set := extensiontest.NewNopSettings()
				return cfg.(*Config), set
			},
			validate: func(o *opampAgent, t *testing.T) {
				assert.Equal(t, len(o.effectiveConfig), 0)

				assert.NoError(t, o.loadEffectiveConfig("testdata"))
				assert.NotEqual(t, len(o.effectiveConfig), 0)
			},
		},
		{
			name: "SaveEffectiveConfig",
			setup: func() (*Config, extension.Settings) {
				cfg := createDefaultConfig()
				set := extensiontest.NewNopSettings()
				return cfg.(*Config), set
			},
			validate: func(o *opampAgent, t *testing.T) {
				d, err := os.MkdirTemp("", "opamp.d")
				assert.NoError(t, err)
				defer os.RemoveAll(d)

				assert.NoError(t, o.saveEffectiveConfig(d))
			},
		},
		{
			name: "UpdateAgentIdentity",
			setup: func() (*Config, extension.Settings) {
				cfg := createDefaultConfig()
				set := extensiontest.NewNopSettings()
				return cfg.(*Config), set
			},
			validate: func(o *opampAgent, t *testing.T) {
				olduid := o.instanceId
				assert.NotEmpty(t, olduid.String())
			
				uid := ulid.Make()
				assert.NotEqual(t, uid, olduid)
			
				o.updateAgentIdentity(uid)
				assert.Equal(t, o.instanceId, uid)
			},
		},
		{
			name: "ComposeEffectiveConfig",
			setup: func() (*Config, extension.Settings) {
				cfg := createDefaultConfig()
				set := extensiontest.NewNopSettings()
				return cfg.(*Config), set
			},
			validate: func(o *opampAgent, t *testing.T) {
				ec := o.composeEffectiveConfig()
				assert.NotNil(t, ec)
			},
		},
		{
			name: "ApplyRemoteConfig",
			setup: func() (*Config, extension.Settings) {
				d, err := os.MkdirTemp("", "opamp.d")
				assert.NoError(t, err)
				defer os.RemoveAll(d)

				cfg := createDefaultConfig().(*Config)
				cfg.RemoteConfigurationDirectory = d
				set := extensiontest.NewNopSettings()
				return cfg, set
			},
			validate: func(o *opampAgent, t *testing.T) {
				path := filepath.Join("testdata", "opamp.d", "opamp-remote-config.yaml")
				rc := remoteConfig(t, path)

				changed, err := o.applyRemoteConfig(rc)
				assert.NoError(t, err)
				assert.True(t, changed)
				assert.NotEqual(t, len(o.effectiveConfig), 0)

				
				o.cfg.AcceptsRemoteConfiguration = false
				changed, err = o.applyRemoteConfig(rc)
				assert.False(t, changed)
				assert.Error(t, err)
				assert.Equal(t, "OpAMP agent does not accept remote configuration", err.Error())
			},
		},

	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg, set := tc.setup()
			o, err := newOpampAgent(cfg, set.Logger, set.BuildInfo, set.Resource)
			assert.NoError(t, err)
			tc.validate(o, t)
		})
	}
}