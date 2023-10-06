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
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/oklog/ulid/v2"
	"github.com/open-telemetry/opamp-go/protobufs"
	"github.com/stretchr/testify/assert"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/extension/extensiontest"
	semconv "go.opentelemetry.io/collector/semconv/v1.18.0"
)

func TestNewOpampAgent(t *testing.T) {
	cfg := createDefaultConfig()
	set := extensiontest.NewNopCreateSettings()
	set.BuildInfo = component.BuildInfo{Version: "test version", Command: "otelcoltest"}
	o, err := newOpampAgent(cfg.(*Config), set.Logger, set.BuildInfo, set.Resource)
	assert.NoError(t, err)
	assert.Equal(t, o.agentType, "otelcoltest")
	assert.Equal(t, o.agentVersion, "test version")
	assert.NotEmpty(t, o.instanceId.String())
	assert.Empty(t, o.effectiveConfig)
	assert.Nil(t, o.agentDescription)
}

func TestNewOpampAgentAttributes(t *testing.T) {
	cfg := createDefaultConfig()
	set := extensiontest.NewNopCreateSettings()
	set.BuildInfo = component.BuildInfo{Version: "test version", Command: "otelcoltest"}
	set.Resource.Attributes().PutStr(semconv.AttributeServiceName, "otelcol-sumo")
	set.Resource.Attributes().PutStr(semconv.AttributeServiceVersion, "sumo.0")
	set.Resource.Attributes().PutStr(semconv.AttributeServiceInstanceID, "01BX5ZZKBKACTAV9WEVGEMMVRZ")
	o, err := newOpampAgent(cfg.(*Config), set.Logger, set.BuildInfo, set.Resource)
	assert.NoError(t, err)
	assert.Equal(t, o.agentType, "otelcol-sumo")
	assert.Equal(t, o.agentVersion, "sumo.0")
	assert.Equal(t, o.instanceId.String(), "01BX5ZZKBKACTAV9WEVGEMMVRZ")
}

func TestGetAgentCapabilities(t *testing.T) {
	cfg := createDefaultConfig().(*Config)
	set := extensiontest.NewNopCreateSettings()
	o, err := newOpampAgent(cfg, set.Logger, set.BuildInfo, set.Resource)
	assert.NoError(t, err)

	assert.Equal(t, o.getAgentCapabilities(), protobufs.AgentCapabilities(4102))

	cfg.AcceptsRemoteConfiguration = false
	assert.Equal(t, o.getAgentCapabilities(), protobufs.AgentCapabilities(4))
}

func TestCreateAgentDescription(t *testing.T) {
	cfg := createDefaultConfig()
	set := extensiontest.NewNopCreateSettings()
	o, err := newOpampAgent(cfg.(*Config), set.Logger, set.BuildInfo, set.Resource)
	assert.NoError(t, err)

	assert.Nil(t, o.agentDescription)
	assert.NoError(t, o.createAgentDescription())
	assert.NotNil(t, o.agentDescription)
}

func TestLoadEffectiveConfig(t *testing.T) {
	cfg := createDefaultConfig()
	set := extensiontest.NewNopCreateSettings()
	o, err := newOpampAgent(cfg.(*Config), set.Logger, set.BuildInfo, set.Resource)
	assert.NoError(t, err)

	assert.Equal(t, len(o.effectiveConfig), 0)

	assert.NoError(t, o.loadEffectiveConfig("testdata"))
	assert.NotEqual(t, len(o.effectiveConfig), 0)
}

func TestSaveEffectiveConfig(t *testing.T) {
	cfg := createDefaultConfig()
	set := extensiontest.NewNopCreateSettings()
	o, err := newOpampAgent(cfg.(*Config), set.Logger, set.BuildInfo, set.Resource)
	assert.NoError(t, err)

	d, err := ioutil.TempDir("", "opamp.d")
	assert.NoError(t, err)
	defer os.RemoveAll(d)

	assert.NoError(t, o.saveEffectiveConfig(d))
}

func TestUpdateAgentIdentity(t *testing.T) {
	cfg := createDefaultConfig()
	set := extensiontest.NewNopCreateSettings()
	o, err := newOpampAgent(cfg.(*Config), set.Logger, set.BuildInfo, set.Resource)
	assert.NoError(t, err)

	olduid := o.instanceId
	assert.NotEmpty(t, olduid.String())

	uid := ulid.Make()
	assert.NotEqual(t, uid, olduid)

	o.updateAgentIdentity(uid)
	assert.Equal(t, o.instanceId, uid)
}

func TestComposeEffectiveConfig(t *testing.T) {
	cfg := createDefaultConfig()
	set := extensiontest.NewNopCreateSettings()
	o, err := newOpampAgent(cfg.(*Config), set.Logger, set.BuildInfo, set.Resource)
	assert.NoError(t, err)

	ec := o.composeEffectiveConfig()
	assert.NotNil(t, ec)
}

func TestApplyRemoteConfig(t *testing.T) {
	d, err := ioutil.TempDir("", "opamp.d")
	assert.NoError(t, err)
	defer os.RemoveAll(d)

	cfg := createDefaultConfig().(*Config)
	cfg.RemoteConfigurationDirectory = d
	set := extensiontest.NewNopCreateSettings()
	o, err := newOpampAgent(cfg, set.Logger, set.BuildInfo, set.Resource)
	assert.NoError(t, err)

	assert.Equal(t, len(o.effectiveConfig), 0)

	path := filepath.Join("testdata", "opamp.d", "opamp-remote-config.yaml")
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
	assert.NoError(t, err)
	assert.True(t, changed)
	assert.NotEqual(t, len(o.effectiveConfig), 0)
}

func TestShutdown(t *testing.T) {
	cfg := createDefaultConfig().(*Config)
	cfg.HTTPClientSettings.Auth = nil
	set := extensiontest.NewNopCreateSettings()
	o, err := newOpampAgent(cfg, set.Logger, set.BuildInfo, set.Resource)
	assert.NoError(t, err)

	// Shutdown with no OpAMP client
	assert.NoError(t, o.Shutdown(context.Background()))
}

func TestStart(t *testing.T) {
	d, err := ioutil.TempDir("", "opamp.d")
	assert.NoError(t, err)
	defer os.RemoveAll(d)

	cfg := createDefaultConfig().(*Config)
	cfg.HTTPClientSettings.Auth = nil
	cfg.RemoteConfigurationDirectory = d
	set := extensiontest.NewNopCreateSettings()
	o, err := newOpampAgent(cfg, set.Logger, set.BuildInfo, set.Resource)
	assert.NoError(t, err)

	assert.NoError(t, o.Start(context.Background(), componenttest.NewNopHost()))
}

func TestReload(t *testing.T) {
	d, err := ioutil.TempDir("", "opamp.d")
	assert.NoError(t, err)
	defer os.RemoveAll(d)

	cfg := createDefaultConfig().(*Config)
	cfg.HTTPClientSettings.Auth = nil
	cfg.RemoteConfigurationDirectory = d
	set := extensiontest.NewNopCreateSettings()
	o, err := newOpampAgent(cfg, set.Logger, set.BuildInfo, set.Resource)
	assert.NoError(t, err)

	ctx := context.Background()
	assert.NoError(t, o.Start(ctx, componenttest.NewNopHost()))
	assert.NoError(t, o.Reload(ctx))
}
