package core

import (
	"context"
	"fmt"
	"github.com/kmgreen2/agglo/pkg/common"
	"time"
)

type Job interface {
	Run(delay time.Duration, sync bool, inData ...interface{}) common.Future
}

type LocalJob struct {
	runnable common.PartialRunnable
	cmdArgs []string
}

func NewLocalJob(runnable common.PartialRunnable) *LocalJob {
	return &LocalJob{
		runnable: runnable,
	}
}

func run(runnable common.PartialRunnable, delay time.Duration, sync bool, inData interface{}) common.Future {
	completable := common.NewCompletable()
	var future common.Future
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
		future = common.CreateFuture(runnable, common.WithDelay(delay))
	} else {
		future = common.CreateFuture(runnable)
	}

	if sync {
		common.WaitAll([]common.Future{future}, -1)
	}
	return future
}

func (j LocalJob) Run(delay time.Duration, sync bool, inData ...interface{}) common.Future {
	if len(inData) > 1 {
		completable := common.NewCompletable()
		msg := fmt.Sprintf("expected 1 inData varadic arg to Run, got %d", len(inData))
		_ = completable.Fail(context.Background(), common.NewInvalidError(msg))
		return completable.Future()
	} else if len(inData) == 1 {
		return run(j.runnable, delay, sync, inData[0])
	}
	return run(j.runnable, delay, sync, nil)
}

type CmdJob struct {
	runnable common.PartialRunnable
}

func NewCmdJob(cmdPath string, cmdArgs ...string) *CmdJob {
	runnable := common.NewExecRunnable(common.WithCmdArgs(cmdArgs...), common.WithPath(cmdPath))
	return &CmdJob{
		runnable: runnable,
	}
}

func (j CmdJob) Run(delay time.Duration, sync bool, inData ...interface{}) common.Future {
	if len(inData) > 1 {
		completable := common.NewCompletable()
		msg := fmt.Sprintf("expected 1 inData varadic arg to Run, got %d", len(inData))
		_ = completable.Fail(context.Background(), common.NewInvalidError(msg))
		return completable.Future()
	} else if len(inData) == 1 {
		return run(j.runnable, delay, sync, inData[0])
	}
	return run(j.runnable, delay, sync, nil)
}
