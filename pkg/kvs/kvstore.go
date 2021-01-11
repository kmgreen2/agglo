package kvs

import "context"

// KVStore is the interface for a key-value store
type KVStore interface {
	AtomicPut(ctx context.Context, key string, prev, value []byte) error
	AtomicDelete(ctx context.Context, key string, prev []byte) error
	Put(ctx context.Context, key string, value []byte) error
	Get(ctx context.Context, key string) ([]byte, error)
	Head(ctx context.Context, key string) error
	Delete(ctx context.Context, key string) error
	List(ctx context.Context, prefix string) ([]string, error)
	ConnectionString() string
	Close() error
}

// ROKVStore is a read-only interface to a KVSTore
type ROKVStore interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Head(ctx context.Context, key string) error
	List(ctx context.Context, prefix string) ([]string, error)
	ConnectionString() string
}
