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
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"regexp"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/config"
	"go.uber.org/zap"

	"github.com/SumoLogic/sumologic-otel-collector/pkg/extension/sumologicextension/api"
	"github.com/SumoLogic/sumologic-otel-collector/pkg/extension/sumologicextension/credentials"
)

const (
	uuidRegex = "[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}"
)

func TestBasicExtensionConstruction(t *testing.T) {
	t.Parallel()

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
				cfg.Credentials.InstallToken = "install_token_123456"
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
	t.Parallel()

	srv := httptest.NewServer(func() http.HandlerFunc {
		var reqCount int32

		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			// TODO Add payload verification - verify if collectorName is set properly
			reqNum := atomic.AddInt32(&reqCount, 1)

			switch reqNum {

			// register
			case 1:
				require.Equal(t, registerUrl, req.URL.Path)
				_, err := w.Write([]byte(`{
					"collectorCredentialId": "collectorId",
					"collectorCredentialKey": "collectorKey",
					"collectorId": "id"
				}`))
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
				}

			// heartbeat
			case 2:
				assert.Equal(t, heartbeatUrl, req.URL.Path)
				w.WriteHeader(204)

			// should not produce any more requests
			default:
				w.WriteHeader(http.StatusInternalServerError)
			}
		})
	}())
	t.Cleanup(func() { srv.Close() })

	dir, err := os.MkdirTemp("", "otelcol-sumo-store-credentials-test-*")
	require.NoError(t, err)
	t.Cleanup(func() { os.RemoveAll(dir) })

	cfg := createDefaultConfig().(*Config)
	cfg.CollectorName = "collector_name"
	cfg.ExtensionSettings = config.ExtensionSettings{}
	cfg.ApiBaseUrl = srv.URL
	cfg.Credentials.InstallToken = "dummy_install_token"
	cfg.CollectorCredentialsDirectory = dir

	se, err := newSumologicExtension(cfg, zap.NewNop())
	require.NoError(t, err)
	require.NoError(t, se.Start(context.Background(), componenttest.NewNopHost()))
	assert.NotEmpty(t, se.registrationInfo.CollectorCredentialId)
	assert.NotEmpty(t, se.registrationInfo.CollectorCredentialKey)
	assert.NotEmpty(t, se.registrationInfo.CollectorId)
	require.NoError(t, se.Shutdown(context.Background()))
}

func TestStoreCredentials(t *testing.T) {
	t.Parallel()

	getServer := func() *httptest.Server {
		var reqCount int32

		return httptest.NewServer(http.HandlerFunc(
			func(w http.ResponseWriter, req *http.Request) {
				// TODO Add payload verification - verify if collectorName is set properly
				reqNum := atomic.AddInt32(&reqCount, 1)

				switch reqNum {

				// register
				case 1:
					require.Equal(t, registerUrl, req.URL.Path)
					_, err := w.Write([]byte(`{
						"collectorCredentialId": "collectorId",
						"collectorCredentialKey": "collectorKey",
						"collectorId": "id"
					}`))
					if err != nil {
						w.WriteHeader(http.StatusInternalServerError)
					}

				// heartbeat
				case 2:
					assert.Equal(t, heartbeatUrl, req.URL.Path)
					w.WriteHeader(204)

				// should not produce any more requests
				default:
					w.WriteHeader(http.StatusInternalServerError)
				}
			}))
	}

	getConfig := func(url string) *Config {
		cfg := createDefaultConfig().(*Config)
		cfg.CollectorName = "collector_name"
		cfg.ExtensionSettings = config.ExtensionSettings{}
		cfg.ApiBaseUrl = url
		cfg.Credentials.InstallToken = "dummy_install_token"
		return cfg
	}

	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	t.Run("dir does not exist", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "otelcol-sumo-store-credentials-test-*")
		require.NoError(t, err)
		t.Cleanup(func() { os.RemoveAll(dir) })

		srv := getServer()
		t.Cleanup(func() { srv.Close() })

		cfg := getConfig(srv.URL)
		cfg.CollectorCredentialsDirectory = dir

		// Ensure the directory doesn't exist before running the extension
		require.NoError(t, os.RemoveAll(dir))

		se, err := newSumologicExtension(cfg, logger)
		require.NoError(t, err)
		key := createHashKey(cfg)
		fileName, err := credentials.HashKeyToFilename(key)
		require.NoError(t, err)
		credsPath := path.Join(dir, fileName)
		require.NoFileExists(t, credsPath)
		require.NoError(t, se.Start(context.Background(), componenttest.NewNopHost()))
		require.NoError(t, se.Shutdown(context.Background()))
		require.FileExists(t, credsPath)
	})

	t.Run("dir exists before launch with 600 permissions", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "otelcol-sumo-store-credentials-test-*")
		require.NoError(t, err)
		t.Cleanup(func() { os.RemoveAll(dir) })

		srv := getServer()
		t.Cleanup(func() { srv.Close() })

		cfg := getConfig(srv.URL)
		cfg.CollectorCredentialsDirectory = dir

		// Ensure the directory has 600 permissions
		require.NoError(t, os.Chmod(dir, 0600))

		se, err := newSumologicExtension(cfg, zap.NewNop())
		require.NoError(t, err)
		key := createHashKey(cfg)
		fileName, err := credentials.HashKeyToFilename(key)
		require.NoError(t, err)
		credsPath := path.Join(dir, fileName)
		require.NoFileExists(t, credsPath)
		require.NoError(t, se.Start(context.Background(), componenttest.NewNopHost()))
		require.NoError(t, se.Shutdown(context.Background()))
		require.FileExists(t, credsPath)
	})

	t.Run("ensure dir gets created with 700 permissions", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "otelcol-sumo-store-credentials-test-*")
		require.NoError(t, err)
		t.Cleanup(func() { os.RemoveAll(dir) })

		srv := getServer()
		t.Cleanup(func() { srv.Close() })
		cfg := getConfig(srv.URL)
		cfg.CollectorCredentialsDirectory = dir

		fi, err := os.Stat(dir)
		require.NoError(t, err)

		// Chceck that directory has 700 permissions
		require.EqualValues(t, 0700, fi.Mode().Perm())

		se, err := newSumologicExtension(cfg, zap.NewNop())
		require.NoError(t, err)
		key := createHashKey(cfg)
		fileName, err := credentials.HashKeyToFilename(key)
		require.NoError(t, err)
		credsPath := path.Join(dir, fileName)
		require.NoFileExists(t, credsPath)
		require.NoError(t, se.Start(context.Background(), componenttest.NewNopHost()))
		require.NoError(t, se.Shutdown(context.Background()))
		require.FileExists(t, credsPath)
	})

	t.Run("by default use sha256 for hashing", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "otelcol-sumo-store-credentials-test-*")
		require.NoError(t, err)
		t.Cleanup(func() { os.RemoveAll(dir) })

		srv := getServer()
		t.Cleanup(func() { srv.Close() })

		cfg := getConfig(srv.URL)
		cfg.CollectorCredentialsDirectory = dir

		se, err := newSumologicExtension(cfg, zap.NewNop())
		require.NoError(t, err)
		key := createHashKey(cfg)
		fileName, err := credentials.HashKeyToFilename(key)
		require.NoError(t, err)
		fileNameSha256, err := credentials.HashKeyToFilenameWith(sha256.New(), key)
		require.NoError(t, err)
		require.Equal(t, fileName, fileNameSha256)

		credsPath := path.Join(dir, fileName)
		require.NoFileExists(t, credsPath)
		require.NoError(t, se.Start(context.Background(), componenttest.NewNopHost()))
		require.NoError(t, se.Shutdown(context.Background()))
		require.FileExists(t, credsPath)

		// Don't create md5 hashed credentials files anymore
		fileNameMd5, err := credentials.HashKeyToFilenameWith(md5.New(), key)
		require.NoError(t, err)
		credsPathMd5 := path.Join(dir, fileNameMd5)
		require.NoFileExists(t, credsPathMd5)
	})
}

func TestStoreCredentials_PreexistingCredentialsAreUsed(t *testing.T) {
	t.Parallel()

	var reqCount int32
	getServer := func() *httptest.Server {
		return httptest.NewServer(http.HandlerFunc(
			func(w http.ResponseWriter, req *http.Request) {
				reqNum := atomic.AddInt32(&reqCount, 1)

				switch reqNum {
				// heartbeat
				case 1:
					require.Equal(t, heartbeatUrl, req.URL.Path)
					w.WriteHeader(204)

				// should not produce any more requests
				default:
					w.WriteHeader(http.StatusInternalServerError)
				}
			}))
	}

	getConfig := func(url string) *Config {
		cfg := createDefaultConfig().(*Config)
		cfg.CollectorName = "collector_name"
		cfg.ExtensionSettings = config.ExtensionSettings{}
		cfg.ApiBaseUrl = url
		cfg.Credentials.InstallToken = "dummy_install_token"
		return cfg
	}

	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	dir, err := os.MkdirTemp("", "otelcol-sumo-store-credentials-test-*")
	require.NoError(t, err)
	t.Cleanup(func() { os.RemoveAll(dir) })
	t.Logf("Using dir: %s", dir)

	store, err := credentials.NewLocalFsStore(
		credentials.WithCredentialsDirectory(dir),
		credentials.WithLogger(logger),
	)
	require.NoError(t, err)

	srv := getServer()
	t.Cleanup(func() { srv.Close() })

	cfg := getConfig(srv.URL)
	cfg.CollectorCredentialsDirectory = dir

	hashKey := createHashKey(cfg)

	require.NoError(t,
		store.Store(hashKey, credentials.CollectorCredentials{
			CollectorName: "collector_name",
			Credentials: api.OpenRegisterResponsePayload{
				CollectorCredentialId:  "collectorId",
				CollectorCredentialKey: "collectorKey",
				CollectorId:            "id",
			},
		}),
	)

	se, err := newSumologicExtension(cfg, logger)
	require.NoError(t, err)

	fileName, err := credentials.HashKeyToFilename(hashKey)
	require.NoError(t, err)
	credsPath := path.Join(dir, fileName)
	// Credentials file exists before starting the extension because we created
	// it directly via store.Store()
	require.FileExists(t, credsPath)

	require.NoError(t, se.Start(context.Background(), componenttest.NewNopHost()))
	require.NoError(t, se.Shutdown(context.Background()))
	require.FileExists(t, credsPath)

	// Don't create md5 hashed credentials files anymore
	fileNameMd5, err := credentials.HashKeyToFilenameWith(md5.New(), hashKey)
	require.NoError(t, err)
	credsPathMd5 := path.Join(dir, fileNameMd5)
	require.NoFileExists(t, credsPathMd5)

	require.EqualValues(t, atomic.LoadInt32(&reqCount), 1)
}

func TestLocalFSCredentialsStore_WorkCorrectlyForMultipleExtensions(t *testing.T) {
	t.Parallel()

	getServer := func() *httptest.Server {
		var reqCount int32

		return httptest.NewServer(http.HandlerFunc(
			func(w http.ResponseWriter, req *http.Request) {
				// TODO Add payload verification - verify if collectorName is set properly
				reqNum := atomic.AddInt32(&reqCount, 1)

				switch reqNum {

				// register
				case 1:
					require.Equal(t, registerUrl, req.URL.Path)
					_, err := w.Write([]byte(`{
						"collectorCredentialId": "collectorId",
						"collectorCredentialKey": "collectorKey",
						"collectorId": "id"
					}`))
					if err != nil {
						w.WriteHeader(http.StatusInternalServerError)
					}

				// heartbeat
				case 2:
					assert.Equal(t, heartbeatUrl, req.URL.Path)
					w.WriteHeader(204)

				// should not produce any more requests
				default:
					w.WriteHeader(http.StatusInternalServerError)
				}
			}))
	}

	getConfig := func(url string) *Config {
		cfg := createDefaultConfig().(*Config)
		cfg.CollectorName = "collector_name"
		cfg.ExtensionSettings = config.ExtensionSettings{}
		cfg.ApiBaseUrl = url
		cfg.Credentials.InstallToken = "dummy_install_token"
		return cfg
	}

	getDir := func(t *testing.T) string {
		dir, err := os.MkdirTemp("", "otelcol-sumo-store-credentials-multiple-extensions-test-*")
		require.NoError(t, err)
		return dir
	}

	dir1 := getDir(t)
	t.Cleanup(func() { os.RemoveAll(dir1) })
	dir2 := getDir(t)
	t.Cleanup(func() { os.RemoveAll(dir2) })

	srv1 := getServer()
	t.Cleanup(func() { srv1.Close() })
	srv2 := getServer()
	t.Cleanup(func() { srv2.Close() })

	cfg1 := getConfig(srv1.URL)
	cfg1.CollectorCredentialsDirectory = dir1

	cfg2 := getConfig(srv2.URL)
	cfg2.CollectorCredentialsDirectory = dir2

	logger1, err := zap.NewDevelopment(zap.Fields(zap.Int("#", 1)))
	require.NoError(t, err)

	logger2, err := zap.NewDevelopment(zap.Fields(zap.Int("#", 2)))
	require.NoError(t, err)

	se1, err := newSumologicExtension(cfg1, logger1)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, se1.Shutdown(context.Background())) })
	fileName1, err := credentials.HashKeyToFilename(createHashKey(cfg1))
	require.NoError(t, err)
	credsPath1 := path.Join(dir1, fileName1)
	require.NoFileExists(t, credsPath1)
	require.NoError(t, se1.Start(context.Background(), componenttest.NewNopHost()))
	require.FileExists(t, credsPath1)

	se2, err := newSumologicExtension(cfg2, logger2)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, se2.Shutdown(context.Background())) })
	fileName2, err := credentials.HashKeyToFilename(createHashKey(cfg2))
	require.NoError(t, err)
	credsPath2 := path.Join(dir2, fileName2)
	require.NoFileExists(t, credsPath2)
	require.NoError(t, se2.Start(context.Background(), componenttest.NewNopHost()))
	require.FileExists(t, credsPath2)

	require.NotEqual(t, credsPath1, credsPath2,
		"credentials files should be different for configs with different apiBaseURLs",
	)
}

func TestRegisterEmptyCollectorName(t *testing.T) {
	t.Parallel()

	hostname, err := os.Hostname()
	require.NoError(t, err)
	srv := httptest.NewServer(func() http.HandlerFunc {
		var reqCount int32

		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			// TODO Add payload verification - verify if collectorName is set properly
			reqNum := atomic.AddInt32(&reqCount, 1)

			switch reqNum {

			// register
			case 1:
				require.Equal(t, registerUrl, req.URL.Path)

				authHeader := req.Header.Get("Authorization")
				assert.Equal(t, "Bearer dummy_install_token", authHeader,
					"collector didn't send correct Authorization header with registration request")

				_, err := w.Write([]byte(`{
					"collectorCredentialId": "aaaaaaaaaaaaaaaaaaaa",
					"collectorCredentialKey": "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
					"collectorId": "000000000FFFFFFF"
				}`))
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
				}

			// heartbeat
			case 2:
				assert.Equal(t, heartbeatUrl, req.URL.Path)
				w.WriteHeader(204)

			// should not produce any more requests
			default:
				w.WriteHeader(http.StatusInternalServerError)
			}
		})
	}())

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
	cfg.Credentials.InstallToken = "dummy_install_token"
	cfg.CollectorCredentialsDirectory = dir

	se, err := newSumologicExtension(cfg, zap.NewNop())
	require.NoError(t, err)
	require.NoError(t, se.Start(context.Background(), componenttest.NewNopHost()))
	regexPattern := fmt.Sprintf("%s-%s", hostname, uuidRegex)
	matched, err := regexp.MatchString(regexPattern, se.collectorName)
	require.NoError(t, err)
	assert.True(t, matched)
}

func TestRegisterEmptyCollectorNameForceRegistration(t *testing.T) {
	t.Parallel()

	hostname, err := os.Hostname()
	require.NoError(t, err)
	srv := httptest.NewServer(func() http.HandlerFunc {
		var reqCount int32

		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			// TODO Add payload verification - verify if collectorName is set properly
			reqNum := atomic.AddInt32(&reqCount, 1)

			switch reqNum {

			// register
			case 1:
				require.Equal(t, registerUrl, req.URL.Path)

				authHeader := req.Header.Get("Authorization")
				assert.Equal(t, "Bearer dummy_install_token", authHeader,
					"collector didn't send correct Authorization header with registration request")

				_, err := w.Write([]byte(`{
					"collectorCredentialId": "aaaaaaaaaaaaaaaaaaaa",
					"collectorCredentialKey": "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
					"collectorId": "000000000FFFFFFF",
					"collectorName": "hostname-test-123456123123"
				}`))
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
				}

			// register again because force registration was set
			case 2:
				require.Equal(t, registerUrl, req.URL.Path)

				authHeader := req.Header.Get("Authorization")
				assert.Equal(t, "Bearer dummy_install_token", authHeader,
					"collector didn't send correct Authorization header with registration request")

				_, err := w.Write([]byte(`{
					"collectorCredentialId": "aaaaaaaaaaaaaaaaaaaa",
					"collectorCredentialKey": "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
					"collectorId": "000000000FFFFFFF",
					"collectorName": "hostname-test-123456123123"
				}`))
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
				}

			// should not produce any more requests
			default:
				w.WriteHeader(http.StatusInternalServerError)
			}
		})
	}())

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
	cfg.Credentials.InstallToken = "dummy_install_token"
	cfg.CollectorCredentialsDirectory = dir
	cfg.ForceRegistration = true

	se, err := newSumologicExtension(cfg, zap.NewNop())
	require.NoError(t, err)
	require.NoError(t, se.Start(context.Background(), componenttest.NewNopHost()))
	require.NoError(t, se.Shutdown(context.Background()))
	assert.NotEmpty(t, se.collectorName)
	regexPattern := fmt.Sprintf("%s-%s", hostname, uuidRegex)
	matched, err := regexp.MatchString(regexPattern, se.collectorName)
	require.NoError(t, err)
	assert.True(t, matched)
	colCreds, err := se.credentialsStore.Get(se.hashKey)
	require.NoError(t, err)
	colName := colCreds.CollectorName
	se, err = newSumologicExtension(cfg, zap.NewNop())
	require.NoError(t, err)
	require.NoError(t, se.Start(context.Background(), componenttest.NewNopHost()))
	assert.Equal(t, se.collectorName, colName)
}

func TestCollectorSendsBasicAuthHeadersOnRegistration(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(func() http.HandlerFunc {
		var reqCount int32

		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			// TODO Add payload verification - verify if collectorName is set properly
			reqNum := atomic.AddInt32(&reqCount, 1)

			switch reqNum {

			// register
			case 1:
				require.Equal(t, registerUrl, req.URL.Path)

				authHeader := req.Header.Get("Authorization")
				assert.Equal(t, "Bearer dummy_install_token", authHeader,
					"collector didn't send correct Authorization header with registration request")

				_, err := w.Write([]byte(`{
					"collectorCredentialId": "aaaaaaaaaaaaaaaaaaaa",
					"collectorCredentialKey": "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
					"collectorId": "000000000FFFFFFF"
				}`))
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
				}

			// heartbeat
			case 2:
				assert.Equal(t, heartbeatUrl, req.URL.Path)
				w.WriteHeader(204)

			// should not produce any more requests
			default:
				w.WriteHeader(http.StatusInternalServerError)
			}
		})
	}())

	t.Cleanup(func() { srv.Close() })

	dir, err := os.MkdirTemp("", "otelcol-sumo-store-credentials-test-*")
	require.NoError(t, err)
	t.Cleanup(func() { os.RemoveAll(dir) })

	cfg := createDefaultConfig().(*Config)
	cfg.CollectorName = ""
	cfg.ApiBaseUrl = srv.URL
	cfg.Credentials.InstallToken = "dummy_install_token"
	cfg.CollectorCredentialsDirectory = dir

	se, err := newSumologicExtension(cfg, zap.NewNop())
	require.NoError(t, err)
	require.NoError(t, se.Start(context.Background(), componenttest.NewNopHost()))
	require.NoError(t, se.Shutdown(context.Background()))
}

func TestCollectorCheckingCredentialsFoundInLocalStorage(t *testing.T) {
	t.Parallel()

	dir, err := os.MkdirTemp("", "otelcol-sumo-store-credentials-test-*")
	require.NoError(t, err)
	t.Cleanup(func() { os.RemoveAll(dir) })

	cStore, err := credentials.NewLocalFsStore(
		credentials.WithCredentialsDirectory(dir),
		credentials.WithLogger(zap.NewNop()),
	)
	require.NoError(t, err)

	storeCredentials := func(t *testing.T, url string) {
		creds := credentials.CollectorCredentials{
			CollectorName: "test-name",
			Credentials: api.OpenRegisterResponsePayload{
				CollectorName:          "test-name",
				CollectorId:            "test-id",
				CollectorCredentialId:  "test-credential-id",
				CollectorCredentialKey: "test-credential-key",
			},
			ApiBaseUrl: url,
		}
		storageKey := createHashKey(&Config{
			CollectorName: "test-name",
			Credentials: accessCredentials{
				InstallToken: "dummy_install_token",
			},
			ApiBaseUrl: url,
		})
		t.Logf("Storing collector credentials in temp dir: %s", dir)
		require.NoError(t, cStore.Store(storageKey, creds))
	}

	testcases := []struct {
		name             string
		expectedReqCount int32
		srvFn            func() (*httptest.Server, *int32)
		configFn         func(url string) *Config
	}{
		{
			name:             "collector checks found credentials via heartbeat call - no registration is done",
			expectedReqCount: 2,
			srvFn: func() (*httptest.Server, *int32) {
				var reqCount int32

				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
						atomic.AddInt32(&reqCount, 1)

						require.NotEqual(t, registerUrl, req.URL.Path,
							"collector shouldn't call the register API when credentials locally retrieved")
						require.Equal(t, heartbeatUrl, req.URL.Path)
						w.WriteHeader(204)

						authHeader := req.Header.Get("Authorization")
						token := base64.StdEncoding.EncodeToString(
							[]byte("test-credential-id:test-credential-key"),
						)
						assert.Equal(t, "Basic "+token, authHeader,
							"collector didn't send correct Authorization header with heartbeat request")
					})),
					&reqCount
			},
			configFn: func(url string) *Config {
				cfg := createDefaultConfig().(*Config)
				cfg.CollectorName = "test-name"
				cfg.ApiBaseUrl = url
				cfg.Credentials.InstallToken = "dummy_install_token"
				cfg.CollectorCredentialsDirectory = dir
				return cfg
			},
		},
		{
			name:             "collector registers when no matching credentials are found in local storage",
			expectedReqCount: 2,
			srvFn: func() (*httptest.Server, *int32) {
				var reqCount int32

				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
						reqNum := atomic.AddInt32(&reqCount, 1)

						switch reqNum {

						// register
						case 1:
							require.Equal(t, registerUrl, req.URL.Path)

							authHeader := req.Header.Get("Authorization")
							assert.Equal(t, "Bearer dummy_install_token", authHeader,
								"collector didn't send correct Authorization header with registration request")

							_, err := w.Write([]byte(`{
							"collectorCredentialId": "aaaaaaaaaaaaaaaaaaaa",
							"collectorCredentialKey": "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
							"collectorId": "000000000FFFFFFF"
						}`))
							if err != nil {
								w.WriteHeader(http.StatusInternalServerError)
							}

						// heartbeat
						case 2:
							w.WriteHeader(204)

						// should not produce any more requests
						default:
							w.WriteHeader(http.StatusInternalServerError)
						}
					})),
					&reqCount
			},
			configFn: func(url string) *Config {
				cfg := createDefaultConfig().(*Config)
				cfg.CollectorName = "test-name-not-in-the-credentials-store"
				cfg.ApiBaseUrl = url
				cfg.Credentials.InstallToken = "dummy_install_token"
				cfg.CollectorCredentialsDirectory = dir
				return cfg
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			tc := tc

			srv, reqCount := tc.srvFn()
			t.Cleanup(func() { srv.Close() })

			cfg := tc.configFn(srv.URL)
			storeCredentials(t, srv.URL)

			logger, err := zap.NewDevelopment()
			require.NoError(t, err)

			se, err := newSumologicExtension(cfg, logger)
			require.NoError(t, err)
			require.NoError(t, se.Start(context.Background(), componenttest.NewNopHost()))

			if !assert.Eventually(t,
				func() bool {
					return atomic.LoadInt32(reqCount) == tc.expectedReqCount
				},
				5*time.Second, 100*time.Millisecond,
			) {
				t.Logf("the expected number of requests (%d) wasn't reached, only got %d",
					tc.expectedReqCount, atomic.LoadInt32(reqCount),
				)
			}

			require.NoError(t, se.Shutdown(context.Background()))
		})
	}
}

func TestRegisterEmptyCollectorNameWithBackoff(t *testing.T) {
	var retriesLimit int32 = 5
	t.Parallel()

	hostname, err := os.Hostname()
	require.NoError(t, err)
	srv := httptest.NewServer(func() http.HandlerFunc {
		var reqCount int32

		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			// TODO Add payload verification - verify if collectorName is set properly
			reqNum := atomic.AddInt32(&reqCount, 1)

			switch true {

			// register
			case reqNum <= retriesLimit:
				require.Equal(t, registerUrl, req.URL.Path)

				authHeader := req.Header.Get("Authorization")
				assert.Equal(t, "Bearer dummy_install_token", authHeader,
					"collector didn't send correct Authorization header with registration request")

				if reqCount < retriesLimit {
					w.WriteHeader(http.StatusTooManyRequests)
				} else {

					_, err := w.Write([]byte(`{
						"collectorCredentialId": "aaaaaaaaaaaaaaaaaaaa",
						"collectorCredentialKey": "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
						"collectorId": "000000000FFFFFFF"
					}`))
					if err != nil {
						w.WriteHeader(http.StatusInternalServerError)
					}
				}

			// heartbeat
			case reqNum == retriesLimit+1:
				assert.Equal(t, heartbeatUrl, req.URL.Path)
				w.WriteHeader(204)

			// should not produce any more requests
			default:
				w.WriteHeader(http.StatusInternalServerError)
			}
		})
	}())

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
	cfg.Credentials.InstallToken = "dummy_install_token"
	cfg.CollectorCredentialsDirectory = dir
	cfg.BackOff.InitialInterval = time.Millisecond
	cfg.BackOff.MaxInterval = time.Millisecond

	se, err := newSumologicExtension(cfg, zap.NewNop())
	require.NoError(t, err)
	require.NoError(t, se.Start(context.Background(), componenttest.NewNopHost()))
	regexPattern := fmt.Sprintf("%s-%s", hostname, uuidRegex)
	matched, err := regexp.MatchString(regexPattern, se.collectorName)
	require.NoError(t, err)
	assert.True(t, matched)
}

func TestRegisterEmptyCollectorNameUnrecoverableError(t *testing.T) {
	t.Parallel()

	hostname, err := os.Hostname()
	require.NoError(t, err)
	srv := httptest.NewServer(func() http.HandlerFunc {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			// TODO Add payload verification - verify if collectorName is set properly
			require.Equal(t, registerUrl, req.URL.Path)

			authHeader := req.Header.Get("Authorization")
			assert.Equal(t, "Bearer dummy_install_token", authHeader,
				"collector didn't send correct Authorization header with registration request")

			w.WriteHeader(http.StatusNotFound)
			_, err := w.Write([]byte(`{
				"id": "XXXXX-XXXXX-XXXXX",
				"errors": [
					{
						"code": "collector-registration:dummy_error",
						"message": "The collector cannot be registered"
					}
				]
			}`))
			require.NoError(t, err)
		})
	}())

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
	cfg.Credentials.InstallToken = "dummy_install_token"
	cfg.CollectorCredentialsDirectory = dir
	cfg.BackOff.InitialInterval = time.Millisecond
	cfg.BackOff.MaxInterval = time.Millisecond

	se, err := newSumologicExtension(cfg, zap.NewNop())
	require.NoError(t, err)
	require.EqualError(t, se.Start(context.Background(), componenttest.NewNopHost()),
		"collector registration failed: failed to register the collector, got HTTP status code: 404")
	regexPattern := fmt.Sprintf("%s-%s", hostname, uuidRegex)
	matched, err := regexp.MatchString(regexPattern, se.collectorName)
	require.NoError(t, err)
	assert.True(t, matched)
}

func TestRegistrationRedirect(t *testing.T) {
	t.Parallel()

	var destReqCount int32
	destSrv := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, req *http.Request) {
			switch atomic.AddInt32(&destReqCount, 1) {

			// register
			case 1:
				require.Equal(t, registerUrl, req.URL.Path)

				authHeader := req.Header.Get("Authorization")

				assert.Equal(t, "Bearer dummy_install_token", authHeader,
					"collector didn't send correct Authorization header with registration request")

				_, err := w.Write([]byte(`{
					"collectorCredentialId": "aaaaaaaaaaaaaaaaaaaa",
					"collectorCredentialKey": "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
					"collectorId": "000000000FFFFFFF"
				}`))
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
				}

			// heartbeat, and 2 heartbeats after restart
			case 2, 3, 4:
				assert.Equal(t, heartbeatUrl, req.URL.Path)
				w.WriteHeader(204)

			// should not produce any more requests
			default:
				require.Fail(t,
					"extension should not make more than 2 requests to the destination server",
				)
			}
		},
	))
	t.Cleanup(func() { destSrv.Close() })

	var origReqCount int32
	origSrv := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, req *http.Request) {
			switch atomic.AddInt32(&origReqCount, 1) {

			// register
			case 1:
				require.Equal(t, registerUrl, req.URL.Path)
				http.Redirect(w, req, destSrv.URL, 301)

			// should not produce any more requests
			default:
				require.Fail(t,
					"extension should not make more than 1 request to the original server",
				)
			}
		},
	))
	t.Cleanup(func() { origSrv.Close() })

	dir, err := os.MkdirTemp("", "otelcol-sumo-redirect-test-*")
	require.NoError(t, err)
	t.Cleanup(func() { os.RemoveAll(dir) })

	configFn := func() *Config {
		cfg := createDefaultConfig().(*Config)
		cfg.CollectorName = ""
		cfg.ExtensionSettings = config.ExtensionSettings{}
		cfg.ApiBaseUrl = origSrv.URL
		cfg.Credentials.InstallToken = "dummy_install_token"
		cfg.CollectorCredentialsDirectory = dir
		return cfg
	}

	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	t.Run("works correctly", func(t *testing.T) {
		se, err := newSumologicExtension(configFn(), logger)
		require.NoError(t, err)
		require.NoError(t, se.Start(context.Background(), componenttest.NewNopHost()))
		assert.Eventually(t, func() bool { return atomic.LoadInt32(&origReqCount) == 1 },
			5*time.Second, 100*time.Millisecond,
			"extension should only make 1 request to the original server before redirect",
		)
		assert.Eventually(t, func() bool { return atomic.LoadInt32(&destReqCount) == 2 },
			5*time.Second, 100*time.Millisecond,
			"extension should make 2 requests (registration + heartbeat) to the destination server",
		)
		require.NoError(t, se.Shutdown(context.Background()))
	})

	t.Run("credentials store retrieves credentials with redirected api url", func(t *testing.T) {
		se, err := newSumologicExtension(configFn(), logger)
		require.NoError(t, err)
		require.NoError(t, se.Start(context.Background(), componenttest.NewNopHost()))

		assert.Eventually(t, func() bool { return atomic.LoadInt32(&origReqCount) == 1 },
			5*time.Second, 100*time.Millisecond,
			"after restarting with locally stored credentials extension shouldn't call the original server",
		)

		assert.Eventually(t, func() bool { return atomic.LoadInt32(&destReqCount) == 4 },
			5*time.Second, 100*time.Millisecond,
			"extension should make 4 requests (registration + heartbeat, after restart "+
				"heartbeat to validate credentials and then the first heartbeat on "+
				"which we wait here) to the destination server",
		)

		require.NoError(t, se.Shutdown(context.Background()))
	})
}

func TestCollectorReregistersAfterHTTPUnathorizedFromHeartbeat(t *testing.T) {
	t.Parallel()

	var reqCount int32
	srv := httptest.NewServer(func() http.HandlerFunc {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			reqNum := atomic.AddInt32(&reqCount, 1)

			t.Logf("request: (#%d) %s", reqNum, req.URL.Path)
			handlerRegister := func() {
				require.Equal(t, registerUrl, req.URL.Path, "request num 1")

				authHeader := req.Header.Get("Authorization")
				assert.Equal(t, "Bearer dummy_install_token", authHeader,
					"collector didn't send correct Authorization header with registration request")

				_, err := w.Write([]byte(`{
					"collectorCredentialId": "aaaaaaaaaaaaaaaaaaaa",
					"collectorCredentialKey": "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
					"collectorId": "000000000FFFFFFF",
					"collectorName": "hostname-test-123456123123"
					}`))
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
				}
			}

			switch reqNum {
			// register
			case 1:
				assert.Equal(t, registerUrl, req.URL.Path)
				handlerRegister()

			// heartbeat
			case 2:
				assert.Equal(t, heartbeatUrl, req.URL.Path)
				w.WriteHeader(204)

			// heartbeat
			case 3:
				assert.Equal(t, heartbeatUrl, req.URL.Path)
				// return unauthorized to mimic collector being removed from API
				w.WriteHeader(http.StatusUnauthorized)

			// register
			case 4:
				assert.Equal(t, registerUrl, req.URL.Path)
				handlerRegister()

			default:
				assert.Equal(t, heartbeatUrl, req.URL.Path)
				w.WriteHeader(204)
			}
		})
	}())

	t.Cleanup(func() { srv.Close() })

	dir, err := os.MkdirTemp("", "otelcol-sumo-reregistration-test-*")
	require.NoError(t, err)
	t.Cleanup(func() { os.RemoveAll(dir) })

	cfg := createDefaultConfig().(*Config)
	cfg.CollectorName = ""
	cfg.ApiBaseUrl = srv.URL
	cfg.Credentials.InstallToken = "dummy_install_token"
	cfg.CollectorCredentialsDirectory = dir
	cfg.HeartBeatInterval = 100 * time.Millisecond

	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	se, err := newSumologicExtension(cfg, logger)
	require.NoError(t, err)
	require.NoError(t, se.Start(context.Background(), componenttest.NewNopHost()))

	const expectedReqCount = 10
	if !assert.Eventually(t,
		func() bool {
			return atomic.LoadInt32(&reqCount) == expectedReqCount
		},
		5*time.Second, 50*time.Millisecond,
	) {
		t.Logf("the expected number of requests (%d) wasn't reached, got %d",
			expectedReqCount, atomic.LoadInt32(&reqCount),
		)
	}

	require.NoError(t, se.Shutdown(context.Background()))
}

func TestRegistrationRequestPayload(t *testing.T) {
	t.Parallel()

	hostname, err := os.Hostname()
	require.NoError(t, err)
	srv := httptest.NewServer(func() http.HandlerFunc {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			require.Equal(t, registerUrl, req.URL.Path)

			var reqPayload api.OpenRegisterRequestPayload
			require.NoError(t, json.NewDecoder(req.Body).Decode(&reqPayload))
			require.True(t, reqPayload.Clobber)
			require.Equal(t, hostname, reqPayload.Hostname)
			require.Equal(t, "my description", reqPayload.Description)
			require.Equal(t, "my category/", reqPayload.Category)
			require.EqualValues(t,
				map[string]interface{}{
					"field1": "value1",
					"field2": "value2",
				},
				reqPayload.Fields,
			)
			require.Equal(t, "PST", reqPayload.TimeZone)

			authHeader := req.Header.Get("Authorization")
			assert.Equal(t, "Bearer dummy_install_token", authHeader,
				"collector didn't send correct Authorization header with registration request")

			_, err = w.Write([]byte(`{
				"collectorCredentialId": "mycredentialID",
				"collectorCredentialKey": "mycredentialKey",
				"collectorId": "0000000001231231",
				"collectorName": "otc-test-123456123123"
			}`))
			require.NoError(t, err)
		})
	}())

	dir, err := os.MkdirTemp("", "otelcol-sumo-registration-payload-test-*")
	t.Cleanup(func() {
		srv.Close()
		os.RemoveAll(dir)
	})
	require.NoError(t, err)

	cfg := createDefaultConfig().(*Config)
	cfg.CollectorName = ""
	cfg.ExtensionSettings = config.ExtensionSettings{}
	cfg.ApiBaseUrl = srv.URL
	cfg.Credentials.InstallToken = "dummy_install_token"
	cfg.CollectorCredentialsDirectory = dir
	cfg.BackOff.InitialInterval = time.Millisecond
	cfg.BackOff.MaxInterval = time.Millisecond
	cfg.Clobber = true
	cfg.CollectorDescription = "my description"
	cfg.CollectorCategory = "my category/"
	cfg.CollectorFields = map[string]interface{}{
		"field1": "value1",
		"field2": "value2",
	}
	cfg.TimeZone = "PST"

	se, err := newSumologicExtension(cfg, zap.NewNop())
	require.NoError(t, err)
	require.NoError(t, se.Start(context.Background(), componenttest.NewNopHost()))
	regexPattern := fmt.Sprintf("%s-%s", hostname, uuidRegex)
	matched, err := regexp.MatchString(regexPattern, se.collectorName)
	require.NoError(t, err)
	assert.True(t, matched)

	require.NoError(t, se.Shutdown(context.Background()))
}
