package utils

import (
	"sync"
	"time"
)

type CacheItem struct {
	Value      interface{}
	Expiration int64
}

type TTLMemoryCache struct {
	items map[string]CacheItem
	mutex sync.RWMutex
}

var GlobalCache = &TTLMemoryCache{
	items: make(map[string]CacheItem),
}

// Set adds an item to the cache with a specified TTL in seconds.
func (c *TTLMemoryCache) Set(key string, value interface{}, ttlSeconds int64) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.items[key] = CacheItem{
		Value:      value,
		Expiration: time.Now().Unix() + ttlSeconds,
	}
}

// Get retrieves an item. Returns value and true if cleanly found. False if expired/missing.
func (c *TTLMemoryCache) Get(key string) (interface{}, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	item, found := c.items[key]
	if !found {
		return nil, false
	}

	if time.Now().Unix() > item.Expiration {
		return nil, false // Expired
	}

	return item.Value, true
}
