//go:build !windows

package sumologic_scripts_tests

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

type testSpec struct {
	name              string
	options           installOptions
	preChecks         []checkFunc
	postChecks        []checkFunc
	preActions        []checkFunc
	conditionalChecks []condCheckFunc
	installCode       int
}

// These checks always have to be true after a script execution
var commonPostChecks = []checkFunc{checkNoBakFilesPresent}

func tearDown(t *testing.T) {
	t.Log("Cleaning up")

	switch runtime.GOOS {
	case "darwin":
		tearDownDarwin(t)
	default:
		tearDownOther(t)
	}
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

func tearDownDarwin(t *testing.T) {
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
	systemUser := getSystemUser()
	if dsclKeyExistsForPath(t, "/Users", systemUser) {
		dsclDeletePath(t, fmt.Sprintf("/Users/%s", systemUser))
	}

	systemGroup := getSystemGroup()
	if dsclKeyExistsForPath(t, "/Groups", systemGroup) {
		dsclDeletePath(t, fmt.Sprintf("/Groups/%s", systemGroup))
	}

	if dsclKeyExistsForPath(t, "/Users", systemUser) {
		panic(fmt.Sprintf("user exists after deletion: %s", systemUser))
	}
	if dsclKeyExistsForPath(t, "/Groups", systemGroup) {
		panic(fmt.Sprintf("group exists after deletion: %s", systemGroup))
	}
}

func tearDownOther(t *testing.T) {
	ch := check{
		test: t,
		installOptions: installOptions{
			uninstall:   true,
			purge:       true,
			autoconfirm: true,
		},
	}

	_, _, _, err := runScript(ch)
	require.NoError(t, err)
}

func cleanCache(t *testing.T) {
	err := os.RemoveAll(cacheDirectory)
	require.NoError(t, err)
}

func runTest(t *testing.T, spec *testSpec) {
	ch := check{
		test:                t,
		installOptions:      spec.options,
		expectedInstallCode: spec.installCode,
	}

	t.Log("Running conditional checks")
	for _, a := range spec.conditionalChecks {
		if !a(ch) {
			t.SkipNow()
		}
	}

	defer tearDown(t)

	t.Log("Running pre actions")
	for _, a := range spec.preActions {
		a(ch)
	}

	t.Log("Running pre checks")
	for _, c := range spec.preChecks {
		c(ch)
	}

	ch.code, ch.output, ch.errorOutput, ch.err = runScript(ch)

	// Remove cache in case of curl issue
	if ch.code == curlTimeoutErrorCode {
		cleanCache(t)
	}

	checkRun(ch)

	t.Log("Running common post checks")
	for _, c := range commonPostChecks {
		c(ch)
	}

	t.Log("Running post checks")
	for _, c := range spec.postChecks {
		c(ch)
	}
}

func getSystemUser() string {
	switch runtime.GOOS {
	case "darwin":
		return fmt.Sprintf("_%s", systemUser)
	case "linux":
		return systemUser
	}

	panic(fmt.Sprintf("Encountered unsupported OS: %s", runtime.GOOS))
}

func getSystemGroup() string {
	switch runtime.GOOS {
	case "darwin":
		return fmt.Sprintf("_%s", systemGroup)
	case "linux":
		return systemGroup
	}

	panic(fmt.Sprintf("Encountered unsupported OS: %s", runtime.GOOS))
}

func getRootGroupName() string {
	if runtime.GOOS == "darwin" {
		return "wheel"
	} else if runtime.GOOS == "linux" {
		return "root"
	}

	panic(fmt.Sprintf("Encountered unsupported OS: %s", runtime.GOOS))
}

func getPackagePath() string {
	tmpDir := os.Getenv("TMPDIR")
	packageName := fmt.Sprintf("%s-package-name-not-defined", runtime.GOOS)

	switch runtime.GOOS {
	case "darwin":
		packageName = darwinPackageName
	default:
		panic(fmt.Sprintf("Encountered unsupported OS: %s", runtime.GOOS))
	}

	return path.Join(tmpDir, packageName)
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

func dsclDeletePath(t *testing.T, path string) {
	cmd := exec.Command("dscl", ".", "-delete", path)
	output, err := cmd.CombinedOutput()
	require.NoErrorf(t, err, "error while using dscl to delete path: %s, path: %s", output, path)
	require.Empty(t, string(output))
}
