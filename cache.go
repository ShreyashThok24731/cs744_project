package main

import (
	"container/list"
	"sync"
)

type entry struct {
	key   string
	value string
}

type KVCache struct {
	mu       sync.RWMutex
	capacity int
	lru      *list.List                  
	items    map[string]*list.Element
}

func NewKVCache(capacity int) *KVCache {
	return &KVCache{
		capacity: capacity,
		lru:      list.New(),
		items:    make(map[string]*list.Element),
	}
}

func (c *KVCache) Get(key string) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	element, ok := c.items[key]
	if !ok {
		return "", false 
	}
	c.lru.MoveToFront(element)
	return element.Value.(*entry).value, true
}

func (c *KVCache) Put(key, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if element, ok := c.items[key]; ok {
		c.lru.MoveToFront(element)
		element.Value.(*entry).value = value
		return
	}
	if c.lru.Len() >= c.capacity {
		oldest := c.lru.Back()
		if oldest != nil {
			c.lru.Remove(oldest)
			delete(c.items, oldest.Value.(*entry).key)
		}
	}
	newEntry := &entry{key: key, value: value}
	element := c.lru.PushFront(newEntry)
	c.items[key] = element
}

func (c *KVCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if element, ok := c.items[key]; ok {
		c.lru.Remove(element)
		delete(c.items, key)
	}
}
