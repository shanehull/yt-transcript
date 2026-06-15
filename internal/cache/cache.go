// Package cache provides a simple in-memory TTL cache for transcript segments.
package cache

import (
	"sync"
	"time"

	yt "github.com/shanehull/yt-transcript"
)

type entry struct {
	segs      []yt.Segment
	expiresAt time.Time
}

// Cache stores transcript segments keyed by video_id:lang with a TTL.
type Cache struct {
	mu    sync.RWMutex
	items map[string]entry
	ttl   time.Duration
}

// New returns a Cache with the given TTL.
func New(ttl time.Duration) *Cache {
	return &Cache{
		items: make(map[string]entry),
		ttl:   ttl,
	}
}

// Get returns the cached segments and true, or nil,false if missing or expired.
func (c *Cache) Get(key string) ([]yt.Segment, bool) {
	c.mu.RLock()
	e, ok := c.items[key]
	c.mu.RUnlock()
	if !ok || time.Now().After(e.expiresAt) {
		return nil, false
	}
	return e.segs, true
}

// Set stores segments under key with the cache's TTL.
func (c *Cache) Set(key string, segs []yt.Segment) {
	c.mu.Lock()
	c.items[key] = entry{segs: segs, expiresAt: time.Now().Add(c.ttl)}
	c.mu.Unlock()
}
