package caching

// Cache has a very simple interface for putting and getting values to/from a
// cache.  This interface is to be used by simple non-LRU local caches and
// hosted caches (e.g. Redis)
type Cache interface {
	Put(key Key, entry Entry) error
	Get(key Key) (Entry, error)
	Delete(key Key) error
	Close() error
}

// LRUCache has a very simple interface for putting and getting values to/from a
// cache.  This interface should be implemented by local LRU caches
type LRUCache interface {
	Put(key LRUKey, entry LRUEntry) error
	Get(key LRUKey) (LRUEntry, error)
	Delete(key LRUKey) error
	Close() error
}
