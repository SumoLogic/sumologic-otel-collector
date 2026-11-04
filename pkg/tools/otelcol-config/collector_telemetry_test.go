package main

import (
	"errors"
	"testing"
	"testing/fstest"
)

const remoteControlledConf = `{"extensions":{"opamp":{"enabled":true}}}`

func TestEnableTelemetryReportingAction(t *testing.T) {
	var called bool
	ctx := &actionContext{
		ConfigDir:              fstest.MapFS{},
		LinkCollectorTelemetry: func() error { called = true; return nil },
	}
	if err := EnableTelemetryReportingAction(ctx); err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Error("collector telemetry linker not called")
	}
	ctx.LinkCollectorTelemetry = func() error { return errors.New("error") }
	if err := EnableTelemetryReportingAction(ctx); err == nil {
		t.Fatal("expected non-nil error")
	}
}

func TestDisableTelemetryReportingAction(t *testing.T) {
	var called bool
	ctx := &actionContext{
		ConfigDir:                fstest.MapFS{},
		UnlinkCollectorTelemetry: func() error { called = true; return nil },
	}
	if err := DisableTelemetryReportingAction(ctx); err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Error("collector telemetry unlinker not called")
	}
	ctx.UnlinkCollectorTelemetry = func() error { return errors.New("error") }
	if err := DisableTelemetryReportingAction(ctx); err == nil {
		t.Fatal("expected non-nil error")
	}
}

func TestTelemetryReportingActionsRemoteControlled(t *testing.T) {
	ctx := &actionContext{
		ConfigDir: fstest.MapFS{
			SumologicRemoteDotYaml: &fstest.MapFile{Data: []byte(remoteControlledConf)},
		},
	}
	if err := EnableTelemetryReportingAction(ctx); err == nil {
		t.Fatal("expected non-nil error for enable")
	}
	if err := DisableTelemetryReportingAction(ctx); err == nil {
		t.Fatal("expected non-nil error for disable")
	}
}
