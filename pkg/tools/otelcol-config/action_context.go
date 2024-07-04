package main

import (
	"io"
	"io/fs"
	"os"
)

type actionContext struct {
	ConfigDir           fs.FS
	Flags               *flagValues
	Stdout              io.Writer
	Stderr              io.Writer
	WriteConfD          func([]byte) error
	WriteConfDOverrides func([]byte) error
}

func makeActionContext(values *flagValues) *actionContext {
	return &actionContext{
		ConfigDir:           os.DirFS(values.ConfigDir),
		Flags:               values,
		Stdout:              os.Stdout,
		Stderr:              os.Stderr,
		WriteConfD:          getConfDWriter(values, ConfDSettings),
		WriteConfDOverrides: getConfDWriter(values, ConfDOverrides),
	}
}
