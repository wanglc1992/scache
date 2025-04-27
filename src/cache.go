package src

import (
	"sync"
	"time"
)

const (
	maxCacheSize      = 1024 //最大元素数量
	defaultExpiration = 3600 * 24 * 30
	cleanInterval     = time.Second * 60 * 5
	expirationNon     = -1 //永久有效
	expirationDefault = 0  //使用默认有效期
)

type Cache interface {
	Get(key string) (interface{}, bool)
	Set(key string, value interface{})
	SetWithExpiration(key string, value interface{}, expiration int64)
	SetNX(key string, value interface{}) bool
	SetNXWithExpiration(key string, value interface{}, expiration int64) bool
	Delete(key string)
	Count() int64
	StopCleanup()
	Flush()
}

type MyCache struct {
	mu            sync.Mutex
	cache         map[string]*Item
	head          *Item
	tail          *Item
	stopCh        chan struct{}
	cleanInterval time.Duration
}

type Item struct {
	Key        string
	Value      interface{}
	Expiration int64
	Pre        *Item
	Next       *Item
}

func NewMyCache() *MyCache {
	c := &MyCache{
		cache:         make(map[string]*Item),
		head:          &Item{},
		tail:          &Item{},
		cleanInterval: cleanInterval,
	}
	c.head.Next = c.tail
	c.tail.Pre = c.head
	c.Run()
	return c
}

func (c *MyCache) Get(key string) (interface{}, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if item, ok := c.cache[key]; ok {
		if !c.isExpired(item) {
			c.moveToHead(item)
			return item.Value, true
		} else {
			c._delete(item)
			return nil, false
		}
	}
	return nil, false
}
func (c *MyCache) Set(key string, value interface{}) {
	c.SetWithExpiration(key, value, expirationNon)
}
func (c *MyCache) SetWithExpiration(key string, value interface{}, expiration int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	exp := c.genExpUnix(expiration)
	if item, ok := c.cache[key]; ok {
		item.Value = value
		item.Expiration = exp
		c.moveToHead(item)
		return
	}
	c.add(key, value, exp)
	return
}

func (c *MyCache) add(key string, value interface{}, exp int64) {
	if c.Count() == maxCacheSize {
		c._delete(c.tail.Pre)
	}
	item := &Item{
		Key:        key,
		Value:      value,
		Expiration: exp,
	}
	c.cache[key] = item
	c.addToHead(item)
	return
}
func (c *MyCache) SetNX(key string, value interface{}) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, ok := c.cache[key]; ok {
		return false
	}
	exp := c.genExpUnix(expirationNon)
	c.add(key, value, exp)
	return true
}
func (c *MyCache) SetNXWithExpiration(key string, value interface{}, expiration int64) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, ok := c.cache[key]; ok {
		return false
	}
	exp := c.genExpUnix(expiration)
	c.add(key, value, exp)
	return true
}

func (c *MyCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if item, ok := c.cache[key]; ok {
		c._delete(item)
		return
	}
}

func (c *MyCache) _delete(item *Item) {
	delete(c.cache, item.Key)
	c.removeItem(item)
	return
}
func (c *MyCache) Count() int64 {
	return int64(len(c.cache))
}

func (c *MyCache) moveToHead(item *Item) {
	c.removeItem(item)
	c.addToHead(item)
}

func (c *MyCache) removeItem(item *Item) {
	item.Pre.Next = item.Next
	item.Next.Pre = item.Pre
}

func (c *MyCache) addToHead(item *Item) {
	item.Next = c.head.Next
	item.Pre = c.head
	c.head.Next.Pre = item
	c.head.Next = item
}

func (c *MyCache) isExpired(item *Item) bool {
	if item.Expiration < 0 {
		return false
	}
	return time.Now().Unix() > item.Expiration
}

func (c *MyCache) genExpUnix(exp int64) int64 {

	if exp < 0 {
		return expirationNon
	} else if exp == expirationDefault {
		return time.Now().Unix() + defaultExpiration
	} else {
		return time.Now().Unix() + exp
	}
}

func (c *MyCache) Cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, item := range c.cache {
		if c.isExpired(item) {
			c._delete(item)
		}
	}
}

func (c *MyCache) StopCleanup() {
	c._stop()
	return
}
func (c *MyCache) Flush() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache = make(map[string]*Item)
	c.head = &Item{}
	c.tail = &Item{}
	c.head.Next = c.tail
	c.tail.Pre = c.head
	return
}
