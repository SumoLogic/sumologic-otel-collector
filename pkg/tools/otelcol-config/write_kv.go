package main

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/mikefarah/yq/v4/pkg/yqlib"
)

// WriteKVAction applies a yq expression to the user configuration in conf.d.
// Any arbitrary write can occur, including deletions.
//
// WriteKVAction applies updates to one of two files. Either:
// conf.d/00-otelcol-config-settings.yaml OR
// conf.d/99-otelcol-config-overrides.yaml,
// depending on if the override flag is set.
func WriteKVAction(ctx *actionContext) error {
	conf, err := ReadConfigDir(ctx.ConfigDir)
	if err != nil {
		return err
	}

	if conf.SumologicRemote != nil {
		return errors.New("write-kv not supported for remote-controlled collectors")
	}

	docName := ConfDSettings
	writeDoc := ctx.WriteConfD
	if ctx.Flags.Override {
		docName = ConfDOverrides
		writeDoc = ctx.WriteConfDOverrides
	}

	encoder := yqlib.YamlFormat.EncoderFactory()
	decoder := yqlib.YamlFormat.DecoderFactory()
	eval := yqlib.NewStringEvaluator()
	doc := conf.ConfD[docName]

	for _, expression := range ctx.Flags.WriteKV {
		if len(doc) == 0 {
			// --null-input
			buf := new(bytes.Buffer)
			writer := yqlib.NewSinglePrinterWriter(buf)
			printer := yqlib.NewPrinter(encoder, writer)
			err := yqlib.NewStreamEvaluator().EvaluateNew(expression, printer)
			if err != nil {
				return fmt.Errorf("error writing %s: %s", docName, err)
			}
			doc = buf.Bytes()
			continue
		}
		result, err := eval.EvaluateAll(expression, string(doc), encoder, decoder)
		if err != nil {
			return fmt.Errorf("error evaluating yq expression: %s", err)
		}
		doc = []byte(result)
	}

	if _, err := writeDoc(doc); err != nil {
		return fmt.Errorf("error writing %s: %s", docName, err)
	}

	return nil
}
