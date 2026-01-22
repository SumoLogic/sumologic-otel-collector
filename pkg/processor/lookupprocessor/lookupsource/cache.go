// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package lookupsource // import "github.com/SumoLogic/sumologic-otel-collector/pkg/processor/lookupprocessor/lookupsource"

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"
)

type CacheConfig struct {
	Enabled bool `mapstructure:"enabled"`

	// Size specifies the maximum number of entries in the cache.
	// Default: 1000
	Size int `mapstructure:"size"`

	// TTL specifies how long successful lookup results are cached.
	// Default: 0 (no expiration)
	TTL time.Duration `mapstructure:"ttl"`

	// NegativeTTL specifies how long failed/not-found lookups are cached.
	// This prevents repeated lookups for non-existent keys.
	// Default: 0 (negative results not cached)
	NegativeTTL time.Duration `mapstructure:"negative_ttl"`
}

// cacheEntry represents a cached value with expiration time.
type cacheEntry struct {
	value      any
	expireAt   time.Time
	isNegative bool // true if this is a cached "not found" result
}

// Cache implements a simple LRU cache with TTL support.
// Uses a map for O(1) lookups and tracks insertion order for LRU eviction.
type Cache struct {
	config  CacheConfig
	mu      sync.RWMutex
	entries map[string]*cacheEntry
	keys    []string // tracks insertion order for LRU
	logger  *zap.Logger
}

func NewCache(cfg CacheConfig) *Cache {
	size := cfg.Size
	if size <= 0 {
		size = 1000 // default
	}
	return &Cache{
		config:  cfg,
		entries: make(map[string]*cacheEntry, size),
		keys:    make([]string, 0, size),
		logger:  zap.NewNop(), // default to no-op logger
	}
}

// SetLogger sets the logger for the cache.
func (c *Cache) SetLogger(logger *zap.Logger) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if logger != nil {
		c.logger = logger
	}
}

func (c *Cache) Get(key string) (any, bool) {
	c.mu.RLock()
	entry, exists := c.entries[key]
	if !exists {
		c.mu.RUnlock()
		return nil, false
	}

	// Check if entry has expired (lazy deletion)
	if !entry.expireAt.IsZero() && time.Now().After(entry.expireAt) {
		c.mu.RUnlock()
		// Upgrade to write lock to remove expired entry
		c.mu.Lock()
		// Double-check entry still exists and is still expired (race condition protection)
		if entry, exists := c.entries[key]; exists && !entry.expireAt.IsZero() && time.Now().After(entry.expireAt) {
			c.logger.Debug("cache entry expired", zap.String("key", key))
			c.removeEntry(key)
		}
		c.mu.Unlock()
		return nil, false
	}

	// Return nil for cached negative results
	if entry.isNegative {
		c.mu.RUnlock()
		c.logger.Debug("cache hit (negative)", zap.String("key", key))
		return nil, false
	}

	c.logger.Debug("cache hit", zap.String("key", key), zap.Any("value", entry.value))
	c.mu.RUnlock()
	return entry.value, true
}

func (c *Cache) Set(key string, value any) {
	c.mu.Lock()
	defer c.mu.Unlock()

	isNegative := value == nil
	ttl := c.config.TTL
	if isNegative {
		ttl = c.config.NegativeTTL
		if ttl == 0 {
			// Don't cache negative results if negative_ttl is 0
			return
		}
	}

	// Calculate expiration time
	var expireAt time.Time
	if ttl > 0 {
		expireAt = time.Now().Add(ttl)
	}

	// If key already exists, update it
	if _, exists := c.entries[key]; exists {
		c.entries[key] = &cacheEntry{
			value:      value,
			expireAt:   expireAt,
			isNegative: isNegative,
		}
		c.logger.Info("cache updated", zap.String("key", key), zap.Bool("negative", isNegative))
		return
	}

	// Evict oldest entry if cache is full
	maxSize := c.config.Size
	if maxSize <= 0 {
		maxSize = 1000
	}
	if len(c.entries) >= maxSize {
		c.evictOldest()
	}

	// Add new entry
	c.entries[key] = &cacheEntry{
		value:      value,
		expireAt:   expireAt,
		isNegative: isNegative,
	}
	c.keys = append(c.keys, key)
	c.logger.Info("cache stored", zap.String("key", key), zap.Bool("negative", isNegative), zap.Duration("ttl", ttl))
}

func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = make(map[string]*cacheEntry, c.config.Size)
	c.keys = make([]string, 0, c.config.Size)
}

// evictOldest removes the oldest entry from the cache (LRU).
// Must be called with lock held.
func (c *Cache) evictOldest() {
	if len(c.keys) == 0 {
		return
	}

	// Remove oldest key
	oldestKey := c.keys[0]
	c.logger.Info("cache evicted oldest entry", zap.String("key", oldestKey))
	c.removeEntry(oldestKey)
}

// removeEntry removes an entry from both the map and keys slice.
// Must be called with write lock held.
func (c *Cache) removeEntry(key string) {
	delete(c.entries, key)

	// Remove key from keys slice
	for i, k := range c.keys {
		if k == key {
			c.keys = append(c.keys[:i], c.keys[i+1:]...)
			break
		}
	}
}

// WrapWithCache wraps a lookup function with caching.
//
// Example:
//
//	cache := lookupsource.NewCache(cfg.Cache)
//	cachedLookup := lookupsource.WrapWithCache(cache, myLookupFunc)
func WrapWithCache(cache *Cache, fn LookupFunc) LookupFunc {
	if cache == nil || !cache.config.Enabled {
		return fn
	}
	return func(ctx context.Context, key string) (any, bool, error) {
		if val, found := cache.Get(key); found {
			return val, true, nil
		}

		val, found, err := fn(ctx, key)
		if err != nil {
			return nil, false, err
		}

		if found {
			cache.Set(key, val)
		}

		return val, found, nil
	}
}
