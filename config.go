package ttlcache

import (
	"fmt"
	"time"
)

type Config struct {
	// TTL specifies how long a cache entry will remain until
	// automatically removed.
	TTL time.Duration
	// GCInterval specifies how regularly the cache will be
	// checked for expired entries.
	GCInterval time.Duration
	// MaxSize specifies the maximum entries in the cache. If
	// exceeded, the oldest entry is removed from the cache.
	// No limit: 0.
	MaxSize int
}

var DefaultConfig = &Config{
	TTL:        30 * time.Second,
	GCInterval: 5 * time.Second,
	MaxSize:    0, // Unlimited
}

func (c *Config) validate() error {
	if c == nil {
		return fmt.Errorf("nil")
	}
	if c.TTL < 1 {
		return fmt.Errorf("non-positive ttl")
	}
	if c.GCInterval < 1 {
		return fmt.Errorf("non-positive gc interval")
	}
	return nil
}
