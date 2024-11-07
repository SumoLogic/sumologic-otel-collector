package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"testing"
)

func discardWriter([]byte) (int, error) {
	return 0, nil
}

type errWriter struct {
}

func (errWriter) Write([]byte) (int, error) {
	return 0, errors.New("writer called")
}

func makeTestActionContext(t testing.TB,
	confD fs.FS,
	flags []string,
	stdout, stderr io.Writer,
	writeConfD, writeConfDOverrides func([]byte) (int, error)) *actionContext {

	t.Helper()

	flagValues := newFlagValues()
	fs := makeFlagSet(flagValues)
	if err := fs.Parse(flags); err != nil {
		t.Fatal(err)
	}

	if stdout == nil {
		stdout = io.Discard
	}

	if stderr == nil {
		stderr = io.Discard
	}

	if writeConfD == nil {
		writeConfD = discardWriter
	}

	if writeConfDOverrides == nil {
		writeConfDOverrides = discardWriter
	}

	return &actionContext{
		ConfigDir:           confD,
		Flags:               flagValues,
		Stdout:              stdout,
		Stderr:              stderr,
		WriteConfD:          writeConfD,
		WriteConfDOverrides: writeConfDOverrides,
	}
}

type testWriter struct {
	exp []byte
}

func (t *testWriter) Write(data []byte) (int, error) {
	if got, want := data, t.exp; !bytes.Equal(got, want) {
		return 0, fmt.Errorf("bad write: got %q, want %q", got, want)
	}
	return len(data), nil
}

func newTestWriter(exp []byte) *testWriter {
	return &testWriter{
		exp: exp,
	}
}
