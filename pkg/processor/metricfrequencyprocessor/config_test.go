package metricfrequencyprocessor

import (
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap"
	"go.opentelemetry.io/collector/confmap/converter/expandconverter"
	"go.opentelemetry.io/collector/confmap/provider/fileprovider"
	"go.opentelemetry.io/collector/confmap/provider/yamlprovider"
	"go.opentelemetry.io/collector/otelcol"
	"go.opentelemetry.io/collector/otelcol/otelcoltest"
)

func TestLoadConfig(t *testing.T) {
	factories, err := otelcoltest.NopFactories()
	require.NoError(t, err)

	factory := NewFactory()
	factories.Processors[factory.Type()] = factory

	cfg, err := otelcoltest.LoadConfigWithSettings(factories, otelcol.ConfigProviderSettings{
		ResolverSettings: confmap.ResolverSettings{
			URIs: []string{path.Join(".", "testdata", "config.yaml")},
			ProviderFactories: []confmap.ProviderFactory{
				fileprovider.NewFactory(),
				yamlprovider.NewFactory(),
			},
			ConverterFactories: []confmap.ConverterFactory{expandconverter.NewFactory()},
		},
	})
	require.NoError(t, err)
	require.NotNil(t, cfg)

	id := component.NewID(Type)

	assert.Equal(t, cfg.Processors[id], createDefaultConfig())
}
