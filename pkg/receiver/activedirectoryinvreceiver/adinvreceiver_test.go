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
	"fmt"
	"testing"
	"time"

	"github.com/go-adsi/adsi"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.uber.org/zap"
)

type MockClient struct {
	mock.Mock
}

func (mc *MockClient) Open(path string, resourceLogs *plog.ResourceLogs) (Container, error) {
	args := mc.Called(path, resourceLogs)
	return args.Get(0).(Container), args.Error(1)
}

type MockContainer struct {
	mock.Mock
}

func (mc *MockContainer) ToObject() (Object, error) {
	args := mc.Called()
	return args.Get(0).(Object), args.Error(1)
}

func (mc *MockContainer) Close() {
	mc.Called()
}

func (mc *MockContainer) Children() (ObjectIter, error) {
	args := mc.Called()
	return args.Get(0).(ObjectIter), args.Error(1)
}

type MockObject struct {
	mock.Mock
}

func (mo *MockObject) Attrs(key string) ([]interface{}, error) {
	args := mo.Called(key)
	return args.Get(0).([]interface{}), args.Error(1)
}

func (mo *MockObject) ToContainer() (Container, error) {
	args := mo.Called()
	return args.Get(0).(Container), args.Error(1)
}

type MockObjectIter struct {
	mock.Mock
}

func (mo *MockObjectIter) Next() (*adsi.Object, error) {
	args := mo.Called()
	return args.Get(0).(*adsi.Object), args.Error(1)
}

func (mo *MockObjectIter) Close() {
	mo.Called()
}

func TestStart(t *testing.T) {
	cfg := CreateDefaultConfig().(*ADConfig)
	cfg.DN = "CN=Guest,CN=Users,DC=exampledomain,DC=com"

	sink := &consumertest.LogsSink{}
	mockClient := &MockClient{}

	logsRcvr := newLogsReceiver(cfg, zap.NewNop(), mockClient, sink)

	err := logsRcvr.Start(context.Background(), componenttest.NewNopHost())
	require.NoError(t, err)

	err = logsRcvr.Shutdown(context.Background())
	require.NoError(t, err)
}

func TestPoll(t *testing.T) {
	cfg := CreateDefaultConfig().(*ADConfig)
	cfg.DN = "CN=Guest,CN=Users,DC=exampledomain,DC=com"
	cfg.PollInterval = 1
	cfg.Attributes = []string{"name"}

	sink := &consumertest.LogsSink{}
	mockClient := defaultMockClient()

	logsRcvr := newLogsReceiver(cfg, zap.NewNop(), mockClient, sink)

	err := logsRcvr.Start(context.Background(), componenttest.NewNopHost())
	require.NoError(t, err)

	require.Eventually(t, func() bool {
		return sink.LogRecordCount() > 0
	}, 2*time.Second, 10*time.Millisecond)

	err = logsRcvr.Shutdown(context.Background())
	require.NoError(t, err)
}

func defaultMockClient() Client {
	mockClient := &MockClient{}
	mockContainer := &MockContainer{}
	mockObject := &MockObject{}
	mockObjectIter := &MockObjectIter{}
	attrs := []interface{}{"Guest", "test"}
	mockContainer.On("ToObject").Return(mockObject, nil)
	mockContainer.On("Children").Return(mockObjectIter, fmt.Errorf("no children"))
	mockContainer.On("Close").Return(nil)
	mockObject.On("Attrs", mock.Anything).Return(attrs, nil)
	mockObject.On("ToContainer").Return(mockContainer, nil)
	mockClient.On("Open", mock.Anything, mock.Anything).Return(mockContainer, nil)
	return mockClient
}
