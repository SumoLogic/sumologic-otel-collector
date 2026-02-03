package main

import (
	"io"
	"io/fs"
	"strings"
	"testing"
	"testing/fstest"
)

func TestSetCollectorNameAction(t *testing.T) {

	longName := strings.Repeat("a", 115)

	tests := []struct {
		Name           string
		Flags          []string
		Conf           fs.FS
		Launchd        fs.FS
		SystemdEnabled bool
		LaunchdEnabled bool
		ExpectedWriter []byte
		ExpectedErr    bool
	}{
		{
			Name:           "[enabled] invalid name",
			Flags:          []string{"--set-collector-name", "my-collector?"},
			Conf:           fstest.MapFS{},
			ExpectedWriter: []byte("extensions:\n  sumologic:\n    collector_name: my-collector?\n"),
			ExpectedErr:    true,
		},
		{
			Name:           "[enabled] valid name with special char =",
			Flags:          []string{"--set-collector-name", "my-collector="},
			Conf:           fstest.MapFS{},
			ExpectedWriter: []byte("extensions:\n  sumologic:\n    collector_name: my-collector=\n"),
			ExpectedErr:    false,
		},
		{
			Name:           "[enabled] no existing setting",
			Flags:          []string{"--set-collector-name", "my-collector"},
			Conf:           fstest.MapFS{},
			ExpectedWriter: []byte("extensions:\n  sumologic:\n    collector_name: my-collector\n"),
		},
		{
			Name:           "[enabled] no existing setting, override",
			Flags:          []string{"--set-collector-name", "my-collector"},
			Conf:           fstest.MapFS{},
			ExpectedWriter: []byte("extensions:\n  sumologic:\n    collector_name: my-collector\n"),
		},
		{
			Name:  "[enabled] remote control with existing file",
			Flags: []string{"--set-collector-name", "my-collector"},
			Conf: fstest.MapFS{
				SumologicRemoteDotYaml: &fstest.MapFile{
					Data: []byte("extensions:\n  opamp:\n    enabled: true\n"),
				},
			},
			ExpectedWriter: []byte("extensions:\n  opamp:\n    enabled: true\n  sumologic:\n    collector_name: my-collector\n"),
		},
		{
			Name:           "[enabled] remote control with no existing file, but flag exists",
			Flags:          []string{"--set-collector-name", "my-collector"},
			Conf:           fstest.MapFS{},
			ExpectedWriter: []byte("extensions:\n  sumologic:\n    collector_name: my-collector\n"),
		},
		{
			Name:        "[enabled]  collector name length limit exceeded",
			Flags:       []string{"--set-collector-name", longName},
			Conf:        fstest.MapFS{},
			ExpectedErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			writer := newTestWriter(test.ExpectedWriter).Write
			errWriter := errWriter{}.Write

			flagValues := newFlagValues()
			flagSet := makeFlagSet(flagValues)

			if err := flagSet.Parse(test.Flags); err != nil {
				t.Fatalf("failed to parse flags: %v", err)
			}

			var (
				settingsWriter, overridesWriter, sumologicRemoteWriter, tokenEnvWriter, launchdWriter func([]byte) (int, error)
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

			err := SetCollectorNameAction(ctx)
			if err != nil && !test.ExpectedErr {
				t.Fatal(err)
			}
			if err == nil && test.ExpectedErr {
				t.Fatal("expected non-nil error")
			}
		})
	}
}
