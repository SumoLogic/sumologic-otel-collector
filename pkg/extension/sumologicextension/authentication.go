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

// Authentication is an interface to get collector authentication data
type Authentication interface {
	Get() (api.OpenRegisterResponsePayload, error)
}

// Authenticator is a common structure for both types of getting collector credentials
// stored in file and registration new collector.
type Authenticator struct {
	conf   *Config
	logger *zap.Logger
}

// Register is registering new collector using API
type Register Authenticator

// Get registers new collector using registration API and returns collector credentials
func (rc Register) Get() (api.OpenRegisterResponsePayload, error) {
	baseUrl := strings.TrimSuffix(rc.conf.ApiBaseUrl, "/")
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
		Ephemeral:     true, // TODO: change that
		CollectorName: rc.conf.CollectorName,
		Description:   rc.conf.CollectorDescription,
		Category:      rc.conf.CollectorCategory,
		Hostname:      hostname,
	}); err != nil {
		return api.OpenRegisterResponsePayload{}, err
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, u.String(), &buff)
	if err != nil {
		return api.OpenRegisterResponsePayload{}, err
	}

	addClientCredentials(req,
		rc.conf.Credentials.AccessID,
		rc.conf.Credentials.AccessKey,
	)
	addJSONHeaders(req)

	rc.logger.Info("Calling register API", zap.String("URL", u.String()))
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
		rc.logger.Debug("Collector registration failed",
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

	rc.logger.Info("Collector registered",
		zap.String("CollectorID", resp.CollectorId),
		zap.Any("response", resp),
	)

	return resp, nil
}

func addClientCredentials(req *http.Request, accessID string, accessKey string) {
	// TODO: What is preferred: headers of basic auth?
	req.Header.Add("accessid", accessID)
	req.Header.Add("accesskey", accessKey)
}

func addJSONHeaders(req *http.Request) {
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
}

// Stored is extracting stored authentication information for
// previously registered collector
type Stored Authenticator

// Get retrieves, decrypts collector credentials using
// hashed collector name as passphrase and then assign it to registrationInfo
// field.
func (sc Stored) Get() (api.OpenRegisterResponsePayload, error) {
	filenameHash, err := hash(sc.conf.CollectorName)
	if err != nil {
		return api.OpenRegisterResponsePayload{}, err
	}

	path := path.Join(sc.conf.CollectorCredentialsPath, filenameHash)
	creds, err := os.Open(path)
	if err != nil {
		return api.OpenRegisterResponsePayload{}, err
	}

	defer creds.Close()
	encryptedCreds, err := ioutil.ReadAll(creds)
	if err != nil {
		return api.OpenRegisterResponsePayload{}, err
	}

	collectorCreds, err := decrypt(encryptedCreds, sc.conf.CollectorName)
	if err != nil {
		return api.OpenRegisterResponsePayload{}, err
	}

	var credentialsInfo api.OpenRegisterResponsePayload
	if err = json.Unmarshal(collectorCreds, &credentialsInfo); err != nil {
		return api.OpenRegisterResponsePayload{}, err
	}

	return credentialsInfo, nil
}
