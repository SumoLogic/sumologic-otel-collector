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
package opampprovider

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"sort"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/oklog/ulid/v2"

	"github.com/open-telemetry/opamp-go/client"
	"github.com/open-telemetry/opamp-go/client/types"
	"github.com/open-telemetry/opamp-go/protobufs"
)

type Agent struct {
	logger types.Logger

	stateManager *stateManager

	agentType    string
	agentVersion string

	serverURL string

	configUpdated chan bool

	agentDescription *protobufs.AgentDescription

	opampClient client.OpAMPClient

	remoteConfigStatus *protobufs.RemoteConfigStatus
}

func newAgent(logger types.Logger) *Agent {
	return &Agent{
		logger:        logger,
		stateManager:  newStateManager(logger),
		agentType:     "sumologic-otel-collector",
		agentVersion:  "0.0.1",
		configUpdated: make(chan bool),
	}
}

func (agent *Agent) Start(serverURL string) error {
	state := agent.stateManager.GetState()

	agent.logger.Debugf("OpAMP agent starting, id=%v, type=%s, version=%s.",
		state.InstanceId, agent.agentType, agent.agentVersion)

	agent.serverURL = serverURL

	err := agent.createAgentDescription()
	if err != nil {
		return err
	}

	agent.opampClient = client.NewWebSocket(agent.logger)

	hostname, err := os.Hostname()
	if err != nil {
		return err
	}

	settings := types.StartSettings{
		OpAMPServerURL: agent.serverURL,
		Header: http.Header{
			"Authorization":  []string{fmt.Sprintf("Secret-Key %s", "foobar")},
			"User-Agent":     []string{fmt.Sprintf("sumologic-otel-collector/%s", "0.0.1")},
			"OpAMP-Version":  []string{"v0.2.0"}, // BindPlane currently requires OpAMP 0.2.0
			"Agent-ID":       []string{state.InstanceId},
			"Agent-Version":  []string{"0.0.1"},
			"Agent-Hostname": []string{hostname},
		},
		InstanceUid: state.InstanceId,
		Callbacks: types.CallbacksStruct{
			OnConnectFunc: func() {
				agent.logger.Debugf("Connected to the OpAMP server.")
			},
			OnConnectFailedFunc: func(err error) {
				agent.logger.Errorf("Failed to connect to the OpAMP server: %v", err)
			},
			OnErrorFunc: func(err *protobufs.ServerErrorResponse) {
				agent.logger.Errorf("OpAMP server returned an error response: %v", err.ErrorMessage)
			},
			SaveRemoteConfigStatusFunc: func(_ context.Context, status *protobufs.RemoteConfigStatus) {
				agent.logger.Debugf("Saving OpAMP remote status: %v", status)
				agent.remoteConfigStatus = status
			},
			GetEffectiveConfigFunc: func(ctx context.Context) (*protobufs.EffectiveConfig, error) {
				state := agent.stateManager.GetState()
				ecp, err := state.EffectiveConfig.composeEffectiveConfigProto()
				agent.logger.Debugf("Getting OpAMP effective config: %v", ecp)
				return ecp, err
			},
			OnMessageFunc: agent.onMessage,
		},
		RemoteConfigStatus: agent.remoteConfigStatus,
		Capabilities: protobufs.AgentCapabilities_AcceptsRemoteConfig |
			protobufs.AgentCapabilities_ReportsRemoteConfig |
			protobufs.AgentCapabilities_ReportsEffectiveConfig,
	}
	err = agent.opampClient.SetAgentDescription(agent.agentDescription)
	if err != nil {
		return err
	}

	agent.logger.Debugf("Starting OpAMP client...")

	err = agent.opampClient.Start(context.Background(), settings)
	if err != nil {
		return err
	}

	agent.logger.Debugf("OpAMP client started.")

	return nil
}

func (agent *Agent) loadState() error {
	if err := agent.stateManager.Load(); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			state := newAgentState()
			agent.stateManager.SetState(state)
			if err := agent.stateManager.Save(); err != nil {
				return err
			}
		} else {
			return err
		}
	}
	return nil
}

func stringKeyValue(key, value string) *protobufs.KeyValue {
	return &protobufs.KeyValue{
		Key: key,
		Value: &protobufs.AnyValue{
			Value: &protobufs.AnyValue_StringValue{StringValue: value},
		},
	}
}

func (agent *Agent) createAgentDescription() error {
	hostname, err := os.Hostname()
	if err != nil {
		return err
	}

	state := agent.stateManager.GetState()

	ident := []*protobufs.KeyValue{
		stringKeyValue("service.instance.id", state.InstanceId),
		stringKeyValue("service.instance.name", hostname),
		stringKeyValue("service.name", agent.agentType),
		stringKeyValue("service.version", agent.agentVersion),
	}

	nonIdent := []*protobufs.KeyValue{
		stringKeyValue("os.arch", runtime.GOARCH),
		stringKeyValue("os.details", runtime.GOOS),
		stringKeyValue("os.family", runtime.GOOS),
		stringKeyValue("host.name", hostname),
	}

	agent.agentDescription = &protobufs.AgentDescription{
		IdentifyingAttributes:    ident,
		NonIdentifyingAttributes: nonIdent,
	}

	return nil
}

func (agent *Agent) updateAgentIdentity(instanceId ulid.ULID) error {
	state := agent.stateManager.GetState()

	agent.logger.Debugf("OpAMP agent identity is being changed from id=%v to id=%v",
		state.InstanceId,
		instanceId.String())

	state.InstanceId = instanceId.String()
	agent.stateManager.SetState(state)

	err := agent.stateManager.Save()

	return err
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

func (agent *Agent) applyRemoteConfig(config *protobufs.AgentRemoteConfig) (configChanged bool, err error) {
	if config == nil {
		return false, nil
	}

	agent.logger.Debugf("OpAMP agent received remote config from server, hash=%x.", config.ConfigHash)

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
	newEffectiveConfigBytes, err := k.Marshal(yaml.Parser())
	if err != nil {
		panic(err)
	}

	state := agent.stateManager.GetState()

	oldEffectiveConfigBytes, err := yaml.Parser().Marshal(state.EffectiveConfig)
	if err != nil {
		panic(err)
	}

	configChanged = false
	if string(oldEffectiveConfigBytes) != string(newEffectiveConfigBytes) {
		agent.logger.Errorf("Effective config changed. Need to report to server.")
		state.EffectiveConfig = k.Raw()
		agent.stateManager.SetState(state)
		if err := agent.stateManager.Save(); err != nil {
			return false, fmt.Errorf("error saving state: %w", err)
		}
		configChanged = true
	}

	return configChanged, nil
}

func (agent *Agent) Shutdown() error {
	agent.logger.Debugf("OpAMP agent shutting down...")
	if agent.opampClient != nil {
		err := agent.opampClient.Stop(context.Background())
		return err
	}
	return nil
}

func (agent *Agent) onMessage(ctx context.Context, msg *types.MessageData) {
	configChanged := false
	if msg.RemoteConfig != nil {
		var err error
		configChanged, err = agent.applyRemoteConfig(msg.RemoteConfig)
		if err != nil {
			err = agent.opampClient.SetRemoteConfigStatus(&protobufs.RemoteConfigStatus{
				LastRemoteConfigHash: msg.RemoteConfig.ConfigHash,
				Status:               protobufs.RemoteConfigStatus_FAILED,
				ErrorMessage:         err.Error(),
			})

			if err != nil {
				agent.logger.Errorf(err.Error())
			}
		} else {
			err = agent.opampClient.SetRemoteConfigStatus(&protobufs.RemoteConfigStatus{
				LastRemoteConfigHash: msg.RemoteConfig.ConfigHash,
				Status:               protobufs.RemoteConfigStatus_APPLIED,
			})

			if err != nil {
				agent.logger.Errorf(err.Error())
			}
		}
	}

	if msg.AgentIdentification != nil {
		newInstanceId, err := ulid.Parse(msg.AgentIdentification.NewInstanceUid)
		if err != nil {
			agent.logger.Errorf(err.Error())
		}
		err = agent.updateAgentIdentity(newInstanceId)
		if err != nil {
			agent.logger.Errorf(err.Error())
		}
	}

	if configChanged {
		err := agent.opampClient.UpdateEffectiveConfig(ctx)
		if err != nil {
			agent.logger.Errorf(err.Error())
		}

		agent.configUpdated <- true
	}
}
