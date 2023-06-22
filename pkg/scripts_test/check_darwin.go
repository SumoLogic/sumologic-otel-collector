package sumologic_scripts_tests

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/stretchr/testify/require"
)

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
					if path == etcPath {
						permissions = etcPathPermissions
					} else {
						permissions = configPathDirPermissions
					}
				} else {
					if filepath.Dir(path) != confDPath || path == userConfigPath {
						permissions = configPathFilePermissions
					} else {
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

func checkUserExists(c check) {
	exists := dsclKeyExistsForPath(c.test, "/Users", systemUser)
	require.True(c.test, exists, "user has not been created")
}

func checkUserNotExists(c check) {
	exists := dsclKeyExistsForPath(c.test, "/Users", systemUser)
	require.False(c.test, exists, "user has been created")
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
