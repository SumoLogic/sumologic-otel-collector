//go:build !windows
// +build !windows

package command

import (
	"context"
	"os/exec"
	"syscall"
)

// command returns a command to execute an executable
func command(ctx context.Context, command string, args []string) *exec.Cmd {
	return exec.CommandContext(ctx, command, args...)
}

// killProcess kills the command process and any child processes
func killProcess(cmd *exec.Cmd) error {
	return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
}
