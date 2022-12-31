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
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"syscall"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/oklog/ulid/v2"
	"go.opentelemetry.io/collector/component"
	"go.uber.org/zap"

	"github.com/open-telemetry/opamp-go/client"
	"github.com/open-telemetry/opamp-go/client/types"
	"github.com/open-telemetry/opamp-go/protobufs"

	"github.com/SumoLogic/sumologic-otel-collector/pkg/extension/sumologicextension"
)

type opampAgent struct {
	cfg    *Config
	host   component.Host
	logger *zap.Logger

	authExtension *sumologicextension.SumologicExtension
	authHeader    http.Header

	agentType    string
	agentVersion string

	instanceId ulid.ULID

	effectiveConfig string

	agentDescription *protobufs.AgentDescription

	opampClient client.OpAMPClient

	remoteConfigStatus *protobufs.RemoteConfigStatus
}

func (o *opampAgent) Start(ctx context.Context, host component.Host) error {
	o.host = host
	o.opampClient = client.NewWebSocket(&Logger{Logger: o.logger.Sugar()})

	path := filepath.Join(o.cfg.RemoteConfigurationDirectory, "opamp-remote-config.yaml")
	if _, err := os.Stat(path); err == nil {
		err := o.loadEffectiveConfig(path)

		if err != nil {
			return err
		}
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

	if err := o.createAuthHeader(); err == nil {
		return o.startClient(ctx)
	}

	go func() {
		// Wait for the authentication extension to start and produce credentials.
		o.authExtension.WaitForCredentials(ctx, "")
		if err := o.createAuthHeader(); err != nil {
			return
		}
		if err := o.startClient(ctx); err != nil {
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
		OpAMPServerURL: o.cfg.Endpoint,
		InstanceUid:    o.instanceId.String(),
		Callbacks: types.CallbacksStruct{
			OnConnectFunc: func() {
				o.logger.Info("Connected to the OpAMP server")
			},
			OnConnectFailedFunc: func(err error) {
				o.logger.Error("Failed to connect to the OpAMP server", zap.Error(err))
			},
			OnErrorFunc: func(err *protobufs.ServerErrorResponse) {
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
		Capabilities: protobufs.AgentCapabilities_AgentCapabilities_AcceptsRemoteConfig |
			protobufs.AgentCapabilities_AgentCapabilities_ReportsRemoteConfig |
			protobufs.AgentCapabilities_AgentCapabilities_ReportsEffectiveConfig,
	}

	o.logger.Debug("Starting OpAMP client...")

	if err := o.opampClient.Start(ctx, settings); err != nil {
		return err
	}

	o.logger.Debug("OpAMP client started")

	return nil
}

func (o *opampAgent) getAuthExtension() error {
	settings := o.cfg.HTTPClientSettings

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
	if o.authExtension == nil {
		return nil
	}

	h, err := o.authExtension.CreateCredentialsHeader()
	if err != nil {
		return err
	}

	o.authHeader = h

	return nil
}

func newOpampAgent(cfg *Config, logger *zap.Logger) (*opampAgent, error) {
	uid := ulid.Make() // TODO: Replace with https://github.com/open-telemetry/opentelemetry-collector/issues/6599

	if cfg.InstanceUID != "" {
		puid, err := ulid.Parse(cfg.InstanceUID)
		if err != nil {
			return nil, err
		}
		uid = puid
	}

	agent := &opampAgent{
		cfg:          cfg,
		logger:       logger,
		agentType:    "io.opentelemetry.collector",
		agentVersion: "1.0.0", // TODO: Replace with actual collector version info.
		instanceId:   uid,
	}

	return agent, nil
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
		stringKeyValue("service.instance.id", o.instanceId.String()),
		stringKeyValue("service.name", o.agentType),
		stringKeyValue("service.version", o.agentVersion),
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

func (o *opampAgent) loadEffectiveConfig(path string) error {
	var k = koanf.New(".")

	rb, err := os.ReadFile(path)

	if err != nil {
		return err
	}

	if err := k.Load(rawbytes.Provider(rb), yaml.Parser()); err != nil {
		return err
	}

	ecb, err := k.Marshal(yaml.Parser())
	if err != nil {
		return err
	}

	o.effectiveConfig = string(ecb)

	return nil
}

func (o *opampAgent) saveEffectiveConfig(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write([]byte(o.effectiveConfig))
	if err != nil {
		return err
	}

	return nil
}

func (o *opampAgent) reloadCollectorConfig() error {
	p, err := os.FindProcess(os.Getpid())

	if err != nil {
		return err
	}

	return p.Signal(syscall.SIGHUP)
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
			ConfigMap: map[string]*protobufs.AgentConfigFile{
				"": {Body: []byte(o.effectiveConfig)},
			},
		},
	}
}

type agentConfigFileItem struct {
	name string
	file *protobufs.AgentConfigFile
}

type agentConfigFileSlice []agentConfigFileItem

func (a agentConfigFileSlice) Less(i, j int) bool {
	return a[i].name < a[j].name
}

func (a agentConfigFileSlice) Swap(i, j int) {
	t := a[i]
	a[i] = a[j]
	a[j] = t
}

func (a agentConfigFileSlice) Len() int {
	return len(a)
}

func (o *opampAgent) applyRemoteConfig(config *protobufs.AgentRemoteConfig) (configChanged bool, err error) {
	o.logger.Debug("Received remote config from OpAMP server", zap.ByteString("hash", config.ConfigHash))

	var k = koanf.New(".")

	orderedConfigs := agentConfigFileSlice{}
	for name, file := range config.Config.ConfigMap {
		if name == "" {
			// skip instance config
			continue
		}
		orderedConfigs = append(orderedConfigs, agentConfigFileItem{
			name: name,
			file: file,
		})
	}

	// Sort to make sure the order of merging is stable.
	sort.Sort(orderedConfigs)

	// Append instance config as the last item.
	instanceConfig := config.Config.ConfigMap[""]
	if instanceConfig != nil {
		orderedConfigs = append(orderedConfigs, agentConfigFileItem{
			name: "",
			file: instanceConfig,
		})
	}

	// Merge received configs.
	for _, item := range orderedConfigs {
		var k2 = koanf.New(".")
		err := k2.Load(rawbytes.Provider(item.file.Body), yaml.Parser())
		if err != nil {
			return false, fmt.Errorf("cannot parse config named %s: %v", item.name, err)
		}
		err = k.Merge(k2)
		if err != nil {
			return false, fmt.Errorf("cannot merge config named %s: %v", item.name, err)
		}
	}

	// The merged final result is our effective config.
	effectiveConfigBytes, err := k.Marshal(yaml.Parser())
	if err != nil {
		return false, fmt.Errorf("cannot marshal the OpAMP effective config: %v", err)
	}

	newEffectiveConfig := string(effectiveConfigBytes)
	configChanged = false
	if o.effectiveConfig != newEffectiveConfig {
		path := filepath.Join(o.cfg.RemoteConfigurationDirectory, "opamp-remote-config.yaml")
		err := o.saveEffectiveConfig(path)
		if err != nil {
			return false, fmt.Errorf("cannot save the OpAMP effective config to %s: %v", path, err)
		}

		o.effectiveConfig = newEffectiveConfig
		configChanged = true
	}

	return configChanged, nil
}

func (o *opampAgent) onMessage(ctx context.Context, msg *types.MessageData) {
	configChanged := false

	if msg.RemoteConfig != nil {
		var err error
		configChanged, err = o.applyRemoteConfig(msg.RemoteConfig)
		if err != nil {
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

		err = o.reloadCollectorConfig()
		if err != nil {
			o.logger.Error("Failed to reload collector configuration via SIGHUP", zap.Error(err))
		}
	}
}
