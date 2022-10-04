package sumologic_scripts_tests

import (
	"os/exec"
	"testing"

	"github.com/stretchr/testify/require"
)

type check struct {
	test                *testing.T
	installOptions      installOptions
	code                int
	err                 error
	expectedInstallCode int
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
	require.NoError(c.test, err)

	code, err := exitCode(cmd)
	require.NoError(c.test, err)
	require.Equal(c.test, 0, code)
}

func checkRun(c check) {
	require.Equal(c.test, c.expectedInstallCode, c.code)
}

func checkConfigCreated(c check) {
	require.FileExists(c.test, configPath)
}

func checkConfigNotCreated(c check) {
	require.NoFileExists(c.test, configPath)
}

func checkTokenInConfig(c check) {
	conf, err := getConfig(configPath)
	require.NoError(c.test, err)

	require.Equal(c.test, c.installOptions.installToken, conf.Extensions.Sumologic.InstallToken)
}
