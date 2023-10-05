// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package activedirectoryinvreceiver

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidate(t *testing.T) {
	cases := []struct {
		name        string
		config      ADConfig
		expectedErr error
	}{
		{
			name: "Valid Config",
			config: ADConfig{
				CN:           "test user",
				OU:           "test",
				Password:     "test",
				DC:           "exampledomain.com",
				Host:         "hostname.exampledomain.com",
				PollInterval: 60,
			},
		},
		{
			name: "Invalid No CN",
			config: ADConfig{
				CN: "",
			},
			expectedErr: errNoCN,
		},
		{
			name: "Invalid Poll Interval",
			config: ADConfig{
				CN:           "test user",
				OU:           "test",
				Password:     "test",
				DC:           "exampledomain.com",
				Host:         "hostname.exampledomain.com",
				PollInterval: -1,
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
