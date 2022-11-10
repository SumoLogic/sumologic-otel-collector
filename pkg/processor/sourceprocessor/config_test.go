// Copyright 2019 OpenTelemetry Authors
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

package sourceprocessor

import (
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
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

	cfgPath := path.Join(".", "testdata", "config.yaml")
	cfg, err := servicetest.LoadConfig(cfgPath, factories)
	assert.NoError(t, err)
	assert.NotNil(t, cfg)

	id1 := component.NewID("source")
	p1 := cfg.Processors[id1]
	assert.Equal(t, p1, factory.CreateDefaultConfig())

	id2 := component.NewIDWithName("source", "2")
	p2, ok := cfg.Processors[id2]
	assert.True(t, ok)

	ps2 := config.NewProcessorSettings(id2)
	assert.Equal(t, p2, &Config{
		ProcessorSettings:         &ps2,
		Collector:                 "somecollector",
		SourceHost:                "%{k8s.pod.hostname}",
		SourceName:                "%{k8s.namespace.name}.%{k8s.pod.name}.%{k8s.container.name}/foo",
		SourceCategory:            "%{k8s.namespace.name}/%{k8s.pod.pod_name}/bar",
		SourceCategoryPrefix:      "kubernetes/",
		SourceCategoryReplaceDash: "/",
		Exclude: map[string]string{
			"k8s.container.name": "excluded_container_regex",
			"k8s.pod.hostname":   "excluded_host_regex",
			"k8s.namespace.name": "excluded_namespace_regex",
			"k8s.pod.name":       "excluded_pod_regex",
			"_SYSTEMD_UNIT":      "excluded_systemd_unit_regex",
		},

		AnnotationPrefix:   "pod_annotation_",
		PodKey:             "k8s.pod.name",
		PodNameKey:         "k8s.pod.pod_name",
		PodTemplateHashKey: "pod_labels_pod-template-hash",

		ContainerAnnotations: ContainerAnnotationsConfig{
			Enabled: false,
			Prefixes: []string{
				"sumologic.com/",
			},
		},
	})
}
