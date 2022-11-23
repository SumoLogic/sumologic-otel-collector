//go:build darwin

package sumologic_scripts_tests

import (
	"testing"
)

func TestInstallScriptDarwin(t *testing.T) {
	for _, spec := range []testSpec{
		{
			name: "download only fips",
			options: installOptions{
				downloadOnly: true,
				fips:         true,
			},
			preChecks:   []checkFunc{checkBinaryNotCreated, checkConfigNotCreated, checkUserConfigNotCreated, checkUserNotExists},
			postChecks:  []checkFunc{checkBinaryNotCreated, checkConfigNotCreated, checkUserConfigNotCreated, checkSystemdConfigNotCreated, checkUserNotExists},
			installCode: 1,
		},
	} {
		t.Run(spec.name, func(t *testing.T) {
			runTest(t, &spec)
		})
	}
}
