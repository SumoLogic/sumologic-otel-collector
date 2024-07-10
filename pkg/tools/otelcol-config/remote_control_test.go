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
	slrWriter := newTestWriter([]byte("extensions:\n  opamp:\n    endpoint: wss://opamp-events.sumologic.com/v1/opamp\n    remote_configuration_directory: /etc/otelcol-sumo/opamp.d\n    enabled: true\nextensions:\n  opamp:\n    endpoint: wss://opamp-events.sumologic.com/v1/opamp\n    remote_configuration_directory: /etc/otelcol-sumo/opamp.d\n    enabled: true\n"))
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
