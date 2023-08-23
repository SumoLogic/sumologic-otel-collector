//go:build !linux && !windows
// +build !linux,!windows

package command

import (
	"os/exec"
	"syscall"
)

// setProcessGroup sets the process group of the command process
func setProcessGroup(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}
