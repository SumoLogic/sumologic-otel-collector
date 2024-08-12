package main

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/google/go-cmp/cmp"
)

func TestConfLoader(t *testing.T) {
	tests := []struct {
		Name        string
		FS          fs.FS
		Expected    ConfDir
		ErrExpected bool
	}{
		{
			Name: "load sumologic.yaml",
			FS: func() fs.FS {
				return fstest.MapFS{
					"sumologic.yaml": &fstest.MapFile{
						Data: []byte(`{"extensions":{"sumologic":{}}}`),
					},
				}
			}(),
			Expected: ConfDir{
				Sumologic: []byte(`{"extensions":{"sumologic":{}}}`),
			},
		},
		{
			Name: "load sumologic-remote.yaml",
			FS: func() fs.FS {
				return fstest.MapFS{
					"sumologic-remote.yaml": &fstest.MapFile{
						Data: []byte(`{"extensions":{"opamp":{}}}`),
					},
				}
			}(),
			Expected: ConfDir{
				SumologicRemote: []byte(`{"extensions":{"opamp":{}}}`),
			},
		},
		{
			Name: "load conf.d",
			FS: func() fs.FS {
				return fstest.MapFS{
					"conf.d/a.yaml": &fstest.MapFile{
						Data: []byte(`{"extensions":{"sumologic":{}}}`),
					},
					"conf.d/b.yaml": &fstest.MapFile{
						Data: []byte(`{"extensions":{"opamp":{}}}`),
					},
					"conf.d/emptydir": &fstest.MapFile{
						Mode: fs.ModeDir,
					},
				}
			}(),
			Expected: ConfDir{
				ConfD: map[string][]byte{
					"a.yaml": []byte(`{"extensions":{"sumologic":{}}}`),
					"b.yaml": []byte(`{"extensions":{"opamp":{}}}`),
				},
			},
		},
		{
			Name: "load conf.d-available",
			FS: func() fs.FS {
				return fstest.MapFS{
					"conf.d-available/a.yaml": &fstest.MapFile{
						Data: []byte(`{"extensions":{"sumologic":{}}}`),
					},
					"conf.d-available/b.yaml": &fstest.MapFile{
						Data: []byte(`{"extensions":{"opamp":{}}}`),
					},
					"conf.d-available/emptydir": &fstest.MapFile{
						Mode: fs.ModeDir,
					},
				}
			}(),
			Expected: ConfDir{
				ConfDAvailable: map[string][]byte{
					"a.yaml": []byte(`{"extensions":{"sumologic":{}}}`),
					"b.yaml": []byte(`{"extensions":{"opamp":{}}}`),
				},
			},
		},
		{
			Name: "load all",
			FS: func() fs.FS {
				return fstest.MapFS{
					"sumologic.yaml": &fstest.MapFile{
						Data: []byte(`{"extensions":{"sumologic":{}}}`),
					},
					"sumologic-remote.yaml": &fstest.MapFile{
						Data: []byte(`{"extensions":{"opamp":{}}}`),
					},
					"conf.d/a.yaml": &fstest.MapFile{
						Data: []byte(`{"extensions":{"sumologic":{}}}`),
					},
					"conf.d/b.yaml": &fstest.MapFile{
						Data: []byte(`{"extensions":{"opamp":{}}}`),
					},
					"conf.d/emptydir": &fstest.MapFile{
						Mode: fs.ModeDir,
					},
					"conf.d-available/a.yaml": &fstest.MapFile{
						Data: []byte(`{"extensions":{"sumologic":{}}}`),
					},
					"conf.d-available/b.yaml": &fstest.MapFile{
						Data: []byte(`{"extensions":{"opamp":{}}}`),
					},
					"conf.d-available/emptydir": &fstest.MapFile{
						Mode: fs.ModeDir,
					},
				}
			}(),
			Expected: ConfDir{
				Sumologic:       []byte(`{"extensions":{"sumologic":{}}}`),
				SumologicRemote: []byte(`{"extensions":{"opamp":{}}}`),
				ConfD: map[string][]byte{
					"a.yaml": []byte(`{"extensions":{"sumologic":{}}}`),
					"b.yaml": []byte(`{"extensions":{"opamp":{}}}`),
				},
				ConfDAvailable: map[string][]byte{
					"a.yaml": []byte(`{"extensions":{"sumologic":{}}}`),
					"b.yaml": []byte(`{"extensions":{"opamp":{}}}`),
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			conf, err := ReadConfigDir(test.FS)
			if err != nil {
				if !test.ErrExpected {
					t.Fatal(err)
				}
			}
			if got, want := conf, test.Expected; !cmp.Equal(got, want) {
				t.Errorf("conf dir not as expected: %s", cmp.Diff(want, got))
			}
		})
	}
}

func TestConfLoaderDanglingSymlinks(t *testing.T) {
	tempdir := t.TempDir()

	if err := os.Mkdir(filepath.Join(tempdir, ConfDotD), 0700); err != nil {
		t.Fatal(err)
	}

	if err := os.Mkdir(filepath.Join(tempdir, ConfDotDAvailable), 0700); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(tempdir, ConfDotD, ConfDSettings), []byte("extensions:\n  sumologic:\n    installation_token: abcdef\n"), 0600); err != nil {
		t.Fatal(err)
	}

	if err := os.Symlink(filepath.Join(tempdir, ConfDotDAvailable, "foobar.yaml"), filepath.Join(tempdir, ConfDotD, "foobar.yaml")); err != nil {
		t.Fatal(err)
	}

	conf, err := ReadConfigDir(os.DirFS(tempdir))
	if err != nil {
		t.Fatal(err)
	}

	exp := ConfDir{
		ConfD: map[string][]byte{
			ConfDSettings: []byte("extensions:\n  sumologic:\n    installation_token: abcdef\n"),
		},
		ConfDAvailable: map[string][]byte{},
	}

	if !cmp.Equal(conf, exp) {
		t.Errorf("conf dir not as expected: %s", cmp.Diff(exp, conf))
	}
}
