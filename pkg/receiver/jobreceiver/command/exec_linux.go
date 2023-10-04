package command

import (
	"os/exec"
	"syscall"
)

// setOptions sets the process group of the command processprocgroup
func setOptions(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true, Pdeathsig: syscall.SIGTERM}
	cmd.Cancel = func() error {
		// Kill process group instead
		return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	}
}
