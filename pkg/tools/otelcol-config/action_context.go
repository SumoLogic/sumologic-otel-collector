package main

import (
	"io"
	"io/fs"
)

// actionContext provides an abstraction of I/O for all actions. The reason
// for this to exist is so I/O operations can be mocked in test.
type actionContext struct {
	ConfigDir            fs.FS
	Flags                *flagValues
	Stdout               io.Writer
	Stderr               io.Writer
	WriteConfD           func([]byte) (int, error)
	WriteConfDOverrides  func([]byte) (int, error)
	WriteSumologicRemote func([]byte) (int, error)
	LinkHostMetrics      func() error
	UnlinkHostMetrics    func() error
}
