// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package opampprovider

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	"github.com/SumoLogic/sumologic-otel-collector/pkg/configprovider/globprovider"
	"go.opentelemetry.io/collector/confmap"
	"gopkg.in/yaml.v2"
)

const (
	SchemeName = "opamp"
)

// ConfigFragment is part of a larger opamp configuration structure.
// To avoid creating a dependency on other packages, we only specify what
// is needed for the OpAmp provider to work.
type ConfigFragment struct {
	RemoteConfigurationDirectory string `yaml:"remote_configuration_directory"`
}

// Provider is an OpAmp configuration provider. It requires a GlobProvider to
// load the contents of the supplied remote configuration directory.
type Provider struct {
	GlobProvider confmap.Provider
}

// New creates a new Provider, with its GlobProvider set to the result of
// globprovider.New().
func New() confmap.Provider {
	return &Provider{GlobProvider: globprovider.New()}
}

func (p *Provider) Retrieve(ctx context.Context, uri string, fn confmap.WatcherFunc) (*confmap.Retrieved, error) {
	var cfg ConfigFragment
	url, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(url.Path)
	if err != nil {
		return nil, fmt.Errorf("couldn't load opamp config file: %s", err)
	}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("couldn't parse opamp config file: %s", err)
	}
	glob := p.GlobProvider
	return glob.Retrieve(ctx, glob.Scheme()+":"+filepath.Join(cfg.RemoteConfigurationDirectory, "*.yaml"), fn)
}

func (*Provider) Scheme() string {
	return SchemeName
}

func (p *Provider) Shutdown(ctx context.Context) error {
	return p.GlobProvider.Shutdown(ctx)
}
