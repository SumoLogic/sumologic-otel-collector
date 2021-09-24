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
	"go.opentelemetry.io/collector/model/pdata"
)

func TestNewSourceCategoryFiller(t *testing.T) {
	cfg := createDefaultConfig().(*Config)
	cfg.SourceCategory = "qwerty-%{k8s.namespace.name}-%{k8s.pod.uid}"

	filler := newSourceCategoryFiller("_sourceCategory", cfg)

	assert.Equal(t, filler.attributeName, "_sourceCategory")
	assert.Equal(t, 2, len(filler.templateAttributes))
	assert.Equal(t, "k8s.namespace.name", filler.templateAttributes[0])
	assert.Equal(t, "k8s.pod.uid", filler.templateAttributes[1])
}

func TestFill(t *testing.T) {
	cfg := createDefaultConfig().(*Config)
	cfg.SourceCategory = "source-%{k8s.namespace.name}-%{k8s.pod.uid}-cat"

	attrs := pdata.NewAttributeMap()
	attrs.InsertString("k8s.namespace.name", "ns-1")
	attrs.InsertString("k8s.pod.uid", "123asd")

	filler := newSourceCategoryFiller("_sourceCategory", cfg)
	filler.fill(&attrs)

	assertAttribute(t, attrs, "_sourceCategory", "kubernetes/source/ns/1/123asd/cat")
}

func TestFillWithAnnotations(t *testing.T) {
	cfg := createDefaultConfig().(*Config)

	attrs := pdata.NewAttributeMap()
	attrs.InsertString("k8s.namespace.name", "ns-1")
	attrs.InsertString("k8s.pod.uid", "123asd")
	attrs.InsertString("k8s.pod.annotation.sumologic.com/sourceCategory", "sc-from-annot-%{k8s.namespace.name}-%{k8s.pod.uid}")
	attrs.InsertString("k8s.pod.annotation.sumologic.com/sourceCategoryPrefix", "annoPrefix:")
	attrs.InsertString("k8s.pod.annotation.sumologic.com/sourceCategoryReplaceDash", "#")

	filler := newSourceCategoryFiller("_sourceCategory", cfg)
	filler.fill(&attrs)

	assertAttribute(t, attrs, "_sourceCategory", "annoPrefix:sc#from#annot#ns#1#123asd")
}
