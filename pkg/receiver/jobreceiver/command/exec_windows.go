package command

import (
	"fmt"
	"os/exec"
)

// setOptions sets the process group of the command process
func setOptions(cmd *exec.Cmd) {
	// Cancel functiona adapted from sensu-go's command package
	// TODO(ck) may be worth looking into the windows Job Object api
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
