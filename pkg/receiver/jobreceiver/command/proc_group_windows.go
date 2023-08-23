package command

import (
	"os/exec"
	"syscall"
)

// setProcessGroup sets the process group of the command process
func setProcessGroup(cmd *exec.Cmd) {
	cmd.SysProcAttr.CreationFlags = syscall.CREATE_NEW_PROCESS_GROUP
}
