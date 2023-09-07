package command

import (
	"fmt"
	"os/exec"
	"syscall"
)

// setOptions sets the process group of the command process
func setOptions(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
	}
	cmd.Cancel = func() error {
		// Try with taskkill first
		taskkill := exec.Command(
			"taskkill",
			"/T", "/F", "/PID", fmt.Sprint(cmd.Process.Pid),
		)
		if err := taskkill.Run(); err == nil {
			return nil
		}
		// Fall back to the default behavior
		return cmd.Process.Kill()
	}
}
