//go:build !(windows || darwin)

package sumologic_scripts_tests

import (
	"fmt"
	"testing"
)

func TestInstallScript(t *testing.T) {
	notInstalledChecks := []checkFunc{
		checkBinaryNotCreated,
		checkConfigNotCreated,
		checkUserConfigNotCreated,
		checkUserNotExists,
		checkGroupNotExists,
		checkUserConfigNotCreated,
		checkSystemdConfigNotCreated,
		checkTokenEnvFileNotCreated,
		checkHostmetricsConfigNotCreated,
	}

	installedWithSystemdChecks := []checkFunc{
		checkBinaryCreated,
		checkConfigCreated,
		checkUserConfigCreated,
		checkSystemdConfigCreated,
		checkUserExists,
		checkGroupExists,
	}

	for _, spec := range []testSpec{
		{
			name:        "no arguments",
			installType: BINARY_INSTALL | PACKAGE_INSTALL,
			options:     installOptions{},
			preChecks:   notInstalledChecks,
			postChecks:  append(notInstalledChecks, checkAbortedDueToNoToken),
			installCode: 1,
		},
		{
			name:        "download only",
			installType: BINARY_INSTALL,
			options: installOptions{
				downloadOnly: true,
			},
			preChecks: notInstalledChecks,
			postChecks: []checkFunc{
				checkBinaryCreated,
				checkConfigNotCreated,
				checkUserConfigNotCreated,
				checkSystemdConfigNotCreated,
				checkUserNotExists,
				checkGroupNotExists,
			},
		},
		{
			name:        "download only",
			installType: PACKAGE_INSTALL,
			options: installOptions{
				downloadOnly: true,
			},
			preChecks:  notInstalledChecks,
			postChecks: append(notInstalledChecks, checkPackageCreated),
		},
		{
			name:        "download only with timeout",
			installType: BINARY_INSTALL | PACKAGE_INSTALL,
			options: installOptions{
				downloadOnly:      true,
				timeout:           1,
				dontKeepDownloads: true,
			},
			// Skip this test as getting binary in github actions takes less than one second
			conditionalChecks: []condCheckFunc{checkSkipTest},
			preChecks:         notInstalledChecks,
			postChecks:        append(notInstalledChecks, checkDownloadTimeout),
			installCode:       curlTimeoutErrorCode,
		},
		{
			name:        "skip config",
			installType: BINARY_INSTALL,
			options: installOptions{
				skipConfig:       true,
				skipInstallToken: true,
			},
			preChecks: notInstalledChecks,
			postChecks: []checkFunc{
				checkBinaryCreated,
				checkConfigNotCreated,
				checkUserConfigNotCreated,
			},
		},
		{
			name:        "skip config",
			installType: PACKAGE_INSTALL,
			options: installOptions{
				skipConfig:       true,
				skipInstallToken: true,
			},
			preChecks: notInstalledChecks,
			postChecks: append(notInstalledChecks, []checkFunc{
				checkAbortedDueToSkipConfigUnsupported,
			}...),
			installCode: 1,
		},
		{
			name:        "skip installation token",
			installType: BINARY_INSTALL,
			options: installOptions{
				skipInstallToken: true,
				skipSystemd:      true,
			},
			preChecks: notInstalledChecks,
			postChecks: []checkFunc{
				checkBinaryCreated,
				checkBinaryIsRunning,
				checkConfigCreated,
				checkConfigFilesOwnershipAndPermissions,
				checkUserConfigNotCreated,
				checkSystemdConfigNotCreated,
			},
		},
		{
			name:        "skip installation token",
			installType: PACKAGE_INSTALL,
			options: installOptions{
				skipInstallToken: true,
			},
			preChecks:   notInstalledChecks,
			postChecks:  notInstalledChecks,
			installCode: 1,
		},
		{
			name:        "override default config",
			installType: BINARY_INSTALL,
			options: installOptions{
				skipInstallToken: true,
				autoconfirm:      true,
			},
			preActions: []checkFunc{
				preActionMockConfig,
			},
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
				checkUserConfigNotCreated,
				checkSystemdConfigNotCreated,
			},
		},
		{
			name:        "override default config",
			installType: PACKAGE_INSTALL,
			options: installOptions{
				installToken: installToken,
				autoconfirm:  true,
			},
			preActions: []checkFunc{
				preActionInstallPreviousVersion,
			},
			preChecks: installedWithSystemdChecks,
			postChecks: []checkFunc{
				checkBinaryCreated,
				checkBinaryIsRunning,
				checkConfigCreated,
				checkConfigOverrided,
				checkUserConfigCreated,
				checkSystemdConfigCreated,
			},
		},
		{
			name:        "installation token only",
			installType: BINARY_INSTALL,
			options: installOptions{
				skipSystemd:  true,
				installToken: installToken,
			},
			preChecks: notInstalledChecks,
			postChecks: []checkFunc{
				checkBinaryCreated,
				checkBinaryIsRunning,
				checkConfigCreated,
				checkConfigFilesOwnershipAndPermissions,
				checkUserConfigCreated,
				checkTokenInConfig,
				checkSystemdConfigNotCreated,
				checkUserNotExists,
				checkHostmetricsConfigNotCreated,
				checkTokenEnvFileNotCreated,
			},
		},
		{
			name:        "installation token only",
			installType: PACKAGE_INSTALL,
			options: installOptions{
				apiBaseURL:   mockAPIBaseURL,
				installToken: installToken,
			},
			preChecks: notInstalledChecks,
			postChecks: []checkFunc{
				checkBinaryCreated,
				checkBinaryIsRunning,
				checkConfigCreated,
				checkConfigFilesOwnershipAndPermissions,
				checkUserConfigCreated,
				checkTokenEnvFileCreated,
				checkTokenInEnvFile,
				checkSystemdConfigCreated,
				checkUserExists,
				checkHostmetricsConfigNotCreated,
			},
		},
		{
			name:        "deprecated installation token only",
			installType: BINARY_INSTALL,
			options: installOptions{
				skipSystemd:            true,
				deprecatedInstallToken: installToken,
			},
			preChecks: notInstalledChecks,
			postChecks: []checkFunc{
				checkBinaryCreated,
				checkBinaryIsRunning,
				checkConfigCreated,
				checkConfigFilesOwnershipAndPermissions,
				checkUserConfigCreated,
				checkSystemdConfigNotCreated,
				checkUserNotExists,
				checkHostmetricsConfigNotCreated,
			},
		},
		{
			name:        "installation token and hostmetrics",
			installType: BINARY_INSTALL | PACKAGE_INSTALL,
			options: installOptions{
				apiBaseURL:         mockAPIBaseURL,
				installToken:       installToken,
				installHostmetrics: true,
			},
			preChecks: notInstalledChecks,
			postChecks: []checkFunc{
				checkBinaryCreated,
				checkBinaryIsRunning,
				checkConfigCreated,
				checkConfigFilesOwnershipAndPermissions,
				checkUserConfigCreated,
				checkTokenInEnvFile,
				checkSystemdConfigCreated,
				checkUserExists,
				checkHostmetricsConfigCreated,
				checkHostmetricsOwnershipAndPermissions,
			},
		},
		{
			name:        "installation token only, binary not in PATH",
			installType: BINARY_INSTALL | PACKAGE_INSTALL,
			options: installOptions{
				apiBaseURL:   mockAPIBaseURL,
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
				checkConfigFilesOwnershipAndPermissions,
				checkUserConfigCreated,
				checkTokenInEnvFile,
				checkSystemdConfigCreated,
				checkUserExists,
			},
		},
		{
			name:        "same installation token",
			installType: BINARY_INSTALL,
			options: installOptions{
				skipSystemd:  true,
				installToken: installToken,
			},
			preActions: []checkFunc{
				preActionMockUserConfig,
				preActionWriteTokenToUserConfig,
			},
			preChecks: []checkFunc{
				checkBinaryNotCreated,
				checkConfigNotCreated,
				checkUserConfigCreated,
				checkUserNotExists,
			},
			postChecks: []checkFunc{
				checkBinaryCreated,
				checkBinaryIsRunning,
				checkConfigCreated,
				checkUserConfigCreated,
				checkTokenInConfig,
				checkSystemdConfigNotCreated,
			},
		},
		{
			name:        "same installation token",
			installType: PACKAGE_INSTALL,
			options: installOptions{
				installToken: installToken,
			},
			preActions: []checkFunc{
				preActionInstallPreviousVersion,
			},
			preChecks: []checkFunc{
				checkBinaryCreated,
				checkConfigCreated,
				checkUserConfigCreated,
				checkTokenInEnvFile,
				checkUserExists,
			},
			postChecks: []checkFunc{
				checkBinaryCreated,
				checkBinaryIsRunning,
				checkConfigCreated,
				checkUserConfigCreated,
				checkTokenInEnvFile,
				checkSystemdConfigCreated,
			},
		},
		{
			name:        "different installation token",
			installType: BINARY_INSTALL,
			options: installOptions{
				skipSystemd:  true,
				installToken: installToken,
			},
			preActions: []checkFunc{
				preActionMockUserConfig,
				preActionWriteDifferentTokenToUserConfig,
			},
			preChecks: []checkFunc{
				checkBinaryNotCreated,
				checkConfigNotCreated,
				checkUserConfigCreated,
				checkUserNotExists,
			},
			postChecks: []checkFunc{
				checkBinaryNotCreated,
				checkConfigNotCreated,
				checkUserConfigCreated,
				checkSystemdConfigNotCreated,
				checkAbortedDueToDifferentToken,
			},
			installCode: 1,
		},
		{
			name:        "different installation token",
			installType: PACKAGE_INSTALL,
			options: installOptions{
				installToken: installToken,
			},
			preActions: []checkFunc{
				preActionInstallPackage,
				preActionWriteDifferentTokenToUserConfig,
			},
			preChecks: installedWithSystemdChecks,
			postChecks: []checkFunc{
				checkBinaryCreated,
				checkConfigCreated,
				checkUserConfigCreated,
				checkSystemdConfigCreated,
				checkAbortedDueToDifferentToken,
			},
			installCode: 1,
		},
		{
			name:        "adding installation token",
			installType: BINARY_INSTALL,
			options: installOptions{
				skipSystemd:  true,
				installToken: installToken,
			},
			preActions: []checkFunc{
				preActionMockUserConfig,
			},
			preChecks: []checkFunc{
				checkBinaryNotCreated,
				checkConfigNotCreated,
				checkUserConfigCreated,
				checkUserNotExists,
			},
			postChecks: []checkFunc{
				checkBinaryCreated,
				checkConfigCreated,
				checkUserConfigCreated,
				checkTokenInConfig,
				checkSystemdConfigNotCreated,
			},
		},
		{
			name:        "adding installation token",
			installType: PACKAGE_INSTALL,
			options: installOptions{
				installToken: installToken,
			},
			preActions: []checkFunc{
				preActionInstallPreviousVersion,
			},
			preChecks: installedWithSystemdChecks,
			postChecks: []checkFunc{
				checkBinaryCreated,
				checkConfigCreated,
				checkUserConfigCreated,
				checkTokenInEnvFile,
				checkSystemdConfigCreated,
			},
		},
		{
			name:        "editing installation token",
			installType: BINARY_INSTALL,
			options: installOptions{
				skipSystemd:  true,
				apiBaseURL:   apiBaseURL,
				installToken: installToken,
			},
			preActions: []checkFunc{
				preActionMockUserConfig,
				preActionWriteEmptyUserConfig,
			},
			preChecks: []checkFunc{
				checkBinaryNotCreated,
				checkConfigNotCreated,
				checkUserConfigCreated,
				checkUserNotExists,
			},
			postChecks: []checkFunc{
				checkBinaryCreated,
				checkConfigCreated,
				checkUserConfigCreated,
				checkTokenInConfig,
				checkSystemdConfigNotCreated,
			},
		},
		{
			name:        "editing installation token",
			installType: PACKAGE_INSTALL,
			options: installOptions{
				apiBaseURL:   apiBaseURL,
				installToken: installToken,
			},
			preActions: []checkFunc{
				preActionInstallPreviousVersion,
				preActionWriteEmptyUserConfig,
			},
			preChecks: installedWithSystemdChecks,
			postChecks: []checkFunc{
				checkBinaryCreated,
				checkConfigCreated,
				checkUserConfigCreated,
				checkTokenInEnvFile,
				checkSystemdConfigCreated,
			},
		},
		{
			name:        "same api base url",
			installType: BINARY_INSTALL,
			options: installOptions{
				skipSystemd:      true,
				apiBaseURL:       apiBaseURL,
				skipInstallToken: true,
			},
			preActions: []checkFunc{
				preActionMockUserConfig,
				preActionWriteAPIBaseURLToUserConfig,
			},
			preChecks: []checkFunc{
				checkBinaryNotCreated,
				checkConfigNotCreated,
				checkUserConfigCreated,
				checkUserNotExists,
			},
			postChecks: []checkFunc{
				checkBinaryCreated,
				checkBinaryIsRunning,
				checkConfigCreated,
				checkUserConfigCreated,
				checkAPIBaseURLInConfig,
				checkSystemdConfigNotCreated,
			},
		},
		{
			name:        "same api base url",
			installType: PACKAGE_INSTALL,
			options: installOptions{
				apiBaseURL:   apiBaseURL,
				installToken: installToken,
			},
			preActions: []checkFunc{
				preActionInstallPreviousVersion,
				preActionWriteAPIBaseURLToUserConfig,
			},
			preChecks: installedWithSystemdChecks,
			postChecks: []checkFunc{
				checkBinaryCreated,
				checkBinaryIsRunning,
				checkConfigCreated,
				checkUserConfigCreated,
				checkAPIBaseURLInConfig,
				checkSystemdConfigCreated,
			},
		},
		{
			name:        "different api base url",
			installType: BINARY_INSTALL,
			options: installOptions{
				skipSystemd:      true,
				apiBaseURL:       apiBaseURL,
				skipInstallToken: true,
			},
			preActions: []checkFunc{
				preActionMockUserConfig,
				preActionWriteDifferentAPIBaseURLToUserConfig,
			},
			preChecks: []checkFunc{
				checkBinaryNotCreated,
				checkConfigNotCreated,
				checkUserConfigCreated,
				checkUserNotExists,
			},
			postChecks: []checkFunc{
				checkBinaryNotCreated,
				checkConfigNotCreated,
				checkUserConfigCreated,
				checkSystemdConfigNotCreated,
				checkAbortedDueToDifferentAPIBaseURL,
			},
			installCode: 1,
		},
		{
			name:        "different api base url",
			installType: PACKAGE_INSTALL,
			options: installOptions{
				apiBaseURL:   apiBaseURL,
				installToken: installToken,
			},
			preActions: []checkFunc{
				preActionInstallPackage,
				preActionWriteDifferentAPIBaseURLToUserConfig,
			},
			preChecks: installedWithSystemdChecks,
			postChecks: []checkFunc{
				checkBinaryCreated,
				checkConfigCreated,
				checkUserConfigCreated,
				checkSystemdConfigCreated,
				checkAbortedDueToDifferentAPIBaseURL,
			},
			installCode: 1,
		},
		{
			name:        "adding api base url",
			installType: BINARY_INSTALL,
			options: installOptions{
				skipSystemd:      true,
				apiBaseURL:       apiBaseURL,
				skipInstallToken: true,
			},
			preActions: []checkFunc{preActionMockUserConfig},
			preChecks: []checkFunc{
				checkBinaryNotCreated,
				checkConfigNotCreated,
				checkUserConfigCreated,
				checkUserNotExists,
			},
			postChecks: []checkFunc{
				checkBinaryCreated,
				checkConfigCreated,
				checkUserConfigCreated,
				checkAPIBaseURLInConfig,
				checkSystemdConfigNotCreated,
			},
		},
		{
			name:        "adding api base url",
			installType: PACKAGE_INSTALL,
			options: installOptions{
				apiBaseURL:   apiBaseURL,
				installToken: installToken,
			},
			preActions: []checkFunc{
				preActionInstallPackageWithOptions(installOptions{
					apiBaseURL:   mockAPIBaseURL,
					installToken: installToken,
					version:      previousPackageVersion,
				}),
				preActionWriteEmptyUserConfig,
			},
			preChecks: installedWithSystemdChecks,
			postChecks: []checkFunc{
				checkBinaryCreated,
				checkConfigCreated,
				checkUserConfigCreated,
				checkAPIBaseURLInConfig,
				checkSystemdConfigCreated,
			},
		},
		{
			name:        "editing api base url",
			installType: BINARY_INSTALL,
			options: installOptions{
				skipSystemd:      true,
				apiBaseURL:       apiBaseURL,
				skipInstallToken: true,
			},
			preActions: []checkFunc{
				preActionMockUserConfig,
				preActionWriteEmptyUserConfig,
			},
			preChecks: []checkFunc{
				checkBinaryNotCreated,
				checkConfigNotCreated,
				checkUserConfigCreated,
				checkUserNotExists,
			},
			postChecks: []checkFunc{
				checkBinaryCreated,
				checkConfigCreated,
				checkUserConfigCreated,
				checkAPIBaseURLInConfig,
				checkSystemdConfigNotCreated,
			},
		},
		{
			name:        "editing api base url",
			installType: PACKAGE_INSTALL,
			options: installOptions{
				apiBaseURL:   apiBaseURL,
				installToken: installToken,
			},
			preActions: []checkFunc{
				preActionInstallPackageWithOptions(installOptions{
					apiBaseURL:   mockAPIBaseURL,
					installToken: installToken,
				}),
			},
			preChecks: installedWithSystemdChecks,
			postChecks: []checkFunc{
				checkBinaryCreated,
				checkConfigCreated,
				checkUserConfigCreated,
				checkSystemdConfigCreated,
			},
			installCode: 1,
		},
		{
			name:        "empty installation token",
			installType: BINARY_INSTALL,
			options: installOptions{
				skipSystemd: true,
			},
			preActions: []checkFunc{
				preActionMockUserConfig,
				preActionWriteDifferentTokenToUserConfig,
			},
			preChecks: []checkFunc{
				checkBinaryNotCreated,
				checkConfigNotCreated,
				checkUserConfigCreated,
				checkUserNotExists,
			},
			postChecks: []checkFunc{
				checkBinaryCreated,
				checkConfigCreated,
				checkUserConfigCreated,
				checkSystemdConfigNotCreated,
				checkDifferentTokenInConfig,
			},
		},
		{
			name:        "upgrade with empty installation token",
			installType: PACKAGE_INSTALL,
			options: installOptions{
				installToken: "",
			},
			preActions: []checkFunc{
				preActionInstallPreviousVersion,
				preActionWriteDifferentTokenToUserConfig,
			},
			preChecks: installedWithSystemdChecks,
			postChecks: []checkFunc{
				checkBinaryCreated,
				checkConfigCreated,
				checkUserConfigCreated,
				checkSystemdConfigCreated,
				checkDifferentTokenInConfig,
			},
		},
		{
			name:        "configuration with tags",
			installType: BINARY_INSTALL,
			options: installOptions{
				skipSystemd:      true,
				skipInstallToken: true,
				tags: map[string]string{
					"lorem":     "ipsum",
					"foo":       "bar",
					"escape_me": "'\\/",
					"slash":     "a/b",
					"numeric":   "1_024",
				},
			},
			preChecks: []checkFunc{
				checkBinaryNotCreated,
				checkConfigNotCreated,
				checkUserConfigNotCreated,
				checkUserNotExists,
			},
			postChecks: []checkFunc{
				checkBinaryCreated,
				checkBinaryIsRunning,
				checkConfigCreated,
				checkConfigFilesOwnershipAndPermissions,
				checkTags,
				checkSystemdConfigNotCreated,
			},
		},
		{
			name:        "configuration with tags",
			installType: PACKAGE_INSTALL,
			options: installOptions{
				installToken: installToken,
				tags: map[string]string{
					"lorem": "ipsum",
					"foo":   "bar",
				},
			},
			preChecks: []checkFunc{
				checkBinaryNotCreated,
				checkConfigNotCreated,
				checkUserConfigNotCreated,
				checkUserNotExists,
			},
			postChecks: []checkFunc{
				checkBinaryCreated,
				checkBinaryIsRunning,
				checkConfigCreated,
				checkConfigFilesOwnershipAndPermissions,
				checkTags,
				checkSystemdConfigCreated,
			},
		},
		{
			name:        "same tags",
			installType: BINARY_INSTALL,
			options: installOptions{
				skipSystemd:      true,
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
				preActionMockUserConfig,
				preActionWriteTagsToUserConfig,
			},
			preChecks: []checkFunc{
				checkBinaryNotCreated,
				checkConfigNotCreated,
				checkUserConfigCreated,
				checkUserNotExists,
			},
			postChecks: []checkFunc{
				checkBinaryCreated,
				checkBinaryIsRunning,
				checkConfigCreated,
				checkUserConfigCreated,
				checkTags,
				checkSystemdConfigNotCreated,
			},
		},
		{
			name:        "same tags",
			installType: PACKAGE_INSTALL,
			options: installOptions{
				installToken: installToken,
				tags: map[string]string{
					"lorem": "ipsum",
					"foo":   "bar",
				},
			},
			preActions: []checkFunc{
				preActionInstallPreviousVersion,
				preActionWriteTagsToUserConfig,
			},
			preChecks: installedWithSystemdChecks,
			postChecks: []checkFunc{
				checkBinaryCreated,
				checkBinaryIsRunning,
				checkConfigCreated,
				checkUserConfigCreated,
				checkTags,
				checkSystemdConfigCreated,
			},
		},
		{
			name:        "different tags",
			installType: BINARY_INSTALL,
			options: installOptions{
				skipSystemd:      true,
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
				preActionMockUserConfig,
				preActionWriteDifferentTagsToUserConfig,
			},
			preChecks: []checkFunc{
				checkBinaryNotCreated,
				checkConfigNotCreated,
				checkUserConfigCreated,
				checkUserNotExists,
			},
			postChecks: []checkFunc{
				checkBinaryNotCreated,
				checkConfigNotCreated,
				checkUserConfigCreated,
				checkDifferentTags,
				checkSystemdConfigNotCreated,

				checkAbortedDueToDifferentTags,
			},
			installCode: 1,
		},
		{
			name:        "different tags",
			installType: PACKAGE_INSTALL,
			options: installOptions{
				tags: map[string]string{
					"lorem": "ipsum",
					"foo":   "bar",
				},
			},
			preActions: []checkFunc{
				preActionInstallPreviousVersion,
				preActionWriteDifferentTagsToUserConfig,
			},
			preChecks: installedWithSystemdChecks,
			postChecks: []checkFunc{
				checkBinaryCreated,
				checkConfigCreated,
				checkUserConfigCreated,
				checkDifferentTags,
				checkSystemdConfigCreated,

				checkAbortedDueToDifferentTags,
			},
			installCode: 1,
		},
		{
			name:        "editing tags",
			installType: BINARY_INSTALL,
			options: installOptions{
				skipSystemd:      true,
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
				preActionMockUserConfig,
				preActionWriteEmptyUserConfig,
			},
			preChecks: []checkFunc{
				checkBinaryNotCreated,
				checkConfigNotCreated,
				checkUserConfigCreated,
				checkUserNotExists,
			},
			postChecks: []checkFunc{
				checkBinaryCreated,
				checkBinaryIsRunning,
				checkConfigCreated,
				checkTags,
				checkSystemdConfigNotCreated,
			},
		},
		{
			name:        "editing tags",
			installType: PACKAGE_INSTALL,
			options: installOptions{
				installToken: installToken,
				tags: map[string]string{
					"lorem": "ipsum",
					"foo":   "bar",
				},
			},
			preActions: []checkFunc{
				preActionInstallPreviousVersion,
				preActionWriteEmptyUserConfig,
			},
			preChecks: installedWithSystemdChecks,
			postChecks: []checkFunc{
				checkBinaryCreated,
				checkBinaryIsRunning,
				checkConfigCreated,
				checkTags,
				checkSystemdConfigCreated,
			},
		},
		{
			name:        "systemd",
			installType: BINARY_INSTALL,
			options: installOptions{
				installToken: installToken,
			},
			preChecks: []checkFunc{
				checkBinaryNotCreated,
				checkConfigNotCreated,
				checkUserConfigNotCreated,
				checkUserNotExists,
				checkTokenEnvFileNotCreated,
			},
			postChecks: []checkFunc{
				checkBinaryCreated,
				checkBinaryIsRunning,
				checkConfigCreated,
				checkConfigFilesOwnershipAndPermissions,
				checkUserConfigNotCreated,
				checkSystemdConfigCreated,
				checkSystemdEnvDirExists,
				checkSystemdEnvDirPermissions,
				checkTokenEnvFileCreated,
				checkTokenInEnvFile,
				checkUserExists,
				checkVarLogACL,
			},
			conditionalChecks: []condCheckFunc{
				checkSystemdAvailability,
			},
		},
		{
			name:        "systemd",
			installType: PACKAGE_INSTALL,
			options: installOptions{
				apiBaseURL:   mockAPIBaseURL,
				installToken: installToken,
			},
			preChecks: notInstalledChecks,
			postChecks: []checkFunc{
				checkBinaryCreated,
				checkBinaryIsRunning,
				checkConfigCreated,
				checkConfigFilesOwnershipAndPermissions,
				checkUserConfigCreated,
				checkSystemdConfigCreated,
				checkSystemdEnvDirExists,
				checkSystemdEnvDirPermissions,
				checkTokenEnvFileCreated,
				checkTokenInEnvFile,
				checkUserExists,
				checkVarLogACL,
			},
			conditionalChecks: []condCheckFunc{
				checkSystemdAvailability,
			},
		},
		{
			name:        "systemd installation token with existing user directory",
			installType: BINARY_INSTALL,
			options: installOptions{
				installToken: installToken,
			},
			preActions: []checkFunc{
				preActionCreateHomeDirectory,
			},
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
				checkConfigFilesOwnershipAndPermissions,
				checkSystemdConfigCreated,
				checkSystemdEnvDirExists,
				checkSystemdEnvDirPermissions,
				checkTokenEnvFileCreated,
				checkTokenInEnvFile,
				checkUserExists,
				checkVarLogACL,
				checkOutputUserAddWarnings,
			},
			conditionalChecks: []condCheckFunc{
				checkSystemdAvailability,
			},
		},
		{
			name:        "systemd installation token with existing user directory",
			installType: PACKAGE_INSTALL,
			options: installOptions{
				installToken: installToken,
			},
			preActions: []checkFunc{
				preActionCreateHomeDirectory,
			},
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
				checkConfigFilesOwnershipAndPermissions,
				checkSystemdConfigCreated,
				checkSystemdEnvDirExists,
				checkSystemdEnvDirPermissions,
				checkTokenEnvFileCreated,
				checkTokenInEnvFile,
				checkUserExists,
				checkVarLogACL,
				checkOutputUserAddWarnings,
			},
			conditionalChecks: []condCheckFunc{
				checkSystemdAvailability,
			},
			installCode: 3, // because of invalid install token
		},
		{
			name:        "systemd existing installation different token env",
			installType: BINARY_INSTALL,
			options: installOptions{
				installToken: installToken,
			},
			preActions: []checkFunc{
				preActionCreateHomeDirectory,
			},
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
				checkConfigFilesOwnershipAndPermissions,
				checkSystemdConfigCreated,
				checkSystemdEnvDirExists,
				checkSystemdEnvDirPermissions,
				checkTokenEnvFileCreated,
				checkTokenInEnvFile,
				checkUserExists,
				checkVarLogACL,
				checkOutputUserAddWarnings,
			},
			conditionalChecks: []condCheckFunc{
				checkSystemdAvailability,
			},
			installCode: 3, // because of invalid install token
		},
		{
			name:        "systemd existing installation different token env",
			installType: PACKAGE_INSTALL,
			options: installOptions{
				installToken: installToken,
			},
			preActions: []checkFunc{
				preActionCreateHomeDirectory,
			},
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
				checkConfigFilesOwnershipAndPermissions,
				checkSystemdConfigCreated,
				checkSystemdEnvDirExists,
				checkSystemdEnvDirPermissions,
				checkTokenEnvFileCreated,
				checkTokenInEnvFile,
				checkUserExists,
				checkVarLogACL,
				checkOutputUserAddWarnings,
			},
			conditionalChecks: []condCheckFunc{
				checkSystemdAvailability,
			},
			installCode: 3, // because of invalid install token
		},
		{
			name:        "installation of hostmetrics in systemd during upgrade",
			installType: BINARY_INSTALL,
			options: installOptions{
				installToken:       installToken,
				installHostmetrics: true,
			},
			preActions: []checkFunc{
				preActionInstallPreviousVersion,
			},
			conditionalChecks: []condCheckFunc{
				checkSystemdAvailability,
			},
			preChecks: installedWithSystemdChecks,
			postChecks: []checkFunc{
				checkBinaryCreated,
				checkBinaryIsRunning,
				checkConfigCreated,
				checkUserConfigCreated,
				checkTokenInEnvFile,
				checkUserExists,
				checkHostmetricsConfigCreated,
				checkHostmetricsOwnershipAndPermissions,
			},
		},
		{
			name:        "installation of hostmetrics in systemd during upgrade",
			installType: PACKAGE_INSTALL,
			options: installOptions{
				installToken:       installToken,
				installHostmetrics: true,
			},
			preActions: []checkFunc{
				preActionInstallPreviousVersion,
			},
			conditionalChecks: []condCheckFunc{
				checkSystemdAvailability,
			},
			preChecks: installedWithSystemdChecks,
			postChecks: []checkFunc{
				checkBinaryCreated,
				checkBinaryIsRunning,
				checkConfigCreated,
				checkUserConfigCreated,
				checkTokenInEnvFile,
				checkUserExists,
				checkHostmetricsConfigCreated,
				checkHostmetricsOwnershipAndPermissions,
			},
		},
		{
			name:        "uninstallation without autoconfirm fails",
			installType: BINARY_INSTALL,
			options: installOptions{
				uninstall: true,
			},
			preActions: []checkFunc{
				preActionMockStructure,
			},
			preChecks: []checkFunc{
				checkBinaryCreated,
				checkConfigCreated,
				checkUserConfigCreated,
				checkUserNotExists,
			},
			postChecks: []checkFunc{
				checkBinaryCreated,
				checkConfigCreated,
				checkUserConfigCreated,
			},
			installCode: 1,
		},
		{
			name:        "uninstallation without autoconfirm fails",
			installType: PACKAGE_INSTALL,
			options: installOptions{
				uninstall: true,
			},
			preActions: []checkFunc{
				preActionInstallPackage,
			},
			preChecks: installedWithSystemdChecks,
			postChecks: []checkFunc{
				checkBinaryCreated,
				checkConfigCreated,
				checkUserConfigCreated,
			},
			installCode: 1,
		},
		{
			name:        "uninstallation with autoconfirm",
			installType: BINARY_INSTALL,
			options: installOptions{
				autoconfirm: true,
				uninstall:   true,
			},
			preActions: []checkFunc{
				preActionMockStructure,
			},
			preChecks: []checkFunc{
				checkBinaryCreated,
				checkConfigCreated,
				checkUserConfigCreated,
				checkUserNotExists,
			},
			postChecks: []checkFunc{
				checkBinaryNotCreated,
				checkConfigCreated,
				checkUserConfigCreated,
				checkUninstallationOutput,
			},
		},
		{
			name:        "uninstallation with autoconfirm",
			installType: PACKAGE_INSTALL,
			options: installOptions{
				autoconfirm: true,
				uninstall:   true,
			},
			preActions: []checkFunc{
				preActionInstallPackage,
			},
			preChecks: installedWithSystemdChecks,
			postChecks: []checkFunc{
				checkBinaryNotCreated,
				checkUserConfigOrBackupCreated,
				checkUninstallationOutput,
			},
		},
		{
			name:        "systemd uninstallation",
			installType: BINARY_INSTALL,
			options: installOptions{
				autoconfirm: true,
				uninstall:   true,
			},
			preActions: []checkFunc{
				preActionMockSystemdStructure,
			},
			preChecks: []checkFunc{
				checkBinaryCreated,
				checkConfigCreated,
				checkUserConfigCreated,
				checkSystemdConfigCreated,
				checkUserNotExists,
			},
			postChecks: []checkFunc{
				checkBinaryNotCreated,
				checkConfigCreated,
				checkUserConfigCreated,
				checkSystemdConfigCreated,
				checkUserNotExists,
			},
			conditionalChecks: []condCheckFunc{
				checkSystemdAvailability,
			},
		},
		{
			name:        "systemd uninstallation",
			installType: PACKAGE_INSTALL,
			options: installOptions{
				autoconfirm: true,
				uninstall:   true,
			},
			preActions: []checkFunc{
				preActionInstallPreviousVersion,
			},
			preChecks: installedWithSystemdChecks,
			postChecks: []checkFunc{
				checkBinaryNotCreated,
				checkUserConfigOrBackupCreated,
				checkSystemdConfigOrBackupCreated,
				checkUserExists,
			},
			conditionalChecks: []condCheckFunc{
				checkSystemdAvailability,
			},
		},
		{
			name:        "purge",
			installType: BINARY_INSTALL,
			options: installOptions{
				uninstall:   true,
				purge:       true,
				autoconfirm: true,
			},
			preActions: []checkFunc{
				preActionMockStructure,
			},
			preChecks: []checkFunc{
				checkBinaryCreated,
				checkConfigCreated,
				checkUserConfigCreated,
				checkUserNotExists,
			},
			postChecks: []checkFunc{
				checkBinaryNotCreated,
				checkConfigNotCreated,
				checkUserConfigNotCreated,
			},
		},
		{
			name:        "purge",
			installType: PACKAGE_INSTALL,
			options: installOptions{
				uninstall:   true,
				purge:       true,
				autoconfirm: true,
			},
			preActions: []checkFunc{
				preActionInstallPackage,
			},
			preChecks: installedWithSystemdChecks,
			postChecks: []checkFunc{
				checkBinaryNotCreated,
				checkConfigNotCreated,
				checkUserConfigNotCreated,
			},
		},
		{
			name:        "systemd purge",
			installType: BINARY_INSTALL,
			options: installOptions{
				uninstall:   true,
				purge:       true,
				autoconfirm: true,
			},
			preActions: []checkFunc{
				preActionMockSystemdStructure,
			},
			preChecks: []checkFunc{
				checkBinaryCreated,
				checkConfigCreated,
				checkUserConfigCreated,
				checkSystemdConfigCreated,
				checkUserNotExists,
			},
			postChecks: notInstalledChecks,
			conditionalChecks: []condCheckFunc{
				checkSystemdAvailability,
			},
		},
		{
			name:        "systemd purge",
			installType: PACKAGE_INSTALL,
			options: installOptions{
				uninstall:   true,
				purge:       true,
				autoconfirm: true,
			},
			preActions: []checkFunc{
				preActionInstallPreviousVersion,
			},
			preChecks:  installedWithSystemdChecks,
			postChecks: notInstalledChecks,
			conditionalChecks: []condCheckFunc{
				checkSystemdAvailability,
			},
		},
		{
			// This only applies to binary installations as packages
			// will always use the token.env file.
			name:        "systemd creation if token in file",
			installType: BINARY_INSTALL,
			options:     installOptions{},
			preActions: []checkFunc{
				preActionMockUserConfig,
				preActionWriteDifferentTokenToUserConfig,
				preActionWriteDefaultAPIBaseURLToUserConfig,
			},
			preChecks: []checkFunc{
				checkBinaryNotCreated,
				checkConfigNotCreated,
				checkUserConfigCreated,
				checkUserNotExists,
			},
			postChecks: []checkFunc{
				checkBinaryCreated,
				checkBinaryIsRunning,
				checkConfigCreated,
				checkDifferentTokenInConfig,
				checkSystemdConfigCreated,
				checkUserExists,
				checkTokenEnvFileNotCreated,
			},
			conditionalChecks: []condCheckFunc{
				checkSystemdAvailability,
			},
		},
		{
			// This only applies to binary installations as packages
			// will always use the token.env file.
			name:        "systemd installation if token in file",
			installType: BINARY_INSTALL,
			options: installOptions{
				installToken: installToken,
			},
			preActions: []checkFunc{
				preActionWriteDifferentTokenToEnvFile,
			},
			preChecks: []checkFunc{
				checkBinaryNotCreated,
				checkUserConfigNotCreated,
				checkUserNotExists,
			},
			postChecks: []checkFunc{
				checkDifferentTokenInEnvFile,
				checkAbortedDueToDifferentToken,
			},
			conditionalChecks: []condCheckFunc{
				checkSystemdAvailability,
			},
			installCode: 1, // because of invalid installation token
		},
		{
			// This only applies to binary installations as packages
			// will always use the token.env file.
			name:        "systemd installation if deprecated token in file",
			installType: BINARY_INSTALL,
			options: installOptions{
				installToken: installToken,
			},
			preActions: []checkFunc{
				preActionWriteDifferentDeprecatedTokenToEnvFile,
			},
			preChecks: []checkFunc{
				checkBinaryNotCreated,
				checkUserConfigNotCreated,
				checkUserNotExists,
			},
			postChecks: []checkFunc{
				checkDifferentTokenInEnvFile,
				checkAbortedDueToDifferentToken,
			},
			conditionalChecks: []condCheckFunc{
				checkSystemdAvailability,
			},
			installCode: 1, // because of invalid installation token
		},
		{
			name:        "don't keep downloads",
			installType: BINARY_INSTALL,
			options: installOptions{
				skipInstallToken:  true,
				dontKeepDownloads: true,
			},
			preChecks: notInstalledChecks,
			postChecks: []checkFunc{
				checkBinaryCreated,
				checkBinaryIsRunning,
				checkConfigCreated,
				checkUserConfigNotCreated,
				checkSystemdConfigNotCreated,
			},
		},
		{
			name:        "don't keep downloads",
			installType: PACKAGE_INSTALL,
			options: installOptions{
				apiBaseURL:        mockAPIBaseURL,
				installToken:      installToken,
				dontKeepDownloads: true,
			},
			preChecks: notInstalledChecks,
			postChecks: []checkFunc{
				checkBinaryCreated,
				checkBinaryIsRunning,
				checkConfigCreated,
				checkUserConfigCreated,
				checkSystemdConfigCreated,
			},
		},
	} {
		binaryTestSpec := NewTestSpecFromTestSpec(spec)
		packageTestSpec := NewTestSpecFromTestSpec(spec)

		if spec.installType&BINARY_INSTALL != 0 {
			// Run spec for binary installation
			binaryTestSpec.name = fmt.Sprintf("binary install -- %s", spec.name)
			binaryTestSpec.options.useNativePackaging = false

			t.Run(binaryTestSpec.name, func(t *testing.T) {
				runTest(t, &binaryTestSpec)
			})
		}

		if spec.installType&PACKAGE_INSTALL != 0 {
			// Run spec for package installation
			packageTestSpec.name = fmt.Sprintf("package install -- %s", spec.name)
			packageTestSpec.options.useNativePackaging = true

			t.Run(packageTestSpec.name, func(t *testing.T) {
				runTest(t, &packageTestSpec)
			})
		}
	}
}
