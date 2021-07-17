// Copyright 2020 OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package sumologicexporter

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer/consumererror"
	"go.opentelemetry.io/collector/consumer/pdata"
	"go.opentelemetry.io/collector/exporter/exporterhelper"

	"github.com/open-telemetry/opentelemetry-collector-contrib/extension/sumologicextension"
)

const (
	logsDataUrl    = "/api/v1/collector/logs"
	metricsDataUrl = "/api/v1/collector/metrics"
	tracesDataUrl  = "/api/v1/collector/traces"
)

type sumologicexporter struct {
	sources             sourceFormats
	config              *Config
	client              *http.Client
	filter              filter
	prometheusFormatter prometheusFormatter
	graphiteFormatter   graphiteFormatter
	dataUrlMetrics      string
	dataUrlLogs         string
	dataUrlTraces       string
}

func initExporter(cfg *Config) (*sumologicexporter, error) {
	switch cfg.LogFormat {
	case JSONFormat:
	case TextFormat:
	case OTLPLogFormat:
	default:
		return nil, fmt.Errorf("unexpected log format: %s", cfg.LogFormat)
	}

	switch cfg.MetricFormat {
	case GraphiteFormat:
	case Carbon2Format:
	case PrometheusFormat:
	case OTLPMetricFormat:
	default:
		return nil, fmt.Errorf("unexpected metric format: %s", cfg.MetricFormat)
	}

	switch cfg.TraceFormat {
	case OTLPTraceFormat:
	default:
		return nil, fmt.Errorf("unexpected trace format: %s", cfg.TraceFormat)
	}

	switch cfg.CompressEncoding {
	case GZIPCompression:
	case DeflateCompression:
	case NoCompression:
	default:
		return nil, fmt.Errorf("unexpected compression encoding: %s", cfg.CompressEncoding)
	}

	if len(cfg.HTTPClientSettings.Endpoint) == 0 && cfg.HTTPClientSettings.Auth == nil {
		return nil, errors.New("no endpoint and no auth extension specified")
	}

	if _, err := url.Parse(cfg.HTTPClientSettings.Endpoint); err != nil {
		return nil, fmt.Errorf("failed parsing endpoint URL: %s; err: %w",
			cfg.HTTPClientSettings.Endpoint, err,
		)
	}

	if cfg.TranslateAttributes {
		cfg.SourceCategory = translateConfigValue(cfg.SourceCategory)
		cfg.SourceHost = translateConfigValue(cfg.SourceHost)
		cfg.SourceName = translateConfigValue(cfg.SourceName)
	}
	sfs, err := newSourceFormats(cfg)
	if err != nil {
		return nil, err
	}

	f, err := newFilter(cfg.MetadataAttributes)
	if err != nil {
		return nil, err
	}

	pf, err := newPrometheusFormatter()
	if err != nil {
		return nil, err
	}

	gf, err := newGraphiteFormatter(cfg.GraphiteTemplate)
	if err != nil {
		return nil, err
	}

	se := &sumologicexporter{
		config:  cfg,
		sources: sfs,
		// NOTE: client is now set in start()
		filter:              f,
		prometheusFormatter: pf,
		graphiteFormatter:   gf,
	}

	return se, nil
}

func newLogsExporter(
	cfg *Config,
	params component.ExporterCreateSettings,
) (component.LogsExporter, error) {
	se, err := initExporter(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize the logs exporter: %w", err)
	}

	return exporterhelper.NewLogsExporter(
		cfg,
		params.Logger,
		se.pushLogsData,
		// Disable exporterhelper Timeout, since we are using a custom mechanism
		// within exporter itself
		exporterhelper.WithTimeout(exporterhelper.TimeoutSettings{Timeout: 0}),
		exporterhelper.WithRetry(cfg.RetrySettings),
		exporterhelper.WithQueue(cfg.QueueSettings),
		exporterhelper.WithStart(se.start),
		exporterhelper.WithShutdown(se.shutdown),
	)
}

func newMetricsExporter(
	cfg *Config,
	params component.ExporterCreateSettings,
) (component.MetricsExporter, error) {
	se, err := initExporter(cfg)
	if err != nil {
		return nil, err
	}

	return exporterhelper.NewMetricsExporter(
		cfg,
		params.Logger,
		se.pushMetricsData,
		// Disable exporterhelper Timeout, since we are using a custom mechanism
		// within exporter itself
		exporterhelper.WithTimeout(exporterhelper.TimeoutSettings{Timeout: 0}),
		exporterhelper.WithRetry(cfg.RetrySettings),
		exporterhelper.WithQueue(cfg.QueueSettings),
		exporterhelper.WithStart(se.start),
		exporterhelper.WithShutdown(se.shutdown),
	)
}

func newTracesExporter(
	cfg *Config,
	params component.ExporterCreateSettings,
) (component.TracesExporter, error) {
	se, err := initExporter(cfg)
	if err != nil {
		return nil, err
	}

	return exporterhelper.NewTracesExporter(
		cfg,
		params.Logger,
		se.pushTracesData,
		// Disable exporterhelper Timeout, since we are using a custom mechanism
		// within exporter itself
		exporterhelper.WithTimeout(exporterhelper.TimeoutSettings{Timeout: 0}),
		exporterhelper.WithRetry(cfg.RetrySettings),
		exporterhelper.WithQueue(cfg.QueueSettings),
		exporterhelper.WithStart(se.start),
		exporterhelper.WithShutdown(se.shutdown),
	)
}

// pushLogsData groups data with common metadata and sends them as separate batched requests.
// It returns the number of unsent logs and an error which contains a list of dropped records
// so they can be handled by OTC retry mechanism
func (se *sumologicexporter) pushLogsData(ctx context.Context, ld pdata.Logs) error {
	var (
		currentMetadata  fields = newFields(pdata.NewAttributeMap())
		previousMetadata fields = newFields(pdata.NewAttributeMap())
		errs             []error
		droppedRecords   []pdata.LogRecord
		err              error
	)

	c, err := newCompressor(se.config.CompressEncoding)
	if err != nil {
		return consumererror.NewLogs(fmt.Errorf("failed to initialize compressor: %w", err), ld)
	}
	sdr := newSender(
		se.config,
		se.client,
		se.filter,
		se.sources,
		c,
		se.prometheusFormatter,
		se.graphiteFormatter,
		se.dataUrlMetrics,
		se.dataUrlLogs,
		se.dataUrlTraces,
	)

	// Iterate over ResourceLogs
	rls := ld.ResourceLogs()
	for i := 0; i < rls.Len(); i++ {
		rl := rls.At(i)

		ills := rl.InstrumentationLibraryLogs()
		// iterate over InstrumentationLibraryLogs
		for j := 0; j < ills.Len(); j++ {
			ill := ills.At(j)

			// iterate over Logs
			logs := ill.Logs()
			for k := 0; k < logs.Len(); k++ {
				log := logs.At(k)

				// copy resource attributes into logs attributes
				// log attributes have precedence over resource attributes
				attributes := log.Attributes()
				rl.Resource().Attributes().Range(func(k string, v pdata.AttributeValue) bool {
					attributes.Insert(k, v)
					return true
				})

				currentMetadata = sdr.filter.filterIn(attributes)

				if se.config.TranslateAttributes {
					translateAttributes(attributes)
					translateAttributes(currentMetadata.orig)
				}

				// If metadata differs from currently buffered, flush the buffer
				if currentMetadata.string() != previousMetadata.string() && previousMetadata.string() != "" {
					var dropped []pdata.LogRecord
					dropped, err = sdr.sendLogs(ctx, previousMetadata)
					if err != nil {
						errs = append(errs, err)
						droppedRecords = append(droppedRecords, dropped...)
					}
					sdr.cleanLogsBuffer()
				}

				// assign metadata
				previousMetadata = currentMetadata

				// add log to the buffer
				var dropped []pdata.LogRecord
				dropped, err = sdr.batchLog(ctx, log, previousMetadata)
				if err != nil {
					droppedRecords = append(droppedRecords, dropped...)
					errs = append(errs, err)
				}
			}
		}
	}

	// Flush pending logs
	dropped, err := sdr.sendLogs(ctx, previousMetadata)
	if err != nil {
		droppedRecords = append(droppedRecords, dropped...)
		errs = append(errs, err)
	}

	if len(droppedRecords) > 0 {
		// Move all dropped records to Logs
		droppedLogs := pdata.NewLogs()
		rls = droppedLogs.ResourceLogs()
		ills := rls.AppendEmpty().InstrumentationLibraryLogs()
		logs := ills.AppendEmpty().Logs()

		for _, log := range droppedRecords {
			logs.Append(log)
		}

		return consumererror.NewLogs(consumererror.Combine(errs), droppedLogs)
	}

	return nil
}

// pushMetricsData groups data with common metadata and send them as separate batched requests
// it returns number of unsent metrics and error which contains list of dropped records
// so they can be handle by the OTC retry mechanism
func (se *sumologicexporter) pushMetricsData(ctx context.Context, md pdata.Metrics) error {
	var (
		currentMetadata  fields = newFields(pdata.NewAttributeMap())
		previousMetadata fields = newFields(pdata.NewAttributeMap())
		errs             []error
		droppedRecords   []metricPair
		attributes       pdata.AttributeMap
	)

	c, err := newCompressor(se.config.CompressEncoding)
	if err != nil {
		return consumererror.NewMetrics(fmt.Errorf("failed to initialize compressor: %w", err), md)
	}
	sdr := newSender(
		se.config,
		se.client,
		se.filter,
		se.sources,
		c,
		se.prometheusFormatter,
		se.graphiteFormatter,
		se.dataUrlMetrics,
		se.dataUrlLogs,
		se.dataUrlTraces,
	)

	// Iterate over ResourceMetrics
	rms := md.ResourceMetrics()
	for i := 0; i < rms.Len(); i++ {
		rm := rms.At(i)

		attributes = rm.Resource().Attributes()
		currentMetadata = sdr.filter.filterIn(attributes)

		if se.config.TranslateAttributes {
			translateAttributes(attributes)
			translateAttributes(currentMetadata.orig)
		}

		// iterate over InstrumentationLibraryMetrics
		ilms := rm.InstrumentationLibraryMetrics()
		for j := 0; j < ilms.Len(); j++ {
			ilm := ilms.At(j)

			// iterate over Metrics
			ms := ilm.Metrics()
			for k := 0; k < ms.Len(); k++ {
				m := ms.At(k)
				mp := metricPair{
					metric:     m,
					attributes: attributes,
				}

				// If metadata differs from currently buffered, flush the buffer
				if currentMetadata.string() != previousMetadata.string() && previousMetadata.string() != "" {
					var dropped []metricPair
					dropped, err = sdr.sendMetrics(ctx, previousMetadata)
					if err != nil {
						errs = append(errs, err)
						droppedRecords = append(droppedRecords, dropped...)
					}
					sdr.cleanMetricBuffer()
				}

				// assign metadata
				previousMetadata = currentMetadata
				var dropped []metricPair
				// add metric to the buffer
				dropped, err = sdr.batchMetric(ctx, mp, currentMetadata)
				if err != nil {
					droppedRecords = append(droppedRecords, dropped...)
					errs = append(errs, err)
				}
			}
		}
	}

	// Flush pending metrics
	dropped, err := sdr.sendMetrics(ctx, previousMetadata)
	if err != nil {
		droppedRecords = append(droppedRecords, dropped...)
		errs = append(errs, err)
	}

	if len(droppedRecords) > 0 {
		// Move all dropped records to Metrics
		droppedMetrics := pdata.NewMetrics()
		rms := droppedMetrics.ResourceMetrics()
		rms.Resize(len(droppedRecords))
		for num, record := range droppedRecords {
			rm := droppedMetrics.ResourceMetrics().At(num)
			record.attributes.CopyTo(rm.Resource().Attributes())

			ilms := rm.InstrumentationLibraryMetrics()
			ilms.AppendEmpty().Metrics().Append(record.metric)
		}

		return consumererror.NewMetrics(consumererror.Combine(errs), droppedMetrics)
	}

	return nil
}

func (se *sumologicexporter) pushTracesData(ctx context.Context, td pdata.Traces) error {
	var currentMetadata fields = newFields(pdata.NewAttributeMap())
	c, err := newCompressor(se.config.CompressEncoding)
	if err != nil {
		return consumererror.NewTraces(fmt.Errorf("failed to initialize compressor: %w", err), td)
	}
	sdr := newSender(
		se.config,
		se.client,
		se.filter,
		se.sources,
		c,
		se.prometheusFormatter,
		se.graphiteFormatter,
		se.dataUrlMetrics,
		se.dataUrlLogs,
		se.dataUrlTraces,
	)
	err = sdr.sendTraces(ctx, td, currentMetadata)
	if err != nil {
		return err
	}

	return nil
}

func (se *sumologicexporter) start(ctx context.Context, host component.Host) error {
	var (
		ext          *sumologicextension.SumologicExtension
		foundSumoExt bool
	)

	httpSettings := se.config.HTTPClientSettings

	for _, e := range host.GetExtensions() {
		v, ok := e.(*sumologicextension.SumologicExtension)
		if ok && httpSettings.Auth.AuthenticatorName == v.ComponentID() {
			ext = v
			foundSumoExt = true
			break
		}
	}

	if httpSettings.Endpoint == "" && httpSettings.Auth != nil &&
		strings.HasPrefix(httpSettings.Auth.AuthenticatorName, "sumologic") {
		// If user specified using sumologicextension as auth but none was
		// found then return an error.
		if !foundSumoExt {
			return fmt.Errorf(
				"sumologic was specified as auth extension (named: %q) but "+
					"a matching extension was not found in the config, "+
					"please re-check the config and/or define the sumologicextension",
				httpSettings.Auth.AuthenticatorName,
			)
		}

		// If we're using sumologicextension as authentication extension and
		// endpoint was not set then send data on a collector generic ingest URL
		// with authentication set by sumologicextension.

		u, err := url.Parse(ext.BaseUrl())
		if err != nil {
			return fmt.Errorf("failed to parse API base URL from sumologicextension: %w", err)
		}
		u.Path = logsDataUrl
		se.dataUrlLogs = u.String()
		u.Path = metricsDataUrl
		se.dataUrlMetrics = u.String()
		u.Path = tracesDataUrl
		se.dataUrlTraces = u.String()
	} else if httpSettings.Endpoint != "" {
		se.dataUrlLogs = httpSettings.Endpoint
		se.dataUrlMetrics = httpSettings.Endpoint
		se.dataUrlTraces = httpSettings.Endpoint

		// Clean authenticator if set to sumologic.
		// Setting to null in configuration doesn't work, so we have to force it that way.
		if httpSettings.Auth != nil && strings.HasPrefix(httpSettings.Auth.AuthenticatorName, "sumologic") {
			httpSettings.Auth = nil
		}
	} else {
		return fmt.Errorf("no auth extension and no endpoint specified")
	}

	client, err := httpSettings.ToClient(host.GetExtensions())
	if err != nil {
		return fmt.Errorf("failed to create HTTP Client: %w", err)
	}

	se.client = client
	return nil
}

func (se *sumologicexporter) shutdown(context.Context) error {
	return nil
}
