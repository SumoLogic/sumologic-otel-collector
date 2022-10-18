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
	"fmt"
	"strings"

	"go.opentelemetry.io/collector/pdata/pcommon"
)

// sourceCategoryFiller adds source category attribute to a collection of attributes.
type sourceCategoryFiller struct {
	valueTemplate                string
	templateAttributes           []string
	prefix                       string
	dashReplacement              string
	annotationPrefix             string
	containerAnnotationsEnabled  bool
	containerAnnotationsPrefixes []string
}

// newSourceCategoryFiller creates a new sourceCategoryFiller.
func newSourceCategoryFiller(cfg *Config) sourceCategoryFiller {

	valueTemplate := cfg.SourceCategory
	templateAttributes := extractTemplateAttributes(valueTemplate)

	return sourceCategoryFiller{
		valueTemplate:                valueTemplate,
		templateAttributes:           templateAttributes,
		prefix:                       cfg.SourceCategoryPrefix,
		dashReplacement:              cfg.SourceCategoryReplaceDash,
		annotationPrefix:             cfg.AnnotationPrefix,
		containerAnnotationsEnabled:  cfg.ContainerAnnotations.Enabled,
		containerAnnotationsPrefixes: cfg.ContainerAnnotations.Prefixes,
	}
}

func extractTemplateAttributes(template string) []string {
	attributeMatches := formatRegex.FindAllStringSubmatch(template, -1)
	attributes := make([]string, 0, len(attributeMatches))
	for _, matchset := range attributeMatches {
		attributes = append(attributes, matchset[1])
	}
	return attributes
}

// fill takes a collection of attributes for a record and adds to it a new attribute with the source category for the record.
//
// The source category is retrieved from one of three places (in the following precedence):
// - the source category container-level annotation (e.g. "k8s.pod.annotation.sumologic.com/container-name.sourceCategory"),
// - the source category pod-level annotation (e.g. "k8s.pod.annotation.sumologic.com/sourceCategory"),
// - the source category configured in the processor's "source_category" configuration option.
func (f *sourceCategoryFiller) fill(attributes *pcommon.Map) {
	containerSourceCategory := f.getSourceCategoryFromContainerAnnotation(attributes)
	if containerSourceCategory != "" {
		attributes.PutString(sourceCategoryKey, containerSourceCategory)
		return
	}

	valueTemplate := getAnnotationAttributeValue(f.annotationPrefix, sourceCategorySpecialAnnotation, attributes)
	var templateAttributes []string
	if valueTemplate != "" {
		templateAttributes = extractTemplateAttributes(valueTemplate)
	} else {
		valueTemplate = f.valueTemplate
		templateAttributes = f.templateAttributes
	}
	sourceCategoryValue := f.replaceTemplateAttributes(valueTemplate, templateAttributes, attributes)

	prefix := getAnnotationAttributeValue(f.annotationPrefix, sourceCategoryPrefixAnnotation, attributes)
	if prefix == "" {
		prefix = f.prefix
	}
	sourceCategoryValue = prefix + sourceCategoryValue

	dashReplacement := getAnnotationAttributeValue(f.annotationPrefix, sourceCategoryReplaceDashAnnotation, attributes)
	if dashReplacement == "" {
		dashReplacement = f.dashReplacement
	}
	sourceCategoryValue = strings.ReplaceAll(sourceCategoryValue, "-", dashReplacement)

	attributes.PutString(sourceCategoryKey, sourceCategoryValue)
}

func (f *sourceCategoryFiller) getSourceCategoryFromContainerAnnotation(attributes *pcommon.Map) string {
	if !f.containerAnnotationsEnabled {
		return ""
	}

	containerName, found := attributes.Get("k8s.container.name")
	if !found || containerName.Str() == "" {
		return ""
	}

	for _, containerAnnotationPrefix := range f.containerAnnotationsPrefixes {
		annotationKey := fmt.Sprintf("%s%s.sourceCategory", containerAnnotationPrefix, containerName.Str())
		annotationValue := getAnnotationAttributeValue(f.annotationPrefix, annotationKey, attributes)
		if annotationValue != "" {
			return annotationValue
		}
	}
	return ""
}

func (f *sourceCategoryFiller) replaceTemplateAttributes(template string, templateAttributes []string, attributes *pcommon.Map) string {
	replacerArgs := make([]string, len(templateAttributes)*2)
	for i, templateAttribute := range templateAttributes {
		attributeValue, found := attributes.Get(templateAttribute)
		var attributeValueString string
		if found {
			attributeValueString = attributeValue.Str()
		} else {
			attributeValueString = "undefined"
		}
		replacerArgs[i*2] = fmt.Sprintf("%%{%s}", templateAttribute)
		replacerArgs[i*2+1] = attributeValueString
	}

	return strings.NewReplacer(replacerArgs...).Replace(template)
}

func getAnnotationAttributeValue(annotationAttributePrefix string, annotation string, attributes *pcommon.Map) string {
	annotationAttribute, found := attributes.Get(annotationAttributePrefix + annotation)
	if found {
		return annotationAttribute.Str()
	}
	return ""
}
