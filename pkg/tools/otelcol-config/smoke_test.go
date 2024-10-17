//go:build !windows

package main

import (
	"os"
	"path/filepath"
	"testing"
)

func createSkeleton(dir string) error {
	if err := os.Mkdir(filepath.Join(dir, ConfDotD), 0770); err != nil {
		return err
	}
	if err := os.Mkdir(filepath.Join(dir, ConfDotDAvailable), 0770); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, ConfDotDAvailable, hostmetricsYAML), []byte("hostmetrics"), 0660); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, ConfDotDAvailable, ephemeralYAML), []byte("ephemeral"), 0660); err != nil {
		return err
	}
	return nil
}

func TestLocallyManagedSmoke(t *testing.T) {
	// run main against a tempdir with all locally managed flags
	configDir := t.TempDir()
	if err := createSkeleton(configDir); err != nil {
		t.Fatal(err)
	}
	flags := []string{
		"otelcol-config",
		"--config", configDir,
		"--add-tag", "foo=bar",
		"--add-tag", "bar=baz",
		"--delete-tag", "bar",
		"--set-installation-token", "abcdef",
		"--enable-hostmetrics",
		"--disable-hostmetrics",
		"--enable-ephemeral",
		"--disable-ephemeral",
		"--set-api-url", "https://example.com",
		"--write-kv", ".hello.world = \"yes\"",
	}

	os.Args = flags

	main()
}

func TestRemotelyManagedSmoke(t *testing.T) {
	// run main against a tempdir with all remotely managed flags
	configDir := t.TempDir()
	if err := createSkeleton(configDir); err != nil {
		t.Fatal(err)
	}
	flags := []string{
		"otelcol-config",
		"--config", configDir,
		"--set-opamp-endpoint", "ws://example.com",
		"--set-api-url", "https://example.com",
		"--enable-remote-control",
	}

	os.Args = flags

	main()
}
