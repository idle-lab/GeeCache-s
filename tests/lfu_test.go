package tests

import (
	"geecache-s/cachePolicy"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testValue struct {
	data string
	size int64
}

func (v testValue) Size() int64 { return v.size }

func TestLFUBasicOperations(t *testing.T) {
	cache := cachePolicy.CreateCache(1000, cachePolicy.CacheCallBack{}, cachePolicy.LfuPolicy)

	_, ok := cache.Get("not_exist")
	assert.False(t, ok)
	assert.Equal(t, 0, cache.Len())
	assert.Equal(t, int64(0), cache.Size())

	val1 := testValue{data: "test1", size: 10}
	assert.NoError(t, cache.Add("key1", val1))
	assert.Equal(t, 1, cache.Len())
	assert.Equal(t, int64(10+len("key1")+4), cache.Size()) // 10 + 4(keylen) + 4(freq)

	v, ok := cache.Get("key1")
	assert.True(t, ok)
	assert.Equal(t, val1, v.(testValue))

	val2 := testValue{data: "test2", size: 20}
	assert.NoError(t, cache.Add("key1", val2))
	assert.Equal(t, 1, cache.Len())
	assert.Equal(t, int64(20+len("key1")+4), cache.Size())

	v, ok = cache.Get("key1")
	assert.True(t, ok)
	assert.Equal(t, val2, v.(testValue))
}

func TestLFUEvictionFlow(t *testing.T) {
	cache := cachePolicy.CreateCache(50, cachePolicy.CacheCallBack{}, cachePolicy.LfuPolicy)

	// Add base entries (each entry size: 4 + 3(key) + 5 = 12)
	assert.NoError(t, cache.Add("key1", testValue{size: 4}))
	assert.NoError(t, cache.Add("key2", testValue{size: 4}))
	assert.NoError(t, cache.Add("key3", testValue{size: 4}))
	assert.NoError(t, cache.Add("key4", testValue{size: 4})) // Total 48
	assert.Equal(t, 4, cache.Len())

	// 48 + 12 > 50
	assert.NoError(t, cache.Add("key5", testValue{size: 4}))

	_, ok := cache.Get("key1")
	assert.False(t, ok, "淘汰 key1")
	assert.True(t, cache.Len() == 4)

	_, ok = cache.Get("key2")
	assert.True(t, ok)

	assert.NoError(t, cache.Add("key6", testValue{size: 4}))

	_, ok = cache.Get("key2")
	assert.True(t, ok)
	_, ok = cache.Get("key3")
	assert.False(t, ok)
}

func TestLFUSameFrequencyEviction(t *testing.T) {
	cache := cachePolicy.CreateCache(40, cachePolicy.CacheCallBack{}, cachePolicy.LfuPolicy)

	assert.NoError(t, cache.Add("key1", testValue{size: 4}))
	assert.NoError(t, cache.Add("key2", testValue{size: 4}))
	assert.NoError(t, cache.Add("key3", testValue{size: 4}))

	var ok bool
	_, ok = cache.Get("key1")
	assert.True(t, ok)
	_, ok = cache.Get("key1")
	assert.True(t, ok)
	_, ok = cache.Get("key1") // 4
	assert.True(t, ok)
	_, ok = cache.Get("key2")
	assert.True(t, ok)
	_, ok = cache.Get("key2") // 3
	assert.True(t, ok)
	_, ok = cache.Get("key3") // 2
	assert.True(t, ok)

	assert.NoError(t, cache.Add("key4", testValue{size: 4}))

	// key3 should be evicted.
	_, ok = cache.Get("key1")
	assert.True(t, ok)
	_, ok = cache.Get("key2")
	assert.True(t, ok)
	_, ok = cache.Get("key3")
	assert.False(t, ok)
	_, ok = cache.Get("key4")
	assert.True(t, ok)
	assert.Equal(t, 3, cache.Len())
}

func TestLFUErrorConditions(t *testing.T) {
	cache := cachePolicy.CreateCache(10, cachePolicy.CacheCallBack{}, cachePolicy.LfuPolicy)
	largeVal := testValue{size: 20} // 需要20 + key长度 + 4 > 10
	err := cache.Add("large", largeVal)
	assert.Error(t, err, "应拒绝过大条目")
	assert.Equal(t, 0, cache.Len())
}

func TestLFUEvictionCallback(t *testing.T) {
	var evictedKey string
	var evictedValue cachePolicy.Value

	cb := cachePolicy.CacheCallBack{
		OnEvicted: func(key string, value cachePolicy.Value) {
			evictedKey = key
			evictedValue = value
		},
	}

	cache := cachePolicy.CreateCache(25, cb, cachePolicy.LfuPolicy)
	assert.NoError(t, cache.Add("key1", testValue{size: 5}))
	assert.NoError(t, cache.Add("key2", testValue{size: 15})) // 触发淘汰

	assert.Equal(t, "key1", evictedKey, "应触发淘汰回调")
	assert.Equal(t, int64(5), evictedValue.Size())
}
