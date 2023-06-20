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

func dsclDeletePath(t *testing.T, path string) {
	cmd := exec.Command("dscl", ".", "-delete", path)
	output, err := cmd.CombinedOutput()
	require.NoErrorf(t, err, "error while using dscl to delete path: %s, path: %s", output, path)
	require.Empty(t, string(output))
}

// The user.Lookup() and user.LookupGroup() functions do not appear to work
// correctly on Darwin. The functions will still return a user or group after it
// has been deleted. There are several GitHub issues in github.com/golang/go
// that describe similar or related behaviour. To work around this issue we use
// the dscl command to determine if a user or group exists.
func dsclKeyExistsForPath(t *testing.T, path, key string) bool {
	cmd := exec.Command("dscl", ".", "-list", path)
	out, err := cmd.StdoutPipe()
	if err != nil {
		require.NoError(t, err)
	}
	defer out.Close()

	bufOut := bufio.NewReader(out)

	if err := cmd.Start(); err != nil {
		require.NoError(t, err)
	}

	for {
		line, _, err := bufOut.ReadLine()

		if string(line) == key {
			return true
		}

		// exit if script finished
		if err == io.EOF {
			break
		}

		// otherwise ensure there is no error
		require.NoError(t, err)
	}

	return false
}

func forgetPackage(t *testing.T, name string) {
	noReceiptMsg := fmt.Sprintf("No receipt for '%s' found at '/'.", name)

	output, err := exec.Command("pkgutil", "--forget", name).CombinedOutput()
	if err != nil && !strings.Contains(string(output), noReceiptMsg) {
		require.NoErrorf(t, err, "error forgetting package: %s", string(output))
	}
}

func removeFileIfExists(t *testing.T, path string) {
	if _, err := os.Stat(path); err != nil {
		return
	}

	require.NoErrorf(t, os.Remove(path), "error removing file: %s", path)
}

func removeDirectoryIfExists(t *testing.T, path string) {
	info, err := os.Stat(path)
	if err != nil {
		return
	}

	require.Truef(t, info.IsDir(), "path is not a directory: %s", path)
	require.NoErrorf(t, os.RemoveAll(path), "error removing directory: %s", path)
}

func tearDown(t *testing.T) {
	// Stop service
	unloadLaunchdService(t)

	// Remove files
	removeFileIfExists(t, binaryPath)
	removeFileIfExists(t, launchdPath)

	// Remove configuration & data
	removeDirectoryIfExists(t, etcPath)
	removeDirectoryIfExists(t, fileStoragePath)
	removeDirectoryIfExists(t, logDirPath)
	removeDirectoryIfExists(t, appSupportDirPath)

	// Remove user & group
	if dsclKeyExistsForPath(t, "/Users", systemUser) {
		dsclDeletePath(t, fmt.Sprintf("/Users/%s", systemUser))
	}

	if dsclKeyExistsForPath(t, "/Groups", systemGroup) {
		dsclDeletePath(t, fmt.Sprintf("/Groups/%s", systemGroup))
	}

	if dsclKeyExistsForPath(t, "/Users", systemUser) {
		panic(fmt.Sprintf("user exists after deletion: %s", systemUser))
	}
	if dsclKeyExistsForPath(t, "/Groups", systemGroup) {
		panic(fmt.Sprintf("group exists after deletion: %s", systemGroup))
	}

	// Remove packages
	forgetPackage(t, "com.sumologic.otelcol-sumo-hostmetrics")
	forgetPackage(t, "com.sumologic.otelcol-sumo")
}

func unloadLaunchdService(t *testing.T) {
	info, err := os.Stat(launchdPath)
	if err != nil {
		return
	}

	require.Falsef(t, info.IsDir(), "launchd config is not a file: %s", launchdPath)

	output, err := exec.Command("launchctl", "unload", "-w", "otelcol-sumo").Output()
	require.NoErrorf(t, err, "error stopping service: %s", string(output))
}
