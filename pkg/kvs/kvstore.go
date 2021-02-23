package kvs

import (
	"context"
	"fmt"
	"github.com/kmgreen2/agglo/pkg/util"
	"strings"
)

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

func NewKVStoreFromConnectionString(connectionString string, opts ...KVStoreOption) (KVStore, error) {
	connectionStringAry := strings.Split(connectionString, ":")
	if len(connectionStringAry) < 2 {
		return nil, util.NewInvalidError(fmt.Sprintf("invalid connection string, expected <type>:<connStr> got: %s",
			connectionString))
	}
	switch connectionStringAry[0] {
	case "mem":
		return NewMemKVStore(opts...), nil
	case "dynamo":
		return NewDynamoKVStoreFromConnectionString(strings.Join(connectionStringAry[1:], ":"))
	}
	return nil, util.NewInvalidError(fmt.Sprintf("invalid backend type: %s", connectionStringAry[0]))
}
