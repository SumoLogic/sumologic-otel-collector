//go:build darwin

package sumologic_scripts_tests

import (
	"os"

	"howett.net/plist"
)

type launchdConfig struct {
	EnvironmentVariables environmentVariables `plist:"EnvironmentVariables"`
	GroupName            string               `plist:"GroupName"`
	KeepAlive            bool                 `plist:"KeepAlive"`
	Label                string               `plist:"Label"`
	ProgramArguments     []string             `plist:"ProgramArguments"`
	RunAtLoad            bool                 `plist:"RunAtLoad"`
	StandardErrorPath    string               `plist:"StandardErrorPath"`
	StandardOutPath      string               `plist:"StandardOutPath"`
	UserName             string               `plist:"UserName"`
}

type environmentVariables struct {
	InstallationToken string `plist:"SUMOLOGIC_INSTALLATION_TOKEN"`
}

func NewLaunchdConfig() launchdConfig {
	return launchdConfig{
		GroupName: "_otelcol-sumo",
		KeepAlive: true,
		Label:     "otelcol-sumo",
		ProgramArguments: []string{
			"/usr/local/bin/otelcol-sumo",
			"--config",
			"/etc/otelcol-sumo/sumologic.yaml",
			"--config",
			"glob:/etc/otelcol-sumo/conf.d/*.yaml",
		},
		RunAtLoad:         true,
		StandardErrorPath: "/var/log/otelcol-sumo/otelcol-sumo.log",
		StandardOutPath:   "/var/log/otelcol-sumo/otelcol-sumo.log",
		UserName:          "_otelcol-sumo",
	}
}

func getLaunchdConfig(path string) (launchdConfig, error) {
	var conf launchdConfig

	plistFile, err := os.ReadFile(path)
	if err != nil {
		return launchdConfig{}, err
	}

	if _, err := plist.Unmarshal(plistFile, &conf); err != nil {
		return launchdConfig{}, err
	}

	return conf, nil
}

func saveLaunchdConfig(path string, conf launchdConfig) error {
	out, err := plist.Marshal(conf, plist.XMLFormat)
	if err != nil {
		return err
	}

	return os.WriteFile(path, out, os.ModePerm)
}
