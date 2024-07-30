package main

import "errors"

// EnableHostmetricsAction links the available hostmetrics configuration to
// conf.d
func EnableHostmetricsAction(ctx *actionContext) error {
	conf, err := ReadConfigDir(ctx.ConfigDir)
	if err != nil {
		return err
	}
	if conf.SumologicRemote != nil {
		return errors.New("enable-hostmetrics not supported for remote-controlled collectors")
	}
	return ctx.LinkHostMetrics()
}

// DisableHostmetrics removes the link to the hostmetrics configuration.
func DisableHostmetricsAction(ctx *actionContext) error {
	conf, err := ReadConfigDir(ctx.ConfigDir)
	if err != nil {
		return err
	}
	if conf.SumologicRemote != nil {
		return errors.New("disable-hostmetrics not supported for remote-controlled collectors")
	}
	return ctx.UnlinkHostMetrics()
}
