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

package cascadingfilterprocessor

import (
	"context"
	"time"

	"go.opencensus.io/stats/view"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/config/configtelemetry"
	"go.opentelemetry.io/collector/consumer"

	cfconfig "github.com/SumoLogic/sumologic-otel-collector/pkg/processor/cascadingfilterprocessor/config"
)

const (
	// The value of "type" Cascading Filter in configuration.
	typeStr        = "cascading_filter"
	stabilityLevel = component.StabilityLevelBeta
)

func init() {
	// TODO: this is hardcoding the metrics level
	err := view.Register(CascadingFilterMetricViews(configtelemetry.LevelNormal)...)
	if err != nil {
		panic("failed to register cascadingfilterprocessor: " + err.Error())
	}
}

// NewFactory returns a new factory for the Cascading Filter processor.
func NewFactory() component.ProcessorFactory {
	return component.NewProcessorFactory(
		typeStr,
		createDefaultConfig,
		component.WithTracesProcessor(createTraceProcessor, stabilityLevel))
}

func createDefaultConfig() component.ProcessorConfig {
	id := component.NewID("cascading_filter")
	ps := config.NewProcessorSettings(id)

	return &cfconfig.Config{
		ProcessorSettings: &ps,
		DecisionWait:      30 * time.Second,
		NumTraces:         100000,
		SpansPerSecond:    0,
	}
}

func createTraceProcessor(
	_ context.Context,
	settings component.ProcessorCreateSettings,
	cfg component.ProcessorConfig,
	nextConsumer consumer.Traces,
) (component.TracesProcessor, error) {
	tCfg := cfg.(*cfconfig.Config)
	return newTraceProcessor(settings.Logger, nextConsumer, *tCfg)
}
