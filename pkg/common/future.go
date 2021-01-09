package common

import (
	"context"
	"sync"
	"time"
)

type futureState int

const (
	undefinedFuture futureState = iota
	succeededFuture
	failedFuture
	canceledFuture
)

type Future interface {
	Get() *FutureResult
	GetWithTimeout(duration time.Duration) *FutureResult
	Then(runnable PartialRunnable) Future
	ThenWithRetry(numRetries int, initialDelay time.Duration, runnable PartialRunnable) Future
	Cancel() error
	IsCancelled() bool
	IsCompleted() bool
	CallbacksCompleted() bool
	OnSuccess(func (interface{})) Future
	OnCancel(func ()) Future
	OnFail(func (err error)) Future
}

type FutureResult struct {
	ctx context.Context
	value interface{}
	err error
}

func (result FutureResult) Error() error {
	return result.err
}

func (result FutureResult) Value() interface{} {
	return result.value
}

func newFutureResult(value interface{}) *FutureResult {
	return &FutureResult{
		value: value,
		err: nil,
	}
}

func newCancelledFuture() *FutureResult {
	return &FutureResult{
		value: nil,
		err: NewCancelledError(),
	}
}

func newFailedFuture(err error) *FutureResult {
	return &FutureResult{
		value: nil,
		err: err,
	}
}

type future struct {
	result chan *FutureResult
	finalResult *FutureResult
	mutex sync.Mutex
	state futureState
	callbacksCompleted bool
	successes []func (interface{})
	fails []func (error)
	cancels []func ()
	ctx context.Context
}

func newFuture() *future {
	return &future{
		result: make(chan *FutureResult, 1),
		mutex: sync.Mutex{},
		state: undefinedFuture,
		callbacksCompleted: false,
	}
}

func completeRunnable(ctx context.Context, completable Completable, runnable Runnable) {
	defer completable.Close()
	result, err := runnable.Run(ctx)
	if err != nil {
		_ = completable.Fail(err)
	} else {
		_ = completable.Success(result)
	}
}


/// ToDo(KMG): Refactor the future creation  code to unify creation under the same helper using options.
/// there is no need to keep the code separate.  Holding off right now while working through another
//  refactor.  Also add in a WithPrepare option that will run a function before running the underlying function

type retryOption struct {
	num int
	initialDelay time.Duration
}

type FutureOptions struct {
	ctx context.Context
	delay time.Duration
	retry retryOption
}

type FutureOption func(opts* FutureOptions)

func createFuture(runnable Runnable, options ...FutureOption) Future {
	completable := NewCompletable()
	futureOptions := &FutureOptions{
		ctx: context.Background(),
	}

	for _, opt := range options {
		opt(futureOptions)
	}
	return completable.Future()
}
//////

func CreateFuture(ctx context.Context, runnable Runnable) Future {
	completable := NewCompletable()

	go func() {
		completeRunnable(ctx, completable, runnable)
	}()
	return completable.Future()
}

func CreateDeferredFuture(ctx context.Context, duration time.Duration, runnable Runnable) Future {
	completable := NewCompletable()

	go func() {
		defer completable.Close()
		t := time.NewTimer(duration)
		<-t.C
		result, err := runnable.Run(ctx)
		if err != nil {
			_ = completable.Fail(err)
		} else {
			_ = completable.Success(result)
		}
	}()

	return completable.Future()
}

func CreateRetryableFuture(ctx context.Context, numRetries int, initialDelay time.Duration, runnable Runnable) Future {
	completable := NewCompletable()

	go func() {
		var err error
		var result interface{}
		defer completable.Close()
		delay := initialDelay
		for i := 0; i <= numRetries; i++ {
			result, err = runnable.Run(ctx)
			if err != nil {
				time.Sleep(delay)
				delay <<= 1
				continue
			} else {
				_ = completable.Success(result)
				return
			}
		}
		_ = completable.Fail(err)
	}()
	return completable.Future()
}

func (f *future) Get() *FutureResult {
	return f.getWithContext(context.Background())
}

func (f *future) GetWithTimeout(timeout time.Duration) *FutureResult {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return f.getWithContext(ctx)
}

func (f *future) ThenWithRetry(numRetries int, initialDelay time.Duration, runnable PartialRunnable) Future {
	next := NewCompletable()

	go func() {
		var err error
		var thenResult interface{}
		defer next.Close()

		delay := initialDelay
		result := f.Get()
		if result.err != nil {
			_ = next.Fail(result.err)
			return
		}
		err = runnable.SetInData(result.value)
		if err != nil {
			_ = next.Fail(result.err)
			return
		}
		for i := 0; i < numRetries; i++ {
			thenResult, err = runnable.Run(result.ctx)
			if err != nil {
				time.Sleep(delay)
				delay <<= 1
				continue
			} else {
				_ = next.Success(thenResult)
				return
			}
		}
		_ = next.Fail(err)
	}()
	return next.Future()
}

func (f *future) Then(runnable PartialRunnable) Future {
	var err error
	next := NewCompletable()

	go func() {
		result := f.Get()
		if result.err != nil {
			_ = next.Fail(result.err)
		} else {
			err = runnable.SetInData(result.value)
			if err != nil {
				_ = next.Fail(err)
			} else {
				completeRunnable(result.ctx, next, runnable)
			}
		}
	}()
	return next.Future()
}

func (f *future) getWithContext(ctx context.Context) *FutureResult {
	select {
	case <-ctx.Done():
		return &FutureResult{
			err: ctx.Err(),
		}
	case <-f.result:
		return f.finalResult
	}
}

func (f *future) IsCancelled() bool {
	return f.state == canceledFuture
}

func (f *future) IsCompleted() bool {
	return f.state == succeededFuture || f.state == failedFuture
}

func (f *future) CallbacksCompleted() bool {
	return f.callbacksCompleted
}

func (f *future) OnSuccess(cb func (interface{})) Future {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	f.successes = append(f.successes, cb)
	return f
}

func (f *future) OnCancel(cb func ()) Future {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	f.cancels = append(f.cancels, cb)
	return f
}

func (f *future) OnFail(cb func (err error)) Future {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	f.fails = append(f.fails, cb)
	return f
}

func (f *future) success(result interface{}) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if f.state != undefinedFuture {
		return NewAlreadyCompletedError("Cannot succeed a future that has already completed")
	}

	f.finalResult = newFutureResult(result)
	f.state = succeededFuture
	f.doCallbacks()
	f.result <- f.finalResult
	return nil
}

func (f *future) fail(err error) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if f.state != undefinedFuture {
		return NewAlreadyCompletedError("Cannot fail a future that has already completed")
	}

	f.finalResult = newFailedFuture(err)
	f.state = failedFuture
	f.doCallbacks()
	f.result <- f.finalResult
	return nil
}

func (f *future) Cancel() error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if f.state != undefinedFuture {
		return NewAlreadyCompletedError("Cannot cancel future that has already completed")
	}

	f.finalResult = newCancelledFuture()
	f.state = canceledFuture
	f.doCallbacks()
	f.result <- f.finalResult
	return nil
}

func (f *future) close() {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	close(f.result)
}

func (f *future) doCallbacks() {
	wg := &sync.WaitGroup{}

	defer func () {
		go func() {
			wg.Wait()
			f.callbacksCompleted = true
		}()
	}()

	switch f.state {
	case succeededFuture:
		for _, cb := range f.successes {
			wg.Add(1)
			go func(callback func(interface{})) {
				callback(f.finalResult.value)
				wg.Done()
			}(cb)
		}
	case failedFuture:
		for _, cb := range f.fails {
			wg.Add(1)
			go func(callback func(error)) {
				callback(f.finalResult.err)
				wg.Done()
			}(cb)
		}
	case canceledFuture:
		for _, cb := range f.cancels {
			wg.Add(1)
			go func(callback func()) {
				callback()
				wg.Done()
			}(cb)
		}
	default:
		return
	}
}
