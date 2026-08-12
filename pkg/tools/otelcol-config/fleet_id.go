package main

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"

	"github.com/mikefarah/yq/v4/pkg/yqlib"
	"gopkg.in/yaml.v3"
)

var validHexPattern = regexp.MustCompile(`^[0-9a-fA-F]{16}$`)

func validateFleetID(id string) error {
	id = strings.TrimSpace(id)

	if id == "" {
		return fmt.Errorf("fleet ID cannot be empty")
	}

	if !validHexPattern.MatchString(id) {
		return fmt.Errorf("fleet ID must be exactly 16 hexadecimal characters (e.g. 000000000ABC1234)")
	}

	return nil
}

func SetFleetIDAction(ctx *actionContext) error {
	conf, err := ReadConfigDir(ctx.ConfigDir)
	if err != nil {
		return err
	}

	fleetID := strings.TrimSpace(ctx.Flags.SetFleetID)

	if err := validateFleetID(fleetID); err != nil {
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
					"fleet_id": fleetID,
				},
			},
		}
		if err := enc.Encode(settings); err != nil {
			panic(err)
		}
		doc = buf.Bytes()
	} else {
		expression := fmt.Sprintf(".extensions.sumologic.fleet_id = %q", fleetID)
		result, err := eval.EvaluateAll(expression, string(doc), encoder, decoder)
		if err != nil {
			return fmt.Errorf("couldn't set fleet ID: error evaluating yq expression: %s", err)
		}
		doc = []byte(result)
	}

	_, err = writer(doc)
	if err != nil {
		return fmt.Errorf("couldn't write updated config: %s", err)
	}

	return nil
}
