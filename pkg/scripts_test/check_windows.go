package sumologic_scripts_tests

import (
	"os/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func checkAbortedDueToNoToken(c check) {
	require.Greater(c.test, len(c.output), 1)
	require.Greater(c.test, len(c.errorOutput), 1)
	require.Contains(c.test, c.errorOutput[0], "Installation token has not been provided.")
	require.Contains(c.test, c.errorOutput[1], "Please set the SUMOLOGIC_INSTALLATION_TOKEN environment variable.")
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

func checkUserExists(c check) {
	_, err := user.Lookup(systemUser)
	require.NoError(c.test, err, "user has not been created")
}

func checkUserNotExists(c check) {
	_, err := user.Lookup(systemUser)
	require.Error(c.test, err, "user has been created")
}
