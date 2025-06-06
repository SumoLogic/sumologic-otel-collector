package main

import (
	"io"
	"io/fs"
	"testing"
	"testing/fstest"
)

func TestSetTimezoneActionAction(t *testing.T) {
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
			Flags:    []string{"--set-timezone", "UTC"},
			Conf:     fstest.MapFS{},
			ExpWrite: []byte("extensions:\n  sumologic:\n    time_zone: UTC\n"),
		},
		{
			Name:     "no existing settings, override",
			Flags:    []string{"--set-timezone", "UTC", "--override"},
			Conf:     fstest.MapFS{},
			ExpWrite: []byte("extensions:\n  sumologic:\n    time_zone: UTC\n"),
		},
		{
			Name:  "remote control with existing file",
			Flags: []string{"--set-timezone", "UTC"},
			Conf: fstest.MapFS{
				SumologicRemoteDotYaml: &fstest.MapFile{
					Data: []byte("extensions:\n  opamp:\n    enabled: true\n"),
				},
			},
			ExpWrite: []byte("extensions:\n  opamp:\n    enabled: true\n  sumologic:\n    time_zone: UTC\n"),
		},
		{
			Name:     "remote control with no existing file, but flag exists",
			Flags:    []string{"--set-timezone", "UTC", "--enable-remote-control"},
			Conf:     fstest.MapFS{},
			ExpWrite: []byte("extensions:\n  sumologic:\n    time_zone: UTC\n"),
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

			err := SetTimezoneAction(ctx)

			if err != nil && !test.ExpErr {
				t.Fatal(err)
			}
			if err == nil && test.ExpErr {
				t.Fatal("expected non-nil error")
			}

		})
	}
}
