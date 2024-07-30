package main

import "testing"

func TestNotImplementedAction(t *testing.T) {
	if got, want := notImplementedAction(nil), errNotImplemented; got != want {
		t.Errorf("bad error: got %q, want %q", got, want)
	}
}

func TestNullAction(t *testing.T) {
	if err := nullAction(nil); err != nil {
		t.Error(err)
	}
}
