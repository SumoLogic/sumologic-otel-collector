package sumologic_scripts_tests

import (
	"os"
	"os/user"
	"testing"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func checkACLAvailability(c check) bool {
	return assert.FileExists(&testing.T{}, "/usr/bin/getfacl", "File ACLS is not supported")
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

func checkTokenEnvFileCreated(c check) {
	require.FileExists(c.test, tokenEnvFilePath, "env token file has not been created")
}

func checkTokenEnvFileNotCreated(c check) {
	require.NoFileExists(c.test, tokenEnvFilePath, "env token file not been created")
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

func preActionMockSystemdStructure(c check) {
	preActionMockStructure(c)

	_, err := os.Create(systemdPath)
	require.NoError(c.test, err)
}
