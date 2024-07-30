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

package activedirectoryinvreceiver

import (
	"context"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver"
)

const (
	// The value of "type" key in configuration.
	typeStr = "active_directory_inv"
)

var Type = component.MustNewType(typeStr)

// NewFactory creates a factory for Active Directory Inventory receiver
func NewFactory() receiver.Factory {
	return receiver.NewFactory(
		Type,
		CreateDefaultConfig,
		receiver.WithLogs(createLogsReceiver, component.StabilityLevelAlpha),
	)
}

// CreateDefaultConfig creates the default configuration for the receiver
func CreateDefaultConfig() component.Config {
	return &ADConfig{
		BaseDN:       "",
		Attributes:   []string{"name", "mail", "department", "manager", "memberOf"},
		PollInterval: 24 * time.Hour,
	}
}

func createLogsReceiver(
	_ context.Context,
	params receiver.Settings,
	rConf component.Config,
	consumer consumer.Logs,
) (receiver.Logs, error) {
	cfg := rConf.(*ADConfig)
	adsiClient := &ADSIClient{}
	adRuntime := &ADRuntimeInfo{}
	rcvr := newLogsReceiver(cfg, params.Logger, adsiClient, adRuntime, consumer)
	return rcvr, nil
}
