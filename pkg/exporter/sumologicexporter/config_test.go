package sumologicexporter

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/confighttp"
)

func TestInitExporterInvalidConfiguration(t *testing.T) {
	testcases := []struct {
		name          string
		cfg           *Config
		expectedError error
	}{
		{
			name:          "unexpected log format",
			expectedError: errors.New("unexpected log format: test_format"),
			cfg: &Config{
				LogFormat:    "test_format",
				MetricFormat: "otlp",
				TraceFormat:  "otlp",
				ClientConfig: confighttp.ClientConfig{
					Timeout:  defaultTimeout,
					Endpoint: "test_endpoint",
				},
			},
		},
		{
			name:          "unexpected metric format",
			expectedError: errors.New("unexpected metric format: test_format"),
			cfg: &Config{
				LogFormat:    "json",
				MetricFormat: "test_format",
				ClientConfig: confighttp.ClientConfig{
					Timeout:     defaultTimeout,
					Endpoint:    "test_endpoint",
					Compression: "gzip",
				},
			},
		},
		{
			name:          "unsupported Carbon2 metrics format",
			expectedError: errors.New("support for the carbon2 metric format was removed, please use prometheus or otlp instead"),
			cfg: &Config{
				LogFormat:    "json",
				MetricFormat: "carbon2",
				ClientConfig: confighttp.ClientConfig{
					Timeout:     defaultTimeout,
					Endpoint:    "test_endpoint",
					Compression: "gzip",
				},
			},
		},
		{
			name:          "unsupported Graphite metrics format",
			expectedError: errors.New("support for the graphite metric format was removed, please use prometheus or otlp instead"),
			cfg: &Config{
				LogFormat:    "json",
				MetricFormat: "graphite",
				ClientConfig: confighttp.ClientConfig{
					Timeout:     defaultTimeout,
					Endpoint:    "test_endpoint",
					Compression: "gzip",
				},
			},
		},
		{
			name:          "unexpected trace format",
			expectedError: errors.New("unexpected trace format: text"),
			cfg: &Config{
				LogFormat:    "json",
				MetricFormat: "otlp",
				TraceFormat:  "text",
				ClientConfig: confighttp.ClientConfig{
					Timeout:     defaultTimeout,
					Endpoint:    "test_endpoint",
					Compression: "gzip",
				},
			},
		},
		{
			name:          "no endpoint and no auth extension specified",
			expectedError: errors.New("no endpoint and no auth extension specified"),
			cfg: &Config{
				LogFormat:    "json",
				MetricFormat: "otlp",
				TraceFormat:  "otlp",
				ClientConfig: confighttp.ClientConfig{
					Timeout:     defaultTimeout,
					Compression: "gzip",
				},
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			err := component.ValidateConfig(tc.cfg)

			if tc.expectedError != nil {
				assert.EqualError(t, err, tc.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
