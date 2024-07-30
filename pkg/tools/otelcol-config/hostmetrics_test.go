package main

import (
	"errors"
	"testing"
	"testing/fstest"
)

func TestEnableHostmetricsAction(t *testing.T) {
	var called bool
	linkFunc := func() error {
		called = true
		return nil
	}
	errLinkFunc := func() error {
		return errors.New("error")
	}
	ctx := &actionContext{
		ConfigDir:       fstest.MapFS{},
		LinkHostMetrics: linkFunc,
	}
	if err := EnableHostmetricsAction(ctx); err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Error("hostmetrics linker not called")
	}
	ctx.LinkHostMetrics = errLinkFunc
	if err := EnableHostmetricsAction(ctx); err == nil {
		t.Fatal("expected non-nil error")
	}
}

func TestEnableHostmetricsActionRemoteControlled(t *testing.T) {
	ctx := &actionContext{
		ConfigDir: fstest.MapFS{
			SumologicRemoteDotYaml: &fstest.MapFile{
				Data: []byte(`{"extensions":{"opamp":{"enabled":true}}}`),
			},
		},
	}
	if err := EnableHostmetricsAction(ctx); err == nil {
		t.Fatal("expected non-nil error")
	}
}

func TestDisableHostmetricsActionRemoteControlled(t *testing.T) {
	ctx := &actionContext{
		ConfigDir: fstest.MapFS{
			SumologicRemoteDotYaml: &fstest.MapFile{
				Data: []byte(`{"extensions":{"opamp":{"enabled":true}}}`),
			},
		},
	}
	if err := DisableHostmetricsAction(ctx); err == nil {
		t.Fatal("expected non-nil error")
	}
}

func TestDisableHostmetricsAction(t *testing.T) {
	var called bool
	unlinkFunc := func() error {
		called = true
		return nil
	}
	errUnlinkFunc := func() error {
		return errors.New("error")
	}
	ctx := &actionContext{
		ConfigDir:         fstest.MapFS{},
		UnlinkHostMetrics: unlinkFunc,
	}
	if err := DisableHostmetricsAction(ctx); err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Error("hostmetrics unlinker not called")
	}
	ctx.UnlinkHostMetrics = errUnlinkFunc
	if err := DisableHostmetricsAction(ctx); err == nil {
		t.Fatal("expected non-nil error")
	}

}
