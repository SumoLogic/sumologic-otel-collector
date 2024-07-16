package main

import (
	"errors"
	"fmt"
	"sort"

	"github.com/mikefarah/yq/v4/pkg/yqlib"
)

const nullResult = "null\n"

// ReadKVAction reads values from conf.d. Since conf.d can comprise multiple
// files, it looks in descending order, and returns the first match it finds.
// This matches the priority order of otelcol-sumo's configuration loader.
// Or at least, it should.
func ReadKVAction(ctx *actionContext) error {
	conf, err := ReadConfigDir(ctx.ConfigDir)
	if err != nil {
		return err
	}

	if conf.SumologicRemote != nil {
		return errors.New("read-kv not supported for remote-controlled collectors")
	}

	readOrder := sort.StringSlice{}
	for key := range conf.ConfD {
		readOrder = append(readOrder, key)
	}
	sort.Sort(sort.Reverse(readOrder))
	encoder := yqlib.YamlFormat.EncoderFactory()
	decoder := yqlib.YamlFormat.DecoderFactory()
	eval := yqlib.NewStringEvaluator()
	for _, key := range ctx.Flags.ReadKV {
		matched := false
		for _, confKey := range readOrder {
			doc := string(conf.ConfD[confKey])
			result, err := eval.Evaluate(key, doc, encoder, decoder)
			if err != nil {
				return fmt.Errorf("error evaluating yq expression: %s", err)
			}
			if len(result) > 0 && result != nullResult {
				// nb: the yaml evaluator result includes a newline,
				// no need to add an additional one.
				matched = true
				_, _ = ctx.Stdout.Write([]byte(result))
				break
			}
		}
		if !matched {
			_, _ = ctx.Stdout.Write([]byte("null\n"))
		}
	}
	return nil
}
