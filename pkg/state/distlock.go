package state

import (
	"context"
	"fmt"
	"github.com/kmgreen2/agglo/pkg/kvs"
	"github.com/kmgreen2/agglo/pkg/util"
	"github.com/pkg/errors"
	"math"
	"strconv"
	"strings"
	"time"
)

func lockPrefix(id string) string {
	prefixFromIdLen := 4

	if len(id) < prefixFromIdLen {
		prefixFromIdLen = len(id)
	}

	return fmt.Sprintf("%slock:%s", id[:prefixFromIdLen], id)
}

type DistributedLock interface {
	Lock(ctx context.Context, timeout time.Duration) (context.Context, error)
	Unlock(ctx context.Context) error
	Locked(ctx context.Context) (bool, error)
	Close(ctx context.Context) error
}

type KVDistributedLock struct {
	id      string
	kvStore kvs.KVStore
}

func NewKVDistributedLock(id string, kvStore kvs.KVStore) *KVDistributedLock {
	return &KVDistributedLock{
		id:      id,
		kvStore: kvStore,
	}
}

func (l KVDistributedLock) getIndexKey(index int) string {
	return fmt.Sprintf("%s:%d", lockPrefix(l.id),index)
}

func (l KVDistributedLock) extractIndex(key string) (int, error) {
	splitKey := strings.Split(key, ":")
	if len(splitKey) < 3 {
		msg := fmt.Sprintf("'%s' not a valid lock key", key)
		return -1, util.NewInternalError(msg)
	}
	return strconv.Atoi(splitKey[len(splitKey) - 1])
}

func (l KVDistributedLock) getMaxIndex(entries []string) (int, error) {
	maxIndex := -1

	if len(entries) == 0 {
		return -1, util.NewInternalError("no entries to process")
	}

	for _, entry := range entries {
		idx, err := l.extractIndex(entry)
		if err != nil {
			return -1, err
		}
		if idx > maxIndex {
			maxIndex = idx
		}
	}
	return maxIndex, nil
}

func (l KVDistributedLock) getMinIndex(entries []string) (int, error) {
	minIndex := math.MaxInt32

	if len(entries) == 0 {
		return -1, util.NewInternalError("no entries to process")
	}

	for _, entry := range entries {
		idx, err := l.extractIndex(entry)
		if err != nil {
			return -1, err
		}
		if idx < minIndex {
			minIndex = idx
		}
	}
	return minIndex, nil
}

func (l KVDistributedLock) getWaiters(ctx context.Context) ([]string, error) {
	lockKey := fmt.Sprintf("%s", lockPrefix(l.id))

	// Get a listing of the current waiters
	entries, err := l.kvStore.List(ctx, lockKey)
	if err != nil && !errors.Is(err, &util.NotFoundError{}) {
		return nil, err
	}

	return entries, nil
}

func (l KVDistributedLock) getMinWaiter(ctx context.Context) (int, error) {
	entries, err := l.getWaiters(ctx)
	if err != nil {
		return -1, err
	}

	minIndex, err := l.getMinIndex(entries)
	if err != nil {
		return -1, err
	}

	return minIndex, nil
}

func (l KVDistributedLock) Lock(ctx context.Context, timeout time.Duration) (context.Context, error) {
	timeoutChannel := make(chan bool, 1)

	if timeout > -1 {
		timeoutTimer := time.NewTimer(timeout)
		defer timeoutTimer.Stop()
		go func() {
			<- timeoutTimer.C
			timeoutChannel <- true
		}()
	}

	// Get the largest index for the lock, and use 0 if there are no entries
	// The index determines the caller's place in line
	entries, err := l.getWaiters(ctx)
	if err != nil && !errors.Is(err, &util.NotFoundError{}){
		return ctx, errors.Wrap(err, "unable to get write lock waiters")
	}

	// Take a guess at the caller's place in line
	myIndex := 0
	if len(entries) > 0 {
		maxIndex, err := l.getMaxIndex(entries)
		if err != nil {
			return ctx, errors.Wrap(err, "unable to get write lock max index")
		}
		myIndex = maxIndex + 1
	}

	// Actually set the callers place in line, moving further back if others have arrived first
	myKey := l.getIndexKey(myIndex)
	for {
		err = l.kvStore.AtomicPut(ctx, myKey, nil, []byte("locked"))
		if err == nil {
			break
		} else if !errors.Is(err, &util.ConflictError{}) {
			return ctx, errors.Wrap(err, fmt.Sprintf("error setting write lock: %s", err.Error()))
		}
		myIndex++
		myKey = l.getIndexKey(myIndex)
	}

	// Wait until the caller is min waiter before returning, and time out if the caller
	// specified a timeout
	minIndex, err := l.getMinWaiter(ctx)
	if err != nil {
		return ctx, errors.Wrap(err, "error getting min write lock waiter")
	}

	for myIndex != minIndex {
		t := time.NewTimer(10 * time.Millisecond)
		select {
		case <- t.C:
		case <- timeoutChannel:
			t.Stop()
			return ctx, errors.Wrap(util.NewTimedOutError("lock timed out"), "lock operation timed out")
		}
		minIndex, err = l.getMinWaiter(ctx)
		if err != nil {
			return ctx, errors.Wrap(err, "error getting min write lock when waiting to acquire write lock")
		}
	}

	ctx = util.InjectDistributedLockIndex(ctx, myIndex)
	return ctx, nil
}

func (l KVDistributedLock) Unlock(ctx context.Context) error {
	entries, err := l.getWaiters(ctx)
	if err != nil {
		return err
	}
	minIndex, err := l.getMinIndex(entries)
	if err != nil {
		return err
	}
	idx := util.ExtractDistributedLockIndex(ctx)
	if idx != minIndex {
		return util.NewInvalidError("cannot call unlock when you do not hold the lock")
	}

	return l.kvStore.AtomicDelete(ctx, l.getIndexKey(minIndex), []byte("locked"))
}

func (l KVDistributedLock) Locked(ctx context.Context) (bool, error) {
	entries, err := l.getWaiters(ctx)
	if err != nil {
		return false, err
	}
	minIndex, err := l.getMinIndex(entries)
	if err != nil {
		return false, err
	}
	idx := util.ExtractDistributedLockIndex(ctx)
	if idx != minIndex {
		return false, nil
	}
	return true, nil
}

func (l KVDistributedLock) Close(ctx context.Context) error {
	return nil
}

