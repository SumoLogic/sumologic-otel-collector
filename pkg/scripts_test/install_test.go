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

func runScript(t *testing.T) {
	cmd := exec.Command("bash", "-c", "../../scripts/install.sh")

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

	err := os.RemoveAll("/var/lib/sumologic/file_storage")
	assert.NoError(t, err, "no permissions to remove storage directory")

	err = os.RemoveAll("/etc/otelcol-sumo")
	assert.NoError(t, err, "no permissions to remove configuration")

	err = os.RemoveAll("/etc/systemd/system/otelcol-sumo.service")
	assert.NoError(t, err, "no permissions to remove systemd configuration")

	err = os.RemoveAll("/usr/local/bin/otelcol-sumo")
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

func TestInstallation(t *testing.T) {
	defer tearDown(t)

	_, err := os.Stat("/usr/local/bin/otelcol-sumo")
	require.ErrorIs(t, err, os.ErrNotExist, "/usr/local/bin/otelcol-sumo is already created")
	runScript(t)

	_, err = os.Stat("/usr/local/bin/otelcol-sumo")
	require.NoError(t, err, "/usr/local/bin/otelcol-sumo has not been created")

	cmd := exec.Command("/usr/local/bin/otelcol-sumo", "--version")
	err = cmd.Start()
	require.NoError(t, err)

	code, err := exitCode(cmd)
	require.NoError(t, err)
	require.Equal(t, 0, code)
}
