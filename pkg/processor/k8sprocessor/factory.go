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

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/processor/processorhelper"

	"github.com/open-telemetry/opentelemetry-collector-contrib/internal/k8sconfig"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/k8sprocessor/kube"
)

const (
	// The value of "type" key in configuration.
	typeStr = "k8s_tagger"
)

var kubeClientProvider = kube.ClientProvider(nil)
var processorCapabilities = consumer.Capabilities{MutatesData: true}
var defaultExcludes = ExcludeConfig{
	Pods: []ExcludePodConfig{
		{Name: "jaeger-agent"},
		{Name: "jaeger-collector"},
		{Name: "otel-collector"},
		{Name: "otel-agent"},
		{Name: "collection-sumologic-otelcol"},
	},
}

// NewFactory returns a new factory for the k8s processor.
func NewFactory() component.ProcessorFactory {
	return processorhelper.NewFactory(
		typeStr,
		createDefaultConfig,
		processorhelper.WithTraces(createTracesProcessor),
		processorhelper.WithMetrics(createMetricsProcessor),
		processorhelper.WithLogs(createLogsProcessor),
	)
}

func createDefaultConfig() config.Processor {
	return &Config{
		ProcessorSettings: config.NewProcessorSettings(config.NewComponentID(typeStr)),
		APIConfig:         k8sconfig.APIConfig{AuthType: k8sconfig.AuthTypeServiceAccount},
		Extract: ExtractConfig{
			Delimiter: DefaultDelimiter,
		},
		Exclude: defaultExcludes,
	}
}

func createTracesProcessor(
	ctx context.Context,
	params component.ProcessorCreateSettings,
	cfg config.Processor,
	next consumer.Traces,
) (component.TracesProcessor, error) {
	return createTracesProcessorWithOptions(ctx, params, cfg, next)
}

func createLogsProcessor(
	ctx context.Context,
	params component.ProcessorCreateSettings,
	cfg config.Processor,
	nextLogsConsumer consumer.Logs,
) (component.LogsProcessor, error) {
	return createLogsProcessorWithOptions(ctx, params, cfg, nextLogsConsumer)
}

func createMetricsProcessor(
	ctx context.Context,
	params component.ProcessorCreateSettings,
	cfg config.Processor,
	nextMetricsConsumer consumer.Metrics,
) (component.MetricsProcessor, error) {
	return createMetricsProcessorWithOptions(ctx, params, cfg, nextMetricsConsumer)
}

func createTracesProcessorWithOptions(
	_ context.Context,
	params component.ProcessorCreateSettings,
	cfg config.Processor,
	next consumer.Traces,
	options ...Option,
) (component.TracesProcessor, error) {
	kp, err := createKubernetesProcessor(params, cfg, options...)
	if err != nil {
		return nil, err
	}

	return processorhelper.NewTracesProcessor(
		cfg,
		next,
		kp.ProcessTraces,
		processorhelper.WithCapabilities(processorCapabilities),
		processorhelper.WithStart(kp.Start),
		processorhelper.WithShutdown(kp.Shutdown))
}

func createMetricsProcessorWithOptions(
	_ context.Context,
	params component.ProcessorCreateSettings,
	cfg config.Processor,
	nextMetricsConsumer consumer.Metrics,
	options ...Option,
) (component.MetricsProcessor, error) {
	kp, err := createKubernetesProcessor(params, cfg, options...)
	if err != nil {
		return nil, err
	}

	return processorhelper.NewMetricsProcessor(
		cfg,
		nextMetricsConsumer,
		kp.ProcessMetrics,
		processorhelper.WithCapabilities(processorCapabilities),
		processorhelper.WithStart(kp.Start),
		processorhelper.WithShutdown(kp.Shutdown))
}

func createLogsProcessorWithOptions(
	_ context.Context,
	params component.ProcessorCreateSettings,
	cfg config.Processor,
	nextLogsConsumer consumer.Logs,
	options ...Option,
) (component.LogsProcessor, error) {
	kp, err := createKubernetesProcessor(params, cfg, options...)
	if err != nil {
		return nil, err
	}

	return processorhelper.NewLogsProcessor(
		cfg,
		nextLogsConsumer,
		kp.ProcessLogs,
		processorhelper.WithCapabilities(processorCapabilities),
		processorhelper.WithStart(kp.Start),
		processorhelper.WithShutdown(kp.Shutdown))
}

func createKubernetesProcessor(
	params component.ProcessorCreateSettings,
	cfg config.Processor,
	options ...Option,
) (*kubernetesprocessor, error) {
	kp := &kubernetesprocessor{logger: params.Logger}

	allOptions := append(createProcessorOpts(cfg), options...)

	for _, opt := range allOptions {
		if err := opt(kp); err != nil {
			return nil, err
		}
	}

	// This might have been set by an option already
	if kp.kc == nil {
		err := kp.initKubeClient(kp.logger, kubeClientProvider)
		if err != nil {
			return nil, err
		}
	}

	return kp, nil
}

func createProcessorOpts(cfg config.Processor) []Option {
	oCfg := cfg.(*Config)
	opts := []Option{}
	if oCfg.Passthrough {
		opts = append(opts, WithPassthrough())
	}

	// extraction rules
	opts = append(opts, WithExtractMetadata(oCfg.Extract.Metadata...))
	opts = append(opts, WithExtractLabels(oCfg.Extract.Labels...))
	opts = append(opts, WithExtractNamespaceLabels(oCfg.Extract.NamespaceLabels...))
	opts = append(opts, WithExtractAnnotations(oCfg.Extract.Annotations...))
	opts = append(opts, WithExtractTags(oCfg.Extract.Tags))

	if oCfg.OwnerLookupEnabled {
		opts = append(opts, WithOwnerLookupEnabled())
	}

	// filters
	opts = append(opts, WithFilterNode(oCfg.Filter.Node, oCfg.Filter.NodeFromEnvVar))
	opts = append(opts, WithFilterNamespace(oCfg.Filter.Namespace))
	opts = append(opts, WithFilterLabels(oCfg.Filter.Labels...))
	opts = append(opts, WithFilterFields(oCfg.Filter.Fields...))
	opts = append(opts, WithAPIConfig(oCfg.APIConfig))

	opts = append(opts, WithExtractPodAssociations(oCfg.Association...))

	opts = append(opts, WithDelimiter(oCfg.Extract.Delimiter))

	opts = append(opts, WithExcludes(oCfg.Exclude))

	return opts
}
