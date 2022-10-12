package sumologic_scripts_tests

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type installOptions struct {
	installToken     string
	autoconfirm      bool
	disableSystemd   bool
	tags             map[string]string
	skipInstallToken bool
	envs             map[string]string
	uninstall        bool
	purge            bool
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

	if io.disableSystemd {
		opts = append(opts, "--disable-systemd-installation")
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

	if len(io.tags) > 0 {
		for k, v := range io.tags {
			opts = append(opts, "--tag", fmt.Sprintf("%s=%s", k, v))
		}
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

func runScript(t *testing.T, opts installOptions) (int, error) {
	cmd := exec.Command("bash", opts.string()...)
	cmd.Env = opts.buildEnvs()

	in, err := cmd.StdinPipe()
	if err != nil {
		require.NoError(t, err)
	}

	defer in.Close()

	out, err := cmd.StdoutPipe()
	if err != nil {
		require.NoError(t, err)
	}

	defer out.Close()

	// We want to read line by line
	bufOut := bufio.NewReader(out)

	// Start the process
	if err = cmd.Start(); err != nil {
		require.NoError(t, err)
	}

	// Read the results from the process
	for {
		line, _, err := bufOut.ReadLine()
		strLine := string(line)
		t.Log(strLine)

		// exit if script finished
		if err == io.EOF {
			break
		}

		// otherwise ensure there is no error
		require.NoError(t, err)

		if opts.autoconfirm {
			continue
		}

		if strings.Contains(strLine, "Showing full changelog") {
			// show changelog
			_, err = in.Write([]byte("\n"))
			require.NoError(t, err)

			// accept changes and proceed with the installation
			_, err = in.Write([]byte("y\n"))
			require.NoError(t, err)
		}

		if strings.Contains(strLine, "We are going to get and set up default configuration for you") {
			// approve installation config
			_, err = in.Write([]byte("y\n"))
			require.NoError(t, err)
		}

		if strings.Contains(strLine, "We are going to set up systemd service") {
			// approve installation config
			_, err = in.Write([]byte("y\n"))
			require.NoError(t, err)
		}

		if strings.Contains(strLine, "Going to remove") {
			// approve installation config
			_, err = in.Write([]byte("y\n"))
			require.NoError(t, err)
		}
	}

	return exitCode(cmd)
}
