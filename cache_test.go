package ttlcache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCache(t *testing.T) {
	t.Parallel()

	cache, err := New[string, string](
		"TestCache", DefaultConfig,
	)
	require.NoError(t, err)

	cache.Set("key", "value")
	require.Len(t, cache.entries, 1)
	require.Len(t, cache.expiries, 1)

	val, ok := cache.Get("key")
	require.True(t, ok)
	require.Equal(t, "value", val)
}

func TestCacheValidatesConfig(t *testing.T) {
	t.Parallel()

	_, err := New[string, string](
		"TestCache", nil,
	)
	require.ErrorContains(t, err, "invalid cache config: nil")
}

func TestCacheMaintainsMaxSize(t *testing.T) {
	t.Parallel()

	cache, err := New[string, string](
		"TestCache", &Config{
			TTL:        DefaultConfig.TTL,
			GCInterval: DefaultConfig.GCInterval,
			MaxSize:    1,
		},
	)
	require.NoError(t, err)

	cache.Set("1", "2")
	require.Len(t, cache.entries, 1)
	require.Len(t, cache.expiries, 1)

	cache.Set("3", "4")
	require.Len(t, cache.entries, 1)
	require.Len(t, cache.expiries, 1)

	val, ok := cache.Get("3")
	require.True(t, ok)
	require.Equal(t, "4", val)

	val, ok = cache.Get("1")
	require.False(t, ok)
	require.Empty(t, val)
}

func TestCacheDelete(t *testing.T) {
	t.Parallel()

	cache, err := New[string, string](
		"TestCache", DefaultConfig,
	)
	require.NoError(t, err)

	cache.Set("key", "value")
	require.Len(t, cache.entries, 1)
	require.Len(t, cache.expiries, 1)

	cache.Delete("key")
	require.Empty(t, cache.entries)
	require.Len(t, cache.expiries, 1)

	val, ok := cache.Get("key")
	require.False(t, ok)
	require.Empty(t, val)
}

func TestCacheGarbageCollection(t *testing.T) {
	t.Parallel()

	cache, err := New[string, string](
		"TestCache", &Config{
			TTL:        200 * time.Millisecond,
			GCInterval: 50 * time.Millisecond,
			MaxSize:    DefaultConfig.MaxSize,
		},
	)
	require.NoError(t, err)

	cache.Set("1", "2")
	require.Len(t, cache.entries, 1)
	require.Len(t, cache.expiries, 1)
	time.Sleep(cache.ttl / 2)

	cache.Set("3", "4")
	require.Len(t, cache.entries, 2)
	require.Len(t, cache.expiries, 2)

	require.Eventually(t, func() bool {
		return len(cache.entries) == 1
	}, time.Second, 20*time.Millisecond)
	require.Len(t, cache.expiries, 1)

	require.Eventually(t, func() bool {
		return len(cache.entries) == 0
	}, time.Second, 20*time.Millisecond)
	require.Empty(t, cache.expiries)
}
