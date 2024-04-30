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
	"go.uber.org/zap"
)

// sourceCategoryFiller adds source category attribute to a collection of attributes.
type sourceCategoryFiller struct {
	logger                       *zap.Logger
	valueTemplate                string
	templateAttributes           []string
	prefix                       string
	dashReplacement              string
	annotationPrefix             string
	namespaceAnnotationPrefix    string
	containerAnnotationsEnabled  bool
	containerNameKey             string
	containerAnnotationsPrefixes []string
}

// newSourceCategoryFiller creates a new sourceCategoryFiller.
func newSourceCategoryFiller(cfg *Config, logger *zap.Logger) sourceCategoryFiller {

	templateAttributes := extractTemplateAttributes(cfg.SourceCategoryPrefix + cfg.SourceCategory)

	return sourceCategoryFiller{
		logger:                       logger,
		valueTemplate:                cfg.SourceCategory,
		templateAttributes:           templateAttributes,
		prefix:                       cfg.SourceCategoryPrefix,
		dashReplacement:              cfg.SourceCategoryReplaceDash,
		annotationPrefix:             cfg.AnnotationPrefix,
		namespaceAnnotationPrefix:    cfg.NamespaceAnnotationPrefix,
		containerAnnotationsEnabled:  cfg.ContainerAnnotations.Enabled,
		containerNameKey:             cfg.ContainerAnnotations.ContainerNameKey,
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
// The source category is retrieved from one of the following locations, listed in descending order of precedence:
// - the source category container-level annotation (e.g. "k8s.pod.annotation.sumologic.com/container-name.sourceCategory"),
// - the source category pod-level annotation (e.g. "k8s.pod.annotation.sumologic.com/sourceCategory"),
// - the source category namespace-level annotation (e.g. "k8s.namespace.annotation.sumologic.com/sourceCategory"),
// - the source category configured in the processor's "source_category" configuration option.
func (f *sourceCategoryFiller) fill(attributes *pcommon.Map) {
	containerSourceCategory := f.getSourceCategoryFromContainerAnnotation(attributes)
	if containerSourceCategory != "" {
		attributes.PutStr(sourceCategoryKey, containerSourceCategory)
		return
	}

	var templateAttributes []string
	doesUseAnnotation := false

	// get sourceCategory and sourceCategoryPrefix from pod annotation
	valueTemplate, doesUseAnnotation := f.getSourceCategoryFromAnnotation(f.annotationPrefix, attributes)

	if !doesUseAnnotation {
		// get sourceCategory and sourceCategoryPrefix from namespace annotation
		valueTemplate, doesUseAnnotation = f.getSourceCategoryFromAnnotation(f.namespaceAnnotationPrefix, attributes)
	}

	if doesUseAnnotation {
		templateAttributes = extractTemplateAttributes(valueTemplate)
	} else {
		templateAttributes = f.templateAttributes
	}

	sourceCategoryValue := f.replaceTemplateAttributes(valueTemplate, templateAttributes, attributes)

	dashReplacement := f.getSourceCategoryDashReplacement(attributes)
	sourceCategoryValue = strings.ReplaceAll(sourceCategoryValue, "-", dashReplacement)

	attributes.PutStr(sourceCategoryKey, sourceCategoryValue)
}

func (f *sourceCategoryFiller) getSourceCategoryDashReplacement(attributes *pcommon.Map) string {
	dashReplacement, found := getAnnotationAttributeValue(f.annotationPrefix, sourceCategoryReplaceDashAnnotation, attributes)
	if found {
		return dashReplacement
	}

	dashReplacement, found = getAnnotationAttributeValue(f.namespaceAnnotationPrefix, sourceCategoryReplaceDashAnnotation, attributes)
	if found {
		return dashReplacement
	}
	return f.dashReplacement
}

func (f *sourceCategoryFiller) getSourceCategoryFromAnnotation(annotationPrefix string, attributes *pcommon.Map) (string, bool) {
	valueTemplate, foundTemplate := getAnnotationAttributeValue(annotationPrefix, sourceCategorySpecialAnnotation, attributes)
	if !foundTemplate {
		valueTemplate = f.valueTemplate
	}

	prefix, foundPrefix := getAnnotationAttributeValue(annotationPrefix, sourceCategoryPrefixAnnotation, attributes)
	if !foundPrefix {
		prefix = f.prefix
	}

	valueTemplate = prefix + valueTemplate
	doesUseAnnotation := foundPrefix || foundTemplate
	return valueTemplate, doesUseAnnotation
}

func (f *sourceCategoryFiller) getSourceCategoryFromContainerAnnotation(attributes *pcommon.Map) string {
	if !f.containerAnnotationsEnabled {
		return ""
	}

	containerName, found := attributes.Get(f.containerNameKey)
	if !found || containerName.Str() == "" {
		f.logger.Debug("Couldn't fill source category from container annotation: container name attribute not found.",
			zap.String("container_name_key", f.containerNameKey))
		return ""
	}

	for _, containerAnnotationPrefix := range f.containerAnnotationsPrefixes {
		annotationKey := fmt.Sprintf("%s%s.sourceCategory", containerAnnotationPrefix, containerName.Str())
		annotationValue, found := getAnnotationAttributeValue(f.annotationPrefix, annotationKey, attributes)
		if found {
			f.logger.Debug("Filled source category from container annotation",
				zap.String("annotation", annotationKey),
				zap.String("source_category", annotationValue),
				zap.String("container", containerName.Str()))
			return annotationValue
		}
	}

	f.logger.Debug("Couldn't fill source category from container annotation: no matching annotation found for container.", zap.String("container", containerName.Str()))
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

func getAnnotationAttributeValue(annotationAttributePrefix string, annotation string, attributes *pcommon.Map) (string, bool) {
	annotationAttribute, found := attributes.Get(annotationAttributePrefix + annotation)
	if found {
		return annotationAttribute.Str(), found
	}
	return "", false
}
