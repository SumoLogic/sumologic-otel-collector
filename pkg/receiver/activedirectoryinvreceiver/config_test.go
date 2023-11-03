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
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestValidate(t *testing.T) {
	cases := []struct {
		name        string
		config      ADConfig
		expectedErr error
	}{
		{
			name: "Valid Config CN",
			config: ADConfig{
				BaseDN:       "CN=Guest,CN=Users,DC=exampledomain,DC=com",
				Attributes:   []string{"name"},
				PollInterval: 60 * time.Second,
			},
		},
		{
			name: "Valid Config DC",
			config: ADConfig{
				BaseDN:       "DC=exampledomain,DC=com",
				Attributes:   []string{"name"},
				PollInterval: 60 * time.Second,
			},
		},
		{
			name: "Valid Config OU",
			config: ADConfig{
				BaseDN:       "CN=Guest,OU=Users,DC=exampledomain,DC=com",
				Attributes:   []string{"name"},
				PollInterval: 24 * time.Hour,
			},
		},
		{
			name: "Invalid DN",
			config: ADConfig{
				BaseDN:       "NA",
				Attributes:   []string{"name"},
				PollInterval: 24 * time.Hour,
			},
			expectedErr: errInvalidDN,
		},
		{
			name: "Invalid Empty DN",
			config: ADConfig{
				BaseDN:       "",
				Attributes:   []string{"name"},
				PollInterval: 24 * time.Hour,
			},
			expectedErr: errInvalidDN,
		},
		{
			name: "Invalid Poll Interval",
			config: ADConfig{
				BaseDN:       "CN=Users,DC=exampledomain,DC=com",
				Attributes:   []string{"name"},
				PollInterval: 0,
			},
			expectedErr: errInvalidPollInterval,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.config.Validate()
			if tc.expectedErr != nil {
				require.ErrorContains(t, err, tc.expectedErr.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}
