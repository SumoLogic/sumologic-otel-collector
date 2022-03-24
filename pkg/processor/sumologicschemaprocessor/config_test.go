package sumologicschemaprocessor

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	p0 := cfg.Processors[config.NewComponentID(typeStr)]
	assert.Equal(t, p0, factory.CreateDefaultConfig())

	p1 := cfg.Processors[config.NewComponentIDWithName(typeStr, "disabled-cloud-namespace")]

	assert.Equal(t, p1,
		&Config{
			ProcessorSettings: config.NewProcessorSettings(config.NewComponentIDWithName(typeStr, "disabled-cloud-namespace")),
			AddCloudNamespace: false,
		})
}
