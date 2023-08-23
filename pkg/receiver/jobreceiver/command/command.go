package command

import (
	"context"
	"fmt"
	"io"
	"math"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"
)

const (
	// TimeoutOutput specifies the command execution output in the
	// event of an execution timeout.
	TimeoutOutput string = "Execution timed out\n"

	// OKExitStatus specifies the command execution exit status
	// that indicates a success, A-OK.
	OKExitStatus int = 0

	// TimeoutExitStatus specifies the command execution exit
	// status in the event of an execution timeout.
	TimeoutExitStatus int = 2

	// FallbackExitStatus specifies the command execution exit
	// status used when golang is unable to determine the exit
	// status.
	FallbackExitStatus int = 3

	// ErrAlreadyExecuted don't do that.
	ErrAlreadyExecuted = cError("invocation has already been started once")
)

// ExecutionRequest provides information about a system command execution,
// somewhat of an abstraction intended to be used for monitoring jobs
type ExecutionRequest struct {
	// Command is the command to be executed.
	Command string

	// Arguments to execute the command with.
	Arguments []string

	// Env ...
	Env []string

	// Execution timeout
	Timeout time.Duration
}

type Invocation interface {
	// Stdout stream. Must by closed by the caller.
	Stdout() io.ReadCloser
	// Stderr stream. Must by closed by the caller.
	Stderr() io.ReadCloser
	// Run the command. Only valid once
	Run(context.Context) (ExecutionResponse, error)
}

// NewInvocation based on an ExecutionRequest
func NewInvocation(request ExecutionRequest) (Invocation, error) {
	c := &invocation{ExecutionRequest: request}
	var err error
	c.stdoutReader, c.stdoutWritter, err = os.Pipe()
	if err != nil {
		return c, err
	}
	c.stderrReader, c.stderrWritter, err = os.Pipe()
	return c, err
}

type invocation struct {
	ExecutionRequest

	stdoutWritter io.WriteCloser
	stderrWritter io.WriteCloser
	stdoutReader  io.ReadCloser
	stderrReader  io.ReadCloser

	mu      sync.Mutex
	started bool
}

func (c *invocation) Stdout() io.ReadCloser {
	return c.stdoutReader
}
func (c *invocation) Stderr() io.ReadCloser {
	return c.stderrReader
}
func (c *invocation) Run(ctx context.Context) (ExecutionResponse, error) {
	resp := ExecutionResponse{}

	c.mu.Lock()
	if c.started {
		defer c.mu.Unlock()
		return resp, ErrAlreadyExecuted
	}
	c.started = true
	c.mu.Unlock()

	// Using a platform specific shell to "cheat", as the shell
	// will handle certain failures for us, where golang exec is
	// known to have troubles, e.g. command not found. We still
	// use a fallback exit status in the unlikely event that the
	// exit status cannot be determined.
	var cmd *exec.Cmd

	// Use context.WithCancel for command execution timeout.
	// context.WithTimeout will not kill child/grandchild processes
	// (see issues tagged in https://github.com/sensu/sensu-go/issues/781).
	// Rather, we will use a timer, CancelFunc and proc functions
	// to perform full cleanup.
	ctx, timeout := context.WithCancel(ctx)
	defer timeout()

	// Taken from Sensu-Spawn (Sensu 1.x.x).
	cmd = command(ctx, c.Command, c.Arguments)

	// Set the ENV for the command if it is set
	if len(c.Env) > 0 {
		cmd.Env = c.Env
	}

	cmd.Stdout = c.stdoutWritter
	cmd.Stderr = c.stderrWritter
	defer func() {
		c.stdoutWritter.Close()
		c.stderrWritter.Close()
	}()

	started := time.Now()
	defer func() {
		resp.Duration = time.Since(started)
	}()

	timer := time.NewTimer(math.MaxInt64)
	defer timer.Stop()
	if c.Timeout > 0 {
		setProcessGroup(cmd)
		timer.Stop()
		timer = time.NewTimer(c.Timeout)
	}
	if err := cmd.Start(); err != nil {
		// Something unexpected happened when attempting to
		// fork/exec, return immediately.
		return resp, err
	}

	waitCh := make(chan struct{})
	var err error
	go func() {
		err = cmd.Wait()
		close(waitCh)
	}()

	// Wait for the process to complete or the timer to trigger, whichever comes first.
	var killErr error
	select {
	case <-waitCh:
		if err != nil {
			// The command most likely return a non-zero exit status.
			if exitError, ok := err.(*exec.ExitError); ok {
				// Best effort to determine the exit status, this
				// should work on Linux, OSX, and Windows.
				if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
					resp.Status = status.ExitStatus()
				} else {
					resp.Status = FallbackExitStatus
				}
			} else {
				resp.Status = FallbackExitStatus
			}
		} else {
			// Everything is A-OK.
			resp.Status = OKExitStatus
		}

	case <-timer.C:
		var killErrOutput string
		if killErr = killProcess(cmd); killErr != nil {
			killErrOutput = fmt.Sprintf("Unable to TERM/KILL the process: #%d\n", cmd.Process.Pid)
		}
		timeout()
		fmt.Fprintf(c.stderrWritter, "%s%s", TimeoutOutput, killErrOutput)
		resp.Status = TimeoutExitStatus
	}

	return resp, nil
}

// ExecutionResponse provides the response information of an ExecutionRequest.
type ExecutionResponse struct {
	// Command execution exit status.
	Status int

	// Duration provides command execution time.
	Duration time.Duration
}

// cError const error type for sentinels
type cError string

func (e cError) Error() string {
	return string(e)
}
