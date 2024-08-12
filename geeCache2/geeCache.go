package geeCache2

import (
	"sync"
)

var (
	mu          sync.RWMutex
	groups      = make(map[string]GeeCache)
	defaultByte = int64(1024)
)

type GeeCache struct {
	name      string
	getter    LocalGetter
	mainCache Cache
}

func NewGeeCache(name string, maxSize int64, getter LocalGetter) *GeeCache {
	if getter == nil {
		panic("nil getter")
	}

	if _, ok := groups[name]; ok {
		panic("is exist " + name)
	}

	c := &GeeCache{
		name:      name,
		getter:    getter,
		mainCache: *NewCache(maxSize),
	}

	groups[name] = c
	return c
}

func getGeeCacheFromGroup(name string) (GeeCache, bool) {
	cache, ok := groups[name]
	return cache, ok
}

func (c *GeeCache) Get(key string) (ByteView, error) {
	ret, ok := c.mainCache.Get(key)
	if !ok {
		return c.load(key)
	}
	return ret, nil
}

func (c *GeeCache) load(key string) (ByteView, error) {
	b, err := c.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}

	bv := ByteView{b: CloneBytes(b)}
	c.populate(key, bv)
	return ByteView{b: b}, nil
}

func (c *GeeCache) populate(key string, val ByteView) {
	c.mainCache.Add(key, val)
}
