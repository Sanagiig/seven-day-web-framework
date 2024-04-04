package geeCache

import (
	"geeCache/byteview"
	"geeCache/lru"
	"sync"
)

type Cache struct {
	mu         sync.RWMutex
	lru        *lru.Cache
	cacheBytes int64
}

func NewCache(maxbytes int64) *Cache {
	return &Cache{
		mu:         sync.RWMutex{},
		cacheBytes: maxbytes,
	}
}

func (c *Cache) Add(key string, value byteview.ByteView) {
	c.mu.RLock()
	if c.lru == nil {
		c.lru = lru.New(c.cacheBytes, nil)
	}
	c.lru.Add(key, value)
	c.mu.RUnlock()
}

func (c *Cache) Get(key string) (value byteview.ByteView, ok bool) {
	var v lru.Value
	c.mu.RLock()
	if c.lru != nil {
		v, ok = c.lru.Get(key)
		if ok {
			value = v.(byteview.ByteView)
		}
	}
	c.mu.RUnlock()
	return value, ok
}
