package main

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/mikefarah/yq/v4/pkg/yqlib"
)

func AddTagAction(ctx *actionContext) error {
	conf, err := ReadConfigDir(ctx.ConfigDir)
	if err != nil {
		return err
	}

	if conf.SumologicRemote != nil {
		return errors.New("add-tag not supported for remote-controlled collectors")
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

	const (
		keyFmt      = ".extensions.sumologic.collector_fields.%s = %s"
		quoteKeyFmt = ".extensions.sumologic.collector_fields.%s = %q"
	)

	for tagName, tag := range ctx.Flags.AddTags {
		expression := fmt.Sprintf(keyFmt, tagName, tag)
		if len(doc) == 0 {
			// --null-input
			buf := new(bytes.Buffer)
			writer := yqlib.NewSinglePrinterWriter(buf)
			printer := yqlib.NewPrinter(encoder, writer)
			seval := yqlib.NewStreamEvaluator()
			err := seval.EvaluateNew(expression, printer)
			if err != nil {
				// perhaps the value needs to be quoted, try again with the value quoted.
				expression = fmt.Sprintf(quoteKeyFmt, tagName, tag)
				qerr := seval.EvaluateNew(expression, printer)
				if qerr != nil {
					return fmt.Errorf("can't add tag %s: error evaluating yq expression: %s", tagName, err)
				}
			}
			doc = buf.Bytes()
			continue
		}
		result, err := eval.EvaluateAll(expression, string(doc), encoder, decoder)
		if err != nil {
			// perhaps the value needs to be quoted, try again with the value quoted.
			expression = fmt.Sprintf(quoteKeyFmt, tagName, tag)
			var qerr error
			result, qerr = eval.EvaluateAll(expression, string(doc), encoder, decoder)
			if qerr != nil {
				return fmt.Errorf("can't add tag %s: error evaluating yq expression: %s", tagName, err)
			}
		}
		doc = []byte(result)
	}

	if _, err := writeDoc(doc); err != nil {
		return fmt.Errorf("error writing %s: %s", docName, err)
	}

	return nil

}
