package mysqlreceiver

import (
        "go.opentelemetry.io/collector/config/confignet"
        "go.opentelemetry.io/collector/config"
)

type Config struct {
        //scraperhelper.ScraperControllerSettings `mapstructure:",squash"`
	*config.ReceiverSettings `mapstructure:"-"`
        Username                                string `mapstructure:"username,omitempty"`
        Password                                string `mapstructure:"password,omitempty"`
        Database                                string `mapstructure:"database,omitempty"`
        AllowNativePasswords                    bool   `mapstructure:"allow_native_passwords,omitempty"`
        confignet.NetAddr                       `mapstructure:",squash"`
        CollectionInterval                      string `mapstructure:"collection_interval,omitempty"`
}