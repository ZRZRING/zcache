package zcache

import (
	"sync"
	"zcache/lru"
)

// 实例化 LRU，并将操作封装为并发安全的方法
type cache struct {
	mutex      sync.Mutex
	lru        *lru.Cache
	cacheBytes int64
}

func (c *cache) add(key string, value ByteView) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.lru == nil {
		c.lru = lru.New(c.cacheBytes, nil)
	}
	c.lru.Add(key, value)
}

func (c *cache) get(key string) (value ByteView, ok bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.lru == nil {
		return
	}
	if v, hit := c.lru.Get(key); hit {
		return v.(ByteView), true
	}
	return
}
