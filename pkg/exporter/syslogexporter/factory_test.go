package syslogexporter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/exporter/exporterhelper"
)

func TestType(t *testing.T) {
	factory := NewFactory()
	pType := factory.Type()
	assert.Equal(t, pType, component.Type("syslog"))
}

func TestCreateDefaultConfig(t *testing.T) {
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig()
	qs := exporterhelper.NewDefaultQueueSettings()
	qs.Enabled = false

	assert.Equal(t, cfg, &Config{
		Endpoint:      "host.domain.com",
		Port:          514,
		Protocol:      "tcp",
		CACertificate: "",
		Format:        "rfc5424",
		QueueSettings: exporterhelper.QueueSettings{
			Enabled:      false,
			NumConsumers: 0,
			QueueSize:    0,
			StorageID:    (*component.ID)(nil)},
		RetrySettings: exporterhelper.RetrySettings{
			Enabled:         false,
			InitialInterval: 0,
			MaxInterval:     0,
			MaxElapsedTime:  0,
		},
	})
	assert.NoError(t, component.ValidateConfig(cfg))
}
