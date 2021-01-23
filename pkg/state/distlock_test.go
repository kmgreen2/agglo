package state_test

import (
	"context"
	gUuid "github.com/google/uuid"
	"github.com/kmgreen2/agglo/pkg/util"
	"github.com/kmgreen2/agglo/pkg/kvs"
	"github.com/kmgreen2/agglo/pkg/state"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
	goSync "sync"
)

func TestKVDistributedLockHappyPath(t *testing.T) {
	numThreads := 100
	kvStore := kvs.NewMemKVStore()
	uuid, err := gUuid.NewRandom()
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	lock := state.NewKVDistributedLock(uuid.String(), kvStore)

	wg := goSync.WaitGroup{}
	wg.Add(numThreads)
	startTime := time.Now()
	for i := 0; i < numThreads; i++ {
		go func() {
			ctx := context.Background()
			ctx, err := lock.Lock(ctx, -1)
			if err != nil {
				assert.FailNow(t, err.Error())
			}
			time.Sleep(10 * time.Millisecond)
			err = lock.Unlock(ctx)
			if err != nil {
				assert.FailNow(t, err.Error())
			}
			wg.Done()
		}()
	}
	wg.Wait()
	assert.True(t, time.Now().Sub(startTime).Milliseconds() > int64(10 * numThreads))
}

func TestKVDistributedLockTimeout(t *testing.T) {
	kvStore := kvs.NewMemKVStore()
	uuid, err := gUuid.NewRandom()
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	lock := state.NewKVDistributedLock(uuid.String(), kvStore)

	ctx := context.Background()
	ctx, err = lock.Lock(ctx, -1)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	ctx = context.Background()
	ctx, err = lock.Lock(ctx, 50 * time.Millisecond)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, &util.TimedOutError{}))
}

func TestKVDistributedLockBadUnlock(t *testing.T) {
	kvStore := kvs.NewMemKVStore()
	uuid, err := gUuid.NewRandom()
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	lock := state.NewKVDistributedLock(uuid.String(), kvStore)

	ctx := context.Background()
	ctx, err = lock.Lock(ctx, -1)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	ctx = context.Background()
	err = lock.Unlock(ctx)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, &util.InvalidError{}))
}

