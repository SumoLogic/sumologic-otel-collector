// Copyright 2021 Sumo Logic, Inc.
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

package sourceprocessor

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.uber.org/zap"
)

func TestNewSourceCategoryFiller(t *testing.T) {
	cfg := createDefaultConfig().(*Config)
	cfg.SourceCategory = "qwerty-%{k8s.namespace.name}-%{k8s.pod.uid}"

	filler := newSourceCategoryFiller(cfg, zap.NewNop())

	assert.Len(t, filler.templateAttributes, 2)
	assert.Equal(t, "k8s.namespace.name", filler.templateAttributes[0])
	assert.Equal(t, "k8s.pod.uid", filler.templateAttributes[1])
}

func TestFill(t *testing.T) {
	cfg := createDefaultConfig().(*Config)
	cfg.SourceCategory = "source-%{k8s.namespace.name}-%{k8s.pod.uid}-cat"
	cfg.SourceCategoryPrefix = "%{k8s.pod.name}/prefix/"

	attrs := pcommon.NewMap()
	attrs.PutStr("k8s.namespace.name", "ns-1")
	attrs.PutStr("k8s.pod.uid", "123asd")
	attrs.PutStr("k8s.pod.name", "ABCD")

	filler := newSourceCategoryFiller(cfg, zap.NewNop())
	filler.fill(&attrs)

	assertAttribute(t, attrs, "_sourceCategory", "ABCD/prefix/source/ns/1/123asd/cat")
}

func TestFillWithAnnotations(t *testing.T) {
	cfg := createDefaultConfig().(*Config)

	attrs := pcommon.NewMap()
	attrs.PutStr("k8s.namespace.name", "ns-1")
	attrs.PutStr("k8s.pod.uid", "123asd")
	attrs.PutStr("k8s.pod.name", "ABC")
	attrs.PutStr("k8s.pod.annotation.sumologic.com/sourceCategory", "sc-from-annot-%{k8s.namespace.name}-%{k8s.pod.uid}")
	attrs.PutStr("k8s.pod.annotation.sumologic.com/sourceCategoryPrefix", "%{k8s.pod.name}-%{k8s.pod.uid}-Prefix:")
	attrs.PutStr("k8s.pod.annotation.sumologic.com/sourceCategoryReplaceDash", "#")

	filler := newSourceCategoryFiller(cfg, zap.NewNop())
	filler.fill(&attrs)

	assertAttribute(t, attrs, "_sourceCategory", "ABC#123asd#Prefix:sc#from#annot#ns#1#123asd")
}

func TestFillWithEmptyAnnotations(t *testing.T) {
	cfg := createDefaultConfig().(*Config)
	cfg.SourceCategoryPrefix = "prefix"

	attrs := pcommon.NewMap()
	attrs.PutStr("k8s.namespace.name", "ns-1")
	attrs.PutStr("k8s.pod.uid", "123asd")
	attrs.PutStr("k8s.pod.pod_name", "ABC")

	filler := newSourceCategoryFiller(cfg, zap.NewNop())

	// can replace prefix with an empty string
	attrs.PutStr("k8s.pod.annotation.sumologic.com/sourceCategoryPrefix", "")
	filler.fill(&attrs)
	assertAttribute(t, attrs, "_sourceCategory", "ns/1/ABC")

	// can replace dash with an empty string
	attrs.PutStr("k8s.pod.annotation.sumologic.com/sourceCategoryReplaceDash", "")
	filler.fill(&attrs)
	assertAttribute(t, attrs, "_sourceCategory", "ns1/ABC")

	// can replace source category with empty string
	attrs.PutStr("k8s.pod.annotation.sumologic.com/sourceCategory", "")
	filler.fill(&attrs)
	assertAttribute(t, attrs, "_sourceCategory", "")
}

func TestFillWithNamespaceAnnotations(t *testing.T) {
	cfg := createDefaultConfig().(*Config)

	attrs := pcommon.NewMap()
	attrs.PutStr("k8s.namespace.name", "ns-1")
	attrs.PutStr("k8s.pod.uid", "123asd")
	attrs.PutStr("k8s.pod.name", "ABC")
	attrs.PutStr("k8s.namespace.annotation.sumologic.com/sourceCategory", "sc-from-annot-%{k8s.namespace.name}-%{k8s.pod.uid}")
	attrs.PutStr("k8s.namespace.annotation.sumologic.com/sourceCategoryPrefix", "%{k8s.pod.name}-%{k8s.pod.uid}-Prefix:")
	attrs.PutStr("k8s.namespace.annotation.sumologic.com/sourceCategoryReplaceDash", "#")

	filler := newSourceCategoryFiller(cfg, zap.NewNop())
	filler.fill(&attrs)

	assertAttribute(t, attrs, "_sourceCategory", "ABC#123asd#Prefix:sc#from#annot#ns#1#123asd")
}

func TestFillWithContainerAnnotations(t *testing.T) {
	t.Run("container annotations are disabled by default", func(t *testing.T) {
		cfg := createDefaultConfig().(*Config)
		cfg.SourceCategory = "my-source-category"

		attrs := pcommon.NewMap()
		attrs.PutStr("k8s.pod.annotation.sumologic.com/container-name-1.sourceCategory", "first_source-category")
		attrs.PutStr("k8s.pod.annotation.sumologic.com/container-name-2.sourceCategory", "another/source-category")
		attrs.PutStr("k8s.container.name", "container-name-1")

		filler := newSourceCategoryFiller(cfg, zap.NewNop())
		filler.fill(&attrs)

		assertAttribute(t, attrs, "_sourceCategory", "kubernetes/my/source/category")
	})

	t.Run("source category for container-name-1", func(t *testing.T) {
		cfg := createDefaultConfig().(*Config)
		cfg.SourceCategory = "my-source-category"
		cfg.ContainerAnnotations.Enabled = true

		attrs := pcommon.NewMap()
		attrs.PutStr("k8s.pod.annotation.sumologic.com/container-name-1.sourceCategory", "first_source-category")
		attrs.PutStr("k8s.pod.annotation.sumologic.com/container-name-2.sourceCategory", "another/source-category")
		attrs.PutStr("k8s.container.name", "container-name-1")

		filler := newSourceCategoryFiller(cfg, zap.NewNop())
		filler.fill(&attrs)

		assertAttribute(t, attrs, "_sourceCategory", "first_source-category")
	})

	t.Run("source category for container-name-2", func(t *testing.T) {
		cfg := createDefaultConfig().(*Config)
		cfg.SourceCategory = "my-source-category"
		cfg.ContainerAnnotations.Enabled = true

		attrs := pcommon.NewMap()
		attrs.PutStr("k8s.pod.annotation.sumologic.com/container-name-1.sourceCategory", "first_source-category")
		attrs.PutStr("k8s.pod.annotation.sumologic.com/container-name-2.sourceCategory", "another/source-category")
		attrs.PutStr("k8s.container.name", "container-name-2")

		filler := newSourceCategoryFiller(cfg, zap.NewNop())
		filler.fill(&attrs)

		assertAttribute(t, attrs, "_sourceCategory", "another/source-category")
	})

	t.Run("custom container annotation prefix", func(t *testing.T) {
		cfg := createDefaultConfig().(*Config)
		cfg.SourceCategory = "my-source-category"
		cfg.ContainerAnnotations.Enabled = true
		cfg.ContainerAnnotations.Prefixes = []string{
			"unused-prefix/",
			"customAnno_prefix:",
		}

		attrs := pcommon.NewMap()
		attrs.PutStr("k8s.pod.annotation.customAnno_prefix:container-name-1.sourceCategory", "first_source-category")
		attrs.PutStr("k8s.pod.annotation.customAnno_prefix:container-name-2.sourceCategory", "another/source-category")
		attrs.PutStr("k8s.pod.annotation.customAnno_prefix:container-name-3.sourceCategory", "THIRD_s-c!")
		attrs.PutStr("k8s.container.name", "container-name-3")

		filler := newSourceCategoryFiller(cfg, zap.NewNop())
		filler.fill(&attrs)

		assertAttribute(t, attrs, "_sourceCategory", "THIRD_s-c!")
	})

	t.Run("set source category prefix", func(t *testing.T) {
		cfg := createDefaultConfig().(*Config)
		cfg.SourceCategory = "my-source-category"
		cfg.SourceCategoryPrefix = "%{k8s.container.name}/"

		attrs := pcommon.NewMap()
		attrs.PutStr("k8s.pod.annotation.sumologic.com/container-name-1.sourceCategory", "first_source-category")
		attrs.PutStr("k8s.pod.annotation.sumologic.com/container-name-2.sourceCategory", "another/source-category")
		attrs.PutStr("k8s.container.name", "container-name-1")

		filler := newSourceCategoryFiller(cfg, zap.NewNop())
		filler.fill(&attrs)

		assertAttribute(t, attrs, "_sourceCategory", "container/name/1/my/source/category")
	})

	t.Run("container annotations do not support templating", func(t *testing.T) {
		cfg := createDefaultConfig().(*Config)
		cfg.ContainerAnnotations.Enabled = true
		cfg.SourceCategory = "my-source-category"
		cfg.SourceCategoryPrefix = "my-prefix"

		attrs := pcommon.NewMap()
		attrs.PutStr("k8s.container.name", "container-name-1")
		attrs.PutStr("k8s.namespace.name", "ns-1")
		attrs.PutStr("k8s.pod.name", "ABC")
		attrs.PutStr("k8s.pod.annotation.sumologic.com/sourceCategory", "sc-from-annot-%{k8s.namespace.name}-%{k8s.pod.uid}")
		attrs.PutStr("k8s.pod.annotation.sumologic.com/sourceCategoryPrefix", "%{k8s.pod.name}-%{k8s.pod.uid}-Prefix:")
		attrs.PutStr("k8s.pod.annotation.sumologic.com/sourceCategoryReplaceDash", "#")
		attrs.PutStr("k8s.pod.annotation.sumologic.com/container-name-1.sourceCategory", "%{k8s.namespace.name}-%{k8s.pod.name}")
		attrs.PutStr("k8s.pod.annotation.sumologic.com/container-name-2.sourceCategory", "another/source-category")
		attrs.PutStr("k8s.container.name", "container-name-1")

		filler := newSourceCategoryFiller(cfg, zap.NewNop())
		filler.fill(&attrs)

		assertAttribute(t, attrs, "_sourceCategory", "%{k8s.namespace.name}-%{k8s.pod.name}")
	})
}
