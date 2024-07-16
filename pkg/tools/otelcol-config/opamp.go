package main

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/mikefarah/yq/v4/pkg/yqlib"
)

var errRemoteControlNotEnabled = errors.New("remote control not enabled")

func SetOpAmpEndpointAction(ctx *actionContext) error {
	endpoint := ctx.Flags.SetOpAmpEndpoint
	if err := validateEndpoint(endpoint); err != nil {
		return err
	}
	conf, err := ReadConfigDir(ctx.ConfigDir)
	if err != nil {
		return err
	}

	writer := ctx.WriteSumologicRemote
	doc := conf.SumologicRemote

	if len(doc) == 0 {
		return fmt.Errorf("cannot set opamp endpoint: %s", errRemoteControlNotEnabled)
	}

	encoder := yqlib.YamlFormat.EncoderFactory()
	decoder := yqlib.YamlFormat.DecoderFactory()
	eval := yqlib.NewStringEvaluator()

	expression := fmt.Sprintf(".extensions.opamp.endpoint = %q", endpoint)
	result, err := eval.EvaluateAll(expression, string(doc), encoder, decoder)
	if err != nil {
		return fmt.Errorf("couldn't write installation token: %s", err)
	}
	doc = []byte(result)

	_, err = writer(doc)
	if err != nil {
		return fmt.Errorf("couldn't write opamp endpoint: %s", err)
	}

	return nil
}

func validateEndpoint(endpoint string) error {
	u, err := url.Parse(endpoint)
	if err != nil {
		return fmt.Errorf("invalid endpoint: %s", err)
	}
	if u.Scheme != "ws" && u.Scheme != "wss" {
		return fmt.Errorf("invalid endpoint: invalid URL scheme %q: want [ws, wss]", u.Scheme)
	}
	return nil
}
