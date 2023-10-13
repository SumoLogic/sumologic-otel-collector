// Copyright 2023, OpenTelemetry Authors
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
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.uber.org/zap"
)

type MockClient struct{}

func (c MockClient) Open(path string) (interface{}, error) {
	return nil, nil
}

func TestStart(t *testing.T) {
	cfg := CreateDefaultConfig().(*ADConfig)
	cfg.DN = "CN=Guest,CN=Users,DC=exampledomain,DC=com"

	sink := &consumertest.LogsSink{}
	mockClient := MockClient{}

	logsRcvr := newLogsReceiver(cfg, zap.NewNop(), mockClient, sink)

	err := logsRcvr.Start(context.Background(), componenttest.NewNopHost())
	require.NoError(t, err)

	err = logsRcvr.Shutdown(context.Background())
	require.NoError(t, err)
}
