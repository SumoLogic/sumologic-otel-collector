// Copyright 2023, OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
		Endpoint: "host.domain.com",
		Port:     514,
		Protocol: "tcp",
		Format:   "rfc5424",
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
