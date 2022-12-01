// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package opampprovider

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_stateManager_StatePath(t *testing.T) {
	type fields struct {
		statePath string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "returns the default state path when field is empty",
			want: func() string {
				return filepath.Join(os.TempDir(), "ot-opamp-state")
			}(),
		},
		{
			name: "returns the field state path when field is set",
			fields: fields{
				statePath: "/tmp/foo",
			},
			want: "/tmp/foo",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &stateManager{
				statePath: tt.fields.statePath,
			}
			if got := m.StatePath(); got != tt.want {
				t.Errorf("stateManager.StatePath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_stateManager_Load(t *testing.T) {
	type fields struct {
		statePath string
	}
	tests := []struct {
		name       string
		platform   string
		fields     fields
		beforeHook func(*testing.T, *stateManager)
		want       *agentState
		wantErr    bool
		wantErrMsg string
	}{
		{
			name:     "returns error when read file fails (linux)",
			platform: "linux",
			fields: fields{
				statePath: "/foo/bar/fake/path",
			},
			wantErr:    true,
			wantErrMsg: "open /foo/bar/fake/path: no such file or directory",
		},
		{
			name:     "returns error when read file fails (windows)",
			platform: "windows",
			fields: fields{
				statePath: "C:\\foo\\bar\\fake\\path",
			},
			wantErr:    true,
			wantErrMsg: "open C:\\foo\\bar\\fake\\path: The system cannot find the path specified.",
		},
		{
			name: "returns error when json unmarshaling fails",
			fields: fields{
				statePath: filepath.Join(t.TempDir(), "ot-opamp-state-test"),
			},
			beforeHook: func(t *testing.T, m *stateManager) {
				f, err := os.OpenFile(m.StatePath(), os.O_RDWR|os.O_CREATE, 0755)
				require.NoError(t, err)
				t.Cleanup(func() {
					f.Close()
				})

				_, err = f.Write([]byte("{42"))
				require.NoError(t, err)
			},
			wantErr:    true,
			wantErrMsg: "invalid character '4' looking for beginning of object key string",
		},
		{
			name: "returns error when validation fails",
			fields: fields{
				statePath: filepath.Join(t.TempDir(), "ot-opamp-state-test"),
			},
			beforeHook: func(t *testing.T, m *stateManager) {
				f, err := os.OpenFile(m.StatePath(), os.O_RDWR|os.O_CREATE, 0755)
				require.NoError(t, err)
				t.Cleanup(func() {
					f.Close()
				})

				state := agentState{}
				bytes, err := json.Marshal(state)
				require.NoError(t, err)

				_, err = f.Write(bytes)
				require.NoError(t, err)
			},
			wantErr:    true,
			wantErrMsg: "instance id is empty",
		},
		{
			name: "loads state without errors",
			fields: fields{
				statePath: filepath.Join(t.TempDir(), "ot-opamp-state-test"),
			},
			beforeHook: func(t *testing.T, m *stateManager) {
				state := &agentState{
					InstanceId: "foo",
				}
				m.SetState(state)
				require.NoError(t, m.Save())
			},
			want: &agentState{
				InstanceId: "foo",
			},
		},
	}
	for _, tt := range tests {
		if tt.platform != "" && tt.platform != runtime.GOOS {
			continue
		}
		t.Run(tt.name, func(t *testing.T) {
			m := &stateManager{
				statePath: tt.fields.statePath,
				logger:    &NoopLogger{},
			}
			if tt.beforeHook != nil {
				tt.beforeHook(t, m)
			}
			err := m.Load()
			got := m.GetState()
			if (err != nil) != tt.wantErr {
				t.Errorf("stateManager.Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && err.Error() != tt.wantErrMsg {
				t.Errorf("stateManager.Load() error msg = %v, wantErrMsg %v", err.Error(), tt.wantErrMsg)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("stateManager.Load() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_stateManager_Save(t *testing.T) {
	type fields struct {
		statePath string
	}
	type args struct {
		state *agentState
	}
	tests := []struct {
		name       string
		platform   string
		fields     fields
		args       args
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "returns error when validation fails",
			fields: fields{
				statePath: filepath.Join(t.TempDir(), "ot-opamp-state-test"),
			},
			args: args{
				state: &agentState{},
			},
			wantErr:    true,
			wantErrMsg: "instance id is empty",
		},
		{
			name:     "returns error when open file fails (linux)",
			platform: "linux",
			fields: fields{
				statePath: "/foo/bar/fake/path",
			},
			args: args{
				state: &agentState{
					InstanceId: newInstanceId(),
				},
			},
			wantErr:    true,
			wantErrMsg: "open /foo/bar/fake/path: no such file or directory",
		},
		{
			name:     "returns error when open file fails (windows)",
			platform: "windows",
			fields: fields{
				statePath: "C:\\foo\\bar\\fake\\path",
			},
			args: args{
				state: &agentState{
					InstanceId: newInstanceId(),
				},
			},
			wantErr:    true,
			wantErrMsg: "open C:\\foo\\bar\\fake\\path: The system cannot find the path specified.",
		},
		{
			name: "saves state without errors",
			args: args{
				state: &agentState{
					InstanceId: newInstanceId(),
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		if tt.platform != "" && tt.platform != runtime.GOOS {
			continue
		}
		t.Run(tt.name, func(t *testing.T) {
			m := &stateManager{
				statePath: tt.fields.statePath,
				logger:    &NoopLogger{},
			}
			m.SetState(tt.args.state)
			err := m.Save()
			if (err != nil) != tt.wantErr {
				t.Errorf("stateManager.Save() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && err.Error() != tt.wantErrMsg {
				t.Errorf("stateManager.Save() error msg = %v, wantErrMsg %v", err.Error(), tt.wantErrMsg)
			}
		})
	}
}
