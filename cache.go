package geecaches

import (
	"geecache-s/cachePolicy"
	"sync"
)

type cache struct {
	mut      sync.Mutex
	policy   cachePolicy.Cache
	maxBytes int64
}

func (c *cache) get(key string) (value ByteView, ok bool) {
	c.mut.Lock()
	defer c.mut.Unlock()
	if c.policy == nil {
		return
	}

	if value, ok := c.policy.Get(key); ok {
		return value.(ByteView), true
	}

	return
}

func (c *cache) add(key string, value cachePolicy.Value) error {
	c.mut.Lock()
	defer c.mut.Unlock()

	if c.policy == nil {
		c.policy = cachePolicy.CreateCache(c.maxBytes, cachePolicy.CacheCallBack{}, cachePolicy.LruPolicy)
	}

	if err := c.policy.Add(key, value); err != nil {
		return err
	}

	return nil
}
