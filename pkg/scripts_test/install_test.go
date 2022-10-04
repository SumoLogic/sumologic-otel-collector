package sumologic_scripts_tests

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	binaryPath      string = "/usr/local/bin/otelcol-sumo"
	fileStoragePath string = "/var/lib/sumologic/file_storage"
	etcPath         string = "/etc/otelcol-sumo"
	systemdPath     string = "/etc/systemd/system/otelcol-sumo.service"
	scriptPath      string = "../../scripts/install.sh"
)

type installOptions struct {
	installToken string
	autoconfirm  bool
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

	return opts
}

type check func(*testing.T, installOptions)

func runScript(t *testing.T, opts installOptions) {
	cmd := exec.Command("bash", opts.string()...)

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
	}

	code, err := exitCode(cmd)
	assert.NoError(t, err)

	t.Logf("%v", code)
}

func tearDown(t *testing.T) {
	t.Log("Cleaning up")

	err := os.RemoveAll(fileStoragePath)
	assert.NoError(t, err, "no permissions to remove storage directory")

	err = os.RemoveAll(etcPath)
	assert.NoError(t, err, "no permissions to remove configuration")

	err = os.RemoveAll(systemdPath)
	assert.NoError(t, err, "no permissions to remove systemd configuration")

	err = os.RemoveAll(binaryPath)
	assert.NoError(t, err, "removing binary")
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

func checkBinaryCreated(t *testing.T, opt installOptions) {
	_, err := os.Stat(binaryPath)
	require.NoError(t, err, "binary has not been created")
}

func checkBinaryNotCreated(t *testing.T, opt installOptions) {
	_, err := os.Stat(binaryPath)
	require.ErrorIs(t, err, os.ErrNotExist, "binary is already created")
}

func checkBinaryIsRunning(t *testing.T, opt installOptions) {
	cmd := exec.Command(binaryPath, "--version")
	err := cmd.Start()
	require.NoError(t, err)

	code, err := exitCode(cmd)
	require.NoError(t, err)
	require.Equal(t, 0, code)
}

func TestInstallScript(t *testing.T) {
	for _, tt := range []struct {
		name       string
		options    installOptions
		code       int
		preChecks  []check
		postChecks []check
		preActions []check
	}{
		{
			name:       "no arguments",
			options:    installOptions{},
			preChecks:  []check{checkBinaryNotCreated},
			postChecks: []check{checkBinaryCreated, checkBinaryIsRunning},
		},
		{
			name:       "autoconfirm",
			options:    installOptions{},
			preChecks:  []check{checkBinaryNotCreated},
			postChecks: []check{checkBinaryCreated, checkBinaryIsRunning},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			defer tearDown(t)

			for _, a := range tt.preActions {
				a(t, tt.options)
			}

			for _, c := range tt.preChecks {
				c(t, tt.options)
			}

			runScript(t, tt.options)

			for _, c := range tt.postChecks {
				c(t, tt.options)
			}

		})
	}
}
