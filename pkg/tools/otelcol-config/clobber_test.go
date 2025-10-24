package main

import (
	"errors"
	"testing"
	"testing/fstest"
)

func TestEnableClobberAction(t *testing.T) {
	var called bool
	linkFunc := func() error {
		called = true
		return nil
	}
	errLinkFunc := func() error {
		return errors.New("error")
	}
	ctx := &actionContext{
		ConfigDir:   fstest.MapFS{},
		LinkClobber: linkFunc,
	}
	if err := EnableClobberAction(ctx); err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Error("clobber linker not called")
	}
	ctx.LinkClobber = errLinkFunc
	if err := EnableClobberAction(ctx); err == nil {
		t.Fatal("expected non-nil error")
	}
}

func TestEnableClobberActionRemoteControlled(t *testing.T) {
	ctx := &actionContext{
		ConfigDir: fstest.MapFS{
			SumologicRemoteDotYaml: &fstest.MapFile{
				Data: []byte("extensions:\n  opamp:\n    enabled: true\n"),
			},
		},
		WriteSumologicRemote: newTestWriter([]byte("extensions:\n  opamp:\n    enabled: true\n  sumologic:\n    clobber: true\n")).Write,
	}
	if err := EnableClobberAction(ctx); err != nil {
		t.Fatal(err)
	}
}

func TestDisableClobberActionRemoteControlled(t *testing.T) {
	ctx := &actionContext{
		ConfigDir: fstest.MapFS{
			SumologicRemoteDotYaml: &fstest.MapFile{
				Data: []byte("extensions:\n  opamp:\n    enabled: true\n"),
			},
		},
		WriteSumologicRemote: newTestWriter([]byte("extensions:\n  opamp:\n    enabled: true\n  sumologic:\n    clobber: false\n")).Write,
	}
	if err := DisableClobberAction(ctx); err != nil {
		t.Fatal(err)
	}
}

func TestDisableClobberAction(t *testing.T) {
	var called bool
	unlinkFunc := func() error {
		called = true
		return nil
	}
	errUnlinkFunc := func() error {
		return errors.New("error")
	}
	ctx := &actionContext{
		ConfigDir:     fstest.MapFS{},
		UnlinkClobber: unlinkFunc,
	}
	if err := DisableClobberAction(ctx); err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Error("clobber unlinker not called")
	}
	ctx.UnlinkClobber = errUnlinkFunc
	if err := DisableClobberAction(ctx); err == nil {
		t.Fatal("expected non-nil error")
	}
}
