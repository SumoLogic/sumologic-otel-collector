package command

import (
	"context"
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

// TestHelperProcess is a target that helps this suite emulate
// platform specific executables in a way that is platform agnostic.
func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	var args []string
	command := os.Args[3]
	if len(os.Args) > 4 {
		args = os.Args[4:]
	}

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
	}

}

//nolint:all
func TestExecute(t *testing.T) {

	ctx := context.Background()
	// Basic command execution
	echo := fakeCommand(t, "echo", "hello", "world")
	outC := eventualOutput(echo)
	resp, err := echo.Run(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, resp.Status)
	require.NoError(t, err)
	assert.Contains(t, <-outC, "hello world")

	// Command exits non-zero
	exitCmd := fakeCommand(t, "exit", "1")
	eventualOutput(exitCmd)
	resp, err = exitCmd.Run(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, resp.Status)
	exitCmd = fakeCommand(t, "exit", "33")
	eventualOutput(exitCmd)
	resp, err = exitCmd.Run(ctx)
	require.NoError(t, err)
	assert.Equal(t, 33, resp.Status)

	// Command canceled by context
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Millisecond*100)
	sleepCmd := fakeCommand(t, "sleep", "1m")
	eventualOutput(sleepCmd)
	done := make(chan struct{})
	go func() {
		resp, err = sleepCmd.Run(timeoutCtx)
		require.NoError(t, err)
		close(done)
	}()
	select {
	case <-time.After(time.Second):
		t.Errorf("command context expired but was not killed")
	case <-done:
		// okay
	}
	cancel()

	// Command exceeds timeout
	sleepCmd = fakeCommand(t, "sleep", "1m")
	sleepCmd.Timeout = time.Millisecond * 100
	eventualOutput(sleepCmd)
	done = make(chan struct{})
	go func() {
		resp, err := sleepCmd.Run(ctx)
		assert.NoError(t, err)
		assert.Equal(t, TimeoutExitStatus, resp.Status)
		close(done)
	}()
	select {
	case <-time.After(5 * time.Second):
		t.Errorf("command timeout exceeded but was not killed")
	case <-done:
		// okay
	}

	// Invocation cannot be spuriously re-invoked
	echo = fakeCommand(t, "echo", "hello", "world")
	outC = eventualOutput(echo)
	echo.Run(ctx)
	_, err = echo.Run(ctx)
	assert.ErrorIs(t, ErrAlreadyExecuted, err)
	assert.Contains(t, <-outC, "hello world")
}

// fakeCommand takes a command and (optionally) command args and will execute
// the TestHelperProcess test within the package FakeCommand is called from.
func fakeCommand(t *testing.T, command string, args ...string) *invocation {
	cargs := []string{"-test.run=TestHelperProcess", "--", command}
	cargs = append(cargs, args...)
	env := []string{"GO_WANT_HELPER_PROCESS=1"}

	execution := ExecutionRequest{
		Command:   os.Args[0],
		Arguments: cargs,
		Env:       env,
	}

	c, err := NewInvocation(execution)
	require.NoError(t, err)
	cmd, ok := c.(*invocation)
	require.True(t, ok)
	return cmd
}

func eventualOutput(i Invocation) <-chan string {
	out := make(chan string, 1)
	go func() {
		var buf SyncBuffer
		defer i.Stdout().Close()
		defer i.Stderr().Close()
		io.Copy(&buf, i.Stdout())
		io.Copy(&buf, i.Stderr())
		out <- buf.String()
		close(out)
	}()
	return out
}
