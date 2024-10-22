package main

import (
	"bytes"
	"fmt"

	"github.com/mikefarah/yq/v4/pkg/yqlib"
)

// EnableEphemeralAction links the available ephemeral configuration to conf.d
func EnableEphemeralAction(ctx *actionContext) error {
	conf, err := ReadConfigDir(ctx.ConfigDir)
	if err != nil {
		return err
	}
	if conf.SumologicRemote != nil {
		return writeEphemeralRemote(ctx, conf.SumologicRemote, ".extensions.sumologic.ephemeral = true")
	}
	return ctx.LinkEphemeral()
}

// DisableEphemeralAction removes the link to the ephemeral configuration.
func DisableEphemeralAction(ctx *actionContext) error {
	conf, err := ReadConfigDir(ctx.ConfigDir)
	if err != nil {
		return err
	}
	if conf.SumologicRemote != nil {
		return writeEphemeralRemote(ctx, conf.SumologicRemote, ".extensions.sumologic.ephemeral = false")
	}
	return ctx.UnlinkEphemeral()
}

func writeEphemeralRemote(ctx *actionContext, doc []byte, expr string) error {
	encoder := yqlib.YamlFormat.EncoderFactory()
	decoder := yqlib.YamlFormat.DecoderFactory()
	eval := yqlib.NewStringEvaluator()
	if len(doc) == 0 {
		// --null-input
		buf := new(bytes.Buffer)
		writer := yqlib.NewSinglePrinterWriter(buf)
		printer := yqlib.NewPrinter(encoder, writer)
		err := yqlib.NewStreamEvaluator().EvaluateNew(expr, printer)
		if err != nil {
			return fmt.Errorf("error writing sumologic-remote.yaml: %s", err)
		}
		doc = buf.Bytes()
	} else {
		result, err := eval.EvaluateAll(expr, string(doc), encoder, decoder)
		if err != nil {
			return fmt.Errorf("error evaluating yq expression: %s", err)
		}
		doc = []byte(result)
	}
	if _, err := ctx.WriteSumologicRemote(doc); err != nil {
		return fmt.Errorf("error writing sumologic-remote.yaml: %s", err)
	}
	return nil
}
