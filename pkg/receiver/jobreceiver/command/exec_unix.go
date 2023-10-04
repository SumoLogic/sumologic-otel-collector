//go:build !linux && !windows
// +build !linux,!windows

package command

import (
	"os/exec"
	"syscall"
)

// setOptions sets the process group of the command process
func setOptions(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error {
		// Kill process group instead
		return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	}
}
