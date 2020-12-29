package common_test

import (
	"github.com/kmgreen2/agglo/pkg/common"
	"github.com/kmgreen2/agglo/test"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func makeFutures(numFast, numSlow, slowDuration int) []common.Future {
	futures := make([]common.Future, numFast+numSlow)

	for i := 0; i < numFast; i++ {
		runnable := test.NewSquareRunnable(2)
		futures[i] = common.CreateFuture(runnable)
	}

	for i := numSlow; i < numSlow+numFast; i++ {
		runnable := test.NewSleepRunnable(slowDuration)
		futures[i] = common.CreateFuture(runnable)
	}

	return futures
}

func TestWaitAll(t *testing.T) {
	futures := makeFutures(10, 10, 1)
	common.WaitAll(futures, -1)

	for _, future := range futures {
		assert.True(t, future.IsCompleted())
	}
}

func TestWaitAllWithTimeOut(t *testing.T) {
	futures := makeFutures(10, 10, 1)
	common.WaitAll(futures, 20 * time.Second)

	for _, future := range futures {
		assert.True(t, future.IsCompleted())
	}
}

func TestWaitAllTimedOut(t *testing.T) {
	futures := makeFutures(10, 10, 1)
	common.WaitAll(futures, 10 * time.Millisecond)

	for i, future := range futures {
		if i >= 10 {
			assert.False(t, future.IsCompleted())
		}
	}
}
