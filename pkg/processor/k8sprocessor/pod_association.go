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

package k8sprocessor

import (
	"context"
	"fmt"
	"net"

	"go.opentelemetry.io/collector/client"
	"go.opentelemetry.io/collector/model/pdata"
	"go.opentelemetry.io/collector/translator/conventions"

	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/k8sprocessor/kube"
)

// extractPodIds extracts IP and pod UID from attributes or request context.
// It returns a value pair containing configured label and IP Address and/or Pod UID.
// If empty value in return it means that attributes does not contains configured label to match resources for Pod.
func extractPodID(ctx context.Context, attrs pdata.AttributeMap, associations []kube.Association) (podIdentifierKey string, podIdentifierValue kube.PodIdentifier) {
	hostname := stringAttributeFromMap(attrs, conventions.AttributeHostName)
	var connectionIP kube.PodIdentifier
	if c, ok := client.FromContext(ctx); ok {
		connectionIP = kube.PodIdentifier(c.IP)
	}
	// If pod association is not set
	if len(associations) == 0 {
		var podIP, labelIP kube.PodIdentifier
		podIP = kube.PodIdentifier(stringAttributeFromMap(attrs, k8sIPLabelName))
		labelIP = kube.PodIdentifier(stringAttributeFromMap(attrs, clientIPLabelName))
		podIdentifierKey = k8sIPLabelName
		if podIP != "" {
			podIdentifierValue = podIP
			return
		} else if labelIP != "" {
			podIdentifierValue = labelIP
			return
		} else if connectionIP != "" {
			podIdentifierValue = connectionIP
			return
		} else if net.ParseIP(hostname) != nil {
			podIdentifierValue = kube.PodIdentifier(hostname)
			return
		}
		podIdentifierKey = ""
		return
	}

	for _, asso := range associations {
		switch {
		// If association configured to take IP address from connection
		case asso.From == "connection" && connectionIP != "":
			podIdentifierKey = k8sIPLabelName
			podIdentifierValue = connectionIP
			return
		case asso.From == "resource_attribute": // If association configured by resource_attribute
			// In k8s environment, host.name label set to a pod IP address.
			// If the value doesn't represent an IP address, we skip it.
			if asso.Name == conventions.AttributeHostName {
				if net.ParseIP(hostname) != nil {
					podIdentifierKey = k8sIPLabelName
					podIdentifierValue = kube.PodIdentifier(hostname)
					return
				}
			} else {
				// Extract values based on configured resource_attribute.
				// Value should be a pod ip, pod uid or `pod_name.namespace_name`
				attributeValue := stringAttributeFromMap(attrs, asso.Name)
				if attributeValue != "" {
					podIdentifierKey = asso.Name
					podIdentifierValue = kube.PodIdentifier(attributeValue)
					return
				}
			}
		case asso.From == "build_hostname":
			// Build hostname from pod k8s.pod.name and k8s.namespace.name attributes
			pod, ok := attrs.Get(conventions.AttributeK8sPod)
			if !ok {
				return "", ""
			}

			namespace, ok := attrs.Get(conventions.AttributeK8sNamespace)
			if !ok {
				return "", ""
			}

			if pod.StringVal() == "" || namespace.StringVal() != "" {
				return asso.Name, kube.PodIdentifier(fmt.Sprintf("%s.%s", pod.StringVal(), namespace.StringVal()))
			}
		}
	}
	return "", kube.PodIdentifier("")
}

func stringAttributeFromMap(attrs pdata.AttributeMap, key string) string {
	if val, ok := attrs.Get(key); ok {
		if val.Type() == pdata.AttributeValueTypeString {
			return val.StringVal()
		}
	}
	return ""
}
