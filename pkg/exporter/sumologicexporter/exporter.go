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
	"sync"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer/consumererror"
	"go.opentelemetry.io/collector/exporter/exporterhelper"
	"go.opentelemetry.io/collector/model/pdata"
	"go.uber.org/multierr"
	"go.uber.org/zap"

	"github.com/SumoLogic/sumologic-otel-collector/pkg/extension/sumologicextension"
)

const (
	logsDataUrl    = "/api/v1/collector/logs"
	metricsDataUrl = "/api/v1/collector/metrics"
	tracesDataUrl  = "/api/v1/collector/traces"
)

type sumologicexporter struct {
	sources sourceFormats
	config  *Config
	host    component.Host
	logger  *zap.Logger

	clientLock sync.RWMutex
	client     *http.Client

	compressorPool sync.Pool

	filter              filter
	prometheusFormatter prometheusFormatter
	graphiteFormatter   graphiteFormatter

	// Lock around data URLs is needed because the reconfiguration of the exporter
	// can happen asynchronously whenever the exporter is re registering.
	dataUrlsLock   sync.RWMutex
	dataUrlMetrics string
	dataUrlLogs    string
	dataUrlTraces  string
}

func initExporter(cfg *Config, createSettings component.ExporterCreateSettings) (*sumologicexporter, error) {
	if cfg.TranslateAttributes {
		cfg.SourceCategory = translateConfigValue(cfg.SourceCategory)
		cfg.SourceHost = translateConfigValue(cfg.SourceHost)
		cfg.SourceName = translateConfigValue(cfg.SourceName)
	}
	sfs, err := newSourceFormats(cfg)
	if err != nil {
		return nil, err
	}

	cfg.MetadataAttributes = addSourceMetadataFields(cfg.MetadataAttributes)
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
		logger:  createSettings.Logger,
		sources: sfs,
		compressorPool: sync.Pool{
			New: func() any {
				c, err := newCompressor(cfg.CompressEncoding)
				if err != nil {
					return fmt.Errorf("failed to initialize compressor: %w", err)
				}
				return c
			},
		},
		// NOTE: client is now set in start()
		filter:              f,
		prometheusFormatter: pf,
		graphiteFormatter:   gf,
	}

	se.logger.Info(
		"Sumo Logic Exporter configured",
		zap.String("log_format", string(cfg.LogFormat)),
		zap.String("metric_format", string(cfg.MetricFormat)),
		zap.String("trace_format", string(cfg.TraceFormat)),
	)

	return se, nil
}

// addSourceMetadataFields adds source related attribute names to the list of
// metadata attributes.
func addSourceMetadataFields(metadataAttributes []string) []string {
	var (
		sourceCategory bool
		sourceHost     bool
		sourceName     bool
	)

	for _, attr := range metadataAttributes {
		switch attr {
		case attributeKeySourceCategory:
			sourceCategory = true
		case attributeKeySourceHost:
			sourceHost = true
		case attributeKeySourceName:
			sourceName = true
		default:
		}
	}

	if !sourceCategory {
		metadataAttributes = append(metadataAttributes, attributeKeySourceCategory)
	}
	if !sourceHost {
		metadataAttributes = append(metadataAttributes, attributeKeySourceHost)
	}
	if !sourceName {
		metadataAttributes = append(metadataAttributes, attributeKeySourceName)
	}

	return metadataAttributes
}

func newLogsExporter(
	cfg *Config,
	params component.ExporterCreateSettings,
) (component.LogsExporter, error) {
	se, err := initExporter(cfg, params)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize the logs exporter: %w", err)
	}

	return exporterhelper.NewLogsExporter(
		cfg,
		params,
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
	se, err := initExporter(cfg, params)
	if err != nil {
		return nil, err
	}

	return exporterhelper.NewMetricsExporter(
		cfg,
		params,
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
	se, err := initExporter(cfg, params)
	if err != nil {
		return nil, err
	}

	return exporterhelper.NewTracesExporter(
		cfg,
		params,
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
	compr, err := se.getCompressor()
	if err != nil {
		return consumererror.NewLogs(err, ld)
	}
	defer se.compressorPool.Put(compr)

	logsUrl, metricsUrl, tracesUrl := se.getDataURLs()
	sdr := newSender(
		se.logger,
		se.config,
		se.getHTTPClient(),
		se.filter,
		se.sources,
		compr,
		se.prometheusFormatter,
		se.graphiteFormatter,
		metricsUrl,
		logsUrl,
		tracesUrl,
	)

	// Follow different execution path for OTLP format
	if sdr.config.LogFormat == OTLPLogFormat {
		if err := sdr.sendOTLPLogs(ctx, ld); err != nil {
			se.handleUnauthorizedErrors(ctx, err)
			return consumererror.NewLogs(err, ld)
		}
		return nil
	}

	var (
		currentMetadata  fields = newFields(pdata.NewAttributeMap())
		previousMetadata fields = newFields(pdata.NewAttributeMap())
		errs             []error
		droppedRecords   []logPair
	)

	// Iterate over ResourceLogs
	rls := ld.ResourceLogs()
	for i := 0; i < rls.Len(); i++ {
		rl := rls.At(i)

		// drop routing attribute
		se.dropRoutingAttribute(rl.Resource().Attributes())

		ills := rl.InstrumentationLibraryLogs()
		// iterate over InstrumentationLibraryLogs
		for j := 0; j < ills.Len(); j++ {
			ill := ills.At(j)

			// iterate over Logs
			logs := ill.LogRecords()
			for k := 0; k < logs.Len(); k++ {
				log := logs.At(k)
				logAttrs := log.Attributes()

				// copy resource attributes into logs attributes
				// log attributes have precedence over resource attributes
				attributes := pdata.NewAttributeMap()
				attributes.EnsureCapacity(
					logAttrs.Len() + rl.Resource().Attributes().Len(),
				)
				logAttrs.CopyTo(attributes)
				rl.Resource().Attributes().Range(func(k string, v pdata.AttributeValue) bool {
					attributes.Insert(k, v)
					return true
				})

				// Put merged attributes into logPair
				lp := logPair{
					log:        log,
					attributes: attributes,
				}

				currentMetadata = sdr.filter.filterIn(attributes)

				if se.config.TranslateAttributes {
					currentMetadata.translateAttributes()
				}

				// If metadata differs from currently buffered, flush the buffer
				if !currentMetadata.equals(previousMetadata) && !previousMetadata.isEmpty() {
					var dropped []logPair
					dropped, err = sdr.sendNonOTLPLogs(ctx, previousMetadata)
					if err != nil {
						errs = append(errs, err)
						droppedRecords = append(droppedRecords, dropped...)
					}
					sdr.cleanLogsBuffer()
				}

				// assign metadata
				previousMetadata = currentMetadata

				// add log to the buffer
				var dropped []logPair
				dropped, err = sdr.batchLog(ctx, lp, previousMetadata)
				if err != nil {
					droppedRecords = append(droppedRecords, dropped...)
					errs = append(errs, err)
				}
			}
		}
	}

	// Flush pending logs
	dropped, err := sdr.sendNonOTLPLogs(ctx, previousMetadata)
	if err != nil {
		droppedRecords = append(droppedRecords, dropped...)
		errs = append(errs, err)
	}

	if len(droppedRecords) > 0 {
		// Move all dropped records to Logs
		droppedLogs := pdata.NewLogs()
		rls = droppedLogs.ResourceLogs()
		ills := rls.AppendEmpty().InstrumentationLibraryLogs()
		logs := ills.AppendEmpty().LogRecords()
		logs.EnsureCapacity(len(droppedRecords))

		for _, lp := range droppedRecords {
			lp.log.CopyTo(logs.AppendEmpty())
		}

		errs = deduplicateErrors(errs)
		se.handleUnauthorizedErrors(ctx, errs...)
		return consumererror.NewLogs(multierr.Combine(errs...), droppedLogs)
	}

	return nil
}

// pushMetricsData groups data with common metadata and send them as separate batched requests
// it returns number of unsent metrics and error which contains list of dropped records
// so they can be handle by the OTC retry mechanism
func (se *sumologicexporter) pushMetricsData(ctx context.Context, md pdata.Metrics) error {
	compr, err := se.getCompressor()
	if err != nil {
		return consumererror.NewMetrics(err, md)
	}
	defer se.compressorPool.Put(compr)

	logsUrl, metricsUrl, tracesUrl := se.getDataURLs()
	sdr := newSender(
		se.logger,
		se.config,
		se.getHTTPClient(),
		se.filter,
		se.sources,
		compr,
		se.prometheusFormatter,
		se.graphiteFormatter,
		metricsUrl,
		logsUrl,
		tracesUrl,
	)

	// Follow different execution path for OTLP format
	if sdr.config.MetricFormat == OTLPMetricFormat {
		if err := sdr.sendOTLPMetrics(ctx, md); err != nil {
			se.handleUnauthorizedErrors(ctx, err)
			return consumererror.NewMetrics(err, md)
		}
		return nil
	}

	var (
		currentMetadata  fields = newFields(pdata.NewAttributeMap())
		previousMetadata fields = newFields(pdata.NewAttributeMap())
		errs             []error
		droppedRecords   []metricPair
	)

	// Iterate over ResourceMetrics
	rms := md.ResourceMetrics()
	for i := 0; i < rms.Len(); i++ {
		rm := rms.At(i)

		attributes := rm.Resource().Attributes()

		// drop routing attribute
		se.dropRoutingAttribute(attributes)

		currentMetadata = sdr.filter.filterIn(attributes)

		if se.config.TranslateAttributes {
			attributes = translateAttributes(attributes)
			currentMetadata.translateAttributes()
		}

		// iterate over InstrumentationLibraryMetrics
		ilms := rm.InstrumentationLibraryMetrics()
		for j := 0; j < ilms.Len(); j++ {
			ilm := ilms.At(j)

			// iterate over Metrics
			ms := ilm.Metrics()
			for k := 0; k < ms.Len(); k++ {
				m := ms.At(k)

				if se.config.TranslateTelegrafMetrics {
					translateTelegrafMetric(m)
				}

				mp := metricPair{
					metric:     m,
					attributes: attributes,
				}

				// If metadata differs from currently buffered, flush the buffer
				if !currentMetadata.equals(previousMetadata) && !previousMetadata.isEmpty() {
					var dropped []metricPair
					var err error
					dropped, err = sdr.sendNonOTLPMetrics(ctx, previousMetadata)
					if err != nil {
						errs = append(errs, err)
						droppedRecords = append(droppedRecords, dropped...)
					}
					sdr.cleanMetricBuffer()
				}

				// assign metadata
				previousMetadata = currentMetadata
				var dropped []metricPair
				var err error
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
	dropped, err := sdr.sendNonOTLPMetrics(ctx, previousMetadata)
	if err != nil {
		droppedRecords = append(droppedRecords, dropped...)
		errs = append(errs, err)
	}

	if len(droppedRecords) > 0 {
		// Move all dropped records to Metrics
		droppedMetrics := pdata.NewMetrics()
		rms := droppedMetrics.ResourceMetrics()
		rms.EnsureCapacity(len(droppedRecords))
		for _, record := range droppedRecords {
			rm := droppedMetrics.ResourceMetrics().AppendEmpty()
			record.attributes.CopyTo(rm.Resource().Attributes())

			ilms := rm.InstrumentationLibraryMetrics()
			record.metric.CopyTo(ilms.AppendEmpty().Metrics().AppendEmpty())
		}

		errs = deduplicateErrors(errs)
		se.handleUnauthorizedErrors(ctx, errs...)
		return consumererror.NewMetrics(multierr.Combine(errs...), droppedMetrics)
	}

	return nil
}

// handleUnauthorizedErrors checks if any of the provided errors is an unauthorized error.
// In which case it triggers exporter reconfiguration which in turn takes the credentials
// from sumologicextension which at this point should already detect the problem with
// authorization (via heartbeats) and prepare new collector credentials to be available.
func (se *sumologicexporter) handleUnauthorizedErrors(ctx context.Context, errs ...error) {
	for _, err := range errs {
		if errors.Is(err, errUnauthorized) {
			se.logger.Warn("Received unauthorized status code, triggering reconfiguration")
			if errC := se.configure(ctx); errC != nil {
				se.logger.Error("Error configuring the exporter with new credentials", zap.Error(err))
			} else {
				// It's enough to successfully reconfigure the exporter just once.
				return
			}
		}
	}
}

func (se *sumologicexporter) pushTracesData(ctx context.Context, td pdata.Traces) error {
	compr, err := se.getCompressor()
	if err != nil {
		return consumererror.NewTraces(err, td)
	}
	defer se.compressorPool.Put(compr)

	logsUrl, metricsUrl, tracesUrl := se.getDataURLs()
	sdr := newSender(
		se.logger,
		se.config,
		se.getHTTPClient(),
		se.filter,
		se.sources,
		compr,
		se.prometheusFormatter,
		se.graphiteFormatter,
		metricsUrl,
		logsUrl,
		tracesUrl,
	)

	// Drop routing attribute from ResourceSpans
	rss := td.ResourceSpans()
	for i := 0; i < rss.Len(); i++ {
		se.dropRoutingAttribute(rss.At(i).Resource().Attributes())
	}

	err = sdr.sendTraces(ctx, td)
	se.handleUnauthorizedErrors(ctx, err)
	return err
}

func (se *sumologicexporter) getCompressor() (compressor, error) {
	switch c := se.compressorPool.Get().(type) {
	case error:
		return compressor{}, fmt.Errorf("%v", c)
	case compressor:
		return c, nil
	default:
		return compressor{}, fmt.Errorf("unknown compressor type: %T", c)
	}
}

func (se *sumologicexporter) start(ctx context.Context, host component.Host) error {
	se.host = host
	return se.configure(ctx)
}

func (se *sumologicexporter) configure(ctx context.Context) error {
	var (
		ext          *sumologicextension.SumologicExtension
		foundSumoExt bool
	)

	httpSettings := se.config.HTTPClientSettings

	for _, e := range se.host.GetExtensions() {
		v, ok := e.(*sumologicextension.SumologicExtension)
		if ok && httpSettings.Auth.AuthenticatorID == v.ComponentID() {
			ext = v
			foundSumoExt = true
			break
		}
	}

	if httpSettings.Endpoint == "" && httpSettings.Auth != nil &&
		string(httpSettings.Auth.AuthenticatorID.Type()) == "sumologic" {
		// If user specified using sumologicextension as auth but none was
		// found then return an error.
		if !foundSumoExt {
			return fmt.Errorf(
				"sumologic was specified as auth extension (named: %q) but "+
					"a matching extension was not found in the config, "+
					"please re-check the config and/or define the sumologicextension",
				httpSettings.Auth.AuthenticatorID.String(),
			)
		}

		// If we're using sumologicextension as authentication extension and
		// endpoint was not set then send data on a collector generic ingest URL
		// with authentication set by sumologicextension.

		u, err := url.Parse(ext.BaseUrl())
		if err != nil {
			return fmt.Errorf("failed to parse API base URL from sumologicextension: %w", err)
		}

		logsUrl := *u
		logsUrl.Path = logsDataUrl
		metricsUrl := *u
		metricsUrl.Path = metricsDataUrl
		tracesUrl := *u
		tracesUrl.Path = tracesDataUrl
		se.setDataURLs(logsUrl.String(), metricsUrl.String(), tracesUrl.String())

	} else if httpSettings.Endpoint != "" {
		se.setDataURLs(httpSettings.Endpoint, httpSettings.Endpoint, httpSettings.Endpoint)

		// Clean authenticator if set to sumologic.
		// Setting to null in configuration doesn't work, so we have to force it that way.
		if httpSettings.Auth != nil && string(httpSettings.Auth.AuthenticatorID.Type()) == "sumologic" {
			httpSettings.Auth = nil
		}
	} else {
		return fmt.Errorf("no auth extension and no endpoint specified")
	}

	client, err := httpSettings.ToClient(se.host.GetExtensions(), component.TelemetrySettings{})
	if err != nil {
		return fmt.Errorf("failed to create HTTP Client: %w", err)
	}

	se.setHTTPClient(client)
	return nil
}

func (se *sumologicexporter) setHTTPClient(client *http.Client) {
	se.clientLock.Lock()
	se.client = client
	se.clientLock.Unlock()
}

func (se *sumologicexporter) getHTTPClient() *http.Client {
	se.clientLock.RLock()
	defer se.clientLock.RUnlock()
	return se.client
}

func (se *sumologicexporter) setDataURLs(logs, metrics, traces string) {
	se.dataUrlsLock.Lock()
	se.dataUrlLogs, se.dataUrlMetrics, se.dataUrlTraces = logs, metrics, traces
	se.dataUrlsLock.Unlock()
}

func (se *sumologicexporter) getDataURLs() (logs, metrics, traces string) {
	se.dataUrlsLock.RLock()
	defer se.dataUrlsLock.RUnlock()
	return se.dataUrlLogs, se.dataUrlMetrics, se.dataUrlTraces
}

func (se *sumologicexporter) shutdown(context.Context) error {
	return nil
}

func (se *sumologicexporter) dropRoutingAttribute(attr pdata.AttributeMap) {
	attr.Delete(se.config.DropRoutingAttribute)
}
