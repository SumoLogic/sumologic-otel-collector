package syslogexporter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidate(t *testing.T) {

	tests := []struct {
		name string
		cfg  *Config
		err  string
	}{
		{
			name: "invalid Port",
			cfg: &Config{
				Port:     515444,
				Endpoint: "host.domain.com",
				Format:   "rfc5424",
				Protocol: "udp",
			},
			err: "Unsupported Port: Port is required, must be in the range 1-65535",
		},

		{
			name: "invalid Endpoint",
			cfg: &Config{
				Port:     514,
				Endpoint: "",
				Format:   "rfc5424",
				Protocol: "udp",
			},
			err: "Invalid FQDN: Endpoint is required, must be a valid FQDN",
		},

		{
			name: "unsupported Protocol",
			cfg: &Config{
				Port:     514,
				Endpoint: "host.domain.com",
				Format:   "rfc5424",
				Protocol: "ftp",
			},
			err: "Unsupported protocol: Protocol is required, only tcp/udp supported",
		},
		{
			name: "Unsupported Format",
			cfg: &Config{
				Port:     514,
				Endpoint: "host.domain.com",
				Protocol: "udp",
				Format:   "rfc",
			},
			err: "Unsupported format: Only rfc5424 and rfc3164 supported",
		},
	}
	for _, testInstance := range tests {
		t.Run(testInstance.name, func(t *testing.T) {
			err := testInstance.cfg.Validate()
			if testInstance.err != "" {
				assert.EqualError(t, err, testInstance.err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
