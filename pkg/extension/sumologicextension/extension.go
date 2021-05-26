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
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.uber.org/zap"
)

type sumologicExtension struct {
	baseUrl          string
	conf             *Config
	logger           *zap.Logger
	registrationInfo OpenRegisterResponsePayload
	ch               chan struct{}
}

const (
	// TODO: fix
	niteBaseUrl = "https://nite-open-events.sumologic.net/"
)

func newSumologicExtension(conf *Config, logger *zap.Logger) (*sumologicExtension, error) {
	if conf.CollectorName == "" {
		return nil, errors.New("collector name is unset")
	}
	ch := make(chan bool)

	return &sumologicExtension{
		// TODO: don't hardcode
		baseUrl: niteBaseUrl,
		conf:    conf,
		logger:  logger,
		ch:      ch,
	}, nil
}

func (pm *sumologicExtension) Start(ctx context.Context, host component.Host) error {
	if err := pm.register(ctx, pm.conf.CollectorName); err != nil {
		return err
	}
	go pm.heartbeat(ctx)

	// -------------------------------------------------------------------------

	// exporters := host.GetExporters()
	// for kk, vv := range exporters {

	// 	for namedEnt, v := range vv {
	// 		pm.logger.Info(fmt.Sprintf("stopping %s exporter, called: %s, type: %s", kk, namedEnt.Name(), namedEnt.Type()))
	// 		pm.logger.Info(fmt.Sprintf("exporter type %T", v))

	// 		if err := v.Shutdown(ctx); err != nil {
	// 			panic(err)
	// 		}
	// 	}

	// 	time.Sleep(time.Second)

	// 	for namedEnt, v := range vv {
	// 		// v =
	// 		pm.logger.Info(fmt.Sprintf("starting %s exporter, called: %s, type: %s", kk, namedEnt.Name(), namedEnt.Type()))

	// 		if err := v.Start(ctx, host); err != nil {
	// 			panic(err)
	// 		}
	// 	}
	// }

	// -------------------------------------------------------------------------

	// factory := host.GetFactory(component.KindExporter, "sumologic").(component.ExporterFactory)
	// _ = factory

	// {
	// cfg1 := factory.CreateDefaultConfig().(*sumologicexporter.Config)
	// _ = cfg1
	// 	cfg1.SourceName = "sourceName1"
	// 	exporter, err := sumologicexporter.NewFactory().CreateMetricsExporter(
	// 		context.TODO(),
	// 		component.ExporterCreateParams{},
	// 		cfg1,
	// 	)
	// 	// exporter, err := factory.CreateMetricsExporter(ctx,
	// 	// 	component.ExporterCreateParams{},
	// 	// 	c,
	// 	// )

	// 	if err != nil {
	// 		panic(err)
	// 	}

	// 	if err = exporter.Start(context.TODO(), host); err != nil {
	// 		panic(err)
	// 	}
	// 	pm.exporters = append(pm.exporters, exporter)
	// }

	// c := &sumologicexporter.Config{
	// 	SourceName: "name2",
	// }
	// exporter, err := factory.CreateMetricsExporter(ctx,
	// 	component.ExporterCreateParams{},
	// 	c,
	// )

	// if err != nil {
	// 	panic(err)
	// }

	// if err = exporter.Start(context.TODO(), host); err != nil {
	// 	panic(err)
	// }
	// pm.exporters = append(pm.exporters, exporter)

	// rfactory := host.GetFactory(component.KindReceiver, "hostmetrics").(component.ReceiverFactory)
	// hostmetricsrec := hostmetricsreceiver.NewFactory().CreateDefaultConfig()
	// // 	rfactory.CreateMetricsReceiver(//ctx context.Context, params component.ReceiverCreateParams,
	// // 	cfg config.Receiver,
	// // 	nextConsumer consumer.Metrics,
	// // )

	// erec, err := rfactory.CreateMetricsReceiver(
	// 	ctx,
	// 	component.ReceiverCreateParams{},
	// 	hostmetricsrec,
	// 	exporter,
	// )

	// if err = erec.Start(context.TODO(), host); err != nil {
	// 	panic(err)
	// }

	return nil
}

// Shutdown is invoked during service shutdown.
func (pm *sumologicExtension) Shutdown(ctx context.Context) error {
	pm.ch <- true
	return nil
}

type FullRegisterKeyResponse struct {
	CollectorID   int64         `json:"collectorId"`
	CredentialID  string        `json:"credentialId"`
	CredentialKey string        `json:"credentialKey"`
	CustomerID    int64         `json:"customerId"`
	Errors        []interface{} `json:"errors"`
	HTTPCode      int64         `json:"httpCode"`
	Warnings      []interface{} `json:"warnings"`
}

func (pm *sumologicExtension) addClientCredentials(req *http.Request) {
	req.Header.Add("accessid", pm.conf.Credentials.AccessID)
	req.Header.Add("accesskey", pm.conf.Credentials.AccessKey)
}

func (pm *sumologicExtension) addJSONHeaders(req *http.Request) {
	req.Header.Add("Content-Type", "application/json")
}

func (pm *sumologicExtension) addCollectorCredentials(req *http.Request) {
	req.Header.Add("collectorCredentialId", pm.registrationInfo.CollectorCredentialId)
	req.Header.Add("collectorCredentialKey", pm.registrationInfo.CollectorCredentialKey)
}

type OpenRegisterRequestPayload struct {
	CollectorName string `json:"collectorName"`
	Ephemeral     bool   `json:"ephemeral"`
	Description   string `json:"description"`
	Hostname      string `json:"hostname"`
	Category      string `json:"category"`
	Timezone      string `json:"timeZone"`
}

type OpenRegisterResponsePayload struct {
	CollectorCredentialId  string `json:"collectorCredentialId"`
	CollectorCredentialKey string `json:"collectorCredentialKey"`
	CollectorId            string `json:"collectorId"`
}

func (pm *sumologicExtension) register(ctx context.Context, collectorName string) error {
	const registerUrl = "api/v1/collector/register"

	u, err := url.Parse(pm.baseUrl + registerUrl)
	if err != nil {
		return err
	}

	// TODO:
	// Just plaing hostname or we want to add some custom logic when setting
	// hostname in request?
	hostname, err := os.Hostname()
	if err != nil {
		return fmt.Errorf("cannot get hostname: %w", err)
	}

	payload := OpenRegisterRequestPayload{
		Ephemeral:     true,
		CollectorName: pm.conf.CollectorName,
		Description:   "Collector for test OTC registration purposes",
		Hostname:      hostname,
	}

	var buff bytes.Buffer
	if err = json.NewEncoder(&buff).Encode(payload); err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), &buff)
	if err != nil {
		return err
	}

	pm.addClientCredentials(req)
	pm.addJSONHeaders(req)

	pm.logger.Info("calling register API",
		zap.String("URL", u.String()),
	)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 400 {
		var buff bytes.Buffer
		if _, err := io.Copy(&buff, res.Body); err != nil {
			return fmt.Errorf(
				"failed to copy collector registration response body, status code: %d, err: %w",
				res.StatusCode, err,
			)
		}
		pm.logger.Error("Collector registration failed",
			zap.Any("response status code", res.StatusCode),
			zap.Any("response", buff.String()),
		)
		return nil
	}

	var resp OpenRegisterResponsePayload
	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
		return err
	}

	pm.logger.Info("Collector registered",
		zap.String("CollectorID", resp.CollectorId),
		zap.Any("response", resp),
	)

	pm.registrationInfo = resp

	return nil
}

func (pm *sumologicExtension) heartbeat(ctx context.Context) {
	const heartbeatUrl = "api/v1/collector/heartbeat"
	if pm.registrationInfo.CollectorCredentialId == "" || pm.registrationInfo.CollectorId == "" {
		pm.logger.Error("Collector not registered, cannot send heartbeat")
		return
	}
	u, err := url.Parse(pm.baseUrl + heartbeatUrl)
	if err != nil {
		pm.logger.Error("Unable to parse heartbeat URL ", zap.String("error: ", err.Error()))
		return
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), nil)
	if err != nil {
		pm.logger.Error("Unable to create HTTP request. ", zap.String("error: ", err.Error()))
		return
	}

	pm.addCollectorCredentials(req)
	pm.addJSONHeaders(req)

	pm.logger.Info("Heartbeat heartbeat API initialized. Starting sending hearbeat requests",
		zap.String("URL", u.String()),
	)
	for {
		select {
		case <-pm.ch:
			pm.logger.Info("Heartbeat sender turn off")
			return
		default:
			res, err := http.DefaultClient.Do(req)
			if err != nil {
				pm.logger.Error("Unable to send HTTP request. ", zap.String("error: ", err.Error()))
				return
			}
			defer res.Body.Close()
			if res.StatusCode != 204 {
				var buff bytes.Buffer
				if _, err := io.Copy(&buff, res.Body); err != nil {
					pm.logger.Error(
						"failed to copy collector heartbeat response body, status code: %d, err: %w",
						zap.Int("response status code", res.StatusCode),
						zap.String("error: ", err.Error()))
					return
				}
				pm.logger.Error("Collector heartbeat request failed",
					zap.Int("response status code", res.StatusCode),
					zap.String("response", buff.String()),
				)
				return
			}
			pm.logger.Info("Heartbeat sent")
			time.Sleep(15 * time.Second)
		}
	}
}

func (pm *sumologicExtension) CollectorID() string {
	return pm.registrationInfo.CollectorId
}

// Implement
// https://github.com/open-telemetry/opentelemetry-collector/blob/2e84285efc665798d76773b9901727e8836e9d8f/config/configauth/clientauth.go#L34-L39
// in order for this extension to be used as custom exporter authenticator.
func (pm *sumologicExtension) RoundTripper(base http.RoundTripper) (http.RoundTripper, error) {
	return roundTripper{
		accessID:  pm.conf.Credentials.AccessID,
		accessKey: pm.conf.Credentials.AccessKey,
		base:      base,
	}, nil
}

type roundTripper struct {
	accessID  string
	accessKey string
	base      http.RoundTripper
}

func (rt roundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	// TODO:
	// What is preferred: headers of basic auth?
	req.Header.Add("accessid", rt.accessID)
	req.Header.Add("accesskey", rt.accessKey)

	return rt.base.RoundTrip(req)
}
