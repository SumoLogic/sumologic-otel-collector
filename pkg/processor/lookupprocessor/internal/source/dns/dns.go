// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

// Package dns provides a DNS lookup source for forward and reverse DNS resolution.
package dns // import "github.com/open-telemetry/opentelemetry-collector-contrib/processor/lookupprocessor/internal/source/dns"

import (
	"context"
	"errors"
	"net"
	"strings"
	"time"

	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/lookupprocessor/lookupsource"
)

const sourceType = "dns"

// Mode specifies the DNS resolution mode.
type Mode string

const (
	// ModeForward performs forward DNS lookup (hostname to IP address)
	ModeForward Mode = "forward"
	// ModeReverse performs reverse DNS lookup (IP address to hostname)
	ModeReverse Mode = "reverse"
)

// Config is the configuration for the DNS lookup source.
type Config struct {
	// Mode specifies the DNS resolution mode: "forward" or "reverse".
	// - forward: resolves hostname to IP address(es)
	// - reverse: resolves IP address to hostname(s)
	// Default: "forward"
	Mode Mode `mapstructure:"mode"`

	// Timeout specifies the maximum time to wait for DNS resolution.
	// Default: 5s
	Timeout time.Duration `mapstructure:"timeout"`

	// Resolver specifies the DNS server to use for lookups.
	// If empty, uses the system's default DNS resolver.
	// Format: "host:port" (e.g., "8.8.8.8:53")
	// Optional.
	Resolver string `mapstructure:"resolver"`

	// MultipleResults specifies how to handle multiple DNS results.
	// - If true, returns all results as a comma-separated string
	// - If false, returns only the first result
	// Default: false
	MultipleResults bool `mapstructure:"multiple_results"`
}

// Validate implements lookupsource.SourceConfig.
func (c *Config) Validate() error {
	switch c.Mode {
	case "", ModeForward, ModeReverse:
		// valid
	default:
		return errors.New("mode must be 'forward' or 'reverse'")
	}

	if c.Timeout < 0 {
		return errors.New("timeout must be non-negative")
	}

	if c.Resolver != "" {
		// Basic validation for resolver format
		if !strings.Contains(c.Resolver, ":") {
			return errors.New("resolver must be in 'host:port' format")
		}
	}

	return nil
}

// NewFactory creates a factory for the DNS source.
func NewFactory() lookupsource.SourceFactory {
	return lookupsource.NewSourceFactory(
		sourceType,
		createDefaultConfig,
		createSource,
	)
}

func createDefaultConfig() lookupsource.SourceConfig {
	return &Config{
		Mode:            ModeForward,
		Timeout:         5 * time.Second,
		MultipleResults: false,
	}
}

func createSource(
	_ context.Context,
	settings lookupsource.CreateSettings,
	cfg lookupsource.SourceConfig,
) (lookupsource.Source, error) {
	dnsCfg := cfg.(*Config)

	// Apply defaults
	mode := dnsCfg.Mode
	if mode == "" {
		mode = ModeForward
	}

	timeout := dnsCfg.Timeout
	if timeout == 0 {
		timeout = 5 * time.Second
	}

	// Create resolver
	var resolver *net.Resolver
	if dnsCfg.Resolver != "" {
		// Use custom DNS server with Go's native resolver
		// PreferGo=true ensures we use the pure Go resolver instead of cgo,
		// which provides better control and consistency across platforms
		resolver = &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				d := net.Dialer{
					Timeout: timeout,
				}
				return d.DialContext(ctx, network, dnsCfg.Resolver)
			},
		}
	} else {
		resolver = net.DefaultResolver
	}

	s := &dnsSource{
		mode:            mode,
		timeout:         timeout,
		resolver:        resolver,
		multipleResults: dnsCfg.MultipleResults,
	}

	// Wrap lookup function with cache if enabled
	lookupFunc := s.lookup
	if settings.Cache.Enabled {
		cache := lookupsource.NewCache(settings.Cache)
		cache.SetLogger(settings.TelemetrySettings.Logger)
		lookupFunc = lookupsource.WrapWithCache(cache, lookupFunc)
	}

	return lookupsource.NewSource(
		lookupFunc,
		func() string { return sourceType },
		nil, // no start needed
		nil, // no shutdown needed
	), nil
}

// dnsSource performs DNS lookups.
type dnsSource struct {
	mode            Mode
	timeout         time.Duration
	resolver        *net.Resolver
	multipleResults bool
}

// lookup performs the DNS resolution based on the configured mode.
func (s *dnsSource) lookup(ctx context.Context, key string) (any, bool, error) {
	// Validate input
	if key == "" {
		return nil, false, errors.New("empty lookup key")
	}

	// Create a context with timeout
	lookupCtx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	switch s.mode {
	case ModeForward:
		return s.forwardLookup(lookupCtx, key)
	case ModeReverse:
		return s.reverseLookup(lookupCtx, key)
	default:
		return nil, false, errors.New("invalid mode")
	}
}

// forwardLookup resolves a hostname to IP address(es).
func (s *dnsSource) forwardLookup(ctx context.Context, hostname string) (any, bool, error) {
	// Use LookupIPAddr instead of LookupHost for better IPv6 handling and zone support
	addrs, err := s.resolver.LookupIPAddr(ctx, hostname)
	if err != nil {
		// Check if it's a DNS error indicating not found
		var dnsErr *net.DNSError
		if errors.As(err, &dnsErr) && dnsErr.IsNotFound {
			return nil, false, nil
		}
		// For other errors, return the error
		return nil, false, err
	}

	if len(addrs) == 0 {
		return nil, false, nil
	}

	// Convert IPAddr to string, handling IPv6 zone identifiers properly
	ips := make([]string, len(addrs))
	for i, addr := range addrs {
		ips[i] = addr.String()
	}

	if s.multipleResults {
		return strings.Join(ips, ","), true, nil
	}
	return ips[0], true, nil
}

// reverseLookup resolves an IP address to hostname(s).
func (s *dnsSource) reverseLookup(ctx context.Context, ip string) (any, bool, error) {
	names, err := s.resolver.LookupAddr(ctx, ip)
	if err != nil {
		// Check if it's a DNS error indicating not found
		var dnsErr *net.DNSError
		if errors.As(err, &dnsErr) && dnsErr.IsNotFound {
			return nil, false, nil
		}
		// For other errors, return the error
		return nil, false, err
	}

	if len(names) == 0 {
		return nil, false, nil
	}

	// Clean up trailing dots from PTR records
	for i := range names {
		names[i] = strings.TrimSuffix(names[i], ".")
	}

	if s.multipleResults {
		return strings.Join(names, ","), true, nil
	}
	return names[0], true, nil
}
