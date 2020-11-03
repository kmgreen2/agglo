package kvs

// ObjectStore is the interface for a key-value store
type KVStore interface {
	Put(key string, value []byte) error
	Get(key string) ([]byte, error)
	Head(key string) error
	Delete(key string) error
	List(prefix string) ([]string, error)
}