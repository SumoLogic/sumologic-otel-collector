package main

import (
	"errors"
	"testing"
	"testing/fstest"
)

func TestEnableEphemeralAction(t *testing.T) {
	var called bool
	linkFunc := func() error {
		called = true
		return nil
	}
	errLinkFunc := func() error {
		return errors.New("error")
	}
	ctx := &actionContext{
		ConfigDir:     fstest.MapFS{},
		LinkEphemeral: linkFunc,
	}
	if err := EnableEphemeralAction(ctx); err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Error("ephemeral linker not called")
	}
	ctx.LinkEphemeral = errLinkFunc
	if err := EnableEphemeralAction(ctx); err == nil {
		t.Fatal("expected non-nil error")
	}
}

func TestEnableEphemeralActionRemoteControlled(t *testing.T) {
	ctx := &actionContext{
		ConfigDir: fstest.MapFS{
			SumologicRemoteDotYaml: &fstest.MapFile{
				Data: []byte("extensions:\n  opamp:\n    enabled: true\n"),
			},
		},
		WriteSumologicRemote: newTestWriter([]byte("extensions:\n  opamp:\n    enabled: true\n  sumologic:\n    ephemeral: true\n")).Write,
	}
	if err := EnableEphemeralAction(ctx); err != nil {
		t.Fatal(err)
	}
}

func TestDisableEphemeralActionRemoteControlled(t *testing.T) {
	ctx := &actionContext{
		ConfigDir: fstest.MapFS{
			SumologicRemoteDotYaml: &fstest.MapFile{
				Data: []byte("extensions:\n  opamp:\n    enabled: true\n"),
			},
		},
		WriteSumologicRemote: newTestWriter([]byte("extensions:\n  opamp:\n    enabled: true\n  sumologic:\n    ephemeral: false\n")).Write,
	}
	if err := DisableEphemeralAction(ctx); err != nil {
		t.Fatal(err)
	}
}

func TestDisableEphemeralAction(t *testing.T) {
	var called bool
	unlinkFunc := func() error {
		called = true
		return nil
	}
	errUnlinkFunc := func() error {
		return errors.New("error")
	}
	ctx := &actionContext{
		ConfigDir:       fstest.MapFS{},
		UnlinkEphemeral: unlinkFunc,
	}
	if err := DisableEphemeralAction(ctx); err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Error("hostmetrics unlinker not called")
	}
	ctx.UnlinkEphemeral = errUnlinkFunc
	if err := DisableEphemeralAction(ctx); err == nil {
		t.Fatal("expected non-nil error")
	}

}
