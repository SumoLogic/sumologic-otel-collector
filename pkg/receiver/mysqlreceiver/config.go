package mysqlreceiver

import (
	"errors"

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
	SetConnMaxLifetime       int         `mapstructure:"setconnmaxlifetimemins,omitempty"`
	SetMaxOpenConns          int         `mapstructure:"setmaxopenconns,omitempty"`
	SetMaxIdleConns          int         `mapstructure:"setmaxidleconns,omitempty"`
	SetMaxNoDatabaseWorkers  int         `mapstructure:"setmaxnodatabaseworkers,omitempty"`
}

type DBQueries struct {
	QueryId                      string `mapstructure:"queryid"`
	Query                        string `mapstructure:"query"`
	IndexColumnName              string `mapstructure:"index_column_name,omitempty"`
	InitialIndexColumnStartValue string `mapstructure:"initial_index_column_start_value,omitempty"`
	IndexColumnType              string `mapstructure:"index_column_type,omitempty"`
}

func (cfg *Config) Validate() error {

	var queryIds []string
	var size = len(cfg.DBQueries)
	for i := 0; i < size; i++ {
		queryIds = append(queryIds, cfg.DBQueries[i].QueryId)
	}
	queryIdCount := make(map[string]int)
	for _, item := range queryIds {
		_, exist := queryIdCount[item]
		if exist {
			queryIdCount[item] += 1
		} else {
			queryIdCount[item] = 1
		}
	}
	for _, count := range queryIdCount {
		if count > 1 {
			err := errors.New("multiple queries have the same queryId which is not allowed")
			return err
		}
	}
	return nil
}
