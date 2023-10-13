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
	"regexp"
	"time"
)

// ADConfig defines configuration for Active Directory Inventory receiver.

type ADConfig struct {
	DN           string        `mapstructure:"base_dn"` // DN is the base distinguished name to search from
	Attributes   []string      `mapstructure:"attributes"`
	PollInterval time.Duration `mapstructure:"poll_interval"`
}

var (
	errInvalidDN           = errors.New("Base DN is incorrect, it must be in the format of CN=Users,DC=exampledomain,DC=com")
	errInvalidPollInterval = errors.New("poll interval is incorrect, it must be a duration greater than one second")
)

// Validate validates all portions of the relevant config
func (c *ADConfig) Validate() error {

	// Define the regular expression pattern for a valid Base DN
	pattern := `^(CN=[^,]+,)?CN=[^,]+,DC=[^,]+(,DC=[^,]+)*$`

	// Compile the regular expression pattern
	regex := regexp.MustCompile(pattern)

	// Check if the Base DN is valid
	if !regex.MatchString(c.DN) {
		return errInvalidDN
	}

	if c.PollInterval < 0 {
		return errInvalidPollInterval
	}

	return nil
}
