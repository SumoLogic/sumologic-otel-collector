package main

import "errors"

// EnableEphemeralAction links the available ephemeral configuration to conf.d
func EnableEphemeralAction(ctx *actionContext) error {
	conf, err := ReadConfigDir(ctx.ConfigDir)
	if err != nil {
		return err
	}
	if conf.SumologicRemote != nil {
		return errors.New("enable-ephemeral not supported for remote-controlled collectors")
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
		return errors.New("disable-ephemeral not supported for remote-controlled collectors")
	}
	return ctx.UnlinkEphemeral()
}
