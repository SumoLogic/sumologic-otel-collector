package sumologic_scripts_tests

import (
	"fmt"
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
		name              string
		options           installOptions
		preChecks         []checkFunc
		postChecks        []checkFunc
		preActions        []checkFunc
		conditionalChecks []condCheckFunc
		installCode       int
	}{
		{
			name:       "no arguments",
			options:    installOptions{},
			preChecks:  []checkFunc{checkBinaryNotCreated},
			postChecks: []checkFunc{checkBinaryCreated, checkBinaryIsRunning, checkConfigNotCreated, checkSystemdConfigNotCreated},
		},
		{
			name:       "autoconfirm",
			options:    installOptions{},
			preChecks:  []checkFunc{checkBinaryNotCreated},
			postChecks: []checkFunc{checkBinaryCreated, checkBinaryIsRunning, checkConfigNotCreated, checkSystemdConfigNotCreated},
		},
		{
			name: "installation token only",
			options: installOptions{
				disableSystemd: true,
				installToken:   installToken,
			},
			preChecks:  []checkFunc{checkBinaryNotCreated},
			postChecks: []checkFunc{checkBinaryCreated, checkBinaryIsRunning, checkConfigCreated, checkTokenInConfig, checkSystemdConfigNotCreated},
		},
		{
			name: "configuration with tags",
			options: installOptions{
				disableSystemd: true,
				installToken:   installToken,
				tags: map[string]string{
					"lorem": "ipsum",
					"foo":   "bar",
				},
			},
			preChecks:  []checkFunc{checkBinaryNotCreated},
			postChecks: []checkFunc{checkBinaryCreated, checkBinaryIsRunning, checkConfigCreated, checkTags, checkSystemdConfigNotCreated},
		},
		{
			name: "systemd",
			options: installOptions{
				installToken: installToken,
			},
			preChecks:         []checkFunc{checkBinaryNotCreated},
			postChecks:        []checkFunc{checkBinaryCreated, checkBinaryIsRunning, checkConfigCreated, checkTokenInConfig, checkSystemdConfigCreated},
			conditionalChecks: []condCheckFunc{checkSystemdAvailability},
			installCode:       3, // because of invalid install token
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			ch := check{
				test:                t,
				installOptions:      tt.options,
				expectedInstallCode: tt.installCode,
			}

			for _, a := range tt.conditionalChecks {
				if !a(ch) {
					t.SkipNow()
				}
			}

			defer tearDown(t)

			for _, a := range tt.preActions {
				a(ch)
			}

			for _, c := range tt.preChecks {
				c(ch)
			}

			ch.code, ch.err = runScript(t, tt.options)
			fmt.Printf("%v:%v", ch.code, ch.expectedInstallCode)
			checkRun(ch)

			for _, c := range tt.postChecks {
				c(ch)
			}

		})
	}
}
