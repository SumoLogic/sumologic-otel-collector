package jobreceiver

import "testing"

func TestConfigValidate(t *testing.T) {
	cfg := Config{}
	if err := cfg.Validate(); err == nil {
		t.Errorf("expected error on empty config")
	}
}
