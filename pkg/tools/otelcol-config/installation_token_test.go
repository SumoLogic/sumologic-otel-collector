package main

import (
	"io"
	"io/fs"
	"path"
	"testing"
	"testing/fstest"
)

func TestSetInstallationTokenAction(t *testing.T) {
	tests := []struct {
		Name           string
		Flags          []string
		Conf           fs.FS
		Launchd        fs.FS
		SystemdEnabled bool
		LaunchdEnabled bool
		ExpWrite       []byte
		ExpErr         bool
	}{
		{
			Name:     "no existing settings",
			Flags:    []string{"--set-installation-token", "abcdef"},
			Conf:     fstest.MapFS{},
			ExpWrite: []byte("extensions:\n  sumologic:\n    installation_token: abcdef\n"),
		},
		{
			Name:  "existing settings",
			Flags: []string{"--set-installation-token", "abcdef"},
			Conf: fstest.MapFS{
				path.Join(ConfDotD, ConfDSettings): &fstest.MapFile{
					Data: []byte("extensions:\n  sumologic:\n    installation_token: token\n    foo: bar\n"),
				},
			},
			ExpWrite: []byte("extensions:\n  sumologic:\n    installation_token: abcdef\n    foo: bar\n"),
		},
		{
			Name:     "no existing settings, override",
			Flags:    []string{"--set-installation-token", "abcdef", "--override"},
			Conf:     fstest.MapFS{},
			ExpWrite: []byte("extensions:\n  sumologic:\n    installation_token: abcdef\n"),
		},
		{
			Name:  "existing settings, override",
			Flags: []string{"--set-installation-token", "abcdef", "--override"},
			Conf: fstest.MapFS{
				path.Join(ConfDotD, ConfDOverrides): &fstest.MapFile{
					Data: []byte("extensions:\n  sumologic:\n    installation_token: token\n    foo: bar\n"),
				},
			},
			ExpWrite: []byte("extensions:\n  sumologic:\n    installation_token: abcdef\n    foo: bar\n"),
		},
		{
			Name:  "remote control with existing file",
			Flags: []string{"--set-installation-token", "abcdef"},
			Conf: fstest.MapFS{
				SumologicRemoteDotYaml: &fstest.MapFile{
					Data: []byte("extensions:\n  opamp:\n    enabled: true\n"),
				},
			},
			ExpWrite: []byte("extensions:\n  opamp:\n    enabled: true\n  sumologic:\n    installation_token: abcdef\n"),
		},
		{
			Name:     "remote control with no existing file, but flag exists",
			Flags:    []string{"--set-installation-token", "abcdef", "--enable-remote-control"},
			Conf:     fstest.MapFS{},
			ExpWrite: []byte("extensions:\n  sumologic:\n    installation_token: abcdef\n"),
		},
		{
			Name:           "systemd",
			Flags:          []string{"--set-installation-token", "abcdef"},
			Conf:           fstest.MapFS{},
			ExpWrite:       []byte("SUMOLOGIC_INSTALLATION_TOKEN=abcdef\n"),
			SystemdEnabled: true,
		},
		{
			Name:  "launchd",
			Flags: []string{"--set-installation-token", "abcdef"},
			Launchd: fstest.MapFS{
				launchdConfigPlist: &fstest.MapFile{
					Data: []byte(defaultLaunchdConfigXML),
				},
			},
			ExpWrite: []byte(xmlPreamble +
				`<plist version="1.0">` +
				`<dict>` +
				`<key>EnvironmentVariables</key>` +
				`<dict>` +
				`<key>SUMOLOGIC_INSTALLATION_TOKEN</key>` +
				`<string>abcdef</string>` +
				`</dict>` +
				`<key>GroupName</key>` +
				`<string>_otelcol-sumo</string>` +
				`<key>KeepAlive</key>` +
				`<true/>` +
				`<key>Label</key>` +
				`<string>otelcol-sumo</string>` +
				`<key>ProgramArguments</key>` +
				`<array>` +
				`<string>/usr/share/otelcol-sumo.sh</string>` +
				`</array>` +
				`<key>RunAtLoad</key>` +
				`<true/>` +
				`<key>StandardErrorPath</key>` +
				`<string>/var/log/otelcol-sumo/otelcol-sumo.log</string>` +
				`<key>StandardOutPath</key>` +
				`<string>/var/log/otelcol-sumo/otelcol-sumo.log</string>` +
				`<key>UserName</key>` +
				`<string>_otelcol-sumo</string>` +
				`</dict>` +
				`</plist>`),
			LaunchdEnabled: true,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			writer := newTestWriter(test.ExpWrite).Write
			errWriter := errWriter{}.Write

			flagValues := newFlagValues()
			flagSet := makeFlagSet(flagValues)

			if err := flagSet.Parse(test.Flags); err != nil {
				t.Fatal(err)
			}

			var (
				settingsWriter,
				overridesWriter,
				sumologicRemoteWriter,
				tokenEnvWriter,
				launchdWriter func([]byte) (int, error)
			)

			tokenEnvWriter = errWriter
			switch {
			case test.SystemdEnabled:
				tokenEnvWriter = writer
				settingsWriter = errWriter
				sumologicRemoteWriter = errWriter
				overridesWriter = errWriter
			case test.LaunchdEnabled:
				launchdWriter = writer
				settingsWriter = errWriter
				sumologicRemoteWriter = errWriter
				overridesWriter = errWriter
			default:
				if flagValues.Override {
					settingsWriter = errWriter
					sumologicRemoteWriter = errWriter
					overridesWriter = writer
				} else if flagValues.EnableRemoteControl || remoteControlEnabled(t, test.Conf) {
					settingsWriter = errWriter
					overridesWriter = errWriter
					sumologicRemoteWriter = writer
				} else {
					overridesWriter = errWriter
					sumologicRemoteWriter = errWriter
					settingsWriter = writer
				}
			}

			ctx := &actionContext{
				ConfigDir:                 test.Conf,
				LaunchdDir:                test.Launchd,
				Flags:                     flagValues,
				Stdout:                    io.Discard,
				Stderr:                    io.Discard,
				WriteConfD:                settingsWriter,
				WriteConfDOverrides:       overridesWriter,
				WriteSumologicRemote:      sumologicRemoteWriter,
				WriteInstallationTokenEnv: tokenEnvWriter,
				WriteLaunchdConfig:        launchdWriter,
				SystemdEnabled:            test.SystemdEnabled,
				LaunchdEnabled:            test.LaunchdEnabled,
			}

			err := SetInstallationTokenAction(ctx)

			if err != nil && !test.ExpErr {
				t.Fatal(err)
			}
			if err == nil && test.ExpErr {
				t.Fatal("expected non-nil error")
			}

		})
	}
}

func remoteControlEnabled(t *testing.T, confFS fs.FS) bool {
	t.Helper()
	conf, err := ReadConfigDir(confFS)
	if err != nil {
		t.Fatal(err)
	}
	return conf.SumologicRemote != nil
}
