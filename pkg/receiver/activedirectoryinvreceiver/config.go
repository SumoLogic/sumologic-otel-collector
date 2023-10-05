package activedirectoryinvreceiver

import (
	"errors"
	"time"
)

// ADConfig defines configuration for Active Directory Inventory receiver.

var (
	defaultPollInterval = 60 * time.Second
)

type ADConfig struct {
	CN           string        `mapstructure:"cn"`
	OU           string        `mapstructure:"ou"`
	Password     string        `mapstructure:"password"`
	DC           string        `mapstructure:"domain"`
	Host         string        `mapstructure:"host"`
	PollInterval time.Duration `mapstructure:"poll_interval"`
}

var (
	errNoCN                = errors.New("no common name configured")
	errNoOU                = errors.New("no organizational unit configured")
	errNoPassword          = errors.New("no password configured")
	errNoDC                = errors.New("no domain configured")
	errNoHost              = errors.New("no host configured")
	errInvalidPollInterval = errors.New("poll interval is incorrect, it must be a duration greater than one second")
)

// Validate validates all portions of the relevant config
func (c *ADConfig) Validate() error {
	if c.CN == "" {
		return errNoCN
	}

	if c.OU == "" {
		return errNoOU
	}

	if c.Password == "" {
		return errNoPassword
	}

	if c.DC == "" {
		return errNoDC
	}

	if c.Host == "" {
		return errNoHost
	}

	if c.PollInterval < 0 {
		return errInvalidPollInterval
	}

	return nil
}
