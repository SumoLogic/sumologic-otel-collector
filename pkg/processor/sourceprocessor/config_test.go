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
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/config/configtest"
)

func TestLoadConfig(t *testing.T) {
	factories, err := componenttest.NopFactories()
	assert.NoError(t, err)

	factory := NewFactory()
	factories.Processors[typeStr] = factory

	cfgPath := path.Join(".", "testdata", "config.yaml")
	cfg, err := configtest.LoadConfig(cfgPath, factories)
	assert.NoError(t, err)
	assert.NotNil(t, cfg)

	id1 := config.NewID("source")
	p1 := cfg.Processors[id1]
	assert.Equal(t, p1, factory.CreateDefaultConfig())

	id2 := config.NewIDWithName("source", "2")
	p2, ok := cfg.Processors[id2]
	assert.True(t, ok)

	ps2 := config.NewProcessorSettings(id2)
	assert.Equal(t, p2, &Config{
		ProcessorSettings:         &ps2,
		Collector:                 "somecollector",
		SourceName:                "%{namespace}.%{pod}.%{container}/foo",
		SourceCategory:            "%{namespace}/%{pod_name}/bar",
		SourceCategoryPrefix:      "kubernetes/",
		SourceCategoryReplaceDash: "/",
		Exclude: map[string]string{
			"container":     "excluded_container_regex",
			"host":          "excluded_host_regex",
			"namespace":     "excluded_namespace_regex",
			"pod":           "excluded_pod_regex",
			"_SYSTEMD_UNIT": "excluded_systemd_unit_regex",
		},

		AnnotationPrefix:   "pod_annotation_",
		ContainerKey:       "container",
		NamespaceKey:       "namespace",
		PodKey:             "pod",
		PodIDKey:           "k8s.pod.uid",
		PodNameKey:         "pod_name",
		PodTemplateHashKey: "pod_labels_pod-template-hash",
		SourceHostKey:      "source_host",
	})
}
