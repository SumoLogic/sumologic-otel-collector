// Copyright  OpenTelemetry Authors
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

package rawk8seventsreceiver

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/consumer/consumertest"
	k8s "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
)

func TestDefaultConfig(t *testing.T) {
	cfg := createDefaultConfig()
	rCfg, ok := cfg.(*Config)
	require.True(t, ok)

	assert.Equal(t, &Config{
		ReceiverSettings: config.NewReceiverSettings(config.NewComponentID(typeStr)),
		APIConfig: APIConfig{
			AuthType: AuthTypeServiceAccount,
		},
		Namespaces:        []string{},
		MaxEventAge:       time.Minute,
		ConsumeMaxRetries: 20,
		ConsumeRetryDelay: time.Millisecond * 500,
	}, rCfg)
}

func TestFactoryType(t *testing.T) {
	assert.Equal(t, config.Type("raw_k8s_events"), NewFactory().Type())
}

func TestCreateReceiver(t *testing.T) {
	rCfg := createDefaultConfig().(*Config)

	// Fails with bad K8s Config.
	r, err := createLogsReceiver(
		context.Background(), componenttest.NewNopReceiverCreateSettings(),
		rCfg, consumertest.NewNop(),
	)
	assert.Error(t, err)
	assert.Nil(t, r)

	// Override for test.
	fakeClientFactory := func(apiConf APIConfig) (k8s.Interface, error) {
		return fake.NewSimpleClientset(), nil
	}
	r, err = createLogsReceiverWithClient(
		context.Background(),
		componenttest.NewNopReceiverCreateSettings(),
		rCfg, consumertest.NewNop(),
		fakeClientFactory,
	)
	assert.NoError(t, err)
	assert.NotNil(t, r)
}
