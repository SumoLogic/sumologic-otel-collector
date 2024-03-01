//go:build windows
// +build windows

package sumologic_scripts_tests

import (
	"testing"
)

// TODO: Check file ownership
// TODO: Set up file permissions to be able to modify config files on Windows

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
			name: "installation token only",
			options: installOptions{
				installToken: installToken,
			},
			preChecks: []checkFunc{checkBinaryNotCreated, checkConfigNotCreated, checkUserConfigNotCreated, checkUserNotExists},
			postChecks: []checkFunc{
				checkBinaryCreated,
				checkBinaryIsRunning,
				checkConfigCreated,
				// checkConfigFilesOwnershipAndPermissions(rootUser, rootGroup),
				checkUserConfigCreated,
				checkEphemeralNotInConfig(userConfigPath),
				checkTokenInConfig,
				checkUserNotExists,
				checkHostmetricsConfigNotCreated,
			},
		},
		{
			name: "installation token and ephemeral",
			options: installOptions{
				installToken: installToken,
				ephemeral:    true,
			},
			preChecks: []checkFunc{checkBinaryNotCreated, checkConfigNotCreated, checkUserConfigNotCreated, checkUserNotExists},
			postChecks: []checkFunc{
				checkBinaryCreated,
				checkBinaryIsRunning,
				checkConfigCreated,
				// checkConfigFilesOwnershipAndPermissions(rootUser, rootGroup),
				checkUserConfigCreated,
				checkTokenInConfig,
				checkEphemeralInConfig(userConfigPath),
				checkUserNotExists,
				checkHostmetricsConfigNotCreated,
			},
		},
		{
			name: "installation token and hostmetrics",
			options: installOptions{
				installToken:       installToken,
				installHostmetrics: true,
			},
			preChecks: []checkFunc{checkBinaryNotCreated, checkConfigNotCreated, checkUserConfigNotCreated, checkUserNotExists},
			postChecks: []checkFunc{
				checkBinaryCreated,
				checkBinaryIsRunning,
				checkConfigCreated,
				// checkConfigFilesOwnershipAndPermissions(rootUser, rootGroup),
				checkUserConfigCreated,
				checkTokenInConfig,
				checkUserNotExists,
				checkHostmetricsConfigCreated,
			},
		},
		{
			name: "installation token and remotely-managed",
			options: installOptions{
				installToken:    installToken,
				remotelyManaged: true,
			},
			preChecks: []checkFunc{checkBinaryNotCreated, checkConfigNotCreated, checkUserConfigNotCreated, checkUserNotExists},
			postChecks: []checkFunc{
				checkBinaryCreated,
				checkBinaryIsRunning,
				checkConfigCreated,
				// checkConfigFilesOwnershipAndPermissions(rootUser, rootGroup),
				checkRemoteConfigDirectoryCreated,
				checkTokenInSumoConfig,
				checkEphemeralNotInConfig(configPath),
				checkUserNotExists,
			},
		},
		{
			name: "installation token, remotely-managed, and ephemeral",
			options: installOptions{
				installToken:    installToken,
				remotelyManaged: true,
				ephemeral:       true,
			},
			preChecks: []checkFunc{checkBinaryNotCreated, checkConfigNotCreated, checkUserConfigNotCreated, checkUserNotExists},
			postChecks: []checkFunc{
				checkBinaryCreated,
				checkBinaryIsRunning,
				checkConfigCreated,
				// checkConfigFilesOwnershipAndPermissions(rootUser, rootGroup),
				checkRemoteConfigDirectoryCreated,
				checkTokenInSumoConfig,
				checkEphemeralInConfig(configPath),
				checkUserNotExists,
			},
		},
		{
			name: "configuration with tags",
			options: installOptions{
				installToken: installToken,
				tags: map[string]string{
					"lorem":     "ipsum",
					"foo":       "bar",
					"escape_me": "'\\/",
					"slash":     "a/b",
					"numeric":   "1_024",
				},
			},
			preChecks: []checkFunc{checkBinaryNotCreated, checkConfigNotCreated, checkUserConfigNotCreated, checkUserNotExists},
			postChecks: []checkFunc{
				checkBinaryCreated,
				checkBinaryIsRunning,
				checkConfigCreated,
				// checkConfigFilesOwnershipAndPermissions(rootUser, rootGroup),
				checkTags,
			},
		},
	} {
		t.Run(spec.name, func(t *testing.T) {
			runTest(t, &spec)
		})
	}
}
