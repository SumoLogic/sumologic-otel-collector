// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package dns

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/SumoLogic/sumologic-otel-collector/pkg/processor/lookupprocessor/lookupsource"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name:    "valid A record type",
			config:  Config{RecordType: RecordTypeA, Timeout: 5 * time.Second},
			wantErr: false,
		},
		{
			name:    "valid AAAA record type",
			config:  Config{RecordType: RecordTypeAAAA, Timeout: 5 * time.Second},
			wantErr: false,
		},
		{
			name:    "valid PTR record type",
			config:  Config{RecordType: RecordTypePTR, Timeout: 5 * time.Second},
			wantErr: false,
		},
		{
			name:    "empty record_type defaults to A",
			config:  Config{Timeout: 5 * time.Second},
			wantErr: false,
		},
		{
			name:    "invalid record type",
			config:  Config{RecordType: "invalid", Timeout: 5 * time.Second},
			wantErr: true,
		},
		{
			name:    "negative timeout",
			config:  Config{RecordType: RecordTypeA, Timeout: -1 * time.Second},
			wantErr: true,
		},
		{
			name:    "valid with resolver",
			config:  Config{RecordType: RecordTypeA, Timeout: 5 * time.Second, Resolver: "8.8.8.8:53"},
			wantErr: false,
		},
		{
			name:    "invalid resolver format",
			config:  Config{RecordType: RecordTypeA, Timeout: 5 * time.Second, Resolver: "8.8.8.8"},
			wantErr: true,
		},
		{
			name:    "valid with multiple results",
			config:  Config{RecordType: RecordTypeA, Timeout: 5 * time.Second, MultipleResults: true},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNewFactory(t *testing.T) {
	factory := NewFactory()
	assert.Equal(t, sourceType, factory.Type())

	cfg := factory.CreateDefaultConfig()
	assert.NotNil(t, cfg)

	defaultCfg, ok := cfg.(*Config)
	require.True(t, ok)
	assert.Equal(t, RecordTypePTR, defaultCfg.RecordType)
	assert.Equal(t, 5*time.Second, defaultCfg.Timeout)
	assert.False(t, defaultCfg.MultipleResults)
}

func TestCreateSource(t *testing.T) {
	factory := NewFactory()
	cfg := &Config{
		RecordType: RecordTypeA,
		Timeout:    5 * time.Second,
	}

	source, err := factory.CreateSource(context.Background(), lookupsource.CreateSettings{}, cfg)
	require.NoError(t, err)
	require.NotNil(t, source)
	assert.Equal(t, sourceType, source.Type())
}

func TestForwardLookup(t *testing.T) {
	factory := NewFactory()
	cfg := &Config{
		RecordType: RecordTypeA,
		Timeout:    10 * time.Second,
	}

	source, err := factory.CreateSource(context.Background(), lookupsource.CreateSettings{}, cfg)
	require.NoError(t, err)

	ctx := context.Background()

	// Test localhost resolution - should work on all systems
	t.Run("resolve localhost", func(t *testing.T) {
		result, found, err := source.Lookup(ctx, "localhost")
		require.NoError(t, err)
		if found {
			// localhost typically resolves to 127.0.0.1 or ::1
			assert.NotEmpty(t, result)
			t.Logf("localhost resolved to: %v", result)
		}
	})

	// Test non-existent domain
	t.Run("non-existent domain", func(t *testing.T) {
		result, found, err := source.Lookup(ctx, "this-domain-should-not-exist-12345.invalid")
		assert.NoError(t, err) // Should not error, just not found
		assert.False(t, found)
		assert.Nil(t, result)
	})
}

func TestReverseLookup(t *testing.T) {
	factory := NewFactory()
	cfg := &Config{
		RecordType: RecordTypePTR,
		Timeout:    10 * time.Second,
	}

	source, err := factory.CreateSource(context.Background(), lookupsource.CreateSettings{}, cfg)
	require.NoError(t, err)

	ctx := context.Background()

	// Test localhost IP reverse resolution
	t.Run("reverse resolve localhost", func(t *testing.T) {
		result, found, err := source.Lookup(ctx, "127.0.0.1")
		require.NoError(t, err)
		if found {
			assert.NotEmpty(t, result)
			t.Logf("127.0.0.1 resolved to: %v", result)
		}
	})

	// Test invalid IP
	t.Run("invalid IP", func(t *testing.T) {
		result, found, err := source.Lookup(ctx, "999.999.999.999")
		// Invalid IP should either error or not be found
		if err == nil {
			assert.False(t, found)
			assert.Nil(t, result)
		}
	})
}

func TestIPv6Support(t *testing.T) {
	ctx := context.Background()
	factory := NewFactory()

	t.Run("AAAA record lookup - IPv6 only", func(t *testing.T) {
		cfg := &Config{
			RecordType:      RecordTypeAAAA,
			Timeout:         10 * time.Second,
			MultipleResults: true,
		}
		source, err := factory.CreateSource(ctx, lookupsource.CreateSettings{}, cfg)
		require.NoError(t, err)

		// dns.google has both A and AAAA records
		result, found, err := source.Lookup(ctx, "dns.google")
		if err != nil {
			t.Skipf("Network not accessible: %v", err)
		}
		if found {
			resultStr, ok := result.(string)
			require.True(t, ok)
			t.Logf("dns.google AAAA records: %v", resultStr)
			// Should contain IPv6 addresses (with colons)
			assert.Contains(t, resultStr, ":")
		}
	})

	t.Run("A record lookup - IPv4 only", func(t *testing.T) {
		cfg := &Config{
			RecordType:      RecordTypeA,
			Timeout:         10 * time.Second,
			MultipleResults: true,
		}
		source, err := factory.CreateSource(ctx, lookupsource.CreateSettings{}, cfg)
		require.NoError(t, err)

		// dns.google has both A and AAAA records
		result, found, err := source.Lookup(ctx, "dns.google")
		if err != nil {
			t.Skipf("Network not accessible: %v", err)
		}
		if found {
			resultStr, ok := result.(string)
			require.True(t, ok)
			t.Logf("dns.google A records: %v", resultStr)
			// Should contain IPv4 addresses (dots, no colons)
			assert.Contains(t, resultStr, ".")
			assert.NotContains(t, resultStr, ":")
		}
	})

	t.Run("PTR record - IPv6 localhost reverse lookup", func(t *testing.T) {
		cfg := &Config{
			RecordType: RecordTypePTR,
			Timeout:    10 * time.Second,
		}
		source, err := factory.CreateSource(ctx, lookupsource.CreateSettings{}, cfg)
		require.NoError(t, err)

		// Test IPv6 localhost (::1)
		result, found, err := source.Lookup(ctx, "::1")
		require.NoError(t, err)
		if found {
			assert.NotEmpty(t, result)
			t.Logf("::1 PTR resolved to: %v", result)
		} else {
			t.Log("::1 reverse lookup not found (may not be configured)")
		}
	})

	t.Run("PTR record - Google DNS IPv6 reverse lookup", func(t *testing.T) {
		cfg := &Config{
			RecordType: RecordTypePTR,
			Timeout:    10 * time.Second,
		}
		source, err := factory.CreateSource(ctx, lookupsource.CreateSettings{}, cfg)
		require.NoError(t, err)

		// Test Google's public DNS IPv6 address
		result, found, err := source.Lookup(ctx, "2001:4860:4860::8888")
		require.NoError(t, err)
		if found {
			assert.NotEmpty(t, result)
			t.Logf("2001:4860:4860::8888 PTR resolved to: %v", result)
			// Should resolve to dns.google or similar
			assert.Contains(t, result.(string), "google")
		} else {
			t.Log("Google DNS IPv6 reverse lookup not found (may require network access)")
		}
	})

	t.Run("PTR record - IPv6 address format validation", func(t *testing.T) {
		cfg := &Config{
			RecordType: RecordTypePTR,
			Timeout:    10 * time.Second,
		}
		source, err := factory.CreateSource(ctx, lookupsource.CreateSettings{}, cfg)
		require.NoError(t, err)

		testCases := []struct {
			ipv6 string
			desc string
		}{
			{"::1", "Loopback"},
			{"fe80::1", "Link-local"},
			{"2001:db8::1", "Documentation prefix"},
			{"2001:4860:4860::8888", "Google DNS full address"},
			{"2001:4860:4860::8844", "Google DNS alternate"},
			{"::ffff:192.0.2.1", "IPv4-mapped IPv6"},
		}

		for _, tc := range testCases {
			t.Run(tc.desc, func(t *testing.T) {
				result, found, err := source.Lookup(ctx, tc.ipv6)
				// We don't assert found or specific result, just that it doesn't panic
				// and handles the IPv6 format correctly
				if err != nil {
					t.Logf("%s PTR lookup error (expected for some addresses): %v", tc.ipv6, err)
				} else if found {
					t.Logf("%s PTR resolved to: %v", tc.ipv6, result)
				} else {
					t.Logf("%s PTR not found (expected for some addresses)", tc.ipv6)
				}
				// No panic means IPv6 is handled correctly
			})
		}
	})

	t.Run("AAAA vs A record type filtering", func(t *testing.T) {
		// Test that AAAA returns only IPv6, A returns only IPv4
		testCases := []struct {
			recordType      RecordType
			hostname        string
			expectIPv6Only  bool
			expectIPv4Only  bool
		}{
			{RecordTypeA, "dns.google", false, true},
			{RecordTypeAAAA, "dns.google", true, false},
		}

		for _, tc := range testCases {
			t.Run(string(tc.recordType), func(t *testing.T) {
				cfg := &Config{
					RecordType: tc.recordType,
					Timeout:    10 * time.Second,
				}
				source, err := factory.CreateSource(ctx, lookupsource.CreateSettings{}, cfg)
				require.NoError(t, err)

				result, found, err := source.Lookup(ctx, tc.hostname)
				if err != nil {
					t.Skipf("Network not accessible: %v", err)
				}
				if found {
					resultStr := result.(string)
					t.Logf("%s %s resolved to: %v", tc.hostname, tc.recordType, resultStr)

					if tc.expectIPv6Only {
						assert.Contains(t, resultStr, ":", "AAAA record should contain IPv6 addresses with colons")
					}
					if tc.expectIPv4Only {
						assert.Contains(t, resultStr, ".", "A record should contain IPv4 addresses with dots")
						assert.NotContains(t, resultStr, ":", "A record should not contain IPv6 addresses")
					}
				}
			})
		}
	})
}

func TestMultipleResults(t *testing.T) {
	factory := NewFactory()
	cfg := &Config{
		RecordType:      RecordTypeA,
		Timeout:         10 * time.Second,
		MultipleResults: true,
	}

	source, err := factory.CreateSource(context.Background(), lookupsource.CreateSettings{}, cfg)
	require.NoError(t, err)

	ctx := context.Background()

	// Test domain that might have multiple A records
	t.Run("multiple results as comma-separated", func(t *testing.T) {
		result, found, err := source.Lookup(ctx, "localhost")
		require.NoError(t, err)
		if found {
			resultStr, ok := result.(string)
			require.True(t, ok)
			assert.NotEmpty(t, resultStr)
			t.Logf("Multiple results: %v", resultStr)
		}
	})
}

func TestTimeout(t *testing.T) {
	factory := NewFactory()
	cfg := &Config{
		RecordType: RecordTypeA,
		Timeout:    1 * time.Nanosecond, // Very short timeout
	}

	source, err := factory.CreateSource(context.Background(), lookupsource.CreateSettings{}, cfg)
	require.NoError(t, err)

	ctx := context.Background()

	// This should timeout quickly
	_, _, err = source.Lookup(ctx, "example.com")
	// We expect either a timeout error or it completes very quickly
	// Don't assert on error as DNS might be cached
	if err != nil {
		t.Logf("Got expected timeout or error: %v", err)
	}
}

func TestLongTimeout(t *testing.T) {
	factory := NewFactory()

	t.Run("DNS resolution with timeout greater than 5 seconds - unreachable resolver", func(t *testing.T) {
		// Use a DNS server that doesn't exist or is unreachable
		// This will cause the lookup to wait up to the timeout period, simulating slow DNS
		cfg := &Config{
			RecordType: RecordTypeA,
			Timeout:    7 * time.Second, // Greater than 5 seconds
			Resolver:   "192.0.2.1:53",  // TEST-NET-1, reserved documentation IP that is unreachable
		}

		source, err := factory.CreateSource(context.Background(), lookupsource.CreateSettings{}, cfg)
		require.NoError(t, err)

		ctx := context.Background()
		start := time.Now()

		// This should timeout after ~7 seconds trying to reach the unreachable resolver
		// This actually simulates a DNS resolution that takes more than 5 seconds
		_, found, err := source.Lookup(ctx, "example.com")
		elapsed := time.Since(start)

		t.Logf("Lookup took %v with unreachable resolver", elapsed)

		// Should timeout or fail, and should take several seconds attempting
		if err != nil || !found {
			// Expected - the resolver is unreachable
			// Verify it took a significant amount of time (at least 7 seconds)
			assert.GreaterOrEqual(t, elapsed.Seconds(), 7.0,
				"DNS resolution should take at least 7 seconds with unreachable resolver")
			assert.LessOrEqual(t, elapsed.Seconds(), 8.0,
				"DNS resolution should timeout around the configured 7 second timeout")
			t.Logf("Successfully tested >7 second DNS resolution behavior")
		}
	})
}

func TestCustomResolver(t *testing.T) {
	factory := NewFactory()
	cfg := &Config{
		RecordType: RecordTypeA,
		Timeout:    10 * time.Second,
		Resolver:   "8.8.8.8:53", // Google's public DNS
	}

	source, err := factory.CreateSource(context.Background(), lookupsource.CreateSettings{}, cfg)
	require.NoError(t, err)

	ctx := context.Background()

	// Test with custom resolver
	t.Run("resolve with custom DNS server", func(t *testing.T) {
		// Use a well-known domain that should resolve
		result, found, err := source.Lookup(ctx, "google.com")
		if err != nil {
			t.Skipf("Skipping test, custom resolver not accessible: %v", err)
		}
		if found {
			assert.NotEmpty(t, result)
			t.Logf("google.com resolved to: %v", result)
		}
	})
}

func TestRecordTypeSwitching(t *testing.T) {
	ctx := context.Background()
	factory := NewFactory()

	t.Run("A record then PTR record", func(t *testing.T) {
		// First, do A record lookup
		forwardCfg := &Config{
			RecordType: RecordTypeA,
			Timeout:    10 * time.Second,
		}
		forwardSource, err := factory.CreateSource(ctx, lookupsource.CreateSettings{}, forwardCfg)
		require.NoError(t, err)

		ip, found, err := forwardSource.Lookup(ctx, "localhost")
		require.NoError(t, err)
		if !found {
			t.Skip("localhost not found, skipping reverse lookup test")
		}

		ipStr, ok := ip.(string)
		require.True(t, ok)
		t.Logf("A record lookup: localhost -> %s", ipStr)

		// Now do PTR record lookup on that IP
		reverseCfg := &Config{
			RecordType: RecordTypePTR,
			Timeout:    10 * time.Second,
		}
		reverseSource, err := factory.CreateSource(ctx, lookupsource.CreateSettings{}, reverseCfg)
		require.NoError(t, err)

		hostname, found, err := reverseSource.Lookup(ctx, ipStr)
		require.NoError(t, err)
		if found {
			t.Logf("PTR record lookup: %s -> %v", ipStr, hostname)
			assert.NotEmpty(t, hostname)
		}
	})
}
