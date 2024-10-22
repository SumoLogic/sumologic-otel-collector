package main

import (
	"io"
	"io/fs"
)

// actionContext provides an abstraction of I/O for all actions. The reason
// for this to exist is so I/O operations can be mocked in test.
//
// actionContext is not like context.Context, and is not for cancellation or
// for carrying values across API boundaries. It is for abstracting I/O.
type actionContext struct {
	ConfigDir                 fs.FS
	Flags                     *flagValues
	Stdout                    io.Writer
	Stderr                    io.Writer
	WriteConfD                func([]byte) (int, error)
	WriteConfDOverrides       func([]byte) (int, error)
	WriteSumologicRemote      func([]byte) (int, error)
	LinkHostMetrics           func() error
	UnlinkHostMetrics         func() error
	LinkEphemeral             func() error
	UnlinkEphemeral           func() error
	SystemdEnabled            bool
	WriteInstallationTokenEnv func([]byte) (int, error)
}
