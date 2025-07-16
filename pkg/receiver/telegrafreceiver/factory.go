// Copyright 2021, OpenTelemetry Authors
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

package telegrafreceiver

import (
	"context"
	"fmt"
	"time"

	telegrafagent "github.com/influxdata/telegraf/agent"
	telegrafconfig "github.com/influxdata/telegraf/config"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver"
)

const (
	typeStr        = "telegraf"
	versionStr     = "v0.1"
	stabilityLevel = component.StabilityLevelBeta
)

var Type = component.MustNewType(typeStr)

// NewFactory creates a factory for telegraf receiver.
func NewFactory() receiver.Factory {
	return receiver.NewFactory(
		Type,
		createDefaultConfig,
		receiver.WithMetrics(createMetricsReceiver, stabilityLevel),
	)
}

func createDefaultConfig() component.Config {
	// TypeVal: config.Type(typeStr),
	// NameVal: typeStr,
	//
	return &Config{
		SeparateField: false,
		// we mostly expect to get recoverable errors from the memory_limiter, which is unlikely to have
		// a check interval less than 1s, so retrying much more frequently is pointless
		ConsumeRetryDelay: 500 * time.Millisecond,
		ConsumeMaxRetries: 10,
	}
}

// createMetricsReceiver creates a metrics receiver based on provided config.
func createMetricsReceiver(
	ctx context.Context,
	params receiver.Settings,
	cfg component.Config,
	nextConsumer consumer.Metrics,
) (receiver.Metrics, error) {
	tCfg, ok := cfg.(*Config)
	if !ok {
		return nil, fmt.Errorf("failed reading telegraf agent config from otc config")
	}

	tConfig := telegrafconfig.NewConfig()
	if err := tConfig.LoadConfigData([]byte(tCfg.AgentConfig), "agentconfig.toml"); err != nil {
		return nil, fmt.Errorf("failed loading telegraf agent config: %w", err)
	}
	tAgent := telegrafagent.NewAgent(tConfig)

	return &telegrafreceiver{
		agent:             tAgent,
		consumer:          nextConsumer,
		logger:            params.Logger,
		metricConverter:   newConverter(tCfg.SeparateField, params.Logger),
		consumeRetryDelay: tCfg.ConsumeRetryDelay,
		consumeMaxRetries: tCfg.ConsumeMaxRetries,
	}, nil
}
