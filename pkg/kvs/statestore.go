package kvs

import (
	"context"
	gUuid "github.com/google/uuid"
	"github.com/kmgreen2/agglo/pkg/common"
	"github.com/pkg/errors"
)

var AppendEntryDelimiter = ":"

type StateStore interface {
	Append(ctx context.Context, key string, value []byte) error
	Checkpoint(ctx context.Context, key string, mapFn func(curr, val []byte)([]byte, error)) error
	Get(ctx context.Context, key string) ([]byte, error)
	Delete(ctx context.Context, key string) error
	AtomicDelete(ctx context.Context, key string, prev []byte) error
}

func NewMemStateStore() *KvStateStore {
	return &KvStateStore{
		NewMemKVStore(),
	}
}

func NewKvStateStore(kvStore KVStore) *KvStateStore {
	return &KvStateStore{
		kvStore,
	}
}

type KvStateStore struct {
	kvStore KVStore
}

func (store KvStateStore) Append(ctx context.Context, key string, value []byte) error {
	randUuid, err := gUuid.NewRandom()
	if err != nil {
		return err
	}
	return store.kvStore.AtomicPut(ctx, key + AppendEntryDelimiter + randUuid.String(), nil, value)
}

func (store *KvStateStore) needsCheckpoint(ctx context.Context, key string) (bool, error) {
	elementKeys, err := store.kvStore.List(ctx, key + AppendEntryDelimiter)
	if err != nil {
		return false, err
	}
	return len(elementKeys) > 0, nil
}

/**
 * ToDo(KMG): This will compute and store a checkpoint of pending append records.
 * It atomically tries to acquire a lock, which is an AtomicPut on a known key that
 * contains the append records to checkpoint.  After storing the new checkpoint, it
 * will delete the append entries and release the lock.
 *
 * We need to make sure we handle the failure case where this function fails between
 * deleting the append entries and releasing the lock.  As of now, the checkpoint can
 * never be updated, since the lock will always be held.
 *
 * Options:
 *
 * 1. Build soft delete into the underlying kvStore.  This can still suffer from the same
 *    problem, since we ultimately have to do deletes.
 * 2. Build bulk, atomic deletes into the underlying kvStore.  This will restrict the types
 *    of kvStores that can be used, since not all support atomic deletes, requiring the
 *    client to implement it on top of the underlying API.
 * 3. Add an additional checkpoint state on top of a shared lock that could allow us to
 *    release the lock after failing to delete all of the append entries.  This entry would contain
 *    the entries that could not be deleted, which could be used to attempt deletion and filter
 *    while computing subsequent checkpoints.
 */
func (store *KvStateStore) Checkpoint(ctx context.Context, key string,
	mapFn func(currCheckpoint, val []byte)([]byte, error)) error {

	/*
	 * Acquire a lock to do the checkpoint.  If this fails, assume someone else is doing it
	 */
	lock := NewKVDistributedLock(key, store.kvStore)
	ctx, err := lock.Lock(ctx, -1)
	if err != nil {
		return err
	}
	defer func() {
		_ = lock.Unlock(ctx)
	}()

	elementKeys, err := store.kvStore.List(ctx, key + AppendEntryDelimiter)
	if err != nil {
		return err
	}

	/*
	 * Get the latest checkpoint value
	 */
	currCheckpoint, err := store.kvStore.Get(ctx, key)
	if err != nil && errors.Is(err, &common.NotFoundError{}) {
		currCheckpoint = nil
	} else if err != nil {
		return err
	}

	/*
	 * Compute the new checkpoint from the append entries
	 */
	newCheckpoint := currCheckpoint
	for _, elementKey := range elementKeys {
		valueBytes, err := store.kvStore.Get(ctx, elementKey)
		if err != nil {
			return errors.Wrap(err, "state store 1")
		}

		newCheckpoint, err = mapFn(newCheckpoint, valueBytes)
		if err != nil {
			return errors.Wrap(err, "state store 2")
		}
	}

	/*
	 * Update the current checkpoint
	 */
	err = store.kvStore.AtomicPut(ctx, key, currCheckpoint, newCheckpoint)
	if err != nil {
		return errors.Wrap(err, "state store 3")
	}

	/*
	 * Delete the append entries
	 */
	for _, elementKey := range elementKeys {
		err = store.kvStore.Delete(ctx, elementKey)
		if err != nil {
			return errors.Wrap(err, "state store 4")
		}
	}

	return nil
}

func (store *KvStateStore)  Get(ctx context.Context, key string) ([]byte, error) {
	return store.kvStore.Get(ctx, key)
}

func (store *KvStateStore)  Delete(ctx context.Context, key string) error {
	lock := NewKVDistributedLock(key, store.kvStore)
	ctx, err := lock.Lock(ctx, -1)
	if err != nil {
		return errors.Wrap(err, "state store 5")
	}
	defer func() {
		_ = lock.Unlock(ctx)
	}()
	return store.kvStore.Delete(ctx, key)
}

func (store *KvStateStore)  AtomicDelete(ctx context.Context, key string, prev []byte) error {
	lock := NewKVDistributedLock(key, store.kvStore)
	ctx, err := lock.Lock(ctx, -1)
	if err != nil {
		return errors.Wrap(err, "state store 6")
	}
	defer func() {
		_ = lock.Unlock(ctx)
	}()
	return store.kvStore.AtomicDelete(ctx, key, prev)
}
