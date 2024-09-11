//go:build unix

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

// use chown to set the owner of sumologic-remote.yaml to otelcol-sumo
// or whatever the user should be, based on the ownership of sumologic.yaml
// or its parent directory
func setSumologicRemoteOwner(values *flagValues) error {
	baseConfigPath := filepath.Join(values.ConfigDir, SumologicDotYaml)
	docPath := filepath.Join(values.ConfigDir, SumologicRemoteDotYaml)

	// check who owns the base configuration file
	stat, err := os.Stat(baseConfigPath)
	if err != nil {
		// maybe it doesn't exist, stat the parent dir instead
		stat, err = os.Stat(values.ConfigDir)
		if err != nil {
			// something is seriously wrong
			return fmt.Errorf("error reading config dir: %s", err)
		}
	}

	sys, ok := stat.Sys().(*syscall.Stat_t)
	if !ok {
		// the platform does not has the expected sys somehow,
		// so just bail out with no error
		return nil
	}

	if int(sys.Uid) == syscall.Getuid() {
		// we're already that user
		return nil
	}

	// set the owner to be consistent with the other configuration
	if err := os.Chown(docPath, int(sys.Uid), int(sys.Gid)); err != nil {
		if err.(*os.PathError).Err == syscall.EPERM {
			// we don't have permission to chown, skip it
			return nil
		}
		return err
	}

	return nil
}