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
