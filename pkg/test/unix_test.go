//go:build integration && !windows
// +build integration,!windows

package sumologic_tests

import (
	"testing"
)

func TestUnix(t *testing.T) {
	testList := []testSpec{
		{
			name:        "ValidateConfigFileExporter",
			validations: []checkFunc{checkValidateOutput},
			args:        []string{validateCommand, configTag, "./testdata/config/config-file-exporter-valid.yaml"},
		},
		{
			name:        "ValidateConfigInvalidConfig",
			validations: []checkFunc{checkInvalidValidateOutput},
			args:        []string{validateCommand, configTag, "./testdata/config/config-file-exporter-invalid.yaml"},
		},
		{
			name:        "ValidateIngestionFileExporter",
			validations: []checkFunc{checkLogFileCreated},
			args:        []string{configTag, "./testdata/config/config-file-exporter-valid.yaml"},
			timeout_ms:  120000,
		},
		{
			name:        "ValidateConfigSumologicExporter",
			validations: []checkFunc{checkValidateOutput},
			args:        []string{validateCommand, configTag, "./testdata/config/config-file-sumologic-exporter-valid.yaml"},
		},
		{
			name:        "ValidateIngestionSumologicExporter",
			validations: []checkFunc{checkValidSumologicExporter, checkLogNumbersViaSumologicMock},
			preActions:  []checkFunc{preActionCreateCredentialsDir},
			args:        []string{configTag, "./testdata/config/config-file-sumologic-exporter-valid.yaml"},
			timeout_ms:  120000,
		},
	}

	for _, tt := range testList {
		t.Run(tt.name, func(t *testing.T) {
			if err := runTest(&tt, t); err != nil {
				t.Errorf("Test %s failed: %v", tt.name, err)
			}
		})
	}
}
