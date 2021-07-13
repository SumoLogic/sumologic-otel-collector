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
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/config"
	"go.uber.org/zap"
)

const (
	uuidRegex = "[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}"
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
			Name: "no_credentials_causes_error",
			Config: func() *Config {
				cfg := createDefaultConfig().(*Config)
				cfg.CollectorName = "collector_name"
				return cfg
			}(),
			WantErr: true,
		},
		{
			Name: "basic",
			Config: func() *Config {
				cfg := createDefaultConfig().(*Config)
				cfg.CollectorName = "collector_name"
				cfg.Credentials.AccessID = "access_id_123456"
				cfg.Credentials.AccessKey = "access_key_123456"
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
		// TODO Add payload verification - verify if collectorName is set properly
		if req.URL.Path == registerUrl {
			_, err := w.Write([]byte(`{
				"collectorCredentialId": "aaaaaaaaaaaaaaaaaaaa",
				"collectorCredentialKey": "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
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
	cfg.Credentials.AccessID = "dummy_access_id"
	cfg.Credentials.AccessKey = "dummy_access_key"

	se, err := newSumologicExtension(cfg, zap.NewNop())
	require.NoError(t, err)
	require.NoError(t, se.Start(context.Background(), componenttest.NewNopHost()))
	assert.NotEmpty(t, se.registrationInfo.CollectorCredentialId)
	assert.NotEmpty(t, se.registrationInfo.CollectorCredentialKey)
	assert.NotEmpty(t, se.registrationInfo.CollectorId)
	require.NoError(t, se.Shutdown(context.Background()))
}

func TestStoreCredentials(t *testing.T) {
	getServer := func() *httptest.Server {
		// TODO Add payload verification - verify if collectorName is set properly
		return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			if req.URL.Path == registerUrl {
				_, err := w.Write([]byte(`{
				"collectorCredentialId": "aaaaaaaaaaaaaaaaaaaa",
				"collectorCredentialKey": "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
				"collectorId": "000000000FFFFFFF"
			}`))

				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				return
			}
		}))
	}

	getConfig := func(url string) *Config {
		cfg := createDefaultConfig().(*Config)
		cfg.CollectorName = "collector_name"
		cfg.ExtensionSettings = config.ExtensionSettings{}
		cfg.ApiBaseUrl = url
		cfg.Credentials.AccessID = "dummy_access_id"
		cfg.Credentials.AccessKey = "dummy_access_key"
		return cfg
	}

	t.Run("dir does not exist", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "otelcol-sumo-store-credentials-test-*")
		require.NoError(t, err)
		t.Cleanup(func() { os.RemoveAll(dir) })

		srv := getServer()
		cfg := getConfig(srv.URL)
		cfg.CollectorCredentialsDirectory = dir

		// Ensure the directory doesn't exist before running the extension
		require.NoError(t, os.RemoveAll(dir))

		se, err := newSumologicExtension(cfg, zap.NewNop())
		require.NoError(t, err)
		key := createHashKey(cfg)
		fileName, err := hash(key)
		require.NoError(t, err)
		credsPath := path.Join(dir, fileName)
		require.NoFileExists(t, credsPath)
		require.NoError(t, se.Start(context.Background(), componenttest.NewNopHost()))
		require.NoError(t, se.Shutdown(context.Background()))
		require.FileExists(t, credsPath)

		// To make sure that collector is using credentials file, turn off the mock registration server
		srv.Close()
		require.NoError(t, se.Start(context.Background(), componenttest.NewNopHost()))
	})

	t.Run("dir exists before launch with 600 permissions", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "otelcol-sumo-store-credentials-test-*")
		require.NoError(t, err)
		t.Cleanup(func() { os.RemoveAll(dir) })

		srv := getServer()
		cfg := getConfig(srv.URL)
		cfg.CollectorCredentialsDirectory = dir

		// Ensure the directory has 600 permissions
		require.NoError(t, os.Chmod(dir, 0600))

		se, err := newSumologicExtension(cfg, zap.NewNop())
		require.NoError(t, err)
		key := createHashKey(cfg)
		fileName, err := hash(key)
		require.NoError(t, err)
		credsPath := path.Join(dir, fileName)
		require.NoFileExists(t, credsPath)
		require.NoError(t, se.Start(context.Background(), componenttest.NewNopHost()))
		require.NoError(t, se.Shutdown(context.Background()))
		require.FileExists(t, credsPath)

		// To make sure that collector is using credentials file, turn off the mock registration server
		srv.Close()
		require.NoError(t, se.Start(context.Background(), componenttest.NewNopHost()))
	})

	t.Run("ensure dir gets created with 700 permissions", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "otelcol-sumo-store-credentials-test-*")
		require.NoError(t, err)
		t.Cleanup(func() { os.RemoveAll(dir) })

		srv := getServer()
		cfg := getConfig(srv.URL)
		cfg.CollectorCredentialsDirectory = dir

		fi, err := os.Stat(dir)
		require.NoError(t, err)

		// Chceck that directory has 700 permissions
		require.EqualValues(t, 0700, fi.Mode().Perm())

		se, err := newSumologicExtension(cfg, zap.NewNop())
		require.NoError(t, err)
		key := createHashKey(cfg)
		fileName, err := hash(key)
		require.NoError(t, err)
		credsPath := path.Join(dir, fileName)
		require.NoFileExists(t, credsPath)
		require.NoError(t, se.Start(context.Background(), componenttest.NewNopHost()))
		require.NoError(t, se.Shutdown(context.Background()))
		require.FileExists(t, credsPath)

		// To make sure that collector is using credentials file, turn off the mock registration server
		srv.Close()
		require.NoError(t, se.Start(context.Background(), componenttest.NewNopHost()))
	})
}

func TestRegisterEmptyCollectorName(t *testing.T) {
	hostname, err := os.Hostname()
	require.NoError(t, err)
	// TODO Add payload verification - verify if collectorName is set properly
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path == registerUrl {
			_, err := w.Write([]byte(`{
				"collectorCredentialId": "aaaaaaaaaaaaaaaaaaaa",
				"collectorCredentialKey": "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
				"collectorId": "000000000FFFFFFF"
			}`))

			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			return
		}
	}))

	dir, err := os.MkdirTemp("", "otelcol-sumo-store-credentials-test-*")
	t.Cleanup(func() {
		srv.Close()
		os.RemoveAll(dir)
	})
	require.NoError(t, err)

	cfg := createDefaultConfig().(*Config)
	cfg.CollectorName = ""
	cfg.ExtensionSettings = config.ExtensionSettings{}
	cfg.ApiBaseUrl = srv.URL
	cfg.Credentials.AccessID = "dummy_access_id"
	cfg.Credentials.AccessKey = "dummy_access_key"
	cfg.CollectorCredentialsDirectory = dir

	se, err := newSumologicExtension(cfg, zap.NewNop())
	require.NoError(t, err)
	require.NoError(t, se.Start(context.Background(), componenttest.NewNopHost()))
	regexPattern := fmt.Sprintf("%s-%s", hostname, uuidRegex)
	matched, err := regexp.MatchString(regexPattern, se.collectorName)
	require.NoError(t, err)
	assert.True(t, matched)
}

func TestRegisterEmptyCollectorNameClobber(t *testing.T) {
	hostname, err := os.Hostname()
	require.NoError(t, err)
	// TODO Add payload verification - verify if collectorName is set properly
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path == registerUrl {
			_, err := w.Write([]byte(`{
				"collectorCredentialId": "aaaaaaaaaaaaaaaaaaaa",
				"collectorCredentialKey": "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
				"collectorId": "000000000FFFFFFF"
			}`))

			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			return
		}
	}))

	dir, err := os.MkdirTemp("", "otelcol-sumo-store-credentials-test-*")
	t.Cleanup(func() {
		srv.Close()
		os.RemoveAll(dir)
	})
	require.NoError(t, err)

	cfg := createDefaultConfig().(*Config)
	cfg.CollectorName = ""
	cfg.ExtensionSettings = config.ExtensionSettings{}
	cfg.ApiBaseUrl = srv.URL
	cfg.Credentials.AccessID = "dummy_access_id"
	cfg.Credentials.AccessKey = "dummy_access_key"
	cfg.CollectorCredentialsDirectory = dir
	cfg.Clobber = true

	se, err := newSumologicExtension(cfg, zap.NewNop())
	require.NoError(t, err)
	require.NoError(t, se.Start(context.Background(), componenttest.NewNopHost()))
	require.NoError(t, se.Shutdown(context.Background()))
	assert.NotEmpty(t, se.collectorName)
	regexPattern := fmt.Sprintf("%s-%s", hostname, uuidRegex)
	matched, err := regexp.MatchString(regexPattern, se.collectorName)
	require.NoError(t, err)
	assert.True(t, matched)
	colCreds, err := se.credentialsGetter.GetStoredCredentials(se.hashKey)
	require.NoError(t, err)
	colName := colCreds.CollectorName
	se, err = newSumologicExtension(cfg, zap.NewNop())
	require.NoError(t, err)
	require.NoError(t, se.Start(context.Background(), componenttest.NewNopHost()))
	assert.Equal(t, se.collectorName, colName)
}

func TestCollectorSendsBasicAuthHeadersOnRegistration(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		assert.Empty(t, req.Header.Get("accessid"))
		assert.Empty(t, req.Header.Get("accesskey"))
		authHeader := req.Header.Get("Authorization")
		token := base64.StdEncoding.EncodeToString(
			[]byte("dummy_access_id:dummy_access_key"),
		)
		assert.Equal(t, "Basic "+token, authHeader)

		if req.URL.Path == registerUrl {
			_, err := w.Write([]byte(`{
				"collectorCredentialId": "aaaaaaaaaaaaaaaaaaaa",
				"collectorCredentialKey": "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
				"collectorId": "000000000FFFFFFF"
			}`))

			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				t.FailNow()
				return
			}
			return
		}
	}))
	t.Cleanup(func() {
		srv.Close()
	})

	dir, err := os.MkdirTemp("", "otelcol-sumo-store-credentials-test-*")
	require.NoError(t, err)
	t.Cleanup(func() {
		os.RemoveAll(dir)
	})

	cfg := createDefaultConfig().(*Config)
	cfg.CollectorName = ""
	cfg.ApiBaseUrl = srv.URL
	cfg.Credentials.AccessID = "dummy_access_id"
	cfg.Credentials.AccessKey = "dummy_access_key"
	cfg.CollectorCredentialsDirectory = dir

	se, err := newSumologicExtension(cfg, zap.NewNop())
	require.NoError(t, err)
	require.NoError(t, se.Start(context.Background(), componenttest.NewNopHost()))
	require.NoError(t, se.Shutdown(context.Background()))
}
