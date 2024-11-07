// Copyright 2020 OpenTelemetry Authors
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

package k8sprocessor

import (
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/otelcol/otelcoltest"

	"github.com/open-telemetry/opentelemetry-collector-contrib/internal/k8sconfig"
)

func TestLoadConfig(t *testing.T) {
	factories, err := otelcoltest.NopFactories()
	require.NoError(t, err)
	factory := NewFactory()
	factories.Processors[Type] = factory

	require.NoError(t, component.ValidateConfig(factory.CreateDefaultConfig()))

	cfg, err := otelcoltest.LoadConfig(path.Join(".", "testdata", "config.yaml"), factories)
	require.NoError(t, err)
	require.NotNil(t, cfg)
	require.NoError(t, cfg.Validate())

	p0 := cfg.Processors[component.NewID(Type)]
	assert.EqualValues(t,
		&Config{
			APIConfig: k8sconfig.APIConfig{AuthType: k8sconfig.AuthTypeServiceAccount},
			Limit:     200,
			Extract:   ExtractConfig{Delimiter: ", "},
		},
		p0,
	)

	p1 := cfg.Processors[component.NewIDWithName(Type, "2")]
	assert.EqualValues(t,
		&Config{
			APIConfig:          k8sconfig.APIConfig{AuthType: k8sconfig.AuthTypeKubeConfig},
			Limit:              200,
			Passthrough:        false,
			OwnerLookupEnabled: true,
			Extract: ExtractConfig{
				Metadata: []string{
					"k8s.pod.name",
					"k8s.pod.uid",
					"k8s.deployment.name",
					"k8s.namespace.name",
					"k8s.node.name",
					"startTime",
				},
				Annotations: []FieldExtractConfig{
					{TagName: "a1", Key: "annotation-one"},
					{TagName: "a2", Key: "annotation-two", Regex: "field=(?P<value>.+)"},
				},
				Labels: []FieldExtractConfig{
					{TagName: "l1", Key: "label1"},
					{TagName: "l2", Key: "label2", Regex: "field=(?P<value>.+)"},
				},
				NamespaceAnnotations: []FieldExtractConfig{
					{TagName: "namespace_annotations_%s", Key: "*"},
				},
				NamespaceLabels: []FieldExtractConfig{
					{TagName: "namespace_labels_%s", Key: "*"},
				},
				Tags: map[string]string{
					"containerId": "my.namespace.containerId",
				},
				Delimiter: ", ",
			},
			Filter: FilterConfig{
				Namespace:      "ns2",
				Node:           "ip-111.us-west-2.compute.internal",
				NodeFromEnvVar: "K8S_NODE",
				Labels: []FieldFilterConfig{
					{Key: "key1", Value: "value1"},
					{Key: "key2", Value: "value2", Op: "not-equals"},
				},
				Fields: []FieldFilterConfig{
					{Key: "key1", Value: "value1"},
					{Key: "key2", Value: "value2", Op: "not-equals"},
				},
			},
			Association: []PodAssociationConfig{
				{
					From: "resource_attribute",
					Name: "ip",
				},
				{
					From: "resource_attribute",
					Name: "k8s.pod.ip",
				},
				{
					From: "resource_attribute",
					Name: "host.name",
				},
				{
					From: "connection",
					Name: "ip",
				},
				{
					From: "resource_attribute",
					Name: "k8s.pod.uid",
				},
			},
			Exclude: ExcludeConfig{
				Pods: []ExcludePodConfig{
					{Name: "jaeger-agent"},
					{Name: "jaeger-collector"},
				},
			},
		},
		p1,
	)
}
