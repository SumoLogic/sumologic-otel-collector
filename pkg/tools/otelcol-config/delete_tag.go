package main

import (
	"fmt"
	"sort"

	"github.com/mikefarah/yq/v4/pkg/yqlib"
)

// DeleteTagAction deletes a collector tag from conf.d.
// If the --override flag is present, it will delete the tag from both the
// overrides file and the settings file. If the --override flag is present and
// the tag exists in a user-controlled file, an error will be returned.
func DeleteTagAction(ctx *actionContext) error {
	conf, err := ReadConfigDir(ctx.ConfigDir)
	if err != nil {
		return err
	}

	if conf.SumologicRemote != nil {
		return deleteTag(ctx, conf.SumologicRemote, ctx.WriteSumologicRemote)
	}

	if ctx.Flags.Override {
		return deleteTagOverride(ctx, conf)
	}

	return deleteTag(ctx, conf.ConfD[ConfDSettings], ctx.WriteConfD)
}

func deleteTagOverride(ctx *actionContext, conf ConfDir) error {
	readOrder := sort.StringSlice{}
	for key := range conf.ConfD {
		readOrder = append(readOrder, key)
	}
	sort.Sort(sort.Reverse(readOrder))

	// Check if there are matching tags in user-controlled files. If there are,
	// refuse to continue on the grounds that we don't support edits to user
	// created files.
	eval := yqlib.NewStringEvaluator()
	encoder := yqlib.YamlFormat.EncoderFactory()
	decoder := yqlib.YamlFormat.DecoderFactory()
	for _, tagName := range ctx.Flags.DeleteTags {
		key := fmt.Sprintf(".extensions.sumologic.collector_fields.%q", tagName)
		for _, confKey := range readOrder {
			doc := string(conf.ConfD[confKey])
			result, err := eval.Evaluate(key, doc, encoder, decoder)
			if err != nil {
				return fmt.Errorf("error evaluating yq expression: %s", err)
			}
			if len(result) > 0 && result != nullResult {
				if confKey != ConfDSettings && confKey != ConfDOverrides {
					return fmt.Errorf("can't delete tag %s: user setting in %s cannot be overridden", tagName, confKey)
				}
			}
		}
	}

	if err := deleteTag(ctx, conf.ConfD[ConfDOverrides], ctx.WriteConfDOverrides); err != nil {
		return err
	}

	return deleteTag(ctx, conf.ConfD[ConfDSettings], ctx.WriteConfD)
}

func deleteTag(ctx *actionContext, doc []byte, writeDoc func([]byte) (int, error)) error {
	encoder := yqlib.YamlFormat.EncoderFactory()
	decoder := yqlib.YamlFormat.DecoderFactory()
	eval := yqlib.NewStringEvaluator()

	if len(doc) == 0 {
		// tag does not exist, nor any other config for that matter
		return nil
	}

	const keyFmt = "del(.extensions.sumologic.collector_fields.%q)"

	for _, tagName := range ctx.Flags.DeleteTags {
		expression := fmt.Sprintf(keyFmt, tagName)

		result, err := eval.EvaluateAll(expression, string(doc), encoder, decoder)
		if err != nil {
			return fmt.Errorf("can't delete tag %s: %s", tagName, err)
		}

		doc = []byte(result)
	}

	if _, err := writeDoc(doc); err != nil {
		return fmt.Errorf("error deleting tag: %s", err)
	}

	return nil
}
