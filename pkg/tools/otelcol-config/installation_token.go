package main

import (
	"bytes"
	"fmt"

	"github.com/mikefarah/yq/v4/pkg/yqlib"
	"gopkg.in/yaml.v3"
)

func SetInstallationTokenAction(ctx *actionContext) error {
	if ctx.SystemdEnabled {
		return setInstallationTokenSystemd(ctx)
	}

	conf, err := ReadConfigDir(ctx.ConfigDir)
	if err != nil {
		return err
	}

	var writer func([]byte) (int, error)
	var doc []byte

	switch {
	case conf.SumologicRemote != nil || ctx.Flags.EnableRemoteControl:
		writer = ctx.WriteSumologicRemote
		doc = conf.SumologicRemote
	case ctx.Flags.Override:
		writer = ctx.WriteConfDOverrides
		doc = conf.ConfD[ConfDOverrides]
	default:
		writer = ctx.WriteConfD
		doc = conf.ConfD[ConfDSettings]
	}

	encoder := yqlib.YamlFormat.EncoderFactory()
	decoder := yqlib.YamlFormat.DecoderFactory()
	eval := yqlib.NewStringEvaluator()

	if len(doc) == 0 {
		buf := new(bytes.Buffer)
		enc := yaml.NewEncoder(buf)
		enc.SetIndent(2)
		settings := map[string]any{
			"extensions": map[string]any{
				"sumologic": map[string]any{
					"installation_token": ctx.Flags.InstallationToken,
				},
			},
		}
		if err := enc.Encode(settings); err != nil {
			// shouldn't ever happen
			panic(err)
		}
		doc = buf.Bytes()
	} else {
		expression := fmt.Sprintf(".extensions.sumologic.installation_token = %q", ctx.Flags.InstallationToken)
		result, err := eval.EvaluateAll(expression, string(doc), encoder, decoder)
		if err != nil {
			return fmt.Errorf("couldn't write installation token: %s", err)
		}
		doc = []byte(result)
	}

	_, err = writer(doc)
	if err != nil {
		return fmt.Errorf("couldn't write installation token: %s", err)
	}

	return nil
}

func setInstallationTokenSystemd(ctx *actionContext) error {
	tokenDoc := fmt.Sprintf("SUMOLOGIC_INSTALLATION_TOKEN=%s\n", ctx.Flags.InstallationToken)
	_, err := ctx.WriteInstallationTokenEnv([]byte(tokenDoc))
	if err != nil {
		return fmt.Errorf("couldn't write token.env: %s", err)
	}
	return nil
}
