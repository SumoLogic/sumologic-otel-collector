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
	"math/rand"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/oklog/ulid/v2"
	"github.com/open-telemetry/opamp-go/client/types"
)

type EmptyInstanceIdError struct{}

func (e *EmptyInstanceIdError) Error() string {
	return "instance id is empty"
}

type agentState struct {
	InstanceId      string `json:"instance_id"`
	EffectiveConfig config `json:"effective_config"`
}

func newAgentState() *agentState {
	return &agentState{
		InstanceId:      newInstanceId(),
		EffectiveConfig: config{},
	}
}

func (s *agentState) validate() error {
	if s.InstanceId == "" {
		return &EmptyInstanceIdError{}
	}
	return nil
}

func newInstanceId() string {
	entropy := ulid.Monotonic(rand.New(rand.NewSource(0)), 0)
	return ulid.MustNew(ulid.Timestamp(time.Now()), entropy).String()
}

type stateManager struct {
	logger    types.Logger
	mu        sync.RWMutex
	statePath string
	state     *agentState
}

func newStateManager(logger types.Logger) *stateManager {
	return &stateManager{
		logger: logger,
		state:  &agentState{},
	}
}

// TODO: set default state path to something other than a temp dir
func (m *stateManager) StatePath() string {
	if m.statePath == "" {
		return filepath.Join(os.TempDir(), "ot-opamp-state")
	}
	return m.statePath
}

func (m *stateManager) GetState() *agentState {
	return m.state
}

func (m *stateManager) SetState(state *agentState) {
	m.state = state
}

func (m *stateManager) Load() error {
	m.logger.Debugf("Loading OpAMP state from path: %s.", m.StatePath())

	m.mu.Lock()
	data, err := os.ReadFile(m.StatePath())
	if err != nil {
		m.mu.Unlock()
		return err
	}
	m.mu.Unlock()

	var state *agentState
	if err := json.Unmarshal(data, &state); err != nil {
		return err
	}

	if err := state.validate(); err != nil {
		return err
	}

	m.SetState(state)

	return nil
}

func (m *stateManager) Save() error {
	m.logger.Debugf("Saving OpAMP state to path: %s.", m.StatePath())

	state := m.GetState()

	if err := state.validate(); err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	f, err := os.OpenFile(m.StatePath(), os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return err
	}
	defer f.Close()

	bytes, err := json.Marshal(state)
	if err != nil {
		return err
	}

	if _, err = f.Write(bytes); err != nil {
		return err
	}

	return nil
}
