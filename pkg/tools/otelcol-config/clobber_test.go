package main

import (
	"io"
	"io/fs"
	"testing"
	"testing/fstest"
)

func TestClobberAction(t *testing.T) {

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
			Name:           "[enabled] no existing setting",
			Flags:          []string{"--enable-clobber"},
			Conf:           fstest.MapFS{},
			ExpectedWriter: []byte("extensions:\n  sumologic:\n    clobber: true\n"),
		},
		{
			Name:           "[enabled] no existing setting, override",
			Flags:          []string{"--enable-clobber", "--override"},
			Conf:           fstest.MapFS{},
			ExpectedWriter: []byte("extensions:\n  sumologic:\n    clobber: true\n"),
		},
		{
			Name:  "[enabled] remote control with existing file",
			Flags: []string{"--enable-clobber"},
			Conf: fstest.MapFS{
				SumologicRemoteDotYaml: &fstest.MapFile{
					Data: []byte("extensions:\n  opamp:\n    enabled: true\n"),
				},
			},
			ExpectedWriter: []byte("extensions:\n  opamp:\n    enabled: true\n  sumologic:\n    clobber: true\n"),
		},
		{
			Name:           "[enabled] remote control with no existing file, but flag exists",
			Flags:          []string{"--enable-clobber", "--enable-remote-control"},
			Conf:           fstest.MapFS{},
			ExpectedWriter: []byte("extensions:\n  sumologic:\n    clobber: true\n"),
		},
		{
			Name:           "[disabled] no existing setting",
			Flags:          []string{"--disable-clobber"},
			Conf:           fstest.MapFS{},
			ExpectedWriter: []byte("extensions:\n  sumologic:\n    clobber: false\n"),
		},
		{
			Name:           "[disabled] no existing setting, override",
			Flags:          []string{"--disable-clobber", "--override"},
			Conf:           fstest.MapFS{},
			ExpectedWriter: []byte("extensions:\n  sumologic:\n    clobber: false\n"),
		},
		{
			Name:  "[disabled] remote control with existing file",
			Flags: []string{"--disable-clobber"},
			Conf: fstest.MapFS{
				SumologicRemoteDotYaml: &fstest.MapFile{
					Data: []byte("extensions:\n  opamp:\n    enabled: true\n"),
				},
			},
			ExpectedWriter: []byte("extensions:\n  opamp:\n    enabled: true\n  sumologic:\n    clobber: false\n"),
		},
		{
			Name:           "[disabled] remote control with no existing file, but flag exists",
			Flags:          []string{"--disable-clobber", "--enable-remote-control"},
			Conf:           fstest.MapFS{},
			ExpectedWriter: []byte("extensions:\n  sumologic:\n    clobber: false\n"),
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

			err := EnableClobberAction(ctx)
			if err != nil && !test.ExpectedErr {
				t.Fatal(err)
			}
			if err == nil && test.ExpectedErr {
				t.Fatal("expected non-nil error")
			}
		})
	}
}
