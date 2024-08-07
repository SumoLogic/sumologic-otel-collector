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

const errMsgRemoteConfigNotAccepted = "OpAMP agent does not accept remote configuration"
const errMsgInvalidConfigName = "cannot validate config named " +
	"service::pipelines::logs/localfilesource/0aa79379-c764-4d3d-9d66-03f6df029a07: " +
	"references processor \"batch\" which is not configured"

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

func TestApplyRemoteConfig(t *testing.T) {
	tests := []struct {
		name          string
		file          string
		acceptsRemote bool
		expectChanged bool
		expectError   bool
		errorMessage  string
	}{
		{"ApplyRemoteConfig", "testdata/opamp.d/opamp-remote-config.yaml", false, false, true, errMsgRemoteConfigNotAccepted},
		{"ApplyRemoteApacheConfig", "testdata/opamp.d/opamp-apache-config.yaml", false, false, true, errMsgRemoteConfigNotAccepted},
		{"ApplyRemoteHostConfig", "testdata/opamp.d/opamp-host-config.yaml", false, false, true, errMsgRemoteConfigNotAccepted},
		{"ApplyRemoteWindowsEventConfig", "testdata/opamp.d/opamp-windows-event-config.yaml", false, false, true, errMsgRemoteConfigNotAccepted},
		{"ApplyRemoteExtensionsConfig", "testdata/opamp.d/opamp-extensions-config.yaml", false, false, true, errMsgRemoteConfigNotAccepted},
		{"ApplyRemoteConfigFailed", "testdata/opamp.d/opamp-invalid-remote-config.yaml", true, false, true, "'max_elapsed_time' must be non-negative"},
		{"ApplyRemoteConfigMissingProcessor", "testdata/opamp.d/opamp-missing-processor.yaml", true, false, true, errMsgInvalidConfigName},
		{"ApplyFilterProcessorConfig", "testdata/opamp.d/opamp-filter-processor.yaml", false, false, true, errMsgRemoteConfigNotAccepted},
		{"ApplyKafkaMetricsConfig", "testdata/opamp.d/opamp-kafkametrics-config.yaml", false, false, true, errMsgRemoteConfigNotAccepted},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d, err := os.MkdirTemp("", "opamp.d")
			assert.NoError(t, err)
			defer os.RemoveAll(d)

			cfg := createDefaultConfig().(*Config)
			cfg.RemoteConfigurationDirectory = d
			set := extensiontest.NewNopSettings()
			o, err := newOpampAgent(cfg, set.Logger, set.BuildInfo, set.Resource)
			assert.NoError(t, err)

			cfg.AcceptsRemoteConfiguration = tt.acceptsRemote

			path := filepath.Join(tt.file)
			rb, err := os.ReadFile(path)
			assert.NoError(t, err)

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

			changed, err := o.applyRemoteConfig(rc)
			if tt.expectError {
				assert.Error(t, err)
				assert.ErrorContains(t, err, tt.errorMessage)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expectChanged, changed)
		})
	}
}

func TestGetAgentCapabilities(t *testing.T) {
	cfg, set := defaultSetup()
	o, err := newOpampAgent(cfg, set.Logger, set.BuildInfo, set.Resource)
	assert.NoError(t, err)

	assert.Equal(t, o.getAgentCapabilities(), protobufs.AgentCapabilities(4102))

	cfg.AcceptsRemoteConfiguration = false
	assert.Equal(t, o.getAgentCapabilities(), protobufs.AgentCapabilities(4))
}

func TestCreateAgentDescription(t *testing.T) {
	cfg, set := defaultSetup()
	o, err := newOpampAgent(cfg, set.Logger, set.BuildInfo, set.Resource)
	assert.NoError(t, err)

	assert.Nil(t, o.agentDescription)
	assert.NoError(t, o.createAgentDescription())
	assert.NotNil(t, o.agentDescription)
}

func TestLoadEffectiveConfig(t *testing.T) {
	cfg, set := defaultSetup()
	o, err := newOpampAgent(cfg, set.Logger, set.BuildInfo, set.Resource)
	assert.NoError(t, err)

	assert.Equal(t, len(o.effectiveConfig), 0)

	assert.NoError(t, o.loadEffectiveConfig("testdata"))
	assert.NotEqual(t, len(o.effectiveConfig), 0)
}

func TestSaveEffectiveConfig(t *testing.T) {
	cfg, set := defaultSetup()
	o, err := newOpampAgent(cfg, set.Logger, set.BuildInfo, set.Resource)
	assert.NoError(t, err)

	d, err := os.MkdirTemp("", "opamp.d")
	assert.NoError(t, err)
	defer os.RemoveAll(d)

	assert.NoError(t, o.saveEffectiveConfig(d))
}

func TestUpdateAgentIdentity(t *testing.T) {
	cfg, set := defaultSetup()
	o, err := newOpampAgent(cfg, set.Logger, set.BuildInfo, set.Resource)
	assert.NoError(t, err)

	olduid := o.instanceId
	assert.NotEmpty(t, olduid.String())

	uid := ulid.Make()
	assert.NotEqual(t, uid, olduid)

	o.updateAgentIdentity(uid)
	assert.Equal(t, o.instanceId, uid)
}

func TestComposeEffectiveConfig(t *testing.T) {
	cfg, set := defaultSetup()
	o, err := newOpampAgent(cfg, set.Logger, set.BuildInfo, set.Resource)
	assert.NoError(t, err)

	ec := o.composeEffectiveConfig()
	assert.NotNil(t, ec)
}

func TestShutdown(t *testing.T) {
	cfg, set := defaultSetup()
	cfg.ClientConfig.Auth = nil

	o, err := newOpampAgent(cfg, set.Logger, set.BuildInfo, set.Resource)
	assert.NoError(t, err)

	// Shutdown with no OpAMP client
	assert.NoError(t, o.Shutdown(context.Background()))
}

// To be removed when the OpAMP backend correctly advertises its URL.
// This test is badly written, it reaches into unexported fields to test
// their contents, but given that it will be removed in a few months, that
// shouldn't matter too much from a big picture perspective.
func TestHackSetEndpoint(t *testing.T) {
	tests := []struct {
		name         string
		url          string
		wantEndpoint string
	}{
		{
			name: "empty url defaults to config endpoint",
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
		}, {

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
}

func TestNewOpampAgent(t *testing.T) {
	cfg, set := setupWithBuildInfo("test version", "otelcoltest")
	o, err := newOpampAgent(cfg, set.Logger, set.BuildInfo, set.Resource)
	assert.NoError(t, err)
	assert.Equal(t, "otelcoltest", o.agentType)
	assert.Equal(t, "test version", o.agentVersion)
	assert.NotEmpty(t, o.instanceId.String())
	assert.Empty(t, o.effectiveConfig)
	assert.Nil(t, o.agentDescription)
}

func TestNewOpampAgentAttributes(t *testing.T) {
	cfg, set := setupWithBuildInfo("test version", "otelcoltest")
	set.Resource.Attributes().PutStr(semconv.AttributeServiceName, "otelcol-sumo")
	set.Resource.Attributes().PutStr(semconv.AttributeServiceVersion, "sumo.0")
	set.Resource.Attributes().PutStr(semconv.AttributeServiceInstanceID, "f8999bc1-4c9b-4619-9bae-7f009d2411ec")
	o, err := newOpampAgent(cfg, set.Logger, set.BuildInfo, set.Resource)
	assert.NoError(t, err)
	assert.Equal(t, "otelcol-sumo", o.agentType)
	assert.Equal(t, "sumo.0", o.agentVersion)
	assert.Equal(t, "7RK6DW2K4V8RCSQBKZ02EJ84FC", o.instanceId.String())
}

func TestStart(t *testing.T) {
	d, err := os.MkdirTemp("", "opamp.d")
	assert.NoError(t, err)
	defer os.RemoveAll(d)
	cfg, set := setupWithRemoteConfig(t, d)
	cfg.ClientConfig.Auth = nil
	cfg.RemoteConfigurationDirectory = d
	o, err := newOpampAgent(cfg, set.Logger, set.BuildInfo, set.Resource)
	assert.NoError(t, err)

	assert.NoError(t, o.Start(context.Background(), componenttest.NewNopHost()))
}

func TestReload(t *testing.T) {
	d, err := os.MkdirTemp("", "opamp.d")
	assert.NoError(t, err)
	defer os.RemoveAll(d)
	cfg, set := setupWithRemoteConfig(t, d)
	cfg.ClientConfig.Auth = nil
	o, err := newOpampAgent(cfg, set.Logger, set.BuildInfo, set.Resource)
	assert.NoError(t, err)

	ctx := context.Background()
	assert.NoError(t, o.Start(ctx, componenttest.NewNopHost()))
	assert.NoError(t, o.Reload(ctx))
}
