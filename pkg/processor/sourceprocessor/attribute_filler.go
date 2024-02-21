// Copyright 2021 OpenTelemetry Authors
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
	"fmt"
	"regexp"
	"strings"

	"go.opentelemetry.io/collector/pdata/pcommon"
)

var (
	formatRegex *regexp.Regexp
)

func init() {
	var err error
	formatRegex, err = regexp.Compile(`\%\{([\w\.\/]+)\}`)
	if err != nil {
		panic("failed to parse regex: " + err.Error())
	}
}

type attributeFiller struct {
	name            string
	compiledFormat  string
	dashReplacement string
	prefix          string
	labels          []string
}

func extractFormat(format string, name string) attributeFiller {
	labels := make([]string, 0)
	matches := formatRegex.FindAllStringSubmatch(format, -1)
	for _, matchset := range matches {
		labels = append(labels, matchset[1])
	}
	template := formatRegex.ReplaceAllString(format, "%s")

	return attributeFiller{
		name:            name,
		compiledFormat:  template,
		dashReplacement: "",
		labels:          labels,
		prefix:          "",
	}
}

func createSourceHostFiller(cfg *Config) attributeFiller {
	filler := extractFormat(cfg.SourceHost, sourceHostKey)
	return filler
}

func createSourceNameFiller(cfg *Config) attributeFiller {
	filler := extractFormat(cfg.SourceName, sourceNameKey)
	return filler
}

func (f *attributeFiller) fillResourceOrUseAnnotation(atts *pcommon.Map, annotationKey string, namespaceAnnotationKey string) bool {
	val, found := atts.Get(annotationKey)
	if found {
		return f.useAnnotation(atts, val)
	}

	val, found = atts.Get(namespaceAnnotationKey)
	if found {
		return f.useAnnotation(atts, val)
	}

	return f.fillAttributes(atts)
}

func (f *attributeFiller) useAnnotation(atts *pcommon.Map, annotation pcommon.Value) bool {
	annotationFiller := extractFormat(annotation.Str(), f.name)
	annotationFiller.dashReplacement = f.dashReplacement
	annotationFiller.compiledFormat = f.prefix + annotationFiller.compiledFormat
	return annotationFiller.fillAttributes(atts)
}

func (f *attributeFiller) fillAttributes(atts *pcommon.Map) bool {
	if len(f.compiledFormat) == 0 {
		return false
	}

	labelValues := f.resourceLabelValues(atts)
	if labelValues != nil {
		str := fmt.Sprintf(f.compiledFormat, labelValues...)
		if f.dashReplacement != "" {
			str = strings.ReplaceAll(str, "-", f.dashReplacement)
		}
		atts.PutStr(f.name, str)
		return true
	}
	return false
}

func (f *attributeFiller) resourceLabelValues(atts *pcommon.Map) []interface{} {
	arr := make([]interface{}, 0)
	for _, label := range f.labels {
		if value, found := atts.Get(label); found {
			arr = append(arr, value.Str())
		} else {
			arr = append(arr, "undefined")
		}
	}
	return arr
}
