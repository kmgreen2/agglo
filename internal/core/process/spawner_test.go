package process_test

import (
	"context"
	"github.com/kmgreen2/agglo/internal/common"
	"github.com/kmgreen2/agglo/internal/core"
	"github.com/kmgreen2/agglo/internal/core/process"
	"github.com/kmgreen2/agglo/test"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestSpawnerHappyPathAsync(t *testing.T) {
	in, _ := test.GenRandomMap(3, 16)
	runnable := test.NewSleepRunnable(1)
	job := core.NewLocalJob(runnable)

	spawner := process.NewSpawner("foo", job, core.TrueCondition, -1, false)

	out, err := spawner.Process(context.Background(), in)

	assert.Nil(t, err)
	delete(out, string(common.SpawnMetadataKey))
	assert.Equal(t, in, out)
}

func TestSpawnerHappyPathFalseCondition(t *testing.T) {
	in, _ := test.GenRandomMap(3, 16)
	runnable := test.NewSleepRunnable(1)
	job := core.NewLocalJob(runnable)

	spawner := process.NewSpawner("foo", job, core.FalseCondition, 10*time.Second, true)

	start := time.Now()
	out, err := spawner.Process(context.Background(), in)
	end := time.Now()

	assert.True(t, end.Sub(start) < 1*time.Second)

	assert.Nil(t, err)
	delete(out, string(common.SpawnMetadataKey))
	assert.Equal(t, in, out)
}

func TestSpawnerHappyPathSync(t *testing.T) {
	in, _ := test.GenRandomMap(3, 16)
	runnable := test.NewSleepRunnable(1)
	job := core.NewLocalJob(runnable)

	spawner := process.NewSpawner("foo", job, core.TrueCondition, -1, true)


	start := time.Now()
	out, err := spawner.Process(context.Background(), in)
	end := time.Now()

	assert.True(t, end.Sub(start) > 1 * time.Second)
	assert.Nil(t, err)
	delete(out, string(common.SpawnMetadataKey))
	assert.Equal(t, in, out)
}

func TestSpawnerFailSync(t *testing.T) {
	in, _ := test.GenRandomMap(3, 16)
	runnable := test.NewFailRunnable()
	job := core.NewLocalJob(runnable)

	spawner := process.NewSpawner("foo", job, core.TrueCondition, -1, true)

	out, err := spawner.Process(context.Background(), in)

	assert.Error(t, err)
	assert.Equal(t, in, out)
}

func TestSpawnerHappyPathDelaySync(t *testing.T) {
	in, _ := test.GenRandomMap(3, 16)
	runnable := test.NewSleepRunnable(1)
	job := core.NewLocalJob(runnable)

	spawner := process.NewSpawner("foo", job, core.TrueCondition, 1*time.Second, true)


	start := time.Now()
	out, err := spawner.Process(context.Background(), in)
	end := time.Now()

	assert.True(t, end.Sub(start) > 2 * time.Second)
	assert.Nil(t, err)
	delete(out, string(common.SpawnMetadataKey))
	assert.Equal(t, in, out)
}