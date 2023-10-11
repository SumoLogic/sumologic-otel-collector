package command

import (
	"bytes"
	"context"
	"io"
	"testing"
	"time"

	"github.com/SumoLogic/sumologic-otel-collector/pkg/receiver/jobreceiver/internal/commandtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMain lets the test binary emulate other processes
func TestMain(m *testing.M) {
	commandtest.WrapTestMain(m)
}

func TestExecute(t *testing.T) {
	ctx := context.Background()
	// Basic command execution
	t.Run("basic", func(t *testing.T) {
		echo := NewExecution(ctx, withTestHelper(t, ExecutionRequest{Command: "echo", Arguments: []string{"hello", "world"}}))
		outC := eventualOutput(t, echo)
		resp, err := echo.Run()
		require.NoError(t, err)
		assert.Equal(t, 0, resp.Status)
		require.NoError(t, err)
		assert.Contains(t, <-outC, "hello world")
	})

	// Command exits non-zero
	t.Run("exit code 1", func(t *testing.T) {
		exitCmd := NewExecution(ctx, withTestHelper(t, ExecutionRequest{Command: "exit", Arguments: []string{"1"}}))
		resp, err := exitCmd.Run()
		require.NoError(t, err)
		assert.Equal(t, 1, resp.Status)
		exitCmd = NewExecution(ctx, withTestHelper(t, ExecutionRequest{Command: "exit", Arguments: []string{"33"}}))
		resp, err = exitCmd.Run()
		require.NoError(t, err)
		assert.Equal(t, 33, resp.Status)
	})

	// Command exceeds timeout
	t.Run("exceeds timeout", func(t *testing.T) {
		timeout := time.Millisecond * 100
		sleepCmd := NewExecution(ctx, withTestHelper(t, ExecutionRequest{Command: "sleep", Arguments: []string{"1m"}, Timeout: timeout}))
		done := make(chan struct{})
		go func() {
			resp, err := sleepCmd.Run()
			assert.NoError(t, err)
			assert.NotEqual(t, OKExitStatus, resp.Status)
			assert.LessOrEqual(t, timeout, resp.Duration)
			close(done)
		}()
		select {
		case <-time.After(5 * time.Second):
			t.Errorf("command timeout exceeded but was not killed")
		case <-done:
			// okay
		}
	})
	// Command exceeds timeout with child process
	t.Run("exceeds timeout with child", func(t *testing.T) {
		timeout := time.Millisecond * 100
		sleepCmd := NewExecution(ctx, withTestHelper(t, ExecutionRequest{Command: "fork", Arguments: []string{"sleep", "1m"}, Timeout: timeout}))
		done := make(chan struct{})
		go func() {
			resp, err := sleepCmd.Run()
			assert.NoError(t, err)
			assert.NotEqual(t, OKExitStatus, resp.Status)
			assert.LessOrEqual(t, timeout, resp.Duration)
			close(done)
		}()
		select {
		case <-time.After(5 * time.Second):
			t.Fatal("command timeout exceeded but was not killed")
		case <-done:
			// okay
		}
	})

	// Invocation cannot be spuriously re-invoked
	t.Run("cannot be ran twice", func(t *testing.T) {
		echo := NewExecution(ctx, withTestHelper(t, ExecutionRequest{Command: "echo", Arguments: []string{"hello", "world"}}))
		outC := eventualOutput(t, echo)
		_, err := echo.Run()
		require.NoError(t, err)
		time.Sleep(time.Millisecond * 100)
		_, err = echo.Run()
		assert.Error(t, err)
		assert.Contains(t, <-outC, "hello world")
	})
}

// withTestHelper takes an ExecutionRequest and adjusts it to run with the
// test binary. TestMain will handle emulating the command.
func withTestHelper(t *testing.T, r ExecutionRequest) ExecutionRequest {
	t.Helper()

	r.Command, r.Arguments = commandtest.WrapCommand(r.Command, r.Arguments)
	return r
}

func eventualOutput(t *testing.T, i *Execution) <-chan string {
	t.Helper()
	out := make(chan string, 1)
	stdout, err := i.Stdout()
	require.NoError(t, err)
	stderr, err := i.Stderr()
	require.NoError(t, err)
	go func() {
		var buf bytes.Buffer
		_, err := io.Copy(&buf, stdout)
		require.NoError(t, err)
		_, err = io.Copy(&buf, stderr)
		require.NoError(t, err)
		out <- buf.String()
		close(out)
	}()
	return out
}
