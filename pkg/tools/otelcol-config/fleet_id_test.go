package main

import (
	"io"
	"io/fs"
	"testing"
	"testing/fstest"
)

func TestSetFleetIDAction(t *testing.T) {
	tests := []struct {
		Name           string
		Flags          []string
		Conf           fs.FS
		ExpectedWriter []byte
		ExpectedErr    bool
	}{
		{
			Name:           "valid fleet ID, no existing config",
			Flags:          []string{"--set-fleet-id", "000000000ABC1234"},
			Conf:           fstest.MapFS{},
			ExpectedWriter: []byte("extensions:\n  sumologic:\n    fleet_id: 000000000ABC1234\n"),
		},
		{
			Name:  "valid fleet ID, remote control enabled",
			Flags: []string{"--set-fleet-id", "000000000ABC1234"},
			Conf: fstest.MapFS{
				SumologicRemoteDotYaml: &fstest.MapFile{
					Data: []byte("extensions:\n  opamp:\n    enabled: true\n"),
				},
			},
			ExpectedWriter: []byte("extensions:\n  opamp:\n    enabled: true\n  sumologic:\n    fleet_id: 000000000ABC1234\n"),
		},
		{
			Name:           "valid fleet ID with override",
			Flags:          []string{"--set-fleet-id", "00000000DEADBEEF", "--override"},
			Conf:           fstest.MapFS{},
			ExpectedWriter: []byte("extensions:\n  sumologic:\n    fleet_id: 00000000DEADBEEF\n"),
		},
		{
			Name:           "valid lowercase hex",
			Flags:          []string{"--set-fleet-id", "abcdef0123456789"},
			Conf:           fstest.MapFS{},
			ExpectedWriter: []byte("extensions:\n  sumologic:\n    fleet_id: abcdef0123456789\n"),
		},
		{
			Name:        "invalid: empty string",
			Flags:       []string{"--set-fleet-id", ""},
			Conf:        fstest.MapFS{},
			ExpectedErr: true,
		},
		{
			Name:        "invalid: too short",
			Flags:       []string{"--set-fleet-id", "ABC1234"},
			Conf:        fstest.MapFS{},
			ExpectedErr: true,
		},
		{
			Name:        "invalid: too long",
			Flags:       []string{"--set-fleet-id", "000000000ABC12345"},
			Conf:        fstest.MapFS{},
			ExpectedErr: true,
		},
		{
			Name:        "invalid: non-hex characters",
			Flags:       []string{"--set-fleet-id", "000000000ABCXYZ1"},
			Conf:        fstest.MapFS{},
			ExpectedErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			writer := newTestWriter(test.ExpectedWriter).Write
			errWriter := errWriter{}.Write

			flagValues := newFlagValues()
			flagSet := makeFlagSet(flagValues)

			if err := flagSet.Parse(test.Flags); err != nil {
				t.Fatalf("failed to parse flags: %v", err)
			}

			var (
				settingsWriter, overridesWriter, sumologicRemoteWriter func([]byte) (int, error)
			)

			switch {
			case flagValues.Override:
				settingsWriter = errWriter
				sumologicRemoteWriter = errWriter
				overridesWriter = writer
			case flagValues.EnableRemoteControl || remoteControlEnabled(t, test.Conf):
				settingsWriter = errWriter
				overridesWriter = errWriter
				sumologicRemoteWriter = writer
			default:
				overridesWriter = errWriter
				sumologicRemoteWriter = errWriter
				settingsWriter = writer
			}

			ctx := &actionContext{
				ConfigDir:            test.Conf,
				Flags:                flagValues,
				Stdout:               io.Discard,
				Stderr:               io.Discard,
				WriteConfD:           settingsWriter,
				WriteConfDOverrides:  overridesWriter,
				WriteSumologicRemote: sumologicRemoteWriter,
			}

			err := SetFleetIDAction(ctx)
			if err != nil && !test.ExpectedErr {
				t.Fatal(err)
			}
			if err == nil && test.ExpectedErr {
				t.Fatal("expected non-nil error")
			}
		})
	}
}

func TestValidateFleetID(t *testing.T) {
	tests := []struct {
		Name    string
		ID      string
		WantErr bool
	}{
		{"valid uppercase", "000000000ABC1234", false},
		{"valid lowercase", "abcdef0123456789", false},
		{"valid mixed case", "00aaBB11ccDD2233", false},
		{"empty", "", true},
		{"too short", "ABC1234", true},
		{"too long", "000000000ABC12345", true},
		{"non-hex chars", "000000000ABCXYZ1", true},
		{"spaces", "  ", true},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			err := validateFleetID(test.ID)
			if (err != nil) != test.WantErr {
				t.Errorf("validateFleetID(%q) error = %v, wantErr %v", test.ID, err, test.WantErr)
			}
		})
	}
}
