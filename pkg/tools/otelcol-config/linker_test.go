//go:build unix

package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

func TestHostMetricsLinker(t *testing.T) {
	dir, err := os.MkdirTemp("", "otelcol-config-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	values := &flagValues{
		ConfigDir: dir,
	}

	availPath := filepath.Join(dir, ConfDotDAvailable)
	if err := os.Mkdir(availPath, 0700); err != nil {
		t.Fatal(err)
	}

	confdPath := filepath.Join(dir, ConfDotD)
	if err := os.Mkdir(confdPath, 0700); err != nil {
		t.Fatal(err)
	}

	configPath := filepath.Join(availPath, hostmetricsFilename)
	if err := os.WriteFile(configPath, []byte("configuration: yes"), 0600); err != nil {
		t.Fatal(err)
	}

	linker := getHostMetricsLinker(values)

	if err := linker(); err != nil {
		t.Fatal(fmt.Errorf("can't link config path %s: %s", configPath, err))
	}

	linkPath := filepath.Join(dir, ConfDotD, hostmetricsFilename)

	stat, err := os.Lstat(linkPath)
	if err != nil {
		t.Fatal(err)
	}

	if stat.Mode().Type() != fs.ModeSymlink {
		t.Error("not a symlink")
	}

	// we expect that calling the linker again succeeds even if the link exists
	if err := linker(); err != nil {
		t.Fatal(fmt.Errorf("can't link config path %s: %s", configPath, err))
	}

	unlinker := getHostMetricsUnlinker(values)

	// unlink the link
	if err := unlinker(); err != nil {
		t.Fatal(err)
	}

	// verify the link is gone
	_, err = os.Lstat(linkPath)
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if _, ok := err.(*os.PathError); !ok {
		t.Fatal(err)
	}

	// verify the linked file is still there
	_, err = os.Stat(configPath)
	if err != nil {
		t.Fatal(err)
	}
}
