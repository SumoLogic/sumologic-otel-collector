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
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/Showmax/go-fqdn"
	"github.com/cenkalti/backoff/v4"
	ps "github.com/mitchellh/go-ps"
	"github.com/shirou/gopsutil/v3/host"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/confighttp"
	"go.opentelemetry.io/collector/extension/auth"
	"go.opentelemetry.io/collector/featuregate"
	"go.uber.org/zap"
	grpccredentials "google.golang.org/grpc/credentials"

	"github.com/SumoLogic/sumologic-otel-collector/pkg/extension/sumologicextension/api"
	"github.com/SumoLogic/sumologic-otel-collector/pkg/extension/sumologicextension/credentials"
)

type SumologicExtension struct {
	collectorName string
	buildVersion  string

	// The lock around baseUrl is needed because sumologicexporter is using
	// it as base URL for API requests and this access has to be coordinated.
	baseUrlLock sync.RWMutex
	baseUrl     string

	credsNotifyLock   sync.Mutex
	credsNotifyUpdate chan struct{}

	host             component.Host
	conf             *Config
	origLogger       *zap.Logger
	logger           *zap.Logger
	credentialsStore credentials.Store
	hashKey          string
	httpClient       *http.Client
	registrationInfo api.OpenRegisterResponsePayload
	updateMetadata   bool

	closeChan chan struct{}
	closeOnce sync.Once
	backOff   *backoff.ExponentialBackOff
	id        component.ID
}

const (
	heartbeatUrl = "/api/v1/collector/heartbeat"
	metadataUrl  = "/api/v1/otCollectors/metadata"
	registerUrl  = "/api/v1/collector/register"

	collectorIdField           = "collector_id"
	collectorNameField         = "collector_name"
	collectorCredentialIdField = "collector_credential_id"
)

const (
	updateCollectorMetadataID    = "extension.sumologic.updateCollectorMetadata"
	updateCollectorMetadataStage = featuregate.StageAlpha

	DefaultHeartbeatInterval = 15 * time.Second
)

var updateCollectorMetadataFeatureGate *featuregate.Gate

func init() {
	updateCollectorMetadataFeatureGate = featuregate.GlobalRegistry().MustRegister(
		updateCollectorMetadataID,
		updateCollectorMetadataStage,
		featuregate.WithRegisterDescription("When enabled, the collector will update its Sumo Logic metadata on startup."),
		featuregate.WithRegisterReferenceURL("https://github.com/SumoLogic/sumologic-otel-collector/pull/858"),
	)
}

var errGRPCNotSupported = fmt.Errorf("gRPC is not supported by sumologicextension")

// SumologicExtension implements ClientAuthenticator
var _ auth.Client = (*SumologicExtension)(nil)

func newSumologicExtension(conf *Config, logger *zap.Logger, id component.ID, buildVersion string) (*SumologicExtension, error) {
	if conf.Credentials.InstallationToken == "" && conf.Credentials.InstallToken == "" {
		return nil, errors.New("access credentials not provided: need installation_token")
	} else if conf.Credentials.InstallationToken == "" {
		logger.Warn("install_token is deprecated. Please use installation_token instead")
		conf.Credentials.InstallationToken = conf.Credentials.InstallToken
	}

	hostname, err := getHostname(logger)
	if err != nil {
		return nil, err
	}

	credentialsStore, err := credentials.NewLocalFsStore(
		credentials.WithCredentialsDirectory(conf.CollectorCredentialsDirectory),
		credentials.WithLogger(logger),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize credentials store: %w", err)
	}

	var (
		collectorName string
		hashKey       = createHashKey(conf)
	)
	if conf.CollectorName == "" {
		// If collector name is not set by the user, check if the collector was restarted
		// and that we can reuse collector name save in credentials store.
		if creds, err := credentialsStore.Get(hashKey); err != nil {
			// If credentials file is not stored on filesystem generate collector name
			collectorName = hostname
		} else {
			collectorName = creds.CollectorName
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
		collectorName:     collectorName,
		buildVersion:      buildVersion,
		baseUrl:           strings.TrimSuffix(conf.ApiBaseUrl, "/"),
		credsNotifyUpdate: make(chan struct{}),
		conf:              conf,
		origLogger:        logger,
		logger:            logger,
		hashKey:           hashKey,
		credentialsStore:  credentialsStore,
		updateMetadata:    updateCollectorMetadataFeatureGate.IsEnabled(),
		closeChan:         make(chan struct{}),
		backOff:           backOff,
		id:                id,
	}, nil
}

func createHashKey(conf *Config) string {
	return fmt.Sprintf("%s%s%s",
		conf.CollectorName,
		conf.Credentials.InstallationToken,
		strings.TrimSuffix(conf.ApiBaseUrl, "/"),
	)
}

func (se *SumologicExtension) Start(ctx context.Context, host component.Host) error {
	se.host = host

	colCreds, err := se.getCredentials(ctx)
	if err != nil {
		return err
	}

	if err = se.injectCredentials(colCreds); err != nil {
		return err
	}

	// Add logger fields based on actual collector name and ID.
	se.logger = se.origLogger.With(
		zap.String(collectorNameField, colCreds.Credentials.CollectorName),
		zap.String(collectorIdField, colCreds.Credentials.CollectorId),
	)

	if se.updateMetadata {
		err = se.updateMetadataWithBackoff(ctx)
		if err != nil {
			return err
		}
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

func (se *SumologicExtension) validateCredentials(
	ctx context.Context,
	colCreds credentials.CollectorCredentials,
) error {
	se.logger.Info("Validating collector credentials...",
		zap.String(collectorCredentialIdField, colCreds.Credentials.CollectorCredentialId),
		zap.String(collectorIdField, colCreds.Credentials.CollectorId),
	)

	if err := se.injectCredentials(colCreds); err != nil {
		return err
	}

	return se.sendHeartbeatWithHTTPClient(ctx, se.httpClient)
}

// injectCredentials injects the collector credentials:
//   - into registration info that's stored in the extension and can be used by roundTripper
//   - into http client and its transport so that each request is using collector
//     credentials as authentication keys
func (se *SumologicExtension) injectCredentials(colCreds credentials.CollectorCredentials) error {
	se.credsNotifyLock.Lock()
	defer se.credsNotifyLock.Unlock()

	// Set the registration info so that it can be used in RoundTripper.
	se.registrationInfo = colCreds.Credentials

	httpClient, err := se.getHTTPClient(se.conf.HTTPClientSettings, colCreds.Credentials)
	if err != nil {
		return err
	}

	se.httpClient = httpClient

	// Let components know that the credentials may have changed.
	close(se.credsNotifyUpdate)
	se.credsNotifyUpdate = make(chan struct{})

	return nil
}

func (se *SumologicExtension) getHTTPClient(
	httpClientSettings confighttp.HTTPClientSettings,
	regInfo api.OpenRegisterResponsePayload,
) (*http.Client, error) {
	httpClient, err := httpClientSettings.ToClient(
		se.host,
		component.TelemetrySettings{},
	)
	if err != nil {
		return nil, fmt.Errorf("couldn't create HTTP client: %w", err)
	}

	// Set the transport so that all requests from httpClient will contain
	// the collector credentials.
	httpClient.Transport, err = se.RoundTripper(httpClient.Transport)
	if err != nil {
		return nil, fmt.Errorf("couldn't create HTTP client transport: %w", err)
	}

	return httpClient, nil
}

// getCredentials retrieves the credentials for the collector.
// It does so by checking the local credentials store and by validating those credentials.
// In case they are invalid or are not available through local credentials store
// then it tries to register the collector using the provided access keys.
func (se *SumologicExtension) getCredentials(ctx context.Context) (credentials.CollectorCredentials, error) {
	var (
		colCreds credentials.CollectorCredentials
		err      error
	)

	if !se.conf.ForceRegistration {
		colCreds, err = se.getLocalCredentials(ctx)
		if err == nil {
			errV := se.validateCredentials(ctx, colCreds)
			if errV == nil {
				se.logger.Info("Found stored credentials, skipping registration",
					zap.String(collectorNameField, colCreds.Credentials.CollectorName),
				)
				return colCreds, nil
			}

			// Credentials might have ended up being invalid or the collector
			// might have been removed in Sumo.
			// Fall back to removing the credentials and recreating them by registering
			// the collector.
			if err := se.credentialsStore.Delete(se.hashKey); err != nil {
				se.logger.Error(
					"Unable to delete old collector credentials", zap.Error(err),
				)
			}

			se.logger.Info("Locally stored credentials invalid. Trying to re-register...",
				zap.String(collectorNameField, colCreds.Credentials.CollectorName),
				zap.String(collectorIdField, colCreds.Credentials.CollectorId),
				zap.Error(errV),
			)
		} else {
			se.logger.Info("Locally stored credentials not found, registering the collector")
		}
	}

	colCreds, err = se.getCredentialsByRegistering(ctx)
	if err != nil {
		return credentials.CollectorCredentials{}, err
	}

	return colCreds, nil
}

// getCredentialsByRegistering registers the collector and returns the credentials
// obtained from the API.
func (se *SumologicExtension) getCredentialsByRegistering(ctx context.Context) (credentials.CollectorCredentials, error) {
	colCreds, err := se.registerCollectorWithBackoff(ctx, se.collectorName)
	if err != nil {
		return credentials.CollectorCredentials{}, err
	}
	if err := se.credentialsStore.Store(se.hashKey, colCreds); err != nil {
		se.logger.Error(
			"Unable to store collector credentials, they will be used now but won't be re-used on next run",
			zap.Error(err),
		)
	}

	se.collectorName = colCreds.CollectorName

	return colCreds, nil
}

// getLocalCredentials returns the credentials retrieved from local credentials
// storage in case they are available there.
func (se *SumologicExtension) getLocalCredentials(ctx context.Context) (credentials.CollectorCredentials, error) {
	colCreds, err := se.credentialsStore.Get(se.hashKey)
	if err != nil {
		return credentials.CollectorCredentials{},
			fmt.Errorf("problem finding local collector credentials (hash key: %s): %w",
				se.hashKey, err,
			)
	}

	se.collectorName = colCreds.CollectorName
	if colCreds.ApiBaseUrl != "" {
		se.SetBaseUrl(colCreds.ApiBaseUrl)
	}

	return colCreds, nil
}

// registerCollector registers the collector using registration API and returns
// the obtained collector credentials.
func (se *SumologicExtension) registerCollector(ctx context.Context, collectorName string) (credentials.CollectorCredentials, error) {
	u, err := url.Parse(se.BaseUrl())
	if err != nil {
		return credentials.CollectorCredentials{}, err
	}
	u.Path = registerUrl

	hostname, err := getHostname(se.logger)
	if err != nil {
		return credentials.CollectorCredentials{}, fmt.Errorf("cannot get hostname: %w", err)
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
		return credentials.CollectorCredentials{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), &buff)
	if err != nil {
		return credentials.CollectorCredentials{}, err
	}

	addClientCredentials(req,
		se.conf.Credentials,
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
		return credentials.CollectorCredentials{}, fmt.Errorf("failed to register the collector: %w", err)
	}

	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 400 {
		return se.handleRegistrationError(res)
	} else if res.StatusCode == 301 {
		// Use the URL from Location header for subsequent requests.
		u := strings.TrimSuffix(res.Header.Get("Location"), "/")
		se.SetBaseUrl(u)
		se.logger.Info("Redirected to a different deployment",
			zap.String("url", u),
		)
		return se.registerCollector(ctx, collectorName)
	}

	var resp api.OpenRegisterResponsePayload
	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
		return credentials.CollectorCredentials{}, err
	}

	if collectorName != resp.CollectorName {
		se.logger.Warn("Collector name already in use, registered modified name", zap.String("registered_name", resp.CollectorName))
	}

	return credentials.CollectorCredentials{
		CollectorName: collectorName,
		Credentials:   resp,
		ApiBaseUrl:    se.BaseUrl(),
	}, nil
}

// handleRegistrationError handles the collector registration errors and returns
// appropriate error for backoff handling and logging purposes.
func (se *SumologicExtension) handleRegistrationError(res *http.Response) (credentials.CollectorCredentials, error) {
	var errResponse api.ErrorResponsePayload
	if err := json.NewDecoder(res.Body).Decode(&errResponse); err != nil {
		var buff bytes.Buffer
		if _, errCopy := io.Copy(&buff, res.Body); errCopy != nil {
			return credentials.CollectorCredentials{}, fmt.Errorf(
				"failed to read the collector registration response body, status code: %d, err: %w",
				res.StatusCode, errCopy,
			)
		}
		return credentials.CollectorCredentials{}, fmt.Errorf(
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
		return credentials.CollectorCredentials{}, backoff.Permanent(fmt.Errorf(
			"failed to register the collector, got HTTP status code: %d",
			res.StatusCode,
		))
	}

	return credentials.CollectorCredentials{}, fmt.Errorf(
		"failed to register the collector, got HTTP status code: %d", res.StatusCode,
	)
}

// callRegisterWithBackoff calls registration using exponential backoff algorithm
// this loosely base on backoff.Retry function
func (se *SumologicExtension) registerCollectorWithBackoff(ctx context.Context, collectorName string) (credentials.CollectorCredentials, error) {
	se.backOff.Reset()
	for {
		creds, err := se.registerCollector(ctx, collectorName)
		if err == nil {
			se.logger = se.origLogger.With(
				zap.String(collectorNameField, creds.Credentials.CollectorName),
				zap.String(collectorIdField, creds.Credentials.CollectorId),
			)
			se.logger.Info("Collector registration finished successfully")

			return creds, nil
		}

		nbo := se.backOff.NextBackOff()
		// Return error if backoff reaches the limit or uncoverable error is spotted
		if _, ok := err.(*backoff.PermanentError); nbo == se.backOff.Stop || ok {
			return credentials.CollectorCredentials{}, fmt.Errorf("collector registration failed: %w", err)
		}

		t := time.NewTimer(nbo)
		defer t.Stop()

		select {
		case <-t.C:
		case <-ctx.Done():
			return credentials.CollectorCredentials{}, fmt.Errorf("collector registration cancelled: %w", ctx.Err())
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

	se.logger.Info("Heartbeat loop initialized. Starting to send hearbeat requests")
	timer := time.NewTimer(se.conf.HeartBeatInterval)
	for {
		select {
		case <-se.closeChan:
			se.logger.Info("Heartbeat sender turned off")
			return

		default:
			err := se.sendHeartbeatWithHTTPClient(ctx, se.httpClient)

			if err != nil {
				if errors.Is(err, errUnauthorizedHeartbeat) {
					se.logger.Warn("Heartbeat request unauthorized, re-registering the collector")
					colCreds, err := se.getCredentialsByRegistering(ctx)
					if err != nil {
						se.logger.Error("Heartbeat error, cannot register the collector", zap.Error(err))
						continue
					}

					// Inject newly received credentials into extension's configuration.
					if err = se.injectCredentials(colCreds); err != nil {
						se.logger.Error("Heartbeat error, cannot inject new collector credentials", zap.Error(err))
						continue
					}

					// Overwrite old logger fields with new collector name and ID.
					se.logger = se.origLogger.With(
						zap.String(collectorNameField, colCreds.Credentials.CollectorName),
						zap.String(collectorIdField, colCreds.Credentials.CollectorId),
					)

				} else {
					se.logger.Error("Heartbeat error", zap.Error(err))
				}
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

var errUnauthorizedHeartbeat = errors.New("heartbeat unauthorized")
var errUnauthorizedMetadata = errors.New("metadata update unauthorized")

type ErrorAPI struct {
	status int
	body   string
}

func (e ErrorAPI) Error() string {
	return fmt.Sprintf("API error (status code: %d): %s", e.status, e.body)
}

func (se *SumologicExtension) sendHeartbeatWithHTTPClient(ctx context.Context, httpClient *http.Client) error {
	u, err := url.Parse(se.BaseUrl() + heartbeatUrl)
	if err != nil {
		return fmt.Errorf("unable to parse heartbeat URL %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), nil)
	if err != nil {
		return fmt.Errorf("unable to create HTTP request %w", err)
	}

	addJSONHeaders(req)
	res, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("unable to send HTTP request: %w", err)
	}
	defer res.Body.Close()

	switch res.StatusCode {
	default:
		var buff bytes.Buffer
		if _, err := io.Copy(&buff, res.Body); err != nil {
			return fmt.Errorf(
				"failed to copy collector heartbeat response body, status code: %d, err: %w",
				res.StatusCode, err,
			)
		}

		return fmt.Errorf("collector heartbeat request failed: %w",
			ErrorAPI{
				status: res.StatusCode,
				body:   buff.String(),
			},
		)

	case http.StatusUnauthorized:
		return errUnauthorizedHeartbeat

	case http.StatusNoContent:
	}

	return nil
}

func getHostIpAddress() (string, error) {
	// This doesn't connect, we just need the connection object.
	c, err := net.Dial("udp", "255.255.255.255:53")
	if err != nil {
		return "", err
	}

	defer c.Close()
	a := c.LocalAddr().(*net.UDPAddr)
	h, _, err := net.SplitHostPort(a.String())
	if err != nil {
		return "", err
	}

	return h, nil
}

var sumoAppProcesses = map[string]string{
	"apache":                "apache",
	"apache2":               "apache",
	"httpd":                 "apache",
	"docker":                "docker",
	"elasticsearch":         "elasticsearch",
	"mysql-server":          "mysql",
	"mysqld":                "mysql",
	"nginx":                 "nginx",
	"postgresql":            "postgres",
	"postgresql-9.5":        "postgres",
	"rabbitmq-server":       "rabbitmq",
	"redis":                 "redis",
	"tomcat":                "tomcat",
	"kafka-server-start.sh": "kafka", // Need to test this, most common shell wrapper.
}

func filteredProcessList() ([]string, error) {
	var pl []string

	p, err := ps.Processes()
	if err != nil {
		return pl, err
	}

	for _, v := range p {
		e := strings.ToLower(v.Executable())
		if a, i := sumoAppProcesses[e]; i {
			pl = append(pl, a)
		}
	}

	return pl, nil
}

func discoverTags() (map[string]interface{}, error) {
	t := map[string]interface{}{
		"sumo.disco.enabled": "true",
	}

	pl, err := filteredProcessList()
	if err != nil {
		return t, err
	}

	for _, v := range pl {
		t["sumo.disco."+v] = "" // We do not currently need a value, save bytes.
	}

	return t, nil
}

func (se *SumologicExtension) updateMetadataWithHTTPClient(ctx context.Context, httpClient *http.Client) error {
	u, err := url.Parse(se.BaseUrl() + metadataUrl)
	if err != nil {
		return fmt.Errorf("unable to parse metadata URL %w", err)
	}

	info, err := host.Info()
	if err != nil {
		return err
	}

	hostname, err := getHostname(se.logger)
	if err != nil {
		return err
	}

	ip, err := getHostIpAddress()
	if err != nil {
		return err
	}

	td := map[string]interface{}{}

	if se.conf.DiscoverCollectorTags {
		td, err = discoverTags()
		if err != nil {
			return err
		}
	}

	for k, v := range se.conf.CollectorFields {
		td[k] = v
	}

	var buff bytes.Buffer
	if err = json.NewEncoder(&buff).Encode(api.OpenMetadataRequestPayload{
		HostDetails: api.OpenMetadataHostDetails{
			Name:        hostname,
			OsName:      info.OS,
			OsVersion:   info.PlatformVersion,
			Environment: se.conf.CollectorEnvironment,
		},
		CollectorDetails: api.OpenMetadataCollectorDetails{
			RunningVersion: se.buildVersion,
		},
		NetworkDetails: api.OpenMetadataNetworkDetails{
			HostIpAddress: ip,
		},
		TagDetails: td,
	}); err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), &buff)
	if err != nil {
		return fmt.Errorf("unable to create HTTP request %w", err)
	}

	addJSONHeaders(req)

	se.logger.Info("Updating collector metadata",
		zap.String("URL", u.String()),
		zap.String("body", buff.String()))

	res, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("unable to send HTTP request: %w", err)
	}
	defer res.Body.Close()

	switch res.StatusCode {
	default:
		var buff bytes.Buffer
		if _, err := io.Copy(&buff, res.Body); err != nil {
			return fmt.Errorf(
				"failed to copy collector metadata response body, status code: %d, err: %w",
				res.StatusCode, err,
			)
		}

		se.logger.Warn("Metadata API error response",
			zap.Int("status", res.StatusCode),
			zap.String("body", buff.String()))

		return fmt.Errorf("collector metadata request failed: %w",
			ErrorAPI{
				status: res.StatusCode,
				body:   buff.String(),
			},
		)

	case http.StatusUnauthorized:
		return errUnauthorizedMetadata
	case http.StatusNoContent:
	case http.StatusOK:
	}

	return nil
}

func (se *SumologicExtension) updateMetadataWithBackoff(ctx context.Context) error {
	se.backOff.Reset()
	for {
		err := se.updateMetadataWithHTTPClient(ctx, se.httpClient)
		if err == nil {
			return nil
		}

		nbo := se.backOff.NextBackOff()
		// Return error if backoff reaches the limit or uncoverable error is spotted
		if _, ok := err.(*backoff.PermanentError); nbo == se.backOff.Stop || ok {
			return fmt.Errorf("collector metadata update failed: %w", err)
		}

		t := time.NewTimer(nbo)
		defer t.Stop()

		select {
		case <-t.C:
		case <-ctx.Done():
			return fmt.Errorf("collector metadata update cancelled: %w", ctx.Err())
		}
	}
}

func (se *SumologicExtension) ComponentID() component.ID {
	return se.id
}

func (se *SumologicExtension) CollectorID() string {
	return se.registrationInfo.CollectorId
}

func (se *SumologicExtension) BaseUrl() string {
	se.baseUrlLock.RLock()
	defer se.baseUrlLock.RUnlock()
	return se.baseUrl
}

func (se *SumologicExtension) SetBaseUrl(baseUrl string) {
	se.baseUrlLock.Lock()
	se.baseUrl = baseUrl
	se.baseUrlLock.Unlock()
}

// WatchCredentialKey watches for credential key updates. It makes use of a
// channel close (done by injectCredentials) and string comparison with a
// known/previous credential key (old). This function allows components to be
// proactive when dealing with changes to authentication.
func (se *SumologicExtension) WatchCredentialKey(ctx context.Context, old string) string {
	se.credsNotifyLock.Lock()
	v, ch := se.registrationInfo.CollectorCredentialKey, se.credsNotifyUpdate
	se.credsNotifyLock.Unlock()

	for v == old {
		select {
		case <-ctx.Done():
			return v
		case <-ch:
			se.credsNotifyLock.Lock()
			v, ch = se.registrationInfo.CollectorCredentialKey, se.credsNotifyUpdate
			se.credsNotifyLock.Unlock()
		}
	}

	return v
}

// CreateCredentialsHeader produces an HTTP header containing authentication
// credentials. This function is for components that do not make use of the
// RoundTripper or have an HTTP request to build upon.
func (se *SumologicExtension) CreateCredentialsHeader() (http.Header, error) {
	id, key := se.registrationInfo.CollectorCredentialId, se.registrationInfo.CollectorCredentialKey

	if id == "" || key == "" {
		return nil, errors.New("collector credentials are not set")
	}

	token := base64.StdEncoding.EncodeToString(
		[]byte(id + ":" + key),
	)

	header := http.Header{}
	header.Set("Authorization", "Basic "+token)

	return header, nil
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

func (se *SumologicExtension) PerRPCCredentials() (grpccredentials.PerRPCCredentials, error) {
	return nil, errGRPCNotSupported
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

	// Delete the existing Authorization header so prevent sending both the old one
	// and the new one.
	req.Header.Del("Authorization")
	req.Header.Add("Authorization", "Basic "+token)
}

func addClientCredentials(req *http.Request, credentials accessCredentials) {
	var authHeaderValue string
	if credentials.InstallationToken != "" {
		authHeaderValue = fmt.Sprintf("Bearer %s", credentials.InstallationToken)
	}

	req.Header.Del("Authorization")
	req.Header.Add("Authorization", authHeaderValue)
}

// getHostname returns the host name consistently with the resource detection processor's defaults
// TODO: try to dynamically extract this from the resource processor in the pipeline
func getHostname(logger *zap.Logger) (string, error) {
	fqdnHostname, err := fqdn.FqdnHostname()
	if err == nil {
		return fqdnHostname, nil
	}
	logger.Debug("failed to get fqdn", zap.Error(err))

	return os.Hostname()
}
