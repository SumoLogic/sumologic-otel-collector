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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/open-telemetry/opentelemetry-collector-contrib/extension/sumologicextension/api"
	"go.uber.org/zap"
)

// CredsGetter is an interface to get collector authentication data
type CredsGetter interface {
	CheckCollectorCredentials() bool
	GetStoredCredentials() (api.OpenRegisterResponsePayload, error)
	RegisterCollector(ctx context.Context) (api.OpenRegisterResponsePayload, error)
}

// credsGetter is a common structure for both types of getting collector credentials
// stored in file and registration new collector.
type credsGetter struct {
	conf   *Config
	logger *zap.Logger
}

// CheckCollectorCredentials checks if collector credentials can be found in path
// configured in the config.
func (cr credsGetter) CheckCollectorCredentials() bool {
	filenameHash, err := hash(cr.conf.CollectorName)
	if err != nil {
		return false
	}
	path := path.Join(cr.conf.CollectorCredentialsPath, filenameHash)
	if _, err := os.Stat(path); err != nil {
		return false
	}
	return true
}

// GetStoredCredentials retrieves collector credentials stored in local file system and then decrypts it
// using hashed collector name as passphrase.
func (cr credsGetter) GetStoredCredentials() (api.OpenRegisterResponsePayload, error) {
	filenameHash, err := hash(cr.conf.CollectorName)
	if err != nil {
		return api.OpenRegisterResponsePayload{}, err
	}

	path := path.Join(cr.conf.CollectorCredentialsPath, filenameHash)
	creds, err := os.Open(path)
	if err != nil {
		return api.OpenRegisterResponsePayload{}, err
	}

	defer creds.Close()
	encryptedCreds, err := ioutil.ReadAll(creds)
	if err != nil {
		return api.OpenRegisterResponsePayload{}, err
	}

	collectorCreds, err := decrypt(encryptedCreds, cr.conf.CollectorName)
	if err != nil {
		return api.OpenRegisterResponsePayload{}, err
	}

	var credentialsInfo api.OpenRegisterResponsePayload
	if err = json.Unmarshal(collectorCreds, &credentialsInfo); err != nil {
		return api.OpenRegisterResponsePayload{}, err
	}

	return credentialsInfo, nil
}

// RegisterCollector registers new collector using registration API and returns collector credentials.
func (cr credsGetter) RegisterCollector(ctx context.Context) (api.OpenRegisterResponsePayload, error) {
	baseUrl := strings.TrimSuffix(cr.conf.ApiBaseUrl, "/")
	u, err := url.Parse(baseUrl)
	if err != nil {
		return api.OpenRegisterResponsePayload{}, err
	}
	u.Path = registerUrl

	// TODO: just plain hostname or we want to add some custom logic when setting
	// hostname in request?
	hostname, err := os.Hostname()
	if err != nil {
		return api.OpenRegisterResponsePayload{}, fmt.Errorf("cannot get hostname: %w", err)
	}

	var buff bytes.Buffer
	if err = json.NewEncoder(&buff).Encode(api.OpenRegisterRequestPayload{
		CollectorName: se.conf.CollectorName,
		Description:   se.conf.CollectorDescription,
		Category:      se.conf.CollectorCategory,
		Hostname:      hostname,
		Ephemeral:     se.conf.Ephemeral,
		Clobber:       se.conf.Clobber,
		TimeZone:      se.conf.TimeZone,
	}); err != nil {
		return api.OpenRegisterResponsePayload{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), &buff)
	if err != nil {
		return api.OpenRegisterResponsePayload{}, err
	}

	addClientCredentials(req,
		cr.conf.Credentials.AccessID,
		cr.conf.Credentials.AccessKey,
	)
	addJSONHeaders(req)

	cr.logger.Info("Calling register API", zap.String("URL", u.String()))
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return api.OpenRegisterResponsePayload{}, fmt.Errorf("failed to register the collector: %w", err)
	}

	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 400 {
		var buff bytes.Buffer
		if _, err := io.Copy(&buff, res.Body); err != nil {
			return api.OpenRegisterResponsePayload{}, fmt.Errorf(
				"failed to copy collector registration response body, status code: %d, err: %w",
				res.StatusCode, err,
			)
		}
		cr.logger.Debug("Collector registration failed",
			zap.Int("status_code", res.StatusCode),
			zap.String("response", buff.String()),
		)
		return api.OpenRegisterResponsePayload{}, fmt.Errorf(
			"failed to register the collector, got HTTP status code: %d",
			res.StatusCode,
		)
	}

	var resp api.OpenRegisterResponsePayload
	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
		return api.OpenRegisterResponsePayload{}, err
	}

	cr.logger.Info("Collector registered",
		zap.String("CollectorID", resp.CollectorId),
		zap.Any("response", resp),
	)

	return resp, nil
}

func addClientCredentials(req *http.Request, accessID string, accessKey string) {
	// TODO: What is preferred: headers or basic auth?
	req.Header.Add("accessid", accessID)
	req.Header.Add("accesskey", accessKey)
}

func addJSONHeaders(req *http.Request) {
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
}
