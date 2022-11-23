//go:build !windows

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
	installToken      string
	autoconfirm       bool
	skipSystemd       bool
	tags              map[string]string
	skipConfig        bool
	skipInstallToken  bool
	fips              bool
	envs              map[string]string
	uninstall         bool
	purge             bool
	apiBaseURL        string
	downloadOnly      bool
	dontKeepDownloads bool
}

func (io *installOptions) string() []string {
	opts := []string{
		scriptPath,
	}

	if io.installToken != "" {
		opts = append(opts, "--installation-token", io.installToken)
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
		opts = append(opts, "--skip-install-token")
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

	if len(io.tags) > 0 {
		for k, v := range io.tags {
			opts = append(opts, "--tag", fmt.Sprintf("%s=%s", k, v))
		}
	}

	if io.apiBaseURL != "" {
		opts = append(opts, "--api", io.apiBaseURL)
	}

	return opts
}

func (io *installOptions) buildEnvs() []string {
	e := os.Environ()

	for k, v := range io.envs {
		e = append(e, fmt.Sprintf("%s=%s", k, v))
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

func runScript(ch check) (int, []string, error) {
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

		if strings.Contains(strLine, "Going to remove") {
			// approve installation config
			_, err = in.Write([]byte("y\n"))
			require.NoError(ch.test, err)
		}
	}

	code, err := exitCode(cmd)
	return code, output, err
}
