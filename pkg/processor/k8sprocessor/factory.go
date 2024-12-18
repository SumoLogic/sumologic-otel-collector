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
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/processor"
	"go.opentelemetry.io/collector/processor/processorhelper"

	"github.com/open-telemetry/opentelemetry-collector-contrib/internal/k8sconfig"

	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/k8sprocessor/kube"
)

const (
	// The value of "type" key in configuration.
	typeStr        = "k8s_tagger"
	stabilityLevel = component.StabilityLevelBeta
)

var Type = component.MustNewType(typeStr)

var kubeClientProvider = kube.ClientProvider(nil)
var processorCapabilities = consumer.Capabilities{MutatesData: true}

// NewFactory returns a new factory for the k8s processor.
func NewFactory() processor.Factory {
	return processor.NewFactory(
		Type,
		createDefaultConfig,
		processor.WithTraces(createTracesProcessor, stabilityLevel),
		processor.WithMetrics(createMetricsProcessor, stabilityLevel),
		processor.WithLogs(createLogsProcessor, stabilityLevel),
	)
}

func createDefaultConfig() component.Config {
	return &Config{
		APIConfig: k8sconfig.APIConfig{AuthType: k8sconfig.AuthTypeServiceAccount},
		Limit:     DefaultLimit,
		Extract: ExtractConfig{
			Delimiter: DefaultDelimiter,
		},
	}
}

func createTracesProcessor(
	ctx context.Context,
	params processor.Settings,
	cfg component.Config,
	next consumer.Traces,
) (processor.Traces, error) {
	return createTracesProcessorWithOptions(ctx, params, cfg, next)
}

func createLogsProcessor(
	ctx context.Context,
	params processor.Settings,
	cfg component.Config,
	nextLogsConsumer consumer.Logs,
) (processor.Logs, error) {
	return createLogsProcessorWithOptions(ctx, params, cfg, nextLogsConsumer)
}

func createMetricsProcessor(
	ctx context.Context,
	params processor.Settings,
	cfg component.Config,
	nextMetricsConsumer consumer.Metrics,
) (processor.Metrics, error) {
	return createMetricsProcessorWithOptions(ctx, params, cfg, nextMetricsConsumer)
}

func createTracesProcessorWithOptions(
	ctx context.Context,
	params processor.Settings,
	cfg component.Config,
	next consumer.Traces,
	options ...Option,
) (processor.Traces, error) {
	kp, err := createKubernetesProcessor(params, cfg, options...)
	if err != nil {
		return nil, err
	}

	return processorhelper.NewTraces(
		ctx,
		params,
		cfg,
		next,
		kp.ProcessTraces,
		processorhelper.WithCapabilities(processorCapabilities),
		processorhelper.WithStart(kp.Start),
		processorhelper.WithShutdown(kp.Shutdown))
}

func createMetricsProcessorWithOptions(
	ctx context.Context,
	params processor.Settings,
	cfg component.Config,
	nextMetricsConsumer consumer.Metrics,
	options ...Option,
) (processor.Metrics, error) {
	kp, err := createKubernetesProcessor(params, cfg, options...)
	if err != nil {
		return nil, err
	}

	return processorhelper.NewMetrics(
		ctx,
		params,
		cfg,
		nextMetricsConsumer,
		kp.ProcessMetrics,
		processorhelper.WithCapabilities(processorCapabilities),
		processorhelper.WithStart(kp.Start),
		processorhelper.WithShutdown(kp.Shutdown))
}

func createLogsProcessorWithOptions(
	ctx context.Context,
	params processor.Settings,
	cfg component.Config,
	nextLogsConsumer consumer.Logs,
	options ...Option,
) (processor.Logs, error) {
	kp, err := createKubernetesProcessor(params, cfg, options...)
	if err != nil {
		return nil, err
	}

	return processorhelper.NewLogs(
		ctx,
		params,
		cfg,
		nextLogsConsumer,
		kp.ProcessLogs,
		processorhelper.WithCapabilities(processorCapabilities),
		processorhelper.WithStart(kp.Start),
		processorhelper.WithShutdown(kp.Shutdown))
}

func createKubernetesProcessor(
	params processor.Settings,
	cfg component.Config,
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

func createProcessorOpts(cfg component.Config) []Option {
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
	opts = append(opts, WithExtractNamespaceAnnotations(oCfg.Extract.NamespaceAnnotations...))
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

	opts = append(opts, WithLimit(oCfg.Limit))

	opts = append(opts, WithExcludes(oCfg.Exclude))

	return opts
}
