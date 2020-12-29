package caching_test

import (
	"github.com/kmgreen2/agglo/pkg/caching"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestStringCacheKeyCompare(t *testing.T) {
	cacheKey1 := caching.NewStringCacheKey("foo")
	cacheKey2 := caching.NewStringCacheKey("bar")
	cacheKey3 := caching.NewStringCacheKey("baz")
	cacheKey4 := caching.NewStringCacheKey("foo")

	assert.True(t, cacheKey1.Compare(cacheKey4) == 0)
	assert.True(t, cacheKey1.Compare(cacheKey2) > 0)
	assert.True(t, cacheKey2.Compare(cacheKey3) < 0)
	assert.True(t, cacheKey3.Compare(cacheKey2) > 0)
	assert.True(t, cacheKey3.Compare(caching.NewLRUStringKey("foo")) == -1)
}

func TestStringCacheKeyToBytes(t *testing.T) {
	cacheKey := caching.NewStringCacheKey("foo")
	keyBytes, err := cacheKey.ToBytes()
	assert.Nil(t, err)
	assert.Equal(t, []byte("foo"), keyBytes)
}

func TestStringCacheKeyToString(t *testing.T) {
	cacheKey := caching.NewStringCacheKey("foo")
	assert.Equal(t, "foo", cacheKey.ToString())
}
