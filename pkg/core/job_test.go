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

	result, err := f.Get()
	assert.Nil(t, err)
	intResult, ok := result.(int)
	assert.True(t, ok)
	assert.Equal(t, 1, intResult)
}

func TestLocalJobSync(t *testing.T) {
	runnable := test.NewSleepRunnable(1)
	job := core.NewLocalJob(runnable)

	start := time.Now()
	f := job.Run(-1, true)
	end := time.Now()

	assert.True(t, end.Sub(start) > 1)

	result, err := f.Get()
	assert.Nil(t, err)
	intResult, ok := result.(int)
	assert.True(t, ok)
	assert.Equal(t, 1, intResult)
}

func TestLocalJobFailed(t *testing.T) {
	runnable := test.NewFailRunnable()
	job := core.NewLocalJob(runnable)

	f := job.Run(-1, false)

	_, err := f.Get()
	assert.Error(t, err)
}

func TestLocalJobDeferred(t *testing.T) {
	runnable := test.NewSleepRunnable(1)
	job := core.NewLocalJob(runnable)

	start := time.Now()
	f := job.Run(1*time.Second, true)
	end := time.Now()

	assert.True(t, end.Sub(start) > 2)

	result, err := f.Get()
	assert.Nil(t, err)
	intResult, ok := result.(int)
	assert.True(t, ok)
	assert.Equal(t, 1, intResult)
}
