package main

import (
	"bytes"
	"fmt"

	"github.com/mikefarah/yq/v4/pkg/yqlib"
)

// EnableClobberAction links the available clobber configuration to conf.d
func EnableClobberAction(ctx *actionContext) error {
	conf, err := ReadConfigDir(ctx.ConfigDir)
	if err != nil {
		return err
	}
	if conf.SumologicRemote != nil {
		return writeClobberRemote(ctx, conf.SumologicRemote, ".extensions.sumologic.clobber = true")
	}
	return ctx.LinkClobber()
}

// DisableClobberAction removes the link to the clobber configuration.
func DisableClobberAction(ctx *actionContext) error {
	conf, err := ReadConfigDir(ctx.ConfigDir)
	if err != nil {
		return err
	}
	if conf.SumologicRemote != nil {
		return writeClobberRemote(ctx, conf.SumologicRemote, ".extensions.sumologic.clobber = false")
	}
	return ctx.UnlinkClobber()
}

func writeClobberRemote(ctx *actionContext, doc []byte, expr string) error {
	encoder := yqlib.YamlFormat.EncoderFactory()
	decoder := yqlib.YamlFormat.DecoderFactory()
	eval := yqlib.NewStringEvaluator()
	if len(doc) == 0 {
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
			return fmt.Errorf("error updating sumologic-remote.yaml: %s", err)
		}
		doc = []byte(result)
	}
	_, err := ctx.WriteSumologicRemote(doc)
	return err
}
