//go:build !windows

package sumologic_scripts_tests

import (
	"testing"
)

func TestInstallScript(t *testing.T) {
	for _, spec := range []testSpec{
		{
			name:        "no arguments",
			options:     installOptions{},
			preChecks:   []checkFunc{checkBinaryNotCreated, checkConfigNotCreated, checkUserConfigNotCreated, checkUserNotExists},
			postChecks:  []checkFunc{checkBinaryNotCreated, checkConfigNotCreated, checkUserConfigNotCreated, checkAbortedDueToNoToken, checkUserNotExists},
			installCode: 1,
		},
		{
			name: "download only",
			options: installOptions{
				downloadOnly: true,
			},
			preChecks:  []checkFunc{checkBinaryNotCreated, checkConfigNotCreated, checkUserConfigNotCreated, checkUserNotExists},
			postChecks: []checkFunc{checkBinaryCreated, checkConfigNotCreated, checkUserConfigNotCreated, checkSystemdConfigNotCreated, checkUserNotExists},
		},
		{
			name: "skip config",
			options: installOptions{
				skipConfig:       true,
				skipInstallToken: true,
			},
			preChecks:  []checkFunc{checkBinaryNotCreated, checkConfigNotCreated, checkUserConfigNotCreated, checkUserNotExists},
			postChecks: []checkFunc{checkBinaryCreated, checkConfigNotCreated, checkUserConfigNotCreated},
		},
		{
			name: "skip installation token",
			options: installOptions{
				skipInstallToken: true,
			},
			preChecks: []checkFunc{checkBinaryNotCreated, checkConfigNotCreated, checkUserConfigNotCreated, checkUserNotExists},
			postChecks: []checkFunc{
				checkBinaryCreated,
				checkBinaryIsRunning,
				checkConfigCreated,
				checkConfigFilesOwnershipAndPermissions("root", getRootGroupName()),
				checkUserConfigNotCreated,
				checkSystemdConfigNotCreated,
			},
		},
		{
			name: "override default config",
			options: installOptions{
				skipInstallToken: true,
				autoconfirm:      true,
			},
			preActions: []checkFunc{preActionMockConfig},
			preChecks:  []checkFunc{checkBinaryNotCreated, checkConfigCreated, checkUserConfigNotCreated, checkUserNotExists},
			postChecks: []checkFunc{checkBinaryCreated, checkBinaryIsRunning, checkConfigCreated, checkConfigOverrided, checkUserConfigNotCreated,
				checkSystemdConfigNotCreated},
		},
		{
			name: "installation token only",
			options: installOptions{
				skipSystemd:  true,
				installToken: installToken,
			},
			preChecks: []checkFunc{checkBinaryNotCreated, checkConfigNotCreated, checkUserConfigNotCreated, checkUserNotExists},
			postChecks: []checkFunc{
				checkBinaryCreated,
				checkBinaryIsRunning,
				checkConfigCreated,
				checkConfigFilesOwnershipAndPermissions("root", getRootGroupName()),
				checkUserConfigCreated,
				checkTokenInConfig,
				checkSystemdConfigNotCreated,
				checkUserNotExists,
				checkHostmetricsConfigNotCreated,
			},
		},
		{
			name: "deprecated installation token only",
			options: installOptions{
				skipSystemd:            true,
				deprecatedInstallToken: installToken,
			},
			preChecks: []checkFunc{checkBinaryNotCreated, checkConfigNotCreated, checkUserConfigNotCreated, checkUserNotExists},
			postChecks: []checkFunc{
				checkBinaryCreated,
				checkBinaryIsRunning,
				checkConfigCreated,
				checkConfigFilesOwnershipAndPermissions("root", getRootGroupName()),
				checkUserConfigCreated,
				checkDeprecatedTokenInConfig,
				checkSystemdConfigNotCreated,
				checkUserNotExists,
				checkHostmetricsConfigNotCreated,
			},
		},
		{
			name: "installation token and hostmetrics",
			options: installOptions{
				skipSystemd:        true,
				installToken:       installToken,
				installHostmetrics: true,
			},
			preChecks: []checkFunc{checkBinaryNotCreated, checkConfigNotCreated, checkUserConfigNotCreated, checkUserNotExists},
			postChecks: []checkFunc{
				checkBinaryCreated,
				checkBinaryIsRunning,
				checkConfigCreated,
				checkConfigFilesOwnershipAndPermissions("root", getRootGroupName()),
				checkUserConfigCreated,
				checkTokenInConfig,
				checkSystemdConfigNotCreated,
				checkUserNotExists,
				checkHostmetricsConfigCreated,
				checkHostmetricsOwnershipAndPermissions("root", getRootGroupName()),
			},
		},
		{
			name: "installation token only, binary not in PATH",
			options: installOptions{
				skipSystemd:  true,
				installToken: installToken,
				envs: map[string]string{
					"PATH": "/sbin:/bin:/usr/sbin:/usr/bin",
				},
			},
			preChecks: []checkFunc{checkBinaryNotCreated, checkConfigNotCreated, checkUserConfigNotCreated, checkUserNotExists},
			postChecks: []checkFunc{
				checkBinaryCreated,
				checkBinaryIsRunning,
				checkConfigCreated,
				checkConfigFilesOwnershipAndPermissions("root", getRootGroupName()),
				checkUserConfigCreated,
				checkTokenInConfig,
				checkSystemdConfigNotCreated,
				checkUserNotExists,
			},
		},
		{
			name: "same installation token",
			options: installOptions{
				skipSystemd:  true,
				installToken: installToken,
			},
			preActions: []checkFunc{preActionMockUserConfig, preActionWriteTokenToUserConfig},
			preChecks:  []checkFunc{checkBinaryNotCreated, checkConfigNotCreated, checkUserConfigCreated, checkUserNotExists},
			postChecks: []checkFunc{checkBinaryCreated, checkBinaryIsRunning, checkConfigCreated, checkUserConfigCreated, checkTokenInConfig, checkSystemdConfigNotCreated},
		},
		{
			name: "different installation token",
			options: installOptions{
				skipSystemd:  true,
				installToken: installToken,
			},
			preActions:  []checkFunc{preActionMockUserConfig, preActionWriteDifferentTokenToUserConfig},
			preChecks:   []checkFunc{checkBinaryNotCreated, checkConfigNotCreated, checkUserConfigCreated, checkUserNotExists},
			postChecks:  []checkFunc{checkBinaryNotCreated, checkConfigNotCreated, checkUserConfigCreated, checkSystemdConfigNotCreated, checkAbortedDueToDifferentToken},
			installCode: 1,
		},
		{
			name: "adding installation token",
			options: installOptions{
				skipSystemd:  true,
				installToken: installToken,
			},
			preActions: []checkFunc{preActionMockUserConfig},
			preChecks:  []checkFunc{checkBinaryNotCreated, checkConfigNotCreated, checkUserConfigCreated, checkUserNotExists},
			postChecks: []checkFunc{checkBinaryCreated, checkConfigCreated, checkUserConfigCreated, checkTokenInConfig, checkSystemdConfigNotCreated},
		},
		{
			name: "editing installation token",
			options: installOptions{
				skipSystemd:  true,
				apiBaseURL:   apiBaseURL,
				installToken: installToken,
			},
			preActions: []checkFunc{preActionMockUserConfig, preActionWriteEmptyUserConfig},
			preChecks:  []checkFunc{checkBinaryNotCreated, checkConfigNotCreated, checkUserConfigCreated, checkUserNotExists},
			postChecks: []checkFunc{checkBinaryCreated, checkConfigCreated, checkUserConfigCreated, checkTokenInConfig, checkSystemdConfigNotCreated},
		},
		{
			name: "same api base url",
			options: installOptions{
				skipSystemd:      true,
				apiBaseURL:       apiBaseURL,
				skipInstallToken: true,
			},
			preActions: []checkFunc{preActionMockUserConfig, preActionWriteAPIBaseURLToUserConfig},
			preChecks:  []checkFunc{checkBinaryNotCreated, checkConfigNotCreated, checkUserConfigCreated, checkUserNotExists},
			postChecks: []checkFunc{checkBinaryCreated, checkBinaryIsRunning, checkConfigCreated, checkUserConfigCreated, checkAPIBaseURLInConfig,
				checkSystemdConfigNotCreated},
		},
		{
			name: "different api base url",
			options: installOptions{
				skipSystemd:      true,
				apiBaseURL:       apiBaseURL,
				skipInstallToken: true,
			},
			preActions: []checkFunc{preActionMockUserConfig, preActionWriteDifferentAPIBaseURLToUserConfig},
			preChecks:  []checkFunc{checkBinaryNotCreated, checkConfigNotCreated, checkUserConfigCreated, checkUserNotExists},
			postChecks: []checkFunc{checkBinaryNotCreated, checkConfigNotCreated, checkUserConfigCreated, checkSystemdConfigNotCreated,
				checkAbortedDueToDifferentAPIBaseURL},
			installCode: 1,
		},
		{
			name: "adding api base url",
			options: installOptions{
				skipSystemd:      true,
				apiBaseURL:       apiBaseURL,
				skipInstallToken: true,
			},
			preActions: []checkFunc{preActionMockUserConfig},
			preChecks:  []checkFunc{checkBinaryNotCreated, checkConfigNotCreated, checkUserConfigCreated, checkUserNotExists},
			postChecks: []checkFunc{checkBinaryCreated, checkConfigCreated, checkUserConfigCreated, checkAPIBaseURLInConfig, checkSystemdConfigNotCreated},
		},
		{
			name: "editing api base url",
			options: installOptions{
				skipSystemd:      true,
				apiBaseURL:       apiBaseURL,
				skipInstallToken: true,
			},
			preActions: []checkFunc{preActionMockUserConfig, preActionWriteEmptyUserConfig},
			preChecks:  []checkFunc{checkBinaryNotCreated, checkConfigNotCreated, checkUserConfigCreated, checkUserNotExists},
			postChecks: []checkFunc{checkBinaryCreated, checkConfigCreated, checkUserConfigCreated, checkAPIBaseURLInConfig, checkSystemdConfigNotCreated},
		},
		{
			name: "empty installation token",
			options: installOptions{
				skipSystemd: true,
			},
			preActions: []checkFunc{preActionMockUserConfig, preActionWriteDifferentTokenToUserConfig},
			preChecks:  []checkFunc{checkBinaryNotCreated, checkConfigNotCreated, checkUserConfigCreated, checkUserNotExists},
			postChecks: []checkFunc{checkBinaryCreated, checkConfigCreated, checkUserConfigCreated, checkSystemdConfigNotCreated, checkDifferentTokenInConfig},
		},
		{
			name: "configuration with tags",
			options: installOptions{
				skipSystemd:      true,
				skipInstallToken: true,
				tags: map[string]string{
					"lorem": "ipsum",
					"foo":   "bar",
				},
			},
			preChecks: []checkFunc{checkBinaryNotCreated, checkConfigNotCreated, checkUserConfigNotCreated, checkUserNotExists},
			postChecks: []checkFunc{
				checkBinaryCreated,
				checkBinaryIsRunning,
				checkConfigCreated,
				checkConfigFilesOwnershipAndPermissions("root", getRootGroupName()),
				checkTags,
				checkSystemdConfigNotCreated,
			},
		},
		{
			name: "same tags",
			options: installOptions{
				skipSystemd:      true,
				skipInstallToken: true,
				tags: map[string]string{
					"lorem": "ipsum",
					"foo":   "bar",
				},
			},
			preActions: []checkFunc{preActionMockUserConfig, preActionWriteTagsToUserConfig},
			preChecks:  []checkFunc{checkBinaryNotCreated, checkConfigNotCreated, checkUserConfigCreated, checkUserNotExists},
			postChecks: []checkFunc{checkBinaryCreated, checkBinaryIsRunning, checkConfigCreated, checkUserConfigCreated, checkTags,
				checkSystemdConfigNotCreated},
		},
		{
			name: "different tags",
			options: installOptions{
				skipSystemd:      true,
				skipInstallToken: true,
				tags: map[string]string{
					"lorem": "ipsum",
					"foo":   "bar",
				},
			},
			preActions: []checkFunc{preActionMockUserConfig, preActionWriteDifferentTagsToUserConfig},
			preChecks:  []checkFunc{checkBinaryNotCreated, checkConfigNotCreated, checkUserConfigCreated, checkUserNotExists},
			postChecks: []checkFunc{checkBinaryNotCreated, checkConfigNotCreated, checkUserConfigCreated, checkDifferentTags, checkSystemdConfigNotCreated,
				checkAbortedDueToDifferentTags},
			installCode: 1,
		},
		{
			name: "editing tags",
			options: installOptions{
				skipSystemd:      true,
				skipInstallToken: true,
				tags: map[string]string{
					"lorem": "ipsum",
					"foo":   "bar",
				},
			},
			preActions: []checkFunc{preActionMockUserConfig, preActionWriteEmptyUserConfig},
			preChecks:  []checkFunc{checkBinaryNotCreated, checkConfigNotCreated, checkUserConfigCreated, checkUserNotExists},
			postChecks: []checkFunc{checkBinaryCreated, checkBinaryIsRunning, checkConfigCreated, checkTags, checkSystemdConfigNotCreated},
		},
		{
			name: "systemd",
			options: installOptions{
				installToken: installToken,
			},
			preChecks: []checkFunc{checkBinaryNotCreated, checkConfigNotCreated, checkUserConfigNotCreated, checkUserNotExists},
			postChecks: []checkFunc{
				checkBinaryCreated,
				checkBinaryIsRunning,
				checkConfigCreated,
				checkConfigFilesOwnershipAndPermissions(systemUser, systemUser),
				checkTokenInConfig,
				checkSystemdConfigCreated,
				checkSystemdEnvDirExists,
				checkSystemdEnvDirPermissions,
				checkUserExists,
				checkVarLogACL,
			},
			conditionalChecks: []condCheckFunc{checkSystemdAvailability},
			installCode:       3, // because of invalid installation token
		},
		{
			name: "installation token with existing user directory",
			options: installOptions{
				installToken: installToken,
			},
			preActions: []checkFunc{preActionCreateHomeDirectory},
			preChecks: []checkFunc{
				checkBinaryNotCreated,
				checkConfigNotCreated,
				checkUserConfigNotCreated,
				checkUserNotExists,
				checkHomeDirectoryCreated,
			},
			postChecks: []checkFunc{
				checkBinaryCreated,
				checkBinaryIsRunning,
				checkConfigCreated,
				checkConfigFilesOwnershipAndPermissions(systemUser, systemUser),
				checkUserConfigCreated,
				checkTokenInConfig,
				checkSystemdConfigCreated,
				checkSystemdEnvDirExists,
				checkSystemdEnvDirPermissions,
				checkUserExists,
				checkVarLogACL,
				checkOutputUserAddWarnings,
			},
			conditionalChecks: []condCheckFunc{checkSystemdAvailability},
			installCode:       3, // because of invalid install token
		},
		{
			name: "installation of hostmetrics in systemd during upgrade",
			options: installOptions{
				installToken:       installToken,
				installHostmetrics: true,
			},
			preActions:        []checkFunc{preActionMockSystemdStructure, preActionCreateUser},
			conditionalChecks: []condCheckFunc{checkSystemdAvailability},
			preChecks:         []checkFunc{checkBinaryCreated, checkConfigCreated, checkUserConfigCreated, checkUserExists},
			postChecks: []checkFunc{
				checkBinaryCreated,
				checkBinaryIsRunning,
				checkConfigCreated,
				checkUserConfigCreated,
				checkTokenInConfig,
				checkUserExists,
				checkHostmetricsConfigCreated,
				checkHostmetricsOwnershipAndPermissions(systemUser, systemUser),
			},
		},
		{
			name: "uninstallation without autoconfirm fails",
			options: installOptions{
				uninstall: true,
			},
			installCode: 1,
			preActions:  []checkFunc{preActionMockStructure},
			preChecks:   []checkFunc{checkBinaryCreated, checkConfigCreated, checkUserConfigCreated, checkUserNotExists},
			postChecks:  []checkFunc{checkBinaryCreated, checkConfigCreated, checkUserConfigCreated},
		},
		{
			name: "uninstallation with autoconfirm",
			options: installOptions{
				autoconfirm: true,
				uninstall:   true,
			},
			preActions: []checkFunc{preActionMockStructure},
			preChecks:  []checkFunc{checkBinaryCreated, checkConfigCreated, checkUserConfigCreated, checkUserNotExists},
			postChecks: []checkFunc{checkBinaryNotCreated, checkConfigCreated, checkUserConfigCreated, checkUninstallationOutput},
		},
		{
			name: "systemd uninstallation",
			options: installOptions{
				autoconfirm: true,
				uninstall:   true,
			},
			preActions:        []checkFunc{preActionMockSystemdStructure},
			preChecks:         []checkFunc{checkBinaryCreated, checkConfigCreated, checkUserConfigCreated, checkSystemdConfigCreated, checkUserNotExists},
			postChecks:        []checkFunc{checkBinaryNotCreated, checkConfigCreated, checkUserConfigCreated, checkSystemdConfigCreated, checkUserNotExists},
			conditionalChecks: []condCheckFunc{checkSystemdAvailability},
		},
		{
			name: "purge",
			options: installOptions{
				uninstall:   true,
				purge:       true,
				autoconfirm: true,
			},
			preActions: []checkFunc{preActionMockStructure},
			preChecks:  []checkFunc{checkBinaryCreated, checkConfigCreated, checkUserConfigCreated, checkUserNotExists},
			postChecks: []checkFunc{checkBinaryNotCreated, checkConfigNotCreated, checkUserConfigNotCreated},
		},
		{
			name: "systemd purge",
			options: installOptions{
				uninstall:   true,
				purge:       true,
				autoconfirm: true,
			},
			preActions:        []checkFunc{preActionMockSystemdStructure},
			preChecks:         []checkFunc{checkBinaryCreated, checkConfigCreated, checkUserConfigCreated, checkSystemdConfigCreated, checkUserNotExists},
			postChecks:        []checkFunc{checkBinaryNotCreated, checkConfigNotCreated, checkUserConfigNotCreated, checkSystemdConfigNotCreated, checkUserNotExists},
			conditionalChecks: []condCheckFunc{checkSystemdAvailability},
		},
		{
			name:       "systemd creation if token in file",
			options:    installOptions{},
			preActions: []checkFunc{preActionMockUserConfig, preActionWriteDifferentTokenToUserConfig, preActionWriteDefaultAPIBaseURLToUserConfig},
			preChecks:  []checkFunc{checkBinaryNotCreated, checkConfigNotCreated, checkUserConfigCreated, checkUserNotExists},
			postChecks: []checkFunc{checkBinaryCreated, checkBinaryIsRunning, checkConfigCreated, checkDifferentTokenInConfig, checkSystemdConfigCreated,
				checkUserExists},
			conditionalChecks: []condCheckFunc{checkSystemdAvailability},
			installCode:       3, // because of invalid installation token
		},
		{
			name: "don't keep downloads",
			options: installOptions{
				skipInstallToken:  true,
				dontKeepDownloads: true,
			},
			preChecks:  []checkFunc{checkBinaryNotCreated, checkConfigNotCreated, checkUserConfigNotCreated, checkUserNotExists},
			postChecks: []checkFunc{checkBinaryCreated, checkBinaryIsRunning, checkConfigCreated, checkUserConfigNotCreated, checkSystemdConfigNotCreated},
		},
	} {
		t.Run(spec.name, func(t *testing.T) {
			runTest(t, &spec)
		})
	}
}
