package lrucache

import (
	"container/list"
	"errors"
	"sync"
	"time"
)

// EvictCallback is used to get a callback when a cache entry is evicted
type EvictCallback func(key interface{}, value interface{})

// LruCache implements a thread safe fixed size Expire LRU cache
type LruCache struct {
	size      int
	evictList *list.List
	cache     map[interface{}]*list.Element
	ttl       time.Duration
	onEvict   EvictCallback
	lock      sync.RWMutex
}

// entry is used to hold a value in the evictList
type entry struct {
	key   interface{}
	value interface{}
	//if tll is nil, entry is not expire auto
	ttl *time.Time
}

func (e *entry) IsExpired() bool {
	if e.ttl == nil {
		return false
	}
	return time.Now().After(*e.ttl)
}

// NewLRUCache creates an expiring cache with the given size
func NewLRUCache(maxSize int, ttl time.Duration, onEvict EvictCallback) (*LruCache, error) {
	if maxSize <= 0 {
		return nil, errors.New("Must provide a positive size to cache")
	}
	c := &LruCache{
		size:      maxSize,
		evictList: list.New(),
		cache:     make(map[interface{}]*list.Element),
		ttl:       ttl,
		onEvict:   onEvict,
	}
	return c, nil
}

// Get a key's value from the cache.
func (c *LruCache) Get(key interface{}) (value interface{}, ok bool) {
	c.lock.Lock()
	defer c.lock.Unlock()
	//exsit
	if ent, ok := c.cache[key]; ok {
		//expired
		if ent.Value.(*entry).IsExpired() {
			c.removeElement(ent)
			return nil, false
		}
		//not expired,movetofront
		c.evictList.MoveToFront(ent)
		return ent.Value.(*entry).value, true
	}
	return nil, false
}

// removeElement is used to remove a given list element from the cache
func (c *LruCache) removeElement(e *list.Element) {
	c.evictList.Remove(e)
	kv := e.Value.(*entry)
	delete(c.cache, kv.key)
	if c.onEvict != nil {
		c.onEvict(kv.key, kv.value)
	}
}

// Add adds the value to the cache at key with the specified maximum duration.
func (c *LruCache) Put(key interface{}, value interface{}, ttl time.Duration) bool {
	c.lock.Lock()
	defer c.lock.Unlock()
	var ex *time.Time = nil
	if ttl > 0 {
		expire := time.Now().Add(ttl)
		ex = &expire
	} else if c.ttl > 0 {
		expire := time.Now().Add(c.ttl)
		ex = &expire
	}
	//Check for existing item
	if ent, ok := c.cache[key]; ok {
		c.evictList.MoveToFront(ent)
		ent.Value.(*entry).value = value
		ent.Value.(*entry).ttl = ex
		return false
	}
	// Add new item
	ent := &entry{
		key:   key,
		value: value,
		ttl:   ex,
	}
	entry := c.evictList.PushFront(ent)
	c.cache[key] = entry
	evict := c.evictList.Len() > c.size
	// Verify size not exceeded
	if evict {
		c.removeOldest()
	}
	return evict
}

// removeOldest removes the oldest item from the cache
func (c *LruCache) removeOldest() {
	ent := c.evictList.Back()
	if ent != nil {
		c.removeElement(ent)
	}
}

// Len returns the number of items in the cache.
func (c *LruCache) Len() int {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.evictList.Len()
}

// Remove removes the provided key from the cache.
func (c *LruCache) Remove(key interface{}) bool {
	c.lock.Lock()
	defer c.lock.Unlock()
	if ent, ok := c.cache[key]; ok {
		c.removeElement(ent)
		return true
	}
	return false
}

// Contains Check if a key exsists in cache without updating the recent-ness.
func (c *LruCache) Contains(key interface{}) (ok bool) {
	c.lock.RLock()
	defer c.lock.RUnlock()
	if ent, ok := c.cache[key]; ok {
		if ent.Value.(*entry).IsExpired() {
			return false
		}
		return ok
	}
	return false
}

// Keys return all the keys in cache, from oldest to newest
func (c *LruCache) Keys() []interface{} {
	c.lock.RLock()
	defer c.lock.RUnlock()
	keys := make([]interface{}, len(c.cache))
	i := 0
	for ent := c.evictList.Back(); ent != nil; ent = ent.Prev() {
		keys[i] = ent.Value.(*entry).key
		i++
	}
	return keys
}

// Clear remove all the keys in cache
func (c *LruCache) Clear() {
	c.lock.Lock()
	defer c.lock.Unlock()
	for k, v := range c.cache {
		if c.onEvict != nil {
			c.onEvict(k, v.Value.(*entry).value)
		}
		delete(c.cache, k)
	}
	c.evictList.Init()
}
