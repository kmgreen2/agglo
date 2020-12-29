package caching_test

import (
	"fmt"
	"github.com/kmgreen2/agglo/pkg/caching"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestLRUStringEntryCompare(t *testing.T) {
	cacheKey1 := caching.NewLRUStringEntry("foo")
	cacheKey2 := caching.NewLRUStringEntry("bar")
	cacheKey3 := caching.NewLRUStringEntry("baz")
	cacheKey4 := caching.NewLRUStringEntry("foo")

	assert.True(t, cacheKey1.Compare(cacheKey4) == 0)
	assert.True(t, cacheKey1.Compare(cacheKey2) > 0)
	assert.True(t, cacheKey2.Compare(cacheKey3) < 0)
	assert.True(t, cacheKey3.Compare(cacheKey2) > 0)
}

func TestLRUStringEntryToBytes(t *testing.T) {
	cacheEntry := caching.NewLRUStringEntry("fizzbuzz")
	entryBytes, err := cacheEntry.ToBytes()
	assert.Nil(t, err)
	assert.Equal(t, []byte("fizzbuzz"), entryBytes)
}

func TestLRUStringEntryToString(t *testing.T) {
	cacheEntry := caching.NewLRUStringEntry("fizzbuzz")
	assert.Equal(t, "fizzbuzz", cacheEntry.ToString())
}

func TestLRUStringKeyCompare(t *testing.T) {
	cacheKey1 := caching.NewLRUStringKey("foo")
	cacheKey2 := caching.NewLRUStringKey("bar")
	cacheKey3 := caching.NewLRUStringKey("baz")
	cacheKey4 := caching.NewLRUStringKey("foo")

	assert.True(t, cacheKey1.Compare(cacheKey4) == 0)
	assert.True(t, cacheKey1.Compare(cacheKey2) > 0)
	assert.True(t, cacheKey2.Compare(cacheKey3) < 0)
	assert.True(t, cacheKey3.Compare(cacheKey2) > 0)
	assert.True(t, cacheKey3.Compare(caching.NewStringCacheKey("foo")) == -1)
}

func TestLRUStringKeyToBytes(t *testing.T) {
	cacheKey := caching.NewLRUStringKey("foo")
	keyBytes, err := cacheKey.ToBytes()
	assert.Nil(t, err)
	assert.Equal(t, []byte("foo"), keyBytes)
}

func TestLRUStringKeyToString(t *testing.T) {
	cacheKey := caching.NewLRUStringKey("foo")
	assert.Equal(t, "foo", cacheKey.ToString())
}

func TestLRUMapCacheAtCapacity(t *testing.T) {
	numEntries := 10
	lruMapCache := caching.NewLRUMapCache(numEntries, -1)
	defer lruMapCache.Close()

	for i := 0; i < numEntries; i++ {
		k, v := caching.NewLRUStringKey(fmt.Sprint(i)), caching.NewLRUStringEntry(fmt.Sprint(i))
		if err := lruMapCache.Put(k, v); err != nil {
			assert.Fail(t, err.Error())
		}

		for j := 0; j <= i; j++ {
			result, err := lruMapCache.Get(caching.NewLRUStringKey(fmt.Sprint(j)))
			if err != nil {
				assert.Fail(t, err.Error())
			}
			assert.Zero(t, result.Compare(caching.NewStringCacheEntry(fmt.Sprint(j))))
		}
	}
}

func TestLRUMapCacheAboveCapacity(t *testing.T) {
	maxEntries := 5
	numEntries := 20
	lruMapCache := caching.NewLRUMapCache(maxEntries, -1)
	defer lruMapCache.Close()

	for i := 0; i < numEntries; i++ {
		k, v := caching.NewLRUStringKey(fmt.Sprint(i)), caching.NewLRUStringEntry(fmt.Sprint(i))
		if err := lruMapCache.Put(k, v); err != nil {
			assert.Fail(t, err.Error())
		}
		// Ensure caching entries are sequenced (no ctime collisions)
		time.Sleep(10 * time.Millisecond)
		if i >= maxEntries {
			_, err := lruMapCache.Get(caching.NewLRUStringKey(fmt.Sprint(i - maxEntries)))
			if err == nil {
				fmt.Print(i)
			}
			assert.Error(t, err, "Not found")
		}
	}
}

func TestLRUMapCacheTTL(t *testing.T) {
	numEntries := 10
	lruMapCache := caching.NewLRUMapCache(numEntries, 2)
	defer lruMapCache.Close()

	for i := 0; i < numEntries; i++ {
		k, v := caching.NewLRUStringKey(fmt.Sprint(i)), caching.NewLRUStringEntry(fmt.Sprint(i))
		if err := lruMapCache.Put(k, v); err != nil {
			assert.Fail(t, err.Error())
		}
	}

	time.Sleep(3 * time.Second)

	for i := 0; i < numEntries; i++ {
		_, err := lruMapCache.Get(caching.NewLRUStringKey(fmt.Sprint(i)))
		assert.Error(t, err)
	}
}

func TestLRUMapCacheBelowTTL(t *testing.T) {
	numEntries := 10
	lruMapCache := caching.NewLRUMapCache(numEntries, 120)
	defer lruMapCache.Close()

	for i := 0; i < numEntries; i++ {
		k, v := caching.NewLRUStringKey(fmt.Sprint(i)), caching.NewLRUStringEntry(fmt.Sprint(i))
		if err := lruMapCache.Put(k, v); err != nil {
			assert.Fail(t, err.Error())
		}

		for j := 0; j <= i; j++ {
			result, err := lruMapCache.Get(caching.NewLRUStringKey(fmt.Sprint(j)))
			if err != nil {
				assert.Fail(t, err.Error())
			}
			assert.Zero(t, result.Compare(caching.NewStringCacheEntry(fmt.Sprint(j))))
		}
	}
}

func TestLRUMapCacheDuplicateKeys(t *testing.T) {
	numEntries := 10
	lruMapCache := caching.NewLRUMapCache(numEntries, -1)
	defer lruMapCache.Close()

	for i := 0; i < numEntries - 1; i++ {
		k, v := caching.NewLRUStringKey(fmt.Sprint(i)), caching.NewLRUStringEntry(fmt.Sprint(i))
		if err := lruMapCache.Put(k, v); err != nil {
			assert.Fail(t, err.Error())
		}
	}

	k, v := caching.NewLRUStringKey("0"), caching.NewLRUStringEntry("1")
	if err := lruMapCache.Put(k, v); err != nil {
		assert.Fail(t, err.Error())
	}

	result, err := lruMapCache.Get(k)
	assert.Nil(t, err)
	assert.Equal(t, v.ToString(), result.ToString())
}

func TestLRUMapCacheDelete(t *testing.T) {
	numEntries := 10
	lruMapCache := caching.NewLRUMapCache(numEntries, -1)
	defer lruMapCache.Close()

	for i := 0; i < numEntries; i++ {
		k, v := caching.NewLRUStringKey(fmt.Sprint(i)), caching.NewLRUStringEntry(fmt.Sprint(i))
		if err := lruMapCache.Put(k, v); err != nil {
			assert.Fail(t, err.Error())
		}
	}
	for i := 0; i < numEntries; i++ {
		k := caching.NewLRUStringKey(fmt.Sprint(i))
		err := lruMapCache.Delete(k)
		assert.Nil(t, err)
		_, err = lruMapCache.Get(k)
		assert.Error(t, err)
		err = lruMapCache.Delete(k)
		assert.Error(t, err)
	}
}
