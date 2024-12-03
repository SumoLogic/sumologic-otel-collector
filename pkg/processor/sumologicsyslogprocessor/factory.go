// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package sumologicsyslogprocessor

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/processor"
	"go.opentelemetry.io/collector/processor/processorhelper"
)

const (
	// The value of "type" Tail Sampling in configuration.
	typeStr        = "sumologic_syslog"
	stabilityLevel = component.StabilityLevelBeta
)

var (
	processorCapabilities = consumer.Capabilities{MutatesData: true}
	Type                  = component.MustNewType(typeStr)
)

// NewFactory returns a new factory for the Tail Sampling processor.
func NewFactory() processor.Factory {
	return processor.NewFactory(
		Type,
		createDefaultConfig,
		processor.WithLogs(createLogProcessor, stabilityLevel))
}

func createDefaultConfig() component.Config {
	return &Config{
		FacilityAttr: defaultFacilityAttr,
	}
}

func createLogProcessor(
	ctx context.Context,
	params processor.Settings,
	cfg component.Config,
	nextConsumer consumer.Logs,
) (processor.Logs, error) {
	tCfg := cfg.(*Config)

	ssp, err := newSumologicSyslogProcessor(tCfg)
	if err != nil {
		return nil, err
	}

	return processorhelper.NewLogs(
		ctx,
		params,
		cfg,
		nextConsumer,
		ssp.ProcessLogs,
		processorhelper.WithCapabilities(processorCapabilities))
}
