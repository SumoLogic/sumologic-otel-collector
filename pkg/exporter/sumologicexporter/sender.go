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
	"reflect"
	"strings"
	"time"

	"go.opentelemetry.io/collector/model/otlp"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.uber.org/multierr"
	"go.uber.org/zap"

	"github.com/SumoLogic/sumologic-otel-collector/pkg/exporter/sumologicexporter/internal/observability"
)

var (
	tracesMarshaler  = otlp.NewProtobufTracesMarshaler()
	metricsMarshaler = otlp.NewProtobufMetricsMarshaler()
	logsMarshaler    = otlp.NewProtobufLogsMarshaler()
)

// metricPair represents information required to send one metric to the Sumo Logic
type metricPair struct {
	attributes pcommon.Map
	metric     pmetric.Metric
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
func (b *bodyBuilder) addLine(line string) {
	b.builder.WriteString(line) // WriteString can't actually return an error
	b.counter += 1
}

// addLine adds multiple lines to builder and increments counter
func (b *bodyBuilder) addLines(lines []string) {
	if len(lines) == 0 {
		return
	}

	// add the first line separately to avoid a conditional in the loop
	b.builder.WriteString(lines[0])

	for _, line := range lines[1:] {
		b.builder.WriteByte('\n')
		b.builder.WriteString(line) // WriteString can't actually return an error
	}
	b.counter += len(lines)
}

// addNewLine adds newline to builder
func (b *bodyBuilder) addNewLine() {
	b.builder.WriteByte('\n') // WriteByte can't actually return an error
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
	metricBuffer        []metricPair
	config              *Config
	client              *http.Client
	sources             sourceFormats
	compressor          compressor
	prometheusFormatter prometheusFormatter
	jsonLogsConfig      JSONLogs
	dataUrlMetrics      string
	dataUrlLogs         string
	dataUrlTraces       string
}

const (
	// maxBufferSize defines size of the logBuffer (maximum number of plog.LogRecord entries)
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
	contentTypeOTLP       string = "application/x-protobuf"

	contentEncodingGzip    string = "gzip"
	contentEncodingDeflate string = "deflate"
)

func newSender(
	logger *zap.Logger,
	cfg *Config,
	cl *http.Client,
	s sourceFormats,
	c compressor,
	pf prometheusFormatter,
	metricsUrl string,
	logsUrl string,
	tracesUrl string,
) *sender {
	return &sender{
		logger:              logger,
		config:              cfg,
		client:              cl,
		sources:             s,
		compressor:          c,
		prometheusFormatter: pf,
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
func (s *sender) logToText(record plog.LogRecord) string {
	return record.Body().AsString()
}

// logToJSON converts LogRecord to a json line, returns it and error eventually
func (s *sender) logToJSON(record plog.LogRecord) (string, error) {
	if s.jsonLogsConfig.AddTimestamp {
		addJSONTimestamp(record.Attributes(), s.jsonLogsConfig.TimestampKey, record.Timestamp())
	}

	// Only append the body when it's not empty to prevent sending 'null' log.
	if body := record.Body(); !isEmptyAttributeValue(body) {
		if s.jsonLogsConfig.FlattenBody && body.Type() == pcommon.ValueTypeMap {
			// Cannot use CopyTo, as it overrides data.orig's values
			body.MapVal().Range(func(k string, v pcommon.Value) bool {
				record.Attributes().Insert(k, v)
				return true
			})
		} else {
			record.Attributes().Upsert(s.jsonLogsConfig.LogKey, body)
		}
	}

	nextLine, err := json.Marshal(record.Attributes().AsRaw())
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
func addJSONTimestamp(attrs pcommon.Map, timestampKey string, pt pcommon.Timestamp) {
	t := pt.AsTime()
	if t == timeZeroUTC {
		attrs.InsertInt(timestampKey, time.Now().UnixMilli())
	} else {
		attrs.InsertInt(timestampKey, t.UnixMilli())
	}
}

func isEmptyAttributeValue(att pcommon.Value) bool {
	t := att.Type()
	return !(t == pcommon.ValueTypeString && len(att.StringVal()) > 0 ||
		t == pcommon.ValueTypeSlice && att.SliceVal().Len() > 0 ||
		t == pcommon.ValueTypeMap && att.MapVal().Len() > 0 ||
		t == pcommon.ValueTypeBytes && len(att.MBytesVal()) > 0)
}

// sendNonOTLPLogs sends log records from the logBuffer formatted according
// to configured LogFormat and as the result of execution
// returns array of records which has not been sent correctly and error
func (s *sender) sendNonOTLPLogs(ctx context.Context, rl plog.ResourceLogs, flds fields) ([]plog.LogRecord, error) {
	if s.config.LogFormat == OTLPLogFormat {
		return nil, fmt.Errorf("Attempting to send OTLP logs as non-OTLP data")
	}

	var (
		body           bodyBuilder = newBodyBuilder()
		errs           []error
		droppedRecords []plog.LogRecord
		currentRecords []plog.LogRecord
	)

	slgs := rl.ScopeLogs()
	for i := 0; i < slgs.Len(); i++ {
		slg := slgs.At(i)
		for j := 0; j < slg.LogRecords().Len(); j++ {
			lr := slg.LogRecords().At(j)
			formattedLine, err := s.formatLogLine(lr)
			if err != nil {
				droppedRecords = append(droppedRecords, lr)
				errs = append(errs, err)
				continue
			}

			sent, err := s.appendAndMaybeSend(ctx, []string{formattedLine}, LogsPipeline, &body, flds)
			if err != nil {
				errs = append(errs, err)
				droppedRecords = append(droppedRecords, currentRecords...)
			}

			// If data was sent and either failed or succeeded, cleanup the currentRecords slice
			if sent {
				currentRecords = currentRecords[:0]
			}

			currentRecords = append(currentRecords, lr)
		}
	}

	if body.Len() > 0 {
		if err := s.send(ctx, LogsPipeline, body.toCountingReader(), flds); err != nil {
			errs = append(errs, err)
			droppedRecords = append(droppedRecords, currentRecords...)
		}
	}

	return droppedRecords, multierr.Combine(errs...)
}

func (s *sender) formatLogLine(lr plog.LogRecord) (string, error) {
	var formattedLine string
	var err error

	switch s.config.LogFormat {
	case TextFormat:
		formattedLine = s.logToText(lr)
	case JSONFormat:
		formattedLine, err = s.logToJSON(lr)
	default:
		err = errors.New("unexpected log format")
	}

	return formattedLine, err
}

// TODO: add support for HTTP limits
func (s *sender) sendOTLPLogs(ctx context.Context, ld plog.Logs) error {
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
func (s *sender) sendNonOTLPMetrics(ctx context.Context, md pmetric.Metrics) (pmetric.Metrics, []error) {
	if s.config.MetricFormat == OTLPMetricFormat {
		return md, []error{fmt.Errorf("Attempting to send OTLP metrics as non-OTLP data")}
	}

	var (
		body             bodyBuilder = newBodyBuilder()
		errs             []error
		currentResources []pmetric.ResourceMetrics
		flds             fields
	)

	rms := md.ResourceMetrics()
	droppedMetrics := pmetric.NewMetrics()
	for i := 0; i < rms.Len(); i++ {
		rm := rms.At(i)
		flds = newFields(rm.Resource().Attributes())
		sms := rm.ScopeMetrics()

		// generally speaking, it's fine to send multiple ResourceMetrics in a single request
		// the only exception is if the computed source headers are different, as those as unique per-request
		// so we check if the headers are different here and send what we have if they are
		if i > 0 {
			currentSourceHeaders := getSourcesHeaders(s.sources, flds)
			previousFields := newFields(rms.At(i - 1).Resource().Attributes())
			previousSourceHeaders := getSourcesHeaders(s.sources, previousFields)
			if !reflect.DeepEqual(previousSourceHeaders, currentSourceHeaders) && body.Len() > 0 {
				if err := s.send(ctx, MetricsPipeline, body.toCountingReader(), previousFields); err != nil {
					errs = append(errs, err)
					for _, resource := range currentResources {
						resource.MoveTo(droppedMetrics.ResourceMetrics().AppendEmpty())
					}
				}
				body.Reset()
				currentResources = currentResources[:0]
			}
		}

		// transform the metrics into formatted lines ready to be sent
		var formattedLines []string
		var err error
		for i := 0; i < sms.Len(); i++ {
			sm := sms.At(i)

			for j := 0; j < sm.Metrics().Len(); j++ {
				m := sm.Metrics().At(j)

				var formattedLine string

				switch s.config.MetricFormat {
				case PrometheusFormat:
					formattedLine = s.prometheusFormatter.metric2String(m, rm.Resource().Attributes())
				default:
					return md, []error{fmt.Errorf("unexpected metric format: %s", s.config.MetricFormat)}
				}

				formattedLines = append(formattedLines, formattedLine)
			}
		}

		sent, err := s.appendAndMaybeSend(ctx, formattedLines, MetricsPipeline, &body, flds)
		if err != nil {
			errs = append(errs, err)
			if sent {
				// failed at sending, add the resource to the dropped metrics
				// move instead of copy here to avoid duplicating data in memory on failure
				for _, resource := range currentResources {
					resource.MoveTo(droppedMetrics.ResourceMetrics().AppendEmpty())
				}
			}
		}

		// If data was sent, cleanup the currentResources slice
		if sent {
			currentResources = currentResources[:0]
		}

		currentResources = append(currentResources, rm)

	}

	if body.Len() > 0 {
		if err := s.send(ctx, MetricsPipeline, body.toCountingReader(), flds); err != nil {
			errs = append(errs, err)
			for _, resource := range currentResources {
				resource.MoveTo(droppedMetrics.ResourceMetrics().AppendEmpty())
			}
		}
	}

	return droppedMetrics, errs
}

func (s *sender) sendOTLPMetrics(ctx context.Context, md pmetric.Metrics) error {
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

// appendAndMaybeSend appends line to the request body that will be sent and sends
// the accumulated data if the internal logBuffer has been filled (with config.MaxRequestBodySize bytes).
// It returns a boolean indicating if the data was sent and an error
func (s *sender) appendAndMaybeSend(
	ctx context.Context,
	lines []string,
	pipeline PipelineType,
	body *bodyBuilder,
	flds fields,
) (sent bool, err error) {

	linesTotalLength := 0
	for _, line := range lines {
		linesTotalLength += len(line) + 1 // count the newline as well
	}

	if body.Len() > 0 && body.Len()+linesTotalLength >= s.config.MaxRequestBodySize {
		sent = true
		err = s.send(ctx, pipeline, body.toCountingReader(), flds)
		body.Reset()
	}

	if body.Len() > 0 {
		// Do not add newline if the body is empty
		body.addNewLine()
	}

	body.addLines(lines)

	return sent, err
}

// sendTraces sends traces in right format basing on the s.config.TraceFormat
func (s *sender) sendTraces(ctx context.Context, td ptrace.Traces) error {
	if s.config.TraceFormat == OTLPTraceFormat {
		return s.sendOTLPTraces(ctx, td)
	}
	return nil
}

// sendOTLPTraces sends trace records in OTLP format
func (s *sender) sendOTLPTraces(ctx context.Context, td ptrace.Traces) error {
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

// cleanMetricBuffer zeroes metricBuffer
func (s *sender) cleanMetricBuffer() {
	s.metricBuffer = (s.metricBuffer)[:0]
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
	sourceHeaderValues := getSourcesHeaders(sources, flds)

	for headerName, headerValue := range sourceHeaderValues {
		req.Header.Add(headerName, headerValue)
	}
}

func getSourcesHeaders(sources sourceFormats, flds fields) map[string]string {
	sourceHeaderValues := map[string]string{}
	if !flds.isInitialized() {
		return sourceHeaderValues
	}

	if sources.host.isSet() {
		sourceHeaderValues[headerHost] = sources.host.format(flds)
	}

	if sources.name.isSet() {
		sourceHeaderValues[headerName] = sources.name.format(flds)
	}

	if sources.category.isSet() {
		sourceHeaderValues[headerCategory] = sources.category.format(flds)
	}
	return sourceHeaderValues
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
func (s *sender) addSourceRelatedResourceAttributesFromFields(attrs pcommon.Map, flds fields) {
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
func (s *sender) addSourceResourceAttributes(attrs pcommon.Map) {
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
