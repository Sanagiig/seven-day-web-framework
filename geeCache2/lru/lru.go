package lru

import (
	"container/list"
	"fmt"
)

type entry struct {
	key   string
	value Value
}

type Value interface {
	Len() int
}

type Cache struct {
	maxBytes  int64
	nBytes    int64
	ll        *list.List
	cache     map[string]*list.Element
	OnEvicted func(key string, value Value)
}

func New(maxBytes int64, OnEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		nBytes:    0,
		ll:        &list.List{},
		cache:     make(map[string]*list.Element),
		OnEvicted: OnEvicted,
	}
}

func (c *Cache) Add(key string, value Value) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		c.nBytes += int64(value.Len()) - int64(kv.value.Len())
		kv.value = value
	} else {
		kv := &entry{key: key, value: value}
		c.ll.PushFront(&list.Element{Value: kv})
		c.nBytes += int64(value.Len())
	}

	if c.nBytes > c.maxBytes {
		c.RemoveOldest()
	}
}

func (c *Cache) Get(key string) (Value, bool) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		return ele.Value.(entry).value, true
	}
	return nil, false
}

func (c *Cache) RemoveOldest() {
	for {
		if c.nBytes <= c.maxBytes {
			break
		}

		ele := c.ll.Back()
		fmt.Printf("%T\n", ele)
		fmt.Printf("%T\n", ele.Value)
		kv := ele.Value.(*entry)
		c.nBytes -= int64(kv.value.Len())
		delete(c.cache, kv.key)
		c.ll.Remove(ele)
	}
}

func (c *Cache) Len() int {
	return c.ll.Len()
}
