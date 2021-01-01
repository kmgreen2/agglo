package core

import (
	"github.com/kmgreen2/agglo/pkg/common"
	"time"
)

type Job interface {
	Run(delay time.Duration, sync bool, args ...interface{}) common.Future
}

type LocalJob struct {
	runnable common.PartialRunnable
}

func NewLocalJob(runnable common.PartialRunnable) *LocalJob {
	return &LocalJob{
		runnable: runnable,
	}
}

func (j LocalJob) Run(delay time.Duration, sync bool, args ...interface{}) common.Future {
	var future common.Future
	err := j.runnable.SetArgs(args...)
	if err != nil {
		completable := common.NewCompletable()
		// Note: This only fails if the completable is already completed.  Since this is
		// brand new, it should never fail here
		_ = completable.Fail(err)
		return completable.Future()
	}

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
