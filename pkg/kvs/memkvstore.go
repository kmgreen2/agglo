package kvs

import (
	"bytes"
	"context"
	"fmt"
	"github.com/kmgreen2/agglo/pkg/common"
	"strings"
	"sync"
)

// MemKVStore is a KVStore implementation that uses an in-memory map
type MemKVStore struct {
	values map[string][]byte
	lock *sync.Mutex
}

// NewMemKVStore will return a new MemKVStore object
func NewMemKVStore() *MemKVStore {
	return &MemKVStore{
		values: make(map[string][]byte),
		lock: &sync.Mutex{},
	}
}

// AtomicPut will atomically map a value to a key in the in-memory map
func (kvStore *MemKVStore) AtomicPut(ctx context.Context, key string, prev, value []byte) error {
	kvStore.lock.Lock()
	defer kvStore.lock.Unlock()

	currBytes, _ := kvStore.values[key]

	if prev == nil && currBytes == nil || (prev != nil && bytes.Compare(currBytes, prev) == 0){
		kvStore.values[key] = value
	} else {
		msg := fmt.Sprintf("state has changed for '%s', cannot apply atomic update", key)
		return common.NewConflictError(msg)
	}

	return nil
}

// AtomicDelete will atomically deletevalue to a key in the in-memory map
func (kvStore *MemKVStore) AtomicDelete(ctx context.Context, key string, prev []byte) error {
	kvStore.lock.Lock()
	defer kvStore.lock.Unlock()

	currBytes, _ := kvStore.values[key]

	if prev != nil && bytes.Compare(currBytes, prev) == 0 {
		delete(kvStore.values, key)
		return nil
	} else if prev == nil && currBytes == nil {
		return nil
	} else {
		msg := fmt.Sprintf("state has changed for '%s', cannot apply atomic update", key)
		return common.NewConflictError(msg)
	}
}

// Put will map a value to a key in the in-memory map
func (kvStore *MemKVStore) Put(ctx context.Context, key string, value []byte) error {
	kvStore.lock.Lock()
	defer kvStore.lock.Unlock()
	kvStore.values[key] = value
	return nil
}

// Get will return a value mapped to the provided key, or error if the mapping does not exist
func (kvStore *MemKVStore) Get(ctx context.Context, key string) ([]byte, error) {
	kvStore.lock.Lock()
	defer kvStore.lock.Unlock()
	if _, ok := kvStore.values[key]; !ok {
		return nil, common.NewNotFoundError(fmt.Sprintf("MemKVStore - key does not exist: %s", key))
	}
	return kvStore.values[key], nil
}

// Head will return an error if the key is not mapped or nil if it is mapped
func (kvStore *MemKVStore) Head(ctx context.Context, key string) error {
	kvStore.lock.Lock()
	defer kvStore.lock.Unlock()
	if _, ok := kvStore.values[key]; !ok {
		return common.NewNotFoundError(fmt.Sprintf("MemKVStore - key does not exist: %s", key))
	}
	return nil
}

// Delete will unmap a key, if it exists; otherwise returns an error
func (kvStore *MemKVStore) Delete(ctx context.Context, key string) error {
	kvStore.lock.Lock()
	defer kvStore.lock.Unlock()
	if _, ok := kvStore.values[key]; !ok {
		return common.NewNotFoundError(fmt.Sprintf("MemKVStore - key does not exist: %s", key))
	}
	delete(kvStore.values, key)
	return nil
}

// List will return all of the keys with the given prefix
func (kvStore *MemKVStore) List(ctx context.Context, prefix string) ([]string, error) {
	kvStore.lock.Lock()
	defer kvStore.lock.Unlock()
	var result []string
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

