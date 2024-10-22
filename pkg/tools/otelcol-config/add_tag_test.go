package main

import (
	"bytes"
	"io/fs"
	"path"
	"testing"
	"testing/fstest"

	"golang.org/x/exp/slices"
)

func TestAddTagAction(t *testing.T) {
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
			Flags:    []string{"--add-tag", `foo=bar`},
			ExpWrite: []byte("extensions:\n  sumologic:\n    collector_fields:\n      foo: bar\n"),
		},
		{
			Name: "create with existing",
			Conf: fstest.MapFS{
				path.Join(ConfDotD, ConfDSettings): &fstest.MapFile{
					Data: []byte("extensions:\n  sumologic:\n    collector_fields:\n      foo: bar\n"),
				},
			},
			Flags:    []string{"--add-tag", `bar=baz`},
			ExpWrite: []byte("extensions:\n  sumologic:\n    collector_fields:\n      foo: bar\n      bar: baz\n"),
		},
		{
			Name: "update",
			Conf: fstest.MapFS{
				path.Join(ConfDotD, ConfDSettings): &fstest.MapFile{
					Data: []byte("extensions:\n  sumologic:\n    collector_fields:\n      foo: bar\n"),
				},
			},
			Flags:    []string{"--add-tag", `foo=baz`},
			ExpWrite: []byte("extensions:\n  sumologic:\n    collector_fields:\n      foo: baz\n"),
		},
		{
			Name:     "numeric",
			Conf:     fstest.MapFS{},
			Flags:    []string{"--add-tag", `foo=5`},
			ExpWrite: []byte("extensions:\n  sumologic:\n    collector_fields:\n      foo: 5\n"),
		},
		{
			Name:     "object",
			Conf:     fstest.MapFS{},
			Flags:    []string{"--add-tag", `foo={"bar":"baz"}`},
			ExpWrite: []byte("extensions:\n  sumologic:\n    collector_fields:\n      foo:\n        bar: baz\n"),
		},
		{
			Name:     "array",
			Conf:     fstest.MapFS{},
			Flags:    []string{"--add-tag", `foo=[1,2,3,4]`},
			ExpWrite: []byte("extensions:\n  sumologic:\n    collector_fields:\n      foo:\n        - 1\n        - 2\n        - 3\n        - 4\n"),
		},
		{
			Name:     "create override",
			Conf:     fstest.MapFS{},
			Flags:    []string{"--add-tag", `foo=bar`, "--override"},
			ExpWrite: []byte("extensions:\n  sumologic:\n    collector_fields:\n      foo: bar\n"),
		},
		{
			Name: "create with existing override",
			Conf: fstest.MapFS{
				path.Join(ConfDotD, ConfDOverrides): &fstest.MapFile{
					Data: []byte("extensions:\n  sumologic:\n    collector_fields:\n      foo: bar\n"),
				},
			},
			Flags:    []string{"--add-tag", `bar=baz`, "--override"},
			ExpWrite: []byte("extensions:\n  sumologic:\n    collector_fields:\n      foo: bar\n      bar: baz\n"),
		},
		{
			Name: "update override",
			Conf: fstest.MapFS{
				path.Join(ConfDotD, ConfDOverrides): &fstest.MapFile{
					Data: []byte("extensions:\n  sumologic:\n    collector_fields:\n      foo: bar\n"),
				},
			},
			Flags:    []string{"--add-tag", `foo=baz`, "--override"},
			ExpWrite: []byte("extensions:\n  sumologic:\n    collector_fields:\n      foo: baz\n"),
		},
		{
			Name:     "numeric override",
			Conf:     fstest.MapFS{},
			Flags:    []string{"--add-tag", `foo=5`, "--override"},
			ExpWrite: []byte("extensions:\n  sumologic:\n    collector_fields:\n      foo: 5\n"),
		},
		{
			Name:     "object override",
			Conf:     fstest.MapFS{},
			Flags:    []string{"--add-tag", `foo={"bar":"baz"}`, "--override"},
			ExpWrite: []byte("extensions:\n  sumologic:\n    collector_fields:\n      foo:\n        bar: baz\n"),
		},
		{
			Name:     "array override",
			Conf:     fstest.MapFS{},
			Flags:    []string{"--add-tag", `foo=[1,2,3,4]`, "--override"},
			ExpWrite: []byte("extensions:\n  sumologic:\n    collector_fields:\n      foo:\n        - 1\n        - 2\n        - 3\n        - 4\n"),
		},
		{
			Name:     "dot in tag name",
			Conf:     fstest.MapFS{},
			Flags:    []string{"--add-tag", "foo.bar=baz"},
			ExpWrite: []byte("extensions:\n  sumologic:\n    collector_fields:\n      foo.bar: baz\n"),
		},
		{
			Name: "add tag to sumologic-remote.yaml",
			Conf: fstest.MapFS{
				SumologicRemoteDotYaml: &fstest.MapFile{
					Data: []byte("extensions:\n  sumologic:\n    collector_fields:\n      foo: bar\n"),
				},
			},
			Flags:    []string{"--add-tag", "bar=baz"},
			ExpWrite: []byte("extensions:\n  sumologic:\n    collector_fields:\n      foo: bar\n      bar: baz\n"),
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
			if _, ok := test.Conf.(fstest.MapFS)[SumologicRemoteDotYaml]; ok {
				actionContext.WriteConfD = errWriter{}.Write
				actionContext.WriteConfDOverrides = errWriter{}.Write
				actionContext.WriteSumologicRemote = newTestWriter(test.ExpWrite).Write
			}
			err := AddTagAction(actionContext)
			if test.ExpErr && err == nil {
				t.Fatal("expected non-nil error")
			}
			if !test.ExpErr && err != nil {
				t.Fatal(err)
			}
		})
	}
}
