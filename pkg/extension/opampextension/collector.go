//go:build !windows

package opampextension

import (
	"os"
	"syscall"
)

func reloadCollectorConfig() error {
	p, err := os.FindProcess(os.Getpid())

	if err != nil {
		return err
	}

	return p.Signal(syscall.SIGHUP)
}
