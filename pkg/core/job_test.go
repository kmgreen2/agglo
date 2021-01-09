package core_test

import (
	"github.com/kmgreen2/agglo/pkg/core"
	"github.com/kmgreen2/agglo/test"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestLocalJob(t *testing.T) {
	runnable := test.NewSleepRunnable(1)
	job := core.NewLocalJob(runnable)

	f := job.Run(-1, false)

	result := f.Get()
	assert.Nil(t, result.Error())
	intResult, ok := result.Value().(int)
	assert.True(t, ok)
	assert.Equal(t, 1, intResult)
}

func TestLocalJobSync(t *testing.T) {
	runnable := test.NewSleepRunnable(1)
	job := core.NewLocalJob(runnable)

	start := time.Now()
	f := job.Run(-1, true)
	end := time.Now()

	assert.True(t, end.Sub(start) > 1 * time.Second)

	result := f.Get()
	assert.Nil(t, result.Error())
	intResult, ok := result.Value().(int)
	assert.True(t, ok)
	assert.Equal(t, 1, intResult)
}

func TestLocalJobFailed(t *testing.T) {
	runnable := test.NewFailRunnable()
	job := core.NewLocalJob(runnable)

	f := job.Run(-1, false)

	result := f.Get()
	assert.Error(t, result.Error())
}

func TestLocalJobDeferred(t *testing.T) {
	runnable := test.NewSleepRunnable(1)
	job := core.NewLocalJob(runnable)

	start := time.Now()
	f := job.Run(1*time.Second, true)
	end := time.Now()

	assert.True(t, end.Sub(start) > 2 * time.Second)

	result := f.Get()
	assert.Nil(t, result.Error())
	intResult, ok := result.Value().(int)
	assert.True(t, ok)
	assert.Equal(t, 1, intResult)
}
