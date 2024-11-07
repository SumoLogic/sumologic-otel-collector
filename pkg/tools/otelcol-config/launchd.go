package main

import (
	"fmt"
	"io/fs"

	"howett.net/plist"
)

const (
	launchdConfigPlist = "com.sumologic.otelcol-sumo.plist"
)

type launchdConfig struct {
	Format int
	Root   launchdRoot
}

type launchdRoot struct {
	Label                string            `plist:"Label"`
	ProgramArguments     []string          `plist:"ProgramArguments"`
	EnvironmentVariables map[string]string `plist:"EnvironmentVariables"`
	UserName             string            `plist:"UserName"`
	GroupName            string            `plist:"GroupName"`
	RunAtLoad            bool              `plist:"RunAtLoad"`
	KeepAlive            bool              `plist:"KeepAlive"`
	StandardOutPath      string            `plist:"StandardOutPath"`
	StandardErrorPath    string            `plist:"StandardErrorPath"`
}

// ReadLaunchdConfig reads and unmarshals the launchd config from
// /Library/LaunchDaemons/com.sumologic.otelcol-sumo.plist. It produces a
// launchdConfig that contains the data read from the launchd config.
func ReadLaunchdConfig(root fs.FS) (launchdConfig, error) {
	conf := launchdConfig{}

	bytes, err := fs.ReadFile(root, launchdConfigPlist)
	if err != nil {
		return conf, fmt.Errorf("error reading launchd config: %s", err)
	}

	confRoot := launchdRoot{}
	format, err := plist.Unmarshal(bytes, &confRoot)
	if err != nil {
		return conf, fmt.Errorf("error unmarshaling launchd config: %s", err)
	}

	conf.Format = format
	conf.Root = confRoot

	return conf, nil
}
