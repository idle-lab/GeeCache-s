package cachePolicy

import (
	"container/list"
	"fmt"
)

type lfuEntry struct {
	freq  int
	key   string
	value Value
}

type LFUCache struct {
	CacheCallBack
	freqList *list.List
	freqMap  map[int]*list.Element
	entryMap map[string]*list.Element

	// The maximum number of bytes available for all pairs.
	maxBytes int64
	// The current number of bytes have been used by pairs.
	curBytes int64
}

func (lfu *LFUCache) Get(key string) (Value, bool) {
	if v, ok := lfu.entryMap[key]; ok {
		lfu.increaseFreq(v)
		return v.Value.(*lfuEntry).value, true
	}
	return nil, false
}

func (lfu *LFUCache) Add(key string, value Value) error {
	insertedBytes := value.Size() + int64(len(key)) + /* freq */ 4
	if lfu.maxBytes != 0 && insertedBytes > lfu.maxBytes {
		return fmt.Errorf("the size of value is too large, size need less than %d which is %d", lfu.maxBytes, value.Size())
	}

	if v, ok := lfu.entryMap[key]; ok {
		entry := v.Value.(*lfuEntry)
		for lfu.maxBytes != 0 && lfu.curBytes+value.Size()-entry.value.Size() > lfu.maxBytes {
			lfu.Evict()
		}
		lfu.curBytes += value.Size() - entry.value.Size()
		entry.value = value
		lfu.increaseFreq(v)
	} else {
		for lfu.maxBytes != 0 && lfu.curBytes+insertedBytes > lfu.maxBytes {
			lfu.Evict()
		}
		entry := &lfuEntry{
			freq:  1,
			key:   key,
			value: value,
		}
		if _, ok := lfu.freqMap[1]; !ok {
			lfu.freqMap[1] = lfu.freqList.PushFront(list.New())
		}
		lfu.entryMap[key] = lfu.freqMap[1].Value.(*list.List).PushBack(entry)
		lfu.curBytes += insertedBytes
	}

	return nil
}

func (lfu *LFUCache) Evict() {
	front := lfu.freqList.Front()
	entryList := front.Value.(*list.List)
	first := entryList.Front()
	firstEntry := first.Value.(*lfuEntry)

	if lfu.CacheCallBack.OnEvicted != nil {
		lfu.OnEvicted(firstEntry.key, firstEntry.value)
	}

	lfu.curBytes -= firstEntry.value.Size() + int64(len(firstEntry.key)) + 4

	delete(lfu.entryMap, firstEntry.key)
	entryList.Remove(first)
	if entryList.Len() == 0 {
		delete(lfu.freqMap, firstEntry.freq)
		lfu.freqList.Remove(front)
	}
}

func (lfu *LFUCache) Len() int {
	return len(lfu.entryMap)
}

func (lfu *LFUCache) Size() int64 {
	return lfu.curBytes
}

func (lfu *LFUCache) increaseFreq(v *list.Element) {
	entry := v.Value.(*lfuEntry)
	elem := lfu.freqMap[entry.freq]
	entryList := elem.Value.(*list.List)

	// insert into the list of freq+1.
	entry.freq++
	if _, ok := lfu.freqMap[entry.freq]; !ok {
		lfu.freqMap[entry.freq] = lfu.freqList.InsertAfter(list.New(), elem)
	}
	lfu.entryMap[entry.key] = lfu.freqMap[entry.freq].Value.(*list.List).PushBack(entry)

	// remove entry from old list.
	entryList.Remove(v)
	if entryList.Len() == 0 {
		lfu.freqList.Remove(elem)
		delete(lfu.freqMap, entry.freq-1)
	}
}
