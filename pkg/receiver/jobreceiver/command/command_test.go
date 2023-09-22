package command

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMain lets the test binary emulate other processes
func TestMain(m *testing.M) {
	flag.Parse()

	pid := os.Getpid()
	if os.Getenv("GO_EXEC_TEST_PID") == "" {
		os.Setenv("GO_EXEC_TEST_PID", strconv.Itoa(pid))
		os.Exit(m.Run())
	}

	args := flag.Args()
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "No command\n")
		os.Exit(2)
	}

	command, args := args[0], args[1:]
	switch command {
	case "echo":
		fmt.Fprintf(os.Stdout, "%s", strings.Join(args, " "))
	case "exit":
		if i, err := strconv.ParseInt(args[0], 10, 32); err == nil {
			os.Exit(int(i))
		}
		panic("unexpected exit argument")
	case "sleep":
		if d, err := time.ParseDuration(args[0]); err == nil {
			time.Sleep(d)
			return
		}
		if i, err := strconv.ParseInt(args[0], 10, 64); err == nil {
			time.Sleep(time.Second * time.Duration(i))
			return
		}
	case "fork":
		childCommand := NewExecution(context.Background(), ExecutionRequest{
			Command:   os.Args[0],
			Arguments: args,
		})
		_, err := childCommand.Run()
		if err != nil {
			fmt.Fprintf(os.Stderr, "fork error: %v", err)
			os.Exit(3)
		}
	}
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
		_, err = echo.Run()
		assert.Error(t, err)
		assert.Contains(t, <-outC, "hello world")
	})
}

// withTestHelper takes an ExecutionRequest and adjusts it to run with the
// test binary. TestMain will handle emulating the command.
func withTestHelper(t *testing.T, request ExecutionRequest) ExecutionRequest {
	t.Helper()

	request.Arguments = append([]string{request.Command}, request.Arguments...)
	request.Command = os.Args[0]
	return request
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
