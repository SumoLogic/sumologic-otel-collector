// Copyright 2022 Sumo Logic, Inc.
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

package sumologicschemaprocessor

import (
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
)

// translateAttributesProcessor translates attribute names from OpenTelemetry to Sumo Logic convention
type translateAttributesProcessor struct {
	shouldTranslate bool
}

// attributeTranslations maps OpenTelemetry attribute names to Sumo Logic attribute names
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

func newTranslateAttributesProcessor(shouldTranslate bool) (*translateAttributesProcessor, error) {
	return &translateAttributesProcessor{
		shouldTranslate: shouldTranslate,
	}, nil
}

func (proc *translateAttributesProcessor) processLogs(logs plog.Logs) error {
	if proc.shouldTranslate {
		for i := 0; i < logs.ResourceLogs().Len(); i++ {
			translateAttributes(logs.ResourceLogs().At(i).Resource().Attributes())
		}
	}

	return nil
}

func (proc *translateAttributesProcessor) processMetrics(metrics pmetric.Metrics) error {
	if proc.shouldTranslate {
		for i := 0; i < metrics.ResourceMetrics().Len(); i++ {
			translateAttributes(metrics.ResourceMetrics().At(i).Resource().Attributes())
		}
	}

	return nil
}

func (proc *translateAttributesProcessor) processTraces(traces ptrace.Traces) error {
	// No-op. Traces should not be translated.
	return nil
}

func (proc *translateAttributesProcessor) isEnabled() bool {
	return proc.shouldTranslate
}

func (*translateAttributesProcessor) ConfigPropertyName() string {
	return "translate_attributes"
}

func translateAttributes(attributes pcommon.Map) {
	result := pcommon.NewMap()
	result.EnsureCapacity(attributes.Len())

	attributes.Range(func(otKey string, value pcommon.Value) bool {
		if sumoKey, ok := attributeTranslations[otKey]; ok {
			// Only insert if it doesn't exist yet to prevent overwriting.
			// We have to do it this way since the final return value is not
			// ready yet to rely on .Insert() not overwriting.
			if _, exists := attributes.Get(sumoKey); !exists {
				if _, ok := result.Get(sumoKey); !ok {
					value.CopyTo(result.PutEmpty(sumoKey))
				}
			} else {
				if _, ok := result.Get(otKey); !ok {
					value.CopyTo(result.PutEmpty(otKey))
				}
			}
		} else {
			if _, ok := result.Get(otKey); !ok {
				value.CopyTo(result.PutEmpty(otKey))
			}
		}
		return true
	})

	result.CopyTo(attributes)
}
