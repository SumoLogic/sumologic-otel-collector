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
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/model/otlp"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type senderTest struct {
	reqCounter *int32
	srv        *httptest.Server
	s          *sender
}

// prepareSenderTest prepares sender test environment.
// Provided cfgOpts additionally configure the sender after the sendible default
// for tests have been applied.
// The enclosed httptest.Server is closed automatically using test.Cleanup.
func prepareSenderTest(t *testing.T, cb []func(w http.ResponseWriter, req *http.Request), cfgOpts ...func(*Config)) *senderTest {
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

	cfg := createDefaultConfig().(*Config)
	cfg.CompressEncoding = NoCompression
	cfg.HTTPClientSettings.Endpoint = testServer.URL
	cfg.LogFormat = TextFormat
	cfg.MetricFormat = Carbon2Format
	cfg.MaxRequestBodySize = 20_971_520
	for _, cfgOpt := range cfgOpts {
		cfgOpt(cfg)
	}

	c, err := newCompressor(cfg.CompressEncoding)
	require.NoError(t, err)

	pf, err := newPrometheusFormatter()
	require.NoError(t, err)

	gf, err := newGraphiteFormatter(cfg.GraphiteTemplate)
	require.NoError(t, err)

	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	return &senderTest{
		reqCounter: &reqCounter,
		srv:        testServer,
		s: newSender(
			logger,
			cfg,
			&http.Client{
				Timeout: cfg.HTTPClientSettings.Timeout,
			},
			sourceFormats{
				host:     getTestSourceFormat(t, "source_host"),
				category: getTestSourceFormat(t, "source_category"),
				name:     getTestSourceFormat(t, "source_name"),
			},
			c,
			pf,
			gf,
			"",
			"",
			"",
		),
	}
}

// prepareOTLPSenderTest prepares sender test environment.
// The enclosed httptest.Server is closed automatically using test.Cleanup.
func prepareOTLPSenderTest(t *testing.T, cb []func(w http.ResponseWriter, req *http.Request)) *senderTest {
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

	cfg := createDefaultConfig().(*Config)
	cfg.CompressEncoding = NoCompression
	cfg.HTTPClientSettings.Endpoint = testServer.URL

	c, err := newCompressor(cfg.CompressEncoding)
	require.NoError(t, err)

	pf, err := newPrometheusFormatter()
	require.NoError(t, err)

	gf, err := newGraphiteFormatter(cfg.GraphiteTemplate)
	require.NoError(t, err)

	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	return &senderTest{
		reqCounter: &reqCounter,
		srv:        testServer,
		s: newSender(
			logger,
			cfg,
			&http.Client{
				Timeout: cfg.HTTPClientSettings.Timeout,
			},
			sourceFormats{
				host:     getTestSourceFormat(t, "source_host"),
				category: getTestSourceFormat(t, "source_category"),
				name:     getTestSourceFormat(t, "source_name"),
			},
			c,
			pf,
			gf,
			testServer.URL,
			testServer.URL,
			testServer.URL,
		),
	}
}

func extractBody(t *testing.T, req *http.Request) string {
	buf := new(strings.Builder)
	_, err := io.Copy(buf, req.Body)
	require.NoError(t, err)
	return buf.String()
}

func exampleLog() []plog.LogRecord {
	buffer := make([]plog.LogRecord, 1)
	buffer[0] = plog.NewLogRecord()
	buffer[0].Body().SetStringVal("Example log")

	return buffer
}

func exampleNLogs(n int) []plog.LogRecord {
	buffer := make([]plog.LogRecord, n)
	for i := 0; i < n; i++ {
		buffer[i] = plog.NewLogRecord()
		buffer[i].Body().SetStringVal("Example log")
	}

	return buffer
}

func exampleTwoLogs() []plog.LogRecord {
	buffer := make([]plog.LogRecord, 2)
	buffer[0] = plog.NewLogRecord()
	buffer[0].Body().SetStringVal("Example log")
	buffer[0].Attributes().InsertString("key1", "value1")
	buffer[0].Attributes().InsertString("key2", "value2")
	buffer[1] = plog.NewLogRecord()
	buffer[1].Body().SetStringVal("Another example log")
	buffer[1].Attributes().InsertString("key1", "value1")
	buffer[1].Attributes().InsertString("key2", "value2")

	return buffer
}

func TestSendTrace(t *testing.T) {
	tracesMarshaler = otlp.NewProtobufTracesMarshaler()
	td := exampleTrace()
	traceBody, err := tracesMarshaler.MarshalTraces(td)
	assert.NoError(t, err)
	test := prepareSenderTest(t, []func(w http.ResponseWriter, req *http.Request){
		func(w http.ResponseWriter, req *http.Request) {
			body := extractBody(t, req)
			assert.Equal(t, string(traceBody), body)
			assert.Equal(t, "otelcol", req.Header.Get("X-Sumo-Client"))
			assert.Equal(t, "application/x-protobuf", req.Header.Get("Content-Type"))
		},
	})

	err = test.s.sendTraces(context.Background(), td)
	assert.NoError(t, err)
}

func TestSendLogs(t *testing.T) {
	test := prepareSenderTest(t, []func(w http.ResponseWriter, req *http.Request){
		func(w http.ResponseWriter, req *http.Request) {
			body := extractBody(t, req)
			assert.Equal(t, "Example log\nAnother example log", body)
			assert.Equal(t, "key1=value, key2=value2", req.Header.Get("X-Sumo-Fields"))
			assert.Equal(t, "otelcol", req.Header.Get("X-Sumo-Client"))
			assert.Equal(t, "application/x-www-form-urlencoded", req.Header.Get("Content-Type"))
		},
	})

	rls := plog.NewResourceLogs()
	slgs := rls.ScopeLogs()
	logsRecords1 := slgs.AppendEmpty().LogRecords()
	logsRecords1.AppendEmpty().Body().SetStringVal("Example log")
	logsRecords2 := slgs.AppendEmpty().LogRecords()
	logsRecords2.AppendEmpty().Body().SetStringVal("Another example log")

	_, err := test.s.sendNonOTLPLogs(context.Background(),
		rls,
		fieldsFromMap(map[string]string{"key1": "value", "key2": "value2"}),
	)
	assert.NoError(t, err)
	assert.EqualValues(t, 1, *test.reqCounter)
}

func TestSendLogsWithEmptyField(t *testing.T) {
	test := prepareSenderTest(t, []func(w http.ResponseWriter, req *http.Request){
		func(w http.ResponseWriter, req *http.Request) {
			body := extractBody(t, req)
			assert.Equal(t, "Example log\nAnother example log", body)
			assert.Equal(t, "key1=value, key2=value2", req.Header.Get("X-Sumo-Fields"))
			assert.Equal(t, "otelcol", req.Header.Get("X-Sumo-Client"))
			assert.Equal(t, "application/x-www-form-urlencoded", req.Header.Get("Content-Type"))
		},
	})

	rls := plog.NewResourceLogs()
	slgs := rls.ScopeLogs()
	logsRecords1 := slgs.AppendEmpty().LogRecords()
	logsRecords1.AppendEmpty().Body().SetStringVal("Example log")
	logsRecords2 := slgs.AppendEmpty().LogRecords()
	logsRecords2.AppendEmpty().Body().SetStringVal("Another example log")

	_, err := test.s.sendNonOTLPLogs(context.Background(),
		rls,
		fieldsFromMap(map[string]string{"key1": "value", "key2": "value2", "service": ""}),
	)
	assert.NoError(t, err)
	assert.EqualValues(t, 1, *test.reqCounter)
}

func TestSendLogsMultitype(t *testing.T) {
	test := prepareSenderTest(t, []func(w http.ResponseWriter, req *http.Request){
		func(w http.ResponseWriter, req *http.Request) {
			body := extractBody(t, req)
			expected := `{"lk1":"lv1","lk2":13}
["lv2",13]`
			assert.Equal(t, expected, body)
			assert.Equal(t, "key1=value, key2=value2", req.Header.Get("X-Sumo-Fields"))
			assert.Equal(t, "otelcol", req.Header.Get("X-Sumo-Client"))
			assert.Equal(t, "application/x-www-form-urlencoded", req.Header.Get("Content-Type"))
		},
	})

	rls := plog.NewResourceLogs()
	slgs := rls.ScopeLogs()
	logsRecords := slgs.AppendEmpty().LogRecords()
	attVal := pcommon.NewValueMap()
	attMap := attVal.MapVal()
	attMap.InsertString("lk1", "lv1")
	attMap.InsertInt("lk2", 13)
	logRecord := logsRecords.AppendEmpty()
	attVal.CopyTo(logRecord.Body())

	attVal = pcommon.NewValueSlice()
	attArr := attVal.SliceVal()
	strVal := pcommon.NewValueString("lv2")
	intVal := pcommon.NewValueInt(13)
	strVal.CopyTo(attArr.AppendEmpty())
	intVal.CopyTo(attArr.AppendEmpty())
	attVal.CopyTo(logsRecords.AppendEmpty().Body())

	_, err := test.s.sendNonOTLPLogs(context.Background(),
		rls,
		fieldsFromMap(map[string]string{"key1": "value", "key2": "value2"}),
	)
	assert.NoError(t, err)

	assert.EqualValues(t, 1, *test.reqCounter)
}

func TestSendLogsSplit(t *testing.T) {
	test := prepareSenderTest(t, []func(w http.ResponseWriter, req *http.Request){
		func(w http.ResponseWriter, req *http.Request) {
			body := extractBody(t, req)
			assert.Equal(t, "Example log", body)
		},
		func(w http.ResponseWriter, req *http.Request) {
			body := extractBody(t, req)
			assert.Equal(t, "Another example log", body)
		},
	})
	test.s.config.MaxRequestBodySize = 10

	rls := plog.NewResourceLogs()
	slgs := rls.ScopeLogs()
	logsRecords1 := slgs.AppendEmpty().LogRecords()
	logsRecords1.AppendEmpty().Body().SetStringVal("Example log")
	logsRecords2 := slgs.AppendEmpty().LogRecords()
	logsRecords2.AppendEmpty().Body().SetStringVal("Another example log")

	_, err := test.s.sendNonOTLPLogs(context.Background(),
		rls,
		fields{},
	)
	assert.NoError(t, err)

	assert.EqualValues(t, 2, *test.reqCounter)
}

func TestSendLogsSplitFailedOne(t *testing.T) {
	test := prepareSenderTest(t, []func(w http.ResponseWriter, req *http.Request){
		func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(500)
			_, err := fmt.Fprintf(
				w,
				`{"id":"1TIRY-KGIVX-TPQRJ","errors":[{"code":"internal.error","message":"Internal server error."}]}`,
			)

			require.NoError(t, err)

			body := extractBody(t, req)
			assert.Equal(t, "Example log", body)
		},
		func(w http.ResponseWriter, req *http.Request) {
			body := extractBody(t, req)
			assert.Equal(t, "Another example log", body)
		},
	})
	test.s.config.MaxRequestBodySize = 10
	test.s.config.LogFormat = TextFormat

	rls := plog.NewResourceLogs()
	slgs := rls.ScopeLogs()
	logsRecords1 := slgs.AppendEmpty().LogRecords()
	logsRecords1.AppendEmpty().Body().SetStringVal("Example log")
	logsRecords2 := slgs.AppendEmpty().LogRecords()
	logsRecords2.AppendEmpty().Body().SetStringVal("Another example log")

	dropped, err := test.s.sendNonOTLPLogs(context.Background(),
		rls,
		fields{},
	)
	assert.EqualError(t, err, "failed sending data: status: 500 Internal Server Error, id: 1TIRY-KGIVX-TPQRJ, errors: [{Code:internal.error Message:Internal server error.}]")
	assert.Len(t, dropped, 1)

	assert.EqualValues(t, 2, *test.reqCounter)
}

func TestSendLogsSplitFailedAll(t *testing.T) {
	test := prepareSenderTest(t, []func(w http.ResponseWriter, req *http.Request){
		func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(500)

			body := extractBody(t, req)
			assert.Equal(t, "Example log", body)
		},
		func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(404)

			body := extractBody(t, req)
			assert.Equal(t, "Another example log", body)
		},
	})
	test.s.config.MaxRequestBodySize = 10
	test.s.config.LogFormat = TextFormat

	rls := plog.NewResourceLogs()
	slgs := rls.ScopeLogs()
	logsRecords1 := slgs.AppendEmpty().LogRecords()
	logsRecords1.AppendEmpty().Body().SetStringVal("Example log")
	logsRecords2 := slgs.AppendEmpty().LogRecords()
	logsRecords2.AppendEmpty().Body().SetStringVal("Another example log")

	dropped, err := test.s.sendNonOTLPLogs(context.Background(), rls, fields{})
	assert.EqualError(
		t,
		err,
		"failed sending data: status: 500 Internal Server Error; failed sending data: status: 404 Not Found",
	)
	assert.Len(t, dropped, 2)

	assert.EqualValues(t, 2, *test.reqCounter)
}

func TestSendLogsJsonConfig(t *testing.T) {
	twoLogsFunc := func() plog.ResourceLogs {
		rls := plog.NewResourceLogs()
		slgs := rls.ScopeLogs().AppendEmpty()
		log := slgs.LogRecords().AppendEmpty()

		log.Body().SetStringVal("Example log")
		log.Attributes().InsertString("key1", "value1")
		log.Attributes().InsertString("key2", "value2")

		log = slgs.LogRecords().AppendEmpty()
		log.Body().SetStringVal("Another example log")
		log.Attributes().InsertString("key1", "value1")
		log.Attributes().InsertString("key2", "value2")

		return rls
	}

	twoComplexBodyLogsFunc := func() plog.ResourceLogs {
		rls := plog.NewResourceLogs()
		slgs := rls.ScopeLogs().AppendEmpty()
		log := slgs.LogRecords().AppendEmpty()

		body := pcommon.NewValueMap().MapVal()
		body.InsertString("a", "b")
		body.InsertBool("c", false)
		body.InsertInt("d", 20)
		body.InsertDouble("e", 20.5)

		f := pcommon.NewValueSlice()
		f.SliceVal().EnsureCapacity(4)
		f.SliceVal().AppendEmpty().SetStringVal("p")
		f.SliceVal().AppendEmpty().SetBoolVal(true)
		f.SliceVal().AppendEmpty().SetIntVal(13)
		f.SliceVal().AppendEmpty().SetDoubleVal(19.3)
		body.Insert("f", f)

		g := pcommon.NewValueMap()
		g.MapVal().InsertString("h", "i")
		g.MapVal().InsertBool("j", false)
		g.MapVal().InsertInt("k", 12)
		g.MapVal().InsertDouble("l", 11.1)

		body.Insert("g", g)

		log.Attributes().InsertString("m", "n")

		pcommon.NewValueMap().CopyTo(log.Body())
		body.CopyTo(log.Body().MapVal())

		return rls
	}

	testcases := []struct {
		name       string
		configOpts []func(*Config)
		bodyRegex  string
		logsFunc   func() plog.ResourceLogs
	}{
		{
			name: "default config",
			configOpts: []func(*Config){
				func(c *Config) {
					c.JSONLogs = JSONLogs{
						LogKey:       DefaultLogKey,
						AddTimestamp: DefaultAddTimestamp,
						TimestampKey: DefaultTimestampKey,
						FlattenBody:  DefaultFlattenBody,
					}
				},
			},
			bodyRegex: `{"key1":"value1","key2":"value2","log":"Example log","timestamp":\d{13}}` +
				`\n` +
				`{"key1":"value1","key2":"value2","log":"Another example log","timestamp":\d{13}}`,
			logsFunc: twoLogsFunc,
		},
		{
			name: "disabled add timestamp",
			configOpts: []func(*Config){
				func(c *Config) {
					c.JSONLogs = JSONLogs{
						LogKey:       DefaultLogKey,
						AddTimestamp: false,
					}
				},
			},
			bodyRegex: `{"key1":"value1","key2":"value2","log":"Example log"}` +
				`\n` +
				`{"key1":"value1","key2":"value2","log":"Another example log"}`,
			logsFunc: twoLogsFunc,
		},
		{
			name: "enabled add timestamp with custom timestamp key",
			configOpts: []func(*Config){
				func(c *Config) {
					c.JSONLogs = JSONLogs{
						LogKey:       DefaultLogKey,
						AddTimestamp: true,
						TimestampKey: "xxyy_zz",
					}
				},
			},
			bodyRegex: `{"key1":"value1","key2":"value2","log":"Example log","xxyy_zz":\d{13}}` +
				`\n` +
				`{"key1":"value1","key2":"value2","log":"Another example log","xxyy_zz":\d{13}}`,
			logsFunc: twoLogsFunc,
		},
		{
			name: "custom log key",
			configOpts: []func(*Config){
				func(c *Config) {
					c.JSONLogs = JSONLogs{
						LogKey:       "log_vendor_key",
						AddTimestamp: DefaultAddTimestamp,
						TimestampKey: DefaultTimestampKey,
						FlattenBody:  DefaultFlattenBody,
					}
				},
			},
			bodyRegex: `{"key1":"value1","key2":"value2","log_vendor_key":"Example log","timestamp":\d{13}}` +
				`\n` +
				`{"key1":"value1","key2":"value2","log_vendor_key":"Another example log","timestamp":\d{13}}`,
			logsFunc: twoLogsFunc,
		},
		{
			name: "flatten body",
			configOpts: []func(*Config){
				func(c *Config) {
					c.JSONLogs = JSONLogs{
						LogKey:       "log_vendor_key",
						AddTimestamp: DefaultAddTimestamp,
						TimestampKey: DefaultTimestampKey,
						FlattenBody:  true,
					}
				},
			},
			bodyRegex: `{"a":"b","c":false,"d":20,"e":20.5,"f":\["p",true,13,19.3\],` +
				`"g":{"h":"i","j":false,"k":12,"l":11.1},"m":"n","timestamp":\d{13}}`,
			logsFunc: twoComplexBodyLogsFunc,
		},
		{
			name: "complex body",
			configOpts: []func(*Config){
				func(c *Config) {
					c.JSONLogs = JSONLogs{
						LogKey:       "log_vendor_key",
						AddTimestamp: DefaultAddTimestamp,
						TimestampKey: DefaultTimestampKey,
						FlattenBody:  DefaultFlattenBody,
					}
				},
			},
			bodyRegex: `{"log_vendor_key":{"a":"b","c":false,"d":20,"e":20.5,"f":\["p",true,13,19.3\],` +
				`"g":{"h":"i","j":false,"k":12,"l":11.1}},"m":"n","timestamp":\d{13}}`,
			logsFunc: twoComplexBodyLogsFunc,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			test := prepareSenderTest(t, []func(w http.ResponseWriter, req *http.Request){
				func(w http.ResponseWriter, req *http.Request) {
					body := extractBody(t, req)
					assert.Regexp(t, tc.bodyRegex, body)
				},
			}, tc.configOpts...)

			test.s.config.LogFormat = JSONFormat

			_, err := test.s.sendNonOTLPLogs(context.Background(),
				tc.logsFunc(),
				fields{},
			)
			assert.NoError(t, err)

			assert.EqualValues(t, 1, *test.reqCounter)
		})
	}
}

func TestSendLogsJson(t *testing.T) {
	test := prepareSenderTest(t, []func(w http.ResponseWriter, req *http.Request){
		func(w http.ResponseWriter, req *http.Request) {
			body := extractBody(t, req)
			var regex string
			regex += `{"key1":"value1","key2":"value2","log":"Example log","timestamp":\d{13}}`
			regex += `\n`
			regex += `{"key1":"value1","key2":"value2","log":"Another example log","timestamp":\d{13}}`
			assert.Regexp(t, regex, body)

			assert.Equal(t, "key=value", req.Header.Get("X-Sumo-Fields"))
			assert.Equal(t, "otelcol", req.Header.Get("X-Sumo-Client"))
			assert.Equal(t, "application/x-www-form-urlencoded", req.Header.Get("Content-Type"))
		},
	})
	test.s.config.LogFormat = JSONFormat

	rls := plog.NewResourceLogs()
	slgs := rls.ScopeLogs().AppendEmpty()
	log := slgs.LogRecords().AppendEmpty()

	log.Body().SetStringVal("Example log")
	log.Attributes().InsertString("key1", "value1")
	log.Attributes().InsertString("key2", "value2")

	log = slgs.LogRecords().AppendEmpty()
	log.Body().SetStringVal("Another example log")
	log.Attributes().InsertString("key1", "value1")
	log.Attributes().InsertString("key2", "value2")

	_, err := test.s.sendNonOTLPLogs(context.Background(),
		rls,
		fieldsFromMap(map[string]string{"key": "value"}),
	)
	assert.NoError(t, err)

	assert.EqualValues(t, 1, *test.reqCounter)
}

func TestSendLogsJsonMultitype(t *testing.T) {
	test := prepareSenderTest(t, []func(w http.ResponseWriter, req *http.Request){
		func(w http.ResponseWriter, req *http.Request) {
			body := extractBody(t, req)
			var regex string
			regex += `{"key1":"value1","key2":"value2","log":{"lk1":"lv1","lk2":13},"timestamp":\d{13}}`
			regex += `\n`
			regex += `{"key1":"value1","key2":"value2","log":\["lv2",13\],"timestamp":\d{13}}`
			assert.Regexp(t, regex, body)

			assert.Equal(t, "key=value", req.Header.Get("X-Sumo-Fields"))
			assert.Equal(t, "otelcol", req.Header.Get("X-Sumo-Client"))
			assert.Equal(t, "application/x-www-form-urlencoded", req.Header.Get("Content-Type"))
		},
	})
	test.s.config.LogFormat = JSONFormat

	rls := plog.NewResourceLogs()
	slgs := rls.ScopeLogs().AppendEmpty()

	attVal := pcommon.NewValueMap()
	attMap := attVal.MapVal()
	attMap.InsertString("lk1", "lv1")
	attMap.InsertInt("lk2", 13)

	log := slgs.LogRecords().AppendEmpty()
	attVal.CopyTo(log.Body())

	log.Attributes().InsertString("key1", "value1")
	log.Attributes().InsertString("key2", "value2")

	log = slgs.LogRecords().AppendEmpty()

	attVal = pcommon.NewValueSlice()
	attArr := attVal.SliceVal()
	strVal := pcommon.NewValueString("lv2")
	intVal := pcommon.NewValueInt(13)

	strVal.CopyTo(attArr.AppendEmpty())
	intVal.CopyTo(attArr.AppendEmpty())

	attVal.CopyTo(log.Body())
	log.Attributes().InsertString("key1", "value1")
	log.Attributes().InsertString("key2", "value2")

	_, err := test.s.sendNonOTLPLogs(context.Background(),
		rls,
		fieldsFromMap(map[string]string{"key": "value"}),
	)
	assert.NoError(t, err)

	assert.EqualValues(t, 1, *test.reqCounter)
}

func TestSendLogsJsonSplit(t *testing.T) {
	test := prepareSenderTest(t, []func(w http.ResponseWriter, req *http.Request){
		func(w http.ResponseWriter, req *http.Request) {
			body := extractBody(t, req)
			var regex string
			regex += `{"key1":"value1","key2":"value2","log":"Example log","timestamp":\d{13}}`
			assert.Regexp(t, regex, body)
		},
		func(w http.ResponseWriter, req *http.Request) {
			body := extractBody(t, req)
			var regex string
			regex += `{"key1":"value1","key2":"value2","log":"Another example log","timestamp":\d{13}}`
			assert.Regexp(t, regex, body)
		},
	})
	test.s.config.LogFormat = JSONFormat
	test.s.config.MaxRequestBodySize = 10

	rls := plog.NewResourceLogs()
	slgs := rls.ScopeLogs().AppendEmpty()
	log := slgs.LogRecords().AppendEmpty()

	log.Body().SetStringVal("Example log")
	log.Attributes().InsertString("key1", "value1")
	log.Attributes().InsertString("key2", "value2")

	log = slgs.LogRecords().AppendEmpty()
	log.Body().SetStringVal("Another example log")
	log.Attributes().InsertString("key1", "value1")
	log.Attributes().InsertString("key2", "value2")

	_, err := test.s.sendNonOTLPLogs(context.Background(),
		rls,
		fieldsFromMap(map[string]string{"key": "value"}),
	)
	assert.NoError(t, err)

	assert.EqualValues(t, 2, *test.reqCounter)
}

func TestSendLogsJsonSplitFailedOne(t *testing.T) {
	test := prepareSenderTest(t, []func(w http.ResponseWriter, req *http.Request){
		func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(500)

			body := extractBody(t, req)

			var regex string
			regex += `{"key1":"value1","key2":"value2","log":"Example log","timestamp":\d{13}}`
			assert.Regexp(t, regex, body)
		},
		func(w http.ResponseWriter, req *http.Request) {
			body := extractBody(t, req)

			var regex string
			regex += `{"key1":"value1","key2":"value2","log":"Another example log","timestamp":\d{13}}`
			assert.Regexp(t, regex, body)
		},
	})
	test.s.config.LogFormat = JSONFormat
	test.s.config.MaxRequestBodySize = 10

	rls := plog.NewResourceLogs()
	slgs := rls.ScopeLogs().AppendEmpty()
	log := slgs.LogRecords().AppendEmpty()

	log.Body().SetStringVal("Example log")
	log.Attributes().InsertString("key1", "value1")
	log.Attributes().InsertString("key2", "value2")

	log = slgs.LogRecords().AppendEmpty()
	log.Body().SetStringVal("Another example log")
	log.Attributes().InsertString("key1", "value1")
	log.Attributes().InsertString("key2", "value2")

	dropped, err := test.s.sendNonOTLPLogs(context.Background(),
		rls,
		fieldsFromMap(map[string]string{"key": "value"}),
	)
	assert.EqualError(t, err, "failed sending data: status: 500 Internal Server Error")
	assert.Len(t, dropped, 1)

	assert.EqualValues(t, 2, *test.reqCounter)
}

func TestSendLogsJsonSplitFailedAll(t *testing.T) {
	test := prepareSenderTest(t, []func(w http.ResponseWriter, req *http.Request){
		func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(500)

			body := extractBody(t, req)

			var regex string
			regex += `{"key1":"value1","key2":"value2","log":"Example log","timestamp":\d{13}}`
			assert.Regexp(t, regex, body)
		},
		func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(404)

			body := extractBody(t, req)

			var regex string
			regex += `{"key1":"value1","key2":"value2","log":"Another example log","timestamp":\d{13}}`
			assert.Regexp(t, regex, body)
		},
	})
	test.s.config.LogFormat = JSONFormat
	test.s.config.MaxRequestBodySize = 10

	rls := plog.NewResourceLogs()
	slgs := rls.ScopeLogs().AppendEmpty()
	log := slgs.LogRecords().AppendEmpty()

	log.Body().SetStringVal("Example log")
	log.Attributes().InsertString("key1", "value1")
	log.Attributes().InsertString("key2", "value2")

	log = slgs.LogRecords().AppendEmpty()
	log.Body().SetStringVal("Another example log")
	log.Attributes().InsertString("key1", "value1")
	log.Attributes().InsertString("key2", "value2")

	dropped, err := test.s.sendNonOTLPLogs(context.Background(),
		rls,
		fields{},
	)

	assert.EqualError(
		t,
		err,
		"failed sending data: status: 500 Internal Server Error; failed sending data: status: 404 Not Found",
	)
	assert.Len(t, dropped, 2)

	assert.EqualValues(t, 2, *test.reqCounter)
}

func TestSendLogsUnexpectedFormat(t *testing.T) {
	test := prepareSenderTest(t, []func(w http.ResponseWriter, req *http.Request){
		func(w http.ResponseWriter, req *http.Request) {
		},
	})
	test.s.config.LogFormat = "dummy"

	rls := plog.NewResourceLogs()
	slgs := rls.ScopeLogs().AppendEmpty()
	log := slgs.LogRecords().AppendEmpty()
	log.Body().SetStringVal("Example log")

	dropped, err := test.s.sendNonOTLPLogs(context.Background(),
		rls,
		fields{},
	)
	assert.Error(t, err)
	assert.Len(t, dropped, 1)
	assert.Equal(t, []plog.LogRecord{log}, dropped)
}

func TestSendLogsOTLP(t *testing.T) {
	test := prepareSenderTest(t, []func(w http.ResponseWriter, req *http.Request){
		func(w http.ResponseWriter, req *http.Request) {
			body := extractBody(t, req)
			//nolint:lll
			assert.Equal(t, "\n\xe6\x01\nb\n\x1c\n\v_sourceHost\x12\r\n\vsource_host\n\x1c\n\v_sourceName\x12\r\n\vsource_name\n$\n\x0f_sourceCategory\x12\x11\n\x0fsource_category\x12;\n\x00\x127*\r\n\vExample log2\x10\n\x04key1\x12\b\n\x06value12\x10\n\x04key2\x12\b\n\x06value2J\x00R\x00\x12C\n\x00\x12?*\x15\n\x13Another example log2\x10\n\x04key1\x12\b\n\x06value12\x10\n\x04key2\x12\b\n\x06value2J\x00R\x00", body)

			assert.Equal(t, "otelcol", req.Header.Get("X-Sumo-Client"))
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
	})

	test.s.config.LogFormat = "otlp"

	l := plog.NewLogs()
	ls := l.ResourceLogs().AppendEmpty()

	logRecords := exampleTwoLogs()
	for i := 0; i < len(logRecords); i++ {
		logRecords[i].MoveTo(ls.ScopeLogs().AppendEmpty().LogRecords().AppendEmpty())
	}

	assert.NoError(t, test.s.sendOTLPLogs(context.Background(), l))
	assert.EqualValues(t, 1, *test.reqCounter)
}

func TestOverrideSourceName(t *testing.T) {
	twoLogsFunc := func() plog.ResourceLogs {
		rls := plog.NewResourceLogs()
		slgs := rls.ScopeLogs().AppendEmpty()
		log := slgs.LogRecords().AppendEmpty()

		log.Body().SetStringVal("Example log")
		log.Attributes().InsertString("key1", "value1")
		log.Attributes().InsertString("key2", "value2")

		log = slgs.LogRecords().AppendEmpty()
		log.Body().SetStringVal("Another example log")
		log.Attributes().InsertString("key1", "value1")
		log.Attributes().InsertString("key2", "value2")

		return rls
	}

	t.Run("text format", func(t *testing.T) {
		test := prepareSenderTest(t, []func(w http.ResponseWriter, req *http.Request){
			func(w http.ResponseWriter, req *http.Request) {
				assert.Equal(t, "Test source name/test_name", req.Header.Get("X-Sumo-Name"))
			},
		})

		test.s.sources.name = getTestSourceFormat(t, "Test source name/%{key1}")

		_, err := test.s.sendNonOTLPLogs(context.Background(),
			twoLogsFunc(),
			fieldsFromMap(map[string]string{"key1": "test_name"}),
		)
		assert.NoError(t, err)

		assert.EqualValues(t, 1, *test.reqCounter)
	})

	t.Run("json format", func(t *testing.T) {
		test := prepareSenderTest(t, []func(w http.ResponseWriter, req *http.Request){
			func(w http.ResponseWriter, req *http.Request) {
				assert.Equal(t, "Test source name/test_name", req.Header.Get("X-Sumo-Name"))
			},
		}, func(c *Config) {
			c.LogFormat = JSONFormat
		})

		test.s.sources.name = getTestSourceFormat(t, "Test source name/%{key1}")

		_, err := test.s.sendNonOTLPLogs(context.Background(),
			twoLogsFunc(),
			fieldsFromMap(map[string]string{"key1": "test_name"}),
		)
		assert.NoError(t, err)

		assert.EqualValues(t, 1, *test.reqCounter)
	})

	t.Run("otlp", func(t *testing.T) {
		test := prepareOTLPSenderTest(t, []func(w http.ResponseWriter, req *http.Request){
			func(w http.ResponseWriter, req *http.Request) {
				unmarshaller := otlp.NewProtobufLogsUnmarshaler()
				b, err := io.ReadAll(req.Body)
				require.NoError(t, err)
				l, err := unmarshaller.UnmarshalLogs(b)
				require.NoError(t, err)

				require.Equal(t, l.ResourceLogs().Len(), 1)
				sourceCategory, ok := l.ResourceLogs().At(0).Resource().Attributes().Get("_sourceName")
				require.True(t, ok)
				require.Equal(t, pcommon.ValueTypeString, sourceCategory.Type())
				require.Equal(t, "Test source name/test_name", sourceCategory.StringVal())
			},
		})

		test.s.sources.name = getTestSourceFormat(t, "Test source name/%{key1}")

		l := plog.NewLogs()
		ls := l.ResourceLogs().AppendEmpty()
		ls.Resource().Attributes().InsertString("key1", "test_name")
		logRecords := exampleTwoLogs()
		for i := 0; i < len(logRecords); i++ {
			logRecords[i].MoveTo(ls.ScopeLogs().AppendEmpty().LogRecords().AppendEmpty())
		}
		assert.NoError(t, test.s.sendOTLPLogs(context.Background(), l))
	})
}

func TestOverrideSourceCategory(t *testing.T) {
	twoLogsFunc := func() plog.ResourceLogs {
		rls := plog.NewResourceLogs()
		slgs := rls.ScopeLogs().AppendEmpty()
		log := slgs.LogRecords().AppendEmpty()

		log.Body().SetStringVal("Example log")
		log.Attributes().InsertString("key1", "value1")
		log.Attributes().InsertString("key2", "value2")

		log = slgs.LogRecords().AppendEmpty()
		log.Body().SetStringVal("Another example log")
		log.Attributes().InsertString("key1", "value1")
		log.Attributes().InsertString("key2", "value2")

		return rls
	}

	t.Run("text format", func(t *testing.T) {
		test := prepareSenderTest(t, []func(w http.ResponseWriter, req *http.Request){
			func(w http.ResponseWriter, req *http.Request) {
				assert.Equal(t, "Test source category/test_name", req.Header.Get("X-Sumo-Category"))
			},
		})

		test.s.sources.category = getTestSourceFormat(t, "Test source category/%{key1}")

		_, err := test.s.sendNonOTLPLogs(context.Background(),
			twoLogsFunc(),
			fieldsFromMap(map[string]string{"key1": "test_name"}),
		)
		assert.NoError(t, err)
	})

	t.Run("json format", func(t *testing.T) {
		test := prepareSenderTest(t, []func(w http.ResponseWriter, req *http.Request){
			func(w http.ResponseWriter, req *http.Request) {
				assert.Equal(t, "Test source category/test_name", req.Header.Get("X-Sumo-Category"))
			},
		}, func(c *Config) {
			c.LogFormat = JSONFormat
		})

		test.s.sources.category = getTestSourceFormat(t, "Test source category/%{key1}")
		_, err := test.s.sendNonOTLPLogs(context.Background(),
			twoLogsFunc(),
			fieldsFromMap(map[string]string{"key1": "test_name"}),
		)
		assert.NoError(t, err)

		assert.EqualValues(t, 1, *test.reqCounter)
	})

	t.Run("otlp", func(t *testing.T) {
		test := prepareOTLPSenderTest(t, []func(w http.ResponseWriter, req *http.Request){
			func(w http.ResponseWriter, req *http.Request) {
				unmarshaller := otlp.NewProtobufLogsUnmarshaler()
				b, err := io.ReadAll(req.Body)
				require.NoError(t, err)
				l, err := unmarshaller.UnmarshalLogs(b)
				require.NoError(t, err)

				require.Equal(t, l.ResourceLogs().Len(), 1)
				sourceCategory, ok := l.ResourceLogs().At(0).Resource().Attributes().Get("_sourceCategory")
				require.True(t, ok)
				require.Equal(t, pcommon.ValueTypeString, sourceCategory.Type())
				require.Equal(t, "Test source category/test_name", sourceCategory.StringVal())
			},
		})

		test.s.sources.category = getTestSourceFormat(t, "Test source category/%{key1}")

		l := plog.NewLogs()
		ls := l.ResourceLogs().AppendEmpty()
		ls.Resource().Attributes().InsertString("key1", "test_name")
		logRecords := exampleTwoLogs()
		for i := 0; i < len(logRecords); i++ {
			logRecords[i].MoveTo(ls.ScopeLogs().AppendEmpty().LogRecords().AppendEmpty())
		}
		assert.NoError(t, test.s.sendOTLPLogs(context.Background(), l))
	})
}

func TestOverrideSourceHost(t *testing.T) {
	twoLogsFunc := func() plog.ResourceLogs {
		rls := plog.NewResourceLogs()
		slgs := rls.ScopeLogs().AppendEmpty()
		log := slgs.LogRecords().AppendEmpty()

		log.Body().SetStringVal("Example log")
		log.Attributes().InsertString("key1", "value1")
		log.Attributes().InsertString("key2", "value2")

		log = slgs.LogRecords().AppendEmpty()
		log.Body().SetStringVal("Another example log")
		log.Attributes().InsertString("key1", "value1")
		log.Attributes().InsertString("key2", "value2")

		return rls
	}

	t.Run("text format", func(t *testing.T) {
		test := prepareSenderTest(t, []func(w http.ResponseWriter, req *http.Request){
			func(w http.ResponseWriter, req *http.Request) {
				assert.Equal(t, "Test source host/test_name", req.Header.Get("X-Sumo-Host"))
			},
		})

		test.s.sources.host = getTestSourceFormat(t, "Test source host/%{key1}")
		_, err := test.s.sendNonOTLPLogs(context.Background(),
			twoLogsFunc(),
			fieldsFromMap(map[string]string{"key1": "test_name"}),
		)
		assert.NoError(t, err)
	})

	t.Run("json format", func(t *testing.T) {
		test := prepareSenderTest(t, []func(w http.ResponseWriter, req *http.Request){
			func(w http.ResponseWriter, req *http.Request) {
				assert.Equal(t, "Test source host/test_name", req.Header.Get("X-Sumo-Host"))
			},
		}, func(c *Config) {
			c.LogFormat = JSONFormat
		})

		test.s.sources.host = getTestSourceFormat(t, "Test source host/%{key1}")

		_, err := test.s.sendNonOTLPLogs(context.Background(),
			twoLogsFunc(),
			fieldsFromMap(map[string]string{"key1": "test_name"}),
		)
		assert.NoError(t, err)

		assert.EqualValues(t, 1, *test.reqCounter)
	})

	t.Run("otlp", func(t *testing.T) {
		test := prepareOTLPSenderTest(t, []func(w http.ResponseWriter, req *http.Request){
			func(w http.ResponseWriter, req *http.Request) {
				unmarshaller := otlp.NewProtobufLogsUnmarshaler()
				b, err := io.ReadAll(req.Body)
				require.NoError(t, err)
				l, err := unmarshaller.UnmarshalLogs(b)
				require.NoError(t, err)

				require.Equal(t, l.ResourceLogs().Len(), 1)
				sourceHost, ok := l.ResourceLogs().At(0).Resource().Attributes().Get("_sourceHost")
				require.True(t, ok)
				require.Equal(t, pcommon.ValueTypeString, sourceHost.Type())
				require.Equal(t, "Test source host/test_name", sourceHost.StringVal())
			},
		})

		test.s.sources.host = getTestSourceFormat(t, "Test source host/%{key1}")

		l := plog.NewLogs()
		ls := l.ResourceLogs().AppendEmpty()
		ls.Resource().Attributes().InsertString("key1", "test_name")
		logRecords := exampleTwoLogs()
		for i := 0; i < len(logRecords); i++ {
			logRecords[i].MoveTo(ls.ScopeLogs().AppendEmpty().LogRecords().AppendEmpty())
		}
		assert.NoError(t, test.s.sendOTLPLogs(context.Background(), l))
	})
}

func TestLogsDontSendSourceFieldsInXSumoFieldsHeader(t *testing.T) {
	twoLogsFunc := func() plog.ResourceLogs {
		rls := plog.NewResourceLogs()
		slgs := rls.ScopeLogs().AppendEmpty()
		log := slgs.LogRecords().AppendEmpty()

		log.Body().SetStringVal("Example log")
		log.Attributes().InsertString("key1", "value1")
		log.Attributes().InsertString("key2", "value2")

		log = slgs.LogRecords().AppendEmpty()
		log.Body().SetStringVal("Another example log")
		log.Attributes().InsertString("key1", "value1")
		log.Attributes().InsertString("key2", "value2")

		return rls
	}

	assertNoSourceFieldsInXSumoFields := func(t *testing.T, fieldsHeader string) {
		for _, field := range strings.Split(fieldsHeader, ",") {
			field = strings.TrimSpace(field)
			split := strings.Split(field, "=")
			require.Len(t, split, 2)

			switch fieldName := split[0]; fieldName {
			case "_sourceName":
				assert.Failf(t, "X-Sumo-Fields header check",
					"%s should be removed from X-Sumo-Fields header when X-Sumo-Name is set", fieldName)
			case "_sourceHost":
				assert.Failf(t, "X-Sumo-Fields header check",
					"%s should be removed from X-Sumo-Fields header when X-Sumo-Host is set", fieldName)
			case "_sourceCategory":
				assert.Failf(t, "X-Sumo-Fields header check",
					"%s should be removed from X-Sumo-Fields header when X-Sumo-Category is set", fieldName)
			default:
			}

			t.Logf("field: %s", field)
		}
	}

	t.Run("json", func(t *testing.T) {
		test := prepareSenderTest(t, []func(w http.ResponseWriter, req *http.Request){
			func(w http.ResponseWriter, req *http.Request) {
				assert.Equal(t,
					"Test source name/key1_val/test_source_name",
					req.Header.Get("X-Sumo-Name"),
				)
				assert.Equal(t, "Test source host/key1_val",
					req.Header.Get("X-Sumo-Host"),
				)
				assert.Equal(t, "Test source category/key1_val",
					req.Header.Get("X-Sumo-Category"),
				)

				body := extractBody(t, req)
				t.Logf("body: %s", body)

				assertNoSourceFieldsInXSumoFields(t, req.Header.Get("X-Sumo-Fields"))
			},
		}, func(c *Config) {
			c.LogFormat = JSONFormat
		})

		test.s.sources.name = getTestSourceFormat(t, "Test source name/%{key1}/%{_sourceName}")
		test.s.sources.host = getTestSourceFormat(t, "Test source host/%{key1}")
		test.s.sources.category = getTestSourceFormat(t, "Test source category/%{key1}")

		_, err := test.s.sendNonOTLPLogs(context.Background(),
			twoLogsFunc(),
			fieldsFromMap(
				map[string]string{
					"key1":            "key1_val",
					"_sourceName":     "test_source_name",
					"_sourceHost":     "test_source_host",
					"_sourceCategory": "test_source_category",
				}),
		)
		assert.NoError(t, err)
		assert.EqualValues(t, 1, *test.reqCounter)
	})
}

func TestLogsHandlesReceiverResponses(t *testing.T) {
	t.Run("json with too many fields logs a warning", func(t *testing.T) {
		test := prepareSenderTest(t, []func(w http.ResponseWriter, req *http.Request){
			func(w http.ResponseWriter, req *http.Request) {
				fmt.Fprintf(w, `{
					"status" : 200,
					"id" : "YBLR1-S2T29-MVXEJ",
					"code" : "bad.http.header.fields",
					"message" : "X-Sumo-Fields Warning: 14 key-value pairs are dropped as they are exceeding maximum key-value pair number limit 30."
				  }`)
			},
		}, func(c *Config) {
			c.LogFormat = JSONFormat
		})

		rls := plog.NewResourceLogs()
		rls.ScopeLogs().AppendEmpty().LogRecords().AppendEmpty().Body().SetStringVal("Example log")

		var buffer bytes.Buffer
		writer := bufio.NewWriter(&buffer)
		test.s.logger = zap.New(
			zapcore.NewCore(
				zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig()),
				zapcore.AddSync(writer),
				zapcore.DebugLevel,
			),
		)

		_, err := test.s.sendNonOTLPLogs(context.Background(),
			rls,
			fieldsFromMap(
				map[string]string{
					"cluster":         "abcaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					"code":            "4222222222222222222222222222222222222222222222222222222222222222222222222222222222222",
					"component":       "apiserver",
					"endpoint":        "httpsaaaaaaaaaaaaaaaaaaa",
					"a":               "a",
					"b":               "b",
					"c":               "c",
					"d":               "d",
					"e":               "e",
					"f":               "f",
					"g":               "g",
					"q":               "q",
					"w":               "w",
					"r":               "r",
					"t":               "t",
					"y":               "y",
					"1":               "1",
					"2":               "2",
					"3":               "3",
					"4":               "4",
					"5":               "5",
					"6":               "6",
					"7":               "7",
					"8":               "8",
					"9":               "9",
					"10":              "10",
					"11":              "11",
					"12":              "12",
					"13":              "13",
					"14":              "14",
					"15":              "15",
					"16":              "16",
					"17":              "17",
					"18":              "18",
					"19":              "19",
					"20":              "20",
					"21":              "21",
					"22":              "22",
					"23":              "23",
					"24":              "24",
					"25":              "25",
					"26":              "26",
					"27":              "27",
					"28":              "28",
					"29":              "29",
					"_sourceName":     "test_source_name",
					"_sourceHost":     "test_source_host",
					"_sourceCategory": "test_source_category",
				}),
		)
		assert.NoError(t, writer.Flush())
		assert.NoError(t, err)
		assert.EqualValues(t, 1, *test.reqCounter)

		assert.Contains(t,
			buffer.String(),
			`There was an issue sending data	{`+
				`"status": "200 OK", `+
				`"id": "YBLR1-S2T29-MVXEJ", `+
				`"code": "bad.http.header.fields", `+
				`"message": "X-Sumo-Fields Warning: 14 key-value pairs are dropped as they are exceeding maximum key-value pair number limit 30."`,
		)
	})
}

func TestInvalidEndpoint(t *testing.T) {
	test := prepareSenderTest(t, []func(w http.ResponseWriter, req *http.Request){})

	test.s.config.HTTPClientSettings.Endpoint = ":"

	rls := plog.NewResourceLogs()
	rls.ScopeLogs().AppendEmpty().LogRecords().AppendEmpty().Body().SetStringVal("Example log")

	_, err := test.s.sendNonOTLPLogs(context.Background(), rls, fields{})
	assert.EqualError(t, err, `parse ":": missing protocol scheme`)
}

func TestInvalidPostRequest(t *testing.T) {
	test := prepareSenderTest(t, []func(w http.ResponseWriter, req *http.Request){})

	test.s.config.HTTPClientSettings.Endpoint = ""
	rls := plog.NewResourceLogs()
	rls.ScopeLogs().AppendEmpty().LogRecords().AppendEmpty().Body().SetStringVal("Example log")

	_, err := test.s.sendNonOTLPLogs(context.Background(), rls, fields{})
	assert.EqualError(t, err, `Post "": unsupported protocol scheme ""`)
}

func TestInvalidMetricFormat(t *testing.T) {
	test := prepareSenderTest(t, []func(w http.ResponseWriter, req *http.Request){})

	test.s.config.MetricFormat = "invalid"

	err := test.s.send(context.Background(), MetricsPipeline, newCountingReader(0).withString(""), fields{})
	assert.EqualError(t, err, `unsupported metrics format: invalid`)
}

func TestInvalidPipeline(t *testing.T) {
	test := prepareSenderTest(t, []func(w http.ResponseWriter, req *http.Request){})

	err := test.s.send(context.Background(), "invalidPipeline", newCountingReader(0).withString(""), fields{})
	assert.EqualError(t, err, `unexpected pipeline: invalidPipeline`)
}

func TestSendCompressGzip(t *testing.T) {
	test := prepareSenderTest(t, []func(res http.ResponseWriter, req *http.Request){
		func(res http.ResponseWriter, req *http.Request) {
			res.WriteHeader(200)
			if _, err := res.Write([]byte("")); err != nil {
				res.WriteHeader(http.StatusInternalServerError)
				assert.FailNow(t, "err: %v", err)
				return
			}
			body := decodeGzip(t, req.Body)
			assert.Equal(t, "gzip", req.Header.Get("Content-Encoding"))
			assert.Equal(t, "Some example log", body)
		},
	})

	test.s.config.CompressEncoding = "gzip"

	c, err := newCompressor("gzip")
	require.NoError(t, err)

	test.s.compressor = c
	reader := newCountingReader(0).withString("Some example log")

	err = test.s.send(context.Background(), LogsPipeline, reader, fields{})
	require.NoError(t, err)
}

func TestSendCompressDeflate(t *testing.T) {
	test := prepareSenderTest(t, []func(res http.ResponseWriter, req *http.Request){
		func(res http.ResponseWriter, req *http.Request) {
			res.WriteHeader(200)

			if _, err := res.Write([]byte("")); err != nil {
				res.WriteHeader(http.StatusInternalServerError)
				assert.FailNow(t, "err: %v", err)
				return
			}
			body := decodeDeflate(t, req.Body)
			assert.Equal(t, "deflate", req.Header.Get("Content-Encoding"))
			assert.Equal(t, "Some example log", body)
		},
	})

	test.s.config.CompressEncoding = "deflate"

	c, err := newCompressor("deflate")
	require.NoError(t, err)

	test.s.compressor = c
	reader := newCountingReader(0).withString("Some example log")

	err = test.s.send(context.Background(), LogsPipeline, reader, fields{})
	require.NoError(t, err)
}

func TestCompressionError(t *testing.T) {
	test := prepareSenderTest(t, []func(w http.ResponseWriter, req *http.Request){})

	test.s.compressor = getTestCompressor(errors.New("read error"), nil)
	reader := newCountingReader(0).withString("Some example log")

	err := test.s.send(context.Background(), LogsPipeline, reader, fields{})
	assert.EqualError(t, err, "read error")
}

func TestInvalidContentEncoding(t *testing.T) {
	// Expect to requests
	test := prepareSenderTest(t, nil)

	test.s.config.CompressEncoding = "test"
	reader := newCountingReader(0).withString("Some example log")

	err := test.s.send(context.Background(), LogsPipeline, reader, fields{})
	assert.EqualError(t, err, "invalid content encoding: test")
}

func TestSendMetrics(t *testing.T) {
	test := prepareSenderTest(t, []func(w http.ResponseWriter, req *http.Request){
		func(w http.ResponseWriter, req *http.Request) {
			body := extractBody(t, req)
			expected := `` +
				`test.metric.data{test="test_value",test2="second_value"} 14500 1605534165000` + "\n" +
				`gauge_metric_name{test="test_value",test2="second_value",remote_name="156920",url="http://example_url"} 124 1608124661166` + "\n" +
				`gauge_metric_name{test="test_value",test2="second_value",remote_name="156955",url="http://another_url"} 245 1608124662166`
			assert.Equal(t, expected, body)
			assert.Equal(t, "otelcol", req.Header.Get("X-Sumo-Client"))
			assert.Equal(t, "application/vnd.sumologic.prometheus", req.Header.Get("Content-Type"))
		},
	})

	test.s.config.MetricFormat = PrometheusFormat

	metricSum, attrs := exampleIntMetric()
	metricGauge, _ := exampleIntGaugeMetric()
	metrics := metricAndAttrsToPdataMetrics(
		attrs,
		metricSum, metricGauge,
	)
	resMetrics := metrics.ResourceMetrics().At(0)

	flds := fieldsFromMap(map[string]string{
		"key1": "value",
		"key2": "value2",
	})
	_, err := test.s.sendNonOTLPMetrics(context.Background(), resMetrics, flds)
	assert.NoError(t, err)
}

func TestSendMetricsSplit(t *testing.T) {
	test := prepareSenderTest(t, []func(w http.ResponseWriter, req *http.Request){
		func(w http.ResponseWriter, req *http.Request) {
			body := extractBody(t, req)
			expected := `test.metric.data{test="test_value",test2="second_value"} 14500 1605534165000`
			assert.Equal(t, expected, body)
		},
		func(w http.ResponseWriter, req *http.Request) {
			body := extractBody(t, req)
			expected := `` +
				`gauge_metric_name{test="test_value",test2="second_value",remote_name="156920",url="http://example_url"} 124 1608124661166` + "\n" +
				`gauge_metric_name{test="test_value",test2="second_value",remote_name="156955",url="http://another_url"} 245 1608124662166`
			assert.Equal(t, expected, body)
		},
	})
	test.s.config.MaxRequestBodySize = 10
	test.s.config.MetricFormat = PrometheusFormat

	metricSum, attrs := exampleIntMetric()
	metricGauge, _ := exampleIntGaugeMetric()
	metrics := metricAndAttrsToPdataMetrics(
		attrs,
		metricSum, metricGauge,
	)
	resMetrics := metrics.ResourceMetrics().At(0)

	_, err := test.s.sendNonOTLPMetrics(context.Background(), resMetrics, fields{})
	assert.NoError(t, err)
}

func TestSendMetricsSplitFailedOne(t *testing.T) {
	test := prepareSenderTest(t, []func(w http.ResponseWriter, req *http.Request){
		func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(500)

			body := extractBody(t, req)
			expected := `test.metric.data{test="test_value",test2="second_value"} 14500 1605534165000`
			assert.Equal(t, expected, body)
		},
		func(w http.ResponseWriter, req *http.Request) {
			body := extractBody(t, req)
			expected := `` +
				`gauge_metric_name{test="test_value",test2="second_value",remote_name="156920",url="http://example_url"} 124 1608124661166` + "\n" +
				`gauge_metric_name{test="test_value",test2="second_value",remote_name="156955",url="http://another_url"} 245 1608124662166`
			assert.Equal(t, expected, body)
		},
	})
	test.s.config.MaxRequestBodySize = 10
	test.s.config.MetricFormat = PrometheusFormat

	metricSum, attrs := exampleIntMetric()
	metricGauge, _ := exampleIntGaugeMetric()
	metrics := metricAndAttrsToPdataMetrics(
		attrs,
		metricSum, metricGauge,
	)
	resMetrics := metrics.ResourceMetrics().At(0)

	dropped, err := test.s.sendNonOTLPMetrics(context.Background(), resMetrics, fields{})
	assert.EqualError(t, err, "failed sending data: status: 500 Internal Server Error")
	require.Len(t, dropped, 1)
	assert.Equal(t, dropped[0], metricSum)
}

func TestSendMetricsSplitFailedAll(t *testing.T) {
	test := prepareSenderTest(t, []func(w http.ResponseWriter, req *http.Request){
		func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(500)

			body := extractBody(t, req)
			expected := `test.metric.data{test="test_value",test2="second_value"} 14500 1605534165000`
			assert.Equal(t, expected, body)
		},
		func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(404)

			body := extractBody(t, req)
			expected := `` +
				`gauge_metric_name{test="test_value",test2="second_value",remote_name="156920",url="http://example_url"} 124 1608124661166` + "\n" +
				`gauge_metric_name{test="test_value",test2="second_value",remote_name="156955",url="http://another_url"} 245 1608124662166`
			assert.Equal(t, expected, body)
		},
	})
	test.s.config.MaxRequestBodySize = 10
	test.s.config.MetricFormat = PrometheusFormat

	metricSum, attrs := exampleIntMetric()
	metricGauge, _ := exampleIntGaugeMetric()
	metrics := metricAndAttrsToPdataMetrics(
		attrs,
		metricSum, metricGauge,
	)
	resMetrics := metrics.ResourceMetrics().At(0)

	dropped, err := test.s.sendNonOTLPMetrics(context.Background(), resMetrics, fields{})
	assert.EqualError(
		t,
		err,
		"failed sending data: status: 500 Internal Server Error; failed sending data: status: 404 Not Found",
	)
	require.Len(t, dropped, 2)
	assert.Equal(t, dropped[0], metricSum)
	assert.Equal(t, dropped[1], metricGauge)
}

func TestSendMetricsUnexpectedFormat(t *testing.T) {
	// Expect no requestes
	test := prepareSenderTest(t, nil)
	test.s.config.MetricFormat = "invalid"

	metricSum, attrs := exampleIntMetric()
	metrics := metricAndAttrsToPdataMetrics(attrs, metricSum)
	resMetrics := metrics.ResourceMetrics().At(0)

	dropped, err := test.s.sendNonOTLPMetrics(context.Background(), resMetrics, fields{})
	assert.EqualError(t, err, "unexpected metric format: invalid")
	assert.Equal(t, dropped, []pmetric.Metric{metricSum})
}

func TestSendCarbon2Metrics(t *testing.T) {
	test := prepareSenderTest(t, []func(w http.ResponseWriter, req *http.Request){
		func(w http.ResponseWriter, req *http.Request) {
			body := extractBody(t, req)
			//nolint:lll
			expected := `` +
				`test=test_value test2=second_value _unit=m/s escape_me=:invalid_ metric=true metric=test.metric.data unit=bytes  14500 1605534165` +
				"\n" +
				`test=test_value test2=second_value _unit=m/s escape_me=:invalid_ metric=true metric=gauge_metric_name  124 1608124661` +
				"\n" +
				`test=test_value test2=second_value _unit=m/s escape_me=:invalid_ metric=true metric=gauge_metric_name  245 1608124662`
			assert.Equal(t, expected, body)
			assert.Equal(t, "otelcol", req.Header.Get("X-Sumo-Client"))
			assert.Equal(t, "application/vnd.sumologic.carbon2", req.Header.Get("Content-Type"))
		},
	})

	test.s.config.MetricFormat = Carbon2Format

	metricSum, attrs := exampleIntMetric()
	metricGauge, _ := exampleIntGaugeMetric()

	// Assign attributes on a resource level here to see it in the serialized output.
	// This is done only because carbon2 data format does not take data point attributes
	// into account just yet: https://github.com/SumoLogic/sumologic-otel-collector/issues/552
	// TODO: change this to append only to a particular data point's attributes
	// and check only for that in expected output when carbon2 serializer gets fixed.
	attrs.InsertString("unit", "m/s")
	attrs.InsertString("escape me", "=invalid\n")
	attrs.InsertBool("metric", true)

	metrics := metricAndAttrsToPdataMetrics(
		attrs,
		metricSum, metricGauge,
	)
	resMetrics := metrics.ResourceMetrics().At(0)

	flds := fieldsFromMap(map[string]string{
		"key1": "value",
		"key2": "value2",
	})
	_, err := test.s.sendNonOTLPMetrics(context.Background(), resMetrics, flds)
	assert.NoError(t, err)
}

func TestSendGraphiteMetrics(t *testing.T) {
	test := prepareSenderTest(t, []func(w http.ResponseWriter, req *http.Request){
		func(w http.ResponseWriter, req *http.Request) {
			body := extractBody(t, req)
			expected := `` +
				`test_metric_data.my_metric_name.kb 14500 1605534165` +
				"\n" +
				`gauge_metric_name.my_metric_name.kb 124 1608124661` +
				"\n" +
				`gauge_metric_name.my_metric_name.kb 245 1608124662`
			assert.Equal(t, expected, body)
			assert.Equal(t, "otelcol", req.Header.Get("X-Sumo-Client"))
			assert.Equal(t, "application/vnd.sumologic.graphite", req.Header.Get("Content-Type"))
		},
	})

	gf, err := newGraphiteFormatter("%{_metric_}.%{metric}.%{unit}")
	require.NoError(t, err)
	test.s.graphiteFormatter = gf

	test.s.config.MetricFormat = GraphiteFormat
	metricSum, attrs := exampleIntMetric()
	metricGauge, _ := exampleIntGaugeMetric()

	// Assign attributes on a resource level here to see it in the serialized output.
	// This is done only because graphite data format does not take data point attributes
	// into account just yet: https://github.com/SumoLogic/sumologic-otel-collector/issues/552
	// TODO: change this to append only to a particular data point's attributes
	// and check only for that in expected output when graphite serializer gets fixed.
	attrs.InsertString("not_in_graphite_template", "stuff")
	attrs.InsertString("unit", "kb")
	attrs.InsertString("metric", "my_metric_name")
	metrics := metricAndAttrsToPdataMetrics(
		attrs,
		metricSum, metricGauge,
	)
	resMetrics := metrics.ResourceMetrics().At(0)

	flds := fieldsFromMap(map[string]string{
		"key1": "value",
		"key2": "value2",
	})
	_, err = test.s.sendNonOTLPMetrics(context.Background(), resMetrics, flds)
	assert.NoError(t, err)
}
