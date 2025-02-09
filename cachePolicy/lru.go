package cachePolicy

import (
	"container/list"
	"fmt"
)

type lruEntry struct {
	key   string
	value Value
}

type LRUCache struct {
	CacheCallBack
	usedMap  map[string]*list.Element
	usedList *list.List

	// The maximum number of bytes available for all pairs.
	maxBytes int64
	// The current number of bytes have been used by pairs.
	curBytes int64
}

func (lru *LRUCache) Get(key string) (Value, bool) {
	if elem, ok := lru.usedMap[key]; ok {
		lru.usedList.MoveToFront(elem)
		return elem.Value.(lruEntry).value, true
	}
	return nil, false
}

func (lru *LRUCache) Add(key string, value Value) error {
	insertedBytes := value.Size() + int64(len(key))
	if lru.maxBytes != 0 && insertedBytes > lru.maxBytes {
		return fmt.Errorf("the size of value is too large, need less than %d which is %d", lru.maxBytes, value.Size())
	}

	if elem, ok := lru.usedMap[key]; ok {
		lru.curBytes += value.Size() - elem.Value.(lruEntry).value.Size()
		elem.Value = lruEntry{key, value}
		lru.usedList.MoveToFront(elem)
	} else {
		lru.usedMap[key] = lru.usedList.PushFront(lruEntry{key, value})
		lru.curBytes += insertedBytes
	}

	for lru.maxBytes != 0 && lru.curBytes >= lru.maxBytes {
		lru.Evict()
	}
	return nil
}

func (lru *LRUCache) Evict() {
	back := lru.usedList.Back().Value.(lruEntry)
	lru.curBytes -= int64(len(back.key)) + back.value.Size()
	delete(lru.usedMap, back.key)
	lru.usedList.Remove(lru.usedList.Back())

	if lru.OnEvicted != nil {
		lru.OnEvicted(back.key, back.value)
	}
}

func (lru *LRUCache) Len() int {
	return len(lru.usedMap)
}

func (lru *LRUCache) Size() int64 {
	return lru.curBytes
}
