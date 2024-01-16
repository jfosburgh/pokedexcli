package pokecache

import (
	"sync"
	"time"
)

type Cache struct {
	cache map[string]cacheEntry
	mu    *sync.Mutex
}

type cacheEntry struct {
	createdAt time.Time
	val       []byte
}

func (c Cache) Add(key string, val []byte) {
	c.mu.Lock()
	c.cache[key] = cacheEntry{time.Now(), val}
	c.mu.Unlock()
	return
}

func (c Cache) Get(key string) ([]byte, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	entry, ok := c.cache[key]
	return entry.val, ok
}

func (c Cache) reapLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		<-ticker.C
		for name, entry := range c.cache {
			if entry.createdAt.Add(interval).Before(time.Now()) {
				c.mu.Lock()
				delete(c.cache, name)
				c.mu.Unlock()
			}
		}
	}

}

func NewCache(interval time.Duration) Cache {
	mu := sync.Mutex{}
	c := Cache{cache: make(map[string]cacheEntry, 0), mu: &mu}
	go c.reapLoop(interval)
	return c
}
