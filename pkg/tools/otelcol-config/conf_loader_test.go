package main

import (
	"io/fs"
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
