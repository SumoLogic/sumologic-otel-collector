package syslogexporter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/exporter"
	"go.uber.org/zap"
)

func createExporterCreateSettings() exporter.CreateSettings {
	return exporter.CreateSettings{
		TelemetrySettings: component.TelemetrySettings{
			Logger: zap.NewNop(),
		},
	}
}

func TestInitExporter(t *testing.T) {
	_, err := initExporter(&Config{Endpoint: "test.com",
		Protocol: "tcp",
		Port:     514,
		Format:   "RFC5424"}, createExporterCreateSettings())
	assert.NoError(t, err)
}
