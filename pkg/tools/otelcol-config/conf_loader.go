package main

import (
	"errors"
	"fmt"
	"io/fs"
)

const (
	SumologicDotYaml       = "sumologic.yaml"
	SumologicRemoteDotYaml = "sumologic-remote.yaml"
	ConfDotD               = "conf.d"
	ConfDotDAvailable      = "conf.d-available"
	ConfDSettings          = "00-otelcol-config-settings.yaml"
	ConfDOverrides         = "99-otelcol-config-overrides.yaml"
)

type ConfDir struct {
	// Sumologic is the contents of sumologic.yaml.
	Sumologic []byte

	// SumologicRemote is the contents of sumologic-remote.yaml.
	SumologicRemote []byte

	// ConfD is a mapping of file name to contents of the conf.d directory.
	ConfD map[string][]byte

	// ConfDAvailable is a mapping of file name to contents of the
	// conf.d-available directory.
	ConfDAvailable map[string][]byte
}

// ReadConfigDir reads the files in /etc/otelcol-sumo according to an expected
// layout. It produces a ConfDir that contains the data read from the files.
func ReadConfigDir(root fs.FS) (conf ConfDir, err error) {
	const errMsg = "error reading otelcol-sumo config dir: %s"

	conf.Sumologic, err = fs.ReadFile(root, SumologicDotYaml)
	if err != nil {
		// sumologic.yaml is not strictly required
		if !errors.Is(err, fs.ErrNotExist) {
			return conf, fmt.Errorf(errMsg, err)
		}
	}

	conf.SumologicRemote, err = fs.ReadFile(root, SumologicRemoteDotYaml)
	if err != nil {
		// sumologic-remote.yaml is not strictly required
		if !errors.Is(err, fs.ErrNotExist) {
			return conf, fmt.Errorf(errMsg, err)
		}
	}

	conf.ConfD, err = getDir(root, ConfDotD)
	if err != nil {
		// conf.d is not strictly required
		if !errors.Is(err, fs.ErrNotExist) {
			return conf, fmt.Errorf(errMsg, err)
		}
	}

	conf.ConfDAvailable, err = getDir(root, ConfDotDAvailable)
	if err != nil {
		// conf.d-available is not strictly required
		if !errors.Is(err, fs.ErrNotExist) {
			return conf, fmt.Errorf(errMsg, err)
		}
	}

	return conf, nil
}

func getDir(root fs.FS, dirName string) (result map[string][]byte, err error) {
	result = make(map[string][]byte)
	dirFS, err := fs.Sub(root, dirName)
	if err != nil {
		return nil, err
	}
	entries, err := fs.ReadDir(dirFS, ".")
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		if entry.Type() == fs.ModeDir {
			continue
		}
		contents, err := fs.ReadFile(dirFS, entry.Name())
		if err != nil {
			if err == fs.ErrNotExist {
				continue
			}
			if _, ok := err.(*fs.PathError); ok {
				continue
			}
			return nil, err
		}
		result[entry.Name()] = contents
	}
	return result, nil
}
