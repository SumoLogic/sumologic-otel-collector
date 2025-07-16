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
	"go.opentelemetry.io/collector/config/configauth"
	"go.opentelemetry.io/collector/config/configoptional"
	"os"
	"path/filepath"
	"testing"

	"github.com/oklog/ulid/v2"
	"github.com/open-telemetry/opamp-go/protobufs"
	"github.com/stretchr/testify/assert"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/extension"
	"go.opentelemetry.io/collector/extension/extensiontest"
	semconv "go.opentelemetry.io/otel/semconv/v1.18.0"
)

const (
	errMsgRemoteConfigNotAccepted = "OpAMP agent does not accept remote configuration"
	errMsgInvalidConfigName       = "cannot validate config: " +
		"service::pipelines::logs/localfilesource/0aa79379-c764-4d3d-9d66-03f6df029a07: " +
		"references processor \"batch\" which is not configured"
	errMsgInvalidType              = "'spike_limit_percentage' expected type 'uint32'"
	errExpectedUncofiguredEndPoint = "expected unconfigured opamp endpoint to result in default sumo opamp url setting"
)

func defaultSetup() (*Config, extension.Settings) {
	cfg := createDefaultConfig().(*Config)
	set := extensiontest.NewNopSettings(extensiontest.NopType)
	set.BuildInfo = component.BuildInfo{Version: "test version", Command: "otelcoltest"}
	return cfg, set
}

func setupWithRemoteConfig(t *testing.T, d string) (*Config, extension.Settings) {
	cfg, set := defaultSetup()
	cfg.RemoteConfigurationDirectory = d
	return cfg, set
}

func TestApplyRemoteConfig(t *testing.T) {
	tests := []struct {
		name         string
		file         string
		expectError  bool
		errorMessage string
	}{
		{"ApplyRemoteConfig", "testdata/opamp.d/opamp-remote-config.yaml", false, ""},
		{"ApplyRemoteApacheConfig", "testdata/opamp.d/opamp-apache-config.yaml", false, ""},
		{"ApplyRemoteHostConfig", "testdata/opamp.d/opamp-host-config.yaml", false, ""},
		{"ApplyRemoteWindowsEventConfig", "testdata/opamp.d/opamp-windows-event-config.yaml", false, ""},
		{"ApplyRemoteExtensionsConfig", "testdata/opamp.d/opamp-extensions-config.yaml", false, ""},
		{"ApplyRemoteConfigFailed", "testdata/opamp.d/opamp-invalid-remote-config.yaml", true, errMsgInvalidType},
		{"ApplyRemoteConfigMissingProcessor", "testdata/opamp.d/opamp-missing-processor.yaml", true, errMsgInvalidConfigName},
		{"ApplyFilterProcessorConfig", "testdata/opamp.d/opamp-filter-processor.yaml", false, ""},
		{"ApplyKafkaMetricsConfig", "testdata/opamp.d/opamp-kafkametrics-config.yaml", false, ""},
		{"ApplyElasticsearchConfig", "testdata/opamp.d/opamp-elastic-config.yaml", false, ""},
		{"ApplyMysqlConfig", "testdata/opamp.d/opamp-mysql-config.yaml", false, ""},
		{"ApplyattributesprocessorConfig", "testdata/opamp.d/opamp-attributes-processor.yaml", false, ""},
		{"ApplyPostgresqlConfig", "testdata/opamp.d/opamp-postgresql-config.yaml", false, ""},
		{"ApplyRabbitmqConfig", "testdata/opamp.d/opamp-rabbitmq-config.yaml", false, ""},
		{"ApplyRedisConfig", "testdata/opamp.d/opamp-redis-config.yaml", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d, err := os.MkdirTemp("", "opamp.d")
			assert.NoError(t, err)
			defer os.RemoveAll(d)
			cfg, set := setupWithRemoteConfig(t, d)
			o, err := newOpampAgent(cfg, set.Logger, set.BuildInfo, set.Resource)
			assert.NoError(t, err)
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

			// Test with an error in configuration
			if tt.expectError {
				changed, err := o.applyRemoteConfig(rc)
				assert.Error(t, err)
				assert.ErrorContains(t, err, tt.errorMessage)
				assert.False(t, changed)
				assert.Equal(t, len(o.effectiveConfig), 0)
			} else {
				// Test with a valid configuration
				changed, err := o.applyRemoteConfig(rc)
				assert.NoError(t, err)
				assert.True(t, changed)
				assert.NotEqual(t, len(o.effectiveConfig), 0)
			}
			// Test with remote configuration disabled
			cfg.AcceptsRemoteConfiguration = false
			changed, err := o.applyRemoteConfig(rc)
			assert.False(t, changed)
			assert.Error(t, err)
			assert.Equal(t, errMsgRemoteConfigNotAccepted, err.Error())
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
	cfg.ClientConfig.Auth = configoptional.None[configauth.Config]()

	o, err := newOpampAgent(cfg, set.Logger, set.BuildInfo, set.Resource)
	assert.NoError(t, err)

	// Shutdown with no OpAMP client
	assert.NoError(t, o.Shutdown(context.Background()))
}

func TestStart(t *testing.T) {
	d, err := os.MkdirTemp("", "opamp.d")
	assert.NoError(t, err)
	defer os.RemoveAll(d)

	cfg := createDefaultConfig().(*Config)
	cfg.ClientConfig.Auth = configoptional.None[configauth.Config]()
	cfg.RemoteConfigurationDirectory = d
	set := extensiontest.NewNopSettings(extensiontest.NopType)
	o, err := newOpampAgent(cfg, set.Logger, set.BuildInfo, set.Resource)
	assert.NoError(t, err)

	assert.NoError(t, o.Start(context.Background(), componenttest.NewNopHost()))
}

func TestReload(t *testing.T) {
	d, err := os.MkdirTemp("", "opamp.d")
	assert.NoError(t, err)
	defer os.RemoveAll(d)

	cfg := createDefaultConfig().(*Config)
	cfg.ClientConfig.Auth = configoptional.None[configauth.Config]()
	cfg.RemoteConfigurationDirectory = d
	set := extensiontest.NewNopSettings(extensiontest.NopType)
	o, err := newOpampAgent(cfg, set.Logger, set.BuildInfo, set.Resource)
	assert.NoError(t, err)

	ctx := context.Background()
	assert.NoError(t, o.Start(ctx, componenttest.NewNopHost()))
	assert.NoError(t, o.Reload(ctx))
}

func TestDefaultEndpointSetOnStart(t *testing.T) {
	cfg := createDefaultConfig().(*Config)
	set := extensiontest.NewNopSettings(extensiontest.NopType)
	o, err := newOpampAgent(cfg, set.Logger, set.BuildInfo, set.Resource)
	if err != nil {
		t.Fatal(err)
	}
	settings := o.startSettings()
	if settings.OpAMPServerURL != DefaultSumoLogicOpAmpURL {
		t.Error(errExpectedUncofiguredEndPoint)
	}
}

func TestNewOpampAgent(t *testing.T) {
	cfg, set := defaultSetup()
	o, err := newOpampAgent(cfg, set.Logger, set.BuildInfo, set.Resource)
	assert.NoError(t, err)
	assert.Equal(t, "otelcoltest", o.agentType)
	assert.Equal(t, "test version", o.agentVersion)
	assert.NotEmpty(t, o.instanceId.String())
	assert.Empty(t, o.effectiveConfig)
	assert.Nil(t, o.agentDescription)
}

func TestNewOpampAgentAttributes(t *testing.T) {
	cfg, set := defaultSetup()
	set.Resource.Attributes().PutStr(string(semconv.ServiceNameKey), "otelcol-sumo")
	set.Resource.Attributes().PutStr(string(semconv.ServiceVersionKey), "sumo.0")
	set.Resource.Attributes().PutStr(string(semconv.ServiceInstanceIDKey), "f8999bc1-4c9b-4619-9bae-7f009d2411ec")
	o, err := newOpampAgent(cfg, set.Logger, set.BuildInfo, set.Resource)
	assert.NoError(t, err)
	assert.Equal(t, "otelcol-sumo", o.agentType)
	assert.Equal(t, "sumo.0", o.agentVersion)
	assert.Equal(t, "7RK6DW2K4V8RCSQBKZ02EJ84FC", o.instanceId.String())
}
