package sumologic_scripts_tests

import (
	"fmt"
	"io/fs"
	"os"
	"os/user"
	"regexp"
	"strings"
	"testing"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func checkAbortedDueToSkipConfigUnsupported(c check) {
	require.Greater(c.test, len(c.output), 0)
	require.Contains(c.test, c.output[len(c.output)-1], "SKIP_CONFIG is not supported")
}

func checkACLAvailability(c check) bool {
	return assert.FileExists(&testing.T{}, "/usr/bin/getfacl", "File ACLS is not supported")
}

func checkDifferentTokenInConfig(c check) {
	conf, err := getConfig(userConfigPath)
	require.NoError(c.test, err, "error while reading configuration")

	require.Equal(c.test, "different"+c.installOptions.installToken, conf.Extensions.Sumologic.InstallToken, "installation token is different than expected")
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

func checkGroupExists(c check) {
	_, err := user.LookupGroup(systemGroup)
	require.NoError(c.test, err, "group has not been created")
}

func checkGroupNotExists(c check) {
	_, err := user.LookupGroup(systemGroup)
	require.Error(c.test, err, "group has been created")
}

func checkOutputUserAddWarnings(c check) {
	output := strings.Join(c.output, "\n")
	require.NotContains(c.test, output, "useradd", "unexpected useradd output")

	errOutput := strings.Join(c.errorOutput, "\n")
	require.NotContains(c.test, errOutput, "useradd", "unexpected useradd output")
}

func checkPackageCreated(c check) {
	re, err := regexp.Compile(`^Package downloaded to: .*/otelcol\-sumo(_|\.).*.(deb|rpm)$`)
	require.NoError(c.test, err)

	matchedLine := ""
	for _, line := range c.output {
		if re.MatchString(line) {
			matchedLine = line
		}
	}
	require.NotEmpty(c.test, matchedLine, "package path not in output")

	packagePath := strings.TrimPrefix(matchedLine, "Package downloaded to: ")
	require.FileExists(c.test, packagePath, "package has not been created")
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

func checkSystemdConfigOrBackupCreated(c check) {
	if configOrBackupMissing(tokenEnvFilePath) {
		c.test.Fatalf("unable to find file: \"%s\"", tokenEnvFilePath)
	}
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

func checkTokenInConfig(c check) {
	require.NotEmpty(c.test, c.installOptions.installToken, "installation token has not been provided")

	conf, err := getConfig(userConfigPath)
	require.NoError(c.test, err, "error while reading configuration")

	require.Equal(c.test, c.installOptions.installToken, conf.Extensions.Sumologic.InstallToken, "installation token is different than expected")
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

func checkUserConfigOrBackupCreated(c check) {
	if configOrBackupMissing(userConfigPath) {
		c.test.Fatalf("unable to find file: \"%s\"", userConfigPath)
	}
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

func configOrBackupMissing(p string) bool {
	backupPath := fmt.Sprintf("%s.rpmsave", p)
	_, cErr := os.Stat(p)
	_, bErr := os.Stat(backupPath)
	return os.IsNotExist(cErr) && os.IsNotExist(bErr)
}

func preActionCreateHomeDirectory(c check) {
	err := os.MkdirAll(libPath, fs.FileMode(etcPathPermissions))
	require.NoError(c.test, err)
}

func preActionInstallPackageWithOptions(o installOptions) checkFunc {
	return func(c check) {
		o.useNativePackaging = true
		c.installOptions = o
		c.code, c.output, c.errorOutput, c.err = runScript(c)
	}
}

func preActionMockConfig(c check) {
	err := os.MkdirAll(etcPath, fs.FileMode(etcPathPermissions))
	require.NoError(c.test, err)

	f, err := os.Create(configPath)
	require.NoError(c.test, err)

	err = f.Chmod(fs.FileMode(configPathFilePermissions))
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

	_, err = os.Create(binaryPath)
	require.NoError(c.test, err)
}

func preActionMockSystemdStructure(c check) {
	preActionMockStructure(c)

	_, err := os.Create(systemdPath)
	require.NoError(c.test, err)
}

func preActionMockUserConfig(c check) {
	err := os.MkdirAll(etcPath, fs.FileMode(etcPathPermissions))
	require.NoError(c.test, err)

	err = os.MkdirAll(confDPath, fs.FileMode(configPathDirPermissions))
	require.NoError(c.test, err)

	f, err := os.Create(userConfigPath)
	require.NoError(c.test, err)

	err = f.Chmod(fs.FileMode(commonConfigPathFilePermissions))
	require.NoError(c.test, err)
}

func preActionWriteAPIBaseURLToUserConfig(c check) {
	conf, err := getConfig(userConfigPath)
	require.NoError(c.test, err)

	conf.Extensions.Sumologic.APIBaseURL = c.installOptions.apiBaseURL
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

func preActionWriteDifferentAPIBaseURLToUserConfig(c check) {
	conf, err := getConfig(userConfigPath)
	require.NoError(c.test, err)

	conf.Extensions.Sumologic.APIBaseURL = "different" + c.installOptions.apiBaseURL
	err = saveConfig(userConfigPath, conf)
	require.NoError(c.test, err)
}

func preActionWriteDifferentDeprecatedTokenToEnvFile(c check) {
	preActionMockEnvFiles(c)

	content := fmt.Sprintf("SUMOLOGIC_INSTALL_TOKEN=different%s", c.installOptions.installToken)
	err := os.WriteFile(tokenEnvFilePath, []byte(content), fs.FileMode(etcPathPermissions))
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

func preActionWriteDifferentTokenToEnvFile(c check) {
	preActionMockEnvFiles(c)

	content := fmt.Sprintf("SUMOLOGIC_INSTALLATION_TOKEN=different%s", c.installOptions.installToken)
	err := os.WriteFile(tokenEnvFilePath, []byte(content), fs.FileMode(etcPathPermissions))
	require.NoError(c.test, err)
}

func preActionWriteDifferentTokenToUserConfig(c check) {
	conf, err := getConfig(userConfigPath)
	require.NoError(c.test, err)

	conf.Extensions.Sumologic.InstallToken = "different" + c.installOptions.installToken
	err = saveConfig(userConfigPath, conf)
	require.NoError(c.test, err)
}

func preActionWriteEmptyUserConfig(c check) {
	err := saveConfig(userConfigPath, config{})
	require.NoError(c.test, err)
}

func preActionWriteTagsToUserConfig(c check) {
	conf, err := getConfig(userConfigPath)
	require.NoError(c.test, err)

	conf.Extensions.Sumologic.Tags = c.installOptions.tags
	err = saveConfig(userConfigPath, conf)
	require.NoError(c.test, err)
}

func preActionWriteTokenToUserConfig(c check) {
	conf, err := getConfig(userConfigPath)
	require.NoError(c.test, err)

	conf.Extensions.Sumologic.InstallToken = c.installOptions.installToken
	err = saveConfig(userConfigPath, conf)
	require.NoError(c.test, err)
}
