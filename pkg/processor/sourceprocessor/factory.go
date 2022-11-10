// Copyright 2019 OpenTelemetry Authors
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
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/processor/processorhelper"
)

const (
	// The value of "type" key in configuration.
	typeStr = "source"

	defaultCollector = ""

	defaultSourceHost                = "%{k8s.pod.hostname}"
	defaultSourceName                = "%{k8s.namespace.name}.%{k8s.pod.name}.%{k8s.container.name}"
	defaultSourceCategory            = "%{k8s.namespace.name}/%{k8s.pod.pod_name}"
	defaultSourceCategoryPrefix      = "kubernetes/"
	defaultSourceCategoryReplaceDash = "/"

	defaultAnnotationPrefix   = "k8s.pod.annotation."
	defaultPodKey             = "k8s.pod.name"
	defaultPodNameKey         = "k8s.pod.pod_name"
	defaultPodTemplateHashKey = "k8s.pod.label.pod-template-hash"

	stabilityLevel = component.StabilityLevelBeta
)

var processorCapabilities = consumer.Capabilities{MutatesData: true}

// NewFactory returns a new factory for the Span processor.
func NewFactory() component.ProcessorFactory {
	return component.NewProcessorFactory(
		typeStr,
		createDefaultConfig,
		component.WithTracesProcessor(createTracesProcessor, stabilityLevel),
		component.WithMetricsProcessor(createMetricsProcessor, stabilityLevel),
		component.WithLogsProcessor(createLogsProcessor, stabilityLevel),
	)
}

// createDefaultConfig creates the default configuration for processor.
func createDefaultConfig() component.ProcessorConfig {
	ps := config.NewProcessorSettings(component.NewID(typeStr))
	return &Config{
		ProcessorSettings:         &ps,
		Collector:                 defaultCollector,
		SourceHost:                defaultSourceHost,
		SourceName:                defaultSourceName,
		SourceCategory:            defaultSourceCategory,
		SourceCategoryPrefix:      defaultSourceCategoryPrefix,
		SourceCategoryReplaceDash: defaultSourceCategoryReplaceDash,

		AnnotationPrefix:   defaultAnnotationPrefix,
		PodKey:             defaultPodKey,
		PodNameKey:         defaultPodNameKey,
		PodTemplateHashKey: defaultPodTemplateHashKey,

		ContainerAnnotations: ContainerAnnotationsConfig{
			Enabled: false,
			Prefixes: []string{
				"sumologic.com/",
			},
		},
	}
}

// CreateTraceProcessor creates a trace processor based on this config.
func createTracesProcessor(
	ctx context.Context,
	params component.ProcessorCreateSettings,
	cfg component.ProcessorConfig,
	next consumer.Traces) (component.TracesProcessor, error) {

	oCfg := cfg.(*Config)

	sp := newSourceProcessor(oCfg)

	return processorhelper.NewTracesProcessor(
		ctx,
		params,
		cfg,
		next,
		sp.ProcessTraces,
		processorhelper.WithCapabilities(processorCapabilities),
	)
}

// createMetricsProcessor creates a metrics processor based on this config
func createMetricsProcessor(
	ctx context.Context,
	params component.ProcessorCreateSettings,
	cfg component.ProcessorConfig,
	next consumer.Metrics,
) (component.MetricsProcessor, error) {
	oCfg := cfg.(*Config)

	sp := newSourceProcessor(oCfg)
	return processorhelper.NewMetricsProcessor(
		ctx,
		params,
		cfg,
		next,
		sp.ProcessMetrics,
		processorhelper.WithCapabilities(processorCapabilities),
	)
}

// createLogsProcessor creates a logs processor based on this config
func createLogsProcessor(
	ctx context.Context,
	params component.ProcessorCreateSettings,
	cfg component.ProcessorConfig,
	next consumer.Logs,
) (component.LogsProcessor, error) {
	oCfg := cfg.(*Config)

	sp := newSourceProcessor(oCfg)
	return processorhelper.NewLogsProcessor(
		ctx,
		params,
		cfg,
		next,
		sp.ProcessLogs,
		processorhelper.WithCapabilities(processorCapabilities),
	)
}
