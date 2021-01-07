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
	Get() (interface{}, error)
	GetWithTimeout(duration time.Duration) (interface{}, error)
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

type futureResult struct {
	value interface{}
	err error
}

func newFutureResult(value interface{}) *futureResult {
	return &futureResult{
		value: value,
		err: nil,
	}
}

func newCancelledFuture() *futureResult {
	return &futureResult{
		value: nil,
		err: NewCancelledError(),
	}
}

func newFailedFuture(err error) *futureResult {
	return &futureResult{
		value: nil,
		err: err,
	}
}

type future struct {
	result chan *futureResult
	finalResult *futureResult
	mutex sync.Mutex
	state futureState
	callbacksCompleted bool
	successes []func (interface{})
	fails []func (error)
	cancels []func ()
}

func newFuture() *future {
	return &future{
		result: make(chan *futureResult, 1),
		mutex: sync.Mutex{},
		state: undefinedFuture,
		callbacksCompleted: false,
	}
}

func completeRunnable(completable Completable, runnable Runnable) {
	defer completable.Close()
	result, err := runnable.Run()
	if err != nil {
		_ = completable.Fail(err)
	} else {
		_ = completable.Success(result)
	}
}

func CreateFuture(runnable Runnable) Future {
	completable := NewCompletable()

	go func() {
		completeRunnable(completable, runnable)
	}()
	return completable.Future()
}

func CreateDeferredFuture(duration time.Duration, runnable Runnable) Future {
	completable := NewCompletable()

	go func() {
		defer completable.Close()
		t := time.NewTimer(duration)
		<-t.C
		result, err := runnable.Run()
		if err != nil {
			_ = completable.Fail(err)
		} else {
			_ = completable.Success(result)
		}
	}()

	return completable.Future()
}

func CreateRetryableFuture(numRetries int, initialDelay time.Duration, runnable Runnable) Future {
	completable := NewCompletable()

	go func() {
		var err error
		var result interface{}
		defer completable.Close()
		delay := initialDelay
		for i := 0; i < numRetries; i++ {
			result, err = runnable.Run()
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

func (f *future) Get() (interface{}, error) {
	return f.getWithContext(context.Background())
}

func (f *future) GetWithTimeout(timeout time.Duration) (interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return f.getWithContext(ctx)
}

func (f *future) ThenWithRetry(numRetries int, initialDelay time.Duration, runnable PartialRunnable) Future {
	next := NewCompletable()

	go func() {
		var err error
		var result interface{}
		defer next.Close()

		delay := initialDelay
		result, err = f.Get()
		if err != nil {
			_ = next.Fail(err)
			return
		}
		err = runnable.SetArgs(result)
		if err != nil {
			_ = next.Fail(err)
			return
		}
		for i := 0; i < numRetries; i++ {
			result, err = runnable.Run()
			if err != nil {
				time.Sleep(delay)
				delay <<= 1
				continue
			} else {
				_ = next.Success(result)
				return
			}
		}
		_ = next.Fail(err)
	}()
	return next.Future()
}

func (f *future) Then(runnable PartialRunnable) Future {
	next := NewCompletable()

	go func() {
		result, err := f.Get()
		if err != nil {
			_ = next.Fail(err)
		} else {
			err = runnable.SetArgs(result)
			if err != nil {
				_ = next.Fail(err)
			} else {
				completeRunnable(next, runnable)
			}
		}
	}()
	return next.Future()
}

func (f *future) getWithContext(ctx context.Context) (interface{}, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-f.result:
		return f.finalResult.value, f.finalResult.err
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
