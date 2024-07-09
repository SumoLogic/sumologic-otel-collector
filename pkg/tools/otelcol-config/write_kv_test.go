package main

import (
	"bytes"
	"io/fs"
	"path"
	"testing"
	"testing/fstest"

	"golang.org/x/exp/slices"
)

func TestWriteKVAction(t *testing.T) {
	tests := []struct {
		Name     string
		Conf     fs.FS
		Flags    []string
		ExpWrite []byte
		ExpErr   bool
	}{
		{
			Name:     "create",
			Conf:     fstest.MapFS{},
			Flags:    []string{"--write-kv", `.foo.bar.baz = "hotdog"`},
			ExpWrite: []byte("foo:\n  bar:\n    baz: hotdog\n"),
		},
		{
			Name: "create_existing",
			Conf: fstest.MapFS{
				path.Join(ConfDotD, ConfDSettings): &fstest.MapFile{
					Data: []byte("foo:\n  bar:\n    foobar: foo\n"),
				},
			},
			Flags:    []string{"--write-kv", `.foo.bar.baz = "hotdog"`},
			ExpWrite: []byte("foo:\n  bar:\n    foobar: foo\n    baz: hotdog\n"),
		},
		{
			Name: "update",
			Conf: fstest.MapFS{
				path.Join(ConfDotD, ConfDSettings): &fstest.MapFile{
					Data: []byte(`{"foo":{"bar":{"baz":"hamburger"}}}`),
				},
			},
			Flags:    []string{"--write-kv", `.foo.bar.baz = "hotdog"`},
			ExpWrite: []byte("{\"foo\": {\"bar\": {\"baz\": \"hotdog\"}}}\n"),
		},
		{
			Name: "append",
			Conf: fstest.MapFS{
				path.Join(ConfDotD, ConfDSettings): &fstest.MapFile{
					Data: []byte("foo:\n  bar:\n    baz:\n      - hamburger"),
				},
			},
			Flags:    []string{"--write-kv", `.foo.bar.baz += ["fries"]`},
			ExpWrite: []byte("foo:\n  bar:\n    baz:\n      - hamburger\n      - fries\n"),
		},
		{
			Name: "create_override",
			Conf: fstest.MapFS{
				path.Join(ConfDotD, ConfDSettings): &fstest.MapFile{ // confd settings file expected to be ignored
					Data: []byte("foo:\n  bar:\n    foobar: foo\n"),
				},
			},
			Flags:    []string{"--write-kv", `.foo.bar.baz = "hotdog"`, "--override"},
			ExpWrite: []byte("foo:\n  bar:\n    baz: hotdog\n"),
		},
		{
			Name: "create_existing_override",
			Conf: fstest.MapFS{
				path.Join(ConfDotD, ConfDSettings): &fstest.MapFile{ // confd settings file expected to be ignored
					Data: []byte("foo: bar\n"),
				},
				path.Join(ConfDotD, ConfDOverrides): &fstest.MapFile{
					Data: []byte("foo:\n  bar:\n    foobar: foo\n"),
				},
			},
			Flags:    []string{"--write-kv", `.foo.bar.baz = "hotdog"`, "--override"},
			ExpWrite: []byte("foo:\n  bar:\n    foobar: foo\n    baz: hotdog\n"),
		},
		{
			Name: "update_override",
			Conf: fstest.MapFS{
				path.Join(ConfDotD, ConfDOverrides): &fstest.MapFile{
					Data: []byte(`{"foo":{"bar":{"baz":"hamburger"}}}`),
				},
			},
			Flags:    []string{"--write-kv", `.foo.bar.baz = "hotdog"`, "--override"},
			ExpWrite: []byte("{\"foo\": {\"bar\": {\"baz\": \"hotdog\"}}}\n"),
		},
		{
			Name: "append_override",
			Conf: fstest.MapFS{
				path.Join(ConfDotD, ConfDOverrides): &fstest.MapFile{
					Data: []byte("foo:\n  bar:\n    baz:\n      - hamburger"),
				},
			},
			Flags:    []string{"--write-kv", `.foo.bar.baz += ["fries"]`, "--override"},
			ExpWrite: []byte("foo:\n  bar:\n    baz:\n      - hamburger\n      - fries\n"),
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			// writes to settings must be as expected
			settingsWriter := newTestWriter(test.ExpWrite).Write

			// writes to overrides are not allowed
			overridesWriter := errWriter{}.Write

			// in override mode, writes to settings are not allowed, and writes
			// to overrides are.
			if slices.Contains(test.Flags, "--override") {
				settingsWriter, overridesWriter = overridesWriter, settingsWriter
			}
			stdoutBuf := new(bytes.Buffer)
			stderrBuf := new(bytes.Buffer)
			actionContext := makeTestActionContext(t,
				test.Conf,
				test.Flags,
				stdoutBuf,
				stderrBuf,
				settingsWriter,
				overridesWriter,
			)
			err := WriteKVAction(actionContext)
			if test.ExpErr && err == nil {
				t.Fatal("expected non-nil error")
			}
			if !test.ExpErr && err != nil {
				t.Fatal(err)
			}
		})
	}
}
