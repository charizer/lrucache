package lrucache

import (
	"container/list"
	"sync"
	"time"
)

//get current time
type Clock struct {
}

func (Clock) Now() time.Time { return time.Now() }

//lrucache
type LruCache struct {
	clock     Clock
	size      int
	evictList *list.List
	cache     map[interface{}]*list.Element
	lock      sync.RWMutex
}

// NewLRUCache creates an expiring cache with the given size
func NewLRUCache(maxSize int) *LruCache {
	c := &LruCache{
		clock:     Clock{},
		size:      maxSize,
		evictList: list.New(),
		cache:     make(map[interface{}]*list.Element),
	}
	return c
}

type entry struct {
	key        interface{}
	value      interface{}
	expireTime time.Time
}

// Get a key's value from the cache.
func (c *LruCache) Get(key interface{}) (value interface{}, ok bool) {
	if c.cache == nil {
		return nil, false
	}
	c.lock.Lock()
	defer c.lock.Unlock()
	//exsit
	if ent, ok := c.cache[key]; ok {
		//expired
		if c.clock.Now().After(ent.Value.(*entry).expireTime) {
			c.removeElement(ent)
			return nil, false
		}
		//not expired,movetofront
		c.evictList.MoveToFront(ent)
		return ent.Value.(*entry).value, true
	}
	return nil, false
}

func (c *LruCache) removeElement(e *list.Element) {
	c.evictList.Remove(e)
	kv := e.Value.(*entry)
	delete(c.cache, kv.key)
}

// Add adds the value to the cache at key with the specified maximum duration.
func (c *LruCache) Put(key interface{}, value interface{}, ttl time.Duration) bool {
	if c.cache == nil {
		return false
	}
	c.lock.Lock()
	defer c.lock.Unlock()
	//Check for existing item
	if ent, ok := c.cache[key]; ok {
		c.evictList.MoveToFront(ent)
		ent.Value.(*entry).value = value
		return true
	}
	// Add new item
	ent := &entry{
		key:        key,
		value:      value,
		expireTime: c.clock.Now().Add(ttl),
	}
	entry := c.evictList.PushFront(ent)
	c.cache[key] = entry
	evict := c.evictList.Len() > c.size
	// Verify size not exceeded
	if evict {
		c.removeOldest()
	}
	return true
}

func (c *LruCache) removeOldest() {
	if c.cache == nil {
		return
	}
	ent := c.evictList.Back()
	if ent != nil {
		c.removeElement(ent)
	}
}

// Len returns the number of items in the cache.
func (c *LruCache) Len() int {
	if c.cache == nil {
		return 0
	}
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.evictList.Len()
}

// Remove removes the provided key from the cache.
func (c *LruCache) Remove(key interface{}) bool {
	if c.cache == nil {
		return false
	}
	c.lock.Lock()
	defer c.lock.Unlock()
	if ent, ok := c.cache[key]; ok {
		c.removeElement(ent)
		return true
	}
	return false
}

// Check if a key exsists in cache without updating the recent-ness.
func (c *LruCache) Contains(key interface{}) (ok bool) {
	if c.cache == nil {
		return false
	}
	c.lock.RLock()
	defer c.lock.RUnlock()
	_, ok = c.cache[key]
	return ok
}

// Keys return all the keys in cache
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

// Flush remove all the keys in cache
func (c *LruCache) Flush() {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.evictList = list.New()
	c.cache = make(map[interface{}]*list.Element)
}

