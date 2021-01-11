package kvs

import (
	"bytes"
	"context"
	"fmt"
	"github.com/kmgreen2/agglo/pkg/common"
	"github.com/kmgreen2/agglo/pkg/observability"
	"strings"
	"sync"
)

type KVStoreOption func(*MemKVStore)

func WithTracing() KVStoreOption {
	return func(kvStore *MemKVStore) {
		kvStore.emitter = observability.NewEmitter("agglo/memKvStore")
	}
}


// MemKVStore is a KVStore implementation that uses an in-memory map
type MemKVStore struct {
	values map[string][]byte
	lock *sync.Mutex
	emitter *observability.Emitter
}

// NewMemKVStore will return a new MemKVStore object
func NewMemKVStore(opts ...KVStoreOption) *MemKVStore {
	kvStore := &MemKVStore{
		values: make(map[string][]byte),
		lock: &sync.Mutex{},
	}

	for _, opt := range opts {
		opt(kvStore)
	}
	return kvStore
}

// AtomicPut will atomically map a value to a key in the in-memory map
func (kvStore *MemKVStore) AtomicPut(ctx context.Context, key string, prev, value []byte) error {
	var err error
	kvStore.lock.Lock()
	defer kvStore.lock.Unlock()

	if kvStore.emitter != nil {
		_, span := kvStore.emitter.CreateSpan(ctx, "atomicPut")
		defer func (err error) {
			if err != nil {
				span.RecordError(err)
			}
			span.End()
		}(err)
	}

	currBytes, _ := kvStore.values[key]

	if prev == nil && currBytes == nil || (prev != nil && bytes.Compare(currBytes, prev) == 0){
		kvStore.values[key] = value
	} else {
		msg := fmt.Sprintf("state has changed for '%s', cannot apply atomic update", key)
		err = common.NewConflictError(msg)
		return err
	}

	return nil
}

// AtomicDelete will atomically deletevalue to a key in the in-memory map
func (kvStore *MemKVStore) AtomicDelete(ctx context.Context, key string, prev []byte) error {
	var err error
	kvStore.lock.Lock()
	defer kvStore.lock.Unlock()

	if kvStore.emitter != nil {
		_, span := kvStore.emitter.CreateSpan(ctx, "atomicDelete")
		defer func (err error) {
			if err != nil {
				span.RecordError(err)
			}
			span.End()
		}(err)
	}

	currBytes, _ := kvStore.values[key]

	if prev != nil && bytes.Compare(currBytes, prev) == 0 {
		delete(kvStore.values, key)
		return nil
	} else if prev == nil && currBytes == nil {
		return nil
	} else {
		msg := fmt.Sprintf("state has changed for '%s', cannot apply atomic update", key)
		err = common.NewConflictError(msg)
		return err
	}
}

// Put will map a value to a key in the in-memory map
func (kvStore *MemKVStore) Put(ctx context.Context, key string, value []byte) error {
	kvStore.lock.Lock()
	defer kvStore.lock.Unlock()

	if kvStore.emitter != nil {
		_, span := kvStore.emitter.CreateSpan(ctx, "put")
		defer func () {
			span.End()
		}()
	}
	kvStore.values[key] = value
	return nil
}

// Get will return a value mapped to the provided key, or error if the mapping does not exist
func (kvStore *MemKVStore) Get(ctx context.Context, key string) ([]byte, error) {
	var err error
	kvStore.lock.Lock()
	defer kvStore.lock.Unlock()

	if kvStore.emitter != nil {
		_, span := kvStore.emitter.CreateSpan(ctx, "get")
		defer func (err error) {
			if err != nil {
				span.RecordError(err)
			}
			span.End()
		}(err)
	}

	if _, ok := kvStore.values[key]; !ok {
		err = common.NewNotFoundError(fmt.Sprintf("MemKVStore - key does not exist: %s", key))
		return nil, err
	}
	return kvStore.values[key], nil
}

// Head will return an error if the key is not mapped or nil if it is mapped
func (kvStore *MemKVStore) Head(ctx context.Context, key string) error {
	var err error
	kvStore.lock.Lock()
	defer kvStore.lock.Unlock()

	if kvStore.emitter != nil {
		_, span := kvStore.emitter.CreateSpan(ctx, "head")
		defer func (err error) {
			if err != nil {
				span.RecordError(err)
			}
			span.End()
		}(err)
	}

	if _, ok := kvStore.values[key]; !ok {
		err = common.NewNotFoundError(fmt.Sprintf("MemKVStore - key does not exist: %s", key))
		return err
	}
	return nil
}

// Delete will unmap a key, if it exists; otherwise returns an error
func (kvStore *MemKVStore) Delete(ctx context.Context, key string) error {
	var err error
	kvStore.lock.Lock()
	defer kvStore.lock.Unlock()

	if kvStore.emitter != nil {
		_, span := kvStore.emitter.CreateSpan(ctx, "delete")
		defer func (err error) {
			if err != nil {
				span.RecordError(err)
			}
			span.End()
		}(err)
	}

	if _, ok := kvStore.values[key]; !ok {
		err = common.NewNotFoundError(fmt.Sprintf("MemKVStore - key does not exist: %s", key))
		return err
	}
	delete(kvStore.values, key)
	return nil
}

// List will return all of the keys with the given prefix
func (kvStore *MemKVStore) List(ctx context.Context, prefix string) ([]string, error) {
	kvStore.lock.Lock()
	defer kvStore.lock.Unlock()
	var result []string

	if kvStore.emitter != nil {
		_, span := kvStore.emitter.CreateSpan(ctx, "list")
		defer func () {
			span.End()
		}()
	}

	prefixLength := len(prefix)
	for s, _ := range kvStore.values {
		if strings.Compare(prefix, s[:prefixLength]) == 0 {
			result = append(result, s)
		}
	}
	return result, nil
}

// ConnectionString will return a string that can be parsed by NewMemKVStore to create an instance
// of KVStore
func (kvStore *MemKVStore) ConnectionString() string {
	return "inMemStore"
}

// Close will flush any in-flight changes and close the connection to the backing system.  In this case, a no-op
func (kvStore *MemKVStore) Close() error {
	return nil
}

