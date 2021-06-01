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

package sumologicextension

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/config"
	"go.uber.org/zap"
)

func TestBasicExtensionConstruction(t *testing.T) {
	testcases := []struct {
		Name    string
		Config  *Config
		WantErr bool
	}{
		{
			Name:    "no_collector_name_causes_error",
			Config:  createDefaultConfig().(*Config),
			WantErr: true,
		},
		{
			Name: "basic",
			Config: func() *Config {
				cfg := createDefaultConfig().(*Config)
				cfg.CollectorName = "collector_name"
				return cfg
			}(),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.Name, func(t *testing.T) {
			se, err := newSumologicExtension(tc.Config, zap.NewNop())
			if tc.WantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, se)
			}
		})
	}
}

func TestBasicStart(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path == registerUrl {
			_, err := w.Write([]byte(`{
				"collectorCredentialId": "aaaaaaaaaaaaaaaaaaaa",
				"collectorCredentialKey": "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
				"collectorId": "000000000FFFFFFF"
			}`))

			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			return
		}
	}))
	t.Cleanup(func() {
		srv.Close()
	})

	cfg := createDefaultConfig().(*Config)
	cfg.CollectorName = "collector_name"
	cfg.ExtensionSettings = config.ExtensionSettings{}
	cfg.ApiBaseUrl = srv.URL
	se, err := newSumologicExtension(cfg, zap.NewNop())
	require.NoError(t, err)
	require.NoError(t, se.Start(context.Background(), componenttest.NewNopHost()))
	assert.NotEmpty(t, se.registrationInfo.CollectorCredentialId)
	assert.NotEmpty(t, se.registrationInfo.CollectorCredentialKey)
	assert.NotEmpty(t, se.registrationInfo.CollectorId)
	require.NoError(t, se.Shutdown(context.Background()))
}
