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
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/SumoLogic/sumologic-otel-collector/pkg/extension/sumologicextension/api"
	"github.com/cenkalti/backoff/v4"
	"github.com/google/uuid"
	"go.opentelemetry.io/collector/component"
	"go.uber.org/zap"
)

type SumologicExtension struct {
	collectorName    string
	baseUrl          string
	httpClient       *http.Client
	conf             *Config
	logger           *zap.Logger
	credentialsStore CredentialsStore
	hashKey          string
	registrationInfo api.OpenRegisterResponsePayload
	closeChan        chan struct{}
	closeOnce        sync.Once
	backOff          *backoff.ExponentialBackOff
}

const (
	heartbeatUrl                  = "/api/v1/collector/heartbeat"
	registerUrl                   = "/api/v1/collector/register"
	collectorCredentialsDirectory = ".sumologic-otel-collector/"

	collectorIdField           = "collector_id"
	collectorNameField         = "collector_name"
	collectorCredentialIdField = "collector_credential_id"

	banner = `
************************************************************************************************************
***    This software is currently in beta and is not recommended for production environments.            ***
***    To participate in this beta, please contact your Sumo Logic account team or Sumo Logic Support.   ***
************************************************************************************************************
`
)

const (
	DefaultHeartbeatInterval = 15 * time.Second
)

func newSumologicExtension(conf *Config, logger *zap.Logger) (*SumologicExtension, error) {
	if conf.Credentials.AccessID == "" || conf.Credentials.AccessKey == "" {
		return nil, errors.New("access_key and/or access_id not provided")
	}
	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	var collectorName string
	credentialsStore := localFsCredentialsStore{
		collectorCredentialsDirectory: conf.CollectorCredentialsDirectory,
		logger:                        logger,
	}
	if conf.CollectorName == "" {
		key := createHashKey(conf)
		// If collector name is not set by the user check if the collector was restarted
		if !credentialsStore.Check(key) {
			// If credentials file is not stored on filesystem generate collector name
			collectorName = fmt.Sprintf("%s-%s", hostname, uuid.New())
		}
	} else {
		collectorName = conf.CollectorName
	}
	if conf.HeartBeatInterval <= 0 {
		conf.HeartBeatInterval = DefaultHeartbeatInterval
	}

	// Prepare ExponentialBackoff
	backOff := backoff.NewExponentialBackOff()
	backOff.InitialInterval = conf.BackOff.InitialInterval
	backOff.MaxElapsedTime = conf.BackOff.MaxElapsedTime
	backOff.MaxInterval = conf.BackOff.MaxInterval

	return &SumologicExtension{
		collectorName:    collectorName,
		baseUrl:          strings.TrimSuffix(conf.ApiBaseUrl, "/"),
		conf:             conf,
		logger:           logger,
		hashKey:          createHashKey(conf),
		credentialsStore: credentialsStore,
		closeChan:        make(chan struct{}),
		backOff:          backOff,
	}, nil
}

func createHashKey(conf *Config) string {
	return fmt.Sprintf("%s%s%s", conf.CollectorName, conf.Credentials.AccessID, conf.Credentials.AccessKey)
}

func (se *SumologicExtension) validateCredenials(
	ctx context.Context,
	creds api.OpenRegisterResponsePayload,
) error {
	return se.sendHeartbeat(ctx)
}

func (se *SumologicExtension) Start(ctx context.Context, host component.Host) error {
	se.logger.Info(banner)
	colCreds, registrationDone, err := se.getCredentials(ctx)
	if err != nil {
		return err
	}

	// Add logger fields based on actual collector name and ID as returned
	// by registration API.
	se.logger = se.logger.With(
		zap.String(collectorNameField, colCreds.Credentials.CollectorName),
		zap.String(collectorIdField, colCreds.Credentials.CollectorId),
	)

	se.registrationInfo = colCreds.Credentials

	se.httpClient, err = se.conf.HTTPClientSettings.ToClient(host.GetExtensions())
	if err != nil {
		return fmt.Errorf("couldn't create HTTP client: %w", err)
	}

	// Set the transport so that all requests from se.httpClient will contain
	// the collector credentials.
	se.httpClient.Transport, err = se.RoundTripper(se.httpClient.Transport)
	if err != nil {
		return fmt.Errorf("couldn't create HTTP client transport: %w", err)
	}

	if !registrationDone {
		se.logger.Info("Checking if locally retrieved credentials are still valid...")
		if err := se.validateCredenials(ctx, colCreds.Credentials); err != nil {
			return fmt.Errorf("locally stored credentials are invalid: %w", err)
		}
		se.logger.Info("Local collector credentials all good, starting up the collector")
	}

	go se.heartbeatLoop()

	return nil
}

// Shutdown is invoked during service shutdown.
func (se *SumologicExtension) Shutdown(ctx context.Context) error {
	se.closeOnce.Do(func() { close(se.closeChan) })
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}

// getCredentials returns the credentials either retrieved from local credentials
// storage in cases they were available there or it registers the collector and
// returns the credentials obtained from the API.
// It also returns a boolean flag indicating if the successful registration
// with the API took place and an error.
func (se *SumologicExtension) getCredentials(ctx context.Context) (CollectorCredentials, bool, error) {
	var (
		colCreds         CollectorCredentials
		registrationDone bool
		err              error
	)

	if se.credentialsStore.Check(se.hashKey) {
		colCreds, err = se.credentialsStore.Get(se.hashKey)
		if err != nil {
			return CollectorCredentials{}, false, err
		}
		se.collectorName = colCreds.CollectorName
		if !se.conf.Clobber {
			if colCreds.ApiBaseUrl != "" {
				se.baseUrl = colCreds.ApiBaseUrl
			}

			se.logger.Info("Found stored credentials, skipping registration",
				zap.String(collectorNameField, colCreds.Credentials.CollectorName),
			)
		} else {
			se.logger.Info(
				"Locally stored credentials found, but clobber flag is set: " +
					"re-registering the collector",
			)
			if colCreds, err = se.registerCollectorWithBackoff(ctx, se.collectorName); err != nil {
				return CollectorCredentials{}, false, err
			}
			registrationDone = true
			if err := se.credentialsStore.Store(se.hashKey, colCreds); err != nil {
				se.logger.Error("Unable to store collector credentials", zap.Error(err))
			}
		}
	} else {
		se.logger.Info("Locally stored credentials not found, registering the collector")
		if colCreds, err = se.registerCollectorWithBackoff(ctx, se.collectorName); err != nil {
			return CollectorCredentials{}, false, err
		}
		registrationDone = true
		if err := se.credentialsStore.Store(se.hashKey, colCreds); err != nil {
			se.logger.Error("Unable to store collector credentials", zap.Error(err))
		}
	}

	return colCreds, registrationDone, err
}

// registerCollector registers the collector using registration API and returns
// the obtained collector credentials.
func (se *SumologicExtension) registerCollector(ctx context.Context, collectorName string) (CollectorCredentials, error) {
	u, err := url.Parse(se.baseUrl)
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
		Description:   se.conf.CollectorDescription,
		Category:      se.conf.CollectorCategory,
		Fields:        se.conf.CollectorFields,
		Hostname:      hostname,
		Ephemeral:     se.conf.Ephemeral,
		Clobber:       se.conf.Clobber,
		TimeZone:      se.conf.TimeZone,
	}); err != nil {
		return CollectorCredentials{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), &buff)
	if err != nil {
		return CollectorCredentials{}, err
	}

	addClientCredentials(req,
		credentials{
			AccessID:  se.conf.Credentials.AccessID,
			AccessKey: se.conf.Credentials.AccessKey,
		},
	)
	addJSONHeaders(req)

	se.logger.Info("Calling register API", zap.String("URL", u.String()))

	client := *http.DefaultClient
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	res, err := client.Do(req)
	if err != nil {
		se.logger.Warn("Collector registration HTTP request failed", zap.Error(err))
		return CollectorCredentials{}, fmt.Errorf("failed to register the collector: %w", err)
	}

	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 400 {
		return se.handleRegistrationError(res)
	} else if res.StatusCode == 301 {
		// Use the URL from Location header for subsequent requests.
		se.baseUrl = strings.TrimSuffix(res.Header.Get("Location"), "/")
		se.logger.Info("Redirected to a different deployment",
			zap.String("url", se.baseUrl),
		)
		return se.registerCollector(ctx, collectorName)
	}

	var resp api.OpenRegisterResponsePayload
	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
		return CollectorCredentials{}, err
	}

	se.logger.Info("Collector registered",
		zap.String(collectorIdField, resp.CollectorId),
		zap.String(collectorNameField, resp.CollectorName),
		zap.String(collectorCredentialIdField, resp.CollectorCredentialId),
	)
	return CollectorCredentials{
		CollectorName: collectorName,
		Credentials:   resp,
		ApiBaseUrl:    se.baseUrl,
	}, nil
}

// handleRegistrationError handles the collector registration errors and returns
// appropriate error for backoff handling and logging purposes.
func (se *SumologicExtension) handleRegistrationError(res *http.Response) (CollectorCredentials, error) {
	var errResponse api.ErrorResponsePayload
	if err := json.NewDecoder(res.Body).Decode(&errResponse); err != nil {
		var buff bytes.Buffer
		if _, errCopy := io.Copy(&buff, res.Body); errCopy != nil {
			return CollectorCredentials{}, fmt.Errorf(
				"failed to read the collector registration response body, status code: %d, err: %w",
				res.StatusCode, errCopy,
			)
		}
		return CollectorCredentials{}, fmt.Errorf(
			"failed to decode collector registration response body: %s, status code: %d, err: %w",
			buff.String(), res.StatusCode, err,
		)
	}

	se.logger.Warn("Collector registration failed",
		zap.Int("status_code", res.StatusCode),
		zap.String("error_id", errResponse.ID),
		zap.Any("errors", errResponse.Errors),
	)

	// Return unrecoverable error for 4xx status codes except 429
	if res.StatusCode >= 400 && res.StatusCode < 500 && res.StatusCode != 429 {
		return CollectorCredentials{}, backoff.Permanent(fmt.Errorf(
			"failed to register the collector, got HTTP status code: %d",
			res.StatusCode,
		))
	}

	return CollectorCredentials{}, fmt.Errorf(
		"failed to register the collector, got HTTP status code: %d", res.StatusCode,
	)
}

// callRegisterWithBackoff calls registration using exponential backoff algorithm
// this loosely base on backoff.Retry function
func (se *SumologicExtension) registerCollectorWithBackoff(ctx context.Context, collectorName string) (CollectorCredentials, error) {
	se.backOff.Reset()
	for {
		creds, err := se.registerCollector(ctx, collectorName)
		if err == nil {
			return creds, nil
		}

		nbo := se.backOff.NextBackOff()
		// Return error if backoff reaches the limit or uncoverable error is spotted
		if _, ok := err.(*backoff.PermanentError); nbo == se.backOff.Stop || ok {
			return CollectorCredentials{}, fmt.Errorf("collector registration failed: %w", err)
		}

		t := time.NewTimer(nbo)
		defer t.Stop()

		select {
		case <-t.C:
		case <-ctx.Done():
			return CollectorCredentials{}, fmt.Errorf("collector registration cancelled: %w", ctx.Err())
		}
	}
}

func (se *SumologicExtension) heartbeatLoop() {
	if se.registrationInfo.CollectorCredentialId == "" || se.registrationInfo.CollectorCredentialKey == "" {
		se.logger.Error("Collector not registered, cannot send heartbeat")
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		// When the close channel is closed ...
		<-se.closeChan
		// ... cancel the ongoing heartbeat request.
		cancel()
	}()

	se.logger.Info("Heartbeat API initialized. Starting sending hearbeat requests")
	timer := time.NewTimer(se.conf.HeartBeatInterval)
	for {
		select {
		case <-se.closeChan:
			se.logger.Info("Heartbeat sender turned off")
			return
		default:
			if err := se.sendHeartbeat(ctx); err != nil {
				se.logger.Error("Heartbeat error", zap.Error(err))
			} else {
				se.logger.Debug("Heartbeat sent")
			}

			select {
			case <-timer.C:
				timer.Stop()
				timer.Reset(se.conf.HeartBeatInterval)
			case <-se.closeChan:
			}

		}
	}
}

func (se *SumologicExtension) sendHeartbeat(ctx context.Context) error {
	u, err := url.Parse(se.baseUrl + heartbeatUrl)
	if err != nil {
		return fmt.Errorf("unable to parse heartbeat URL %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), nil)
	if err != nil {
		return fmt.Errorf("unable to create HTTP request %w", err)
	}

	addJSONHeaders(req)
	res, err := se.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("unable to send HTTP request: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode != 204 {
		var buff bytes.Buffer
		if _, err := io.Copy(&buff, res.Body); err != nil {
			return fmt.Errorf(
				"failed to copy collector heartbeat response body, status code: %d, err: %w",
				res.StatusCode, err,
			)
		}
		return fmt.Errorf(
			"collector heartbeat request failed, status code: %d, body: %s",
			res.StatusCode, buff.String(),
		)
	}

	return nil
}

func (se *SumologicExtension) ComponentID() string {
	return se.conf.ExtensionSettings.ID().String()
}

func (se *SumologicExtension) CollectorID() string {
	return se.registrationInfo.CollectorId
}

func (se *SumologicExtension) BaseUrl() string {
	return se.baseUrl
}

// Implement [1] in order for this extension to be used as custom exporter
// authenticator.
//
// [1]: https://github.com/open-telemetry/opentelemetry-collector/blob/2e84285efc665798d76773b9901727e8836e9d8f/config/configauth/clientauth.go#L34-L39
func (se *SumologicExtension) RoundTripper(base http.RoundTripper) (http.RoundTripper, error) {
	return roundTripper{
		collectorCredentialId:  se.registrationInfo.CollectorCredentialId,
		collectorCredentialKey: se.registrationInfo.CollectorCredentialKey,
		base:                   base,
	}, nil
}

type roundTripper struct {
	collectorCredentialId  string
	collectorCredentialKey string
	base                   http.RoundTripper
}

func (rt roundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	addCollectorCredentials(req, rt.collectorCredentialId, rt.collectorCredentialKey)
	return rt.base.RoundTrip(req)
}

func addCollectorCredentials(req *http.Request, collectorCredentialId string, collectorCredentialKey string) {
	token := base64.StdEncoding.EncodeToString(
		[]byte(collectorCredentialId + ":" + collectorCredentialKey),
	)

	req.Header.Add("Authorization", "Basic "+token)
}

func addClientCredentials(req *http.Request, credentials credentials) {
	token := base64.StdEncoding.EncodeToString(
		[]byte(credentials.AccessID + ":" + credentials.AccessKey),
	)

	req.Header.Add("Authorization", "Basic "+token)
}
