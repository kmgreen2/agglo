package common

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestCreateFuture(t *testing.T) {
	testSuccess := false
	runnable := NewSquareRunnable(10)
	f := CreateFuture(runnable).OnSuccess(func (x interface{}) {
		testSuccess = true
	})

	result, err := f.Get()
	assert.Nil(t, err)
	assert.Equal(t, 100, result)
	assert.True(t, testSuccess)
}

func TestCreateDeferredFuture(t *testing.T) {
	testSuccess := false
	runnable := NewSquareRunnable(10)
	before := time.Now()
	f := CreateDeferredFuture(2*time.Second, runnable).
		OnSuccess(func (x interface{}) {
		testSuccess = true
	})

	result, err := f.Get()
	after := time.Now()
	assert.Nil(t, err)
	assert.Equal(t, 100, result)
	assert.True(t, testSuccess)
	assert.True(t, after.Sub(before) > 2 * time.Second)
}

func TestCreateFutureTimeout(t *testing.T) {
	testSuccess := false
	runnable := NewSleepRunnable(2)
	f := CreateFuture(runnable).OnSuccess(func (x interface{}) {
		testSuccess = true
	})

	_, err := f.GetWithTimeout(time.Second)
	assert.Error(t, err)
	assert.False(t, testSuccess)
}

func TestCreateFutureThen(t *testing.T) {
	testSuccess := false
	runnable := NewSquareRunnable(10)
	f := CreateFuture(runnable).OnSuccess(func (x interface{}) {
		testSuccess = true
	}).Then(runnable)

	result, err := f.Get()
	assert.Nil(t, err)
	assert.Equal(t, 10000, result)
	assert.True(t, testSuccess)
}

