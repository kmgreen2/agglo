package util_test

import (
	"context"
	"github.com/kmgreen2/agglo/pkg/util"
	"github.com/kmgreen2/agglo/test"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func waitForCallbacks(f util.Future) {
	for !f.CallbacksCompleted() {

	}
}

func TestCreateFuture(t *testing.T) {
	testSuccess := false
	testFailure := false
	testCancel := false
	runnable := test.NewSquareRunnable(10)
	f := util.CreateFuture(runnable).OnSuccess(func (context.Context, interface{}) {
		testSuccess = true
	}).OnFail(func (context.Context, error) {
		testFailure = true
	}).OnCancel(func (context.Context) {
		testCancel = true
	})


	result := f.Get()
	assert.Nil(t, result.Error())
	assert.Equal(t, 100, result.Value())
	assert.True(t, f.IsCompleted())

	waitForCallbacks(f)
	assert.True(t, testSuccess)
	assert.False(t, testCancel)
	assert.False(t, testFailure)
}

func TestCreateFutureFail(t *testing.T) {
	testSuccess := false
	testFailure := false
	testCancel := false
	runnable := test.NewSleepAndFailRunnable(0)
	f := util.CreateFuture(runnable).OnSuccess(func (context.Context, interface{}) {
		testSuccess = true
	}).OnFail(func (context.Context, error) {
		testFailure = true
	}).OnCancel(func (context.Context) {
		testCancel = true
	})

	result := f.Get()
	assert.Error(t, result.Error())
	assert.Nil(t, result.Value())
	assert.True(t, f.IsCompleted())

	waitForCallbacks(f)
	assert.False(t, testSuccess)
	assert.False(t, testCancel)
	assert.True(t, testFailure)
}

func TestCreateFutureCancel(t *testing.T) {
	testSuccess := false
	testFailure := false
	testCancel := false
	runnable := test.NewSleepRunnable(2)
	f := util.CreateFuture(runnable).OnSuccess(func (context.Context, interface{}) {
		testSuccess = true
	}).OnFail(func (context.Context, error) {
		testFailure = true
	}).OnCancel(func (context.Context) {
		testCancel = true
	})

	_ = f.Cancel(context.Background())

	result := f.Get()
	assert.Error(t, result.Error())
	assert.Nil(t, result.Value())
	assert.True(t, f.IsCancelled())

	waitForCallbacks(f)
	assert.False(t, testSuccess)
	assert.False(t, testFailure)
	assert.True(t, testCancel)
}

func TestCreateDeferredFuture(t *testing.T) {
	testSuccess := false
	testFailure := false
	testCancel := false
	runnable := test.NewSquareRunnable(10)
	before := time.Now()
	f := util.CreateFuture(runnable, util.WithDelay(2*time.Second)).
		OnSuccess(func (context.Context, interface{}) {
		testSuccess = true
	}).OnFail(func (context.Context, error) {
		testFailure = true
	}).OnCancel(func (context.Context) {
		testCancel = true
	})

	result := f.Get()
	after := time.Now()
	assert.Nil(t, result.Error())
	assert.Equal(t, 100, result.Value())
	assert.True(t, f.IsCompleted())
	assert.True(t, after.Sub(before) > 2 * time.Second)

	waitForCallbacks(f)
	assert.True(t, testSuccess)
	assert.False(t, testFailure)
	assert.False(t, testCancel)
}

func TestCreateFutureTimeout(t *testing.T) {
	testSuccess := false
	testFailure := false
	testCancel := false
	runnable := test.NewSleepRunnable(2)
	f := util.CreateFuture(runnable).OnSuccess(func (context.Context, interface{}) {
		testSuccess = true
	}).OnFail(func (context.Context, error) {
		testFailure = true
	}).OnCancel(func (context.Context) {
		testCancel = true
	})

	result := f.GetWithTimeout(100 * time.Millisecond)
	assert.Error(t, result.Error())
	assert.False(t, f.IsCompleted())

	assert.False(t, testSuccess)
	assert.False(t, testFailure)
	assert.False(t, testCancel)
}

func TestCreateFutureThen(t *testing.T) {
	testSuccess := false
	testFailure := false
	testCancel := false
	testThenSuccess := false
	testThenFailure := false
	testThenCancel := false
	runnable := test.NewSquareRunnable(10)
	f1 := util.CreateFuture(runnable).OnSuccess(func (context.Context, interface{}) {
		testSuccess = true
	}).OnFail(func (context.Context, error) {
		testFailure = true
	}).OnCancel(func (context.Context) {
		testCancel = true
	})

	f2 := f1.Then(runnable).OnSuccess(func (context.Context, interface{}) {
		testThenSuccess = true
	}).OnFail(func (context.Context, error) {
		testThenFailure = true
	}).OnCancel(func (context.Context) {
		testThenCancel = true
	})

	result := f2.Get()
	assert.Nil(t, result.Error())
	assert.Equal(t, 10000, result.Value())

	waitForCallbacks(f1)
	waitForCallbacks(f2)
	assert.True(t, testSuccess)
	assert.False(t, testFailure)
	assert.False(t, testCancel)
	assert.True(t, testThenSuccess)
	assert.False(t, testThenFailure)
	assert.False(t, testThenCancel)
	assert.True(t, f1.IsCompleted())
	assert.True(t, f2.IsCompleted())
}

func TestCreateFutureThenWithFailure(t *testing.T) {
	testFirstSuccess := false
	testSecondSuccess := false
	testFirstFailure := false
	testSecondFailure := false
	failRunnable := test.NewFailRunnable()
	runnable := test.NewSquareRunnable(10)
	f1 := util.CreateFuture(runnable).OnSuccess(func (context.Context, interface{}) {
		testFirstSuccess = true
	})
	f2 := f1.Then(failRunnable).OnFail(func (context.Context, error) {
		testFirstFailure = true
	})
	f3 := f2.Then(runnable).OnSuccess(func (context.Context, interface{}) {
		testSecondSuccess = true
	}).OnFail(func (context.Context, error) {
		testSecondFailure = true
	})

	result := f3.Get()
	assert.Error(t, result.Error())

	waitForCallbacks(f1)
	waitForCallbacks(f2)
	waitForCallbacks(f3)
	assert.True(t, testFirstSuccess)
	assert.True(t, testFirstFailure)
	assert.True(t, testSecondFailure)
	assert.False(t, testSecondSuccess)
}

func TestRetryableFuture(t *testing.T) {
	testSuccess := false
	testFailure := false
	testCancel := false
	runnable := test.NewFailThenSucceedRunnable(2)
	f := util.CreateFuture(
		runnable, util.WithRetry(3, 100 * time.Millisecond)).OnSuccess(func (context.Context, interface{}) {

		testSuccess = true
	}).OnFail(func (context.Context, error) {
		testFailure = true
	}).OnCancel(func (context.Context) {
		testCancel = true
	})


	result := f.Get()
	assert.Nil(t, result.Error())
	assert.Equal(t, 2, result.Value())
	assert.True(t, f.IsCompleted())

	waitForCallbacks(f)
	assert.True(t, testSuccess)
	assert.False(t, testCancel)
	assert.False(t, testFailure)
}

func TestThenRetryableFuture(t *testing.T) {
	testSuccess := false
	testFailure := false
	testCancel := false
	runnable := test.NewSquareRunnable(2)
	thenRunnable := test.NewFailThenSucceedRunnable(2)


	f := util.CreateFuture(runnable).Then(thenRunnable, util.WithRetry(3, 100 * time.Millisecond)).
	OnSuccess(func (context.Context, interface{}) {
		testSuccess = true
	}).OnFail(func (context.Context, error) {
		testFailure = true
	}).OnCancel(func (context.Context) {
		testCancel = true
	})


	result := f.Get()
	assert.Nil(t, result.Error())
	assert.Equal(t, 2, result.Value())
	assert.True(t, f.IsCompleted())

	waitForCallbacks(f)
	assert.True(t, testSuccess)
	assert.False(t, testCancel)
	assert.False(t, testFailure)
}

func TestThenRetryableFutureFailed(t *testing.T) {
	testSuccess := false
	testFailure := false
	testCancel := false
	runnable := test.NewSquareRunnable(2)
	thenRunnable := test.NewFailThenSucceedRunnable(4)


	f := util.CreateFuture(runnable).Then(thenRunnable, util.WithRetry(3, 100 * time.Millisecond)).
		OnSuccess(func (context.Context, interface{}) {
			testSuccess = true
		}).OnFail(func (context.Context, error) {
		testFailure = true
	}).OnCancel(func (context.Context) {
		testCancel = true
	})


	result := f.Get()
	assert.Error(t, result.Error())
	assert.Nil(t, result.Value())
	assert.True(t, f.IsCompleted())

	waitForCallbacks(f)
	assert.False(t, testSuccess)
	assert.False(t, testCancel)
	assert.True(t, testFailure)
}

func TestCallbackAfterFutureCompletes(t *testing.T) {
	testSuccess := false
	testFailure := false
	testCancel := false
	runnable := test.NewSquareRunnable(10)
	f := util.CreateFuture(runnable)

	waitForCallbacks(f)

	f.OnSuccess(func (context.Context, interface{}) {
		testSuccess = true
	}).OnFail(func (context.Context, error) {
		testFailure = true
	}).OnCancel(func (context.Context) {
		testCancel = true
	})

	result := f.Get()
	assert.Nil(t, result.Error())
	assert.Equal(t, 100, result.Value())
	assert.True(t, f.IsCompleted())

	assert.True(t, testSuccess)
	assert.False(t, testCancel)
	assert.False(t, testFailure)
}

func TestThenAfterFutureComplete(t *testing.T) {
	testSuccess := false
	testFailure := false
	testCancel := false
	testThenSuccess := false
	testThenFailure := false
	testThenCancel := false
	runnable := test.NewSquareRunnable(10)
	f1 := util.CreateFuture(runnable).OnSuccess(func (context.Context, interface{}) {
		testSuccess = true
	}).OnFail(func (context.Context, error) {
		testFailure = true
	}).OnCancel(func (context.Context) {
		testCancel = true
	})

	waitForCallbacks(f1)

	f2 := f1.Then(runnable).OnSuccess(func (context.Context, interface{}) {
		testThenSuccess = true
	}).OnFail(func (context.Context, error) {
		testThenFailure = true
	}).OnCancel(func (context.Context) {
		testThenCancel = true
	})

	result := f2.Get()
	assert.Nil(t, result.Error())
	assert.Equal(t, 10000, result.Value())

	waitForCallbacks(f2)
	assert.True(t, testSuccess)
	assert.False(t, testFailure)
	assert.False(t, testCancel)
	assert.True(t, testThenSuccess)
	assert.False(t, testThenFailure)
	assert.False(t, testThenCancel)
	assert.True(t, f1.IsCompleted())
	assert.True(t, f2.IsCompleted())
}

