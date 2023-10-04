package command

import (
	"context"
	"io"
	"os/exec"
	"syscall"
	"time"
)

const (
	// OKExitStatus specifies the command execution exit status
	// that indicates a success, A-OK.
	OKExitStatus int = 0

	// FallbackExitStatus specifies the command execution exit
	// status used when golang is unable to determine the exit
	// status.
	FallbackExitStatus int = 3
)

// ExecutionRequest describes an external command to be executed.
type ExecutionRequest struct {
	// Command is the command to be executed. This is the only field that must
	// be set non-zero. Behaves like os/exec.Cmd's Path field.
	Command string

	// Arguments to execute the command with.
	Arguments []string

	// Env variables to include
	Env []string

	// Timeout when set non-zero functions as a timer for starting and running
	// a command on Execution.Run
	Timeout time.Duration
}

// NewExecution based on an ExecutionRequest
func NewExecution(ctx context.Context, request ExecutionRequest) *Execution {
	ctx, cancel := context.WithCancel(ctx)
	cmd := exec.CommandContext(ctx, request.Command, request.Arguments...)
	if len(request.Env) > 0 {
		cmd.Env = request.Env
	}

	setOptions(cmd)
	return &Execution{
		cmd:     cmd,
		ctx:     ctx,
		cancel:  cancel,
		timeout: request.Timeout,
	}
}

type Execution struct {
	cmd     *exec.Cmd
	ctx     context.Context
	cancel  func()
	timeout time.Duration
}

// Stdout and Stderr return Readers for the underlying execution's output
// streams. Run must be called subsequently or file descriptors may be leaked.
func (c *Execution) Stdout() (io.Reader, error) {
	return c.cmd.StdoutPipe()
}
func (c *Execution) Stderr() (io.Reader, error) {
	return c.cmd.StderrPipe()
}

// Run the command. May only be invoked once.
func (c *Execution) Run() (resp ExecutionResponse, err error) {
	defer c.cancel()

	started := time.Now()
	defer func() {
		resp.Duration = time.Since(started)
	}()

	if c.timeout > 0 {
		time.AfterFunc(c.timeout, c.cancel)
	}
	if err := c.cmd.Start(); err != nil {
		// Something unexpected happened when attempting to
		// fork/exec, return immediately.
		return resp, err
	}

	// Wait for the process to complete then attempt to determine the result
	err = c.cmd.Wait()
	if err != nil {
		// The command most likely return a non-zero exit status.
		if exitError, ok := err.(*exec.ExitError); ok {
			// Best effort to determine the exit status, this
			// should work on Linux, OSX, and Windows.
			if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
				resp.Status = status.ExitStatus()
			} else {
				resp.Status = FallbackExitStatus
				resp.Error = exitError
			}
		} else {
			// Probably an I/O error
			resp.Status = FallbackExitStatus
			resp.Error = err
		}
	} else {
		// Everything is A-OK.
		resp.Status = OKExitStatus
	}

	return resp, nil
}

// ExecutionResponse provides the response information of an ExecutionRequest.
type ExecutionResponse struct {
	// Command execution exit status.
	Status int

	// Duration provides command execution time.
	Duration time.Duration

	// Error is passed when the outcome of the execution is uncertain
	Error error
}
