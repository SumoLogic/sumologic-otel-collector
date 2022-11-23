//go:build linux && amd64

package sumologic_scripts_tests

import (
	"os/exec"
	"testing"

	"github.com/stretchr/testify/require"
)

func checkBinaryIsFIPS(c check) {
	cmd := exec.Command(binaryPath, "--version")

	output, err := cmd.Output()
	require.NoError(c.test, err, "error while checking version")
	require.Contains(c.test, string(output), "fips")
}

func TestInstallScriptLinuxAmd64(t *testing.T) {
	for _, spec := range []testSpec{
		{
			name: "download only fips",
			options: installOptions{
				downloadOnly: true,
				fips:         true,
			},
			preChecks: []checkFunc{checkBinaryNotCreated, checkConfigNotCreated, checkUserConfigNotCreated, checkUserNotExists},
			postChecks: []checkFunc{
				checkBinaryCreated,
				checkBinaryIsFIPS,
				checkConfigNotCreated,
				checkUserConfigNotCreated,
				checkSystemdConfigNotCreated,
				checkUserNotExists,
			},
		},
	} {
		t.Run(spec.name, func(t *testing.T) {
			runTest(t, &spec)
		})
	}
}
