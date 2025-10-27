package main

import (
	"bytes"
	"fmt"

	"github.com/mikefarah/yq/v4/pkg/yqlib"
	"gopkg.in/yaml.v3"
)

func ClobberAction(ctx *actionContext) error {
	conf, err := ReadConfigDir(ctx.ConfigDir)
	if err != nil {
		return err
	}
	var writer func([]byte) (int, error)
	var config []byte

	switch {
	case conf.SumologicRemote != nil || ctx.Flags.EnableRemoteControl:
		writer = ctx.WriteSumologicRemote
		config = conf.SumologicRemote
	case ctx.Flags.Override:
		writer = ctx.WriteConfDOverrides
		config = conf.ConfD[ConfDOverrides]
	default:
		writer = ctx.WriteConfD
		config = conf.ConfD[ConfDSettings]
	}

	return writeYAML(ctx, config, writer)
}

func writeYAML(ctx *actionContext, config []byte, writer func([]byte) (int, error)) error {
	encoder := yqlib.YamlFormat.EncoderFactory()
	decoder := yqlib.YamlFormat.DecoderFactory()
	eval := yqlib.NewStringEvaluator()

	if len(config) == 0 {
		buff := new(bytes.Buffer)
		enc := yaml.NewEncoder(buff)
		enc.SetIndent(2)
		settings := map[string]any{
			"extensions": map[string]any{
				"sumologic": map[string]any{
					"clobber": ctx.Flags.Clobber,
				},
			},
		}
		if err := enc.Encode(settings); err != nil {
			panic(err)
		}

		config = buff.Bytes()
	} else {
		expression := fmt.Sprintf(".extensions.sumologic.clobber = %t", ctx.Flags.Clobber)
		result, err := eval.EvaluateAll(expression, string(config), encoder, decoder)
		if err != nil {
			return fmt.Errorf("evaluate: %w", err)
		}
		config = []byte(result)

	}
	_, err := writer(config)
	if err != nil {
		return fmt.Errorf("Error encountered while setting clobber: %w", err)
	}
	return nil
}
