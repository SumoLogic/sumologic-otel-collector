//go:build !windows

package sumologic_scripts_tests

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
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

func checkRun(c check) {
	require.Equal(c.test, c.expectedInstallCode, c.code, "unexpected installation script error code")
}

func checkConfigDirectoryOwnershipAndPermissions(c check) {
	PathHasOwner(c.test, etcPath, "root", rootGroup)
	// TODO: fix mismatch between package permissions & expected permissions
	if runtime.GOOS == "darwin" {
		PathHasPermissions(c.test, etcPath, 0755)
	} else {
		PathHasPermissions(c.test, etcPath, etcPathPermissions)
	}
}

func checkConfigCreated(c check) {
	require.FileExists(c.test, configPath, "configuration has not been created properly")
	checkConfigDirectoryOwnershipAndPermissions(c)
}

func checkConfigFilesOwnershipAndPermissions(ownerName string, ownerGroup string) func(c check) {
	return func(c check) {
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
					// TODO: fix mismatch between package permissions & expected permissions
					if runtime.GOOS == "darwin" {
						permissions = 0770
					} else {
						permissions = configPathDirPermissions
					}
				} else {
					// TODO: fix mismatch between package permissions & expected permissions
					if runtime.GOOS == "darwin" {
						switch path {
						case filepath.Join(etcPath, "sumologic.yaml"):
							permissions = 0440
						case filepath.Join(etcPath, "conf.d", "common.yaml"):
							permissions = 0440
						default:
							permissions = 0644
						}
					} else {
						permissions = configPathFilePermissions
					}
				}
				PathHasPermissions(c.test, path, permissions)
				PathHasOwner(c.test, configPath, ownerName, ownerGroup)
			}
		}
		PathHasPermissions(c.test, configPath, configPathFilePermissions)
	}
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
	}, "invalid value for install token")
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

func checkTokenInConfig(c check) {
	require.NotEmpty(c.test, c.installOptions.installToken, "installation token has not been provided")

	conf, err := getConfig(userConfigPath)
	require.NoError(c.test, err, "error while reading configuration")

	require.Equal(c.test, c.installOptions.installToken, conf.Extensions.Sumologic.InstallToken, "installation token is different than expected")
}

func checkDeprecatedTokenInConfig(c check) {
	require.NotEmpty(c.test, c.installOptions.deprecatedInstallToken, "installation token has not been provided")

	conf, err := getConfig(userConfigPath)
	require.NoError(c.test, err, "error while reading configuration")

	require.Equal(c.test, c.installOptions.deprecatedInstallToken, conf.Extensions.Sumologic.InstallToken, "installation token is different than expected")
}

func checkDifferentTokenInConfig(c check) {
	conf, err := getConfig(userConfigPath)
	require.NoError(c.test, err, "error while reading configuration")

	require.Equal(c.test, "different"+c.installOptions.installToken, conf.Extensions.Sumologic.InstallToken, "installation token is different than expected")
}

func checkHostmetricsConfigCreated(c check) {
	require.FileExists(c.test, hostmetricsConfigPath, "hostmetrics configuration has not been created properly")
}

func checkHostmetricsOwnershipAndPermissions(ownerName string, ownerGroup string) func(c check) {
	return func(c check) {
		PathHasOwner(c.test, hostmetricsConfigPath, ownerName, ownerGroup)
		PathHasPermissions(c.test, hostmetricsConfigPath, configPathFilePermissions)
	}
}

func checkHostmetricsConfigNotCreated(c check) {
	require.NoFileExists(c.test, hostmetricsConfigPath, "hostmetrics configuration has been created")
}

func checkTags(c check) {
	conf, err := getConfig(userConfigPath)
	require.NoError(c.test, err, "error while reading configuration")

	for k, v := range c.installOptions.tags {
		require.Equal(c.test, v, conf.Extensions.Sumologic.Tags[k], "installation token is different than expected")
	}
}

func checkDifferentTags(c check) {
	conf, err := getConfig(userConfigPath)
	require.NoError(c.test, err, "error while reading configuration")

	require.Equal(c.test, "tag", conf.Extensions.Sumologic.Tags["some"], "installation token is different than expected")
}

func preActionMockStructure(c check) {
	preActionMockConfigs(c)

	err := os.MkdirAll(fileStoragePath, os.ModePerm)
	require.NoError(c.test, err)

	_, err = os.Create(binaryPath)
	require.NoError(c.test, err)
}

func preActionMockConfigs(c check) {
	preActionMockConfig(c)
	preActionMockUserConfig(c)
}

func preActionMockConfig(c check) {
	err := os.MkdirAll(etcPath, fs.FileMode(etcPathPermissions))
	require.NoError(c.test, err)

	f, err := os.Create(configPath)
	require.NoError(c.test, err)

	err = f.Chmod(fs.FileMode(configPathFilePermissions))
	require.NoError(c.test, err)
}

func preActionMockEnvFiles(c check) {
	err := os.MkdirAll(envDirectoryPath, fs.FileMode(etcPathPermissions))
	require.NoError(c.test, err)

	f, err := os.Create(configPath)
	require.NoError(c.test, err)

	err = f.Chmod(fs.FileMode(configPathFilePermissions))
	require.NoError(c.test, err)
}

func preActionMockUserConfig(c check) {
	err := os.MkdirAll(confDPath, fs.FileMode(configPathDirPermissions))
	require.NoError(c.test, err)

	f, err := os.Create(userConfigPath)
	require.NoError(c.test, err)

	err = f.Chmod(fs.FileMode(configPathFilePermissions))
	require.NoError(c.test, err)
}

func preActionCreateHomeDirectory(c check) {
	err := os.MkdirAll(libPath, fs.FileMode(etcPathPermissions))
	require.NoError(c.test, err)
}

func preActionWriteTokenToUserConfig(c check) {
	conf, err := getConfig(userConfigPath)
	require.NoError(c.test, err)

	conf.Extensions.Sumologic.InstallToken = c.installOptions.installToken
	err = saveConfig(userConfigPath, conf)
	require.NoError(c.test, err)
}

func preActionWriteDifferentTokenToUserConfig(c check) {
	conf, err := getConfig(userConfigPath)
	require.NoError(c.test, err)

	conf.Extensions.Sumologic.InstallToken = "different" + c.installOptions.installToken
	err = saveConfig(userConfigPath, conf)
	require.NoError(c.test, err)
}

func preActionWriteDifferentTokenToEnvFile(c check) {
	preActionMockEnvFiles(c)

	content := fmt.Sprintf("SUMOLOGIC_INSTALLATION_TOKEN=different%s", c.installOptions.installToken)
	err := os.WriteFile(tokenEnvFilePath, []byte(content), fs.FileMode(etcPathPermissions))
	require.NoError(c.test, err)
}

func preActionWriteDifferentDeprecatedTokenToEnvFile(c check) {
	preActionMockEnvFiles(c)

	content := fmt.Sprintf("SUMOLOGIC_INSTALL_TOKEN=different%s", c.installOptions.installToken)
	err := os.WriteFile(tokenEnvFilePath, []byte(content), fs.FileMode(etcPathPermissions))
	require.NoError(c.test, err)
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

func checkOutputUserAddWarnings(c check) {
	output := strings.Join(c.output, "\n")
	require.NotContains(c.test, output, "useradd", "unexpected useradd output")

	errOutput := strings.Join(c.errorOutput, "\n")
	require.NotContains(c.test, errOutput, "useradd", "unexpected useradd output")
}

func checkDownloadTimeout(c check) {
	output := strings.Join(c.errorOutput, "\n")
	count := strings.Count(output, "Operation timed out after")
	require.Equal(c.test, 6, count)
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

func preActionWriteDefaultAPIBaseURLToUserConfig(c check) {
	conf, err := getConfig(userConfigPath)
	require.NoError(c.test, err)

	conf.Extensions.Sumologic.APIBaseURL = apiBaseURL
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

func preActionWriteDifferentTagsToUserConfig(c check) {
	conf, err := getConfig(userConfigPath)
	require.NoError(c.test, err)

	conf.Extensions.Sumologic.Tags = map[string]string{
		"some": "tag",
	}
	err = saveConfig(userConfigPath, conf)
	require.NoError(c.test, err)
}

func checkAbortedDueToDifferentTags(c check) {
	require.Greater(c.test, len(c.output), 0)
	require.Contains(c.test, c.output[len(c.output)-1], "You are trying to install with different tags than in your configuration file!")
}

// preActionCreateUser creates the system user and then set it as owner of configPath
func preActionCreateUser(c check) {
	preActionMockUserConfig(c)

	cmd := exec.Command("useradd", systemUser)
	_, err := cmd.CombinedOutput()
	require.NoError(c.test, err)

	f, err := os.Open(configPath)
	require.NoError(c.test, err)

	user, err := user.Lookup(systemUser)
	require.NoError(c.test, err)

	uid, err := strconv.Atoi(user.Uid)
	require.NoError(c.test, err)

	gid, err := strconv.Atoi(user.Gid)
	require.NoError(c.test, err)

	err = f.Chown(uid, gid)
	require.NoError(c.test, err)
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
