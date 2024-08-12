package geeCache2

import (
	"seven-day-web-framework/geeCache2/lru"
	"sync"
)

type Cache struct {
	lru     lru.Cache
	maxSize int64
	mu      sync.RWMutex
}

func NewCache(maxSize int64) *Cache {
	return &Cache{
		lru:     *lru.New(maxSize, nil),
		maxSize: maxSize,
		mu:      sync.RWMutex{},
	}
}

func (c *Cache) Add(key string, val ByteView) {
	c.mu.Lock()
	c.lru.Add(key, val)
	c.mu.Unlock()
}

func (c *Cache) Get(key string) (ByteView, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	bv, ok := c.lru.Get(key)
	if !ok {
		return ByteView{}, ok
	}
	return bv.(ByteView), ok
}
