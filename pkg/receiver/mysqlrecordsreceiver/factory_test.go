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
package mysqlrecordsreceiver

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/config/confignet"
	"go.opentelemetry.io/collector/consumer/consumertest"
)

func TestValidType(t *testing.T) {
	factory := NewFactory()
	ft := factory.Type()
	require.EqualValues(t, "mysqlrecords", ft)
}

func TestInvalidType(t *testing.T) {
	factory := NewFactory()
	ft := factory.Type()
	require.NotEqualValues(t, "garbage", ft)
}

func TestCreateLogsReceiver(t *testing.T) {
	factory := NewFactory()
	rs := config.NewReceiverSettings(config.NewComponentID("mysql"))
	logsReceiver, err := factory.CreateLogsReceiver(
		context.Background(),
		componenttest.NewNopReceiverCreateSettings(),
		&Config{
			ReceiverSettings:   &rs,
			CollectionInterval: "10s",
			Username:           "mysqluser",
			Password:           "userpass",
			NetAddr: confignet.NetAddr{
				Endpoint:  "localhost:3306",
				Transport: "tcp",
			},
		},
		consumertest.NewNop(),
	)
	require.NoError(t, err)
	require.NotNil(t, logsReceiver)
}
