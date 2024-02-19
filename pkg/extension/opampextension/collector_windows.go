//go:build windows

package opampextension

import (
	"os"
)

func reloadCollectorConfig() error {
	// On Windows we want to kill the process and let the service manager restart it.
	os.Exit(-9)
	return nil
}
