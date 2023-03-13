// Copyright 2023, OpenTelemetry Authors
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

package syslogexporter

import (
	"context"
	"errors"
	"io"
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer/consumererror"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.uber.org/zap"
)

var expectedForm = "<165>1 2003-08-24T12:14:15Z 192.0.2.1 myproc 8710 - - It's time to make the do-nuts.\n"
var originalForm = "<165>1 2003-08-24T05:14:15-07:00 192.0.2.1 myproc 8710 - - It's time to make the do-nuts."

type exporterTest struct {
	srv net.TCPListener
	exp *syslogexporter
}

func exampleLog(t *testing.T) plog.LogRecord {
	buffer := plog.NewLogRecord()
	buffer.Body().SetStr(originalForm)
	timestamp := "2003-08-24T05:14:15-07:00"
	timeStr, err := time.Parse(time.RFC3339, timestamp)
	require.NoError(t, err, "failed to start test syslog server")
	ts := pcommon.NewTimestampFromTime(timeStr)
	buffer.SetTimestamp(ts)
	attrMap := map[string]any{"proc_id": "8710", "message": "It's time to make the do-nuts.",
		"appname": "myproc", "hostname": "192.0.2.1", "priority": int64(165),
		"version": int64(1)}
	for k, v := range attrMap {
		if _, ok := v.(string); ok {
			buffer.Attributes().PutStr(k, v.(string))
		} else {
			buffer.Attributes().PutInt(k, v.(int64))
		}
	}
	return buffer
}

func LogRecordsToLogs(record plog.LogRecord) plog.Logs {
	logs := plog.NewLogs()
	logsSlice := logs.ResourceLogs().AppendEmpty().ScopeLogs().AppendEmpty().LogRecords()
	ls := logsSlice.AppendEmpty()
	record.CopyTo(ls)
	return logs
}

func createExporterCreateSettings() exporter.CreateSettings {
	return exporter.CreateSettings{
		TelemetrySettings: component.TelemetrySettings{
			Logger: zap.NewNop(),
		},
	}
}

func TestInitExporter(t *testing.T) {
	_, err := initExporter(&Config{Endpoint: "test.com",
		Protocol: "tcp",
		Port:     514,
		Format:   "rfc5424"}, createExporterCreateSettings())
	assert.NoError(t, err)
}

func buildValidExporter(t *testing.T, server net.TCPListener, cfg *Config) (*syslogexporter, error) {
	var port string
	var err error
	hostPort := server.Addr().String()
	cfg.Endpoint, port, err = net.SplitHostPort(hostPort)
	require.NoError(t, err, "could not parse port")
	cfg.Port, err = strconv.Atoi(port)
	require.NoError(t, err, "type error")
	exp, err := initExporter(cfg, createExporterCreateSettings())
	require.NoError(t, err)
	return exp, err
}

func buildInvalidExporter(t *testing.T, server net.TCPListener, cfg *Config) (*syslogexporter, error) {
	var port string
	var err error
	hostPort := server.Addr().String()
	cfg.Endpoint, port, err = net.SplitHostPort(hostPort)
	require.NoError(t, err, "could not parse endpoint")
	require.NotNil(t, port)
	invalidPort := "112" // Assign invalid port
	cfg.Port, err = strconv.Atoi(invalidPort)
	require.NoError(t, err, "type error")
	exp, err := initExporter(cfg, createExporterCreateSettings())
	require.NoError(t, err)
	return exp, err
}

func createServer() (net.TCPListener, error) {
	var addr net.TCPAddr
	addr.IP = net.IP{127, 0, 0, 1}
	addr.Port = 0
	testServer, err := net.ListenTCP("tcp", &addr)
	return *testServer, err
}

func prepareExporterTest(t *testing.T, cfg *Config, invalidExporter bool) *exporterTest {
	// Start a test syslog server
	var err error
	testServer, err := createServer()
	require.NoError(t, err, "failed to start test syslog server")
	var exp *syslogexporter
	if invalidExporter {
		exp, err = buildInvalidExporter(t, testServer, cfg)
	} else {
		exp, err = buildValidExporter(t, testServer, cfg)
	}
	require.NoError(t, err, "Error building exporter")
	require.NotNil(t, exp)
	return &exporterTest{
		srv: testServer,
		exp: exp,
	}

}

func createTestConfig() *Config {
	config := createDefaultConfig().(*Config)
	config.Protocol = "tcp"
	config.TLSSetting.Insecure = true
	return config
}

func TestSyslogExportSuccess(t *testing.T) {
	test := prepareExporterTest(t, createTestConfig(), false)
	require.NotNil(t, test.exp)
	defer test.srv.Close()
	go func() {
		buffer := exampleLog(t)
		logs := LogRecordsToLogs(buffer)
		err := test.exp.pushLogsData(context.Background(), logs)
		require.NoError(t, err, "could not send message")
	}()
	err := test.srv.SetDeadline(time.Now().Add(time.Second * 1))
	require.NoError(t, err, "cannot set deadline")
	conn, err := test.srv.AcceptTCP()
	require.NoError(t, err, "could not accept connection")
	defer conn.Close()
	b, err := io.ReadAll(conn)
	require.NoError(t, err, "could not read all")
	assert.Equal(t, string(b), expectedForm)
}

func TestSyslogExportFail(t *testing.T) {
	test := prepareExporterTest(t, createTestConfig(), true)
	defer test.srv.Close()
	buffer := exampleLog(t)
	logs := LogRecordsToLogs(buffer)
	consumerErr := test.exp.pushLogsData(context.Background(), logs)
	var consumerErrorLogs consumererror.Logs
	ok := errors.As(consumerErr, &consumerErrorLogs)
	assert.Equal(t, ok, true)
	consumerLogs := consumererror.Logs.Data(consumerErrorLogs)
	rls := consumerLogs.ResourceLogs()
	require.Equal(t, 1, rls.Len())
	scl := rls.At(0).ScopeLogs()
	require.Equal(t, 1, scl.Len())
	lrs := scl.At(0).LogRecords()
	require.Equal(t, 1, lrs.Len())
	droppedLog := lrs.At(0).Body().AsString()
	err := test.srv.SetDeadline(time.Now().Add(time.Second * 1))
	require.NoError(t, err, "cannot set deadline")
	conn, err := test.srv.AcceptTCP()
	require.ErrorContains(t, err, "i/o timeout")
	require.Nil(t, conn)
	assert.ErrorContains(t, consumerErr, "dial tcp 127.0.0.1:112: connect")
	assert.Equal(t, droppedLog, originalForm)
}
