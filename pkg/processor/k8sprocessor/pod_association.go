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
	"errors"
	"fmt"
	"net"
	"strings"

	"go.opentelemetry.io/collector/client"
	"go.opentelemetry.io/collector/pdata/pcommon"
	conventions "go.opentelemetry.io/collector/semconv/v1.18.0"

	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/k8sprocessor/kube"
)

// extractPodIds extracts IP and pod UID from attributes or request context.
// It returns a value pair containing configured label and IP Address and/or Pod UID.
// If empty value in return it means that attributes does not contains configured label to match resources for Pod.
func extractPodID(
	ctx context.Context,
	attrs pcommon.Map,
	associations []kube.Association,
) (podIdentifierKey string, podIdentifierValue kube.PodIdentifier, returnErr error) {
	connectionIP := getConnectionIP(ctx)
	hostname := stringAttributeFromMap(attrs, conventions.AttributeHostName)

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

		return "", kube.PodIdentifier(""), errors.New("pod association not set, could not assign other pod id")
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
			pod, ok := attrs.Get(conventions.AttributeK8SPodName)
			if !ok {
				return "", kube.PodIdentifier(""), errors.New("pod name not found in attributes")
			}

			namespace, ok := attrs.Get(conventions.AttributeK8SNamespaceName)
			if !ok {
				return "", kube.PodIdentifier(""), errors.New("namespace name not found in attributes")
			}

			if pod.Str() == "" || namespace.Str() != "" {
				podIdentifierKey = asso.Name
				podIdentifierValue = kube.PodIdentifier(fmt.Sprintf("%s.%s", pod.Str(), namespace.Str()))
				return
			}
		}
	}

	return "", kube.PodIdentifier(""), errors.New("could not assign pod id basing on associations")
}

func getConnectionIP(ctx context.Context) kube.PodIdentifier {
	c := client.FromContext(ctx)
	if c.Addr == nil {
		return ""
	}
	switch addr := c.Addr.(type) {
	case *net.UDPAddr:
		return kube.PodIdentifier(addr.IP.String())
	case *net.TCPAddr:
		return kube.PodIdentifier(addr.IP.String())
	case *net.IPAddr:
		return kube.PodIdentifier(addr.IP.String())
	}

	// If this is not a known address type, check for known "untyped" formats.
	// 1.1.1.1:<port>

	lastColonIndex := strings.LastIndex(c.Addr.String(), ":")
	if lastColonIndex != -1 {
		ipString := c.Addr.String()[:lastColonIndex]
		ip := net.ParseIP(ipString)
		if ip != nil {
			return kube.PodIdentifier(ip.String())
		}
	}

	return kube.PodIdentifier(c.Addr.String())
}

func stringAttributeFromMap(attrs pcommon.Map, key string) string {
	if val, ok := attrs.Get(key); ok {
		if val.Type() == pcommon.ValueTypeStr {
			return val.Str()
		}
	}
	return ""
}
