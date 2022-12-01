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
	"reflect"
	"testing"

	"github.com/oklog/ulid/v2"
	"github.com/open-telemetry/opamp-go/client"
	"github.com/open-telemetry/opamp-go/client/types"
	"github.com/open-telemetry/opamp-go/protobufs"
	"github.com/stretchr/testify/require"
)

func TestAgent_Start(t *testing.T) {
	type fields struct {
		logger             types.Logger
		state              *agentState
		agentType          string
		agentVersion       string
		serverURL          string
		configUpdated      chan bool
		agentDescription   *protobufs.AgentDescription
		opampClient        client.OpAMPClient
		remoteConfigStatus *protobufs.RemoteConfigStatus
	}
	tests := []struct {
		name       string
		fields     fields
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "returns error when starting the client fails",
			fields: fields{
				logger: &NoopLogger{},
				state:  &agentState{},
				agentDescription: &protobufs.AgentDescription{
					IdentifyingAttributes: []*protobufs.KeyValue{},
				},
			},
			wantErr:    true,
			wantErrMsg: "cannot set instance uid to empty value",
		},
		{
			name: "can successfully start the client",
			fields: fields{
				logger: &NoopLogger{},
				agentDescription: &protobufs.AgentDescription{
					IdentifyingAttributes: []*protobufs.KeyValue{},
				},
				state: &agentState{
					InstanceId: newInstanceId(),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent := &Agent{
				logger:             tt.fields.logger,
				stateManager:       newStateManager(tt.fields.logger),
				agentType:          tt.fields.agentType,
				agentVersion:       tt.fields.agentVersion,
				serverURL:          tt.fields.serverURL,
				configUpdated:      tt.fields.configUpdated,
				agentDescription:   tt.fields.agentDescription,
				opampClient:        tt.fields.opampClient,
				remoteConfigStatus: tt.fields.remoteConfigStatus,
			}
			agent.stateManager.SetState(tt.fields.state)
			err := agent.Start(tt.fields.serverURL)
			if (err != nil) != tt.wantErr {
				t.Errorf("Agent.Start() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && err.Error() != tt.wantErrMsg {
				t.Errorf("Agent.Start() error msg = %v, wantErrMsg %v", err.Error(), tt.wantErrMsg)
			}
		})
	}
}

func TestAgent_loadState(t *testing.T) {
	type fields struct {
		logger       types.Logger
	}
	tests := []struct {
		name       string
		fields     fields
		beforeHook func(t *testing.T, m *stateManager)
	}{
		{
			name: "generates state when none exists",
			beforeHook: func(t *testing.T, m *stateManager) {
				err := m.Delete()
				require.NoError(t, err)
			},
			fields: fields{
				logger: &NoopLogger{},
			},
		},
		{
			name: "loads state when it exists",
			beforeHook: func(t *testing.T, m *stateManager) {
				m.SetState(newAgentState())
				err := m.Save()
				require.NoError(t, err)
			},
			fields: fields{
				logger: &NoopLogger{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent := &Agent{
				logger:       tt.fields.logger,
				stateManager: newStateManager(tt.fields.logger),
			}
			if tt.beforeHook != nil {
				tt.beforeHook(t, agent.stateManager)
			}
			err := agent.loadState()
			require.NoError(t, err)
			state := agent.stateManager.GetState()
			if state.InstanceId == "" {
				t.Error("Agent.loadState() instanceId is empty, want non-empty string")
			}
		})
	}
}

func Test_stringKeyValue(t *testing.T) {
	type args struct {
		key   string
		value string
	}
	tests := []struct {
		name string
		args args
		want *protobufs.KeyValue
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := stringKeyValue(tt.args.key, tt.args.value); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("stringKeyValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAgent_createAgentDescription(t *testing.T) {
	type fields struct {
		logger             types.Logger
		state              *agentState
		agentType          string
		agentVersion       string
		serverURL          string
		configUpdated      chan bool
		agentDescription   *protobufs.AgentDescription
		opampClient        client.OpAMPClient
		remoteConfigStatus *protobufs.RemoteConfigStatus
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent := &Agent{
				logger:             tt.fields.logger,
				agentType:          tt.fields.agentType,
				agentVersion:       tt.fields.agentVersion,
				serverURL:          tt.fields.serverURL,
				configUpdated:      tt.fields.configUpdated,
				agentDescription:   tt.fields.agentDescription,
				opampClient:        tt.fields.opampClient,
				remoteConfigStatus: tt.fields.remoteConfigStatus,
			}
			agent.stateManager.SetState(tt.fields.state)
			agent.createAgentDescription()
		})
	}
}

func TestAgent_updateAgentIdentity(t *testing.T) {
	type fields struct {
		logger             types.Logger
		state              *agentState
		agentType          string
		agentVersion       string
		serverURL          string
		configUpdated      chan bool
		agentDescription   *protobufs.AgentDescription
		opampClient        client.OpAMPClient
		remoteConfigStatus *protobufs.RemoteConfigStatus
	}
	type args struct {
		instanceId ulid.ULID
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent := &Agent{
				logger:             tt.fields.logger,
				agentType:          tt.fields.agentType,
				agentVersion:       tt.fields.agentVersion,
				serverURL:          tt.fields.serverURL,
				configUpdated:      tt.fields.configUpdated,
				agentDescription:   tt.fields.agentDescription,
				opampClient:        tt.fields.opampClient,
				remoteConfigStatus: tt.fields.remoteConfigStatus,
			}
			agent.stateManager.SetState(tt.fields.state)
			agent.updateAgentIdentity(tt.args.instanceId)
		})
	}
}

func TestAgent_composeEffectiveConfig(t *testing.T) {
	type fields struct {
		logger             types.Logger
		state              *agentState
		agentType          string
		agentVersion       string
		serverURL          string
		configUpdated      chan bool
		agentDescription   *protobufs.AgentDescription
		opampClient        client.OpAMPClient
		remoteConfigStatus *protobufs.RemoteConfigStatus
	}
	tests := []struct {
		name   string
		fields fields
		want   *protobufs.EffectiveConfig
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent := &Agent{
				logger:             tt.fields.logger,
				agentType:          tt.fields.agentType,
				agentVersion:       tt.fields.agentVersion,
				serverURL:          tt.fields.serverURL,
				configUpdated:      tt.fields.configUpdated,
				agentDescription:   tt.fields.agentDescription,
				opampClient:        tt.fields.opampClient,
				remoteConfigStatus: tt.fields.remoteConfigStatus,
			}
			agent.stateManager.SetState(tt.fields.state)
			state := agent.stateManager.GetState()
			if got := state.EffectiveConfig; !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Agent.composeEffectiveConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAgent_effectiveConfigMap(t *testing.T) {
	type fields struct {
		logger             types.Logger
		state              *agentState
		agentType          string
		agentVersion       string
		serverURL          string
		configUpdated      chan bool
		agentDescription   *protobufs.AgentDescription
		opampClient        client.OpAMPClient
		remoteConfigStatus *protobufs.RemoteConfigStatus
	}
	tests := []struct {
		name    string
		fields  fields
		want    map[string]interface{}
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent := &Agent{
				logger:             tt.fields.logger,
				agentType:          tt.fields.agentType,
				agentVersion:       tt.fields.agentVersion,
				serverURL:          tt.fields.serverURL,
				configUpdated:      tt.fields.configUpdated,
				agentDescription:   tt.fields.agentDescription,
				opampClient:        tt.fields.opampClient,
				remoteConfigStatus: tt.fields.remoteConfigStatus,
			}
			agent.stateManager.SetState(tt.fields.state)
			state := agent.stateManager.GetState()
			got, err := state.EffectiveConfig.composeOtConfig()
			if (err != nil) != tt.wantErr {
				t.Errorf("Agent.effectiveConfigMap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Agent.effectiveConfigMap() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_agentConfigFileSlice_Less(t *testing.T) {
	type args struct {
		i int
		j int
	}
	tests := []struct {
		name string
		a    agentConfigFileSlice
		args args
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.a.Less(tt.args.i, tt.args.j); got != tt.want {
				t.Errorf("agentConfigFileSlice.Less() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_agentConfigFileSlice_Swap(t *testing.T) {
	type args struct {
		i int
		j int
	}
	tests := []struct {
		name string
		a    agentConfigFileSlice
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.a.Swap(tt.args.i, tt.args.j)
		})
	}
}

func Test_agentConfigFileSlice_Len(t *testing.T) {
	tests := []struct {
		name string
		a    agentConfigFileSlice
		want int
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.a.Len(); got != tt.want {
				t.Errorf("agentConfigFileSlice.Len() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAgent_applyRemoteConfig(t *testing.T) {
	type fields struct {
		logger             types.Logger
		state              *agentState
		agentType          string
		agentVersion       string
		serverURL          string
		configUpdated      chan bool
		agentDescription   *protobufs.AgentDescription
		opampClient        client.OpAMPClient
		remoteConfigStatus *protobufs.RemoteConfigStatus
	}
	type args struct {
		config *protobufs.AgentRemoteConfig
	}
	tests := []struct {
		name              string
		fields            fields
		args              args
		wantConfigChanged bool
		wantErr           bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent := &Agent{
				logger:             tt.fields.logger,
				agentType:          tt.fields.agentType,
				agentVersion:       tt.fields.agentVersion,
				serverURL:          tt.fields.serverURL,
				configUpdated:      tt.fields.configUpdated,
				agentDescription:   tt.fields.agentDescription,
				opampClient:        tt.fields.opampClient,
				remoteConfigStatus: tt.fields.remoteConfigStatus,
			}
			agent.stateManager.SetState(tt.fields.state)
			gotConfigChanged, err := agent.applyRemoteConfig(tt.args.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("Agent.applyRemoteConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotConfigChanged != tt.wantConfigChanged {
				t.Errorf("Agent.applyRemoteConfig() = %v, want %v", gotConfigChanged, tt.wantConfigChanged)
			}
		})
	}
}

func TestAgent_Shutdown(t *testing.T) {
	type fields struct {
		logger             types.Logger
		state              *agentState
		agentType          string
		agentVersion       string
		serverURL          string
		configUpdated      chan bool
		agentDescription   *protobufs.AgentDescription
		opampClient        client.OpAMPClient
		remoteConfigStatus *protobufs.RemoteConfigStatus
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent := &Agent{
				logger:             tt.fields.logger,
				agentType:          tt.fields.agentType,
				agentVersion:       tt.fields.agentVersion,
				serverURL:          tt.fields.serverURL,
				configUpdated:      tt.fields.configUpdated,
				agentDescription:   tt.fields.agentDescription,
				opampClient:        tt.fields.opampClient,
				remoteConfigStatus: tt.fields.remoteConfigStatus,
			}
			agent.stateManager.SetState(tt.fields.state)
			agent.Shutdown()
		})
	}
}

func TestAgent_onMessage(t *testing.T) {
	type fields struct {
		logger             types.Logger
		state              *agentState
		agentType          string
		agentVersion       string
		serverURL          string
		configUpdated      chan bool
		agentDescription   *protobufs.AgentDescription
		opampClient        client.OpAMPClient
		remoteConfigStatus *protobufs.RemoteConfigStatus
	}
	type args struct {
		ctx context.Context
		msg *types.MessageData
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent := &Agent{
				logger:             tt.fields.logger,
				agentType:          tt.fields.agentType,
				agentVersion:       tt.fields.agentVersion,
				serverURL:          tt.fields.serverURL,
				configUpdated:      tt.fields.configUpdated,
				agentDescription:   tt.fields.agentDescription,
				opampClient:        tt.fields.opampClient,
				remoteConfigStatus: tt.fields.remoteConfigStatus,
			}
			agent.stateManager.SetState(tt.fields.state)
			agent.onMessage(tt.args.ctx, tt.args.msg)
		})
	}
}
