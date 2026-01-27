// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package lookupsource

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCache(t *testing.T) {
	cfg := CacheConfig{
		Enabled: true,
		Size:    100,
		TTL:     5 * time.Minute,
	}
	cache := NewCache(cfg)
	require.NotNil(t, cache)
}

func TestCacheStub(t *testing.T) {
	// The cache now has a real implementation
	cache := NewCache(CacheConfig{Enabled: true, Size: 100})

	// Get returns not found initially
	val, found := cache.Get("key")
	assert.False(t, found)
	assert.Nil(t, val)

	// Set a value
	cache.Set("key", "value")

	// Now it should be found
	val, found = cache.Get("key")
	assert.True(t, found)
	assert.Equal(t, "value", val)

	// Clear removes all entries
	cache.Clear()

	// After clear, should not be found
	val, found = cache.Get("key")
	assert.False(t, found)
	assert.Nil(t, val)
}

func TestWrapWithCacheDisabled(t *testing.T) {
	lookupCount := 0
	baseFn := func(_ context.Context, key string) (any, bool, error) {
		lookupCount++
		return "value-" + key, true, nil
	}

	// Disabled cache should not wrap
	cache := NewCache(CacheConfig{Enabled: false})
	wrappedFn := WrapWithCache(cache, baseFn)

	// Multiple calls should all hit the base function
	_, _, _ = wrappedFn(t.Context(), "key1")
	_, _, _ = wrappedFn(t.Context(), "key1")
	_, _, _ = wrappedFn(t.Context(), "key1")

	assert.Equal(t, 3, lookupCount)
}

func TestWrapWithCacheNil(t *testing.T) {
	lookupCount := 0
	baseFn := func(_ context.Context, key string) (any, bool, error) {
		lookupCount++
		return "value-" + key, true, nil
	}

	// Nil cache should not wrap
	wrappedFn := WrapWithCache(nil, baseFn)

	val, found, err := wrappedFn(t.Context(), "key1")
	require.NoError(t, err)
	require.True(t, found)
	assert.Equal(t, "value-key1", val)
	assert.Equal(t, 1, lookupCount)
}

func TestWrapWithCacheEnabled(t *testing.T) {
	lookupCount := 0
	baseFn := func(_ context.Context, key string) (any, bool, error) {
		lookupCount++
		return "value-" + key, true, nil
	}

	cache := NewCache(CacheConfig{Enabled: true, Size: 100})
	wrappedFn := WrapWithCache(cache, baseFn)

	// First call should hit base function
	val, found, err := wrappedFn(context.Background(), "key1")
	require.NoError(t, err)
	require.True(t, found)
	assert.Equal(t, "value-key1", val)
	assert.Equal(t, 1, lookupCount)

	// Second call should hit cache
	val, found, err = wrappedFn(context.Background(), "key1")
	require.NoError(t, err)
	require.True(t, found)
	assert.Equal(t, "value-key1", val)
	assert.Equal(t, 1, lookupCount) // Should still be 1 because cache was hit

	// Different key should hit base function
	val, found, err = wrappedFn(context.Background(), "key2")
	require.NoError(t, err)
	require.True(t, found)
	assert.Equal(t, "value-key2", val)
	assert.Equal(t, 2, lookupCount)
}

func TestCacheTTL(t *testing.T) {
	cache := NewCache(CacheConfig{
		Enabled: true,
		Size:    100,
		TTL:     100 * time.Millisecond,
	})

	// Set a value
	cache.Set("key", "value")

	// Should be found immediately
	val, found := cache.Get("key")
	assert.True(t, found)
	assert.Equal(t, "value", val)

	// Wait for TTL to expire
	time.Sleep(150 * time.Millisecond)

	// Should not be found after expiration
	val, found = cache.Get("key")
	assert.False(t, found)
	assert.Nil(t, val)
}

func TestCacheNegativeTTL(t *testing.T) {
	cache := NewCache(CacheConfig{
		Enabled:     true,
		Size:        100,
		NegativeTTL: 100 * time.Millisecond,
	})

	// Set a negative (nil) value
	cache.Set("notfound", nil)

	// Should not be found but should be in cache
	val, found := cache.Get("notfound")
	assert.False(t, found)
	assert.Nil(t, val)

	// Wait for negative TTL to expire
	time.Sleep(150 * time.Millisecond)

	// Should still not be found after expiration
	val, found = cache.Get("notfound")
	assert.False(t, found)
	assert.Nil(t, val)
}

func TestCacheLRUEviction(t *testing.T) {
	cache := NewCache(CacheConfig{
		Enabled: true,
		Size:    3, // Small size to test eviction
	})

	// Add 3 entries (fills cache)
	cache.Set("key1", "value1")
	cache.Set("key2", "value2")
	cache.Set("key3", "value3")

	// All should be found
	val, found := cache.Get("key1")
	assert.True(t, found)
	assert.Equal(t, "value1", val)

	// Add 4th entry (should evict oldest: key1)
	cache.Set("key4", "value4")

	// key1 should be evicted
	val, found = cache.Get("key1")
	assert.False(t, found)

	// Others should still be there
	val, found = cache.Get("key2")
	assert.True(t, found)
	assert.Equal(t, "value2", val)

	val, found = cache.Get("key4")
	assert.True(t, found)
	assert.Equal(t, "value4", val)
}

func TestCacheConcurrency(t *testing.T) {
	cache := NewCache(CacheConfig{
		Enabled: true,
		Size:    1000,
	})

	// Test concurrent reads and writes
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 100; j++ {
				key := "key" + string(rune('0'+id))
				cache.Set(key, id)
				cache.Get(key)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}
