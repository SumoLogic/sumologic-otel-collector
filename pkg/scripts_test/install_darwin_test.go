//go:build darwin

package sumologic_scripts_tests

import (
	"testing"
)

func TestInstallScriptDarwin(t *testing.T) {
	notInstalledChecks := []checkFunc{
		checkBinaryNotCreated,
		checkConfigNotCreated,
		checkUserConfigNotCreated,
		checkUserNotExists,
		checkGroupNotExists,
		checkLaunchdConfigNotCreated,
	}

	for _, spec := range []testSpec{
		{
			name:        "no arguments",
			options:     installOptions{},
			preChecks:   notInstalledChecks,
			postChecks:  append(notInstalledChecks, checkAbortedDueToNoToken),
			installCode: 1,
		},
		{
			name: "download only",
			options: installOptions{
				downloadOnly: true,
			},
			preChecks:  notInstalledChecks,
			postChecks: append(notInstalledChecks, checkPackageCreated),
		},
		{
			name: "download only with timeout",
			options: installOptions{
				downloadOnly:      true,
				timeout:           1,
				dontKeepDownloads: true,
			},
			// Skip this test as getting binary in github actions takes less than one second
			conditionalChecks: []condCheckFunc{checkSkipTest},
			preChecks:         notInstalledChecks,
			postChecks:        notInstalledChecks,
			installCode:       curlTimeoutErrorCode,
		},
		{
			name: "download only fips",
			options: installOptions{
				downloadOnly: true,
				fips:         true,
			},
			preChecks:   notInstalledChecks,
			postChecks:  notInstalledChecks,
			installCode: 1,
		},
		{
			// Skip config is not supported on Darwin
			name: "skip config",
			options: installOptions{
				skipConfig:       true,
				skipInstallToken: true,
			},
			preChecks:   notInstalledChecks,
			postChecks:  notInstalledChecks,
			installCode: 1,
		},
		{
			name: "skip installation token",
			options: installOptions{
				skipInstallToken: true,
			},
			preChecks: notInstalledChecks,
			postChecks: []checkFunc{
				checkBinaryCreated,
				checkBinaryIsRunning,
				checkConfigCreated,
				checkConfigFilesOwnershipAndPermissions(systemUser, systemGroup),
				checkUserConfigCreated,
				checkLaunchdConfigCreated,
				checkHomeDirectoryCreated,
			},
			installCode: 1, // because of invalid installation token
		},
		{
			name: "installation token only",
			options: installOptions{
				installToken: installToken,
			},
			preChecks: notInstalledChecks,
			postChecks: []checkFunc{
				checkBinaryCreated,
				checkBinaryIsRunning,
				checkLatestAppVersion,
				checkConfigCreated,
				checkConfigFilesOwnershipAndPermissions(systemUser, systemGroup),
				checkUserConfigCreated,
				checkEphemeralNotInConfig(userConfigPath),
				checkLaunchdConfigCreated,
				checkTokenInLaunchdConfig,
				checkUserExists,
				checkGroupExists,
				checkHostmetricsConfigNotCreated,
				checkHomeDirectoryCreated,
			},
			installCode: 1, // because of invalid installation token
		},
		{
			name: "installation token and ephemeral",
			options: installOptions{
				installToken: installToken,
				ephemeral:    true,
			},
			preChecks: notInstalledChecks,
			postChecks: []checkFunc{
				checkBinaryCreated,
				checkBinaryIsRunning,
				checkConfigCreated,
				checkConfigFilesOwnershipAndPermissions(systemUser, systemGroup),
				checkUserConfigCreated,
				checkEphemeralInConfig(userConfigPath),
				checkLaunchdConfigCreated,
				checkTokenInLaunchdConfig,
				checkUserExists,
				checkGroupExists,
				checkHostmetricsConfigNotCreated,
				checkHomeDirectoryCreated,
			},
			installCode: 1, // because of invalid installation token
		},
		{
			name: "override default config",
			options: installOptions{
				skipInstallToken: true,
				autoconfirm:      true,
			},
			preActions: []checkFunc{preActionMockConfig},
			preChecks: []checkFunc{
				checkBinaryNotCreated,
				checkConfigCreated,
				checkUserConfigNotCreated,
				checkUserNotExists,
			},
			postChecks: []checkFunc{
				checkBinaryCreated,
				checkBinaryIsRunning,
				checkConfigCreated,
				checkConfigOverrided,
				checkUserConfigCreated,
				checkLaunchdConfigCreated,
			},
			installCode: 1, // because of invalid installation token
		},
		{
			name: "installation token and hostmetrics",
			options: installOptions{
				installToken:       installToken,
				installHostmetrics: true,
			},
			preChecks: notInstalledChecks,
			postChecks: []checkFunc{
				checkBinaryCreated,
				checkBinaryIsRunning,
				checkConfigCreated,
				checkConfigFilesOwnershipAndPermissions(systemUser, systemGroup),
				checkUserConfigCreated,
				checkLaunchdConfigCreated,
				checkTokenInLaunchdConfig,
				checkUserExists,
				checkGroupExists,
				checkHostmetricsConfigCreated,
				checkHostmetricsOwnershipAndPermissions(systemUser, systemGroup),
				checkHomeDirectoryCreated,
			},
			installCode: 1, // because of invalid installation token
		},
		{
			name: "installation token and remotely-managed",
			options: installOptions{
				installToken:    installToken,
				remotelyManaged: true,
			},
			preChecks: notInstalledChecks,
			postChecks: []checkFunc{
				checkBinaryCreated,
				checkBinaryIsRunning,
				checkConfigCreated,
				checkRemoteConfigDirectoryCreated,
				checkConfigFilesOwnershipAndPermissions(systemUser, systemGroup),
				checkUserConfigCreated,
				checkEphemeralNotInConfig(configPath),
				checkLaunchdConfigCreated,
				checkTokenInLaunchdConfig,
				checkUserExists,
				checkGroupExists,
				checkHomeDirectoryCreated,
			},
			installCode: 1, // because of invalid installation token
		},
		{
			name: "installation token, remotely-managed, and ephemeral",
			options: installOptions{
				installToken:    installToken,
				remotelyManaged: true,
				ephemeral:       true,
			},
			preChecks: notInstalledChecks,
			postChecks: []checkFunc{
				checkBinaryCreated,
				checkBinaryIsRunning,
				checkConfigCreated,
				checkRemoteConfigDirectoryCreated,
				checkConfigFilesOwnershipAndPermissions(systemUser, systemGroup),
				checkUserConfigCreated,
				checkEphemeralInConfig(configPath),
				checkLaunchdConfigCreated,
				checkTokenInLaunchdConfig,
				checkUserExists,
				checkGroupExists,
				checkHomeDirectoryCreated,
			},
			installCode: 1, // because of invalid installation token
		},
		{
			name: "installation token only, binary not in PATH",
			options: installOptions{
				installToken: installToken,
				envs: map[string]string{
					"PATH": "/sbin:/bin:/usr/sbin:/usr/bin",
				},
			},
			preChecks: notInstalledChecks,
			postChecks: []checkFunc{
				checkBinaryCreated,
				checkBinaryIsRunning,
				checkConfigCreated,
				checkConfigFilesOwnershipAndPermissions(systemUser, systemGroup),
				checkUserConfigCreated,
				checkLaunchdConfigCreated,
				checkTokenInLaunchdConfig,
				checkUserExists,
				checkGroupExists,
				checkHomeDirectoryCreated,
			},
			installCode: 1, // because of invalid installation token
		},
		{
			name: "same installation token in launchd config",
			options: installOptions{
				installToken: installToken,
			},
			preActions: []checkFunc{preActionMockLaunchdConfig},
			preChecks: []checkFunc{
				checkBinaryNotCreated,
				checkConfigNotCreated,
				checkUserConfigNotCreated,
				checkUserNotExists,
				checkGroupNotExists,
				checkLaunchdConfigCreated,
			},
			postChecks: []checkFunc{
				checkBinaryCreated,
				checkBinaryIsRunning,
				checkConfigCreated,
				checkUserConfigCreated,
				checkLaunchdConfigCreated,
				checkTokenInLaunchdConfig,
				checkHomeDirectoryCreated,
			},
			installCode: 1, // because of invalid installation token
		},
		{
			name: "different installation token in launchd config",
			options: installOptions{
				installToken: installToken,
			},
			preActions: []checkFunc{
				preActionMockLaunchdConfig,
				preActionWriteDifferentTokenToLaunchdConfig,
			},
			preChecks: []checkFunc{
				checkBinaryNotCreated,
				checkConfigNotCreated,
				checkUserConfigNotCreated,
				checkUserNotExists,
				checkGroupNotExists,
				checkLaunchdConfigCreated,
			},
			postChecks: []checkFunc{
				checkBinaryNotCreated,
				checkConfigNotCreated,
				checkUserConfigNotCreated,
				checkUserNotExists,
				checkGroupNotExists,
				checkLaunchdConfigCreated,
				checkAbortedDueToDifferentToken,
				checkDifferentTokenInLaunchdConfig,
			},
			installCode: 1,
		},
		{
			name: "same api base url",
			options: installOptions{
				apiBaseURL:       apiBaseURL,
				skipInstallToken: true,
			},
			preActions: []checkFunc{
				preActionInstallPackage,
			},
			preChecks: []checkFunc{
				checkBinaryCreated,
				checkConfigCreated,
				checkUserConfigCreated,
				checkUserExists,
			},
			postChecks: []checkFunc{
				checkBinaryCreated,
				checkBinaryIsRunning,
				checkConfigCreated,
				checkUserConfigCreated,
				checkAPIBaseURLInConfig,
			},
			installCode: 1, // because of invalid installation token
		},
		{
			name: "different api base url",
			options: installOptions{
				apiBaseURL:       apiBaseURL,
				skipInstallToken: true,
			},
			preActions: []checkFunc{
				preActionInstallPackageWithDifferentAPIBaseURL,
			},
			preChecks: []checkFunc{
				checkBinaryCreated,
				checkConfigCreated,
				checkUserConfigCreated,
				checkUserExists,
			},
			postChecks: []checkFunc{
				checkBinaryCreated,
				checkConfigCreated,
				checkUserConfigCreated,
				checkLaunchdConfigCreated,
				checkAbortedDueToDifferentAPIBaseURL,
			},
			installCode: 1,
		},
		{
			name: "adding api base url",
			options: installOptions{
				apiBaseURL:       apiBaseURL,
				skipInstallToken: true,
			},
			preActions: []checkFunc{preActionInstallPackageWithNoAPIBaseURL},
			preChecks: []checkFunc{
				checkBinaryCreated,
				checkConfigCreated,
				checkUserConfigCreated,
				checkUserExists,
			},
			postChecks: []checkFunc{
				checkBinaryCreated,
				checkConfigCreated,
				checkUserConfigCreated,
				checkAPIBaseURLInConfig,
			},
			installCode: 1, // because of invalid installation token
		},
		{
			name: "editing api base url",
			options: installOptions{
				apiBaseURL:       apiBaseURL,
				skipInstallToken: true,
			},
			preActions: []checkFunc{
				preActionInstallPackageWithNoAPIBaseURL,
			},
			preChecks: []checkFunc{
				checkBinaryCreated,
				checkConfigCreated,
				checkUserConfigCreated,
				checkUserExists,
			},
			postChecks: []checkFunc{
				checkBinaryCreated,
				checkConfigCreated,
				checkUserConfigCreated,
				checkAPIBaseURLInConfig,
			},
			installCode: 1, // because of invalid installation token
		},
		{
			name: "configuration with tags",
			options: installOptions{
				skipInstallToken: true,
				tags: map[string]string{
					"lorem":     "ipsum",
					"foo":       "bar",
					"escape_me": "'\\/",
					"slash":     "a/b",
					"numeric":   "1_024",
				},
			},
			preChecks: notInstalledChecks,
			postChecks: []checkFunc{
				checkBinaryCreated,
				checkBinaryIsRunning,
				checkConfigCreated,
				checkConfigFilesOwnershipAndPermissions(systemUser, systemGroup),
				checkTags,
				checkLaunchdConfigCreated,
			},
			installCode: 1, // because of invalid installation token
		},
		{
			name: "same tags",
			options: installOptions{
				skipInstallToken: true,
				tags: map[string]string{
					"lorem":     "ipsum",
					"foo":       "bar",
					"escape_me": "'\\/",
					"slash":     "a/b",
					"numeric":   "1_024",
				},
			},
			preActions: []checkFunc{
				preActionInstallPackage,
			},
			preChecks: []checkFunc{
				checkBinaryCreated,
				checkConfigCreated,
				checkUserConfigCreated,
				checkUserExists,
			},
			postChecks: []checkFunc{
				checkBinaryCreated,
				checkBinaryIsRunning,
				checkConfigCreated,
				checkUserConfigCreated,
				checkTags,
				checkLaunchdConfigCreated,
			},
			installCode: 1, // because of invalid installation token
		},
		{
			name: "different tags",
			options: installOptions{
				skipInstallToken: true,
				tags: map[string]string{
					"lorem":     "ipsum",
					"foo":       "bar",
					"escape_me": "'\\/",
					"slash":     "a/b",
					"numeric":   "1_024",
				},
			},
			preActions: []checkFunc{
				preActionInstallPackageWithDifferentTags,
			},
			preChecks: []checkFunc{
				checkBinaryCreated,
				checkConfigCreated,
				checkUserConfigCreated,
				checkUserExists,
			},
			postChecks: []checkFunc{
				checkBinaryCreated,
				checkConfigCreated,
				checkUserConfigCreated,
				checkDifferentTags,
				checkLaunchdConfigCreated,
				checkAbortedDueToDifferentTags,
			},
			installCode: 1,
		},
		{
			name: "editing tags",
			options: installOptions{
				skipInstallToken: true,
				tags: map[string]string{
					"lorem":     "ipsum",
					"foo":       "bar",
					"escape_me": "'\\/",
					"slash":     "a/b",
					"numeric":   "1_024",
				},
			},
			preActions: []checkFunc{
				preActionInstallPackageWithNoTags,
			},
			preChecks: []checkFunc{
				checkBinaryCreated,
				checkConfigCreated,
				checkUserConfigCreated,
				checkUserExists,
			},
			postChecks: []checkFunc{
				checkBinaryCreated,
				checkBinaryIsRunning,
				checkConfigCreated,
				checkTags,
				checkLaunchdConfigCreated,
			},
			installCode: 1, // because of invalid installation token
		},
	} {
		t.Run(spec.name, func(t *testing.T) {
			runTest(t, &spec)
		})
	}
}
