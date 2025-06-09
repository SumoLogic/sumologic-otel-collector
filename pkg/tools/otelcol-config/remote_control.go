package main

import (
	"bytes"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	DefaultSumoLogicOpampEndpoint       = "wss://opamp-collectors.sumologic.com/v1/opamp"
	DefaultConfigurationDirectory       = "/etc/otelcol-sumo"
	DefaultRemoteConfigurationDirectory = "opamp.d"
)

func EnableRemoteControlAction(ctx *actionContext) error {
	conf, err := ReadConfigDir(ctx.ConfigDir)
	if err != nil {
		return err
	}

	if conf.SumologicRemote != nil {
		// remote config already enabled
		return nil
	}

	return makeNewSumologicRemoteYAML(ctx, conf)
}

func DisableRemoteControlAction(ctx *actionContext) error {
	conf, err := ReadConfigDir(ctx.ConfigDir)
	if err != nil {
		return err
	}

	if conf.SumologicRemote == nil {
		return nil
	}

	_, err = ctx.WriteSumologicRemote(nil)

	return err
}

func makeNewSumologicRemoteYAML(ctx *actionContext, conf ConfDir) error {
	confBase := DefaultConfigurationDirectory
	if ctx.Flags.ConfigDir != "" {
		confBase = ctx.Flags.ConfigDir
	}

	remoteConfigDir := filepath.Join(confBase, DefaultRemoteConfigurationDirectory)

	var sumoRemoteConfig = map[string]any{
		"extensions": map[string]any{
			"file_storage": map[string]any{
				"compaction": map[string]any{
					"directory":  "/var/lib/otelcol-sumo/file_storage",
					"on_rebound": true,
				},
				"directory": "/var/lib/otelcol-sumo/file_storage",
			},
			"health_check": map[string]any{
				"endpoint": "localhost:13133",
			},
			"opamp": map[string]any{
				"remote_configuration_directory": remoteConfigDir,
				"endpoint":                       DefaultSumoLogicOpampEndpoint,
			},
			"sumologic": map[string]any{
				"installation_token":              "${SUMOLOGIC_INSTALLATION_TOKEN}",
				"collector_credentials_directory": "/var/lib/otelcol-sumo/credentials",
				"time_zone":                       "UTC",
			},
		},
		"receivers": map[string]any{
			"nop": map[string]any{},
		},
		"exporters": map[string]any{
			"nop": map[string]any{},
		},
		"service": map[string]any{
			"extensions": []string{
				"sumologic",
				"health_check",
				"file_storage",
				"opamp",
			},
			"pipelines": map[string]any{
				"metrics/default": map[string]any{
					"receivers": []string{"nop"},
					"exporters": []string{"nop"},
				},
				"logs/default": map[string]any{
					"receivers": []string{"nop"},
					"exporters": []string{"nop"},
				},
				"traces/default": map[string]any{
					"receivers": []string{"nop"},
					"exporters": []string{"nop"},
				},
			},
		},
	}

	buf := new(bytes.Buffer)
	enc := yaml.NewEncoder(buf)
	enc.SetIndent(2)
	if err := enc.Encode(sumoRemoteConfig); err != nil {
		// this should never happen even under abnormal circumstances
		panic(err)
	}

	_, err := ctx.WriteSumologicRemote(buf.Bytes())

	return err
}
