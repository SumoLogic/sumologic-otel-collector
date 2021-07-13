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
	"encoding/base64"
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

// CollectorCredentials are used for storing the credentials received on registration
type CollectorCredentials struct {
	CollectorName string                          `json:"collectorName"`
	Credentials   api.OpenRegisterResponsePayload `json:"collectorCredentials"`
}

// CredsGetter is an interface to get collector authentication data
type CredsGetter interface {
	CheckCollectorCredentials(key string) bool
	GetStoredCredentials(key string) (CollectorCredentials, error)
	RegisterCollector(ctx context.Context, collectorName string) (CollectorCredentials, error)
}

// credsGetter is a common structure for both types of getting collector credentials
// stored in file and registration new collector.
type credsGetter struct {
	conf   *Config
	logger *zap.Logger
}

// CheckCollectorCredentials checks if collector credentials can be found in path
// configured in the config.
func (cr credsGetter) CheckCollectorCredentials(key string) bool {
	filenameHash, err := hash(key)
	if err != nil {
		return false
	}
	path := path.Join(cr.conf.CollectorCredentialsDirectory, filenameHash)
	if _, err := os.Stat(path); err != nil {
		return false
	}
	return true
}

// GetStoredCredentials retrieves collector credentials stored in local file system and then decrypts it
// using hashed collector name as passphrase.
func (cr credsGetter) GetStoredCredentials(key string) (CollectorCredentials, error) {
	filenameHash, err := hash(key)
	if err != nil {
		return CollectorCredentials{}, err
	}

	path := path.Join(cr.conf.CollectorCredentialsDirectory, filenameHash)
	creds, err := os.Open(path)
	if err != nil {
		return CollectorCredentials{}, err
	}

	defer creds.Close()
	encryptedCreds, err := ioutil.ReadAll(creds)
	if err != nil {
		return CollectorCredentials{}, err
	}

	collectorCreds, err := decrypt(encryptedCreds, key)
	if err != nil {
		return CollectorCredentials{}, err
	}

	var credentialsInfo CollectorCredentials
	if err = json.Unmarshal(collectorCreds, &credentialsInfo); err != nil {
		return CollectorCredentials{}, err
	}

	return credentialsInfo, nil
}

// RegisterCollector registers new collector using registration API and returns collector credentials.
func (cr credsGetter) RegisterCollector(ctx context.Context, collectorName string) (CollectorCredentials, error) {
	baseUrl := strings.TrimSuffix(cr.conf.ApiBaseUrl, "/")
	u, err := url.Parse(baseUrl)
	if err != nil {
		return CollectorCredentials{}, err
	}
	u.Path = registerUrl

	// TODO: just plain hostname or we want to add some custom logic when setting
	// hostname in request?
	hostname, err := os.Hostname()
	if err != nil {
		return CollectorCredentials{}, fmt.Errorf("cannot get hostname: %w", err)
	}

	var buff bytes.Buffer
	if err = json.NewEncoder(&buff).Encode(api.OpenRegisterRequestPayload{
		CollectorName: collectorName,
		Description:   cr.conf.CollectorDescription,
		Category:      cr.conf.CollectorCategory,
		Fields:        cr.conf.CollectorFields,
		Hostname:      hostname,
		Ephemeral:     cr.conf.Ephemeral,
		Clobber:       cr.conf.Clobber,
		TimeZone:      cr.conf.TimeZone,
	}); err != nil {
		return CollectorCredentials{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), &buff)
	if err != nil {
		return CollectorCredentials{}, err
	}

	addClientCredentials(req,
		credentials{
			AccessID:  cr.conf.Credentials.AccessID,
			AccessKey: cr.conf.Credentials.AccessKey,
		},
	)
	addJSONHeaders(req)

	cr.logger.Info("Calling register API", zap.String("URL", u.String()))
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return CollectorCredentials{}, fmt.Errorf("failed to register the collector: %w", err)
	}

	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 400 {
		var buff bytes.Buffer
		if _, err := io.Copy(&buff, res.Body); err != nil {
			return CollectorCredentials{}, fmt.Errorf(
				"failed to copy collector registration response body, status code: %d, err: %w",
				res.StatusCode, err,
			)
		}
		cr.logger.Debug("Collector registration failed",
			zap.Int("status_code", res.StatusCode),
			zap.String("response", buff.String()),
		)
		return CollectorCredentials{}, fmt.Errorf(
			"failed to register the collector, got HTTP status code: %d",
			res.StatusCode,
		)
	}

	var resp api.OpenRegisterResponsePayload
	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
		return CollectorCredentials{}, err
	}

	cr.logger.Info("Collector registered",
		zap.String("CollectorID", resp.CollectorId),
		zap.Any("response", resp),
	)

	return CollectorCredentials{
		// TODO: When registration API will return registered name use it instead of collectorName
		CollectorName: collectorName,
		Credentials:   resp,
	}, nil
}

func addClientCredentials(req *http.Request, credentials credentials) {
	token := base64.StdEncoding.EncodeToString(
		[]byte(credentials.AccessID + ":" + credentials.AccessKey),
	)

	req.Header.Add("Authorization", "Basic "+token)
}

func addJSONHeaders(req *http.Request) {
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
}
