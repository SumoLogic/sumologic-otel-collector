package syslogexporter

import (
	"context"
	"fmt"
	"io"
	"net"
	"strconv"
	"testing"
	"time"

	"go.opentelemetry.io/collector/pdata/pcommon"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.uber.org/zap"
)

var expectedForm = "<165>1 2003-08-24T12:14:15Z 192.0.2.1 myproc 8710 - - It's time to make the do-nuts.\n"

type exporterTest struct {
	srv net.TCPListener
	exp *syslogexporter
}

type testPort struct {
	isIncorrect bool
}

func exampleLog() plog.LogRecord {
	buffer := plog.NewLogRecord()
	buffer.Body().SetStr("<165>1 2003-08-24T05:14:15-07:00 192.0.2.1 myproc 8710 - - It's time to make the do-nuts.")
	timestamp := "2003-08-24T05:14:15-07:00"
	t, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		fmt.Print(err)
	}
	ts := pcommon.NewTimestampFromTime(t)
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
	tgt := logsSlice.AppendEmpty()
	record.CopyTo(tgt)
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

func prepareExporterTest(t *testing.T, cfg *Config, userPort ...testPort) *exporterTest {
	// Start a test syslog server
	var port string
	var addr net.TCPAddr
	addr.IP = net.IP{127, 0, 0, 1}
	addr.Port = 0
	testServer, err := net.ListenTCP("tcp", &addr)
	if err != nil {
		t.Fatal("failed to start test syslog server:", err)
	}
	t.Cleanup(func() { testServer.Close() })
	hostPort := testServer.Addr().String()
	cfg.Endpoint, port, err = net.SplitHostPort(hostPort)
	if err != nil {
		t.Fatalf("could not parse port: %s", err)
	}
	if userPort[0].isIncorrect {
		port = "112"
	}
	cfg.Port, err = strconv.Atoi(port)
	if err != nil {
		t.Fatalf("type error: %s", err)
	}
	exp, err := initExporter(cfg, createExporterCreateSettings())
	require.NoError(t, err)
	return &exporterTest{
		srv: *testServer,
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
	userPort := testPort{isIncorrect: false}
	test := prepareExporterTest(t, createTestConfig(), userPort)
	defer test.srv.Close()
	go func() {
		buffer := exampleLog()
		logs := LogRecordsToLogs(buffer)
		err := test.exp.pushLogsData(context.Background(), logs)
		if err != nil {
			fmt.Printf("could not send message: %s", err)
		}
	}()
	err := test.srv.SetDeadline(time.Now().Add(time.Second * 1))
	if err != nil {
		t.Fatalf("could not accept connection: %s", err)
	}
	conn, err := test.srv.AcceptTCP()
	if err != nil {
		t.Fatalf("could not accept connection: %s", err)
	}
	defer conn.Close()
	b, err := io.ReadAll(conn)
	if err != nil {
		t.Fatalf("could not read all: %s", err)
	}
	assert.Equal(t, string(b), expectedForm)
}

func TestSyslogExportFail(t *testing.T) {
	userPort := testPort{isIncorrect: true}
	test := prepareExporterTest(t, createTestConfig(), userPort)
	defer test.srv.Close()
	buffer := exampleLog()
	logs := LogRecordsToLogs(buffer)
	consumerErr := test.exp.pushLogsData(context.Background(), logs)
	if consumerErr != nil {
		t.Logf("could not send message: %s", consumerErr)
	}
	err := test.srv.SetDeadline(time.Now().Add(time.Second * 1))
	if err != nil {
		t.Fatalf("could not accept connection: %s", err)
	}
	conn, err := test.srv.AcceptTCP()
	if err != nil {
		t.Logf("could not accept connection: %s", err)
		t.Log(conn)
	}
	assert.Contains(t, consumerErr.Error(), "dial tcp 127.0.0.1:112: connect")
}
