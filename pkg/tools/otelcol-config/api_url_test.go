package main

import (
	"io"
	"io/fs"
	"path"
	"testing"
	"testing/fstest"
)

func TestSetSetAPIURLAction(t *testing.T) {
	tests := []struct {
		Name     string
		Flags    []string
		Conf     fs.FS
		ExpWrite []byte
		ExpErr   bool
	}{
		{
			Name:     "no existing settings",
			Flags:    []string{"--set-api-url", "http://example.com"},
			Conf:     fstest.MapFS{},
			ExpWrite: []byte("extensions:\n  sumologic:\n    api_base_url: http://example.com\n"),
		},
		{
			Name:  "existing settings",
			Flags: []string{"--set-api-url", "https://example.com"},
			Conf: fstest.MapFS{
				path.Join(ConfDotD, ConfDSettings): &fstest.MapFile{
					Data: []byte("extensions:\n  sumologic:\n    api_base_url: https://example.com\n    foo: bar\n"),
				},
			},
			ExpWrite: []byte("extensions:\n  sumologic:\n    api_base_url: https://example.com\n    foo: bar\n"),
		},
		{
			Name:     "no existing settings, override",
			Flags:    []string{"--set-api-url", "http://example.com", "--override"},
			Conf:     fstest.MapFS{},
			ExpWrite: []byte("extensions:\n  sumologic:\n    api_base_url: http://example.com\n"),
		},
		{
			Name:  "existing settings, override",
			Flags: []string{"--set-api-url", "http://example.com", "--override"},
			Conf: fstest.MapFS{
				path.Join(ConfDotD, ConfDOverrides): &fstest.MapFile{
					Data: []byte("extensions:\n  sumologic:\n    api_base_url: http://example.com\n    foo: bar\n"),
				},
			},
			ExpWrite: []byte("extensions:\n  sumologic:\n    api_base_url: http://example.com\n    foo: bar\n"),
		},
		{
			Name:  "remote control with existing file",
			Flags: []string{"--set-api-url", "http://example.com"},
			Conf: fstest.MapFS{
				SumologicRemoteDotYaml: &fstest.MapFile{
					Data: []byte("extensions:\n  opamp:\n    enabled: true\n"),
				},
			},
			ExpWrite: []byte("extensions:\n  opamp:\n    enabled: true\n  sumologic:\n    api_base_url: http://example.com\n"),
		},
		{
			Name:     "remote control with no existing file, but flag exists",
			Flags:    []string{"--set-api-url", "http://example.com", "--enable-remote-control"},
			Conf:     fstest.MapFS{},
			ExpWrite: []byte("extensions:\n  sumologic:\n    api_base_url: http://example.com\n"),
		},
		{
			Name:   "invalid URL scheme",
			Flags:  []string{"--set-api-url", "ws://example.com"},
			Conf:   fstest.MapFS{},
			ExpErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			writer := newTestWriter(test.ExpWrite).Write
			errWriter := errWriter{}.Write

			flagValues := newFlagValues()
			fs := makeFlagSet(flagValues)

			if err := fs.Parse(test.Flags); err != nil {
				t.Fatal(err)
			}

			var settingsWriter, overridesWriter, sumologicRemoteWriter func([]byte) (int, error)

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

			ctx := &actionContext{
				ConfigDir:            test.Conf,
				Flags:                flagValues,
				Stdout:               io.Discard,
				Stderr:               io.Discard,
				WriteConfD:           settingsWriter,
				WriteConfDOverrides:  overridesWriter,
				WriteSumologicRemote: sumologicRemoteWriter,
			}

			err := SetAPIURLAction(ctx)

			if err != nil && !test.ExpErr {
				t.Fatal(err)
			}
			if err == nil && test.ExpErr {
				t.Fatal("expected non-nil error")
			}

		})
	}
}
