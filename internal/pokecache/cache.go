package pokecache

import (
	"sync"
	"time"
)

type cacheEntry struct {
	lastAccessed time.Time
	val          []byte
}

type Cache struct {
	mu       sync.Mutex
	entries  map[string]cacheEntry
	interval time.Duration
}

// NewCache creates a new cache and starts the reap loop
func NewCache(interval time.Duration) *Cache {
	c := &Cache{
		entries:  make(map[string]cacheEntry),
		interval: interval,
	}

	go c.reapLoop()
	return c
}

// Add inserts or updates an entry in the cache
func (c *Cache) Add(key string, val []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries[key] = cacheEntry{
		lastAccessed: time.Now(),
		val:          val,
	}
}

func (c *Cache) Get(key string) ([]byte, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, exists := c.entries[key]
	if !exists {
		return nil, false
	}

	entry.lastAccessed = time.Now()
	c.entries[key] = entry
	
	return entry.val, true
}

func (c *Cache) reapLoop() {
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		for k, e := range c.entries {
			if time.Since(e.lastAccessed) > c.interval {
				delete(c.entries, k)
			}
		}
		c.mu.Unlock()
	}
}
