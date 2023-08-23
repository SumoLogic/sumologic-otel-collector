package command

import (
	"os/exec"
	"syscall"
)

// setProcessGroup sets the process group of the command processprocgroup
func setProcessGroup(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true, Pdeathsig: syscall.SIGTERM}
}
