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
	"log.file.path_resolved":  "_sourceName",
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
