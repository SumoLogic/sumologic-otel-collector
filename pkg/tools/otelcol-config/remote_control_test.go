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
	const expData = `exporters:
  nop: {}
extensions:
  file_storage:
    compaction:
      directory: /var/lib/otelcol-sumo/file_storage
      on_rebound: true
    directory: /var/lib/otelcol-sumo/file_storage
  health_check:
    endpoint: localhost:13133
  opamp:
    endpoint: wss://opamp-collectors.sumologic.com/v1/opamp
    remote_configuration_directory: /etc/otelcol-sumo/opamp.d
  sumologic:
    collector_credentials_directory: /var/lib/otelcol-sumo/credentials
    installation_token: ${SUMOLOGIC_INSTALLATION_TOKEN}
    time_zone: UTC
receivers:
  nop: {}
service:
  extensions:
    - sumologic
    - health_check
    - file_storage
    - opamp
  pipelines:
    logs/default:
      exporters:
        - nop
      receivers:
        - nop
    metrics/default:
      exporters:
        - nop
      receivers:
        - nop
    traces/default:
      exporters:
        - nop
      receivers:
        - nop
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
