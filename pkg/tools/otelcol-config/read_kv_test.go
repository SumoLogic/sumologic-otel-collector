package main

import (
	"bytes"
	"io/fs"
	"path"
	"testing"
	"testing/fstest"
)

func TestReadKVAction(t *testing.T) {
	tests := []struct {
		Name      string
		Conf      fs.FS
		Flags     []string
		ExpStdout string
		ExpStderr string
		ExpErr    bool
	}{
		{
			Name: "precedence",
			Conf: fstest.MapFS{
				// note that foo is overridden and read-kv should return baz
				path.Join(ConfDotD, "00-fst.yaml"): &fstest.MapFile{
					Data: []byte(`{"sumologic":{"foo":"bar"}}`),
				},
				path.Join(ConfDotD, "01-snd.yaml"): &fstest.MapFile{
					Data: []byte(`{"sumologic":{"foo":"baz"}}`),
				},
			},
			Flags: []string{
				"--read-kv",
				".sumologic.foo",
			},
			ExpStdout: "baz\n",
		},
		{
			Name: "missing values ignored",
			Conf: fstest.MapFS{
				path.Join(ConfDotD, "00-fst.yaml"): &fstest.MapFile{
					Data: []byte(`{"sumologic":{"foo":"bar"}}`),
				},
				path.Join(ConfDotD, "01-snd.yaml"): &fstest.MapFile{
					Data: []byte(`{"sumologic":{}}`),
				},
			},
			Flags: []string{
				"--read-kv",
				".sumologic.foo",
			},
			ExpStdout: "bar\n",
		},
		{
			Name: "explicit nulls ignored",
			Conf: fstest.MapFS{
				path.Join(ConfDotD, "00-fst.yaml"): &fstest.MapFile{
					Data: []byte(`{"sumologic":{"foo":"bar"}}`),
				},
				path.Join(ConfDotD, "01-snd.yaml"): &fstest.MapFile{
					Data: []byte(`{"sumologic":{"foo":null}}`),
				},
			},
			Flags: []string{
				"--read-kv",
				".sumologic.foo",
			},
			ExpStdout: "bar\n",
		},
		{
			Name: "key missing",
			Conf: fstest.MapFS{
				path.Join(ConfDotD, "00-fst.yaml"): &fstest.MapFile{
					Data: []byte(`{"sumologic":{"foo":"bar"}}`),
				},
			},
			Flags: []string{
				"--read-kv",
				".sumologic.bar",
			},
			ExpStdout: "null\n",
		},
		{
			Name: "complex result",
			Conf: fstest.MapFS{
				path.Join(ConfDotD, "00-fst.yaml"): &fstest.MapFile{
					Data: []byte(`{"sumologic":{"foo":"bar"}}`),
				},
			},
			Flags: []string{
				"--read-kv",
				".sumologic",
			},
			ExpStdout: "{\"foo\": \"bar\"}\n",
		},
		{
			Name: "invalid expression",
			Conf: fstest.MapFS{
				path.Join(ConfDotD, "00-fst.yaml"): &fstest.MapFile{
					Data: []byte(`{"sumologic":{"foo":"bar"}}`),
				},
			},
			Flags: []string{
				"--read-kv",
				"asodfijasld!!!fi",
			},
			ExpErr: true,
		},
	}
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			stdoutBuf := new(bytes.Buffer)
			stderrBuf := new(bytes.Buffer)
			actionContext := makeTestActionContext(t,
				test.Conf,
				test.Flags,
				stdoutBuf,
				stderrBuf,
				errWriter{}.Write,
				errWriter{}.Write,
			)
			err := ReadKVAction(actionContext)
			if test.ExpErr && err == nil {
				t.Fatal("expected non-nil error")
			}
			if !test.ExpErr && err != nil {
				t.Fatal(err)
			}
			if got, want := stdoutBuf.String(), test.ExpStdout; got != want {
				t.Errorf("bad read-kv: got %q, want %q", got, want)
			}
			if got, want := stderrBuf.String(), test.ExpStderr; got != want {
				t.Errorf("%q", got)
			}
		})
	}
}
