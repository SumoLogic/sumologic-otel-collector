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
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/SumoLogic/sumologic-otel-collector/pkg/configprovider/globprovider"
	"go.opentelemetry.io/collector/confmap"
	"go.opentelemetry.io/collector/confmap/provider/fileprovider"
	"gopkg.in/yaml.v2"
)

const (
	SchemeName = "opamp"
)

// ConfigFragment is part of a larger opamp configuration structure.
// To avoid creating a dependency on other packages, we only specify what
// is needed for the OpAmp provider to work.
type ConfigFragment struct {
	Extensions struct {
		OpAmp struct {
			RemoteConfigurationDirectory string `yaml:"remote_configuration_directory"`
		} `yaml:"opamp"`
	} `yaml:"extensions"`
}

func (c ConfigFragment) ConfigDir() string {
	return c.Extensions.OpAmp.RemoteConfigurationDirectory
}

func (c ConfigFragment) Validate() error {
	if c.ConfigDir() == "" {
		return errors.New("remote_configuration_directory missing from opamp extension")
	}
	return nil
}

// Provider is an OpAmp configuration provider. It requires a GlobProvider to
// load the contents of the supplied remote configuration directory.
type Provider struct {
	GlobProvider confmap.Provider
}

// New creates a new Provider, with its GlobProvider set to the result of
// globprovider.New().
func NewWithSettings(settings confmap.ProviderSettings) confmap.Provider {
	return &Provider{GlobProvider: globprovider.NewWithSettings(settings)}
}

func NewFactory() confmap.ProviderFactory {
	return confmap.NewProviderFactory(NewWithSettings)
}

// Remove specific keys from srcMap if same keys present in mergeMap under same path
func removeMatchingKeysFromSrcMap(srcMap map[string]interface{}, mergeMap map[string]interface{}) map[string]interface{} {
	srcMapVal := srcMap
	mergeMapVal := mergeMap
	for key, mergeValue := range mergeMapVal {
		srcValue, exists := srcMapVal[key]

		//if key not exists in merge map, skip the path
		if !exists {
			continue
		}

		if key == "collector_fields" {
			delete(srcMapVal, key)
		}

		mergeNestedMap, isMergeMap := mergeValue.(map[string]interface{})
		srcNestedMap, isSrcMap := srcValue.(map[string]interface{})

		if isMergeMap && isSrcMap {
			// Recursively handle nested maps
			removeMatchingKeysFromSrcMap(srcNestedMap, mergeNestedMap)
		}
	}
	return srcMap
}

// Checks and prepares scrConf for replace behavior instead of map merge
// Hack for confmap.Conf.Merge method to replace specific fields instead of merging them
// Merge method merges field values from source and mergeConf, so by removing existing values from
// source map, we can achieve replace behavior
func PrepareForReplaceBehavior(srcConf *confmap.Conf, mergeConf *confmap.Conf) {
	*srcConf = *confmap.NewFromStringMap(removeMatchingKeysFromSrcMap(srcConf.ToStringMap(), mergeConf.ToStringMap()))
}

func (p *Provider) Retrieve(ctx context.Context, configPath string, fn confmap.WatcherFunc) (*confmap.Retrieved, error) {
	var cfg ConfigFragment
	configPath = strings.TrimPrefix(configPath, "opamp:")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("couldn't load opamp config file: %s", err)
	}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("couldn't parse opamp config file: %s", err)
	}
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid opamp config file: %s", err)
	}
	conf := confmap.New()
	glob := p.GlobProvider
	retrieved, err := glob.Retrieve(ctx, glob.Scheme()+":"+filepath.Join(cfg.ConfigDir(), "*.yaml"), fn)
	if err != nil {
		return nil, err
	}
	retConf, err := retrieved.AsConf()
	if err != nil {
		return nil, err
	}
	addl, err := fileprovider.NewFactory().Create(confmap.ProviderSettings{}).Retrieve(ctx, "file:"+configPath, fn)
	if err != nil {
		return nil, err
	}
	addlConf, err := addl.AsConf()
	if err != nil {
		return nil, err
	}
	//Order of conf parameter is imp, see method comments
	PrepareForReplaceBehavior(addlConf, retConf)
	// merge the file config in
	if err := conf.Merge(addlConf); err != nil {
		return nil, err
	}
	// merge the glob config in, potentially overriding file config
	if err := conf.Merge(retConf); err != nil {
		return nil, err
	}
	return confmap.NewRetrieved(conf.ToStringMap())
}

func (*Provider) Scheme() string {
	return SchemeName
}

func (p *Provider) Shutdown(ctx context.Context) error {
	return p.GlobProvider.Shutdown(ctx)
}
