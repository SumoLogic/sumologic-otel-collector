// Copyright 2021, OpenTelemetry Authors
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
	"errors"
	"time"
)

// ADConfig defines configuration for Active Directory Inventory receiver.

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
