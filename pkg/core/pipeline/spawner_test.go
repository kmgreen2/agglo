package pipeline_test

import (
	"github.com/kmgreen2/agglo/pkg/core"
	"github.com/kmgreen2/agglo/pkg/core/pipeline"
	"github.com/kmgreen2/agglo/test"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestSpawnerHappyPathAsync(t *testing.T) {
	in, _ := test.GenRandomMap(3, 16)
	runnable := test.NewSleepRunnable(1)
	job := core.NewLocalJob(runnable)

	spawner := pipeline.NewSpawner(job, core.TrueCondition, -1, false)

	out, err := spawner.Process(in)

	assert.Nil(t, err)
	assert.Equal(t, in, out)
}

func TestSpawnerHappyPathFalseCondition(t *testing.T) {
	in, _ := test.GenRandomMap(3, 16)
	runnable := test.NewSleepRunnable(1)
	job := core.NewLocalJob(runnable)

	spawner := pipeline.NewSpawner(job, core.FalseCondition, 10*time.Second, true)

	start := time.Now()
	out, err := spawner.Process(in)
	end := time.Now()

	assert.True(t, end.Sub(start) < 1*time.Second)

	assert.Nil(t, err)
	assert.Equal(t, in, out)
}

func TestSpawnerHappyPathSync(t *testing.T) {
	in, _ := test.GenRandomMap(3, 16)
	runnable := test.NewSleepRunnable(1)
	job := core.NewLocalJob(runnable)

	spawner := pipeline.NewSpawner(job, core.TrueCondition, -1, true)


	start := time.Now()
	out, err := spawner.Process(in)
	end := time.Now()

	assert.True(t, end.Sub(start) > 1 * time.Second)
	assert.Nil(t, err)
	assert.Equal(t, in, out)
}

func TestSpawnerFailSync(t *testing.T) {
	in, _ := test.GenRandomMap(3, 16)
	runnable := test.NewFailRunnable()
	job := core.NewLocalJob(runnable)

	spawner := pipeline.NewSpawner(job, core.TrueCondition, -1, true)

	out, err := spawner.Process(in)

	assert.Error(t, err)
	assert.Equal(t, in, out)
}

func TestSpawnerHappyPathDelaySync(t *testing.T) {
	in, _ := test.GenRandomMap(3, 16)
	runnable := test.NewSquareRunnable(2)
	job := core.NewLocalJob(runnable)

	spawner := pipeline.NewSpawner(job, core.TrueCondition, 1*time.Second, true)


	start := time.Now()
	out, err := spawner.Process(in)
	end := time.Now()

	assert.True(t, end.Sub(start) > 1 * time.Second)
	assert.Nil(t, err)
	assert.Equal(t, in, out)
}