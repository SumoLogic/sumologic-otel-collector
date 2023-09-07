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
	"sync"
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

//nolint:all
func TestExecute(t *testing.T) {
	ctx := context.Background()
	// Basic command execution
	t.Run("basic", func(t *testing.T) {
		echo := NewExecution(ctx, withTestHelper(t, ExecutionRequest{Command: "echo", Arguments: []string{"hello", "world"}}))
		outC := eventualOutput(echo)
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

	// Invocation cannot be spuriously re-invoked
	t.Run("cannot be ran twice", func(t *testing.T) {
		echo := NewExecution(ctx, withTestHelper(t, ExecutionRequest{Command: "echo", Arguments: []string{"hello", "world"}}))
		outC := eventualOutput(echo)
		echo.Run()
		_, err := echo.Run()
		assert.ErrorIs(t, ErrAlreadyExecuted, err)
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

//nolint:all
func eventualOutput(i *Execution) <-chan string {
	out := make(chan string, 1)
	stdout, _ := i.Stdout()
	stderr, _ := i.Stderr()
	go func() {
		var buf syncBuffer
		io.Copy(&buf, stdout)
		io.Copy(&buf, stderr)
		out <- buf.String()
		close(out)
	}()
	return out
}

type syncBuffer struct {
	buf bytes.Buffer
	mu  sync.Mutex
}

func (s *syncBuffer) Write(p []byte) (n int, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.buf.Write(p)
}

func (s *syncBuffer) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.buf.String()
}
