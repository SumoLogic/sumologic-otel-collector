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
		name              string
		options           installOptions
		preChecks         []checkFunc
		postChecks        []checkFunc
		preActions        []checkFunc
		conditionalChecks []condCheckFunc
		installCode       int
	}{
		{
			name:        "no arguments",
			options:     installOptions{},
			preChecks:   []checkFunc{checkBinaryNotCreated, checkConfigNotCreated, checkUserConfigNotCreated},
			postChecks:  []checkFunc{checkBinaryNotCreated, checkConfigNotCreated, checkUserConfigNotCreated},
			installCode: 1,
		},
		{
			name: "skip install token",
			options: installOptions{
				skipInstallToken: true,
			},
			preChecks:  []checkFunc{checkBinaryNotCreated, checkConfigNotCreated, checkUserConfigNotCreated},
			postChecks: []checkFunc{checkBinaryCreated, checkBinaryIsRunning, checkConfigCreated, checkUserConfigNotCreated, checkSystemdConfigNotCreated},
		},
		{
			name: "autoconfirm",
			options: installOptions{
				skipInstallToken: true,
			},
			preChecks:  []checkFunc{checkBinaryNotCreated, checkConfigNotCreated, checkUserConfigNotCreated},
			postChecks: []checkFunc{checkBinaryCreated, checkBinaryIsRunning, checkConfigCreated, checkUserConfigNotCreated, checkSystemdConfigNotCreated},
		},
		{
			name: "override default config",
			options: installOptions{
				skipInstallToken: true,
				autoconfirm:      true,
			},
			preActions: []checkFunc{preActionMockConfig},
			preChecks:  []checkFunc{checkBinaryNotCreated, checkConfigCreated, checkUserConfigNotCreated},
			postChecks: []checkFunc{checkBinaryCreated, checkBinaryIsRunning, checkConfigCreated, checkConfigOverrided, checkUserConfigNotCreated,
				checkSystemdConfigNotCreated},
		},
		{
			name: "installation token only",
			options: installOptions{
				disableSystemd: true,
				installToken:   installToken,
			},
			preChecks:  []checkFunc{checkBinaryNotCreated, checkConfigNotCreated, checkUserConfigNotCreated},
			postChecks: []checkFunc{checkBinaryCreated, checkBinaryIsRunning, checkConfigCreated, checkUserConfigCreated, checkTokenInConfig, checkSystemdConfigNotCreated},
		},
		{
			name: "installation token only (envs)",
			options: installOptions{
				disableSystemd: true,
				envs: map[string]string{
					"SUMOLOGIC_INSTALL_TOKEN": installToken,
				},
			},
			preChecks: []checkFunc{checkBinaryNotCreated, checkConfigNotCreated, checkUserConfigNotCreated},
			postChecks: []checkFunc{
				checkBinaryCreated, checkBinaryIsRunning, checkConfigCreated, checkUserConfigCreated,
				checkEnvTokenInConfig, checkSystemdConfigNotCreated},
		},
		{
			name: "same installation token",
			options: installOptions{
				disableSystemd: true,
				installToken:   installToken,
			},
			preActions: []checkFunc{preActionMockUserConfig, preActionWriteTokenToUserConfig},
			preChecks:  []checkFunc{checkBinaryNotCreated, checkConfigNotCreated, checkUserConfigCreated},
			postChecks: []checkFunc{checkBinaryCreated, checkBinaryIsRunning, checkConfigCreated, checkUserConfigCreated, checkTokenInConfig, checkSystemdConfigNotCreated},
		},
		{
			name: "different installation token",
			options: installOptions{
				disableSystemd: true,
				installToken:   installToken,
			},
			preActions:  []checkFunc{preActionMockUserConfig, preActionWriteDifferentTokenToUserConfig},
			preChecks:   []checkFunc{checkBinaryNotCreated, checkConfigNotCreated, checkUserConfigCreated},
			postChecks:  []checkFunc{checkBinaryNotCreated, checkConfigNotCreated, checkUserConfigCreated, checkSystemdConfigNotCreated, checkAbortedDueToDifferentToken},
			installCode: 1,
		},
		{
			name: "adding installation token",
			options: installOptions{
				disableSystemd: true,
				installToken:   installToken,
			},
			preActions: []checkFunc{preActionMockUserConfig},
			preChecks:  []checkFunc{checkBinaryNotCreated, checkConfigNotCreated, checkUserConfigCreated},
			postChecks: []checkFunc{checkBinaryCreated, checkConfigCreated, checkUserConfigCreated, checkTokenInConfig, checkSystemdConfigNotCreated},
		},
		{
			name: "editing installation token",
			options: installOptions{
				disableSystemd: true,
				apiBaseURL:     apiBaseURL,
			},
			preActions: []checkFunc{preActionMockUserConfig, preActionWriteEmptyUserConfig},
			preChecks:  []checkFunc{checkBinaryNotCreated, checkConfigNotCreated, checkUserConfigCreated},
			postChecks: []checkFunc{checkBinaryCreated, checkConfigCreated, checkUserConfigCreated, checkTokenInConfig, checkSystemdConfigNotCreated},
		},
		{
			name: "same api base url",
			options: installOptions{
				disableSystemd: true,
				apiBaseURL:     apiBaseURL,
			},
			preActions: []checkFunc{preActionMockUserConfig, preActionWriteAPIBaseURLToUserConfig},
			preChecks:  []checkFunc{checkBinaryNotCreated, checkConfigNotCreated, checkUserConfigCreated},
			postChecks: []checkFunc{checkBinaryCreated, checkBinaryIsRunning, checkConfigCreated, checkUserConfigCreated, checkAPIBaseURLInConfig,
				checkSystemdConfigNotCreated},
		},
		{
			name: "different api base url",
			options: installOptions{
				disableSystemd: true,
				apiBaseURL:     apiBaseURL,
			},
			preActions: []checkFunc{preActionMockUserConfig, preActionWriteDifferentAPIBaseURLToUserConfig},
			preChecks:  []checkFunc{checkBinaryNotCreated, checkConfigNotCreated, checkUserConfigCreated},
			postChecks: []checkFunc{checkBinaryNotCreated, checkConfigNotCreated, checkUserConfigCreated, checkSystemdConfigNotCreated,
				checkAbortedDueToDifferentAPIBaseURL},
			installCode: 1,
		},
		{
			name: "adding api base url",
			options: installOptions{
				disableSystemd: true,
				apiBaseURL:     apiBaseURL,
			},
			preActions: []checkFunc{preActionMockUserConfig},
			preChecks:  []checkFunc{checkBinaryNotCreated, checkConfigNotCreated, checkUserConfigCreated},
			postChecks: []checkFunc{checkBinaryCreated, checkConfigCreated, checkUserConfigCreated, checkAPIBaseURLInConfig, checkSystemdConfigNotCreated},
		},
		{
			name: "editing api base url",
			options: installOptions{
				disableSystemd: true,
				apiBaseURL:     apiBaseURL,
			},
			preActions: []checkFunc{preActionMockUserConfig, preActionWriteEmptyUserConfig},
			preChecks:  []checkFunc{checkBinaryNotCreated, checkConfigNotCreated, checkUserConfigCreated},
			postChecks: []checkFunc{checkBinaryCreated, checkConfigCreated, checkUserConfigCreated, checkAPIBaseURLInConfig, checkSystemdConfigNotCreated},
		},
		{
			name: "empty installation token",
			options: installOptions{
				disableSystemd: true,
			},
			preActions: []checkFunc{preActionMockUserConfig, preActionWriteDifferentTokenToUserConfig},
			preChecks:  []checkFunc{checkBinaryNotCreated, checkConfigNotCreated, checkUserConfigCreated},
			postChecks: []checkFunc{checkBinaryCreated, checkConfigCreated, checkUserConfigCreated, checkSystemdConfigNotCreated, checkDifferentTokenInConfig},
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
			preChecks:  []checkFunc{checkBinaryNotCreated, checkConfigNotCreated, checkUserConfigNotCreated},
			postChecks: []checkFunc{checkBinaryCreated, checkBinaryIsRunning, checkConfigCreated, checkTags, checkSystemdConfigNotCreated},
		},
		{
			name: "same tags",
			options: installOptions{
				disableSystemd: true,
				tags: map[string]string{
					"lorem": "ipsum",
					"foo":   "bar",
				},
			},
			preActions: []checkFunc{preActionMockUserConfig, preActionWriteTagsToUserConfig},
			preChecks:  []checkFunc{checkBinaryNotCreated, checkConfigNotCreated, checkUserConfigCreated},
			postChecks: []checkFunc{checkBinaryCreated, checkBinaryIsRunning, checkConfigCreated, checkUserConfigCreated, checkTags,
				checkSystemdConfigNotCreated},
		},
		{
			name: "different tags",
			options: installOptions{
				disableSystemd: true,
				tags: map[string]string{
					"lorem": "ipsum",
					"foo":   "bar",
				},
			},
			preActions: []checkFunc{preActionMockUserConfig, preActionWriteDifferentTagsToUserConfig},
			preChecks:  []checkFunc{checkBinaryNotCreated, checkConfigNotCreated, checkUserConfigCreated},
			postChecks: []checkFunc{checkBinaryNotCreated, checkConfigNotCreated, checkUserConfigCreated, checkDifferentTags, checkSystemdConfigNotCreated,
				checkAbortedDueToDifferentTags},
			installCode: 1,
		},
		{
			name: "editing tags",
			options: installOptions{
				disableSystemd: true,
				tags: map[string]string{
					"lorem": "ipsum",
					"foo":   "bar",
				},
			},
			preActions: []checkFunc{preActionMockUserConfig, preActionWriteEmptyUserConfig},
			preChecks:  []checkFunc{checkBinaryNotCreated, checkConfigNotCreated, checkUserConfigCreated},
			postChecks: []checkFunc{checkBinaryCreated, checkBinaryIsRunning, checkConfigCreated, checkTags, checkSystemdConfigNotCreated},
		},
		{
			name: "systemd",
			options: installOptions{
				installToken: installToken,
			},
			preChecks:         []checkFunc{checkBinaryNotCreated, checkConfigNotCreated, checkUserConfigNotCreated},
			postChecks:        []checkFunc{checkBinaryCreated, checkBinaryIsRunning, checkConfigCreated, checkTokenInConfig, checkSystemdConfigCreated},
			conditionalChecks: []condCheckFunc{checkSystemdAvailability},
			installCode:       3, // because of invalid install token
		},
		{
			name: "uninstallation",
			options: installOptions{
				uninstall: true,
			},
			preActions: []checkFunc{preActionMockStructure},
			preChecks:  []checkFunc{checkBinaryCreated, checkConfigCreated, checkUserConfigCreated},
			postChecks: []checkFunc{checkBinaryNotCreated, checkConfigCreated, checkUserConfigCreated},
		},
		{
			name: "systemd uninstallation",
			options: installOptions{
				uninstall: true,
			},
			preActions:        []checkFunc{preActionMockSystemdStructure},
			preChecks:         []checkFunc{checkBinaryCreated, checkConfigCreated, checkUserConfigCreated, checkSystemdConfigCreated},
			postChecks:        []checkFunc{checkBinaryNotCreated, checkConfigCreated, checkUserConfigCreated, checkSystemdConfigCreated},
			conditionalChecks: []condCheckFunc{checkSystemdAvailability},
		},
		{
			name: "purge",
			options: installOptions{
				uninstall: true,
				purge:     true,
			},
			preActions: []checkFunc{preActionMockStructure},
			preChecks:  []checkFunc{checkBinaryCreated, checkConfigCreated, checkUserConfigCreated},
			postChecks: []checkFunc{checkBinaryNotCreated, checkConfigNotCreated, checkUserConfigNotCreated},
		},
		{
			name: "systemd purge",
			options: installOptions{
				uninstall: true,
				purge:     true,
			},
			preActions:        []checkFunc{preActionMockSystemdStructure},
			preChecks:         []checkFunc{checkBinaryCreated, checkConfigCreated, checkUserConfigCreated, checkSystemdConfigCreated},
			postChecks:        []checkFunc{checkBinaryNotCreated, checkConfigNotCreated, checkUserConfigNotCreated, checkSystemdConfigNotCreated},
			conditionalChecks: []condCheckFunc{checkSystemdAvailability},
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

			ch.code, ch.output, ch.err = runScript(ch)
			checkRun(ch)

			for _, c := range tt.postChecks {
				c(ch)
			}

		})
	}
}
