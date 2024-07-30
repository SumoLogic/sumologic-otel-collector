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

package rawk8seventsreceiver

import (
	"context"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver"
	"k8s.io/apimachinery/pkg/fields"
	k8s "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

const (
	// Value of "type" key in configuration.
	typeStr        = "raw_k8s_events"
	stabilityLevel = component.StabilityLevelBeta
)

var Type = component.MustNewType(typeStr)

// NewFactory creates a factory for rawk8sevents receiver.
func NewFactory() receiver.Factory {
	return receiver.NewFactory(
		Type,
		createDefaultConfig,
		receiver.WithLogs(createLogsReceiver, stabilityLevel))
}

func createDefaultConfig() component.Config {
	return &Config{
		APIConfig: APIConfig{
			AuthType: AuthTypeServiceAccount,
		},
		Namespaces:        []string{},
		MaxEventAge:       time.Minute,
		ConsumeMaxRetries: 20,
		ConsumeRetryDelay: time.Millisecond * 500,
	}
}

func createLogsReceiver(
	ctx context.Context,
	params receiver.Settings,
	cfg component.Config,
	consumer consumer.Logs,
) (receiver.Logs, error) {

	k8sClientFactory := MakeClient
	return createLogsReceiverWithClient(ctx, params, cfg, consumer, k8sClientFactory)
}

func createLogsReceiverWithClient(
	_ context.Context,
	params receiver.Settings,
	cfg component.Config,
	consumer consumer.Logs,
	clientFactory func(APIConfig) (k8s.Interface, error),
) (receiver.Logs, error) {
	rCfg := cfg.(*Config)

	k8sClient, err := clientFactory(rCfg.APIConfig)
	if err != nil {
		return nil, err
	}

	listerWatcherFactory := func(
		c cache.Getter,
		resource string,
		namespace string,
		fieldSelector fields.Selector,
	) cache.ListerWatcher {
		return cache.NewListWatchFromClient(c, resource, namespace, fieldSelector)
	}

	return newRawK8sEventsReceiver(params, rCfg, consumer, k8sClient, listerWatcherFactory)
}
