package main

import (
	"io"
	"io/fs"
)

// actionContext provides an abstraction of I/O for all actions
type actionContext struct {
	ConfigDir            fs.FS
	Flags                *flagValues
	Stdout               io.Writer
	Stderr               io.Writer
	WriteConfD           func([]byte) (int, error)
	WriteConfDOverrides  func([]byte) (int, error)
	WriteSumologicRemote func([]byte) (int, error)
}
