package jobreceiver

import (
	"context"
	"testing"
	"time"

	"github.com/SumoLogic/sumologic-otel-collector/pkg/receiver/jobreceiver/internal/commandtest"
	"github.com/SumoLogic/sumologic-otel-collector/pkg/receiver/jobreceiver/output"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/receiver/receivertest"
)

// TestMain enable command emulation from commandtest
func TestMain(m *testing.M) {
	commandtest.WrapTestMain(m)
}

func TestMonitoringJob(t *testing.T) {
	// basic test
	f := NewFactory()
	cfg := testdataConfigSimple()

	sink := new(consumertest.LogsSink)

	rec, err := f.CreateLogs(context.Background(), receivertest.NewNopSettings(f.Type()), cfg, sink)
	require.NoError(t, err)

	require.NoError(t, rec.Start(context.Background(), componenttest.NewNopHost()))

	if !assert.Eventually(t, expectNLogs(sink, 1), time.Second*5, time.Millisecond*50, "expected one log entry") {
		t.Fatalf("actual %d, %v", sink.LogRecordCount(), sink.AllLogs())
	}
	require.NoError(t, rec.Shutdown(context.Background()))

	first := sink.AllLogs()[0]
	firstRecord := first.ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(0)
	assert.Equal(t, "hello world", firstRecord.Body().AsString())
}

func testdataConfigSimple() *Config {
	cfg := &Config{
		Exec:     newDefaultExecutionConfig(),
		Schedule: ScheduleConfig{Interval: time.Millisecond * 100},
		Output:   output.NewDefaultConfig(),
	}
	cmd, args := commandtest.WrapCommand("echo", []string{"hello world"})
	cfg.Exec.Command, cfg.Exec.Arguments = cmd, args
	return cfg
}

func expectNLogs(sink *consumertest.LogsSink, expected int) func() bool {
	return func() bool {
		return expected <= sink.LogRecordCount()
	}
}
