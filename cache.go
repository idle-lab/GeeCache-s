package geecaches

import (
	"geecache-s/cachePolicy"
	"sync"
)

type cache struct {
	mut      sync.Mutex
	cache    cachePolicy.Cache
	policy   cachePolicy.CachePolicy
	maxBytes int64
}

func (c *cache) get(key string) (value ByteView, ok bool) {
	c.mut.Lock()
	defer c.mut.Unlock()
	if c.cache == nil {
		return
	}

	if value, ok := c.cache.Get(key); ok {
		return value.(ByteView), true
	}

	return
}

func (c *cache) add(key string, value cachePolicy.Value) error {
	c.mut.Lock()
	defer c.mut.Unlock()

	if c.cache == nil {
		c.cache = cachePolicy.CreateCache(c.maxBytes, cachePolicy.CacheCallBack{}, c.policy)
	}

	if err := c.cache.Add(key, value); err != nil {
		return err
	}

	return nil
}
