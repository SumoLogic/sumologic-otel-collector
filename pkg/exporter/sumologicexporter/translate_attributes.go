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

package sumologicexporter

import (
	"fmt"
	"strings"

	"go.opentelemetry.io/collector/model/pdata"
)

// attributeTranslations maps OpenTelemetry attribute names to Sumo attribute names
var attributeTranslations = map[string]string{
	"cloud.account.id":        "AccountId",
	"cloud.availability_zone": "AvailabilityZone",
	"cloud.platform":          "aws_service",
	"cloud.region":            "Region",
	"host.id":                 "InstanceId",
	"host.name":               "host",
	"host.type":               "InstanceType",
	"k8s.cluster.name":        "Cluster",
	"k8s.container.name":      "container",
	"k8s.daemonset.name":      "daemonset",
	"k8s.deployment.name":     "deployment",
	"k8s.namespace.name":      "namespace",
	"k8s.node.name":           "node",
	"k8s.service.name":        "service",
	"k8s.pod.hostname":        "host",
	"k8s.pod.name":            "pod",
	"k8s.pod.uid":             "pod_id",
	"k8s.replicaset.name":     "replicaset",
	"k8s.statefulset.name":    "statefulset",
	"service.name":            "service",
	"file.path.resolved":      "_sourceName",
}

func translateAttributes(attributes pdata.AttributeMap) pdata.AttributeMap {
	ret := pdata.NewAttributeMap()
	ret.EnsureCapacity(attributes.Len())

	attributes.Range(func(otKey string, value pdata.AttributeValue) bool {
		if sumoKey, ok := attributeTranslations[otKey]; ok {
			// Only insert if it doesn't exist yet to prevent overwriting.
			// We have to do it this way since the final return value is not
			// ready yet to rely on .Insert() not overwriting.
			if _, exists := attributes.Get(sumoKey); !exists {
				ret.Insert(sumoKey, value)
			} else {
				ret.Insert(otKey, value)
			}
		} else {
			ret.Insert(otKey, value)
		}
		return true
	})

	return ret
}

// translateAttributesInPlace renames attribute keys according to attributeTranslations.
//
// DEPRECATED: Please use translateAttributes instead.
func translateAttributesInPlace(attributes pdata.AttributeMap) {
	attributes.Range(func(otKey string, value pdata.AttributeValue) bool {
		if sumoKey, ok := attributeTranslations[otKey]; ok {
			// do not rename attribute if target name already exists
			if _, ok := attributes.Get(sumoKey); ok {
				return true
			}
			attributes.Insert(sumoKey, value)
			attributes.Remove(otKey)
		}
		return true
	})
}

// translateConfigValue renames attribute keys in config values according to attributeTranslations.
// Example:
// * '%{k8s.container.name}' would translate to '%{container}'
// * '%{k8s.pod.name}-%{custom_attr}' would translate to '%{pod}-%{custom_attr}'
// * '%{pod}' would translate to '%{pod}'
func translateConfigValue(value string) string {
	for otKey, sumoKey := range attributeTranslations {
		value = strings.ReplaceAll(value, fmt.Sprintf("%%{%v}", otKey), fmt.Sprintf("%%{%v}", sumoKey))
	}
	return value
}
