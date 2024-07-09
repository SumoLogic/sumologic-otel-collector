package main

import (
	"io"
	"io/fs"
)

// actionContext provides an abstraction of I/O for all actions
type actionContext struct {
	ConfigDir           fs.FS
	Flags               *flagValues
	Stdout              io.Writer
	Stderr              io.Writer
	WriteConfD          func([]byte) error
	WriteConfDOverrides func([]byte) error
}
