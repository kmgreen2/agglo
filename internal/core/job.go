package core

import (
	"context"
	"fmt"
	"github.com/kmgreen2/agglo/pkg/util"
	"time"
)

type Job interface {
	Run(delay time.Duration, sync bool, inData ...interface{}) util.Future
}

type LocalJob struct {
	runnable util.PartialRunnable
	cmdArgs  []string
}

func NewLocalJob(runnable util.PartialRunnable) *LocalJob {
	return &LocalJob{
		runnable: runnable,
	}
}

func run(runnable util.PartialRunnable, delay time.Duration, sync bool, inData interface{}) util.Future {
	completable := util.NewCompletable()
	var future util.Future
	if inData != nil {
		err := runnable.SetInData(inData)
		if err != nil {
			// Note: This only fails if the completable is already completed.  Since this is
			// brand new, it should never fail here
			_ = completable.Fail(context.Background(), err)
			return completable.Future()
		}
	}

	if delay > 0 {
		future = util.CreateFuture(runnable, util.WithDelay(delay))
	} else {
		future = util.CreateFuture(runnable)
	}

	if sync {
		util.WaitAll([]util.Future{future}, -1)
	}
	return future
}

func (j LocalJob) Run(delay time.Duration, sync bool, inData ...interface{}) util.Future {
	if len(inData) > 1 {
		completable := util.NewCompletable()
		msg := fmt.Sprintf("expected 1 inData varadic arg to Run, got %d", len(inData))
		_ = completable.Fail(context.Background(), util.NewInvalidError(msg))
		return completable.Future()
	} else if len(inData) == 1 {
		return run(j.runnable, delay, sync, inData[0])
	}
	return run(j.runnable, delay, sync, nil)
}

type CmdJob struct {
	runnable util.PartialRunnable
}

func NewCmdJob(cmdPath string, cmdArgs ...string) *CmdJob {
	runnable := util.NewExecRunnable(util.WithCmdArgs(cmdArgs...), util.WithPath(cmdPath))
	return &CmdJob{
		runnable: runnable,
	}
}

func (j CmdJob) Run(delay time.Duration, sync bool, inData ...interface{}) util.Future {
	if len(inData) > 1 {
		completable := util.NewCompletable()
		msg := fmt.Sprintf("expected 1 inData varadic arg to Run, got %d", len(inData))
		_ = completable.Fail(context.Background(), util.NewInvalidError(msg))
		return completable.Future()
	} else if len(inData) == 1 {
		return run(j.runnable, delay, sync, inData[0])
	}
	return run(j.runnable, delay, sync, nil)
}
