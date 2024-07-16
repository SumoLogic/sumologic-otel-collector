package main

import (
	"io"
	"io/fs"
	"testing"
	"testing/fstest"
)

func TestSetOpAmpEndpointAction(t *testing.T) {
	tests := []struct {
		Name     string
		Flags    []string
		Conf     fs.FS
		ExpWrite []byte
		ExpErr   bool
	}{
		{
			Name:   "no sumologic-remote.yaml",
			Flags:  []string{"--set-opamp-endpoint", "wss://example.com"},
			Conf:   fstest.MapFS{},
			ExpErr: true,
		},
		{
			Name:  "success",
			Flags: []string{"--set-opamp-endpoint", "wss://example.com"},
			Conf: fstest.MapFS{
				SumologicRemoteDotYaml: &fstest.MapFile{
					Data: []byte("extensions:\n  opamp:\n    enabled: true\n    endpoint: wss://example.com/wrong\n"),
				},
			},
			ExpWrite: []byte("extensions:\n  opamp:\n    enabled: true\n    endpoint: wss://example.com\n"),
		},
		{
			Name:   "invalid URL",
			Flags:  []string{"--set-opamp-endpoint", "https://example.com"},
			Conf:   fstest.MapFS{},
			ExpErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			flagValues := newFlagValues()
			fs := makeFlagSet(flagValues)

			if err := fs.Parse(test.Flags); err != nil {
				t.Fatal(err)
			}
			ctx := &actionContext{
				ConfigDir:            test.Conf,
				Flags:                flagValues,
				Stdout:               io.Discard,
				Stderr:               io.Discard,
				WriteConfD:           errWriter{}.Write,
				WriteConfDOverrides:  errWriter{}.Write,
				WriteSumologicRemote: newTestWriter(test.ExpWrite).Write,
			}

			err := SetOpAmpEndpointAction(ctx)
			if err != nil && !test.ExpErr {
				t.Fatal(err)
			}
			if err == nil && test.ExpErr {
				t.Fatal("expected non-nil error")
			}

		})
	}
}
