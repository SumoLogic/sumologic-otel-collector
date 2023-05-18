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
				checkConfigFilesOwnershipAndPermissions(getSystemUser(), getSystemGroup()),
				checkUserConfigCreated,
				checkLaunchdConfigCreated,
			},
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
				checkConfigCreated,
				checkConfigFilesOwnershipAndPermissions(getSystemUser(), getSystemGroup()),
				checkUserConfigCreated,
				checkLaunchdConfigCreated,
				checkTokenInLaunchdConfig,
				checkUserExists,
				checkHostmetricsConfigNotCreated,
			},
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
				checkConfigFilesOwnershipAndPermissions(getSystemUser(), getSystemGroup()),
				checkUserConfigCreated,
				checkLaunchdConfigCreated,
				checkTokenInLaunchdConfig,
				checkUserExists,
				checkHostmetricsConfigCreated,
				checkConfigFilesOwnershipAndPermissions(getSystemUser(), getSystemGroup()),
			},
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
				checkConfigFilesOwnershipAndPermissions(getSystemUser(), getSystemGroup()),
				checkUserConfigCreated,
				checkLaunchdConfigCreated,
				checkTokenInLaunchdConfig,
				checkUserExists,
			},
		},
		{
			name: "same installation token",
			options: installOptions{
				installToken: installToken,
			},
			preActions: []checkFunc{preActionMockLaunchdConfig},
			preChecks:  []checkFunc{checkBinaryNotCreated, checkConfigNotCreated, checkUserConfigNotCreated, checkUserNotExists, checkGroupNotExists, checkLaunchdConfigCreated},
			postChecks: []checkFunc{checkBinaryCreated, checkBinaryIsRunning, checkConfigCreated, checkUserConfigCreated, checkLaunchdConfigCreated, checkTokenInLaunchdConfig},
		},
		{
			name: "different installation token",
			options: installOptions{
				installToken: installToken,
			},
			preActions:  []checkFunc{preActionMockLaunchdConfig, preActionWriteDifferentTokenToLaunchdConfig},
			preChecks:   []checkFunc{checkBinaryNotCreated, checkConfigNotCreated, checkUserConfigNotCreated, checkUserNotExists, checkGroupNotExists, checkLaunchdConfigCreated},
			postChecks:  []checkFunc{checkBinaryNotCreated, checkConfigNotCreated, checkUserConfigNotCreated, checkUserNotExists, checkGroupNotExists, checkLaunchdConfigCreated, checkAbortedDueToDifferentToken},
			installCode: 1,
		},
	} {
		t.Run(spec.name, func(t *testing.T) {
			runTest(t, &spec)
		})
	}
}
