package main

import (
	"bytes"
	"errors"
	"fmt"
	"net/url"

	"github.com/mikefarah/yq/v4/pkg/yqlib"
	"gopkg.in/yaml.v3"
)

func SetAPIURLAction(ctx *actionContext) error {
	u, err := url.Parse(ctx.Flags.SetAPIURL)
	if err != nil {
		return fmt.Errorf("couldn't set api base url: invalid url: %s", err)
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return errors.New("couldn't set api base url: url must have http or https scheme")
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
					"api_base_url": ctx.Flags.SetAPIURL,
				},
			},
		}
		if err := enc.Encode(settings); err != nil {
			// shouldn't ever happen
			panic(err)
		}
		doc = buf.Bytes()
	} else {
		expression := fmt.Sprintf(".extensions.sumologic.api_base_url = %q", ctx.Flags.SetAPIURL)
		result, err := eval.EvaluateAll(expression, string(doc), encoder, decoder)
		if err != nil {
			return fmt.Errorf("couldn't write api base url: %s", err)
		}
		doc = []byte(result)
	}

	_, err = writer(doc)
	if err != nil {
		return fmt.Errorf("couldn't write api base url: %s", err)
	}

	return nil
}
