package activedirectoryinvreceiver

import "time"

// ADConfig defines configuration for Active Directory Inventory receiver.

type ADConfig struct {
	CN           string        `mapstructure:"cn"`
	OU           string        `mapstructure:"ou"`
	Password     string        `mapstructure:"password"`
	DC           string        `mapstructure:"domain"`
	Host         string        `mapstructure:"host"`
	PollInterval time.Duration `mapstructure:"poll_interval"`
}
