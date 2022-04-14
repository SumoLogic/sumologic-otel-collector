package mysqlreceiver 

import (
	"context"
	//"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/config/confignet"
	"go.opentelemetry.io/collector/consumer"
)

const (
	typeStr = "mysql"
)


func NewFactory() component.ReceiverFactory {
	return component.NewReceiverFactory(
		typeStr,
		createDefaultConfig,
		component.WithLogsReceiver(CreateLogsReceiver))
}

func createDefaultConfig() config.Receiver {
	rs := config.NewReceiverSettings(config.NewComponentID(typeStr))

	return &Config{
		ReceiverSettings:   &rs,
		CollectionInterval: "10s",
		AllowNativePasswords: true,
		Username:             "Username",
		NetAddr: confignet.NetAddr{
			Endpoint:  "localhost:3306",
			Transport: "tcp",
		},
	}
}


func CreateLogsReceiver(
	_ context.Context,
	params component.ReceiverCreateSettings,
	rConf config.Receiver,
	consumer consumer.Logs,
) (component.LogsReceiver, error) {
	
	cfg := rConf.(*Config)
	return newMySQLReceiver(params.Logger, cfg, consumer)
}
