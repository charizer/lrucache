# lrucache

## 功能：golang实现的包含过期时间线程安全的lru(Least recently used，最近最少使用)缓存，具有以下特点：

##### 1、新数据插入到list头部，每当缓存命中时，则将数据移动到头部

##### 2、当数据达到最大长度时，则自动淘汰list尾部数据

##### 3、数据可以设置过期时间，获取数据时，如果数据已经过期，则自动删除数据，并返回数据不命中

##### 4、可以设置淘汰数据时的回调函数

### 基本数据结构:

```
type LruCache struct {
    //缓存容量
	size      int
	//缓存list节点
	evictList *list.List
	//缓存元素
	cache     map[interface{}]*list.Element
	//元素过期时间
	ttl       time.Duration
	//淘汰元素回调函数
	onEvict   EvictCallback
	//读写锁
	lock      sync.RWMutex
}
```

### 使用示例:
```
//过期时间
const (
	Expired = 60 * time.Second
)
//数据淘汰时的回调函数
onEvicted := func(k interface{}, v interface{}) {
		
}
//创建lrucache
l, err := NewLRUCache(16, Expired, onEvicted)
if err != nil {
	log.Errorf("err: %v", err)
	return
}
//存放32个数据
for i := 0; i < 32; i++ {
	l.Put(i, i, Expired)
}
//长度为16，0—15已经被淘汰
log.Infof("len:%d",l.Len())
l.Get(2)     //没有此元素，已经被淘汰
l.Get(30)    //有此元素
```

