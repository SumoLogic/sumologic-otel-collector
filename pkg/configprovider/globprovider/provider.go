// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package globprovider

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/SumoLogic/sumologic-otel-collector/pkg/configprovider/providerutil"
	"go.opentelemetry.io/collector/confmap"
	"gopkg.in/yaml.v3"
)

const (
	schemeName   = "glob"
	schemePrefix = schemeName + ":"
)

type Provider struct{
	remotelyManagedMergeFlow bool
}

func NewWithSettings(_ confmap.ProviderSettings) confmap.Provider {
	return &Provider{}
}

func NewFactory() confmap.ProviderFactory {
	return confmap.NewProviderFactory(NewWithSettings)
}

func (fmp *Provider) SetRemotelyManagedMergeFlow(enable bool) {
    fmp.remotelyManagedMergeFlow = enable
}

func (fmp *Provider) Retrieve(ctx context.Context, uri string, _ confmap.WatcherFunc) (*confmap.Retrieved, error) {
	var rawConf map[string]interface{}
	if !strings.HasPrefix(uri, schemePrefix) {
		return &confmap.Retrieved{}, fmt.Errorf("%q uri is not supported by %q provider", uri, schemeName)
	}

	globPattern := uri[len(schemePrefix):]
	paths, err := filepath.Glob(globPattern)
	if err != nil {
		return &confmap.Retrieved{}, err
	}

	// sort the paths alphabetically to have consistent ordering
	sort.Strings(paths)

	conf := confmap.New()
	for _, path := range paths {
		content, err := os.ReadFile(path)
		if err != nil {
			return &confmap.Retrieved{}, err
		}

		if err := yaml.Unmarshal(content, &rawConf); err != nil {
			return &confmap.Retrieved{}, err
		}
		pathConf := confmap.NewFromStringMap(rawConf)
		if fmp.remotelyManagedMergeFlow {
			providerutil.PrepareForReplaceBehavior(conf, pathConf)
		}
		if err := conf.Merge(pathConf); err != nil {
			return &confmap.Retrieved{}, err
		}

	}
	return confmap.NewRetrieved(conf.ToStringMap())
}

func (*Provider) Scheme() string {
	return schemeName
}

func (fmp *Provider) Shutdown(context.Context) error {
	return nil
}
