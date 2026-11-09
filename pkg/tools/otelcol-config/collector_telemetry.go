package main

import "errors"

// EnableTelemetryReportingAction links the available collector telemetry
// configuration to conf.d
func EnableTelemetryReportingAction(ctx *actionContext) error {
	conf, err := ReadConfigDir(ctx.ConfigDir)
	if err != nil {
		return err
	}
	if conf.SumologicRemote != nil {
		return errors.New("enable-telemetry-reporting not supported for remote-controlled collectors")
	}
	return ctx.LinkCollectorTelemetry()
}

// DisableTelemetryReportingAction removes the link to the collector telemetry
// configuration.
func DisableTelemetryReportingAction(ctx *actionContext) error {
	conf, err := ReadConfigDir(ctx.ConfigDir)
	if err != nil {
		return err
	}
	if conf.SumologicRemote != nil {
		return errors.New("disable-telemetry-reporting not supported for remote-controlled collectors")
	}
	return ctx.UnlinkCollectorTelemetry()
}
