package sumologic_scripts_tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/windows"
)

func checkAbortedDueToNoToken(c check) {
	require.Greater(c.test, len(c.output), 1)
	require.Greater(c.test, len(c.errorOutput), 1)
	// The exact formatting of the error message can be different depending on Powershell version
	errorOutput := strings.Join(c.errorOutput, " ")
	require.Contains(c.test, errorOutput, "Installation token has not been provided.")
	require.Contains(c.test, errorOutput, "Please set the SUMOLOGIC_INSTALLATION_TOKEN environment variable.")
}

func checkEphemeralNotInConfig(p string) func(c check) {
	return func(c check) {
		assert.False(c.test, c.installOptions.ephemeral, "ephemeral was specified")

		conf, err := getConfig(p)
		require.NoError(c.test, err, "error while reading configuration")

		assert.False(c.test, conf.Extensions.Sumologic.Ephemeral, "ephemeral is true")
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

func checkConfigFilesOwnershipAndPermissions(ownerSid string) func(c check) {
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
				PathHasOwner(c.test, configPath, ownerSid)
			}
		}
		PathHasPermissions(c.test, configPath, configPathFilePermissions)
	}
}

func PathHasOwner(t *testing.T, path string, ownerSID string) {
	securityDescriptor, err := windows.GetNamedSecurityInfo(
		path,
		windows.SE_FILE_OBJECT,
		windows.OWNER_SECURITY_INFORMATION,
	)
	require.NoError(t, err)

	// get the owning user
	owner, _, err := securityDescriptor.Owner()
	require.NoError(t, err)

	require.Equal(t, ownerSID, owner.String(), "%s should be owned by user '%s'", path, ownerSID)
}
