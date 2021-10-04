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
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/config/confighttp"
	"go.opentelemetry.io/collector/consumer/consumererror"
	"go.opentelemetry.io/collector/model/otlp"
	"go.opentelemetry.io/collector/model/pdata"
	"go.uber.org/zap"
)

func LogRecordsToLogs(records []pdata.LogRecord) pdata.Logs {
	logs := pdata.NewLogs()
	logsSlice := logs.ResourceLogs().AppendEmpty().InstrumentationLibraryLogs().AppendEmpty().Logs()
	for _, record := range records {
		record.CopyTo(logsSlice.AppendEmpty())
	}

	return logs
}

func logRecordsToLogPair(records []pdata.LogRecord) []logPair {
	logs := make([]logPair, len(records))
	for num, record := range records {
		logs[num] = logPair{
			log:        record,
			attributes: record.Attributes(),
		}
	}

	return logs
}

type exporterTest struct {
	srv *httptest.Server
	exp *sumologicexporter
}

func createTestConfig() *Config {
	config := createDefaultConfig().(*Config)
	config.CompressEncoding = NoCompression
	config.LogFormat = TextFormat
	config.MaxRequestBodySize = 20_971_520
	config.MetricFormat = Carbon2Format
	config.TraceFormat = OTLPTraceFormat
	return config
}

func createExporterCreateSettings() component.ExporterCreateSettings {
	return component.ExporterCreateSettings{
		TelemetrySettings: component.TelemetrySettings{
			Logger: zap.NewNop(),
		},
	}
}

// prepareExporterTest prepares an exporter test object using provided config
// and a slice of callbacks to be called for subsequent requests coming being
// sent to the server.
// The enclosed *httptest.Server is automatically closed on test cleanup.
func prepareExporterTest(t *testing.T, cfg *Config, cb []func(w http.ResponseWriter, req *http.Request), cfgOpts ...func(*Config)) *exporterTest {
	var reqCounter int32
	// generate a test server so we can capture and inspect the request
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if len(cb) == 0 {
			return
		}

		if c := int(atomic.LoadInt32(&reqCounter)); assert.Greater(t, len(cb), c) {
			cb[c](w, req)
			atomic.AddInt32(&reqCounter, 1)
		}
	}))
	t.Cleanup(func() { testServer.Close() })

	cfg.HTTPClientSettings.Endpoint = testServer.URL
	cfg.HTTPClientSettings.Auth = nil
	for _, cfgOpt := range cfgOpts {
		cfgOpt(cfg)
	}

	exp, err := initExporter(cfg, createExporterCreateSettings())
	require.NoError(t, err)

	require.NoError(t, exp.start(context.Background(), componenttest.NewNopHost()))

	return &exporterTest{
		srv: testServer,
		exp: exp,
	}
}

func TestInitExporter(t *testing.T) {
	_, err := initExporter(&Config{
		LogFormat:        "json",
		MetricFormat:     "carbon2",
		CompressEncoding: "gzip",
		TraceFormat:      "otlp",
		HTTPClientSettings: confighttp.HTTPClientSettings{
			Timeout:  defaultTimeout,
			Endpoint: "test_endpoint",
		},
	}, createExporterCreateSettings())
	assert.NoError(t, err)
}

func TestInitExporterInvalidLogFormat(t *testing.T) {
	_, err := initExporter(&Config{
		LogFormat:        "test_format",
		MetricFormat:     "carbon2",
		CompressEncoding: "gzip",
		TraceFormat:      "otlp",
		HTTPClientSettings: confighttp.HTTPClientSettings{
			Timeout:  defaultTimeout,
			Endpoint: "test_endpoint",
		},
	}, createExporterCreateSettings())

	assert.EqualError(t, err, "unexpected log format: test_format")
}

func TestInitExporterInvalidMetricFormat(t *testing.T) {
	_, err := initExporter(&Config{
		LogFormat:    "json",
		MetricFormat: "test_format",
		HTTPClientSettings: confighttp.HTTPClientSettings{
			Timeout:  defaultTimeout,
			Endpoint: "test_endpoint",
		},
		CompressEncoding: "gzip",
	}, createExporterCreateSettings())

	assert.EqualError(t, err, "unexpected metric format: test_format")
}

func TestInitExporterInvalidTraceFormat(t *testing.T) {
	_, err := initExporter(&Config{
		LogFormat:    "json",
		MetricFormat: "carbon2",
		TraceFormat:  "text",
		HTTPClientSettings: confighttp.HTTPClientSettings{
			Timeout:  defaultTimeout,
			Endpoint: "test_endpoint",
		},
		CompressEncoding: "gzip",
	}, createExporterCreateSettings())

	assert.EqualError(t, err, "unexpected trace format: text")
}

func TestInitExporterInvalidCompressEncoding(t *testing.T) {
	_, err := initExporter(&Config{
		LogFormat:        "json",
		MetricFormat:     "carbon2",
		CompressEncoding: "test_format",
		TraceFormat:      "otlp",
		HTTPClientSettings: confighttp.HTTPClientSettings{
			Timeout:  defaultTimeout,
			Endpoint: "test_endpoint",
		},
	}, createExporterCreateSettings())

	assert.EqualError(t, err, "unexpected compression encoding: test_format")
}

func TestInitExporterInvalidEndpointAndNoAuth(t *testing.T) {
	_, err := initExporter(&Config{
		LogFormat:        "json",
		MetricFormat:     "carbon2",
		CompressEncoding: "gzip",
		TraceFormat:      "otlp",
		HTTPClientSettings: confighttp.HTTPClientSettings{
			Timeout: defaultTimeout,
		},
	}, createExporterCreateSettings())

	assert.EqualError(t, err, "no endpoint and no auth extension specified")
}

func TestAllSuccess(t *testing.T) {
	test := prepareExporterTest(t, createTestConfig(), []func(w http.ResponseWriter, req *http.Request){
		func(w http.ResponseWriter, req *http.Request) {
			body := extractBody(t, req)
			assert.Equal(t, `Example log`, body)
			assert.Equal(t, "", req.Header.Get("X-Sumo-Fields"))
		},
	})

	logs := LogRecordsToLogs(exampleLog())

	err := test.exp.pushLogsData(context.Background(), logs)
	assert.NoError(t, err)
}

func TestResourceMerge(t *testing.T) {
	test := prepareExporterTest(t, createTestConfig(), []func(w http.ResponseWriter, req *http.Request){
		func(w http.ResponseWriter, req *http.Request) {
			body := extractBody(t, req)
			assert.Equal(t, `Example log`, body)
			assert.Equal(t, "key1=original_value, key2=additional_value", req.Header.Get("X-Sumo-Fields"))
		},
	})

	f, err := newFilter([]string{`key\d`})
	require.NoError(t, err)
	test.exp.filter = f

	logs := LogRecordsToLogs(exampleLog())
	logs.ResourceLogs().At(0).InstrumentationLibraryLogs().At(0).Logs().At(0).Attributes().InsertString("key1", "original_value")
	logs.ResourceLogs().At(0).Resource().Attributes().InsertString("key1", "overwrite_value")
	logs.ResourceLogs().At(0).Resource().Attributes().InsertString("key2", "additional_value")

	err = test.exp.pushLogsData(context.Background(), logs)
	assert.NoError(t, err)
}

func TestAllFailed(t *testing.T) {
	test := prepareExporterTest(t, createTestConfig(), []func(w http.ResponseWriter, req *http.Request){
		func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(500)

			body := extractBody(t, req)
			assert.Equal(t, "Example log\nAnother example log", body)
			assert.Equal(t, "", req.Header.Get("X-Sumo-Fields"))
		},
	})

	logs := LogRecordsToLogs(exampleTwoLogs())

	err := test.exp.pushLogsData(context.Background(), logs)
	assert.EqualError(t, err, "error during sending data: 500 Internal Server Error")

	var partial consumererror.Logs
	require.True(t, consumererror.AsLogs(err, &partial))
	assert.Equal(t, logs, partial.GetLogs())
}

func TestPartiallyFailed(t *testing.T) {
	test := prepareExporterTest(t, createTestConfig(), []func(w http.ResponseWriter, req *http.Request){
		func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(500)

			body := extractBody(t, req)
			assert.Equal(t, "Example log", body)
			assert.Equal(t, "key1=value1, key2=value2", req.Header.Get("X-Sumo-Fields"))
		},
		func(w http.ResponseWriter, req *http.Request) {
			body := extractBody(t, req)
			assert.Equal(t, "Another example log", body)
			assert.Equal(t, "key3=value3, key4=value4", req.Header.Get("X-Sumo-Fields"))
		},
	})

	f, err := newFilter([]string{`key\d`})
	require.NoError(t, err)
	test.exp.filter = f

	records := exampleTwoDifferentLogs()
	logs := LogRecordsToLogs(records)
	expected := LogRecordsToLogs(records[:1])

	err = test.exp.pushLogsData(context.Background(), logs)
	assert.EqualError(t, err, "error during sending data: 500 Internal Server Error")

	var partial consumererror.Logs
	require.True(t, consumererror.AsLogs(err, &partial))
	assert.Equal(t, expected, partial.GetLogs())
}

func TestInvalidSourceFormats(t *testing.T) {
	_, err := initExporter(&Config{
		LogFormat:        "json",
		MetricFormat:     "carbon2",
		CompressEncoding: "gzip",
		TraceFormat:      "otlp",
		HTTPClientSettings: confighttp.HTTPClientSettings{
			Timeout:  defaultTimeout,
			Endpoint: "test_endpoint",
		},
		MetadataAttributes: []string{"[a-z"},
	}, createExporterCreateSettings())
	assert.EqualError(t, err, "error parsing regexp: missing closing ]: `[a-z`")
}

func TestInvalidHTTPCLient(t *testing.T) {
	exp, err := initExporter(&Config{
		LogFormat:        "json",
		MetricFormat:     "carbon2",
		CompressEncoding: "gzip",
		TraceFormat:      "otlp",
		HTTPClientSettings: confighttp.HTTPClientSettings{
			Endpoint: "test_endpoint",
			CustomRoundTripper: func(next http.RoundTripper) (http.RoundTripper, error) {
				return nil, errors.New("roundTripperException")
			},
		},
	}, createExporterCreateSettings())
	assert.NoError(t, err)

	assert.EqualError(t,
		exp.start(context.Background(), componenttest.NewNopHost()),
		"failed to create HTTP Client: roundTripperException",
	)
}

func TestPushInvalidCompressor(t *testing.T) {
	test := prepareExporterTest(t, createTestConfig(), []func(w http.ResponseWriter, req *http.Request){
		func(w http.ResponseWriter, req *http.Request) {
			body := extractBody(t, req)
			assert.Equal(t, `Example log`, body)
			assert.Equal(t, "", req.Header.Get("X-Sumo-Fields"))
		},
	})

	logs := LogRecordsToLogs(exampleLog())

	test.exp.config.CompressEncoding = "invalid"

	err := test.exp.pushLogsData(context.Background(), logs)
	assert.EqualError(t, err, "failed to initialize compressor: invalid format: invalid")
}

func TestPushFailedBatch(t *testing.T) {
	test := prepareExporterTest(t, createTestConfig(), []func(w http.ResponseWriter, req *http.Request){
		func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(500)
			body := extractBody(t, req)

			expected := fmt.Sprintf(
				"%s%s",
				strings.Repeat("Example log\n", maxBufferSize-1),
				"Example log",
			)

			assert.Equal(t, expected, body)
			assert.Equal(t, "", req.Header.Get("X-Sumo-Fields"))
		},
		func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(200)
			body := extractBody(t, req)

			assert.Equal(t, `Example log`, body)
			assert.Equal(t, "", req.Header.Get("X-Sumo-Fields"))
		},
	})

	logs := LogRecordsToLogs(exampleLog())
	logs.ResourceLogs().EnsureCapacity(maxBufferSize + 1)
	log := logs.ResourceLogs().At(0)
	rLogs := logs.ResourceLogs()

	for i := 0; i < maxBufferSize; i++ {
		log.CopyTo(rLogs.AppendEmpty())
	}

	err := test.exp.pushLogsData(context.Background(), logs)
	assert.EqualError(t, err, "error during sending data: 500 Internal Server Error")
}

func TestPushOTLPLogsClearTimestamp(t *testing.T) {
	createLogs := func() pdata.Logs {
		exampleLogs := exampleLog()
		exampleLogs[0].SetTimestamp(12345)
		logs := LogRecordsToLogs(exampleLogs)
		return logs
	}

	testcases := []struct {
		name         string
		configFunc   func() *Config
		expectedBody string
	}{
		{
			name: "enabled",
			configFunc: func() *Config {
				config := createTestConfig()
				config.ClearLogsTimestamp = true
				config.LogFormat = OTLPLogFormat
				return config
			},
			expectedBody: "\n\x1b\n\x00\x12\x17\n\x00\x12\x13*\r\n\vExample logJ\x00R\x00",
		},
		{
			name: "disabled",
			configFunc: func() *Config {
				config := createTestConfig()
				config.ClearLogsTimestamp = false
				config.LogFormat = OTLPLogFormat
				return config
			},
			expectedBody: "\n$\n\x00\x12 \n\x00\x12\x1c\t90\x00\x00\x00\x00\x00\x00*\r\n\vExample logJ\x00R\x00",
		},
		{
			name: "default does clear the timestamp",
			configFunc: func() *Config {
				config := createTestConfig()
				// Don't set the clear timestamp config value
				config.LogFormat = OTLPLogFormat
				return config
			},
			expectedBody: "\n\x1b\n\x00\x12\x17\n\x00\x12\x13*\r\n\vExample logJ\x00R\x00",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			expectedRequests := []func(w http.ResponseWriter, req *http.Request){
				func(w http.ResponseWriter, req *http.Request) {
					body := extractBody(t, req)
					assert.Equal(t, tc.expectedBody, body)
				},
			}
			test := prepareExporterTest(t, tc.configFunc(), expectedRequests)

			err := test.exp.pushLogsData(context.Background(), createLogs())
			assert.NoError(t, err)
		})
	}
}

func TestPushTextLogsWithAttributeTranslation(t *testing.T) {
	logs := LogRecordsToLogs(exampleLog())
	logs.ResourceLogs().At(0).Resource().Attributes().InsertString("host.name", "harry-potter")
	logs.ResourceLogs().At(0).Resource().Attributes().InsertString("host.type", "wizard")

	config := createTestConfig()
	config.MetadataAttributes = []string{`host\.name`}
	config.SourceCategory = "%{host.name}"
	config.SourceHost = "%{host}"

	expectedRequests := []func(w http.ResponseWriter, req *http.Request){
		func(w http.ResponseWriter, req *http.Request) {
			body := extractBody(t, req)
			assert.Equal(t, `Example log`, body)
			assert.Equal(t, "host=harry-potter", req.Header.Get("X-Sumo-Fields"))
			assert.Equal(t, "harry-potter", req.Header.Get("X-Sumo-Category"))

			// This gets the value from 'host.name' because we do not disallow
			// using Sumo schema and 'host.name' from OT convention
			// translates into 'host' in Sumo convention
			assert.Equal(t, "harry-potter", req.Header.Get("X-Sumo-Host"))
		},
	}
	test := prepareExporterTest(t, config, expectedRequests)

	err := test.exp.pushLogsData(context.Background(), logs)
	assert.NoError(t, err)
}

func TestPushTextLogsWithAttributeTranslationDisabled(t *testing.T) {
	logs := LogRecordsToLogs(exampleLog())
	logs.ResourceLogs().At(0).Resource().Attributes().InsertString("host.name", "harry-potter")
	logs.ResourceLogs().At(0).Resource().Attributes().InsertString("host.type", "wizard")

	config := createTestConfig()
	config.MetadataAttributes = []string{`host\.name`}
	config.SourceCategory = "%{host.name}"
	config.SourceHost = "%{host}"
	config.TranslateAttributes = false

	expectedRequests := []func(w http.ResponseWriter, req *http.Request){
		func(w http.ResponseWriter, req *http.Request) {
			body := extractBody(t, req)
			assert.Equal(t, `Example log`, body)
			assert.Equal(t, "host.name=harry-potter", req.Header.Get("X-Sumo-Fields"))
			assert.Equal(t, "harry-potter", req.Header.Get("X-Sumo-Category"))
			assert.Equal(t, "undefined", req.Header.Get("X-Sumo-Host"))
		},
	}
	test := prepareExporterTest(t, config, expectedRequests)

	err := test.exp.pushLogsData(context.Background(), logs)
	assert.NoError(t, err)
}

func TestPushJSONLogsWithAttributeTranslation(t *testing.T) {
	logs := LogRecordsToLogs(exampleLog())
	logs.ResourceLogs().At(0).Resource().Attributes().InsertString("host.name", "harry-potter")
	logs.ResourceLogs().At(0).Resource().Attributes().InsertString("host.type", "wizard")

	config := createTestConfig()
	config.MetadataAttributes = []string{`host\.name`}
	config.SourceCategory = "%{host.name}"
	config.SourceHost = "%{host}"
	config.LogFormat = JSONFormat

	expectedRequests := []func(w http.ResponseWriter, req *http.Request){
		func(w http.ResponseWriter, req *http.Request) {
			body := extractBody(t, req)
			var regex string
			// Mind that host attribute is not being send in log body
			regex += `{"InstanceType":"wizard","log":"Example log","timestamp":\d{13}}`
			assert.Regexp(t, regex, body)

			assert.Equal(t, "host=harry-potter", req.Header.Get("X-Sumo-Fields"))
			assert.Equal(t, "harry-potter", req.Header.Get("X-Sumo-Category"))

			// This gets the value from 'host.name' because we do not disallow
			// using Sumo schema and 'host.name' from OT convention
			// translates into 'host' in Sumo convention.
			assert.Equal(t, "harry-potter", req.Header.Get("X-Sumo-Host"))
		},
	}
	test := prepareExporterTest(t, config, expectedRequests)

	err := test.exp.pushLogsData(context.Background(), logs)
	assert.NoError(t, err)
}

func TestPushJSONLogsWithAttributeTranslationDisabled(t *testing.T) {
	logs := LogRecordsToLogs(exampleLog())
	logs.ResourceLogs().At(0).Resource().Attributes().InsertString("host.name", "harry-potter")
	logs.ResourceLogs().At(0).Resource().Attributes().InsertString("host.type", "wizard")

	config := createTestConfig()
	config.MetadataAttributes = []string{`host\.name`}
	config.SourceCategory = "%{host.name}"
	config.SourceHost = "%{host}"
	config.LogFormat = JSONFormat
	config.TranslateAttributes = false

	expectedRequests := []func(w http.ResponseWriter, req *http.Request){
		func(w http.ResponseWriter, req *http.Request) {
			body := extractBody(t, req)
			var regex string
			regex += `{"host.type":"wizard","log":"Example log","timestamp":\d{13}}`
			assert.Regexp(t, regex, body)

			assert.Equal(t, "host.name=harry-potter", req.Header.Get("X-Sumo-Fields"))
			assert.Equal(t, "harry-potter", req.Header.Get("X-Sumo-Category"))
			assert.Equal(t, "undefined", req.Header.Get("X-Sumo-Host"))
		},
	}
	test := prepareExporterTest(t, config, expectedRequests)

	err := test.exp.pushLogsData(context.Background(), logs)
	assert.NoError(t, err)
}

func TestAllMetricsSuccess(t *testing.T) {
	test := prepareExporterTest(t, createTestConfig(), []func(w http.ResponseWriter, req *http.Request){
		func(w http.ResponseWriter, req *http.Request) {
			body := extractBody(t, req)
			expected := `test.metric.data{test="test_value",test2="second_value"} 14500 1605534165000
gauge_metric_name{foo="bar",remote_name="156920",url="http://example_url"} 124 1608124661166
gauge_metric_name{foo="bar",remote_name="156955",url="http://another_url"} 245 1608124662166`
			assert.Equal(t, expected, body)
			assert.Equal(t, "application/vnd.sumologic.prometheus", req.Header.Get("Content-Type"))
		},
	})
	test.exp.config.MetricFormat = PrometheusFormat

	metrics := metricPairToMetrics([]metricPair{
		exampleIntMetric(),
		exampleIntGaugeMetric(),
	})

	err := test.exp.pushMetricsData(context.Background(), metrics)
	assert.NoError(t, err)
}

func TestAllMetricsOTLP(t *testing.T) {
	test := prepareExporterTest(t, createTestConfig(), []func(w http.ResponseWriter, req *http.Request){
		func(w http.ResponseWriter, req *http.Request) {
			body := extractBody(t, req)

			md, err := otlp.NewProtobufMetricsUnmarshaler().UnmarshalMetrics([]byte(body))
			assert.NoError(t, err)
			assert.NotNil(t, md)

			//nolint:lll
			expected := "\nf\n/\n\x14\n\x04test\x12\f\n\ntest_value\n\x17\n\x05test2\x12\x0e\n\fsecond_value\x123\n\x00\x12/\n\x10test.metric.data\x1a\x05bytes:\x14\n\x12\x19\x00\x12\x94\v\xd1\x00H\x161\xa48\x00\x00\x00\x00\x00\x00\n\xc2\x01\n\x0e\n\f\n\x03foo\x12\x05\n\x03bar\x12\xaf\x01\n\x00\x12\xaa\x01\n\x11gauge_metric_name*\x94\x01\nH\x19\x80GX\xef\xdb4Q\x161|\x00\x00\x00\x00\x00\x00\x00:\x17\n\vremote_name\x12\b\n\x06156920:\x1b\n\x03url\x12\x14\n\x12http://example_url\nH\x19\x80\x11\xf3*\xdc4Q\x161\xf5\x00\x00\x00\x00\x00\x00\x00:\x17\n\vremote_name\x12\b\n\x06156955:\x1b\n\x03url\x12\x14\n\x12http://another_url"
			assert.Equal(t, expected, body)
			assert.Equal(t, "application/x-protobuf", req.Header.Get("Content-Type"))
		},
	})
	test.exp.config.MetricFormat = OTLPMetricFormat

	metrics := metricPairToMetrics([]metricPair{
		exampleIntMetric(),
		exampleIntGaugeMetric(),
	})

	err := test.exp.pushMetricsData(context.Background(), metrics)
	assert.NoError(t, err)
}

func TestAllMetricsFailed(t *testing.T) {
	test := prepareExporterTest(t, createTestConfig(), []func(w http.ResponseWriter, req *http.Request){
		func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(500)

			body := extractBody(t, req)
			expected := `test.metric.data{test="test_value",test2="second_value"} 14500 1605534165000
gauge_metric_name{foo="bar",remote_name="156920",url="http://example_url"} 124 1608124661166
gauge_metric_name{foo="bar",remote_name="156955",url="http://another_url"} 245 1608124662166`
			assert.Equal(t, expected, body)
			assert.Equal(t, "application/vnd.sumologic.prometheus", req.Header.Get("Content-Type"))
		},
	})
	test.exp.config.MetricFormat = PrometheusFormat

	metrics := metricPairToMetrics([]metricPair{
		exampleIntMetric(),
		exampleIntGaugeMetric(),
	})

	err := test.exp.pushMetricsData(context.Background(), metrics)
	assert.EqualError(t, err, "error during sending data: 500 Internal Server Error")

	var partial consumererror.Metrics
	require.True(t, consumererror.AsMetrics(err, &partial))
	assert.Equal(t, metrics, partial.GetMetrics())
}

func TestMetricsPartiallyFailed(t *testing.T) {
	test := prepareExporterTest(t, createTestConfig(), []func(w http.ResponseWriter, req *http.Request){
		func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(500)

			body := extractBody(t, req)
			expected := `test.metric.data{test="test_value",test2="second_value"} 14500 1605534165000`
			assert.Equal(t, expected, body)
			assert.Equal(t, "application/vnd.sumologic.prometheus", req.Header.Get("Content-Type"))
		},
		func(w http.ResponseWriter, req *http.Request) {
			body := extractBody(t, req)
			expected := `gauge_metric_name{foo="bar",remote_name="156920",url="http://example_url"} 124 1608124661166
gauge_metric_name{foo="bar",remote_name="156955",url="http://another_url"} 245 1608124662166`
			assert.Equal(t, expected, body)
			assert.Equal(t, "application/vnd.sumologic.prometheus", req.Header.Get("Content-Type"))
		},
	})
	test.exp.config.MetricFormat = PrometheusFormat
	test.exp.config.MaxRequestBodySize = 1

	records := []metricPair{
		exampleIntMetric(),
		exampleIntGaugeMetric(),
	}
	metrics := metricPairToMetrics(records)
	expected := metricPairToMetrics(records[:1])

	err := test.exp.pushMetricsData(context.Background(), metrics)
	assert.EqualError(t, err, "error during sending data: 500 Internal Server Error")

	var partial consumererror.Metrics
	require.True(t, consumererror.AsMetrics(err, &partial))
	assert.Equal(t, expected, partial.GetMetrics())
}

func TestPushMetricsInvalidCompressor(t *testing.T) {
	test := prepareExporterTest(t, createTestConfig(), []func(w http.ResponseWriter, req *http.Request){
		func(w http.ResponseWriter, req *http.Request) {
			body := extractBody(t, req)
			assert.Equal(t, `Example log`, body)
			assert.Equal(t, "", req.Header.Get("X-Sumo-Fields"))
		},
	})

	metrics := metricPairToMetrics([]metricPair{
		exampleIntMetric(),
		exampleIntGaugeMetric(),
	})

	test.exp.config.CompressEncoding = "invalid"

	err := test.exp.pushMetricsData(context.Background(), metrics)
	assert.EqualError(t, err, "failed to initialize compressor: invalid format: invalid")
}

func TestMetricsDifferentMetadata(t *testing.T) {
	test := prepareExporterTest(t, createTestConfig(), []func(w http.ResponseWriter, req *http.Request){
		func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(500)

			body := extractBody(t, req)
			expected := `test.metric.data{test="test_value",test2="second_value",key1="value1"} 14500 1605534165000`
			assert.Equal(t, expected, body)
			assert.Equal(t, "application/vnd.sumologic.prometheus", req.Header.Get("Content-Type"))
		},
		func(w http.ResponseWriter, req *http.Request) {
			body := extractBody(t, req)
			expected := `gauge_metric_name{foo="bar",key2="value2",remote_name="156920",url="http://example_url"} 124 1608124661166
gauge_metric_name{foo="bar",key2="value2",remote_name="156955",url="http://another_url"} 245 1608124662166`
			assert.Equal(t, expected, body)
			assert.Equal(t, "application/vnd.sumologic.prometheus", req.Header.Get("Content-Type"))
		},
	})
	test.exp.config.MetricFormat = PrometheusFormat
	test.exp.config.MaxRequestBodySize = 1

	f, err := newFilter([]string{`key\d`})
	require.NoError(t, err)
	test.exp.filter = f

	records := []metricPair{
		exampleIntMetric(),
		exampleIntGaugeMetric(),
	}

	records[0].attributes.InsertString("key1", "value1")
	records[1].attributes.InsertString("key2", "value2")

	metrics := metricPairToMetrics(records)
	expected := metricPairToMetrics(records[:1])

	err = test.exp.pushMetricsData(context.Background(), metrics)
	assert.EqualError(t, err, "error during sending data: 500 Internal Server Error")

	var partial consumererror.Metrics
	require.True(t, consumererror.AsMetrics(err, &partial))
	assert.Equal(t, expected, partial.GetMetrics())
}

func TestPushMetricsFailedBatch(t *testing.T) {
	t.Skip("Skip test due to prometheus format complexity. Execution can take over 30s")
	test := prepareExporterTest(t, createTestConfig(), []func(w http.ResponseWriter, req *http.Request){
		func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(500)
			body := extractBody(t, req)

			expected := fmt.Sprintf(
				"%s%s",
				strings.Repeat("test_metric_data{test=\"test_value\",test2=\"second_value\"} 14500 1605534165000\n", maxBufferSize-1),
				`test_metric_data{test="test_value",test2="second_value"} 14500 1605534165000`,
			)

			assert.Equal(t, expected, body)
		},
		func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(200)
			body := extractBody(t, req)

			assert.Equal(t, `test_metric_data{test="test_value",test2="second_value"} 14500 1605534165000`, body)
		},
	})
	test.exp.config.MetricFormat = PrometheusFormat
	test.exp.config.MaxRequestBodySize = 1024 * 1024 * 1024 * 1024

	metrics := metricPairToMetrics([]metricPair{exampleIntMetric()})
	metrics.ResourceMetrics().EnsureCapacity(maxBufferSize + 1)
	rMetrics := metrics.ResourceMetrics()
	metric := rMetrics.AppendEmpty()

	for i := 0; i < maxBufferSize; i++ {
		metric.CopyTo(rMetrics.AppendEmpty())
	}

	err := test.exp.pushMetricsData(context.Background(), metrics)
	assert.EqualError(t, err, "error during sending data: 500 Internal Server Error")
}

func TestLogsJsonFormatMetadataFilter(t *testing.T) {
	testcases := []struct {
		name                  string
		logResourceAttributes map[string]pdata.AttributeValue
		cfgFn                 func(c *Config)
		handler               func(w http.ResponseWriter, req *http.Request)
	}{
		{
			name: "basic",
			logResourceAttributes: map[string]pdata.AttributeValue{
				"_sourceCategory": pdata.NewAttributeValueString("dummy"),
				"key1":            pdata.NewAttributeValueString("value1"),
				"key2":            pdata.NewAttributeValueString("value2"),
			},
			cfgFn: func(c *Config) {
				c.LogFormat = JSONFormat
				c.SourceCategory = "testing_source_templates %{_sourceCategory}"
				c.MetadataAttributes = []string{
					"key1",
					"_source.*",
				}
			},
			handler: func(w http.ResponseWriter, req *http.Request) {
				body := extractBody(t, req)

				var regex string
				regex += `{"key2":"value2","log":"Example log","timestamp":\d{13}}`
				assert.Regexp(t, regex, body)

				assert.Equal(t, "key1=value1", req.Header.Get("X-Sumo-Fields"),
					"X-Sumo-Fields is not as expected",
				)

				assert.Equal(t, "testing_source_templates dummy",
					req.Header.Get("X-Sumo-Category"),
					"X-Sumo-Category header is not set correctly",
				)
			},
		},
		{
			name: "source related attributes available for templating even without specifying in metadata attributes",
			logResourceAttributes: map[string]pdata.AttributeValue{
				"_sourceCategory": pdata.NewAttributeValueString("dummy_category"),
				"_sourceHost":     pdata.NewAttributeValueString("dummy_host"),
				"_sourceName":     pdata.NewAttributeValueString("dummy_name"),
				"key1":            pdata.NewAttributeValueString("value1"),
				"key2":            pdata.NewAttributeValueString("value2"),
			},
			cfgFn: func(c *Config) {
				c.LogFormat = JSONFormat
				c.SourceCategory = "testing_source_templates %{_sourceCategory}"
				c.SourceHost = "testing_source_templates %{_sourceHost}"
				c.SourceName = "testing_source_templates %{_sourceName}"
				c.MetadataAttributes = []string{
					"key1",
				}
			},
			handler: func(w http.ResponseWriter, req *http.Request) {
				body := extractBody(t, req)

				var regex string
				regex += `{"key2":"value2","log":"Example log","timestamp":\d{13}}`
				assert.Regexp(t, regex, body)

				assert.Equal(t, "key1=value1", req.Header.Get("X-Sumo-Fields"),
					"X-Sumo-Fields is not as expected",
				)

				assert.Equal(t, "testing_source_templates dummy_category",
					req.Header.Get("X-Sumo-Category"),
					"X-Sumo-Category header is not set correctly",
				)

				assert.Equal(t, "testing_source_templates dummy_host",
					req.Header.Get("X-Sumo-Host"),
					"X-Sumo-Host header is not set correctly",
				)

				assert.Equal(t, "testing_source_templates dummy_name",
					req.Header.Get("X-Sumo-Name"),
					"X-Sumo-Name header is not set correctly",
				)
			},
		},
		{
			name: "unavailable source metadata rendered as undefined",
			logResourceAttributes: map[string]pdata.AttributeValue{
				"_sourceCategory": pdata.NewAttributeValueString("cat"),
				"key1":            pdata.NewAttributeValueString("value1"),
				"key2":            pdata.NewAttributeValueString("value2"),
			},
			cfgFn: func(c *Config) {
				c.LogFormat = JSONFormat
				c.SourceCategory = "dummy %{_sourceCategory}"
				c.SourceHost = "dummy %{_sourceHost}"
				c.SourceName = "dummy %{_sourceName}"
				c.MetadataAttributes = []string{
					"key1",
					"_source.*",
				}
			},
			handler: func(w http.ResponseWriter, req *http.Request) {
				body := extractBody(t, req)

				var regex string
				regex += `{"key2":"value2","log":"Example log","timestamp":\d{13}}`
				assert.Regexp(t, regex, body)

				assert.Equal(t, "key1=value1", req.Header.Get("X-Sumo-Fields"),
					"X-Sumo-Fields is not as expected",
				)

				assert.Equal(t, "dummy cat",
					req.Header.Get("X-Sumo-Category"),
					"X-Sumo-Category header is not set correctly",
				)

				assert.Equal(t, "dummy undefined",
					req.Header.Get("X-Sumo-Host"),
					"X-Sumo-Host header is not set correctly",
				)

				assert.Equal(t, "dummy undefined",
					req.Header.Get("X-Sumo-Name"),
					"X-Sumo-Name header is not set correctly",
				)
			},
		},
		{
			name: "empty attribute",
			logResourceAttributes: map[string]pdata.AttributeValue{
				"_sourceCategory": pdata.NewAttributeValueString("dummy"),
				"key1":            pdata.NewAttributeValueString("value1"),
				"key2":            pdata.NewAttributeValueString("value2"),
				"key3":            pdata.NewAttributeValueString(""),
			},
			cfgFn: func(c *Config) {
				c.LogFormat = JSONFormat
				c.SourceCategory = "testing_source_templates %{_sourceCategory}"
				c.MetadataAttributes = []string{
					"key1",
					"_source.*",
					"key3",
				}
			},
			handler: func(w http.ResponseWriter, req *http.Request) {
				body := extractBody(t, req)

				var regex string
				regex += `{"key2":"value2","log":"Example log","timestamp":\d{13}}`
				assert.Regexp(t, regex, body)

				assert.Equal(t, "key1=value1", req.Header.Get("X-Sumo-Fields"),
					"X-Sumo-Fields is not as expected",
				)

				assert.Equal(t, "testing_source_templates dummy",
					req.Header.Get("X-Sumo-Category"),
					"X-Sumo-Category header is not set correctly",
				)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			test := prepareExporterTest(t, createTestConfig(),
				[]func(http.ResponseWriter, *http.Request){tc.handler},
				tc.cfgFn,
			)

			logs := LogRecordsToLogs(exampleLog())
			logResourceAttrs := logs.ResourceLogs().At(0).Resource().Attributes()
			logResourceAttrs.InitFromMap(tc.logResourceAttributes)

			err := test.exp.pushLogsData(context.Background(), logs)
			assert.NoError(t, err)
		})
	}
}

func TestLogsTextFormatMetadataFilter(t *testing.T) {
	test := prepareExporterTest(t, createTestConfig(), []func(w http.ResponseWriter, req *http.Request){
		func(w http.ResponseWriter, req *http.Request) {
			body := extractBody(t, req)
			assert.Equal(t, `Example log`, body)
			assert.Equal(t, "key1=value1", req.Header.Get("X-Sumo-Fields"))
		},
	})
	test.exp.config.LogFormat = TextFormat

	f, err := newFilter([]string{`key1`})
	require.NoError(t, err)
	test.exp.filter = f

	logs := LogRecordsToLogs(exampleLog())
	logs.ResourceLogs().At(0).Resource().Attributes().InsertString("key1", "value1")
	logs.ResourceLogs().At(0).Resource().Attributes().InsertString("key2", "value2")

	err = test.exp.pushLogsData(context.Background(), logs)
	assert.NoError(t, err)
}

func TestMetricsCarbon2FormatMetadataFilter(t *testing.T) {
	test := prepareExporterTest(t, createTestConfig(), []func(w http.ResponseWriter, req *http.Request){
		func(w http.ResponseWriter, req *http.Request) {
			body := extractBody(t, req)
			expected := `test=test_value test2=second_value key1=value1 key2=value2 metric=test.metric.data unit=bytes  14500 1605534165`
			assert.Equal(t, expected, body)
			assert.Equal(t, "application/vnd.sumologic.carbon2", req.Header.Get("Content-Type"))
		},
	})
	test.exp.config.MetricFormat = Carbon2Format

	f, err := newFilter([]string{`key1`})
	require.NoError(t, err)
	test.exp.filter = f

	records := []metricPair{
		exampleIntMetric(),
	}

	records[0].attributes.InsertString("key1", "value1")
	records[0].attributes.InsertString("key2", "value2")

	metrics := metricPairToMetrics(records)

	err = test.exp.pushMetricsData(context.Background(), metrics)
	assert.NoError(t, err)
}

func TestMetricsGraphiteFormatMetadataFilter(t *testing.T) {
	test := prepareExporterTest(t, createTestConfig(), []func(w http.ResponseWriter, req *http.Request){
		func(w http.ResponseWriter, req *http.Request) {
			body := extractBody(t, req)
			expected := `test_metric_data.test_value.second_value.value1.value2 14500 1605534165`
			assert.Equal(t, expected, body)
			assert.Equal(t, "application/vnd.sumologic.graphite", req.Header.Get("Content-Type"))
		},
	})
	test.exp.config.MetricFormat = GraphiteFormat
	graphiteFormatter, err := newGraphiteFormatter("%{_metric_}.%{test}.%{test2}.%{key1}.%{key2}")
	assert.NoError(t, err)
	test.exp.graphiteFormatter = graphiteFormatter

	f, err := newFilter([]string{`key1`})
	require.NoError(t, err)
	test.exp.filter = f

	records := []metricPair{
		exampleIntMetric(),
	}

	records[0].attributes.InsertString("key1", "value1")
	records[0].attributes.InsertString("key2", "value2")

	metrics := metricPairToMetrics(records)

	err = test.exp.pushMetricsData(context.Background(), metrics)
	assert.NoError(t, err)
}

func TestMetricsPrometheusFormatMetadataFilter(t *testing.T) {
	test := prepareExporterTest(t, createTestConfig(), []func(w http.ResponseWriter, req *http.Request){
		func(w http.ResponseWriter, req *http.Request) {
			body := extractBody(t, req)
			expected := `test.metric.data{test="test_value",test2="second_value",key1="value1",key2="value2"} 14500 1605534165000`
			assert.Equal(t, expected, body)
			assert.Equal(t, "application/vnd.sumologic.prometheus", req.Header.Get("Content-Type"))
		},
	})
	test.exp.config.MetricFormat = PrometheusFormat

	f, err := newFilter([]string{`key1`})
	require.NoError(t, err)
	test.exp.filter = f

	records := []metricPair{
		exampleIntMetric(),
	}

	records[0].attributes.InsertString("key1", "value1")
	records[0].attributes.InsertString("key2", "value2")

	metrics := metricPairToMetrics(records)

	err = test.exp.pushMetricsData(context.Background(), metrics)
	assert.NoError(t, err)
}

func TestPushMetrics_AttributeTranslation(t *testing.T) {
	createConfig := func() *Config {
		config := createDefaultConfig().(*Config)
		config.CompressEncoding = NoCompression
		config.LogFormat = TextFormat
		config.MetricFormat = PrometheusFormat
		return config
	}

	testcases := []struct {
		name            string
		cfgFn           func() *Config
		expectedHeaders map[string]string
		expectedBody    string
	}{
		{
			name: "enabled",
			cfgFn: func() *Config {
				cfg := createConfig()
				cfg.MetadataAttributes = []string{`host\.name`}
				cfg.SourceCategory = "%{host.name}"
				cfg.SourceHost = "%{host}"
				// This is and should be done by default:
				// cfg.TranslateAttributes = true
				return cfg
			},
			expectedHeaders: map[string]string{
				"Content-Type":    "application/vnd.sumologic.prometheus",
				"X-Sumo-Category": "harry-potter",

				// This gets the value from 'host.name' because we do not disallow
				// using Sumo schema and 'host.name' from OT convention
				// translates into 'host' in Sumo convention
				"X-Sumo-Host": "harry-potter",
			},
			expectedBody: `test.metric.data{test="test_value",test2="second_value",host="harry-potter",InstanceType="wizard"} 14500 1605534165000`,
		},
		{
			name: "enabled_with_ot_host_name_template_set_in_source_host",
			cfgFn: func() *Config {
				cfg := createConfig()
				cfg.MetadataAttributes = []string{`host\.name`}
				cfg.SourceCategory = "%{host.name}"
				cfg.SourceHost = "%{host.name}"
				// This is and should be done by default:
				// cfg.TranslateAttributes = true
				return cfg
			},
			expectedHeaders: map[string]string{
				"Content-Type":    "application/vnd.sumologic.prometheus",
				"X-Sumo-Category": "harry-potter",
				"X-Sumo-Host":     "harry-potter",
			},
			expectedBody: `test.metric.data{test="test_value",test2="second_value",host="harry-potter",InstanceType="wizard"} 14500 1605534165000`,
		},
		{
			name: "enabled_with_proper_host_template_set_in_source_host_but_not_specified_in_metadata_attributes",
			cfgFn: func() *Config {
				cfg := createConfig()
				// This is the default
				// cfg.MetadataAttributes = []string{}
				cfg.SourceCategory = "%{host.name}"
				cfg.SourceHost = "%{host.name}"
				// This is and should be done by default:
				// cfg.TranslateAttributes = true
				return cfg
			},
			expectedHeaders: map[string]string{
				"Content-Type":    "application/vnd.sumologic.prometheus",
				"X-Sumo-Category": "undefined",
				"X-Sumo-Host":     "undefined",
			},
			expectedBody: `test.metric.data{test="test_value",test2="second_value",host="harry-potter",InstanceType="wizard"} 14500 1605534165000`,
		},
		{
			name: "disabled",
			cfgFn: func() *Config {
				cfg := createConfig()
				cfg.MetadataAttributes = []string{`host\.name`}
				cfg.SourceCategory = "%{host.name}"
				cfg.SourceHost = "%{host}"
				cfg.TranslateAttributes = false
				return cfg
			},
			expectedHeaders: map[string]string{
				"Content-Type":    "application/vnd.sumologic.prometheus",
				"X-Sumo-Category": "harry-potter",
				"X-Sumo-Host":     "undefined",
			},
			expectedBody: `test.metric.data{test="test_value",test2="second_value",host.name="harry-potter",host.type="wizard"} 14500 1605534165000`,
		},
		{
			name: "disabled_with_ot_host_name_template_set_in_source_host",
			cfgFn: func() *Config {
				cfg := createConfig()
				cfg.MetadataAttributes = []string{`host\.name`}
				cfg.SourceCategory = "%{host.name}"
				cfg.SourceHost = "%{host.name}"
				cfg.TranslateAttributes = false
				return cfg
			},
			expectedHeaders: map[string]string{
				"Content-Type":    "application/vnd.sumologic.prometheus",
				"X-Sumo-Category": "harry-potter",
				"X-Sumo-Host":     "harry-potter",
			},
			expectedBody: `test.metric.data{test="test_value",test2="second_value",host.name="harry-potter",host.type="wizard"} 14500 1605534165000`,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			records := []metricPair{
				exampleIntMetric(),
			}
			records[0].attributes.InsertString("host.name", "harry-potter")
			records[0].attributes.InsertString("host.type", "wizard")

			metrics := metricPairToMetrics(records)

			config := tc.cfgFn()
			callbacks := []func(w http.ResponseWriter, req *http.Request){
				func(w http.ResponseWriter, req *http.Request) {
					assert.Equal(t, tc.expectedBody, extractBody(t, req))
					for header, expectedValue := range tc.expectedHeaders {
						assert.Equalf(t, expectedValue, req.Header.Get(header),
							"Unexpected value in header: %s", header,
						)
					}
				},
			}
			test := prepareExporterTest(t, config, callbacks)

			err := test.exp.pushMetricsData(context.Background(), metrics)
			assert.NoError(t, err)
		})
	}
}

func TestPushMetrics_MetricsTranslation(t *testing.T) {
	createConfig := func() *Config {
		config := createDefaultConfig().(*Config)
		config.CompressEncoding = NoCompression
		config.MetricFormat = PrometheusFormat
		return config
	}

	testcases := []struct {
		name         string
		cfgFn        func() *Config
		metricsFn    func() pdata.Metrics
		expectedBody string
	}{
		{
			name: "enabled by default translated metrics successfully",
			cfgFn: func() *Config {
				cfg := createConfig()
				// This is and should be done by default:
				// cfg.TranslateTelegrafMetrics = true
				return cfg
			},
			metricsFn: func() pdata.Metrics {
				metrics := pdata.NewMetrics()
				{
					m := metrics.ResourceMetrics().AppendEmpty().
						InstrumentationLibraryMetrics().AppendEmpty().
						Metrics().AppendEmpty()
					m.SetName("cpu_usage_active")
					m.SetDataType(pdata.MetricDataTypeGauge)
					dp := m.Gauge().DataPoints().AppendEmpty()
					dp.SetDoubleVal(123.456)
					dp.SetTimestamp(pdata.NewTimestampFromTime(time.Unix(1605534165, 0)))
					dp.Attributes().InsertString("test", "test_value")
				}
				return metrics
			},
			expectedBody: `CPU_Total{test="test_value"} 123.456 1605534165000`,
		},
		{
			name: "enabled 3 metrics with 1 not to be translated",
			cfgFn: func() *Config {
				cfg := createConfig()
				// This is and should be done by default:
				// cfg.TranslateTelegrafMetrics = true
				return cfg
			},
			metricsFn: func() pdata.Metrics {
				metrics := pdata.NewMetrics()
				{
					m := metrics.ResourceMetrics().AppendEmpty().
						InstrumentationLibraryMetrics().AppendEmpty().
						Metrics().AppendEmpty()
					m.SetName("cpu_usage_active")
					m.SetDataType(pdata.MetricDataTypeGauge)
					dp := m.Gauge().DataPoints().AppendEmpty()
					dp.SetDoubleVal(123.456)
					dp.SetTimestamp(pdata.NewTimestampFromTime(time.Unix(1605534165, 0)))
					dp.Attributes().InsertString("test", "test_value")
				}
				{
					m := metrics.ResourceMetrics().AppendEmpty().
						InstrumentationLibraryMetrics().AppendEmpty().
						Metrics().AppendEmpty()
					m.SetName("diskio_reads")
					m.SetDataType(pdata.MetricDataTypeGauge)
					dp := m.Gauge().DataPoints().AppendEmpty()
					dp.SetIntVal(123456)
					dp.SetTimestamp(pdata.NewTimestampFromTime(time.Unix(1605534165, 1000000)))
					dp.Attributes().InsertString("test", "test_value")
				}
				{
					m := metrics.ResourceMetrics().AppendEmpty().
						InstrumentationLibraryMetrics().AppendEmpty().
						Metrics().AppendEmpty()
					m.SetName("dummy_metric")
					m.SetDataType(pdata.MetricDataTypeGauge)
					dp := m.Gauge().DataPoints().AppendEmpty()
					dp.SetIntVal(10)
					dp.SetTimestamp(pdata.NewTimestampFromTime(time.Unix(1605534165, 2000000)))
					dp.Attributes().InsertString("test", "test_value")
				}
				return metrics
			},
			expectedBody: `CPU_Total{test="test_value"} 123.456 1605534165000
Disk_Reads{test="test_value"} 123456 1605534165001
dummy_metric{test="test_value"} 10 1605534165002`,
		},
		{
			name: "disabled does not translate metric names",
			cfgFn: func() *Config {
				cfg := createConfig()
				cfg.TranslateTelegrafMetrics = false
				return cfg
			},
			metricsFn: func() pdata.Metrics {
				metrics := pdata.NewMetrics()
				{
					m := metrics.ResourceMetrics().AppendEmpty().
						InstrumentationLibraryMetrics().AppendEmpty().
						Metrics().AppendEmpty()
					m.SetName("cpu_usage_active")
					m.SetDataType(pdata.MetricDataTypeGauge)
					dp := m.Gauge().DataPoints().AppendEmpty()
					dp.SetDoubleVal(123.456)
					dp.SetTimestamp(pdata.NewTimestampFromTime(time.Unix(1605534165, 0)))
					dp.Attributes().InsertString("test", "test_value")
				}
				return metrics
			},
			expectedBody: `cpu_usage_active{test="test_value"} 123.456 1605534165000`,
		},
		{
			name: "disabled does not translate metric names - 3 metrics",
			cfgFn: func() *Config {
				cfg := createConfig()
				cfg.TranslateTelegrafMetrics = false
				return cfg
			},
			metricsFn: func() pdata.Metrics {
				metrics := pdata.NewMetrics()
				{
					m := metrics.ResourceMetrics().AppendEmpty().
						InstrumentationLibraryMetrics().AppendEmpty().
						Metrics().AppendEmpty()
					m.SetName("cpu_usage_active")
					m.SetDataType(pdata.MetricDataTypeGauge)
					dp := m.Gauge().DataPoints().AppendEmpty()
					dp.SetDoubleVal(123.456)
					dp.SetTimestamp(pdata.NewTimestampFromTime(time.Unix(1605534165, 0)))
					dp.Attributes().InsertString("test", "test_value")
				}
				{
					m := metrics.ResourceMetrics().AppendEmpty().
						InstrumentationLibraryMetrics().AppendEmpty().
						Metrics().AppendEmpty()
					m.SetName("diskio_reads")
					m.SetDataType(pdata.MetricDataTypeGauge)
					dp := m.Gauge().DataPoints().AppendEmpty()
					dp.SetIntVal(123456)
					dp.SetTimestamp(pdata.NewTimestampFromTime(time.Unix(1605534165, 1000000)))
					dp.Attributes().InsertString("test", "test_value")
				}
				{
					m := metrics.ResourceMetrics().AppendEmpty().
						InstrumentationLibraryMetrics().AppendEmpty().
						Metrics().AppendEmpty()
					m.SetName("dummy_metric")
					m.SetDataType(pdata.MetricDataTypeGauge)
					dp := m.Gauge().DataPoints().AppendEmpty()
					dp.SetIntVal(10)
					dp.SetTimestamp(pdata.NewTimestampFromTime(time.Unix(1605534165, 2000000)))
					dp.Attributes().InsertString("test", "test_value")
				}
				return metrics
			},
			expectedBody: `cpu_usage_active{test="test_value"} 123.456 1605534165000
diskio_reads{test="test_value"} 123456 1605534165001
dummy_metric{test="test_value"} 10 1605534165002`,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			metrics := tc.metricsFn()
			config := tc.cfgFn()

			callbacks := []func(w http.ResponseWriter, req *http.Request){
				func(w http.ResponseWriter, req *http.Request) {
					assert.Equal(t, tc.expectedBody, extractBody(t, req))
				},
			}
			test := prepareExporterTest(t, config, callbacks)

			err := test.exp.pushMetricsData(context.Background(), metrics)
			assert.NoError(t, err)
		})
	}
}
