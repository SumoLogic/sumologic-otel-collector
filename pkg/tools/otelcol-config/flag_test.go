package main

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/spf13/pflag"
)

func TestFlagActions(t *testing.T) {
	// test that flagActions is built correctly
	fv := newFlagValues()
	fs := makeFlagSet(fv)
	fs.VisitAll(func(flag *pflag.Flag) {
		action := flagActions[flag.Name]
		if action == nil {
			t.Errorf("undefined flag action: %s", flag.Name)
		}
	})
}

func TestAddTag(t *testing.T) {
	tests := []struct {
		name     string
		flags    []string
		wantTags map[string]string
	}{
		{
			name:     "simple",
			flags:    []string{"otelcol-config", "--add-tag", "foo=bar"},
			wantTags: map[string]string{"foo": "bar"},
		},
		{
			name:     "multiple",
			flags:    []string{"otelcol-config", "--add-tag", "foo=bar", "--add-tag", "bar=biff"},
			wantTags: map[string]string{"foo": "bar", "bar": "biff"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fv := newFlagValues()
			fs := makeFlagSet(fv)
			if err := fs.Parse(test.flags); err != nil {
				t.Fatal(err)
			}
			if got, want := fv.AddTags, test.wantTags; !cmp.Equal(got, want) {
				t.Errorf("bad flag values: got %v, want %v", got, want)
			}
		})
	}
}

func TestDeleteTag(t *testing.T) {
	tests := []struct {
		name     string
		flags    []string
		wantTags []string
	}{
		{
			name:     "simple",
			flags:    []string{"otelcol-config", "--delete-tag", "foo"},
			wantTags: []string{"foo"},
		},
		{
			name:     "multiple",
			flags:    []string{"otelcol-config", "--delete-tag", "foo", "--delete-tag", "bar"},
			wantTags: []string{"foo", "bar"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fv := newFlagValues()
			fs := makeFlagSet(fv)
			if err := fs.Parse(test.flags); err != nil {
				t.Fatal(err)
			}
			if got, want := fv.DeleteTags, test.wantTags; !cmp.Equal(got, want) {
				t.Errorf("bad flag values: got %v, want %v", got, want)
			}
		})
	}
}

func TestSetInstallationToken(t *testing.T) {
	fv := newFlagValues()
	fs := makeFlagSet(fv)
	if err := fs.Parse([]string{"otelcol-config", "--set-installation-token", "abcdef"}); err != nil {
		t.Fatal(err)
	}
	if got, want := fv.InstallationToken, "abcdef"; !cmp.Equal(got, want) {
		t.Errorf("bad flag values: got %v, want %v", got, want)
	}
}

func TestEnableHostmetrics(t *testing.T) {
	fv := newFlagValues()
	fs := makeFlagSet(fv)
	if err := fs.Parse([]string{"otelcol-config", "--enable-hostmetrics"}); err != nil {
		t.Fatal(err)
	}
	if got, want := fv.EnableHostmetrics, true; !cmp.Equal(got, want) {
		t.Errorf("bad flag values: got %v, want %v", got, want)
	}
}

func TestDisableHostmetrics(t *testing.T) {
	fv := newFlagValues()
	fs := makeFlagSet(fv)
	if err := fs.Parse([]string{"otelcol-config", "--disable-hostmetrics"}); err != nil {
		t.Fatal(err)
	}
	if got, want := fv.DisableHostmetrics, true; !cmp.Equal(got, want) {
		t.Errorf("bad flag values: got %v, want %v", got, want)
	}
}

func TestEnableEphemeral(t *testing.T) {
	fv := newFlagValues()
	fs := makeFlagSet(fv)
	if err := fs.Parse([]string{"otelcol-config", "--enable-ephemeral"}); err != nil {
		t.Fatal(err)
	}
	if got, want := fv.EnableEphemeral, true; !cmp.Equal(got, want) {
		t.Errorf("bad flag values: got %v, want %v", got, want)
	}
}

func TestDisableEphemeral(t *testing.T) {
	fv := newFlagValues()
	fs := makeFlagSet(fv)
	if err := fs.Parse([]string{"otelcol-config", "--disable-ephemeral"}); err != nil {
		t.Fatal(err)
	}
	if got, want := fv.DisableEphemeral, true; !cmp.Equal(got, want) {
		t.Errorf("bad flag values: got %v, want %v", got, want)
	}
}

func TestEnableRemoteControl(t *testing.T) {
	fv := newFlagValues()
	fs := makeFlagSet(fv)
	if err := fs.Parse([]string{"otelcol-config", "--enable-remote-control"}); err != nil {
		t.Fatal(err)
	}
	if got, want := fv.EnableRemoteControl, true; !cmp.Equal(got, want) {
		t.Errorf("bad flag values: got %v, want %v", got, want)
	}
}

func TestDisableRemoteControl(t *testing.T) {
	fv := newFlagValues()
	fs := makeFlagSet(fv)
	if err := fs.Parse([]string{"otelcol-config", "--disable-remote-control"}); err != nil {
		t.Fatal(err)
	}
	if got, want := fv.DisableRemoteControl, true; !cmp.Equal(got, want) {
		t.Errorf("bad flag values: got %v, want %v", got, want)
	}
}

func TestSetOpampEndpoint(t *testing.T) {
	fv := newFlagValues()
	fs := makeFlagSet(fv)
	if err := fs.Parse([]string{"otelcol-config", "--set-opamp-endpoint", "wss://example.com"}); err != nil {
		t.Fatal(err)
	}
	if got, want := fv.SetOpAmpEndpoint, "wss://example.com"; !cmp.Equal(got, want) {
		t.Errorf("bad flag values: got %v, want %v", got, want)
	}
}

func TestSetTimezone(t *testing.T) {
	fv := newFlagValues()
	fs := makeFlagSet(fv)
	if err := fs.Parse([]string{"otelcol-config", "--set-timezone", "UTC"}); err != nil {
		t.Fatal(err)
	}
	if got, want := fv.SetTimezone, "UTC"; !cmp.Equal(got, want) {
		t.Errorf("bad flag values: got %v, want %v", got, want)
	}
}

func TestConfigDir(t *testing.T) {
	fv := newFlagValues()
	fs := makeFlagSet(fv)
	if err := fs.Parse([]string{"otelcol-config", "--config", "/etc/foosumo"}); err != nil {
		t.Fatal(err)
	}
	if got, want := fv.ConfigDir, "/etc/foosumo"; !cmp.Equal(got, want) {
		t.Errorf("bad flag values: got %v, want %v", got, want)
	}
}

func TestConfigDirUnset(t *testing.T) {
	fv := newFlagValues()
	fs := makeFlagSet(fv)
	if err := fs.Parse([]string{"otelcol-config"}); err != nil {
		t.Fatal(err)
	}
	if got, want := fv.ConfigDir, "/etc/otelcol-sumo"; !cmp.Equal(got, want) {
		t.Errorf("bad flag values: got %v, want %v", got, want)
	}
}
