package caching

import (
	"bytes"
	"container/heap"
	"fmt"
	"github.com/kmgreen2/agglo/pkg/common"
	"strings"
	"time"
)

// LRUStringEntry is a string-based caching entry used
// to manage entries in an LRU caching
type LRUStringEntry struct {
	val string
	ctime int64
}

// NewLRUStringEntry will create a new LRUStringEntry
func NewLRUStringEntry(val string) *LRUStringEntry {
	return &LRUStringEntry{
		val,
		0,
	}
}

// CTime is the creation time of a caching entry used when
// TTL is set
func (entry *LRUStringEntry) CTime() int64 {
	return entry.ctime
}

// SetCTime will set the creation time of a caching entry
func (entry *LRUStringEntry) SetCTime(ctime int64) {
	entry.ctime = ctime
}

// ToBytes will convert the internal representation to bytes
func (entry *LRUStringEntry) ToBytes() ([]byte, error) {
	return []byte(entry.val), nil
}

// ToString will convert the internal representation to a string
func (entry *LRUStringEntry) ToString() string {
	return entry.val
}

// ToType will convert the entry to a specific type specified by the
// argument
func (entry *LRUStringEntry) ToType(elem interface{}) error {
	switch v := elem.(type) {
	case *string:
		elem = entry.val
		return nil
	default:
		return fmt.Errorf("Converting to %T is not supported for LRUStringEntry", v)
	}
}

// Compare will compare *this* caching entry to another entry
func (entry *LRUStringEntry) Compare(other Entry) int {
	otherBytes, err := other.ToBytes()
	if err != nil {
		return -1
	}
	return bytes.Compare([]byte(entry.val), otherBytes)
}

// LRUStringKey is string-based key used to track LRU caching entries
type LRUStringKey struct {
	val string
	atime int64
	index int
}

// NewLRUStringKey will create a new LRUStringKey
func NewLRUStringKey(val string) *LRUStringKey {
	return &LRUStringKey{
		val: val,
		atime: 0,
		index: 0,
	}
}

// ATime is the last access time of a LRUStringKey
func (key *LRUStringKey) ATime() int64 {
	return key.atime
}

// SetATime will set the last access time of a LRUStringKey
func (key *LRUStringKey) SetATime(atime int64) {
	key.atime = atime
}

// Index is the index of the key in the LRU queue
func (key *LRUStringKey) Index() int {
	return key.index
}

// SetIndex will set the index of the key in the LRU queue
func (key *LRUStringKey) SetIndex(index int) {
	key.index = index
}

// ToBytes will convert the string key to bytes
func (key *LRUStringKey) ToBytes() ([]byte, error) {
	return []byte(key.val), nil
}

// ToString will return the string key
func (key *LRUStringKey) ToString() string {
	return key.val
}

// Compare will compare *this* key to another caching key
func (key *LRUStringKey) Compare(other Key) int {
	switch v := other.(type) {
	case *LRUStringKey:
		return strings.Compare(key.val, v.ToString())
	default:
		return -1
	}
}

// LRUMapCache is an in-memory LRU caching with the ability to
// set a caching-wide TTL for entries
type LRUMapCache struct {
	cache map[string]LRUEntry
	keys map[string]LRUKey
	lruQueue *LRUQueue
	maxEntries int
	ttl int64
}

// NewLRUMapCache will create a new LRUMapCache
func NewLRUMapCache(maxEntries int, ttl int64) *LRUMapCache {
	lruQueue := make(LRUQueue, 0)
	heap.Init(&lruQueue)
	return &LRUMapCache{
		make(map[string]LRUEntry),
		make(map[string]LRUKey),
		&lruQueue,
		maxEntries,
		ttl,
	}
}

// Put will put a key,value pair into the caching, overwriting other values
// with the same key.  If the caching has hit capacity, an entry with the
// oldest atime will be evicted
func (lruMapCache *LRUMapCache) Put(key LRUKey, entry LRUEntry) error {
	currTime := time.Now().UnixNano()
	lruKey := NewLRUStringKey(key.ToString())
	lruKey.SetATime(currTime)
	lruEntry := NewLRUStringEntry(entry.ToString())
	lruEntry.SetCTime(currTime)
	if lruMapCache.lruQueue.Len() >= lruMapCache.maxEntries {
		evictedKey := heap.Pop(lruMapCache.lruQueue).(LRUKey)
		delete(lruMapCache.cache, evictedKey.ToString())
		delete(lruMapCache.keys, evictedKey.ToString())
	}
	lruMapCache.cache[key.ToString()] = lruEntry

	// Remove an existing key from the queue, if exists
	if existingKey, ok := lruMapCache.keys[key.ToString()]; ok {
		lruMapCache.lruQueue.Delete(existingKey)
	}

	lruMapCache.keys[key.ToString()] = lruKey
	heap.Push(lruMapCache.lruQueue, lruKey)
	return nil
}

// Get will get an entry based on a given value.  Values exceeding the
// caching-wide TTL will be evicted.
func (lruMapCache *LRUMapCache) Get(key LRUKey) (LRUEntry, error) {
	if elem, ok := lruMapCache.cache[key.ToString()]; ok {
		currTime := time.Now().UnixNano()
		ctime := elem.CTime()
		ageInSeconds := (currTime - ctime) / 1e9
		if lruMapCache.ttl > 0 && ageInSeconds > lruMapCache.ttl {
			lruItem := lruMapCache.keys[key.ToString()]
			lruMapCache.lruQueue.Delete(lruItem)
			delete(lruMapCache.cache, key.ToString())
			delete(lruMapCache.keys, key.ToString())
			return nil, common.WithStack(common.NewEvictedError(fmt.Sprintf("key: %s", key.ToString())))
		}

		lruKey := lruMapCache.keys[key.ToString()]
		lruKey.SetATime(time.Now().UnixNano())
		lruMapCache.lruQueue.Update(lruKey)
		return elem, nil
	}

	return nil, common.WithStack(common.NewNotFoundError(fmt.Sprintf("key: %s", key.ToString())))
}

// Delete will delete a key and entry from the LRU caching
func (lruMapCache *LRUMapCache) Delete(key LRUKey) error {
	if _, ok := lruMapCache.cache[key.ToString()]; ok {
		if lruItem, ok := lruMapCache.keys[key.ToString()]; ok {
			delete(lruMapCache.cache, key.ToString())
			delete(lruMapCache.keys, key.ToString())
			lruMapCache.lruQueue.Delete(lruItem)
		} else {
			return common.NewNotFoundError(fmt.Sprintf("Could not find key in LRU queue: %s", key.ToString()))
		}
	} else {
		return common.NewNotFoundError(fmt.Sprintf("Could not find key in caching: %s", key.ToString()))
	}
	return nil
}

// Close is a no-op for the local cache
func (lruMapCache *LRUMapCache) Close() error {
	for k := range lruMapCache.keys {
		delete(lruMapCache.keys, k)
	}
	for k := range lruMapCache.cache {
		delete(lruMapCache.cache, k)
	}
	v := lruMapCache.lruQueue.Pop()
	for v != nil {
		v = lruMapCache.lruQueue.Pop()
	}
	return nil
}

