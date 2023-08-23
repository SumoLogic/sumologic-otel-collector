//go:build windows
// +build windows

package command

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

// command returns a command to execute an executable
func command(ctx context.Context, command string, args []string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, "cmd")
	cl := append([]string{command}, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		// Manually set the command line arguments so they are not escaped
		// https://github.com/golang/go/commit/f18a4e9609aac3aa83d40920c12b9b45f9376aea
		// http://www.josephspurrier.com/prevent-escaping-exec-command-arguments-in-go/
		CmdLine: strings.Join(cl, " "),
	}
	return cmd
}

// killProcess kills the command process and any child processes
func killProcess(cmd *exec.Cmd) error {
	process := cmd.Process
	if process == nil {
		return nil
	}

	args := []string{fmt.Sprintf("/T /F /PID %d", process.Pid)}
	err := command(context.Background(), "taskkill", args).Run()
	if err == nil {
		return nil
	}

	err = forceKill(process)
	if err == nil {
		return nil
	}
	err = process.Signal(os.Kill)

	return fmt.Errorf("could not kill process")
}

func forceKill(process *os.Process) error {
	handle, err := syscall.OpenProcess(syscall.PROCESS_TERMINATE, true, uint32(process.Pid))
	if err != nil {
		return err
	}

	err = syscall.TerminateProcess(handle, 0)
	_ = syscall.CloseHandle(handle)

	return err
}
