package sumologic_scripts_tests

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func checkACLAvailability(c check) bool {
	return assert.FileExists(&testing.T{}, "/usr/bin/getfacl", "File ACLS is not supported")
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
					if path == opampDPath {
						permissions = opampDPermissions
					} else {
						permissions = configPathDirPermissions
					}
				} else {
					permissions = configPathFilePermissions
				}
				PathHasPermissions(c.test, path, permissions)
				PathHasOwner(c.test, configPath, ownerName, ownerGroup)
			}
		}
		PathHasPermissions(c.test, configPath, configPathFilePermissions)
	}
}

func checkDifferentTokenInConfig(c check) {
	conf, err := getConfig(userConfigPath)
	require.NoError(c.test, err, "error while reading configuration")

	require.Equal(c.test, "different"+c.installOptions.installToken, conf.Extensions.Sumologic.InstallationToken, "installation token is different than expected")
}

func checkDifferentTokenInEnvFile(c check) {
	require.NotEmpty(c.test, c.installOptions.installToken, "installation token has not been provided")

	envs, err := godotenv.Read(tokenEnvFilePath)

	require.NoError(c.test, err)
	if _, ok := envs["SUMOLOGIC_INSTALL_TOKEN"]; ok {
		require.Equal(c.test, "different"+c.installOptions.installToken, envs["SUMOLOGIC_INSTALL_TOKEN"], "installation token is different than expected")
	} else {
		require.Equal(c.test, "different"+c.installOptions.installToken, envs["SUMOLOGIC_INSTALLATION_TOKEN"], "installation token is different than expected")
	}
}

func checkDownloadTimeout(c check) {
	output := strings.Join(c.errorOutput, "\n")
	count := strings.Count(output, "Operation timed out after")
	require.Equal(c.test, 6, count)
}

func checkHostmetricsOwnershipAndPermissions(ownerName string, ownerGroup string) func(c check) {
	return func(c check) {
		PathHasOwner(c.test, hostmetricsConfigPath, ownerName, ownerGroup)
		PathHasPermissions(c.test, hostmetricsConfigPath, configPathFilePermissions)
	}
}

func checkOutputUserAddWarnings(c check) {
	output := strings.Join(c.output, "\n")
	require.NotContains(c.test, output, "useradd", "unexpected useradd output")

	errOutput := strings.Join(c.errorOutput, "\n")
	require.NotContains(c.test, errOutput, "useradd", "unexpected useradd output")
}

func checkSystemdAvailability(c check) bool {
	return assert.DirExists(&testing.T{}, systemdDirectoryPath, "systemd is not supported")
}

func checkSystemdConfigCreated(c check) {
	require.FileExists(c.test, systemdPath, "systemd configuration has not been created properly")
}

func checkSystemdConfigNotCreated(c check) {
	require.NoFileExists(c.test, systemdPath, "systemd configuration has been created")
}

func checkSystemdEnvDirExists(c check) {
	require.DirExists(c.test, etcPath+"/env", "systemd env directory does not exist")
}

func checkSystemdEnvDirPermissions(c check) {
	PathHasPermissions(c.test, etcPath+"/env", configPathDirPermissions)
}

func checkRemoteFlagInSystemdFile(c check) {
	contents, err := getSystemdConfig(systemdPath)

	require.NoError(c.test, err)

	assert.Contains(c.test, contents, "--remote-config")
	assert.NotContains(c.test, contents, "--config")
}

func checkTokenEnvFileCreated(c check) {
	require.FileExists(c.test, tokenEnvFilePath, "env token file has not been created")
}

func checkTokenEnvFileNotCreated(c check) {
	require.NoFileExists(c.test, tokenEnvFilePath, "env token file not been created")
}

func checkTokenInConfig(c check) {
	require.NotEmpty(c.test, c.installOptions.installToken, "installation token has not been provided")

	conf, err := getConfig(userConfigPath)
	require.NoError(c.test, err, "error while reading configuration")

	require.Equal(c.test, c.installOptions.installToken, conf.Extensions.Sumologic.InstallationToken, "installation token is different than expected")
}

func checkTokenInSumoConfig(c check) {
	require.NotEmpty(c.test, c.installOptions.installToken, "installation token has not been provided")

	conf, err := getConfig(configPath)
	require.NoError(c.test, err, "error while reading configuration")

	require.Equal(c.test, c.installOptions.installToken, conf.Extensions.Sumologic.InstallationToken, "installation token is different than expected")
}

func checkTokenInEnvFile(c check) {
	require.NotEmpty(c.test, c.installOptions.installToken, "installation token has not been provided")

	envs, err := godotenv.Read(tokenEnvFilePath)

	require.NoError(c.test, err)
	if _, ok := envs["SUMOLOGIC_INSTALL_TOKEN"]; ok {
		require.Equal(c.test, c.installOptions.installToken, envs["SUMOLOGIC_INSTALL_TOKEN"], "installation token is different than expected")
	} else {
		require.Equal(c.test, c.installOptions.installToken, envs["SUMOLOGIC_INSTALLATION_TOKEN"], "installation token is different than expected")
	}
}

func checkEphemeralInConfig(p string) func(c check) {
	return func(c check) {
		assert.True(c.test, c.installOptions.ephemeral, "ephemeral was not specified")

		conf, err := getConfig(p)
		require.NoError(c.test, err, "error while reading configuration")

		assert.True(c.test, conf.Extensions.Sumologic.Ephemeral, "ephemeral is not true")
	}
}

func checkEphemeralNotInConfig(p string) func(c check) {
	return func(c check) {
		assert.False(c.test, c.installOptions.ephemeral, "ephemeral was specified")

		conf, err := getConfig(p)
		require.NoError(c.test, err, "error while reading configuration")

		assert.False(c.test, conf.Extensions.Sumologic.Ephemeral, "ephemeral is true")
	}
}

func checkUninstallationOutput(c check) {
	require.Greater(c.test, len(c.output), 1)
	require.Contains(c.test, c.output[len(c.output)-1], "Uninstallation completed")
}

func checkUserExists(c check) {
	_, err := user.Lookup(systemUser)
	require.NoError(c.test, err, "user has not been created")
}

func checkUserNotExists(c check) {
	_, err := user.Lookup(systemUser)
	require.Error(c.test, err, "user has been created")
}

func checkVarLogACL(c check) {
	if !checkACLAvailability(c) {
		return
	}

	PathHasUserACL(c.test, "/var/log", systemUser, "r-x")
}

func preActionCreateHomeDirectory(c check) {
	err := os.MkdirAll(libPath, fs.FileMode(etcPathPermissions))
	require.NoError(c.test, err)
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

func preActionMockConfigs(c check) {
	preActionMockConfig(c)
	preActionMockUserConfig(c)
}

func preActionMockEnvFiles(c check) {
	err := os.MkdirAll(envDirectoryPath, fs.FileMode(etcPathPermissions))
	require.NoError(c.test, err)

	f, err := os.Create(configPath)
	require.NoError(c.test, err)

	err = f.Chmod(fs.FileMode(configPathFilePermissions))
	require.NoError(c.test, err)
}

func preActionMockStructure(c check) {
	preActionMockConfigs(c)

	err := os.MkdirAll(fileStoragePath, os.ModePerm)
	require.NoError(c.test, err)

	content := []byte("#!/bin/sh\necho hello world\n")
	err = os.WriteFile(binaryPath, content, 0755)
	require.NoError(c.test, err)
}

func preActionMockSystemdStructure(c check) {
	preActionMockStructure(c)

	_, err := os.Create(systemdPath)
	require.NoError(c.test, err)
}

func preActionWriteDefaultAPIBaseURLToUserConfig(c check) {
	conf, err := getConfig(userConfigPath)
	require.NoError(c.test, err)

	conf.Extensions.Sumologic.APIBaseURL = apiBaseURL
	err = saveConfig(userConfigPath, conf)
	require.NoError(c.test, err)
}

func preActionWriteDifferentDeprecatedTokenToEnvFile(c check) {
	preActionMockEnvFiles(c)

	content := fmt.Sprintf("SUMOLOGIC_INSTALL_TOKEN=different%s", c.installOptions.installToken)
	err := os.WriteFile(tokenEnvFilePath, []byte(content), fs.FileMode(etcPathPermissions))
	require.NoError(c.test, err)
}

func preActionWriteDifferentTokenToEnvFile(c check) {
	preActionMockEnvFiles(c)

	content := fmt.Sprintf("SUMOLOGIC_INSTALLATION_TOKEN=different%s", c.installOptions.installToken)
	err := os.WriteFile(tokenEnvFilePath, []byte(content), fs.FileMode(etcPathPermissions))
	require.NoError(c.test, err)
}

func preActionWriteDifferentTokenToUserConfig(c check) {
	conf, err := getConfig(userConfigPath)
	require.NoError(c.test, err)

	conf.Extensions.Sumologic.InstallationToken = "different" + c.installOptions.installToken
	err = saveConfig(userConfigPath, conf)
	require.NoError(c.test, err)
}

func preActionWriteTokenToUserConfig(c check) {
	conf, err := getConfig(userConfigPath)
	require.NoError(c.test, err)

	conf.Extensions.Sumologic.InstallationToken = c.installOptions.installToken
	err = saveConfig(userConfigPath, conf)
	require.NoError(c.test, err)
}
