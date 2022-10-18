//go:build go1.19
// +build go1.19

package sumologic_tests

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type timestampTestCase struct {
	timestamp string
	layout    string
	expected  time.Time
	name      string
}

func TestTimestamps(t *testing.T) {
	testCases := []timestampTestCase{
		{
			layout:    "2006-01-02 15:04:05,000 -07:00",
			timestamp: "2021-01-14 22:31:45,987 +13:30",
			expected:  time.Date(2021, 1, 14, 22, 31, 45, 987_000_000, time.FixedZone("", 13*60*60+30*60)),
			name:      "timestamp with comma parses successfully",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := time.Parse(tc.layout, tc.timestamp)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, result)
		})
	}
}
