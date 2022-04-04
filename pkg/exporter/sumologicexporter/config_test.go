package sumologicexporter

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/config/confighttp"
)

func TestInitExporterInvalidLogFormat(t *testing.T) {
	testcases := []struct {
		name          string
		cfg           *Config
		expectedError error
	}{
		{
			name:          "unexpected log format",
			expectedError: errors.New("unexpected log format: test_format"),
			cfg: &Config{
				LogFormat:        "test_format",
				MetricFormat:     "carbon2",
				CompressEncoding: "gzip",
				TraceFormat:      "otlp",
				HTTPClientSettings: confighttp.HTTPClientSettings{
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
				HTTPClientSettings: confighttp.HTTPClientSettings{
					Timeout:  defaultTimeout,
					Endpoint: "test_endpoint",
				},
				CompressEncoding: "gzip",
			},
		},
		{
			name:          "unexpected trace format",
			expectedError: errors.New("unexpected trace format: text"),
			cfg: &Config{
				LogFormat:    "json",
				MetricFormat: "carbon2",
				TraceFormat:  "text",
				HTTPClientSettings: confighttp.HTTPClientSettings{
					Timeout:  defaultTimeout,
					Endpoint: "test_endpoint",
				},
				CompressEncoding: "gzip",
			},
		},
		{
			name:          "unexpected compression encoding",
			expectedError: errors.New("invalid compression encoding type: test_format"),
			cfg: &Config{
				LogFormat:        "json",
				MetricFormat:     "carbon2",
				CompressEncoding: "test_format",
				TraceFormat:      "otlp",
				HTTPClientSettings: confighttp.HTTPClientSettings{
					Timeout:  defaultTimeout,
					Endpoint: "test_endpoint",
				},
			},
		},
		{
			name:          "no endpoint and no auth extension specified",
			expectedError: errors.New("no endpoint and no auth extension specified"),
			cfg: &Config{
				LogFormat:        "json",
				MetricFormat:     "carbon2",
				CompressEncoding: "gzip",
				TraceFormat:      "otlp",
				HTTPClientSettings: confighttp.HTTPClientSettings{
					Timeout: defaultTimeout,
				},
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			err := tc.cfg.Validate()

			if tc.expectedError != nil {
				assert.EqualError(t, err, tc.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
