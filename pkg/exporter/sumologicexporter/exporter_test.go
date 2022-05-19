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
	"sync"
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
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.uber.org/zap"
)

func LogRecordsToLogs(records []plog.LogRecord) plog.Logs {
	logs := plog.NewLogs()
	logsSlice := logs.ResourceLogs().AppendEmpty().ScopeLogs().AppendEmpty().LogRecords()
	for _, record := range records {
		record.CopyTo(logsSlice.AppendEmpty())
	}

	return logs
}

type exporterTest struct {
	srv        *httptest.Server
	exp        *sumologicexporter
	reqCounter *int32
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
		c := int(atomic.LoadInt32(&reqCounter))
		if assert.Greaterf(t, len(cb), c, "Exporter sent more requests (%d) than the number of test callbacks defined: %d", c+1, len(cb)) {
			cb[c](w, req)
			atomic.AddInt32(&reqCounter, 1)
		}
	}))
	t.Cleanup(func() {
		testServer.Close()

		// Ensure we got all required requests
		assert.Eventuallyf(t, func() bool {
			return int(atomic.LoadInt32(&reqCounter)) == len(cb)
		}, 2*time.Second, 100*time.Millisecond,
			"HTTP server didn't receive all the expected requests; got: %d, expected: %d",
			atomic.LoadInt32(&reqCounter), len(cb),
		)
	})

	cfg.HTTPClientSettings.Endpoint = testServer.URL
	cfg.HTTPClientSettings.Auth = nil
	for _, cfgOpt := range cfgOpts {
		cfgOpt(cfg)
	}

	exp, err := initExporter(cfg, createExporterCreateSettings())
	require.NoError(t, err)

	require.NoError(t, exp.start(context.Background(), componenttest.NewNopHost()))

	return &exporterTest{
		srv:        testServer,
		exp:        exp,
		reqCounter: &reqCounter,
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

func TestLogsResourceAttributesSentAsFields(t *testing.T) {
	testcases := []struct {
		name       string
		configFunc func() *Config
		callbacks  []func(w http.ResponseWriter, req *http.Request)
		logsFunc   func() plog.Logs
	}{
		{
			name: "text",
			configFunc: func() *Config {
				config := createTestConfig()
				config.LogFormat = TextFormat
				// config.MetadataAttributes = []string{".*"}
				return config
			},
			callbacks: []func(w http.ResponseWriter, req *http.Request){
				func(w http.ResponseWriter, req *http.Request) {
					// b, err := httputil.DumpRequest(req, true)
					// assert.NoError(t, err)
					// fmt.Printf("body:\n%s\n", string(b))
					body := extractBody(t, req)
					assert.Equal(t, "Example log\nAnother example log", body)
					assert.Equal(t, "res_attr1=1, res_attr2=2", req.Header.Get("X-Sumo-Fields"))
				},
			},
			logsFunc: func() plog.Logs {
				buffer := make([]plog.LogRecord, 2)
				buffer[0] = plog.NewLogRecord()
				buffer[0].Body().SetStringVal("Example log")
				buffer[0].Attributes().InsertString("key1", "value1")
				buffer[0].Attributes().InsertString("key2", "value2")
				buffer[1] = plog.NewLogRecord()
				buffer[1].Body().SetStringVal("Another example log")
				buffer[1].Attributes().InsertString("key1", "value1")
				buffer[1].Attributes().InsertString("key2", "value2")
				buffer[1].Attributes().InsertString("key3", "value3")

				logs := LogRecordsToLogs(buffer)
				logs.ResourceLogs().At(0).Resource().Attributes().InsertString("res_attr1", "1")
				logs.ResourceLogs().At(0).Resource().Attributes().InsertString("res_attr2", "2")
				return logs
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := tc.configFunc()
			test := prepareExporterTest(t, cfg, tc.callbacks)

			logs := tc.logsFunc()
			assert.NoError(t, test.exp.pushLogsData(context.Background(), logs))
			assert.EqualValues(t, len(tc.callbacks), atomic.LoadInt32(test.reqCounter))
		})
	}
}

func TestAllFailed(t *testing.T) {
	test := prepareExporterTest(t, createTestConfig(), []func(w http.ResponseWriter, req *http.Request){
		func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(500)

			body := extractBody(t, req)
			assert.Equal(t, "Example log\nAnother example log", body)
			assert.Empty(t, req.Header.Get("X-Sumo-Fields"))
		},
	})

	logs := plog.NewLogs()
	logsSlice := logs.ResourceLogs().AppendEmpty()
	logsRecords1 := logsSlice.ScopeLogs().AppendEmpty().LogRecords()
	logsRecords1.AppendEmpty().Body().SetStringVal("Example log")

	logsRecords2 := logsSlice.ScopeLogs().AppendEmpty().LogRecords()
	logsRecords2.AppendEmpty().Body().SetStringVal("Another example log")

	logsExpected := plog.NewLogs()
	logsSlice.CopyTo(logsExpected.ResourceLogs().AppendEmpty())

	err := test.exp.pushLogsData(context.Background(), logs)
	assert.EqualError(t, err, "failed sending data: status: 500 Internal Server Error")

	var partial consumererror.Logs
	require.True(t, errors.As(err, &partial))
	assert.Equal(t, logsExpected, partial.GetLogs())
}

func TestPartiallyFailed(t *testing.T) {
	test := prepareExporterTest(t, createTestConfig(), []func(w http.ResponseWriter, req *http.Request){
		func(w http.ResponseWriter, req *http.Request) {
			body := extractBody(t, req)
			assert.Equal(t, "Example log", body)
			// No resource attributes for those logs hence no fields
			assert.Empty(t, req.Header.Get("X-Sumo-Fields"))
		},
		func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(500)

			body := extractBody(t, req)
			assert.Equal(t, "Another example log", body)
			// No resource attributes for those logs hence no fields
			assert.Empty(t, req.Header.Get("X-Sumo-Fields"))
		},
	})

	logs := plog.NewLogs()
	logsSlice1 := logs.ResourceLogs().AppendEmpty()
	logsRecords1 := logsSlice1.ScopeLogs().AppendEmpty().LogRecords()
	logsRecords1.AppendEmpty().Body().SetStringVal("Example log")
	logsSlice2 := logs.ResourceLogs().AppendEmpty()
	logsRecords2 := logsSlice2.ScopeLogs().AppendEmpty().LogRecords()
	logsRecords2.AppendEmpty().Body().SetStringVal("Another example log")

	logsExpected := plog.NewLogs()
	logsSlice2.CopyTo(logsExpected.ResourceLogs().AppendEmpty())

	err := test.exp.pushLogsData(context.Background(), logs)
	assert.EqualError(t, err, "failed sending data: status: 500 Internal Server Error")

	var partial consumererror.Logs
	require.True(t, errors.As(err, &partial))
	assert.Equal(t, logsExpected, partial.GetLogs())
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
	// Expect no requests
	test := prepareExporterTest(t, createTestConfig(), nil)
	test.exp.config.CompressEncoding = "invalid"

	logs := LogRecordsToLogs(exampleLog())

	err := test.exp.pushLogsData(context.Background(), logs)
	assert.EqualError(t, err, "failed to initialize compressor: invalid format: invalid")
}

func TestPushFailedBatch(t *testing.T) {
	t.Skip()

	t.Parallel()

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
	assert.EqualError(t, err, "failed sending data: status: 500 Internal Server Error")
}

func TestPushOTLPLogsClearTimestamp(t *testing.T) {
	createLogs := func() plog.Logs {
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

func TestPushOTLPLogs_AttributeTranslation(t *testing.T) {
	createLogs := func() plog.Logs {
		logs := LogRecordsToLogs(exampleLog())
		resourceAttrs := logs.ResourceLogs().At(0).Resource().Attributes()
		resourceAttrs.InsertString("host.name", "harry-potter")
		resourceAttrs.InsertString("host.type", "wizard")
		resourceAttrs.InsertString("log.file.path_resolved", "/tmp/log.log")
		return logs
	}

	testcases := []struct {
		name       string
		configFunc func() *Config
		callbacks  []func(w http.ResponseWriter, req *http.Request)
	}{
		{
			name: "enabled",
			configFunc: func() *Config {
				config := createTestConfig()
				config.SourceCategory = "category_with_host_template_%{host.name}"
				config.SourceHost = "%{host.name}"
				config.LogFormat = OTLPLogFormat
				config.TranslateAttributes = true
				return config
			},
			callbacks: []func(w http.ResponseWriter, req *http.Request){
				func(w http.ResponseWriter, req *http.Request) {
					body := extractBody(t, req)

					//nolint:lll
					assert.Equal(t, "\n\xcb\x01\n\xaf\x01\n\x16\n\x04host\x12\x0e\n\fharry-potter\n\x18\n\fInstanceType\x12\b\n\x06wizard\n\x1d\n\v_sourceName\x12\x0e\n\f/tmp/log.log\n\x1d\n\v_sourceHost\x12\x0e\n\fharry-potter\n=\n\x0f_sourceCategory\x12*\n(category_with_host_template_harry-potter\x12\x17\n\x00\x12\x13*\r\n\vExample logJ\x00R\x00", body)

					assert.Empty(t, req.Header.Get("X-Sumo-Fields"),
						"We should not get X-Sumo-Fields header when sending data with OTLP",
					)
					assert.Empty(t, req.Header.Get("X-Sumo-Category"),
						"We should not get X-Sumo-Category header when sending data with OTLP",
					)
					assert.Empty(t, req.Header.Get("X-Sumo-Name"),
						"We should not get X-Sumo-Name header when sending data with OTLP",
					)
					assert.Empty(t, req.Header.Get("X-Sumo-Host"),
						"We should not get X-Sumo-Host header when sending data with OTLP",
					)
				},
			},
		},
		{
			name: "disabled",
			configFunc: func() *Config {
				config := createTestConfig()
				config.SourceCategory = "category_with_host_template_%{host.name}"
				config.SourceHost = "%{host.name}"
				config.LogFormat = OTLPLogFormat
				config.TranslateAttributes = false
				return config
			},
			callbacks: []func(w http.ResponseWriter, req *http.Request){
				func(w http.ResponseWriter, req *http.Request) {
					body := extractBody(t, req)

					//nolint:lll
					assert.Equal(t, "\n\xd8\x01\n\xbc\x01\n\x1b\n\thost.name\x12\x0e\n\fharry-potter\n\x15\n\thost.type\x12\b\n\x06wizard\n(\n\x16log.file.path_resolved\x12\x0e\n\f/tmp/log.log\n\x1d\n\v_sourceHost\x12\x0e\n\fharry-potter\n=\n\x0f_sourceCategory\x12*\n(category_with_host_template_harry-potter\x12\x17\n\x00\x12\x13*\r\n\vExample logJ\x00R\x00", body)

					assert.Empty(t, req.Header.Get("X-Sumo-Fields"),
						"We should not get X-Sumo-Fields header when sending data with OTLP",
					)
					assert.Empty(t, req.Header.Get("X-Sumo-Category"),
						"We should not get X-Sumo-Category header when sending data with OTLP",
					)
					assert.Empty(t, req.Header.Get("X-Sumo-Name"),
						"We should not get X-Sumo-Name header when sending data with OTLP",
					)
					assert.Empty(t, req.Header.Get("X-Sumo-Host"),
						"We should not get X-Sumo-Host header when sending data with OTLP",
					)
				},
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			tc := tc

			test := prepareExporterTest(t, tc.configFunc(), tc.callbacks)

			err := test.exp.pushLogsData(context.Background(), createLogs())
			assert.NoError(t, err)
		})
	}
}

func TestPushTextLogs_AttributeTranslation(t *testing.T) {
	createLogs := func() plog.Logs {
		logs := LogRecordsToLogs(exampleLog())
		resourceAttrs := logs.ResourceLogs().At(0).Resource().Attributes()
		resourceAttrs.InsertString("host.name", "harry-potter")
		resourceAttrs.InsertString("host.type", "wizard")
		return logs
	}

	testcases := []struct {
		name       string
		configFunc func() *Config
		callbacks  []func(w http.ResponseWriter, req *http.Request)
	}{
		{
			name: "enabled",
			configFunc: func() *Config {
				config := createTestConfig()
				config.SourceCategory = "%{host.name}"
				config.SourceHost = "%{host}"
				config.LogFormat = TextFormat
				config.TranslateAttributes = true
				return config
			},
			callbacks: []func(w http.ResponseWriter, req *http.Request){
				func(w http.ResponseWriter, req *http.Request) {
					body := extractBody(t, req)
					assert.Equal(t, `Example log`, body)
					assert.Equal(t, "InstanceType=wizard, host=harry-potter", req.Header.Get("X-Sumo-Fields"), "X-Sumo-Fields")
					assert.Equal(t, "harry-potter", req.Header.Get("X-Sumo-Category"), "X-Sumo-Category")

					// This gets the value from 'host.name' because we do not disallow
					// using Sumo schema and 'host.name' from OT convention
					// translates into 'host' in Sumo convention
					assert.Equal(t, "harry-potter", req.Header.Get("X-Sumo-Host"), "X-Sumo-Host")
				},
			},
		},
		{
			name: "disabled",
			configFunc: func() *Config {
				config := createTestConfig()
				config.SourceCategory = "%{host.name}"
				config.SourceHost = "%{host}"
				config.LogFormat = TextFormat
				config.TranslateAttributes = false
				return config
			},
			callbacks: []func(w http.ResponseWriter, req *http.Request){
				func(w http.ResponseWriter, req *http.Request) {
					body := extractBody(t, req)
					assert.Equal(t, `Example log`, body)
					assert.Equal(t, "host.name=harry-potter, host.type=wizard", req.Header.Get("X-Sumo-Fields"), "X-Sumo-Fields")
					assert.Equal(t, "harry-potter", req.Header.Get("X-Sumo-Category"), "X-Sumo-Category")
					assert.Equal(t, "undefined", req.Header.Get("X-Sumo-Host"), "X-Sumo-Host")
				},
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			tc := tc

			test := prepareExporterTest(t, tc.configFunc(), tc.callbacks)

			err := test.exp.pushLogsData(context.Background(), createLogs())
			assert.NoError(t, err)
		})
	}
}

func TestPushJSONLogs_AttributeTranslation(t *testing.T) {
	createLogs := func() plog.Logs {
		logs := LogRecordsToLogs(exampleLog())
		resourceAttrs := logs.ResourceLogs().At(0).Resource().Attributes()
		resourceAttrs.InsertString("host.name", "harry-potter")
		resourceAttrs.InsertString("host.type", "wizard")
		return logs
	}

	testcases := []struct {
		name       string
		logs       plog.Logs
		configFunc func() *Config
		callbacks  []func(w http.ResponseWriter, req *http.Request)
	}{
		{
			name: "enabled",
			logs: createLogs(),
			configFunc: func() *Config {
				config := createTestConfig()
				config.SourceCategory = "%{host.name}"
				config.SourceHost = "%{host}"
				config.LogFormat = JSONFormat
				config.TranslateAttributes = true
				return config
			},
			callbacks: []func(w http.ResponseWriter, req *http.Request){
				func(w http.ResponseWriter, req *http.Request) {
					body := extractBody(t, req)

					var regex string
					// Mind that host attribute is not being send in log body
					regex += `{"log":"Example log","timestamp":\d{13}}`
					assert.Regexp(t, regex, body)

					assert.Equal(t, "InstanceType=wizard, host=harry-potter", req.Header.Get("X-Sumo-Fields"), "X-Sumo-Fields")
					assert.Equal(t, "harry-potter", req.Header.Get("X-Sumo-Category"), "X-Sumo-Category")

					// This gets the value from 'host.name' because we do not disallow
					// using Sumo schema and 'host.name' from OT convention
					// translates into 'host' in Sumo convention
					assert.Equal(t, "harry-potter", req.Header.Get("X-Sumo-Host"), "X-Sumo-Host")
				},
			},
		},
		{
			name: "disabled",
			logs: createLogs(),
			configFunc: func() *Config {
				config := createTestConfig()
				config.SourceCategory = "%{host.name}"
				config.SourceHost = "%{host}"
				config.LogFormat = JSONFormat
				config.TranslateAttributes = false
				return config
			},
			callbacks: []func(w http.ResponseWriter, req *http.Request){
				func(w http.ResponseWriter, req *http.Request) {
					body := extractBody(t, req)
					var regex string
					regex += `{"log":"Example log","timestamp":\d{13}}`
					assert.Regexp(t, regex, body)

					assert.Equal(t, "host.name=harry-potter, host.type=wizard", req.Header.Get("X-Sumo-Fields"), "X-Sumo-Fields")
					assert.Equal(t, "harry-potter", req.Header.Get("X-Sumo-Category"), "X-Sumo-Category")
					assert.Equal(t, "undefined", req.Header.Get("X-Sumo-Host"), "X-Sumo-Host")
				},
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			tc := tc

			test := prepareExporterTest(t, tc.configFunc(), tc.callbacks)

			err := test.exp.pushLogsData(context.Background(), tc.logs)
			assert.NoError(t, err)
		})
	}
}

func TestPushLogs_DontRemoveSourceAttributes(t *testing.T) {
	createLogs := func() plog.Logs {
		logs := plog.NewLogs()
		resourceLogs := logs.ResourceLogs().AppendEmpty()
		logsSlice := resourceLogs.ScopeLogs().AppendEmpty().LogRecords()

		logRecords := make([]plog.LogRecord, 2)
		logRecords[0] = plog.NewLogRecord()
		logRecords[0].Body().SetStringVal("Example log aaaaaaaaaaaaaaaaaaaaaa 1")
		logRecords[0].CopyTo(logsSlice.AppendEmpty())
		logRecords[1] = plog.NewLogRecord()
		logRecords[1].Body().SetStringVal("Example log aaaaaaaaaaaaaaaaaaaaaa 2")
		logRecords[1].CopyTo(logsSlice.AppendEmpty())

		resourceAttrs := resourceLogs.Resource().Attributes()
		resourceAttrs.InsertString("hostname", "my-host-name")
		resourceAttrs.InsertString("hosttype", "my-host-type")
		resourceAttrs.InsertString("_sourceCategory", "my-source-category")
		resourceAttrs.InsertString("_sourceHost", "my-source-host")
		resourceAttrs.InsertString("_sourceName", "my-source-name")

		return logs
	}

	callbacks := []func(w http.ResponseWriter, req *http.Request){
		func(w http.ResponseWriter, req *http.Request) {
			body := extractBody(t, req)
			assert.Equal(t, "Example log aaaaaaaaaaaaaaaaaaaaaa 1", body)
			assert.Equal(t, "hostname=my-host-name, hosttype=my-host-type", req.Header.Get("X-Sumo-Fields"))
			assert.Equal(t, "my-source-category", req.Header.Get("X-Sumo-Category"))
			assert.Equal(t, "my-source-host", req.Header.Get("X-Sumo-Host"))
			assert.Equal(t, "my-source-name", req.Header.Get("X-Sumo-Name"))
			for k, v := range req.Header {
				t.Logf("request #1 header: %v=%v", k, v)
			}
		},
		func(w http.ResponseWriter, req *http.Request) {
			body := extractBody(t, req)
			assert.Equal(t, "Example log aaaaaaaaaaaaaaaaaaaaaa 2", body)
			assert.Equal(t, "hostname=my-host-name, hosttype=my-host-type", req.Header.Get("X-Sumo-Fields"))
			assert.Equal(t, "my-source-category", req.Header.Get("X-Sumo-Category"))
			assert.Equal(t, "my-source-host", req.Header.Get("X-Sumo-Host"))
			assert.Equal(t, "my-source-name", req.Header.Get("X-Sumo-Name"))
			for k, v := range req.Header {
				t.Logf("request #2 header: %v=%v", k, v)
			}
		},
	}

	config := createTestConfig()
	config.SourceCategory = "%{_sourceCategory}"
	config.SourceName = "%{_sourceName}"
	config.SourceHost = "%{_sourceHost}"
	config.LogFormat = TextFormat
	config.TranslateAttributes = false
	config.MaxRequestBodySize = 32

	test := prepareExporterTest(t, config, callbacks)
	assert.NoError(t, test.exp.pushLogsData(context.Background(), createLogs()))
}

func TestAllMetricsSuccess(t *testing.T) {
	testcases := []struct {
		name         string
		expectedBody string
		metricFunc   func() (pmetric.Metric, pcommon.Map)
	}{
		{
			name:         "sum",
			expectedBody: `test.metric.data{test="test_value",test2="second_value"} 14500 1605534165000`,
			metricFunc:   exampleIntMetric,
		},
		{
			name: "gauge",
			expectedBody: `gauge_metric_name{foo="bar",remote_name="156920",url="http://example_url"} 124 1608124661166
gauge_metric_name{foo="bar",remote_name="156955",url="http://another_url"} 245 1608124662166`,
			metricFunc: exampleIntGaugeMetric,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			test := prepareExporterTest(t, createTestConfig(), []func(w http.ResponseWriter, req *http.Request){
				func(w http.ResponseWriter, req *http.Request) {
					body := extractBody(t, req)
					assert.Equal(t, tc.expectedBody, body)
					assert.Equal(t, "application/vnd.sumologic.prometheus", req.Header.Get("Content-Type"))
				},
			})
			test.exp.config.MetricFormat = PrometheusFormat

			metric := metricAndAttributesToPdataMetrics(tc.metricFunc())

			err := test.exp.pushMetricsData(context.Background(), metric)
			assert.NoError(t, err)
		})
	}
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

	metricSum, attrsSum := exampleIntMetric()
	metricGauge, attrsGauge := exampleIntGaugeMetric()
	metrics := metricPairToMetrics(
		metricPair{
			attributes: attrsSum,
			metric:     metricSum,
		},
		metricPair{
			attributes: attrsGauge,
			metric:     metricGauge,
		},
	)

	err := test.exp.pushMetricsData(context.Background(), metrics)
	assert.NoError(t, err)
}

func TestAllMetricsFailed(t *testing.T) {
	testcases := []struct {
		name          string
		callbacks     []func(w http.ResponseWriter, req *http.Request)
		metricFunc    func() pmetric.Metrics
		expectedError string
	}{
		{
			name: "sent together when metrics under the same resource",
			callbacks: []func(w http.ResponseWriter, req *http.Request){
				func(w http.ResponseWriter, req *http.Request) {
					w.WriteHeader(500)

					body := extractBody(t, req)
					expected := `test.metric.data{test="test_value",test2="second_value"} 14500 1605534165000
gauge_metric_name{test="test_value",test2="second_value",remote_name="156920",url="http://example_url"} 124 1608124661166
gauge_metric_name{test="test_value",test2="second_value",remote_name="156955",url="http://another_url"} 245 1608124662166`
					assert.Equal(t, expected, body)
					assert.Equal(t, "application/vnd.sumologic.prometheus", req.Header.Get("Content-Type"))
				},
			},
			metricFunc: func() pmetric.Metrics {
				metricSum, attrs := exampleIntMetric()
				metricGauge, _ := exampleIntGaugeMetric()
				metrics := metricAndAttrsToPdataMetrics(
					attrs,
					metricSum, metricGauge,
				)
				return metrics
			},
			expectedError: "failed sending data: status: 500 Internal Server Error",
		},
		{
			name: "sent together when metrics under different resources",
			callbacks: []func(w http.ResponseWriter, req *http.Request){
				func(w http.ResponseWriter, req *http.Request) {
					w.WriteHeader(500)

					body := extractBody(t, req)
					expected := `test.metric.data{test="test_value",test2="second_value"} 14500 1605534165000
gauge_metric_name{foo="bar",remote_name="156920",url="http://example_url"} 124 1608124661166
gauge_metric_name{foo="bar",remote_name="156955",url="http://another_url"} 245 1608124662166`
					assert.Equal(t, expected, body)
					assert.Equal(t, "application/vnd.sumologic.prometheus", req.Header.Get("Content-Type"))
				},
			},
			metricFunc: func() pmetric.Metrics {
				metricSum, attrsSum := exampleIntMetric()
				metricGauge, attrsGauge := exampleIntGaugeMetric()
				metrics := metricPairToMetrics(
					metricPair{
						attributes: attrsSum,
						metric:     metricSum,
					},
					metricPair{
						attributes: attrsGauge,
						metric:     metricGauge,
					},
				)
				return metrics
			},
			expectedError: "failed sending data: status: 500 Internal Server Error",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			test := prepareExporterTest(t, createTestConfig(), tc.callbacks)
			test.exp.config.MetricFormat = PrometheusFormat

			metrics := tc.metricFunc()
			err := test.exp.pushMetricsData(context.Background(), metrics)

			assert.EqualError(t, err, tc.expectedError)

			var partial consumererror.Metrics
			require.True(t, errors.As(err, &partial))
			// TODO fix
			// assert.Equal(t, metrics, partial.GetMetrics())
		})
	}
}

func TestPushMetricsInvalidCompressor(t *testing.T) {
	metrics := metricAndAttributesToPdataMetrics(exampleIntMetric())

	// Expect no requests
	test := prepareExporterTest(t, createTestConfig(), nil)
	test.exp.config.CompressEncoding = "invalid"

	err := test.exp.pushMetricsData(context.Background(), metrics)
	assert.EqualError(t, err, "failed to initialize compressor: invalid format: invalid")
}

func TestLogsJsonFormatMetadataFilter(t *testing.T) {
	testcases := []struct {
		name                  string
		logResourceAttributes map[string]interface{}
		cfgFn                 func(c *Config)
		handler               func(w http.ResponseWriter, req *http.Request)
	}{
		{
			name: "basic",
			logResourceAttributes: map[string]interface{}{
				"_sourceCategory": "dummy",
				"key1":            "value1",
				"key2":            "value2",
			},
			cfgFn: func(c *Config) {
				c.LogFormat = JSONFormat
				c.SourceCategory = "testing_source_templates %{_sourceCategory}"
			},
			handler: func(w http.ResponseWriter, req *http.Request) {
				body := extractBody(t, req)
				regex := `{"log":"Example log","timestamp":\d{13}}`
				assert.Regexp(t, regex, body)

				assert.Equal(t, "key1=value1, key2=value2", req.Header.Get("X-Sumo-Fields"),
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
			logResourceAttributes: map[string]interface{}{
				"_sourceCategory": "dummy_category",
				"_sourceHost":     "dummy_host",
				"_sourceName":     "dummy_name",
				"key1":            "value1",
				"key2":            "value2",
			},
			cfgFn: func(c *Config) {
				c.LogFormat = JSONFormat
				c.SourceCategory = "testing_source_templates %{_sourceCategory}"
				c.SourceHost = "testing_source_templates %{_sourceHost}"
				c.SourceName = "testing_source_templates %{_sourceName}"
			},
			handler: func(w http.ResponseWriter, req *http.Request) {
				body := extractBody(t, req)
				regex := `{"log":"Example log","timestamp":\d{13}}`
				assert.Regexp(t, regex, body)

				assert.Equal(t, "key1=value1, key2=value2", req.Header.Get("X-Sumo-Fields"),
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
			logResourceAttributes: map[string]interface{}{
				"_sourceCategory": "cat",
				"key1":            "value1",
				"key2":            "value2",
			},
			cfgFn: func(c *Config) {
				c.LogFormat = JSONFormat
				c.SourceCategory = "dummy %{_sourceCategory}"
				c.SourceHost = "dummy %{_sourceHost}"
				c.SourceName = "dummy %{_sourceName}"
			},
			handler: func(w http.ResponseWriter, req *http.Request) {
				body := extractBody(t, req)
				regex := `{"log":"Example log","timestamp":\d{13}}`
				assert.Regexp(t, regex, body)

				assert.Equal(t, "key1=value1, key2=value2", req.Header.Get("X-Sumo-Fields"),
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
			logResourceAttributes: map[string]interface{}{
				"_sourceCategory": "dummy",
				"key1":            "value1",
				"key2":            "value2",
				"key3":            "",
			},
			cfgFn: func(c *Config) {
				c.LogFormat = JSONFormat
				c.SourceCategory = "testing_source_templates %{_sourceCategory}"
			},
			handler: func(w http.ResponseWriter, req *http.Request) {
				body := extractBody(t, req)

				var regex string
				regex += `{"log":"Example log","timestamp":\d{13}}`
				assert.Regexp(t, regex, body)

				assert.Equal(t, "key1=value1, key2=value2", req.Header.Get("X-Sumo-Fields"),
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
			pcommon.NewMapFromRaw(tc.logResourceAttributes).CopyTo(logResourceAttrs)

			err := test.exp.pushLogsData(context.Background(), logs)
			assert.NoError(t, err)
		})
	}
}

func TestLogsTextFormatMetadataFilterWithDroppedAttribute(t *testing.T) {
	test := prepareExporterTest(t, createTestConfig(), []func(w http.ResponseWriter, req *http.Request){
		func(w http.ResponseWriter, req *http.Request) {
			body := extractBody(t, req)
			assert.Equal(t, `Example log`, body)
			assert.Equal(t, "key2=value2", req.Header.Get("X-Sumo-Fields"))
		},
	})
	test.exp.config.LogFormat = TextFormat
	test.exp.config.DropRoutingAttribute = "key1"

	logs := LogRecordsToLogs(exampleLog())
	logs.ResourceLogs().At(0).Resource().Attributes().InsertString("key1", "value1")
	logs.ResourceLogs().At(0).Resource().Attributes().InsertString("key2", "value2")

	err := test.exp.pushLogsData(context.Background(), logs)
	assert.NoError(t, err)
}

func TestMetricsCarbon2(t *testing.T) {
	test := prepareExporterTest(t, createTestConfig(), []func(w http.ResponseWriter, req *http.Request){
		func(w http.ResponseWriter, req *http.Request) {
			body := extractBody(t, req)
			expected := `test=test_value test2=second_value key1=value1 key2=value2 metric=test.metric.data unit=bytes  14500 1605534165`
			assert.Equal(t, expected, body)
			assert.Equal(t, "application/vnd.sumologic.carbon2", req.Header.Get("Content-Type"))
		},
	})
	test.exp.config.MetricFormat = Carbon2Format

	metrics := metricAndAttributesToPdataMetrics(exampleIntMetric())

	attrs := metrics.ResourceMetrics().At(0).Resource().Attributes()
	attrs.InsertString("key1", "value1")
	attrs.InsertString("key2", "value2")

	err := test.exp.pushMetricsData(context.Background(), metrics)
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

	metrics := metricAndAttributesToPdataMetrics(exampleIntMetric())

	attrs := metrics.ResourceMetrics().At(0).Resource().Attributes()
	attrs.InsertString("key1", "value1")
	attrs.InsertString("key2", "value2")

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

	metrics := metricAndAttributesToPdataMetrics(exampleIntMetric())

	attrs := metrics.ResourceMetrics().At(0).Resource().Attributes()
	attrs.InsertString("key1", "value1")
	attrs.InsertString("key2", "value2")

	err := test.exp.pushMetricsData(context.Background(), metrics)
	assert.NoError(t, err)
}

func TestMetricsPrometheusWithDroppedRoutingAttribute(t *testing.T) {
	test := prepareExporterTest(t, createTestConfig(), []func(w http.ResponseWriter, req *http.Request){
		func(w http.ResponseWriter, req *http.Request) {
			body := extractBody(t, req)
			expected := `test.metric.data{test="test_value",test2="second_value",key1="value1",key2="value2"} 14500 1605534165000`
			assert.Equal(t, expected, body)
			assert.Equal(t, "application/vnd.sumologic.prometheus", req.Header.Get("Content-Type"))
		},
	})
	test.exp.config.MetricFormat = PrometheusFormat
	test.exp.config.DropRoutingAttribute = "http_listener_v2_path_custom"

	metrics := metricAndAttributesToPdataMetrics(exampleIntMetric())

	attrs := metrics.ResourceMetrics().At(0).Resource().Attributes()
	attrs.InsertString("key1", "value1")
	attrs.InsertString("key2", "value2")
	attrs.InsertString("http_listener_v2_path_custom", "prometheus.metrics")

	err := test.exp.pushMetricsData(context.Background(), metrics)
	assert.NoError(t, err)
}

func TestPushPrometheusMetrics_AttributeTranslation(t *testing.T) {
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
			name: "disabled",
			cfgFn: func() *Config {
				cfg := createConfig()
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
			metricSum, attrsSum := exampleIntMetric()
			attrsSum.InsertString("host.name", "harry-potter")
			attrsSum.InsertString("host.type", "wizard")

			metrics := metricPairToMetrics(
				metricPair{
					attributes: attrsSum,
					metric:     metricSum,
				},
			)

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

func TestPushOTLPMetrics_AttributeTranslation(t *testing.T) {
	createConfig := func() *Config {
		config := createDefaultConfig().(*Config)
		config.CompressEncoding = NoCompression
		config.MetricFormat = OTLPMetricFormat
		return config
	}

	testcases := []struct {
		name         string
		cfgFn        func() *Config
		expectedBody string
	}{
		{
			name: "enabled",
			cfgFn: func() *Config {
				cfg := createConfig()
				cfg.SourceCategory = "source_category_with_hostname_%{host.name}"
				cfg.SourceHost = "%{host}"
				// This is and should be done by default:
				// cfg.TranslateAttributes = true
				return cfg
			},
			//nolint:lll
			expectedBody: "\n\xf9\x01\n\xc1\x01\n\x14\n\x04test\x12\f\n\ntest_value\n\x17\n\x05test2\x12\x0e\n\fsecond_value\n\x16\n\x04host\x12\x0e\n\fharry-potter\n\x18\n\fInstanceType\x12\b\n\x06wizard\n\x1d\n\v_sourceHost\x12\x0e\n\fharry-potter\n?\n\x0f_sourceCategory\x12,\n*source_category_with_hostname_harry-potter\x123\n\x00\x12/\n\x10test.metric.data\x1a\x05bytes:\x14\n\x12\x19\x00\x12\x94\v\xd1\x00H\x161\xa48\x00\x00\x00\x00\x00\x00",
		},
		{
			name: "disabled",
			cfgFn: func() *Config {
				cfg := createConfig()
				cfg.SourceCategory = "source_category_with_hostname_%{host.name}"
				cfg.SourceHost = "%{host.name}"
				cfg.TranslateAttributes = false
				return cfg
			},
			//nolint:lll
			expectedBody: "\n\xfb\x01\n\xc3\x01\n\x14\n\x04test\x12\f\n\ntest_value\n\x17\n\x05test2\x12\x0e\n\fsecond_value\n\x1b\n\thost.name\x12\x0e\n\fharry-potter\n\x15\n\thost.type\x12\b\n\x06wizard\n\x1d\n\v_sourceHost\x12\x0e\n\fharry-potter\n?\n\x0f_sourceCategory\x12,\n*source_category_with_hostname_harry-potter\x123\n\x00\x12/\n\x10test.metric.data\x1a\x05bytes:\x14\n\x12\x19\x00\x12\x94\v\xd1\x00H\x161\xa48\x00\x00\x00\x00\x00\x00",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			metricSum, attrsSum := exampleIntMetric()
			attrsSum.InsertString("host.name", "harry-potter")
			attrsSum.InsertString("host.type", "wizard")

			metrics := metricPairToMetrics(
				metricPair{
					attributes: attrsSum,
					metric:     metricSum,
				},
			)

			config := tc.cfgFn()
			callbacks := []func(w http.ResponseWriter, req *http.Request){
				func(w http.ResponseWriter, req *http.Request) {
					assert.Equal(t, tc.expectedBody, extractBody(t, req))

					assert.Equal(t, "application/x-protobuf", req.Header.Get("Content-Type"))

					assert.Empty(t, req.Header.Get("X-Sumo-Fields"),
						"We should not get X-Sumo-Fields header when sending data with OTLP",
					)
					assert.Empty(t, req.Header.Get("X-Sumo-Category"),
						"We should not get X-Sumo-Category header when sending data with OTLP",
					)
					assert.Empty(t, req.Header.Get("X-Sumo-Name"),
						"We should not get X-Sumo-Name header when sending data with OTLP",
					)
					assert.Empty(t, req.Header.Get("X-Sumo-Host"),
						"We should not get X-Sumo-Host header when sending data with OTLP",
					)
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
		metricsFn    func() pmetric.Metrics
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
			metricsFn: func() pmetric.Metrics {
				metrics := pmetric.NewMetrics()
				{
					m := metrics.ResourceMetrics().AppendEmpty().
						ScopeMetrics().AppendEmpty().
						Metrics().AppendEmpty()
					m.SetName("cpu_usage_active")
					m.SetDataType(pmetric.MetricDataTypeGauge)
					dp := m.Gauge().DataPoints().AppendEmpty()
					dp.SetDoubleVal(123.456)
					dp.SetTimestamp(pcommon.NewTimestampFromTime(time.Unix(1605534165, 0)))
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
			metricsFn: func() pmetric.Metrics {
				metrics := pmetric.NewMetrics()
				scopeMetrics := metrics.ResourceMetrics().AppendEmpty().ScopeMetrics()
				{
					m := scopeMetrics.AppendEmpty().Metrics().AppendEmpty()
					m.SetName("cpu_usage_active")
					m.SetDataType(pmetric.MetricDataTypeGauge)
					dp := m.Gauge().DataPoints().AppendEmpty()
					dp.SetDoubleVal(123.456)
					dp.SetTimestamp(pcommon.NewTimestampFromTime(time.Unix(1605534165, 0)))
					dp.Attributes().InsertString("test", "test_value")
				}
				{
					m := scopeMetrics.AppendEmpty().Metrics().AppendEmpty()
					m.SetName("diskio_reads")
					m.SetDataType(pmetric.MetricDataTypeGauge)
					dp := m.Gauge().DataPoints().AppendEmpty()
					dp.SetIntVal(123456)
					dp.SetTimestamp(pcommon.NewTimestampFromTime(time.Unix(1605534165, 1000000)))
					dp.Attributes().InsertString("test", "test_value")
				}
				{
					m := scopeMetrics.AppendEmpty().Metrics().AppendEmpty()
					m.SetName("dummy_metric")
					m.SetDataType(pmetric.MetricDataTypeGauge)
					dp := m.Gauge().DataPoints().AppendEmpty()
					dp.SetIntVal(10)
					dp.SetTimestamp(pcommon.NewTimestampFromTime(time.Unix(1605534165, 2000000)))
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
			metricsFn: func() pmetric.Metrics {
				metrics := pmetric.NewMetrics()
				{
					m := metrics.ResourceMetrics().AppendEmpty().
						ScopeMetrics().AppendEmpty().
						Metrics().AppendEmpty()
					m.SetName("cpu_usage_active")
					m.SetDataType(pmetric.MetricDataTypeGauge)
					dp := m.Gauge().DataPoints().AppendEmpty()
					dp.SetDoubleVal(123.456)
					dp.SetTimestamp(pcommon.NewTimestampFromTime(time.Unix(1605534165, 0)))
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
			metricsFn: func() pmetric.Metrics {
				metrics := pmetric.NewMetrics()
				scopeMetrics := metrics.ResourceMetrics().AppendEmpty().ScopeMetrics()
				{
					m := scopeMetrics.AppendEmpty().Metrics().AppendEmpty()
					m.SetName("cpu_usage_active")
					m.SetDataType(pmetric.MetricDataTypeGauge)
					dp := m.Gauge().DataPoints().AppendEmpty()
					dp.SetDoubleVal(123.456)
					dp.SetTimestamp(pcommon.NewTimestampFromTime(time.Unix(1605534165, 0)))
					dp.Attributes().InsertString("test", "test_value")
				}
				{
					m := scopeMetrics.AppendEmpty().Metrics().AppendEmpty()
					m.SetName("diskio_reads")
					m.SetDataType(pmetric.MetricDataTypeGauge)
					dp := m.Gauge().DataPoints().AppendEmpty()
					dp.SetIntVal(123456)
					dp.SetTimestamp(pcommon.NewTimestampFromTime(time.Unix(1605534165, 1000000)))
					dp.Attributes().InsertString("test", "test_value")
				}
				{
					m := scopeMetrics.AppendEmpty().Metrics().AppendEmpty()
					m.SetName("dummy_metric")
					m.SetDataType(pmetric.MetricDataTypeGauge)
					dp := m.Gauge().DataPoints().AppendEmpty()
					dp.SetIntVal(10)
					dp.SetTimestamp(pcommon.NewTimestampFromTime(time.Unix(1605534165, 2000000)))
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

func TestTracesWithDroppedAttribute(t *testing.T) {
	// Prepare data to compare (trace without routing attribute)
	traces := exampleTrace()
	traces.ResourceSpans().At(0).Resource().Attributes().InsertString("key2", "value2")
	tracesMarshaler = otlp.NewProtobufTracesMarshaler()
	bytes, err := tracesMarshaler.MarshalTraces(traces)
	require.NoError(t, err)

	test := prepareExporterTest(t, createTestConfig(), []func(w http.ResponseWriter, req *http.Request){
		func(w http.ResponseWriter, req *http.Request) {
			body := extractBody(t, req)
			assert.Equal(t, string(bytes), body)
		},
	})
	test.exp.config.DropRoutingAttribute = "key1"

	// add routing attribute and check if after marshalling it's different
	traces.ResourceSpans().At(0).Resource().Attributes().InsertString("key1", "value1")
	bytesWithAttribute, err := tracesMarshaler.MarshalTraces(traces)
	require.NoError(t, err)
	require.NotEqual(t, bytes, bytesWithAttribute)

	err = test.exp.pushTracesData(context.Background(), traces)
	assert.NoError(t, err)
}

func Benchmark_ExporterPushLogs(b *testing.B) {
	createConfig := func() *Config {
		config := createDefaultConfig().(*Config)
		config.CompressEncoding = GZIPCompression
		config.MetricFormat = PrometheusFormat
		config.LogFormat = TextFormat
		config.SourceCategory = "testing_source_templates %{_sourceCategory}"
		config.HTTPClientSettings.Auth = nil
		return config
	}

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
	}))
	b.Cleanup(func() { testServer.Close() })

	cfg := createConfig()
	cfg.HTTPClientSettings.Endpoint = testServer.URL

	exp, err := initExporter(cfg, createExporterCreateSettings())
	require.NoError(b, err)
	require.NoError(b, exp.start(context.Background(), componenttest.NewNopHost()))
	defer func() {
		require.NoError(b, exp.shutdown(context.Background()))
	}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wg := sync.WaitGroup{}
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				err := exp.pushLogsData(context.Background(), LogRecordsToLogs(exampleNLogs(128)))
				if err != nil {
					b.Logf("Failed pushing logs: %v", err)
				}
				wg.Done()
			}()
		}

		wg.Wait()
	}
}
