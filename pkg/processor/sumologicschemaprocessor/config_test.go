// Copyright 2022 Sumo Logic, Inc.
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

package sumologicschemaprocessor

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/service/servicetest"
)

func TestLoadConfig(t *testing.T) {
	factories, err := componenttest.NopFactories()
	assert.NoError(t, err)

	factory := NewFactory()
	factories.Processors[typeStr] = factory
	cfg, err := servicetest.LoadConfigAndValidate(filepath.Join("testdata", "config.yaml"), factories)

	require.Nil(t, err)
	require.NotNil(t, cfg)

	p0 := cfg.Processors[component.NewID(typeStr)]
	assert.Equal(t, p0, factory.CreateDefaultConfig())

	p1 := cfg.Processors[component.NewIDWithName(typeStr, "disabled-cloud-namespace")]

	assert.Equal(t, p1,
		&Config{
			ProcessorSettings: config.NewProcessorSettings(component.NewIDWithName(
				typeStr,
				"disabled-cloud-namespace",
			)),
			AddCloudNamespace:           false,
			TranslateAttributes:         true,
			TranslateTelegrafAttributes: true,
		})

	p2 := cfg.Processors[component.NewIDWithName(typeStr, "disabled-attribute-translation")]

	assert.Equal(t, p2,
		&Config{
			ProcessorSettings: config.NewProcessorSettings(component.NewIDWithName(
				typeStr,
				"disabled-attribute-translation",
			)),
			AddCloudNamespace:           true,
			TranslateAttributes:         false,
			TranslateTelegrafAttributes: true,
		})

	p3 := cfg.Processors[component.NewIDWithName(typeStr, "disabled-telegraf-attribute-translation")]

	assert.Equal(t, p3,
		&Config{
			ProcessorSettings: config.NewProcessorSettings(component.NewIDWithName(
				typeStr,
				"disabled-telegraf-attribute-translation",
			)),
			AddCloudNamespace:           true,
			TranslateAttributes:         true,
			TranslateTelegrafAttributes: false,
		})
}
