package main

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"

	"github.com/mikefarah/yq/v4/pkg/yqlib"
	"gopkg.in/yaml.v3"
)

// For better performance, compile regex once at package level
var validNamePattern = regexp.MustCompile(`[^A-Za-z0-9_./=+\-@]`)

func isValidName(name string) bool {
	return !validNamePattern.MatchString(name)
}

func SetCollectorNameAction(ctx *actionContext) error {
	conf, err := ReadConfigDir(ctx.ConfigDir)
	collectorName := ctx.Flags.SetCollectorName
	collectorName = strings.TrimSpace(collectorName)

	if collectorName == "" {
		return fmt.Errorf("collector name cannot be empty")
	}
	// only Letters, numbers and _. / = + - @ are allowed
	if !isValidName(collectorName) {
		return fmt.Errorf("collector name contains invalid characters; only letters, numbers and _. / = + - @ are allowed")
	}

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
					"collector_name": ctx.Flags.SetCollectorName,
				},
			},
		}
		if err := enc.Encode(settings); err != nil {
			panic(err)
		}
		doc = buf.Bytes()
	} else {
		expression := fmt.Sprintf(".extensions.sumologic.collector_name = %q", ctx.Flags.SetCollectorName)
		result, err := eval.EvaluateAll(expression, string(doc), encoder, decoder)
		if err != nil {
			return fmt.Errorf("couldn't set collector name: error evaluating yq expression: %s", err)
		}

		doc = []byte(result)
	}
	_, err = writer(doc)
	if err != nil {
		return fmt.Errorf("couldn't write updated config: %s", err)
	}

	return nil
}
