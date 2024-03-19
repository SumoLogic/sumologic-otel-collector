//go:build linux || darwin

package sumologic_scripts_tests

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/stretchr/testify/require"
)

type installOptions struct {
	installToken       string
	autoconfirm        bool
	skipSystemd        bool
	tags               map[string]string
	skipConfig         bool
	skipInstallToken   bool
	fips               bool
	envs               map[string]string
	uninstall          bool
	purge              bool
	apiBaseURL         string
	configBranch       string
	downloadOnly       bool
	dontKeepDownloads  bool
	installHostmetrics bool
	remotelyManaged    bool
	ephemeral          bool
	timeout            float64
}

func (io *installOptions) string() []string {
	opts := []string{
		scriptPath,
	}

	if io.autoconfirm {
		opts = append(opts, "--yes")
	}

	if io.fips {
		opts = append(opts, "--fips")
	}

	if io.skipSystemd {
		opts = append(opts, "--skip-systemd")
	}

	if io.skipConfig {
		opts = append(opts, "--skip-config")
	}

	if io.skipInstallToken {
		opts = append(opts, "--skip-installation-token")
	}

	if io.uninstall {
		opts = append(opts, "--uninstall")
	}

	if io.purge {
		opts = append(opts, "--purge")
	}

	if io.downloadOnly {
		opts = append(opts, "--download-only")
	}

	if !io.dontKeepDownloads {
		opts = append(opts, "--keep-downloads")
	}

	if io.installHostmetrics {
		opts = append(opts, "--install-hostmetrics")
	}

	if io.remotelyManaged {
		opts = append(opts, "--remotely-managed")
	}

	if io.ephemeral {
		opts = append(opts, "--ephemeral")
	}

	if len(io.tags) > 0 {
		for k, v := range io.tags {
			opts = append(opts, "--tag", fmt.Sprintf("%s=%s", k, v))
		}
	}

	if io.apiBaseURL != "" {
		opts = append(opts, "--api", io.apiBaseURL)
	}

	if io.configBranch != "" {
		opts = append(opts, "--config-branch", io.configBranch)
	}

	if io.timeout != 0 {
		opts = append(opts, "--download-timeout", fmt.Sprintf("%f", io.timeout))
	}

	return opts
}

func (io *installOptions) buildEnvs() []string {
	e := os.Environ()

	for k, v := range io.envs {
		e = append(e, fmt.Sprintf("%s=%s", k, v))
	}

	if io.installToken != "" {
		e = append(e, fmt.Sprintf("%s=%s", installTokenEnv, io.installToken))
	}

	return e
}

func exitCode(cmd *exec.Cmd) (int, error) {
	err := cmd.Wait()

	if err == nil {
		return cmd.ProcessState.ExitCode(), nil
	}

	if exiterr, ok := err.(*exec.ExitError); ok {
		return exiterr.ExitCode(), nil
	}

	return 0, fmt.Errorf("cannot obtain exit code: %v", err)
}

func runScript(ch check) (int, []string, []string, error) {
	cmd := exec.Command("bash", ch.installOptions.string()...)
	cmd.Env = ch.installOptions.buildEnvs()
	output := []string{}

	in, err := cmd.StdinPipe()
	if err != nil {
		require.NoError(ch.test, err)
	}

	defer in.Close()

	out, err := cmd.StdoutPipe()
	if err != nil {
		require.NoError(ch.test, err)
	}
	defer out.Close()

	errOut, err := cmd.StderrPipe()
	if err != nil {
		require.NoError(ch.test, err)
	}
	defer errOut.Close()

	// We want to read line by line
	bufOut := bufio.NewReader(out)

	// Start the process
	if err = cmd.Start(); err != nil {
		require.NoError(ch.test, err)
	}

	// Read the results from the process
	for {
		line, _, err := bufOut.ReadLine()
		strLine := strings.TrimSpace(string(line))

		if len(strLine) > 0 {
			output = append(output, strLine)
		}
		ch.test.Log(strLine)

		// exit if script finished
		if err == io.EOF {
			break
		}

		// otherwise ensure there is no error
		require.NoError(ch.test, err)

		if ch.installOptions.autoconfirm {
			continue
		}

	}

	// Handle stderr separately
	bufErrOut := bufio.NewReader(errOut)
	errorOutput := []string{}
	for {
		line, _, err := bufErrOut.ReadLine()
		strLine := strings.TrimSpace(string(line))

		if len(strLine) > 0 {
			errorOutput = append(errorOutput, strLine)
		}
		ch.test.Log(strLine)

		// exit if script finished
		if err == io.EOF {
			break
		}
	}

	code, err := exitCode(cmd)
	return code, output, errorOutput, err
}
