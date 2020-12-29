package caching_test

import (
	"github.com/kmgreen2/agglo/pkg/caching"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestStringCacheEntryCompare(t *testing.T) {
	cacheEntry1 := caching.NewStringCacheEntry("fizzbuzz")
	cacheEntry2 := caching.NewStringCacheEntry("baz")
	cacheEntry3 := caching.NewStringCacheEntry("bar")
	cacheEntry4 := caching.NewStringCacheEntry("fizzbuzz")

	assert.True(t, cacheEntry1.Compare(cacheEntry2) > 0)
	assert.True(t, cacheEntry1.Compare(cacheEntry4) == 0)
	assert.True(t, cacheEntry2.Compare(cacheEntry3) > 0)
	assert.True(t, cacheEntry3.Compare(cacheEntry2) < 0)
}

func TestStringCacheEntryToBytes(t *testing.T) {
	cacheEntry := caching.NewStringCacheEntry("fizzbuzz")
	entryBytes, err := cacheEntry.ToBytes()
	assert.Nil(t, err)
	assert.Equal(t, []byte("fizzbuzz"), entryBytes)
}

func TestStringCacheEntryToString(t *testing.T) {
	cacheEntry := caching.NewStringCacheEntry("fizzbuzz")
	assert.Equal(t, "fizzbuzz", cacheEntry.ToString())
}

func TestStringCacheEntryToType(t *testing.T) {
	cacheEntry := caching.NewStringCacheEntry("fizzbuzz")
	var str	string
	var bad int

	err := cacheEntry.ToType(&str)
	assert.Nil(t, err)

	err = cacheEntry.ToType(&bad)
	assert.Error(t, err)
}
