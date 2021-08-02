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

package sumologicexporter

import (
	"fmt"
	"sort"
	"strings"

	"go.opentelemetry.io/collector/model/pdata"
	tracetranslator "go.opentelemetry.io/collector/translator/trace"
)

// fields represents metadata
type fields struct {
	orig     pdata.AttributeMap
	replacer *strings.Replacer
}

func newFields(attrMap pdata.AttributeMap) fields {
	return fields{
		orig:     attrMap,
		replacer: strings.NewReplacer(",", "_", "=", ":", "\n", "_"),
	}
}

// string returns fields as ordered key=value string with `, ` as separator
func (f fields) string() string {
	returnValue := make([]string, 0, f.orig.Len())
	f.orig.Range(func(k string, v pdata.AttributeValue) bool {
		returnValue = append(
			returnValue,
			fmt.Sprintf(
				"%s=%s",
				f.sanitizeField(k),
				f.sanitizeField(tracetranslator.AttributeValueToString(v)),
			),
		)
		return true
	})
	sort.Strings(returnValue)

	return strings.Join(returnValue, ", ")
}

// sanitizeFields sanitize field (key or value) to be correctly parsed by sumologic receiver
func (f fields) sanitizeField(fld string) string {
	return f.replacer.Replace(fld)
}
