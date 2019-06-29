package lrucache

import (
	"fmt"
	"time"
)

const (
	Expired = 5*time.Second
)

func expectEntry(c *lrucache.LruCache, key interface{}) {
	result, ok := c.Get(key)
	fmt.Printf("key:%v,result:%v,ok:%v\n", key, result, ok)
	if !ok {
		fmt.Printf("key:%v not exit\n", key)
	}
}

func keys(c *lrucache.LruCache) {
	s := c.Keys()
	for k, v := range s {
		fmt.Printf("k:%v,v:%v,sv:%s\n", k, v, v.(string))
	}
}

func contains(c *lrucache.LruCache, key interface{}) {
	b := c.Contains(key)
	if b {
		fmt.Println("contains key:", key)
	} else {
		fmt.Println("not contains key:", key)
	}
}

func TestCache() {
	var m time.Time = time.Now()
	c := lrucache.NewLRUCache(4)
	c.Put("elem1", "1", Expired)
	c.Put("elem2", "2", 5*time.Second)
	c.Put("elem3", "3", Expired)
	c.Put("elem4", "4", 5*time.Second)
	c.Put("elem5", "5", 5*time.Second)
	expectEntry(c, "elem1")
	expectEntry(c, "elem3")
	c.Remove("elem4")
	keys(c)
	fmt.Println(c.Len())
	time.Sleep(10 * time.Second)
	var n time.Time = time.Now()
	if n.After(m) {
		fmt.Println("pass after 5")
	}
	expectEntry(c, "elem3")
	fmt.Println(c.Len())
	expectEntry(c, "elem2")
	fmt.Println(c.Len())
	keys(c)
	contains(c, "elem5")
	c.Flush()
	contains(c, "elem5")
}
