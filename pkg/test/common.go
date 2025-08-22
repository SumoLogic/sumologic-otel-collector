package sumologic_tests

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"
)

type testSpec struct {
	name        string
	validations []checkFunc
	args        []string
	preActions  []checkFunc
	timeout_ms  int
}

func runTest(tt *testSpec, t *testing.T) (fErr error) {
	ch := testContext{
		test: t,
	}
	defer tearDown()

	for _, v := range tt.preActions {
		if ok := v(ch); !ok {
			return nil
		}
	}

	t.Log("Starting binary execution")
	ctx, cmd, cancel, err := prepareExecuteBinary(tt)
	if err != nil {
		return fmt.Errorf("failed to create command: %w", err)
	}
	ch.commandOutput, ch.commandExitCode, ch.commandErr = getCommandOutput(cmd, ctx, cancel)
	t.Log("Running post validations")
	for _, v := range tt.validations {
		if ok := v(ch); !ok {
			return nil
		}
	}

	return nil
}

func prepareExecuteBinary(tt *testSpec) (context.Context, *exec.Cmd, context.CancelFunc, error) {
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		return nil, nil, nil, fmt.Errorf("binary does not exist: %s", binaryPath)
	}
	if err := isExecutable(binaryPath); err != nil {
		return nil, nil, nil, fmt.Errorf("binary is not executable: %w", err)
	}
	timeout := tt.timeout_ms
	if timeout == 0 {
		timeout = 5000
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Millisecond)
	cmd := exec.CommandContext(ctx, binaryPath, tt.args...)
	if wd, err := os.Getwd(); err == nil {
		cmd.Dir = wd
	}

	return ctx, cmd, cancel, nil
}

func isExecutable(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	mode := info.Mode()
	if mode&0111 == 0 {
		return fmt.Errorf("file is not executable")
	}
	return nil
}

func getCommandOutput(c *exec.Cmd, ctx context.Context, cancel context.CancelFunc) ([]byte, int, error) {
	defer cancel()
	output, err := c.CombinedOutput()
	exitCode := 0
	if c.ProcessState != nil {
		exitCode = c.ProcessState.ExitCode()
	}
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return output, exitCode, nil
	}
	return output, exitCode, err
}

func tearDown() {
	if _, err := os.Stat(logFilePath); err == nil {
		_ = os.Remove(logFilePath)
	}
	killProcessesByName("otelcol-sumo")
	if _, err := os.Stat(credentialsDir); err == nil {
		_ = os.RemoveAll(credentialsDir)
	}
}

func killProcessesByName(name string) {
	out, err := exec.Command("pgrep", name).Output()
	if err != nil || len(out) == 0 {
		return
	}

	pidsStr := strings.TrimSpace(string(out))
	pids := strings.Split(pidsStr, "\n")

	for _, pidStr := range pids {
		if pid, err := strconv.Atoi(strings.TrimSpace(pidStr)); err == nil {
			exec.Command("kill", strconv.Itoa(pid)).Run()
		}
	}
}
