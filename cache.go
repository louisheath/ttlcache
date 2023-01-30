// Package ttlcache implements an in-memory TTL cache,
// holding entries in a thread-safe map for a configurable
// amount of time.
package ttlcache

import (
	"fmt"
	"sync"
	"time"
)

type Cache[K comparable, V any] struct {
	mu   sync.RWMutex
	stop chan struct{}

	name     string
	entries  map[K]V
	expiries []expiryEntry[K]

	ttl        time.Duration
	gcInterval time.Duration
	maxSize    int
}

type expiryEntry[K comparable] struct {
	key    K
	expiry time.Time
}

func New[K comparable, V any](
	name string, cfg *Config,
) (*Cache[K, V], error) {
	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid cache config: %w", err)
	}
	c := &Cache[K, V]{
		name:       name,
		entries:    map[K]V{},
		ttl:        cfg.TTL,
		gcInterval: cfg.GCInterval,
		maxSize:    cfg.MaxSize,
	}
	go c.garbageCollectForever()
	return c, nil
}

func (c *Cache[K, V]) Get(k K) (V, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	v, ok := c.entries[k]
	return v, ok
}

func (c *Cache[K, V]) Set(k K, v V) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries[k] = v
	c.expiries = append(c.expiries, expiryEntry[K]{
		key:    k,
		expiry: time.Now().Add(c.ttl),
	})

	if c.maxSize > 0 && len(c.entries) > c.maxSize {
		delete(c.entries, c.expiries[0].key)
		c.expiries = c.expiries[1:]
	}
}

// Delete removes the provided key from the cache, but
// the key will remain in the TTL queue, therefore its
// memory footprint is still non-zero until TTL expiry.
func (c *Cache[K, V]) Delete(k K) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.entries, k)
}

func (c *Cache[K, V]) StopGarbageCollection() {
	close(c.stop)
}

func (c *Cache[K, V]) garbageCollectForever() {
	ticker := time.NewTicker(c.gcInterval)
	go func() {
		for {
			select {
			case <-ticker.C:
				c.garbageCollect(time.Now())
			case <-c.stop:
				ticker.Stop()
				return
			}
		}
	}()
}

func (c *Cache[K, V]) garbageCollect(now time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()

	i := 0
	for ; i < len(c.expiries); i++ {
		if c.expiries[i].expiry.Before(now) {
			delete(c.entries, c.expiries[i].key)
			continue
		}
		break
	}

	c.expiries = c.expiries[i:]
}
