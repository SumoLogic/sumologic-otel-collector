package mysqlreceiver

import (
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/config/confignet"
)

type Config struct {
	*config.ReceiverSettings `mapstructure:",squash"`
	Username                 string `mapstructure:"username,omitempty"`
	Password                 string `mapstructure:"password,omitempty"`
	Database                 string `mapstructure:"database,omitempty"`
	Endpoint                 string `mapstructure:"endpoint,omitempty"`
	Transport                string `mapstructure:"transport,omitempty"`
	AllowNativePasswords     bool   `mapstructure:"allow_native_passwords,omitempty"`
	confignet.NetAddr        `mapstructure:",squash"`
	CollectionInterval       string      `mapstructure:"collection_interval,omitempty"`
	DBQueries                []DBQueries `mapstructure:"db_queries"`
}

type DBQueries struct {
	QueryId                      string `mapstructure:"queryid"`
	Query                        string `mapstructure:"query"`
	IndexColumnName              string `mapstructure:"index_column_name,omitempty"`
	InitialIndexColumnStartValue string `mapstructure:"initial_index_column_start_value,omitempty"`
	IndexColumnType              string `mapstructure:"index_column_type,omitempty"`
}
