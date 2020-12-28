package common

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func waitForCallbacks(f Future) {
	for !f.CallbacksCompleted() {

	}
}

func TestCreateFuture(t *testing.T) {
	testSuccess := false
	testFailure := false
	testCancel := false
	runnable := NewSquareRunnable(10)
	f := CreateFuture(runnable).OnSuccess(func (x interface{}) {
		testSuccess = true
	}).OnFail(func (err error) {
		testFailure = true
	}).OnCancel(func () {
		testCancel = true
	})


	result, err := f.Get()
	assert.Nil(t, err)
	assert.Equal(t, 100, result)
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
	runnable := NewSleepAndFailRunnable(0)
	f := CreateFuture(runnable).OnSuccess(func (x interface{}) {
		testSuccess = true
	}).OnFail(func (err error) {
		testFailure = true
	}).OnCancel(func () {
		testCancel = true
	})

	result, err := f.Get()
	assert.Error(t, err)
	assert.Nil(t, result)
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
	runnable := NewSleepRunnable(2)
	f := CreateFuture(runnable).OnSuccess(func (x interface{}) {
		testSuccess = true
	}).OnFail(func (err error) {
		testFailure = true
	}).OnCancel(func () {
		testCancel = true
	})

	_ = f.Cancel()

	result, err := f.Get()
	assert.Error(t, err)
	assert.Nil(t, result)
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
	runnable := NewSquareRunnable(10)
	before := time.Now()
	f := CreateDeferredFuture(2*time.Second, runnable).
		OnSuccess(func (x interface{}) {
		testSuccess = true
	}).OnFail(func (err error) {
		testFailure = true
	}).OnCancel(func () {
		testCancel = true
	})

	result, err := f.Get()
	after := time.Now()
	assert.Nil(t, err)
	assert.Equal(t, 100, result)
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
	runnable := NewSleepRunnable(2)
	f := CreateFuture(runnable).OnSuccess(func (x interface{}) {
		testSuccess = true
	}).OnFail(func (err error) {
		testFailure = true
	}).OnCancel(func () {
		testCancel = true
	})

	_, err := f.GetWithTimeout(100 * time.Millisecond)
	assert.Error(t, err)
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
	runnable := NewSquareRunnable(10)
	f := CreateFuture(runnable).OnSuccess(func (x interface{}) {
		testSuccess = true
	}).OnFail(func (err error) {
		testFailure = true
	}).OnCancel(func () {
		testCancel = true
	}).Then(runnable).OnSuccess(func (x interface{}) {
		testThenSuccess = true
	}).OnFail(func (err error) {
		testThenFailure = true
	}).OnCancel(func () {
		testThenCancel = true
	})

	result, err := f.Get()
	assert.Nil(t, err)
	assert.Equal(t, 10000, result)

	waitForCallbacks(f)
	assert.True(t, testSuccess)
	assert.False(t, testFailure)
	assert.False(t, testCancel)
	assert.True(t, testThenSuccess)
	assert.False(t, testThenFailure)
	assert.False(t, testThenCancel)
	assert.True(t, f.IsCompleted())
}

func TestCreateFutureThenWithFailure(t *testing.T) {
	testFirstSuccess := false
	testSecondSuccess := false
	testFirstFailure := false
	testSecondFailure := false
	failRunnable := NewFailRunnable()
	runnable := NewSquareRunnable(10)
	f := CreateFuture(runnable).OnSuccess(func (x interface{}) {
		testFirstSuccess = true
	}).Then(failRunnable).OnFail(func (err error) {
		testFirstFailure = true
	}).Then(runnable).OnSuccess(func (x interface{}) {
		testSecondSuccess = true
	}).OnFail(func (err error) {
		testSecondFailure = true
	})

	_, err := f.Get()
	assert.Error(t, err)

	waitForCallbacks(f)
	assert.True(t, testFirstSuccess)
	assert.True(t, testFirstFailure)
	assert.True(t, testSecondFailure)
	assert.False(t, testSecondSuccess)
}

