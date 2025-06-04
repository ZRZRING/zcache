package lru

import (
	"container/list"
)

type Value interface {
	Len() int
}

// Cache LRU 缓存
type Cache struct {
	maxBytes  int64                    // 允许使用的最大内存
	nbytes    int64                    // 当前已使用的内存
	cache     map[string]*list.Element // 键是字符串，值是双向链表中对应节点的指针
	list      *list.List               // Go 语言标准库实现的双向链表
	OnEvicted func(k string, v Value)  // 某条记录被移除时的回调函数，可以为 nil
}

// entry kv 数据载体
type entry struct {
	key   string // 钦定为 string 类型，方便查找
	value Value  // 需要实现 Len 函数获取数据大小
}

func New(maxBytes int64, onEvicted func(k string, v Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		list:      list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

// removeElement 删除元素
func (c *Cache) removeElement(e *list.Element) {
	c.list.Remove(e)
	kv := e.Value.(*entry)
	c.nbytes -= int64(len(kv.key))
	c.nbytes -= int64(kv.value.Len())
	delete(c.cache, kv.key)
	if c.OnEvicted != nil {
		c.OnEvicted(kv.key, kv.value)
	}
}

// RemoveOldest 移除最近最少访问的节点
func (c *Cache) RemoveOldest() {
	if c.cache == nil {
		return
	}
	e := c.list.Back()
	if e != nil {
		c.removeElement(e)
	}
}

// Remove 删除指定的 key 对应的节点
func (c *Cache) Remove(key string) {
	if c.cache == nil {
		return
	}
	if e, ok := c.cache[key]; ok {
		c.removeElement(e)
	}
}

// Add 添加或更新节点
func (c *Cache) Add(k string, v Value) {
	if c.cache == nil {
		return
	}
	if e, hit := c.cache[k]; hit {
		c.list.MoveToFront(e)
		kv := e.Value.(*entry)
		c.nbytes -= int64(kv.value.Len())
		kv.value = v
		c.nbytes += int64(kv.value.Len())
	} else {
		e = c.list.PushFront(&entry{k, v})
		c.cache[k] = e
	}
	for c.maxBytes != 0 && int64(c.list.Len()) > c.maxBytes {
		c.RemoveOldest()
	}
}

// Get 查找指定的 key 对应的节点
func (c *Cache) Get(k string) (value Value, ok bool) {
	if c.cache == nil {
		return
	}
	if e, hit := c.cache[k]; hit {
		c.list.MoveToFront(e)
		kv := e.Value.(*entry)
		return kv.value, true
	}
	return
}

// Clear 清空缓存
func (c *Cache) Clear() {
	if c.OnEvicted != nil {
		for _, e := range c.cache {
			kv := e.Value.(*entry)
			c.OnEvicted(kv.key, kv.value)
		}
	}
	c.list = nil
	c.cache = nil
}

// Len 返回缓存中节点的数量
func (c *Cache) Len() int {
	if c.cache == nil {
		return 0
	}
	return c.list.Len()
}
