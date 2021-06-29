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

	"go.opentelemetry.io/collector/consumer/pdata"
)

// metadataTranslations maps OpenTelemetry attribute names to Sumo attribute names
var metadataTranslations = map[string]string{
	"cloud.account.id":        "accountId",
	"cloud.availability_zone": "availabilityZone",
	"cloud.platform":          "aws_service",
	"cloud.region":            "region",
	"host.id":                 "instanceId",
	"host.name":               "host",
	"host.type":               "instanceType",
	"k8s.cluster.name":        "cluster",
	"k8s.container.name":      "container",
	"k8s.daemonset.name":      "daemonset",
	"k8s.deployment.name":     "deployment",
	"k8s.namespace.name":      "namespace",
	"k8s.node.name":           "node",
	"k8s.pod.hostname":        "host",
	"k8s.pod.name":            "pod",
	"k8s.pod.uid":             "pod_id",
	"k8s.replicaset.name":     "replicaset",
	"k8s.statefulset.name":    "statefulset",
	"service.name":            "service",
	"file.path.resolved":      "_sourceName",
}

// translateMetadata renames metadata keys according to metadataTranslations.
func translateMetadata(attributes pdata.AttributeMap) {
	attributes.Range(func(otKey string, value pdata.AttributeValue) bool {
		if sumoKey, ok := metadataTranslations[otKey]; ok {
			attributes.Upsert(sumoKey, value)
			attributes.Delete(otKey)
		}
		return true
	})
}

// translateConfigValue renames metadata keys in config values according to metadataTranslations.
func translateConfigValue(value string) string {
	for _, sumoKey := range metadataTranslations {
		value = strings.ReplaceAll(value, fmt.Sprintf("%%{%v}", sumoKey), "")
	}
	for otKey, sumoKey := range metadataTranslations {
		value = strings.ReplaceAll(value, fmt.Sprintf("%%{%v}", otKey), fmt.Sprintf("%%{%v}", sumoKey))
	}
	return value
}
