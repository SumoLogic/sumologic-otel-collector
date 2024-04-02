package sumologic_scripts_tests

import (
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

type check struct {
	test                *testing.T
	installOptions      installOptions
	code                int
	err                 error
	expectedInstallCode int
	output              []string
	errorOutput         []string
}

type condCheckFunc func(check) bool

func checkSkipTest(c check) bool {
	return false
}

type checkFunc func(check)

func checkBinaryCreated(c check) {
	require.FileExists(c.test, binaryPath, "binary has not been created")
}

func checkBinaryNotCreated(c check) {
	require.NoFileExists(c.test, binaryPath, "binary is already created")
}

func checkBinaryIsRunning(c check) {
	cmd := exec.Command(binaryPath, "--version")
	err := cmd.Start()
	require.NoError(c.test, err, "error while checking version")

	code, err := exitCode(cmd)
	require.NoError(c.test, err, "error while checking exit code")
	require.Equal(c.test, 0, code, "got error code while checking version")
}

func checkLatestAppVersion(c check) {
	cmd := exec.Command(binaryPath, "--version")
	output, err := cmd.Output()
	c.test.Logf("latest app version: %s", latestAppVersion)
	require.NoError(c.test, err, "error while checking version")
	require.Contains(c.test, string(output), latestAppVersion, "must install latest app version")
}

func checkRun(c check) {
	require.Equal(c.test, c.expectedInstallCode, c.code, "unexpected installation script error code")
}

func checkConfigCreated(c check) {
	require.FileExists(c.test, configPath, "configuration has not been created properly")
}

func checkConfigNotCreated(c check) {
	require.NoFileExists(c.test, configPath, "configuration has been created")
}

func checkConfigOverrided(c check) {
	conf, err := getConfig(configPath)
	require.NoError(c.test, err)

	require.Condition(c.test, func() (success bool) {
		switch conf.Extensions.Sumologic.InstallationToken {
		case "${SUMOLOGIC_INSTALLATION_TOKEN}":
			return true
		default:
			return false
		}
	}, "invalid value for installation token")
}

func checkUserConfigCreated(c check) {
	require.FileExists(c.test, userConfigPath, "user configuration has not been created properly")
}

func checkUserConfigNotCreated(c check) {
	require.NoFileExists(c.test, userConfigPath, "user configuration has been created")
}

func checkHomeDirectoryCreated(c check) {
	require.DirExists(c.test, libPath, "home directory has not been created properly")
}

func checkNoBakFilesPresent(c check) {
	cwd, err := os.Getwd()
	require.NoError(c.test, err)
	cwdGlob := filepath.Join(cwd, "*.bak")
	etcPathGlob := filepath.Join(etcPath, "*.bak")
	etcPathNestedGlob := filepath.Join(etcPath, "*", "*.bak")

	for _, bakGlob := range []string{cwdGlob, etcPathGlob, etcPathNestedGlob} {
		bakFiles, err := filepath.Glob(bakGlob)
		require.NoError(c.test, err)
		require.Empty(c.test, bakFiles)
	}
}

func checkHostmetricsConfigCreated(c check) {
	require.FileExists(c.test, hostmetricsConfigPath, "hostmetrics configuration has not been created properly")
}

func checkHostmetricsConfigNotCreated(c check) {
	require.NoFileExists(c.test, hostmetricsConfigPath, "hostmetrics configuration has been created")
}

func checkRemoteConfigDirectoryCreated(c check) {
	require.DirExists(c.test, opampDPath, "remote configuration directory has not been created properly")
}

func checkRemoteConfigDirectoryNotCreated(c check) {
	require.NoDirExists(c.test, opampDPath, "remote configuration directory has been created")
}

func checkTags(c check) {
	conf, err := getConfig(userConfigPath)
	require.NoError(c.test, err, "error while reading configuration")

	for k, v := range c.installOptions.tags {
		require.Equal(c.test, v, conf.Extensions.Sumologic.Tags[k], "tag is different than expected")
	}
}

func checkDifferentTags(c check) {
	conf, err := getConfig(userConfigPath)
	require.NoError(c.test, err, "error while reading configuration")

	require.Equal(c.test, "tag", conf.Extensions.Sumologic.Tags["some"], "tag is different than expected")
}

func checkAbortedDueToDifferentToken(c check) {
	require.Greater(c.test, len(c.output), 0)
	require.Contains(c.test, c.output[len(c.output)-1], "You are trying to install with different token than in your configuration file!")
}

func preActionWriteAPIBaseURLToUserConfig(c check) {
	conf, err := getConfig(userConfigPath)
	require.NoError(c.test, err)

	conf.Extensions.Sumologic.APIBaseURL = c.installOptions.apiBaseURL
	err = saveConfig(userConfigPath, conf)
	require.NoError(c.test, err)
}

func preActionWriteDifferentAPIBaseURLToUserConfig(c check) {
	conf, err := getConfig(userConfigPath)
	require.NoError(c.test, err)

	conf.Extensions.Sumologic.APIBaseURL = "different" + c.installOptions.apiBaseURL
	err = saveConfig(userConfigPath, conf)
	require.NoError(c.test, err)
}

func preActionWriteDifferentTagsToUserConfig(c check) {
	conf, err := getConfig(userConfigPath)
	require.NoError(c.test, err)

	conf.Extensions.Sumologic.Tags = map[string]string{
		"some": "tag",
	}
	err = saveConfig(userConfigPath, conf)
	require.NoError(c.test, err)
}

func preActionWriteEmptyUserConfig(c check) {
	conf, err := getConfig(userConfigPath)
	require.NoError(c.test, err)

	err = saveConfig(userConfigPath, conf)
	require.NoError(c.test, err)
}

func preActionWriteTagsToUserConfig(c check) {
	conf, err := getConfig(userConfigPath)
	require.NoError(c.test, err)

	conf.Extensions.Sumologic.Tags = c.installOptions.tags
	err = saveConfig(userConfigPath, conf)
	require.NoError(c.test, err)
}

func checkAbortedDueToDifferentAPIBaseURL(c check) {
	require.Greater(c.test, len(c.output), 0)
	require.Contains(c.test, c.output[len(c.output)-1], "You are trying to install with different api base url than in your configuration file!")
}

func checkAPIBaseURLInConfig(c check) {
	conf, err := getConfig(userConfigPath)
	require.NoError(c.test, err, "error while reading configuration")

	require.Equal(c.test, c.installOptions.apiBaseURL, conf.Extensions.Sumologic.APIBaseURL, "api base url is different than expected")
}

func checkAbortedDueToDifferentTags(c check) {
	require.Greater(c.test, len(c.output), 0)
	require.Contains(c.test, c.output[len(c.output)-1], "You are trying to install with different tags than in your configuration file!")
}

func PathHasPermissions(t *testing.T, path string, perms uint32) {
	info, err := os.Stat(path)
	require.NoError(t, err)
	expected := fs.FileMode(perms)
	got := info.Mode().Perm()
	require.Equal(t, expected, got, "%s should have %o permissions but has %o", path, expected, got)
}

func PathHasUserACL(t *testing.T, path string, ownerName string, perms string) {
	cmd := exec.Command("/usr/bin/getfacl", path)

	output, err := cmd.Output()
	require.NoError(t, err, "error while checking "+path+" acl")
	require.Contains(t, string(output), "user:"+ownerName+":"+perms)
}
