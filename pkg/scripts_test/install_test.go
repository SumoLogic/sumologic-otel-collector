package sumologic_scripts_tests

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func tearDown(t *testing.T) {
	t.Log("Cleaning up")

	err := os.RemoveAll(fileStoragePath)
	assert.NoError(t, err, "no permissions to remove storage directory")

	err = os.RemoveAll(etcPath)
	assert.NoError(t, err, "no permissions to remove configuration")

	err = os.RemoveAll(systemdPath)
	assert.NoError(t, err, "no permissions to remove systemd configuration")

	err = os.RemoveAll(binaryPath)
	assert.NoError(t, err, "removing binary")
}

func TestInstallScript(t *testing.T) {
	for _, tt := range []struct {
		name        string
		options     installOptions
		preChecks   []checkFunc
		postChecks  []checkFunc
		preActions  []checkFunc
		installCode int
	}{
		{
			name:       "no arguments",
			options:    installOptions{},
			preChecks:  []checkFunc{checkBinaryNotCreated},
			postChecks: []checkFunc{checkBinaryCreated, checkBinaryIsRunning, checkConfigNotCreated},
		},
		{
			name:       "autoconfirm",
			options:    installOptions{},
			preChecks:  []checkFunc{checkBinaryNotCreated},
			postChecks: []checkFunc{checkBinaryCreated, checkBinaryIsRunning, checkConfigNotCreated},
		},
		{
			name: "installation token only",
			options: installOptions{
				disableSystemd: true,
				installToken:   installToken,
			},
			preChecks:  []checkFunc{checkBinaryNotCreated},
			postChecks: []checkFunc{checkBinaryCreated, checkBinaryIsRunning, checkConfigCreated, checkTokenInConfig},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			defer tearDown(t)

			ch := check{
				test:                t,
				installOptions:      tt.options,
				expectedInstallCode: tt.installCode,
			}

			for _, a := range tt.preActions {
				a(ch)
			}

			for _, c := range tt.preChecks {
				c(ch)
			}

			ch.code, ch.err = runScript(t, tt.options)
			checkRun(ch)

			for _, c := range tt.postChecks {
				c(ch)
			}

		})
	}
}
