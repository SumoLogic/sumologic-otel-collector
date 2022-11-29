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

package opampprovider

import (
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/open-telemetry/opamp-go/protobufs"
	"go.opentelemetry.io/collector/confmap"
)

type config map[string]interface{}

func (c *config) composeOtConfig() (map[string]interface{}, error) {
	// Make a copy of the string map.
	sm := confmap.NewFromStringMap(*c).ToStringMap()

	// Bindplane returning configuration that doesn't work with our OT distribution.
	delete(sm, "labels")

	return sm, nil
}

func (c *config) composeEffectiveConfigProto() (*protobufs.EffectiveConfig, error) {
	bytes, err := yaml.Parser().Marshal(*c)
	if err != nil {
		return nil, err
	}

	return &protobufs.EffectiveConfig{
		ConfigMap: &protobufs.AgentConfigMap{
			ConfigMap: map[string]*protobufs.AgentConfigFile{
				"": {Body: bytes},
			},
		},
	}, nil
}
