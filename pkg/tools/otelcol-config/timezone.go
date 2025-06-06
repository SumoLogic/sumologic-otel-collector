package main

import (
	"bytes"
	"fmt"
	"github.com/mikefarah/yq/v4/pkg/yqlib"
	"gopkg.in/yaml.v3"
)

func SetTimezoneAction(ctx *actionContext) error {
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

	encoder := yqlib.YamlFormat.EncoderFactory()
	decoder := yqlib.YamlFormat.DecoderFactory()
	eval := yqlib.NewStringEvaluator()

	if len(config) == 0 {
		// If the config is empty, we create a new document with the timezone setting.
		buf := new(bytes.Buffer)
		enc := yaml.NewEncoder(buf)
		enc.SetIndent(2)
		settings := map[string]any{
			"extensions": map[string]any{
				"sumologic": map[string]any{
					"time_zone": ctx.Flags.SetTimezone,
				},
			},
		}
		if err := enc.Encode(settings); err != nil {
			// shouldn't ever happen
			panic(err)
		}
		config = buf.Bytes()
	} else {
		expression := fmt.Sprintf(".extensions.sumologic.time_zone = %q", ctx.Flags.SetTimezone)
		result, err := eval.EvaluateAll(expression, string(config), encoder, decoder)
		if err != nil {
			return fmt.Errorf("couldn't write timezone: %s", err)
		}
		config = []byte(result)
	}

	_, err = writer(config)
	if err != nil {
		return fmt.Errorf("couldn't write timezone: %s", err)
	}

	return nil
}
