package main

import (
	"bytes"
	"fmt"
	"path/filepath"

	"github.com/mikefarah/yq/v4/pkg/yqlib"
)

const (
	extensionsOpampEndpoint                = ".extensions.opamp.endpoint"
	extensionsRemoteConfigurationDirectory = ".extensions.opamp.remote_configuration_directory"
	extensionsOpampEnabled                 = ".extensions.opamp.enabled"
)

const (
	DefaultSumoLogicOpampEndpoint       = "wss://opamp-events.sumologic.com/v1/opamp"
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

	encoder := yqlib.YamlFormat.EncoderFactory()
	remoteConfigDir := filepath.Join(confBase, DefaultRemoteConfigurationDirectory)
	buf := new(bytes.Buffer)
	writer := yqlib.NewSinglePrinterWriter(buf)
	printer := yqlib.NewPrinter(encoder, writer)
	expression := fmt.Sprintf(
		"%s = %q, %s = %q, %s = %t",
		extensionsOpampEndpoint, DefaultSumoLogicOpampEndpoint,
		extensionsRemoteConfigurationDirectory, remoteConfigDir,
		extensionsOpampEnabled, true,
	)
	err := yqlib.NewStreamEvaluator().EvaluateNew(expression, printer)
	if err != nil {
		err = fmt.Errorf("developer error: %s", err)
		return err
	}

	_, err = ctx.WriteSumologicRemote(buf.Bytes())

	return err
}
