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
	"os"
	"path/filepath"
	"testing"

	"github.com/oklog/ulid/v2"
	"github.com/open-telemetry/opamp-go/protobufs"
	"github.com/stretchr/testify/assert"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/config/confighttp"
	"go.opentelemetry.io/collector/extension"
	"go.opentelemetry.io/collector/extension/extensiontest"

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
func defaultSetup() (*Config, extension.Settings) {
	cfg := createDefaultConfig().(*Config)
	set := extensiontest.NewNopSettings()
	return cfg, set
}

func setupWithRemoteConfig(t *testing.T, d string) (*Config, extension.Settings) {
	cfg, set := defaultSetup()
	cfg.RemoteConfigurationDirectory = d
	return cfg, set
}

func setupWithBuildInfo(version, command string) (*Config, extension.Settings) {
	cfg, set := defaultSetup()
	set.BuildInfo = component.BuildInfo{Version: version, Command: command}
	return cfg, set
}

func TestOpampAgenRemoteConfig(t *testing.T) {
	testCases := []struct {
		name     string
		validate func(*opampAgent, *testing.T)
	}{

		{
			name: "ApplyRemoteConfig",
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
		{
			name: "ApplyRemoteApacheConfig",
			validate: func(o *opampAgent, t *testing.T) {
				path := filepath.Join("testdata", "opamp.d", "opamp-apache-config.yaml")
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
		{
			name: "ApplyRemoteHostConfig",
			validate: func(o *opampAgent, t *testing.T) {
				path := filepath.Join("testdata", "opamp.d", "opamp-host-config.yaml")
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
		{
			name: "ApplyRemoteWindowsEventConfig",
			validate: func(o *opampAgent, t *testing.T) {
				path := filepath.Join("testdata", "opamp.d", "opamp-windows-event-config.yaml")
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
		{
			name: "ApplyRemoteExtensionsConfig",
			validate: func(o *opampAgent, t *testing.T) {
				path := filepath.Join("testdata", "opamp.d", "opamp-extensions-config.yaml")
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
		{
			name: "ApplyRemoteConfigFailed",
			validate: func(o *opampAgent, t *testing.T) {
				path := filepath.Join("testdata", "opamp.d", "opamp-invalid-remote-config.yaml")
				rc := remoteConfig(t, path)

				changed, err := o.applyRemoteConfig(rc)
				assert.Error(t, err)
				assert.ErrorContains(t, err, "'max_elapsed_time' must be non-negative")
				assert.False(t, changed)
				assert.Equal(t, len(o.effectiveConfig), 0)
			},
		},
		{
			name: "ApplyRemoteConfigMissingProcessor",
			validate: func(o *opampAgent, t *testing.T) {
				path := filepath.Join("testdata", "opamp.d", "opamp-missing-processor.yaml")
				rc := remoteConfig(t, path)

				changed, err := o.applyRemoteConfig(rc)
				assert.Error(t, err)
				expectedErrSubstring := "cannot validate config named service::pipelines::logs/localfilesource/0aa79379-c764-4d3d-9d66-03f6df029a07:" +
					" references processor \"batch\" which is not configured"
				assert.ErrorContains(t, err, expectedErrSubstring)
				assert.False(t, changed)
				assert.Equal(t, len(o.effectiveConfig), 0)
			},
		},
		{
			name: "ApplyFilterProcessorConfig",
			validate: func(o *opampAgent, t *testing.T) {
				path := filepath.Join("testdata", "opamp.d", "opamp-filter-processor.yaml")
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
		{
			name: "ApplyKafkaMetricsConfig",
			validate: func(o *opampAgent, t *testing.T) {
				path := filepath.Join("testdata", "opamp.d", "opamp-kafkametrics-config.yaml")
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
			d, err := os.MkdirTemp("", "opamp.d")
			assert.NoError(t, err)
			defer os.RemoveAll(d)
			cfg, set := setupWithRemoteConfig(t, d)
			o, err := newOpampAgent(cfg, set.Logger, set.BuildInfo, set.Resource)
			assert.NoError(t, err)
			tc.validate(o, t)
		})
	}
}
func TestOpampAgenAuthNil(t *testing.T) {
	testCases := []struct {
		name     string
		validate func(*opampAgent, *testing.T)
	}{
		{
			name: "Start",
			validate: func(o *opampAgent, t *testing.T) {
				assert.NoError(t, o.Start(context.Background(), componenttest.NewNopHost()))
			},
		},
		{
			name: "Reload",
			validate: func(o *opampAgent, t *testing.T) {
				ctx := context.Background()
				assert.NoError(t, o.Start(ctx, componenttest.NewNopHost()))
				assert.NoError(t, o.Reload(ctx))
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			d, err := os.MkdirTemp("", "opamp.d")
			assert.NoError(t, err)
			defer os.RemoveAll(d)
			cfg, set := setupWithRemoteConfig(t, d)
			cfg.ClientConfig.Auth = nil
			o, err := newOpampAgent(cfg, set.Logger, set.BuildInfo, set.Resource)
			assert.NoError(t, err)
			tc.validate(o, t)
		})
	}
}
func TestDefaultOpampAgent(t *testing.T) {
	testCases := []struct {
		name     string
		validate func(*opampAgent, *testing.T)
	}{
		{
			name: "GetAgentCapabilities",
			validate: func(o *opampAgent, t *testing.T) {
				assert.Equal(t, o.getAgentCapabilities(), protobufs.AgentCapabilities(4102))

				o.cfg.AcceptsRemoteConfiguration = false
				assert.Equal(t, o.getAgentCapabilities(), protobufs.AgentCapabilities(4))
			},
		},
		{
			name: "CreateAgentDescription",
			validate: func(o *opampAgent, t *testing.T) {
				assert.Nil(t, o.agentDescription)
				assert.NoError(t, o.createAgentDescription())
				assert.NotNil(t, o.agentDescription)
			},
		},
		{
			name: "LoadEffectiveConfig",
			validate: func(o *opampAgent, t *testing.T) {
				assert.Equal(t, len(o.effectiveConfig), 0)

				assert.NoError(t, o.loadEffectiveConfig("testdata"))
				assert.NotEqual(t, len(o.effectiveConfig), 0)
			},
		},
		{
			name: "SaveEffectiveConfig",
			validate: func(o *opampAgent, t *testing.T) {
				d, err := os.MkdirTemp("", "opamp.d")
				assert.NoError(t, err)
				defer os.RemoveAll(d)

				assert.NoError(t, o.saveEffectiveConfig(d))
			},
		},
		{
			name: "UpdateAgentIdentity",
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
			validate: func(o *opampAgent, t *testing.T) {
				ec := o.composeEffectiveConfig()
				assert.NotNil(t, ec)
			},
		},
		{
			name: "Shutdown",
			validate: func(o *opampAgent, t *testing.T) {
				cfg := createDefaultConfig().(*Config)
				cfg.ClientConfig.Auth = nil

				// Shutdown with no OpAMP client
				assert.NoError(t, o.Shutdown(context.Background()))
			},
		},
		{
			name: "HackSetEndpoint",
			validate: func(o *opampAgent, t *testing.T) {
				tests := []struct {
					name         string
					url          string
					wantEndpoint string
				}{
					{
						name:         "empty url defaults to config endpoint",
						wantEndpoint: "wss://example.com",
					},
					{
						name:         "url variant a",
						url:          "https://sumo-open-events.example.com",
						wantEndpoint: "wss://sumo-opamp-events.example.com/v1/opamp",
					},
					{
						name:         "url variant b",
						url:          "https://sumo-open-collectors.example.com",
						wantEndpoint: "wss://sumo-opamp-collectors.example.com/v1/opamp",
					},
					{
						name:         "url variant c",
						url:          "https://example.com",
						wantEndpoint: "wss://example.com/v1/opamp",
					},
					{
						name:         "dev sumologic url",
						url:          "https://long-open-events.sumologic.net/api/v1",
						wantEndpoint: "wss://long-opamp-events.sumologic.net/v1/opamp",
					},
					{
						name:         "prod sumologic url",
						url:          "https://open-collectors.sumologic.com/api/v1",
						wantEndpoint: "wss://opamp-collectors.sumologic.com/v1/opamp",
					},
					{
						name:         "prod sumologic url with region",
						url:          "https://open-collectors.au.sumologic.com/api/v1/",
						wantEndpoint: "wss://opamp-collectors.au.sumologic.com/v1/opamp",
					},
				}

				for _, test := range tests {
					t.Run(test.name, func(t *testing.T) {
						agent := &opampAgent{cfg: &Config{
							ClientConfig: confighttp.ClientConfig{
								Endpoint: "wss://example.com",
							},
						}}
						if err := agent.setEndpoint(test.url); err != nil {
							// can only happen with an invalid URL, which is quite hard
							// to even come up with for Go's URL package
							t.Fatal(err)
						}
						if got, want := agent.endpoint, test.wantEndpoint; got != want {
							t.Errorf("didn't get expected endpoint: got %q, want %q", got, want)
						}
					})
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg, set := defaultSetup()
			o, err := newOpampAgent(cfg, set.Logger, set.BuildInfo, set.Resource)
			assert.NoError(t, err)
			tc.validate(o, t)
		})
	}
}

func TestBuildInfoOpampAgent(t *testing.T) {
	testCases := []struct {
		name     string
		setup    func() (*Config, extension.Settings)
		validate func(*opampAgent, *testing.T)
	}{
		{
			name: "NewOpampAgent",
			setup: func() (*Config, extension.Settings) {
				cfg, set := setupWithBuildInfo("test version", "otelcoltest")
				return cfg, set
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
				cfg, set := setupWithBuildInfo("test version", "otelcoltest")
				set.Resource.Attributes().PutStr(semconv.AttributeServiceName, "otelcol-sumo")
				set.Resource.Attributes().PutStr(semconv.AttributeServiceVersion, "sumo.0")
				set.Resource.Attributes().PutStr(semconv.AttributeServiceInstanceID, "f8999bc1-4c9b-4619-9bae-7f009d2411ec")
				return cfg, set
			},
			validate: func(o *opampAgent, t *testing.T) {
				assert.Equal(t, "otelcol-sumo", o.agentType)
				assert.Equal(t, "sumo.0", o.agentVersion)
				assert.Equal(t, "7RK6DW2K4V8RCSQBKZ02EJ84FC", o.instanceId.String())
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
