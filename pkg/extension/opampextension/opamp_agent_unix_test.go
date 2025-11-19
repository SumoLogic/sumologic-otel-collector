//go:build !windows

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
	"os"
	"path/filepath"
	"testing"

	"github.com/open-telemetry/opamp-go/protobufs"
	"github.com/stretchr/testify/assert"
)

func TestApplyRemoteConfigUnix(t *testing.T) {
	tests := []struct {
		name         string
		file         string
		expectError  bool
		errorMessage string
	}{
		{"ApplyChronyConfig", "testdata/opamp.d/opamp-chrony-receiver-config.yaml", false, ""},
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
