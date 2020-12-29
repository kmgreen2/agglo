package core

import (
	"github.com/kmgreen2/agglo/pkg/common"
	"time"
)

type Job interface {
	Run(delay time.Duration, sync bool) error
}

type LocalJob struct {
	runnable common.Runnable
}

func NewLocalJob(runnable common.Runnable) *LocalJob {
	return &LocalJob{
		runnable: runnable,
	}
}

func (j LocalJob) Run(delay time.Duration, sync bool) common.Future {
	var future common.Future
	if delay > 0 {
		future = common.CreateDeferredFuture(delay, j.runnable)
	} else {
		future = common.CreateFuture(j.runnable)
	}

	if sync {
		common.WaitAll([]common.Future{future}, -1)
	}
	return future
}
