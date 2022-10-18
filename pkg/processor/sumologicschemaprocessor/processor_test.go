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
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.uber.org/zap"
)

func TestAddCloudNamespaceForLogs(t *testing.T) {
	testCases := []struct {
		name              string
		addCloudNamespace bool
		createLogs        func() plog.Logs
		test              func(plog.Logs)
	}{
		{
			name:              "adds cloud.namespace attribute for EC2",
			addCloudNamespace: true,
			createLogs: func() plog.Logs {
				inputLogs := plog.NewLogs()
				inputLogs.ResourceLogs().AppendEmpty().Resource().Attributes().PutString("cloud.platform", "aws_ec2")
				return inputLogs
			},
			test: func(outputLogs plog.Logs) {
				cloudNamespaceAttribute, found := outputLogs.ResourceLogs().At(0).Resource().Attributes().Get("cloud.namespace")
				assert.True(t, found)
				assert.Equal(t, "aws/ec2", cloudNamespaceAttribute.Str())
			},
		},
		{
			name:              "adds cloud.namespace attribute for ECS",
			addCloudNamespace: true,
			createLogs: func() plog.Logs {
				inputLogs := plog.NewLogs()
				inputLogs.ResourceLogs().AppendEmpty().Resource().Attributes().PutString("cloud.platform", "aws_ecs")
				return inputLogs
			},
			test: func(outputLogs plog.Logs) {
				cloudNamespaceAttribute, found := outputLogs.ResourceLogs().At(0).Resource().Attributes().Get("cloud.namespace")
				assert.True(t, found)
				assert.Equal(t, "ecs", cloudNamespaceAttribute.Str())
			},
		},
		{
			name:              "adds cloud.namespace attribute for Beanstalk",
			addCloudNamespace: true,
			createLogs: func() plog.Logs {
				inputLogs := plog.NewLogs()
				inputLogs.ResourceLogs().AppendEmpty().Resource().Attributes().PutString("cloud.platform", "aws_elastic_beanstalk")
				return inputLogs
			},
			test: func(outputLogs plog.Logs) {
				cloudNamespaceAttribute, found := outputLogs.ResourceLogs().At(0).Resource().Attributes().Get("cloud.namespace")
				assert.True(t, found)
				assert.Equal(t, "ElasticBeanstalk", cloudNamespaceAttribute.Str())
			},
		},
		{
			name:              "does not add cloud.namespace attribute for unknown cloud.platform attribute values",
			addCloudNamespace: false,
			createLogs: func() plog.Logs {
				inputLogs := plog.NewLogs()
				inputLogs.ResourceLogs().AppendEmpty().Resource().Attributes().PutString("cloud.platform", "aws_eks")
				inputLogs.ResourceLogs().AppendEmpty().Resource().Attributes().PutString("cloud.platform", "aws_lambda")
				inputLogs.ResourceLogs().AppendEmpty().Resource().Attributes().PutString("cloud.platform", "azure_vm")
				inputLogs.ResourceLogs().AppendEmpty().Resource().Attributes().PutString("cloud.platform", "gcp_app_engine")
				return inputLogs
			},
			test: func(outputLogs plog.Logs) {
				for i := 0; i < outputLogs.ResourceLogs().Len(); i++ {
					_, found := outputLogs.ResourceLogs().At(i).Resource().Attributes().Get("cloud.namespace")
					assert.False(t, found)
				}
			},
		},
		{
			name:              "does not add cloud.namespce attribute when disabled",
			addCloudNamespace: false,
			createLogs: func() plog.Logs {
				inputLogs := plog.NewLogs()
				inputLogs.ResourceLogs().AppendEmpty().Resource().Attributes().PutString("cloud.platform", "aws_ec2")
				return inputLogs
			},
			test: func(outputLogs plog.Logs) {
				cloudNamespaceAttribute, found := outputLogs.ResourceLogs().At(0).Resource().Attributes().Get("cloud.namespace")
				assert.True(t, found)
				assert.Equal(t, "aws/ec2", cloudNamespaceAttribute.Str())
			},
		},
		{
			name:              "adds different cloud.namespace attributes to different resources",
			addCloudNamespace: true,
			createLogs: func() plog.Logs {
				inputLogs := plog.NewLogs()
				inputLogs.ResourceLogs().AppendEmpty().Resource().Attributes().PutString("cloud.platform", "aws_ec2")
				inputLogs.ResourceLogs().AppendEmpty().Resource().Attributes().PutString("cloud.platform", "aws_lambda")
				inputLogs.ResourceLogs().AppendEmpty().Resource().Attributes().PutString("cloud.platform", "aws_ecs")
				inputLogs.ResourceLogs().AppendEmpty().Resource().Attributes().PutString("cloud.platform", "aws_elastic_beanstalk")
				return inputLogs
			},
			test: func(outputLogs plog.Logs) {
				ec2ResourceAttribute, found := outputLogs.ResourceLogs().At(0).Resource().Attributes().Get("cloud.namespace")
				assert.True(t, found)
				assert.Equal(t, "aws/ec2", ec2ResourceAttribute.Str())

				_, found = outputLogs.ResourceLogs().At(1).Resource().Attributes().Get("cloud.namespace")
				assert.False(t, found)

				ecsResourceAttribute, found := outputLogs.ResourceLogs().At(2).Resource().Attributes().Get("cloud.namespace")
				assert.True(t, found)
				assert.Equal(t, "ecs", ecsResourceAttribute.Str())

				beanstalkResourceAttribute, found := outputLogs.ResourceLogs().At(3).Resource().Attributes().Get("cloud.namespace")
				assert.True(t, found)
				assert.Equal(t, "ElasticBeanstalk", beanstalkResourceAttribute.Str())
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			// Arrange
			processor, err := newSumologicSchemaProcessor(newProcessorCreateSettings(), newCloudNamespaceConfig(testCase.addCloudNamespace))
			require.NoError(t, err)

			// Act
			outputLogs, err := processor.processLogs(context.Background(), testCase.createLogs())
			require.NoError(t, err)

			// Assert
			testCase.test(outputLogs)
		})
	}
}

func TestAddCloudNamespaceForMetrics(t *testing.T) {
	testCases := []struct {
		name              string
		addCloudNamespace bool
		createMetrics     func() pmetric.Metrics
		test              func(pmetric.Metrics)
	}{
		{
			name:              "adds cloud.namespace attribute for EC2",
			addCloudNamespace: true,
			createMetrics: func() pmetric.Metrics {
				inputMetrics := pmetric.NewMetrics()
				inputMetrics.ResourceMetrics().AppendEmpty().Resource().Attributes().PutString("cloud.platform", "aws_ec2")
				return inputMetrics
			},
			test: func(outputMetrics pmetric.Metrics) {
				cloudNamespaceAttribute, found := outputMetrics.ResourceMetrics().At(0).Resource().Attributes().Get("cloud.namespace")
				assert.True(t, found)
				assert.Equal(t, "aws/ec2", cloudNamespaceAttribute.Str())
			},
		},
		{
			name:              "adds cloud.namespace attribute for ECS",
			addCloudNamespace: true,
			createMetrics: func() pmetric.Metrics {
				inputMetrics := pmetric.NewMetrics()
				inputMetrics.ResourceMetrics().AppendEmpty().Resource().Attributes().PutString("cloud.platform", "aws_ecs")
				return inputMetrics
			},
			test: func(outputMetrics pmetric.Metrics) {
				cloudNamespaceAttribute, found := outputMetrics.ResourceMetrics().At(0).Resource().Attributes().Get("cloud.namespace")
				assert.True(t, found)
				assert.Equal(t, "ecs", cloudNamespaceAttribute.Str())
			},
		},
		{
			name:              "adds cloud.namespace attribute for Beanstalk",
			addCloudNamespace: true,
			createMetrics: func() pmetric.Metrics {
				inputMetrics := pmetric.NewMetrics()
				inputMetrics.ResourceMetrics().AppendEmpty().Resource().Attributes().PutString("cloud.platform", "aws_elastic_beanstalk")
				return inputMetrics
			},
			test: func(outputMetrics pmetric.Metrics) {
				cloudNamespaceAttribute, found := outputMetrics.ResourceMetrics().At(0).Resource().Attributes().Get("cloud.namespace")
				assert.True(t, found)
				assert.Equal(t, "ElasticBeanstalk", cloudNamespaceAttribute.Str())
			},
		},
		{
			name:              "does not add cloud.namespace attribute for unknown cloud.platform attribute values",
			addCloudNamespace: false,
			createMetrics: func() pmetric.Metrics {
				inputMetrics := pmetric.NewMetrics()
				inputMetrics.ResourceMetrics().AppendEmpty().Resource().Attributes().PutString("cloud.platform", "aws_eks")
				inputMetrics.ResourceMetrics().AppendEmpty().Resource().Attributes().PutString("cloud.platform", "aws_lambda")
				inputMetrics.ResourceMetrics().AppendEmpty().Resource().Attributes().PutString("cloud.platform", "azure_vm")
				inputMetrics.ResourceMetrics().AppendEmpty().Resource().Attributes().PutString("cloud.platform", "gcp_app_engine")
				return inputMetrics
			},
			test: func(outputMetrics pmetric.Metrics) {
				for i := 0; i < outputMetrics.ResourceMetrics().Len(); i++ {
					_, found := outputMetrics.ResourceMetrics().At(i).Resource().Attributes().Get("cloud.namespace")
					assert.False(t, found)
				}
			},
		},
		{
			name:              "does not add cloud.namespce attribute when disabled",
			addCloudNamespace: false,
			createMetrics: func() pmetric.Metrics {
				inputMetrics := pmetric.NewMetrics()
				inputMetrics.ResourceMetrics().AppendEmpty().Resource().Attributes().PutString("cloud.platform", "aws_ec2")
				return inputMetrics
			},
			test: func(outputMetrics pmetric.Metrics) {
				cloudNamespaceAttribute, found := outputMetrics.ResourceMetrics().At(0).Resource().Attributes().Get("cloud.namespace")
				assert.True(t, found)
				assert.Equal(t, "aws/ec2", cloudNamespaceAttribute.Str())
			},
		},
		{
			name:              "adds different cloud.namespace attributes to different resources",
			addCloudNamespace: true,
			createMetrics: func() pmetric.Metrics {
				inputMetrics := pmetric.NewMetrics()
				inputMetrics.ResourceMetrics().AppendEmpty().Resource().Attributes().PutString("cloud.platform", "aws_ec2")
				inputMetrics.ResourceMetrics().AppendEmpty().Resource().Attributes().PutString("cloud.platform", "aws_lambda")
				inputMetrics.ResourceMetrics().AppendEmpty().Resource().Attributes().PutString("cloud.platform", "aws_ecs")
				inputMetrics.ResourceMetrics().AppendEmpty().Resource().Attributes().PutString("cloud.platform", "aws_elastic_beanstalk")
				return inputMetrics
			},
			test: func(outputMetrics pmetric.Metrics) {
				ec2ResourceAttribute, found := outputMetrics.ResourceMetrics().At(0).Resource().Attributes().Get("cloud.namespace")
				assert.True(t, found)
				assert.Equal(t, "aws/ec2", ec2ResourceAttribute.Str())

				_, found = outputMetrics.ResourceMetrics().At(1).Resource().Attributes().Get("cloud.namespace")
				assert.False(t, found)

				ecsResourceAttribute, found := outputMetrics.ResourceMetrics().At(2).Resource().Attributes().Get("cloud.namespace")
				assert.True(t, found)
				assert.Equal(t, "ecs", ecsResourceAttribute.Str())

				beanstalkResourceAttribute, found := outputMetrics.ResourceMetrics().At(3).Resource().Attributes().Get("cloud.namespace")
				assert.True(t, found)
				assert.Equal(t, "ElasticBeanstalk", beanstalkResourceAttribute.Str())
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			// Arrange
			processor, err := newSumologicSchemaProcessor(newProcessorCreateSettings(), newCloudNamespaceConfig(testCase.addCloudNamespace))
			require.NoError(t, err)

			// Act
			outputMetrics, err := processor.processMetrics(context.Background(), testCase.createMetrics())
			require.NoError(t, err)

			// Assert
			testCase.test(outputMetrics)
		})
	}
}

func TestAddCloudNamespaceForTraces(t *testing.T) {
	testCases := []struct {
		name              string
		addCloudNamespace bool
		createTraces      func() ptrace.Traces
		test              func(ptrace.Traces)
	}{
		{
			name:              "adds cloud.namespace attribute for EC2",
			addCloudNamespace: true,
			createTraces: func() ptrace.Traces {
				inputTraces := ptrace.NewTraces()
				inputTraces.ResourceSpans().AppendEmpty().Resource().Attributes().PutString("cloud.platform", "aws_ec2")
				return inputTraces
			},
			test: func(outputTraces ptrace.Traces) {
				cloudNamespaceAttribute, found := outputTraces.ResourceSpans().At(0).Resource().Attributes().Get("cloud.namespace")
				assert.True(t, found)
				assert.Equal(t, "aws/ec2", cloudNamespaceAttribute.Str())
			},
		},
		{
			name:              "adds cloud.namespace attribute for ECS",
			addCloudNamespace: true,
			createTraces: func() ptrace.Traces {
				inputTraces := ptrace.NewTraces()
				inputTraces.ResourceSpans().AppendEmpty().Resource().Attributes().PutString("cloud.platform", "aws_ecs")
				return inputTraces
			},
			test: func(outputTraces ptrace.Traces) {
				cloudNamespaceAttribute, found := outputTraces.ResourceSpans().At(0).Resource().Attributes().Get("cloud.namespace")
				assert.True(t, found)
				assert.Equal(t, "ecs", cloudNamespaceAttribute.Str())
			},
		},
		{
			name:              "adds cloud.namespace attribute for Beanstalk",
			addCloudNamespace: true,
			createTraces: func() ptrace.Traces {
				inputTraces := ptrace.NewTraces()
				inputTraces.ResourceSpans().AppendEmpty().Resource().Attributes().PutString("cloud.platform", "aws_elastic_beanstalk")
				return inputTraces
			},
			test: func(outputTraces ptrace.Traces) {
				cloudNamespaceAttribute, found := outputTraces.ResourceSpans().At(0).Resource().Attributes().Get("cloud.namespace")
				assert.True(t, found)
				assert.Equal(t, "ElasticBeanstalk", cloudNamespaceAttribute.Str())
			},
		},
		{
			name:              "does not add cloud.namespace attribute for unknown cloud.platform attribute values",
			addCloudNamespace: false,
			createTraces: func() ptrace.Traces {
				inputTraces := ptrace.NewTraces()
				inputTraces.ResourceSpans().AppendEmpty().Resource().Attributes().PutString("cloud.platform", "aws_eks")
				inputTraces.ResourceSpans().AppendEmpty().Resource().Attributes().PutString("cloud.platform", "aws_lambda")
				inputTraces.ResourceSpans().AppendEmpty().Resource().Attributes().PutString("cloud.platform", "azure_vm")
				inputTraces.ResourceSpans().AppendEmpty().Resource().Attributes().PutString("cloud.platform", "gcp_app_engine")
				return inputTraces
			},
			test: func(outputTraces ptrace.Traces) {
				for i := 0; i < outputTraces.ResourceSpans().Len(); i++ {
					_, found := outputTraces.ResourceSpans().At(i).Resource().Attributes().Get("cloud.namespace")
					assert.False(t, found)
				}
			},
		},
		{
			name:              "does not add cloud.namespce attribute when disabled",
			addCloudNamespace: false,
			createTraces: func() ptrace.Traces {
				inputTraces := ptrace.NewTraces()
				inputTraces.ResourceSpans().AppendEmpty().Resource().Attributes().PutString("cloud.platform", "aws_ec2")
				return inputTraces
			},
			test: func(outputTraces ptrace.Traces) {
				cloudNamespaceAttribute, found := outputTraces.ResourceSpans().At(0).Resource().Attributes().Get("cloud.namespace")
				assert.True(t, found)
				assert.Equal(t, "aws/ec2", cloudNamespaceAttribute.Str())
			},
		},
		{
			name:              "adds different cloud.namespace attributes to different resources",
			addCloudNamespace: true,
			createTraces: func() ptrace.Traces {
				inputTraces := ptrace.NewTraces()
				inputTraces.ResourceSpans().AppendEmpty().Resource().Attributes().PutString("cloud.platform", "aws_ec2")
				inputTraces.ResourceSpans().AppendEmpty().Resource().Attributes().PutString("cloud.platform", "aws_lambda")
				inputTraces.ResourceSpans().AppendEmpty().Resource().Attributes().PutString("cloud.platform", "aws_ecs")
				inputTraces.ResourceSpans().AppendEmpty().Resource().Attributes().PutString("cloud.platform", "aws_elastic_beanstalk")
				return inputTraces
			},
			test: func(outputTraces ptrace.Traces) {
				ec2ResourceAttribute, found := outputTraces.ResourceSpans().At(0).Resource().Attributes().Get("cloud.namespace")
				assert.True(t, found)
				assert.Equal(t, "aws/ec2", ec2ResourceAttribute.Str())

				_, found = outputTraces.ResourceSpans().At(1).Resource().Attributes().Get("cloud.namespace")
				assert.False(t, found)

				ecsResourceAttribute, found := outputTraces.ResourceSpans().At(2).Resource().Attributes().Get("cloud.namespace")
				assert.True(t, found)
				assert.Equal(t, "ecs", ecsResourceAttribute.Str())

				beanstalkResourceAttribute, found := outputTraces.ResourceSpans().At(3).Resource().Attributes().Get("cloud.namespace")
				assert.True(t, found)
				assert.Equal(t, "ElasticBeanstalk", beanstalkResourceAttribute.Str())
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			// Arrange
			processor, err := newSumologicSchemaProcessor(newProcessorCreateSettings(), newCloudNamespaceConfig(testCase.addCloudNamespace))
			require.NoError(t, err)

			// Act
			outputTraces, err := processor.processTraces(context.Background(), testCase.createTraces())
			require.NoError(t, err)

			// Assert
			testCase.test(outputTraces)
		})
	}
}

func TestTranslateAttributesForLogs(t *testing.T) {
	testCases := []struct {
		name                string
		translateAttributes bool
		createLogs          func() plog.Logs
		test                func(plog.Logs)
	}{
		{
			name:                "translates one attribute",
			translateAttributes: true,
			createLogs: func() plog.Logs {
				inputLogs := plog.NewLogs()
				inputLogs.ResourceLogs().AppendEmpty().Resource().Attributes().PutString("cloud.account.id", "MyId1")
				inputLogs.ResourceLogs().AppendEmpty().Resource().Attributes().PutString("cloud.account.id", "MyId2")

				return inputLogs
			},
			test: func(outputLogs plog.Logs) {
				attribute1, found := outputLogs.ResourceLogs().At(0).Resource().Attributes().Get("AccountId")
				assert.True(t, found)
				assert.Equal(t, "MyId1", attribute1.Str())

				attribute2, found := outputLogs.ResourceLogs().At(1).Resource().Attributes().Get("AccountId")
				assert.True(t, found)
				assert.Equal(t, "MyId2", attribute2.Str())
			},
		},
		{
			name:                "does not translate",
			translateAttributes: false,
			createLogs: func() plog.Logs {
				inputLogs := plog.NewLogs()
				inputLogs.ResourceLogs().AppendEmpty().Resource().Attributes().PutString("cloud.account.id", "MyId1")
				inputLogs.ResourceLogs().AppendEmpty().Resource().Attributes().PutString("cloud.account.id", "MyId2")

				return inputLogs
			},
			test: func(outputLogs plog.Logs) {
				attribute1, found := outputLogs.ResourceLogs().At(0).Resource().Attributes().Get("cloud.account.id")
				assert.True(t, found)
				assert.Equal(t, "MyId1", attribute1.Str())

				attribute2, found := outputLogs.ResourceLogs().At(1).Resource().Attributes().Get("cloud.account.id")
				assert.True(t, found)
				assert.Equal(t, "MyId2", attribute2.Str())
			},
		},
		{
			name:                "translates no attributes",
			translateAttributes: true,
			createLogs: func() plog.Logs {
				inputLogs := plog.NewLogs()
				inputLogs.ResourceLogs().AppendEmpty().Resource().Attributes().PutString("not.actual.attr", "a1")
				inputLogs.ResourceLogs().AppendEmpty().Resource().Attributes().PutString("maybe.an.attr", "a2")
				return inputLogs
			},
			test: func(outputLogs plog.Logs) {
				attribute1, found := outputLogs.ResourceLogs().At(0).Resource().Attributes().Get("not.actual.attr")
				assert.True(t, found)
				assert.Equal(t, "a1", attribute1.Str())

				attribute2, found := outputLogs.ResourceLogs().At(1).Resource().Attributes().Get("maybe.an.attr")
				assert.True(t, found)
				assert.Equal(t, "a2", attribute2.Str())
			},
		},
		{
			name:                "translates many attributes, but not all",
			translateAttributes: true,
			createLogs: func() plog.Logs {
				inputLogs := plog.NewLogs()
				inputLogs.ResourceLogs().AppendEmpty().Resource().Attributes().PutString("cloud.account.id", "MyId")
				inputLogs.ResourceLogs().AppendEmpty().Resource().Attributes().PutString("maybe.an.attr", "a2")
				inputLogs.ResourceLogs().AppendEmpty().Resource().Attributes().PutString("k8s.cluster.name", "A cool cluster")

				return inputLogs
			},
			test: func(outputLogs plog.Logs) {
				attribute1, found := outputLogs.ResourceLogs().At(0).Resource().Attributes().Get("AccountId")
				assert.True(t, found)
				assert.Equal(t, "MyId", attribute1.Str())

				attribute2, found := outputLogs.ResourceLogs().At(1).Resource().Attributes().Get("maybe.an.attr")
				assert.True(t, found)
				assert.Equal(t, "a2", attribute2.Str())

				attribute3, found := outputLogs.ResourceLogs().At(2).Resource().Attributes().Get("Cluster")
				assert.True(t, found)
				assert.Equal(t, "A cool cluster", attribute3.Str())
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			// Arrange
			processor, err := newSumologicSchemaProcessor(newProcessorCreateSettings(), newTranslateAttributesConfig(testCase.translateAttributes))
			require.NoError(t, err)

			// Act
			outputLogs, err := processor.processLogs(context.Background(), testCase.createLogs())
			require.NoError(t, err)

			// Assert
			testCase.test(outputLogs)
		})
	}
}

func TestTranslateAttributesForMetrics(t *testing.T) {
	testCases := []struct {
		name                string
		translateAttributes bool
		createMetrics       func() pmetric.Metrics
		test                func(pmetric.Metrics)
	}{
		{
			name:                "translates many attributes, but not all",
			translateAttributes: true,
			createMetrics: func() pmetric.Metrics {
				inputMetrics := pmetric.NewMetrics()
				inputMetrics.ResourceMetrics().AppendEmpty().Resource().Attributes().PutString("cloud.account.id", "MyId")
				inputMetrics.ResourceMetrics().AppendEmpty().Resource().Attributes().PutString("maybe.an.attr", "a2")
				inputMetrics.ResourceMetrics().AppendEmpty().Resource().Attributes().PutString("k8s.cluster.name", "A cool cluster")

				return inputMetrics
			},
			test: func(outputMetrics pmetric.Metrics) {
				attribute1, found := outputMetrics.ResourceMetrics().At(0).Resource().Attributes().Get("AccountId")
				assert.True(t, found)
				assert.Equal(t, "MyId", attribute1.Str())

				attribute2, found := outputMetrics.ResourceMetrics().At(1).Resource().Attributes().Get("maybe.an.attr")
				assert.True(t, found)
				assert.Equal(t, "a2", attribute2.Str())

				attribute3, found := outputMetrics.ResourceMetrics().At(2).Resource().Attributes().Get("Cluster")
				assert.True(t, found)
				assert.Equal(t, "A cool cluster", attribute3.Str())
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			// Arrange
			processor, err := newSumologicSchemaProcessor(newProcessorCreateSettings(), newTranslateAttributesConfig(testCase.translateAttributes))
			require.NoError(t, err)

			// Act
			outputMetrics, err := processor.processMetrics(context.Background(), testCase.createMetrics())
			require.NoError(t, err)

			// Assert
			testCase.test(outputMetrics)
		})
	}
}

func TestTranslateAttributesForTraces(t *testing.T) {
	// Traces are NOT translated.
	testCases := []struct {
		name                string
		translateAttributes bool
		createTraces        func() ptrace.Traces
		test                func(ptrace.Traces)
	}{
		{
			name:                "does not translate even translatable attributes",
			translateAttributes: true,
			createTraces: func() ptrace.Traces {
				inputTraces := ptrace.NewTraces()
				inputTraces.ResourceSpans().AppendEmpty().Resource().Attributes().PutString("cloud.account.id", "MyId")
				inputTraces.ResourceSpans().AppendEmpty().Resource().Attributes().PutString("maybe.an.attr", "a2")
				inputTraces.ResourceSpans().AppendEmpty().Resource().Attributes().PutString("k8s.cluster.name", "A cool cluster")

				return inputTraces
			},
			test: func(outputTraces ptrace.Traces) {
				attribute1, found := outputTraces.ResourceSpans().At(0).Resource().Attributes().Get("cloud.account.id")
				assert.True(t, found)
				assert.Equal(t, "MyId", attribute1.Str())

				attribute2, found := outputTraces.ResourceSpans().At(1).Resource().Attributes().Get("maybe.an.attr")
				assert.True(t, found)
				assert.Equal(t, "a2", attribute2.Str())

				attribute3, found := outputTraces.ResourceSpans().At(2).Resource().Attributes().Get("k8s.cluster.name")
				assert.True(t, found)
				assert.Equal(t, "A cool cluster", attribute3.Str())
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			// Arrange
			processor, err := newSumologicSchemaProcessor(newProcessorCreateSettings(), newTranslateAttributesConfig(testCase.translateAttributes))
			require.NoError(t, err)

			// Act
			outputTraces, err := processor.processTraces(context.Background(), testCase.createTraces())
			require.NoError(t, err)

			// Assert
			testCase.test(outputTraces)
		})
	}
}

func TestTranslateTelegrafMetrics(t *testing.T) {
	testCases := []struct {
		testName        string
		originalNames   []string
		translatedNames []string
		shouldTranslate bool
	}{
		{
			testName:        "translates two names",
			originalNames:   []string{"cpu_usage_irq", "system_load1"},
			translatedNames: []string{"CPU_Irq", "CPU_LoadAvg_1min"},
			shouldTranslate: true,
		},
		{
			testName:        "does not translate",
			originalNames:   []string{"cpu_usage_irq", "system_load1"},
			translatedNames: []string{"cpu_usage_irq", "system_load1"},
			shouldTranslate: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.testName, func(t *testing.T) {
			// Arrange
			processor, err := newSumologicSchemaProcessor(newProcessorCreateSettings(), newTranslateTelegrafAttributesConfig(testCase.shouldTranslate))
			require.NoError(t, err)

			// Prepare metrics
			metrics := pmetric.NewMetrics()
			for _, name := range testCase.originalNames {
				metrics.ResourceMetrics().AppendEmpty().ScopeMetrics().AppendEmpty().Metrics().AppendEmpty().SetName(name)
			}

			// Act
			resultMetrics, err := processor.processMetrics(context.Background(), metrics)
			require.NoError(t, err)

			// Assert
			for index, name := range testCase.translatedNames {
				assert.Equal(t, name, resultMetrics.ResourceMetrics().At(index).ScopeMetrics().At(0).Metrics().At(0).Name())
			}
		})
	}
}

func newProcessorCreateSettings() component.ProcessorCreateSettings {
	return component.ProcessorCreateSettings{
		TelemetrySettings: component.TelemetrySettings{
			Logger: zap.NewNop(),
		},
	}
}

func newCloudNamespaceConfig(addCloudNamespace bool) *Config {
	config := createDefaultConfig().(*Config)
	config.AddCloudNamespace = addCloudNamespace
	config.TranslateAttributes = false
	config.TranslateTelegrafAttributes = false
	return config
}

func newTranslateAttributesConfig(translateAttributes bool) *Config {
	config := createDefaultConfig().(*Config)
	config.AddCloudNamespace = false
	config.TranslateAttributes = translateAttributes
	config.TranslateTelegrafAttributes = false
	return config
}

func newTranslateTelegrafAttributesConfig(translateTelegrafAttributes bool) *Config {
	config := createDefaultConfig().(*Config)
	config.AddCloudNamespace = false
	config.TranslateAttributes = false
	config.TranslateTelegrafAttributes = translateTelegrafAttributes
	return config
}
