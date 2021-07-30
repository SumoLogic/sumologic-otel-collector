// Copyright 2020, OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package sumologicexporter

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"go.opentelemetry.io/collector/consumer/consumererror"
	"go.opentelemetry.io/collector/model/otlp"
	"go.opentelemetry.io/collector/model/pdata"
	tracetranslator "go.opentelemetry.io/collector/translator/trace"
)

var (
	tracesMarshaler  = otlp.NewProtobufTracesMarshaler()
	metricsMarshaler = otlp.NewProtobufMetricsMarshaler()
	logsMarshaler    = otlp.NewProtobufLogsMarshaler()
)

type appendResponse struct {
	// sent gives information if the data was sent or not
	sent bool
	// appended keeps state of appending new log line to the body
	appended bool
}

// metricPair represents information required to send one metric to the Sumo Logic
type metricPair struct {
	attributes pdata.AttributeMap
	metric     pdata.Metric
}

type sender struct {
	logBuffer           []pdata.LogRecord
	metricBuffer        []metricPair
	config              *Config
	client              *http.Client
	filter              filter
	sources             sourceFormats
	compressor          compressor
	prometheusFormatter prometheusFormatter
	graphiteFormatter   graphiteFormatter
	dataUrlMetrics      string
	dataUrlLogs         string
	dataUrlTraces       string
}

const (
	logKey string = "log"
	// maxBufferSize defines size of the logBuffer (maximum number of pdata.LogRecord entries)
	maxBufferSize int = 1024 * 1024

	headerContentType     string = "Content-Type"
	headerContentEncoding string = "Content-Encoding"
	headerClient          string = "X-Sumo-Client"
	headerHost            string = "X-Sumo-Host"
	headerName            string = "X-Sumo-Name"
	headerCategory        string = "X-Sumo-Category"
	headerFields          string = "X-Sumo-Fields"

	contentTypeLogs       string = "application/x-www-form-urlencoded"
	contentTypePrometheus string = "application/vnd.sumologic.prometheus"
	contentTypeCarbon2    string = "application/vnd.sumologic.carbon2"
	contentTypeGraphite   string = "application/vnd.sumologic.graphite"
	contentTypeOTLP       string = "application/x-protobuf"

	contentEncodingGzip    string = "gzip"
	contentEncodingDeflate string = "deflate"
)

func newAppendResponse() appendResponse {
	return appendResponse{
		appended: true,
	}
}

func newSender(
	cfg *Config,
	cl *http.Client,
	f filter,
	s sourceFormats,
	c compressor,
	pf prometheusFormatter,
	gf graphiteFormatter,
	metricsUrl string,
	logsUrl string,
	tracesUrl string,
) *sender {
	return &sender{
		config:              cfg,
		client:              cl,
		filter:              f,
		sources:             s,
		compressor:          c,
		prometheusFormatter: pf,
		graphiteFormatter:   gf,
		dataUrlMetrics:      metricsUrl,
		dataUrlLogs:         logsUrl,
		dataUrlTraces:       tracesUrl,
	}
}

// send sends data to sumologic
func (s *sender) send(ctx context.Context, pipeline PipelineType, body io.Reader, flds fields) error {
	data, err := s.compressor.compress(body)
	if err != nil {
		return err
	}

	req, err := s.createRequest(ctx, pipeline, data)
	if err != nil {
		return err
	}

	if err := s.addRequestHeaders(req, pipeline, flds); err != nil {
		return err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return fmt.Errorf("error during sending data: %s", resp.Status)
	}
	return nil
}

func (s *sender) createRequest(ctx context.Context, pipeline PipelineType, data io.Reader) (*http.Request, error) {
	var url string
	if s.config.HTTPClientSettings.Endpoint == "" {
		switch pipeline {
		case MetricsPipeline:
			url = s.dataUrlMetrics
		case LogsPipeline:
			url = s.dataUrlLogs
		case TracesPipeline:
			url = s.dataUrlTraces
		default:
			return nil, fmt.Errorf("unknown pipeline type: %s", pipeline)
		}
	} else {
		url = s.config.HTTPClientSettings.Endpoint
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, data)
	if err != nil {
		return req, err
	}

	return req, err
}

// logToText converts LogRecord to a plain text line, returns it and error eventually
func (s *sender) logToText(record pdata.LogRecord) string {
	return tracetranslator.AttributeValueToString(record.Body())
}

// logToJSON converts LogRecord to a json line, returns it and error eventually
func (s *sender) logToJSON(record pdata.LogRecord) (string, error) {
	data := s.filter.filterOut(record.Attributes())
	data.orig.Upsert(logKey, record.Body())

	nextLine, err := json.Marshal(tracetranslator.AttributeMapToMap(data.orig))
	if err != nil {
		return "", err
	}

	return bytes.NewBuffer(nextLine).String(), nil
}

// sendLogs sends log records from the logBuffer formatted according
// to configured LogFormat and as the result of execution
// returns array of records which has not been sent correctly and error
func (s *sender) sendLogs(ctx context.Context, flds fields) ([]pdata.LogRecord, error) {

	// Follow different execution path for OTLP format
	if s.config.LogFormat == OTLPLogFormat {
		return s.sendOTLPLogs(ctx, flds)
	}

	var (
		body           strings.Builder
		errs           []error
		droppedRecords []pdata.LogRecord
		currentRecords []pdata.LogRecord
	)

	for _, record := range s.logBuffer {
		var formattedLine string
		var err error

		switch s.config.LogFormat {
		case TextFormat:
			formattedLine = s.logToText(record)
		case JSONFormat:
			formattedLine, err = s.logToJSON(record)
		default:
			err = errors.New("unexpected log format")
		}

		if err != nil {
			droppedRecords = append(droppedRecords, record)
			errs = append(errs, err)
			continue
		}

		ar, err := s.appendAndSend(ctx, formattedLine, LogsPipeline, &body, flds)
		if err != nil {
			errs = append(errs, err)
			if ar.sent {
				droppedRecords = append(droppedRecords, currentRecords...)
			}

			if !ar.appended {
				droppedRecords = append(droppedRecords, record)
			}
		}

		// If data was sent, cleanup the currentTimeSeries counter
		if ar.sent {
			currentRecords = currentRecords[:0]
		}

		// If log has been appended to body, increment the currentTimeSeries
		if ar.appended {
			currentRecords = append(currentRecords, record)
		}
	}

	if body.Len() > 0 {
		if err := s.send(ctx, LogsPipeline, strings.NewReader(body.String()), flds); err != nil {
			errs = append(errs, err)
			droppedRecords = append(droppedRecords, currentRecords...)
		}
	}

	if len(errs) > 0 {
		return droppedRecords, consumererror.Combine(errs)
	}
	return droppedRecords, nil
}

// sendLogs sends log records from the logBuffer in OTLP format and as a result
// it returns an array of records which has not been sent correctly and an error.
// TODO: add support for HTTP limits
func (s *sender) sendOTLPLogs(ctx context.Context, flds fields) ([]pdata.LogRecord, error) {
	ld := pdata.NewLogs()
	rl := ld.ResourceLogs().AppendEmpty()
	ill := rl.InstrumentationLibraryLogs().AppendEmpty()
	logs := ill.Logs()
	logs.EnsureCapacity(len(s.logBuffer))
	for _, record := range s.logBuffer {
		record.CopyTo(logs.AppendEmpty())
	}

	body, err := logsMarshaler.MarshalLogs(ld)
	if err != nil {
		return s.logBuffer, err
	}

	if err := s.send(ctx, LogsPipeline, bytes.NewReader(body), flds); err != nil {
		return s.logBuffer, err
	}
	return nil, nil
}

// sendMetrics sends metrics in right format basing on the s.config.MetricFormat
func (s *sender) sendMetrics(ctx context.Context, flds fields) ([]metricPair, error) {

	// Follow different execution path for OTLP format
	if s.config.MetricFormat == OTLPMetricFormat {
		return s.sendOTLPMetrics(ctx, flds)
	}

	var (
		body           strings.Builder
		errs           []error
		droppedRecords []metricPair
		currentRecords []metricPair
	)

	for _, record := range s.metricBuffer {
		var formattedLine string
		var err error

		switch s.config.MetricFormat {
		case PrometheusFormat:
			formattedLine = s.prometheusFormatter.metric2String(record)
		case Carbon2Format:
			formattedLine = carbon2Metric2String(record)
		case GraphiteFormat:
			formattedLine = s.graphiteFormatter.metric2String(record)
		default:
			err = fmt.Errorf("unexpected metric format: %s", s.config.MetricFormat)
		}

		if err != nil {
			droppedRecords = append(droppedRecords, record)
			errs = append(errs, err)
			continue
		}

		ar, err := s.appendAndSend(ctx, formattedLine, MetricsPipeline, &body, flds)
		if err != nil {
			errs = append(errs, err)
			if ar.sent {
				droppedRecords = append(droppedRecords, currentRecords...)
			}

			if !ar.appended {
				droppedRecords = append(droppedRecords, record)
			}
		}

		// If data was sent, cleanup the currentTimeSeries counter
		if ar.sent {
			currentRecords = currentRecords[:0]
		}

		// If log has been appended to body, increment the currentTimeSeries
		if ar.appended {
			currentRecords = append(currentRecords, record)
		}
	}

	if body.Len() > 0 {
		if err := s.send(ctx, MetricsPipeline, strings.NewReader(body.String()), flds); err != nil {
			errs = append(errs, err)
			droppedRecords = append(droppedRecords, currentRecords...)
		}
	}

	if len(errs) > 0 {
		return droppedRecords, consumererror.Combine(errs)
	}
	return droppedRecords, nil
}

// sendMetrics sends metric records from the metricBuffer in OTLP format and as a result
// it returns an array of records which has not been sent correctly and an error.
// TODO: add support for HTTP limits
func (s *sender) sendOTLPMetrics(ctx context.Context, flds fields) ([]metricPair, error) {
	md := pdata.NewMetrics()
	rms := md.ResourceMetrics()
	rms.EnsureCapacity(len(s.metricBuffer))
	for _, record := range s.metricBuffer {
		rm := rms.AppendEmpty()
		record.attributes.CopyTo(rm.Resource().Attributes())
		ilm := rm.InstrumentationLibraryMetrics().AppendEmpty()
		ms := ilm.Metrics().AppendEmpty()
		record.metric.CopyTo(ms)
	}

	body, err := metricsMarshaler.MarshalMetrics(md)
	if err != nil {
		return s.metricBuffer, err
	}

	if err := s.send(ctx, MetricsPipeline, bytes.NewReader(body), flds); err != nil {
		return s.metricBuffer, err
	}
	return nil, nil
}

// appendAndSend appends line to the request body that will be sent and sends
// the accumulated data if the internal logBuffer has been filled (with maxBufferSize elements).
// It returns appendResponse
func (s *sender) appendAndSend(
	ctx context.Context,
	line string,
	pipeline PipelineType,
	body *strings.Builder,
	flds fields,
) (appendResponse, error) {
	var errors []error
	ar := newAppendResponse()

	if body.Len() > 0 && body.Len()+len(line) >= s.config.MaxRequestBodySize {
		ar.sent = true
		if err := s.send(ctx, pipeline, strings.NewReader(body.String()), flds); err != nil {
			errors = append(errors, err)
		}
		body.Reset()
	}

	if body.Len() > 0 {
		// Do not add newline if the body is empty
		if _, err := body.WriteString("\n"); err != nil {
			errors = append(errors, err)
			ar.appended = false
		}
	}

	if ar.appended {
		// Do not append new line if separator was not appended
		if _, err := body.WriteString(line); err != nil {
			errors = append(errors, err)
			ar.appended = false
		}
	}

	if len(errors) > 0 {
		return ar, consumererror.Combine(errors)
	}
	return ar, nil
}

// sendTraces sends traces in right format basing on the s.config.TraceFormat
func (s *sender) sendTraces(ctx context.Context, td pdata.Traces, flds fields) error {
	if s.config.TraceFormat == OTLPTraceFormat {
		return s.sendOTLPTraces(ctx, td, flds)
	}
	return nil
}

// sendOTLPTraces sends trace records in OTLP format
func (s *sender) sendOTLPTraces(ctx context.Context, td pdata.Traces, flds fields) error {
	body, err := tracesMarshaler.MarshalTraces(td)
	if err != nil {
		return err
	}
	if err := s.send(ctx, TracesPipeline, bytes.NewReader(body), flds); err != nil {
		return err
	}
	return nil
}

// cleanLogsBuffer zeroes logBuffer
func (s *sender) cleanLogsBuffer() {
	s.logBuffer = (s.logBuffer)[:0]
}

// batchLog adds log to the logBuffer and flushes them if logBuffer is full to avoid overflow
// returns list of log records which were not sent successfully
func (s *sender) batchLog(ctx context.Context, log pdata.LogRecord, metadata fields) ([]pdata.LogRecord, error) {
	s.logBuffer = append(s.logBuffer, log)

	if s.countLogs() >= maxBufferSize {
		dropped, err := s.sendLogs(ctx, metadata)
		s.cleanLogsBuffer()
		return dropped, err
	}

	return nil, nil
}

// countLogs returns number of logs in logBuffer
func (s *sender) countLogs() int {
	return len(s.logBuffer)
}

// cleanMetricBuffer zeroes metricBuffer
func (s *sender) cleanMetricBuffer() {
	s.metricBuffer = (s.metricBuffer)[:0]
}

// batchMetric adds metric to the metricBuffer and flushes them if metricBuffer is full to avoid overflow
// returns list of metric records which were not sent successfully
func (s *sender) batchMetric(ctx context.Context, metric metricPair, metadata fields) ([]metricPair, error) {
	s.metricBuffer = append(s.metricBuffer, metric)

	if s.countMetrics() >= maxBufferSize {
		dropped, err := s.sendMetrics(ctx, metadata)
		s.cleanMetricBuffer()
		return dropped, err
	}

	return nil, nil
}

// countMetrics returns number of metrics in metricBuffer
func (s *sender) countMetrics() int {
	return len(s.metricBuffer)
}

func addCompressHeader(req *http.Request, enc CompressEncodingType) error {
	switch enc {
	case GZIPCompression:
		req.Header.Set(headerContentEncoding, contentEncodingGzip)
	case DeflateCompression:
		req.Header.Set(headerContentEncoding, contentEncodingDeflate)
	case NoCompression:
	default:
		return fmt.Errorf("invalid content encoding: %s", enc)
	}

	return nil
}

func addSourcesHeaders(req *http.Request, sources sourceFormats, flds fields) {
	if sources.host.isSet() {
		req.Header.Add(headerHost, sources.host.format(flds))
	}

	if sources.name.isSet() {
		req.Header.Add(headerName, sources.name.format(flds))
	}

	if sources.category.isSet() {
		req.Header.Add(headerCategory, sources.category.format(flds))
	}
}

func addLogsHeaders(req *http.Request, lf LogFormatType, flds fields) {
	switch lf {
	case OTLPLogFormat:
		req.Header.Add(headerContentType, contentTypeOTLP)
	default:
		req.Header.Add(headerContentType, contentTypeLogs)
	}
	req.Header.Add(headerFields, flds.string())
}

func addMetricsHeaders(req *http.Request, mf MetricFormatType) error {
	switch mf {
	case PrometheusFormat:
		req.Header.Add(headerContentType, contentTypePrometheus)
	case Carbon2Format:
		req.Header.Add(headerContentType, contentTypeCarbon2)
	case GraphiteFormat:
		req.Header.Add(headerContentType, contentTypeGraphite)
	case OTLPMetricFormat:
		req.Header.Add(headerContentType, contentTypeOTLP)
	default:
		return fmt.Errorf("unsupported metrics format: %s", mf)
	}
	return nil
}

func addTracesHeaders(req *http.Request, tf TraceFormatType) error {
	switch tf {
	case OTLPTraceFormat:
		req.Header.Add(headerContentType, contentTypeOTLP)
	default:
		return fmt.Errorf("unsupported traces format: %s", tf)
	}
	return nil
}

func (s *sender) addRequestHeaders(req *http.Request, pipeline PipelineType, flds fields) error {
	req.Header.Add(headerClient, s.config.Client)

	if err := addCompressHeader(req, s.config.CompressEncoding); err != nil {
		return err
	}
	addSourcesHeaders(req, s.sources, flds)

	switch pipeline {
	case LogsPipeline:
		addLogsHeaders(req, s.config.LogFormat, flds)
	case MetricsPipeline:
		if err := addMetricsHeaders(req, s.config.MetricFormat); err != nil {
			return err
		}
	case TracesPipeline:
		if err := addTracesHeaders(req, s.config.TraceFormat); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unexpected pipeline: %v", pipeline)
	}
	return nil
}
