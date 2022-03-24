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
	"go.opentelemetry.io/collector/model/pdata"
	conventions "go.opentelemetry.io/collector/model/semconv/v1.6.1"
)

type cloudNamespaceProcessor struct {
	addCloudNamespace bool
}

const (
	cloudNamespaceAttributeName = "cloud.namespace"
	cloudNamespaceAwsEc2        = "aws/ec2"
	cloudNamespaceAwsEcs        = "ecs"
	cloudNamespaceAwsBeanstalk  = "ElasticBeanstalk"
)

func newCloudNamespaceProcessor(addCloudNamespace bool) (*cloudNamespaceProcessor, error) {
	return &cloudNamespaceProcessor{
		addCloudNamespace: addCloudNamespace,
	}, nil
}

func (*cloudNamespaceProcessor) processLogs(logs pdata.Logs) (pdata.Logs, error) {
	for i := 0; i < logs.ResourceLogs().Len(); i++ {
		addCloudNamespaceAttribute(logs.ResourceLogs().At(i).Resource().Attributes())
	}
	return logs, nil
}

func (*cloudNamespaceProcessor) processMetrics(metrics pdata.Metrics) (pdata.Metrics, error) {
	for i := 0; i < metrics.ResourceMetrics().Len(); i++ {
		addCloudNamespaceAttribute(metrics.ResourceMetrics().At(i).Resource().Attributes())
	}
	return metrics, nil
}

func (*cloudNamespaceProcessor) processTraces(traces pdata.Traces) (pdata.Traces, error) {
	for i := 0; i < traces.ResourceSpans().Len(); i++ {
		addCloudNamespaceAttribute(traces.ResourceSpans().At(i).Resource().Attributes())
	}
	return traces, nil
}

func addCloudNamespaceAttribute(attributes pdata.AttributeMap) {
	cloudPlatformAttributeValue, found := attributes.Get(conventions.AttributeCloudPlatform)
	if !found {
		return
	}

	switch cloudPlatformAttributeValue.StringVal() {
	case conventions.AttributeCloudPlatformAWSEC2:
		attributes.InsertString(cloudNamespaceAttributeName, cloudNamespaceAwsEc2)
	case conventions.AttributeCloudPlatformAWSECS:
		attributes.InsertString(cloudNamespaceAttributeName, cloudNamespaceAwsEcs)
	case conventions.AttributeCloudPlatformAWSElasticBeanstalk:
		attributes.InsertString(cloudNamespaceAttributeName, cloudNamespaceAwsBeanstalk)
	}
}
