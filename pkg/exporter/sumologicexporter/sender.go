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
	"time"

	"go.opentelemetry.io/collector/model/otlp"
	"go.opentelemetry.io/collector/model/pdata"
	"go.uber.org/multierr"
	"go.uber.org/zap"

	"github.com/SumoLogic/sumologic-otel-collector/pkg/exporter/sumologicexporter/internal/observability"
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

// logPair keeps information about logRecord and attributes,
// where attributes are record and resource attributes
type logPair struct {
	attributes pdata.AttributeMap
	log        pdata.LogRecord
}

// countingReader keeps number of records related to reader
type countingReader struct {
	counter int64
	reader  io.Reader
}

// newCountingReader creates countingReader with given number of records
func newCountingReader(records int) *countingReader {
	return &countingReader{
		counter: int64(records),
	}
}

// withBytes sets up reader to read from bytes data
func (c *countingReader) withBytes(data []byte) *countingReader {
	c.reader = bytes.NewReader(data)
	return c
}

// withString sets up reader to read from string data
func (c *countingReader) withString(data string) *countingReader {
	c.reader = strings.NewReader(data)
	return c
}

// bodyBuilder keeps information about number of records related to data it keeps
type bodyBuilder struct {
	builder strings.Builder
	counter int
}

// newBodyBuilder returns empty bodyBuilder
func newBodyBuilder() bodyBuilder {
	return bodyBuilder{}
}

// Reset resets both counter and builder content
func (b *bodyBuilder) Reset() {
	b.counter = 0
	b.builder.Reset()
}

// addLine adds line to builder and increments counter
func (b *bodyBuilder) addLine(line string) error {
	_, err := b.builder.WriteString(line)
	if err != nil {
		return err
	}
	b.counter += 1
	return nil
}

// addNewLine adds newline to builder
func (b *bodyBuilder) addNewLine() error {
	err := b.builder.WriteByte('\n')
	return err
}

// Len returns builder content length
func (b *bodyBuilder) Len() int {
	return b.builder.Len()
}

// toCountingReader converts bodyBuilder to countingReader
func (b *bodyBuilder) toCountingReader() *countingReader {
	return newCountingReader(b.counter).withString(b.builder.String())
}

type sender struct {
	logger              *zap.Logger
	logBuffer           []logPair
	metricBuffer        []metricPair
	config              *Config
	client              *http.Client
	filter              filter
	sources             sourceFormats
	compressor          compressor
	prometheusFormatter prometheusFormatter
	graphiteFormatter   graphiteFormatter
	jsonLogsConfig      JSONLogs
	dataUrlMetrics      string
	dataUrlLogs         string
	dataUrlTraces       string
}

const (
	// maxBufferSize defines size of the logBuffer (maximum number of pdata.LogRecord entries)
	maxBufferSize int = 1024 * 1024

	headerContentType     string = "Content-Type"
	headerContentEncoding string = "Content-Encoding"
	headerClient          string = "X-Sumo-Client"
	headerHost            string = "X-Sumo-Host"
	headerName            string = "X-Sumo-Name"
	headerCategory        string = "X-Sumo-Category"
	headerFields          string = "X-Sumo-Fields"

	attributeKeySourceHost     = "_sourceHost"
	attributeKeySourceName     = "_sourceName"
	attributeKeySourceCategory = "_sourceCategory"

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
	logger *zap.Logger,
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
		logger:              logger,
		config:              cfg,
		client:              cl,
		filter:              f,
		sources:             s,
		compressor:          c,
		prometheusFormatter: pf,
		graphiteFormatter:   gf,
		jsonLogsConfig:      cfg.JSONLogs,
		dataUrlMetrics:      metricsUrl,
		dataUrlLogs:         logsUrl,
		dataUrlTraces:       tracesUrl,
	}
}

var errUnauthorized = errors.New("unauthorized")

// send sends data to sumologic
func (s *sender) send(ctx context.Context, pipeline PipelineType, reader *countingReader, flds fields) error {
	data, err := s.compressor.compress(reader.reader)
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

	s.logger.Debug("Sending data",
		zap.String("pipeline", string(pipeline)),
		zap.Any("headers", req.Header),
	)

	start := time.Now()
	resp, err := s.client.Do(req)
	if err != nil {
		s.recordMetrics(time.Since(start), reader.counter, req, nil, pipeline)
		return err
	}
	defer resp.Body.Close()

	s.recordMetrics(time.Since(start), reader.counter, req, resp, pipeline)

	return s.handleReceiverResponse(resp)
}

func (s *sender) handleReceiverResponse(resp *http.Response) error {
	// API responds with a 200 or 204 with ConentLength set to 0 when all data
	// has been successfully ingested.
	if resp.ContentLength == 0 && (resp.StatusCode == 200 || resp.StatusCode == 204) {
		return nil
	}

	type ReceiverResponseCore struct {
		Status  int    `json:"status,omitempty"`
		ID      string `json:"id,omitempty"`
		Code    string `json:"code,omitempty"`
		Message string `json:"message,omitempty"`
	}

	// API responds with a 200 or 204 with a JSON body describing what issues
	// were encountered when processing the sent data.
	switch resp.StatusCode {
	case 200, 204:
		var rResponse ReceiverResponseCore
		var (
			b  = bytes.NewBuffer(make([]byte, 0, resp.ContentLength))
			tr = io.TeeReader(resp.Body, b)
		)

		if err := json.NewDecoder(tr).Decode(&rResponse); err != nil {
			s.logger.Warn("Error decoding receiver response", zap.ByteString("body", b.Bytes()))
			return nil
		}

		l := s.logger.With(zap.String("status", resp.Status))
		if len(rResponse.ID) > 0 {
			l = l.With(zap.String("id", rResponse.ID))
		}
		if len(rResponse.Code) > 0 {
			l = l.With(zap.String("code", rResponse.Code))
		}
		if len(rResponse.Message) > 0 {
			l = l.With(zap.String("message", rResponse.Message))
		}
		l.Warn("There was an issue sending data")
		return nil

	case 401:
		return errUnauthorized

	default:
		type ReceiverErrorResponse struct {
			ReceiverResponseCore
			Errors []struct {
				Code    string `json:"code"`
				Message string `json:"message"`
			} `json:"errors,omitempty"`
		}

		var rResponse ReceiverErrorResponse
		if resp.ContentLength > 0 {
			var (
				b  = bytes.NewBuffer(make([]byte, 0, resp.ContentLength))
				tr = io.TeeReader(resp.Body, b)
			)

			if err := json.NewDecoder(tr).Decode(&rResponse); err != nil {
				return fmt.Errorf("failed to decode API response (status: %s): %s",
					resp.Status, b.String(),
				)
			}
		}

		errMsgs := []string{
			fmt.Sprintf("status: %s", resp.Status),
		}

		if len(rResponse.ID) > 0 {
			errMsgs = append(errMsgs, fmt.Sprintf("id: %s", rResponse.ID))
		}
		if len(rResponse.Code) > 0 {
			errMsgs = append(errMsgs, fmt.Sprintf("code: %s", rResponse.Code))
		}
		if len(rResponse.Message) > 0 {
			errMsgs = append(errMsgs, fmt.Sprintf("message: %s", rResponse.Message))
		}
		if len(rResponse.Errors) > 0 {
			errMsgs = append(errMsgs, fmt.Sprintf("errors: %+v", rResponse.Errors))
		}

		return fmt.Errorf("failed sending data: %s", strings.Join(errMsgs, ", "))
	}
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
	return record.Body().AsString()
}

// logToJSON converts LogRecord to a json line, returns it and error eventually
func (s *sender) logToJSON(record logPair) (string, error) {
	data := s.filter.filterOut(record.attributes)
	if s.jsonLogsConfig.AddTimestamp {
		addJSONTimestamp(data.orig, s.jsonLogsConfig.TimestampKey, record.log.Timestamp())
	}

	if s.config.TranslateAttributes {
		data.translateAttributes()
	}

	// Only append the body when it's not empty to prevent sending 'null' log.
	if body := record.log.Body(); !isEmptyAttributeValue(body) {
		if s.jsonLogsConfig.FlattenBody && body.Type() == pdata.AttributeValueTypeMap {
			// Cannot use CopyTo, as it overrides data.orig's values
			body.MapVal().Range(func(k string, v pdata.AttributeValue) bool {
				data.orig.Insert(k, v)
				return true
			})
		} else {
			data.orig.Upsert(s.jsonLogsConfig.LogKey, body)
		}
	}

	nextLine, err := json.Marshal(data.orig.AsRaw())
	if err != nil {
		return "", err
	}

	return bytes.NewBuffer(nextLine).String(), nil
}

var timeZeroUTC = time.Unix(0, 0).UTC()

// addJSONTimestamp adds a timestamp field to record attribtues before sending
// out the logs as JSON.
// If the attached timestamp is equal to 0 (millisecond based UNIX timestamp)
// then send out current time formatted as milliseconds since January 1, 1970.
func addJSONTimestamp(attrs pdata.AttributeMap, timestampKey string, pt pdata.Timestamp) {
	t := pt.AsTime()
	if t == timeZeroUTC {
		attrs.InsertInt(timestampKey, time.Now().UnixMilli())
	} else {
		attrs.InsertInt(timestampKey, t.UnixMilli())
	}
}

func isEmptyAttributeValue(att pdata.AttributeValue) bool {
	t := att.Type()
	return !(t == pdata.AttributeValueTypeString && len(att.StringVal()) > 0 ||
		t == pdata.AttributeValueTypeArray && att.SliceVal().Len() > 0 ||
		t == pdata.AttributeValueTypeMap && att.MapVal().Len() > 0 ||
		t == pdata.AttributeValueTypeBytes && len(att.BytesVal()) > 0)
}

// sendNonOTLPLogs sends log records from the logBuffer formatted according
// to configured LogFormat and as the result of execution
// returns array of records which has not been sent correctly and error
func (s *sender) sendNonOTLPLogs(ctx context.Context, flds fields) ([]logPair, error) {
	if s.config.LogFormat == OTLPLogFormat {
		return nil, fmt.Errorf("Attempting to send OTLP logs as non-OTLP data")
	}

	var (
		body           bodyBuilder = newBodyBuilder()
		errs           []error
		droppedRecords []logPair
		currentRecords []logPair
	)

	for _, record := range s.logBuffer {
		var formattedLine string
		var err error

		switch s.config.LogFormat {
		case TextFormat:
			formattedLine = s.logToText(record.log)
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
		if err := s.send(ctx, LogsPipeline, body.toCountingReader(), flds); err != nil {
			errs = append(errs, err)
			droppedRecords = append(droppedRecords, currentRecords...)
		}
	}

	if len(errs) > 0 {
		return droppedRecords, multierr.Combine(errs...)
	}
	return droppedRecords, nil
}

// TODO: add support for HTTP limits
func (s *sender) sendOTLPLogs(ctx context.Context, ld pdata.Logs) error {
	rls := ld.ResourceLogs()
	for i := 0; i < rls.Len(); i++ {
		rl := rls.At(i)

		if s.config.TranslateAttributes {
			translateAttributes(rl.Resource().Attributes()).
				CopyTo(rl.Resource().Attributes())
		}

		// Clear timestamps if required
		if s.config.ClearLogsTimestamp {
			slgs := rl.ScopeLogs()
			for j := 0; j < slgs.Len(); j++ {
				log := slgs.At(j)
				for k := 0; k < log.LogRecords().Len(); k++ {
					log.LogRecords().At(k).SetTimestamp(0)
				}
			}
		}

		s.addSourceResourceAttributes(rl.Resource().Attributes())
	}

	body, err := logsMarshaler.MarshalLogs(ld)
	if err != nil {
		return err
	}

	return s.send(ctx, LogsPipeline, newCountingReader(ld.LogRecordCount()).withBytes(body), fields{})
}

// sendNonOTLPMetrics sends metrics in right format basing on the s.config.MetricFormat
func (s *sender) sendNonOTLPMetrics(ctx context.Context, flds fields) ([]metricPair, error) {
	if s.config.MetricFormat == OTLPMetricFormat {
		return nil, fmt.Errorf("Attempting to send OTLP metrics as non-OTLP data")
	}

	var (
		body           bodyBuilder
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
		if err := s.send(ctx, MetricsPipeline, body.toCountingReader(), flds); err != nil {
			errs = append(errs, err)
			droppedRecords = append(droppedRecords, currentRecords...)
		}
	}

	if len(errs) > 0 {
		return droppedRecords, multierr.Combine(errs...)
	}
	return droppedRecords, nil
}

func (s *sender) sendOTLPMetrics(ctx context.Context, md pdata.Metrics) error {
	rms := md.ResourceMetrics()
	for i := 0; i < rms.Len(); i++ {
		rm := rms.At(i)

		if s.config.TranslateAttributes {
			translateAttributes(rm.Resource().Attributes()).
				CopyTo(rm.Resource().Attributes())
		}

		s.addSourceResourceAttributes(rm.Resource().Attributes())
	}

	body, err := metricsMarshaler.MarshalMetrics(md)
	if err != nil {
		return err
	}

	return s.send(ctx, MetricsPipeline, newCountingReader(md.DataPointCount()).withBytes(body), fields{})
}

// appendAndSend appends line to the request body that will be sent and sends
// the accumulated data if the internal logBuffer has been filled (with maxBufferSize elements).
// It returns appendResponse
func (s *sender) appendAndSend(
	ctx context.Context,
	line string,
	pipeline PipelineType,
	body *bodyBuilder,
	flds fields,
) (appendResponse, error) {
	var errors []error
	ar := newAppendResponse()

	if body.Len() > 0 && body.Len()+len(line) >= s.config.MaxRequestBodySize {
		ar.sent = true
		if err := s.send(ctx, pipeline, body.toCountingReader(), flds); err != nil {
			errors = append(errors, err)
		}
		body.Reset()
	}

	if body.Len() > 0 {
		// Do not add newline if the body is empty
		if err := body.addNewLine(); err != nil {
			errors = append(errors, err)
			ar.appended = false
		}
	}

	if ar.appended {
		// Do not append new line if separator was not appended
		if err := body.addLine(line); err != nil {
			errors = append(errors, err)
			ar.appended = false
		}
	}

	if len(errors) > 0 {
		return ar, multierr.Combine(errors...)
	}
	return ar, nil
}

// sendTraces sends traces in right format basing on the s.config.TraceFormat
func (s *sender) sendTraces(ctx context.Context, td pdata.Traces) error {
	if s.config.TraceFormat == OTLPTraceFormat {
		return s.sendOTLPTraces(ctx, td)
	}
	return nil
}

// sendOTLPTraces sends trace records in OTLP format
func (s *sender) sendOTLPTraces(ctx context.Context, td pdata.Traces) error {
	capacity := td.SpanCount()
	for i := 0; i < td.ResourceSpans().Len(); i++ {
		s.addSourceResourceAttributes(td.ResourceSpans().At(i).Resource().Attributes())
	}

	body, err := tracesMarshaler.MarshalTraces(td)
	if err != nil {
		return err
	}
	if err := s.send(ctx, TracesPipeline, newCountingReader(capacity).withBytes(body), fields{}); err != nil {
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
func (s *sender) batchLog(ctx context.Context, log logPair, metadata fields) ([]logPair, error) {
	s.logBuffer = append(s.logBuffer, log)

	if s.countLogs() >= maxBufferSize {
		dropped, err := s.sendNonOTLPLogs(ctx, metadata)
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
		dropped, err := s.sendNonOTLPMetrics(ctx, metadata)
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
	if !flds.isInitialized() {
		return
	}

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

	if fieldsStr := flds.string(); fieldsStr != "" {
		req.Header.Add(headerFields, fieldsStr)
	}
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

// addSourceResourceAttributes adds source related attributes:
// * source category
// * source host
// * source name
// to the provided attribute map using the provided fields as values source and using
// the source templates for formatting.
func (s *sender) addSourceRelatedResourceAttributesFromFields(attrs pdata.AttributeMap, flds fields) {
	if s.sources.host.isSet() {
		attrs.InsertString(attributeKeySourceHost, s.sources.host.format(flds))
	}
	if s.sources.name.isSet() {
		attrs.InsertString(attributeKeySourceName, s.sources.name.format(flds))
	}
	if s.sources.category.isSet() {
		attrs.InsertString(attributeKeySourceCategory, s.sources.category.format(flds))
	}
}

// addSourceResourceAttributes adds source related attributes:
// * source category
// * source host
// * source name
// to the provided attribute map, according to the corresponding templates.
//
// When those attributes are already in the attribute map then nothing is
// changed since attributes that are provided with data have precedence over
// exporter configuration.
func (s *sender) addSourceResourceAttributes(attrs pdata.AttributeMap) {
	if s.sources.host.isSet() {
		if _, ok := attrs.Get(attributeKeySourceHost); !ok {
			attrs.InsertString(attributeKeySourceHost, s.sources.host.formatPdataMap(attrs))
		}
	}
	if s.sources.name.isSet() {
		if _, ok := attrs.Get(attributeKeySourceName); !ok {
			attrs.InsertString(attributeKeySourceName, s.sources.name.formatPdataMap(attrs))
		}
	}
	if s.sources.category.isSet() {
		if _, ok := attrs.Get(attributeKeySourceCategory); !ok {
			attrs.InsertString(attributeKeySourceCategory, s.sources.category.formatPdataMap(attrs))
		}
	}
}

func (s *sender) recordMetrics(duration time.Duration, count int64, req *http.Request, resp *http.Response, pipeline PipelineType) {
	statusCode := 0

	if resp != nil {
		statusCode = resp.StatusCode
	}

	id := s.config.ID().String()

	if err := observability.RecordRequestsDuration(duration, statusCode, req.URL.String(), string(pipeline), id); err != nil {
		s.logger.Debug("error for recording metric for request duration", zap.Error(err))
	}

	if err := observability.RecordRequestsBytes(req.ContentLength, statusCode, req.URL.String(), string(pipeline), id); err != nil {
		s.logger.Debug("error for recording metric for sent bytes", zap.Error(err))
	}

	if err := observability.RecordRequestsRecords(count, statusCode, req.URL.String(), string(pipeline), id); err != nil {
		s.logger.Debug("error for recording metric for sent records", zap.Error(err))
	}

	if err := observability.RecordRequestsSent(statusCode, req.URL.String(), string(pipeline), id); err != nil {
		s.logger.Debug("error for recording metric for sent request", zap.Error(err))
	}
}
