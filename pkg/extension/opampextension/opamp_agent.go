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
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap"
	"go.opentelemetry.io/collector/confmap/converter/expandconverter"
	"go.opentelemetry.io/collector/confmap/provider/fileprovider"
	"go.opentelemetry.io/collector/confmap/provider/yamlprovider"
	"go.opentelemetry.io/collector/otelcol"
	"go.opentelemetry.io/collector/pdata/pcommon"
	semconv "go.opentelemetry.io/collector/semconv/v1.18.0"
	"go.uber.org/zap"

	"github.com/google/uuid"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/knadh/koanf/v2"
	"github.com/oklog/ulid/v2"
	"github.com/open-telemetry/opamp-go/client"
	"github.com/open-telemetry/opamp-go/client/types"
	"github.com/open-telemetry/opamp-go/protobufs"
	"github.com/open-telemetry/opentelemetry-collector-contrib/extension/sumologicextension"
)

type opampAgent struct {
	cfg    *Config
	host   component.Host
	logger *zap.Logger

	authExtension *sumologicextension.SumologicExtension
	authHeader    http.Header

	endpoint string

	agentType    string
	agentVersion string

	instanceId ulid.ULID

	effectiveConfig map[string]*protobufs.AgentConfigFile

	agentDescription *protobufs.AgentDescription

	opampClient client.OpAMPClient

	remoteConfigStatus *protobufs.RemoteConfigStatus
}

// opampLogShim translates between zap and opamp-go's bespoke logging interface
type opampLogShim struct {
	logger *zap.SugaredLogger
}

func (o opampLogShim) Debugf(_ context.Context, fmt string, v ...interface{}) {
	o.logger.Debugf(fmt, v...)
}

func (o opampLogShim) Errorf(_ context.Context, fmt string, v ...interface{}) {
	o.logger.Errorf(fmt, v...)
}

func (o *opampAgent) Start(ctx context.Context, host component.Host) error {
	o.host = host
	o.opampClient = client.NewWebSocket(opampLogShim{o.logger.Sugar()})

	if err := o.loadEffectiveConfig(o.cfg.RemoteConfigurationDirectory); err != nil {
		return err
	}

	if err := o.createAgentDescription(); err != nil {
		return err
	}

	if err := o.opampClient.SetAgentDescription(o.agentDescription); err != nil {
		return err
	}

	if err := o.getAuthExtension(); err != nil {
		return err
	}

	var baseURL string
	if o.authExtension != nil {
		baseURL = o.authExtension.BaseURL()
	}

	if err := o.setEndpoint(baseURL); err != nil {
		return err
	}

	if o.authExtension == nil {
		return o.startClient(ctx)
	}

	if err := o.createAuthHeader(); err == nil {
		return o.startClient(ctx)
	}

	go func() {
		// Wait for the authentication extension to start and produce credentials.
		o.authExtension.WatchCredentialKey(ctx, "")
		if err := o.createAuthHeader(); err != nil {
			o.logger.Error("Failed to start OpAMP agent", zap.Error(err))
			return
		}
		if err := o.startClient(ctx); err != nil {
			o.logger.Error("Failed to start OpAMP agent", zap.Error(err))
			return
		}
	}()

	return nil
}

func (o *opampAgent) Shutdown(ctx context.Context) error {
	o.logger.Debug("OpAMP agent shutting down...")
	if o.opampClient != nil {
		o.logger.Debug("Stopping OpAMP client...")
		err := o.opampClient.Stop(ctx)
		return err
	}
	return nil
}

func (o *opampAgent) Reload(ctx context.Context) error {
	err := o.Shutdown(ctx)

	if err != nil {
		return err
	}

	return o.Start(ctx, o.host)
}

func (o *opampAgent) startClient(ctx context.Context) error {
	settings := types.StartSettings{
		Header:         o.authHeader,
		OpAMPServerURL: o.endpoint,
		InstanceUid:    o.instanceId.String(),
		Callbacks: types.CallbacksStruct{
			OnConnectFunc: func(ctx context.Context) {
				o.logger.Info("Connected to the OpAMP server")
			},
			OnConnectFailedFunc: func(ctx context.Context, err error) {
				o.logger.Error("Failed to connect to the OpAMP server", zap.Error(err))
			},
			OnErrorFunc: func(ctx context.Context, err *protobufs.ServerErrorResponse) {
				o.logger.Error("OpAMP server returned an error response", zap.String("message", err.ErrorMessage))
			},
			SaveRemoteConfigStatusFunc: func(_ context.Context, status *protobufs.RemoteConfigStatus) {
				o.remoteConfigStatus = status
			},
			GetEffectiveConfigFunc: func(ctx context.Context) (*protobufs.EffectiveConfig, error) {
				o.logger.Info("OpAMP server requested the effective configuration")
				return o.composeEffectiveConfig(), nil
			},
			OnMessageFunc: o.onMessage,
		},
		RemoteConfigStatus: o.remoteConfigStatus,
		Capabilities:       o.getAgentCapabilities(),
	}

	o.logger.Debug("Starting OpAMP client...")

	if err := o.opampClient.Start(ctx, settings); err != nil {
		return err
	}

	o.logger.Debug("OpAMP client started")

	if o.authExtension != nil {
		if err := o.watchCredentials(ctx, o.Reload); err != nil {
			return err
		}
	}

	return nil
}

func (o *opampAgent) getAuthExtension() error {
	settings := o.cfg.ClientConfig

	if settings.Auth == nil {
		return nil
	}

	for _, e := range o.host.GetExtensions() {
		v, ok := e.(*sumologicextension.SumologicExtension)
		if ok && settings.Auth.AuthenticatorID == v.ComponentID() {
			o.authExtension = v
			break
		}
	}

	if o.authExtension == nil {
		return fmt.Errorf(
			"sumologic was specified as auth extension (named: %q) but "+
				"a matching extension was not found in the config, "+
				"please re-check the config and/or define the sumologicextension",
			settings.Auth.AuthenticatorID.String(),
		)
	}

	return nil
}

func (o *opampAgent) createAuthHeader() error {
	h, err := o.authExtension.CreateCredentialsHeader()
	if err != nil {
		return err
	}

	o.authHeader = h

	return nil
}

func (o *opampAgent) watchCredentials(ctx context.Context, callback func(ctx context.Context) error) error {
	k := o.authExtension.WatchCredentialKey(ctx, "")

	go func() {
		o.authExtension.WatchCredentialKey(ctx, k)
		if err := callback(ctx); err != nil {
			o.logger.Error("Failed to execute watch credential key callback", zap.Error(err))
		}
	}()

	return nil
}

// setEndpoint sets the OpAMP endpoint based on the collector endpoint.
// This is a hack, and it should be removed when the backend is able to
// correctly redirect our OpAMP client to the correct URL.
func (o *opampAgent) setEndpoint(baseURL string) error {
	if baseURL == "" {
		o.endpoint = o.cfg.Endpoint
		return nil
	}
	u, err := url.Parse(baseURL)
	if err != nil {
		return fmt.Errorf("url error, cannot set opamp endpoint: %s", err)
	}

	u.Scheme = "wss"
	u.Path = "/v1/opamp"

	// These replacements are specific to Sumo Logic's current domain naming,
	// and are made provisionally for the OTRM beta. In the future, the backend
	// will inform the agent of the correct OpAMP URL to use.
	u.Host = strings.Replace(u.Host, "open-events", "opamp-events", 1)
	u.Host = strings.Replace(u.Host, "open-collectors", "opamp-collectors", 1)

	o.endpoint = u.String()

	return nil
}

func newOpampAgent(cfg *Config, logger *zap.Logger, build component.BuildInfo, res pcommon.Resource) (*opampAgent, error) {
	agentType := build.Command

	sn, ok := res.Attributes().Get(semconv.AttributeServiceName)
	if ok {
		agentType = sn.AsString()
	}

	agentVersion := build.Version

	sv, ok := res.Attributes().Get(semconv.AttributeServiceVersion)
	if ok {
		agentVersion = sv.AsString()
	}

	uid := ulid.Make()

	if cfg.InstanceUID != "" {
		puid, err := ulid.Parse(cfg.InstanceUID)
		if err != nil {
			return nil, err
		}
		uid = puid
	} else {
		sid, ok := res.Attributes().Get(semconv.AttributeServiceInstanceID)
		if ok {
			uuid, err := uuid.Parse(sid.AsString())
			if err != nil {
				return nil, err
			}
			uid = ulid.ULID(uuid)
		}
	}

	agent := &opampAgent{
		cfg:          cfg,
		logger:       logger,
		agentType:    agentType,
		agentVersion: agentVersion,
		instanceId:   uid,
	}

	return agent, nil
}

func (o *opampAgent) getAgentCapabilities() protobufs.AgentCapabilities {
	c := protobufs.AgentCapabilities_AgentCapabilities_ReportsEffectiveConfig

	if o.cfg.AcceptsRemoteConfiguration {
		c = c |
			protobufs.AgentCapabilities_AgentCapabilities_AcceptsRemoteConfig |
			protobufs.AgentCapabilities_AgentCapabilities_ReportsRemoteConfig
	}

	return c
}

func stringKeyValue(key, value string) *protobufs.KeyValue {
	return &protobufs.KeyValue{
		Key: key,
		Value: &protobufs.AnyValue{
			Value: &protobufs.AnyValue_StringValue{StringValue: value},
		},
	}
}

func (o *opampAgent) createAgentDescription() error {
	hostname, err := os.Hostname()
	if err != nil {
		return err
	}

	ident := []*protobufs.KeyValue{
		stringKeyValue(semconv.AttributeServiceInstanceID, o.instanceId.String()),
		stringKeyValue(semconv.AttributeServiceName, o.agentType),
		stringKeyValue(semconv.AttributeServiceVersion, o.agentVersion),
	}

	nonIdent := []*protobufs.KeyValue{
		stringKeyValue("os.arch", runtime.GOARCH),
		stringKeyValue("os.family", runtime.GOOS),
		stringKeyValue("host.name", hostname),
	}

	o.agentDescription = &protobufs.AgentDescription{
		IdentifyingAttributes:    ident,
		NonIdentifyingAttributes: nonIdent,
	}

	return nil
}

func (o *opampAgent) loadEffectiveConfig(dir string) error {
	if _, err := os.Stat(dir); err != nil {
		return err
	}

	ec := map[string]*protobufs.AgentConfigFile{}

	paths, err := filepath.Glob(filepath.Join(dir, "*.yaml"))
	if err != nil {
		return err
	}

	for _, p := range paths {
		var k = koanf.New(".")

		rb, err := os.ReadFile(p)

		if err != nil {
			return err
		}

		if err := k.Load(rawbytes.Provider(rb), yaml.Parser()); err != nil {
			return err
		}

		fb, err := k.Marshal(yaml.Parser())
		if err != nil {
			return err
		}

		ec[filepath.Base(p)] = &protobufs.AgentConfigFile{Body: fb}
	}

	o.effectiveConfig = ec

	return nil
}

func (o *opampAgent) saveEffectiveConfig(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}

	files, err := d.Readdir(0)
	if err != nil {
		return err
	}

	for _, f := range files {
		err := os.Remove(filepath.Join(dir, f.Name()))
		if err != nil {
			return err
		}
	}

	for k, v := range o.effectiveConfig {
		p := filepath.Join(dir, k)

		// OpenFile the same way os.Create does it, but with 0600 perms
		f, err := os.OpenFile(p, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
		if err != nil {
			return err
		}
		defer f.Close()

		_, err = f.Write(v.Body)
		if err != nil {
			return err
		}
		o.logger.Debug("Loading Component Factories...")
		factories, err := Components()
		if err != nil {
			return fmt.Errorf("cannot get the list of factories: %v", err)
		}
		o.logger.Info("Loading Configuration to Validate...")

		_, errValidate := loadConfigAndValidateWithSettings(factories, otelcol.ConfigProviderSettings{
			ResolverSettings: confmap.ResolverSettings{
				URIs: []string{p},
				ProviderFactories: []confmap.ProviderFactory{
					fileprovider.NewFactory(),
					yamlprovider.NewFactory(),
				},
				ConverterFactories: []confmap.ConverterFactory{expandconverter.NewFactory()},
			},
		})
		if errValidate != nil {
			o.logger.Error("Validation Failed... %v", zap.Error(errValidate))
			return fmt.Errorf("cannot validate config named %v", errValidate)
		}
		o.logger.Info("Config Validation Successful...")
	}

	return nil
}

func (o *opampAgent) updateAgentIdentity(instanceId ulid.ULID) {
	o.logger.Debug("OpAMP agent identity is being changed",
		zap.String("old_id", o.instanceId.String()),
		zap.String("new_id", instanceId.String()))
	o.instanceId = instanceId
}

func (o *opampAgent) composeEffectiveConfig() *protobufs.EffectiveConfig {
	return &protobufs.EffectiveConfig{
		ConfigMap: &protobufs.AgentConfigMap{
			ConfigMap: o.effectiveConfig,
		},
	}
}

func (o *opampAgent) applyRemoteConfig(config *protobufs.AgentRemoteConfig) (configChanged bool, err error) {
	o.logger.Info("Received remote config from OpAMP server", zap.ByteString("hash", config.ConfigHash))

	if !o.cfg.AcceptsRemoteConfiguration {
		return false, fmt.Errorf("OpAMP agent does not accept remote configuration")
	}

	nec := map[string]*protobufs.AgentConfigFile{}

	for n, _ := range config.Config.ConfigMap {
		var k = koanf.New(".")

		data := `
extensions:
  sumologic:
    collector_fields:
      gg: testing1234
      sumo.disco.enabled: "true"
      trst: "1323"
`

		// Load the YAML data into the Koanf instance
		err := k.Load(rawbytes.Provider([]byte(data)), yaml.Parser())
		//err := k.Load(rawbytes.Provider(f.Body), yaml.Parser())
		if err != nil {
			return false, fmt.Errorf("cannot parse config named %s: %v", n, err)
		}
		k.Set("receivers.nop", map[string]interface{}{})
		k.Set("exporters.nop", map[string]interface{}{})
		fb, err := k.Marshal(yaml.Parser())
		o.logger.Info("Agent config yaml", zap.String("config", string(fb)))
		
		if err != nil {
			return false, fmt.Errorf("cannot marshal config named %s: %v", n, err)
		}

		nec[n] = &protobufs.AgentConfigFile{Body: fb}
	}

	configChanged = false
	if !reflect.DeepEqual(o.effectiveConfig, nec) {
		o.logger.Info("Start saving effective config")
		oec := o.effectiveConfig
		o.effectiveConfig = nec

		err := o.saveEffectiveConfig(o.cfg.RemoteConfigurationDirectory)
		if err != nil {
			o.effectiveConfig = oec
			return false, fmt.Errorf("cannot save the OpAMP effective config to %s: %v", o.cfg.RemoteConfigurationDirectory, err)
		}

		configChanged = true
	}

	return configChanged, nil
}

func (o *opampAgent) onMessage(ctx context.Context, msg *types.MessageData) {
	configChanged := false

	if msg.RemoteConfig != nil {
		var err error
		configChanged, err = o.applyRemoteConfig(msg.RemoteConfig)
		if err == nil {
			err = fmt.Errorf("Simulated test error")
			o.logger.Error("Failed to apply OpAMP agent remote config", zap.Error(err))
			err = o.opampClient.SetRemoteConfigStatus(&protobufs.RemoteConfigStatus{
				LastRemoteConfigHash: msg.RemoteConfig.ConfigHash,
				Status:               protobufs.RemoteConfigStatuses_RemoteConfigStatuses_FAILED,
				ErrorMessage:         err.Error(),
			})

			if err != nil {
				o.logger.Error("Failed to set OpAMP agent remote config status", zap.Error(err))
			}
		} else {
			err = o.opampClient.SetRemoteConfigStatus(&protobufs.RemoteConfigStatus{
				LastRemoteConfigHash: msg.RemoteConfig.ConfigHash,
				Status:               protobufs.RemoteConfigStatuses_RemoteConfigStatuses_APPLIED,
			})

			if err != nil {
				o.logger.Error("Failed to set OpAMP agent remote config status", zap.Error(err))
			}
		}
	}

	if msg.AgentIdentification != nil {
		instanceId, err := ulid.Parse(msg.AgentIdentification.NewInstanceUid)
		if err != nil {
			o.logger.Error("Failed to parse a new OpAMP agent identity", zap.Error(err))
		}
		o.updateAgentIdentity(instanceId)
	}

	if configChanged {
		err := o.opampClient.UpdateEffectiveConfig(ctx)
		if err != nil {
			o.logger.Error("Failed to update OpAMP agent effective configuration", zap.Error(err))
		}

		err = reloadCollectorConfig()
		if err != nil {
			o.logger.Error("Failed to reload collector configuration via SIGHUP", zap.Error(err))
		}
	}
}

func loadConfigWithSettings(factories otelcol.Factories, set otelcol.ConfigProviderSettings) (*otelcol.Config, error) {
	// Read yaml config from file
	provider, err := otelcol.NewConfigProvider(set)
	if err != nil {
		return nil, err
	}
	return provider.Get(context.Background(), factories)
}

func loadConfigAndValidateWithSettings(factories otelcol.Factories, set otelcol.ConfigProviderSettings) (*otelcol.Config, error) {
	cfg, err := loadConfigWithSettings(factories, set)
	if err != nil {
		return nil, err
	}
	return cfg, cfg.Validate()
}
