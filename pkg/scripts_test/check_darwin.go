package sumologic_scripts_tests

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func checkConfigFilesOwnershipAndPermissions(ownerName string, ownerGroup string) func(c check) {
	return func(c check) {
		PathHasPermissions(c.test, etcPath, etcPathPermissions)
		PathHasOwner(c.test, etcPath, ownerName, ownerGroup)

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
					switch path {
					case etcPath:
						permissions = etcPathPermissions
					case opampDPath:
						// /etc/otelcol-sumo/opamp.d
						permissions = opampDPermissions
					default:
						permissions = configPathDirPermissions
					}
				} else {
					switch path {
					case configPath:
						// /etc/otelcol-sumo/sumologic.yaml
						permissions = configPathFilePermissions
					case userConfigPath:
						// /etc/otelcol-sumo/conf.d/common.yaml
						permissions = commonConfigPathFilePermissions
					default:
						// /etc/otelcol-sumo/conf.d/*
						permissions = confDPathFilePermissions
					}
				}
				PathHasPermissions(c.test, path, permissions)
				PathHasOwner(c.test, configPath, ownerName, ownerGroup)
			}
		}
		PathHasPermissions(c.test, configPath, configPathFilePermissions)
	}
}

func checkDifferentTokenInLaunchdConfig(c check) {
	require.NotEmpty(c.test, c.installOptions.installToken, "installation token has not been provided")

	conf, err := getLaunchdConfig(launchdPath)
	require.NoError(c.test, err)

	require.Equal(c.test, "different"+c.installOptions.installToken, conf.EnvironmentVariables.InstallationToken, "installation token is different than expected")
}

func checkGroupExists(c check) {
	exists := dsclKeyExistsForPath(c.test, "/Groups", systemGroup)
	require.True(c.test, exists, "group has not been created")
}

func checkGroupNotExists(c check) {
	exists := dsclKeyExistsForPath(c.test, "/Groups", systemGroup)
	require.False(c.test, exists, "group has been created")
}

func checkHostmetricsOwnershipAndPermissions(ownerName string, ownerGroup string) func(c check) {
	return func(c check) {
		PathHasOwner(c.test, hostmetricsConfigPath, ownerName, ownerGroup)
		PathHasPermissions(c.test, hostmetricsConfigPath, confDPathFilePermissions)
	}
}

func checkLaunchdConfigCreated(c check) {
	require.FileExists(c.test, launchdPath, "launchd configuration has not been created properly")
}

func checkLaunchdConfigNotCreated(c check) {
	require.NoFileExists(c.test, launchdPath, "launchd configuration has been created")
}

func checkPackageCreated(c check) {
	re, err := regexp.Compile("Package downloaded to: .*/otelcol-sumo.pkg")
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

func checkTokenInLaunchdConfig(c check) {
	require.NotEmpty(c.test, c.installOptions.installToken, "installation token has not been provided")

	conf, err := getLaunchdConfig(launchdPath)
	require.NoError(c.test, err)

	require.Equal(c.test, c.installOptions.installToken, conf.EnvironmentVariables.InstallationToken, "installation token is different than expected")
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

func checkUserExists(c check) {
	exists := dsclKeyExistsForPath(c.test, "/Users", systemUser)
	require.True(c.test, exists, "user has not been created")
}

func checkUserNotExists(c check) {
	exists := dsclKeyExistsForPath(c.test, "/Users", systemUser)
	require.False(c.test, exists, "user has been created")
}

func preActionInstallPackage(c check) {
	c.code, c.output, c.errorOutput, c.err = runScript(c)
}

func preActionInstallPackageWithDifferentAPIBaseURL(c check) {
	c.installOptions.apiBaseURL = "different" + c.installOptions.apiBaseURL
	c.code, c.output, c.errorOutput, c.err = runScript(c)
}

func preActionInstallPackageWithDifferentTags(c check) {
	c.installOptions.tags = map[string]string{
		"some": "tag",
	}
	c.code, c.output, c.errorOutput, c.err = runScript(c)
}

func preActionInstallPackageWithNoAPIBaseURL(c check) {
	c.installOptions.apiBaseURL = ""
	c.code, c.output, c.errorOutput, c.err = runScript(c)
}

func preActionInstallPackageWithNoTags(c check) {
	c.installOptions.tags = nil
	c.code, c.output, c.errorOutput, c.err = runScript(c)
}

func preActionMockLaunchdConfig(c check) {
	f, err := os.Create(launchdPath)
	require.NoError(c.test, err)

	err = f.Chmod(fs.FileMode(launchdPathFilePermissions))
	require.NoError(c.test, err)

	conf := NewLaunchdConfig()
	err = saveLaunchdConfig(launchdPath, conf)
	require.NoError(c.test, err)
}

func preActionWriteDifferentTokenToLaunchdConfig(c check) {
	conf, err := getLaunchdConfig(launchdPath)
	require.NoError(c.test, err)

	conf.EnvironmentVariables.InstallationToken = "different" + c.installOptions.installToken
	err = saveLaunchdConfig(launchdPath, conf)
	require.NoError(c.test, err)
}
