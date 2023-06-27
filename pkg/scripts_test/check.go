//go:build !windows

package sumologic_scripts_tests

import (
	"io/fs"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
	"syscall"
	"testing"

	"github.com/stretchr/testify/require"
)

type check struct {
	test                *testing.T
	installOptions      installOptions
	installType         installType
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

func checkRun(c check) {
	require.Equal(c.test, c.expectedInstallCode, c.code, "unexpected installation script error code")
}

func checkConfigFilesOwnershipAndPermissions(c check) {
	ownerName := getConfigFilesOwner(c)
	groupName := getConfigFilesGroup(c)
	etcPathGlob := filepath.Join(etcPath, "*")
	etcPathNestedGlob := filepath.Join(etcPath, "*", "*")

	for _, glob := range []string{etcPathGlob, etcPathNestedGlob} {
		paths, err := filepath.Glob(glob)
		require.NoError(c.test, err)
		for _, path := range paths {
			var permissions uint32
			info, err := os.Stat(path)
			require.NoError(c.test, err)
			if info.IsDir() {
				permissions = configPathDirPermissions
			} else {
				switch path {
				case configPath:
					// /etc/otelcol-sumo/sumologic.yaml
					permissions = configPathFilePermissions
				case userConfigPath:
					// /etc/otelcol-sumo/conf.d/common.yaml
					permissions = commonConfigPathFilePermissions
				case tokenEnvFilePath:
					// /etc/otelcol-sumo/env/token.env
					permissions = commonConfigPathFilePermissions
				default:
					// /etc/otelcol-sumo/conf.d/**/
					permissions = confDPathFilePermissions
				}
			}
			PathHasPermissions(c.test, path, permissions)
			PathHasOwner(c.test, configPath, ownerName, groupName)
		}
	}
	PathHasPermissions(c.test, configPath, configPathFilePermissions)
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

func checkHostmetricsOwnershipAndPermissions(c check) {
	ownerName := getConfigFilesOwner(c)
	groupName := getConfigFilesGroup(c)
	PathHasOwner(c.test, hostmetricsConfigPath, ownerName, groupName)
	PathHasPermissions(c.test, hostmetricsConfigPath, confDPathFilePermissions)
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

func checkAbortedDueToNoToken(c check) {
	require.Greater(c.test, len(c.output), 1)
	require.Contains(c.test, c.output[len(c.output)-2], "Installation token has not been provided. Please set the 'SUMOLOGIC_INSTALLATION_TOKEN' environment variable.")
	require.Contains(c.test, c.output[len(c.output)-1], "You can ignore this requirement by adding '--skip-installation-token argument.")
}

func getConfigFilesOwner(c check) string {
	if c.installOptions.uninstall {
		if c.installType&PACKAGE_INSTALL != 0 {
			return systemUser
		}
		return rootUser
	}
	if c.installOptions.skipSystemd || c.installOptions.skipInstallToken {
		return rootUser
	}
	return systemUser
}

func getConfigFilesGroup(c check) string {
	if c.installOptions.uninstall {
		if c.installType&PACKAGE_INSTALL != 0 {
			return systemGroup
		}
		return rootGroup
	}
	if c.installOptions.skipSystemd || c.installOptions.skipInstallToken {
		return rootGroup
	}
	return systemGroup
}

func preActionInstallPackage(c check) {
	c.installOptions.installToken = installToken
	c.installOptions.uninstall = false
	c.installOptions.purge = false
	c.installOptions.apiBaseURL = mockAPIBaseURL
	c.code, c.output, c.errorOutput, c.err = runScript(c)
}

func preActionInstallPackageVersion(version string) checkFunc {
	return func(c check) {
		c.installOptions.installToken = installToken
		c.installOptions.uninstall = false
		c.installOptions.purge = false
		c.installOptions.version = version
		c.installOptions.apiBaseURL = mockAPIBaseURL
		c.code, c.output, c.errorOutput, c.err = runScript(c)
	}
}

func preActionInstallPreviousVersion(c check) {
	if c.installType&PACKAGE_INSTALL != 0 {
		preActionInstallPackageVersion(previousPackageVersion)(c)
	}
	preActionInstallVersion(previousBinaryVersion)(c)
}

func preActionInstallVersion(version string) checkFunc {
	return func(c check) {
		c.installOptions.installToken = installToken
		c.installOptions.uninstall = false
		c.installOptions.purge = false
		c.installOptions.version = version
		c.installOptions.apiBaseURL = mockAPIBaseURL
		c.code, c.output, c.errorOutput, c.err = runScript(c)
	}
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

func PathHasOwner(t *testing.T, path string, ownerName string, groupName string) {
	info, err := os.Stat(path)
	require.NoError(t, err)

	// get the owning user and group
	stat := info.Sys().(*syscall.Stat_t)
	uid := strconv.FormatUint(uint64(stat.Uid), 10)
	gid := strconv.FormatUint(uint64(stat.Gid), 10)

	usr, err := user.LookupId(uid)
	require.NoError(t, err)

	group, err := user.LookupGroupId(gid)
	require.NoError(t, err)

	require.Equal(t, ownerName, usr.Username, "%s should be owned by user '%s'", path, ownerName)
	require.Equal(t, groupName, group.Name, "%s should be owned by group '%s'", path, groupName)
}

func PathHasUserACL(t *testing.T, path string, ownerName string, perms string) {
	cmd := exec.Command("/usr/bin/getfacl", path)

	output, err := cmd.Output()
	require.NoError(t, err, "error while checking "+path+" acl")
	require.Contains(t, string(output), "user:"+ownerName+":"+perms)
}
