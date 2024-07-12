//go:build unix

package main

import (
	"testing"
	"testing/fstest"
)

func TestDisableRemoteControlActionNoConfigFilePresent(t *testing.T) {
	values := &flagValues{
		DisableRemoteControl: true,
	}
	ctx := &actionContext{
		ConfigDir:            fstest.MapFS{},
		Flags:                values,
		Stdout:               errWriter{},
		Stderr:               errWriter{},
		WriteConfD:           errWriter{}.Write,
		WriteConfDOverrides:  errWriter{}.Write,
		WriteSumologicRemote: errWriter{}.Write,
	}

	if err := DisableRemoteControlAction(ctx); err != nil {
		t.Fatal(err)
	}
}

func TestDisableRemoteControlActionConfigFilePresent(t *testing.T) {
	values := &flagValues{
		DisableRemoteControl: true,
	}
	ctx := &actionContext{
		ConfigDir: fstest.MapFS{
			SumologicRemoteDotYaml: &fstest.MapFile{
				Data: []byte(`{"extensions":{"opamp":{"enabled":true}}}`),
			},
		},
		Flags:                values,
		Stdout:               errWriter{},
		Stderr:               errWriter{},
		WriteConfD:           errWriter{}.Write,
		WriteConfDOverrides:  errWriter{}.Write,
		WriteSumologicRemote: newTestWriter(nil).Write,
	}

	if err := DisableRemoteControlAction(ctx); err != nil {
		t.Fatal(err)
	}
}

func TestEnableRemoteControlConfigFilePresent(t *testing.T) {
	values := &flagValues{
		EnableRemoteControl: true,
	}
	ctx := &actionContext{
		ConfigDir: fstest.MapFS{
			SumologicRemoteDotYaml: &fstest.MapFile{
				Data: []byte(`{"extensions":{"opamp":{"enabled":true}}}`),
			},
		},
		Flags:                values,
		Stdout:               errWriter{},
		Stderr:               errWriter{},
		WriteConfD:           errWriter{}.Write,
		WriteConfDOverrides:  errWriter{}.Write,
		WriteSumologicRemote: errWriter{}.Write,
	}

	if err := EnableRemoteControlAction(ctx); err != nil {
		t.Fatal(err)
	}
}

func TestEnableRemoteControlConfigFileNotPresent(t *testing.T) {
	values := &flagValues{
		EnableRemoteControl: true,
	}
	const expData = `extensions:
  opamp:
    enabled: true
    endpoint: wss://opamp-events.sumologic.com/v1/opamp
    remote_configuration_directory: /etc/otelcol-sumo/opamp.d
`
	slrWriter := newTestWriter([]byte(expData))
	ctx := &actionContext{
		ConfigDir:            fstest.MapFS{},
		Flags:                values,
		Stdout:               errWriter{},
		Stderr:               errWriter{},
		WriteConfD:           errWriter{}.Write,
		WriteConfDOverrides:  errWriter{}.Write,
		WriteSumologicRemote: slrWriter.Write,
	}

	if err := EnableRemoteControlAction(ctx); err != nil {
		t.Fatal(err)
	}
}
