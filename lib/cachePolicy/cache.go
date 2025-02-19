package cachePolicy

import (
	"container/list"
	"fmt"
)

// Value use Len to count how many bytes it takes
type Value interface {
	Size() int64
}

const (
	LruPolicy = iota
	LfuPolicy
)

type CachePolicy int

// It is not safe for concurrent access.
type Cache interface {
	// Retrn the value corresponding to the key
	Get(key string) (Value, bool)

	// If %key is not exists, insert a (k, v) pair.
	// If %key is in cache, update old value to incoming value.
	// If cache if full, evicted one of (k, v) pairs in cache.
	// When length of value is larger than maxBytes, Add will return a error.
	Add(key string, value Value) error

	// Evict a (k, v) pair.
	Evict()

	// Return number of (k, v) pairs.
	Len() int

	// Return how many bytes have been used by pairs.
	Size() int64
}

type CacheCallBack struct {
	// optional and executed when an entry is purged.
	OnEvicted func(key string, value Value)
}

// New creates a new Cache.
// If maxEntries is zero, the cache has no limit and it's assumed
// that eviction is done by the caller.
func CreateCache(maxBytes int64, callBacks CacheCallBack, cacheType CachePolicy) Cache {
	switch cacheType {
	case LruPolicy:
		return &LRUCache{
			usedMap:       make(map[string]*list.Element),
			usedList:      list.New(),
			maxBytes:      maxBytes,
			curBytes:      0,
			CacheCallBack: callBacks,
		}
	case LfuPolicy:
		return &LFUCache{
			freqList:      list.New(),
			freqMap:       make(map[int]*list.Element),
			entryMap:      make(map[string]*list.Element),
			maxBytes:      maxBytes,
			curBytes:      0,
			CacheCallBack: callBacks,
		}
	default:
		panic(fmt.Sprintf("This cache replacement policy is not supported, which cache policy code is %d", cacheType))
	}
}
