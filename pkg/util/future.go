package util

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
	Then(runnable PartialRunnable, options ...FutureOption) Future
	Cancel(ctx context.Context) error
	IsCancelled() bool
	IsCompleted() bool
	IsSucceeded() bool
	CallbacksCompleted() bool
	OnSuccess(func (context.Context, interface{})) Future
	OnCancel(func (context.Context)) Future
	OnFail(func (context.Context, error)) Future
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

func (result FutureResult) Context() context.Context {
	return result.ctx
}

func newFutureResult(ctx context.Context, value interface{}) *FutureResult {
	return &FutureResult{
		ctx: ctx,
		value: value,
		err: nil,
	}
}

func newCancelledFuture(ctx context.Context) *FutureResult {
	return &FutureResult{
		ctx: ctx,
		value: nil,
		err: NewCancelledError(),
	}
}

func newFailedFuture(ctx context.Context, err error) *FutureResult {
	return &FutureResult{
		ctx: ctx,
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
	callbacksCalled bool
	successes []func (context.Context, interface{})
	fails []func (context.Context, error)
	cancels []func (context.Context)
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

type retryOption struct {
	num int
	initialDelay time.Duration
}

type FutureOptions struct {
	ctx context.Context
	delay time.Duration
	retry *retryOption
	prepareFn func(ctx, prev context.Context) context.Context
}

type FutureOption func(opts* FutureOptions)

func SetContext(ctx context.Context) FutureOption {
	return func(opts *FutureOptions) {
		opts.ctx = ctx
	}
}

func WithDelay(delay time.Duration) FutureOption {
	return func(opts *FutureOptions) {
		opts.delay = delay
	}
}

func WithPrepare(prepareFn func(ctx, prev context.Context) context.Context) FutureOption {
	return func(opts *FutureOptions) {
		opts.prepareFn = prepareFn
	}
}

func WithRetry(num int, initialDelay time.Duration) FutureOption {
	return func(opts *FutureOptions) {
		opts.retry = &retryOption{
			initialDelay: initialDelay,
			num: num,
		}
	}
}

func CreateFuture(runnable Runnable, options ...FutureOption) Future {
	completable := NewCompletable()
	futureOptions := &FutureOptions{
		ctx: context.Background(),
		retry: &retryOption{
			initialDelay: 0,
			num: 0,
		},
		delay: 0,
	}

	for _, opt := range options {
		opt(futureOptions)
	}

	go func() {
		var err error
		var result interface{}
		defer completable.Close()
		defer func() {
			if r := recover(); r != nil {
				_ = completable.Fail(futureOptions.ctx, err)
			}
		}()
		if futureOptions.delay > 0 {
			t := time.NewTimer(futureOptions.delay)
			<-t.C
		}

		if futureOptions.prepareFn != nil {
			futureOptions.ctx = futureOptions.prepareFn(futureOptions.ctx, context.Background())
		}

		delay := futureOptions.retry.initialDelay
		for i := 0; i <= futureOptions.retry.num; i++ {
			result, err = runnable.Run(futureOptions.ctx)
			if err != nil && futureOptions.retry.num > 0 {
				time.Sleep(delay)
				delay <<= 1
				continue
			} else if err != nil {
				break
			} else {
				_ = completable.Success(futureOptions.ctx, result)
				return
			}
		}
		_ = completable.Fail(futureOptions.ctx, err)
	}()
	return completable.Future()
}

func (f *future) Then(runnable PartialRunnable, options ...FutureOption) Future {
	next := NewCompletable()
	futureOptions := &FutureOptions{
		ctx: context.Background(),
		retry: &retryOption{
			initialDelay: 0,
			num: 0,
		},
		delay: 0,
	}

	for _, opt := range options {
		opt(futureOptions)
	}

	go func() {
		var err error
		var thenResult interface{}
		defer next.Close()
		defer func() {
			if r := recover(); r != nil {
				_ = next.Fail(futureOptions.ctx, err)
			}
		}()

		result := f.Get()
		if result.err != nil {
			_ = next.Fail(result.ctx, result.err)
			return
		}
		err = runnable.SetInData(result.value)
		if err != nil {
			_ = next.Fail(result.ctx, result.err)
			return
		}

		if futureOptions.prepareFn != nil {
			futureOptions.ctx = futureOptions.prepareFn(futureOptions.ctx, result.ctx)
		}

		delay := futureOptions.retry.initialDelay
		for i := 0; i <= futureOptions.retry.num; i++ {
			thenResult, err = runnable.Run(futureOptions.ctx)
			if err != nil && futureOptions.retry.num > 0 {
				time.Sleep(delay)
				delay <<= 1
				continue
			} else if err != nil {
				break
			} else {
				_ = next.Success(futureOptions.ctx, thenResult)
				return
			}
		}
		_ = next.Fail(futureOptions.ctx, err)
	}()
	return next.Future()
}

func (f *future) Get() *FutureResult {
	return f.getWithContext(context.Background())
}

func (f *future) GetWithTimeout(timeout time.Duration) *FutureResult {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return f.getWithContext(ctx)
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

func (f *future) IsSucceeded() bool {
	return f.state == succeededFuture
}

func (f *future) CallbacksCompleted() bool {
	return f.callbacksCompleted
}

func (f *future) CallbacksCalled() bool {
	return f.callbacksCalled
}

func (f *future) OnSuccess(cb func (context.Context, interface{})) Future {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if f.CallbacksCalled() && f.state == succeededFuture {
		cb(f.finalResult.ctx, f.finalResult.value)
	} else if !f.CallbacksCalled() {
		f.successes = append(f.successes, cb)
	}
	return f
}

func (f *future) OnCancel(cb func (context.Context)) Future {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if f.CallbacksCalled() && f.IsCancelled() {
		cb(f.finalResult.ctx)
	} else if !f.CallbacksCalled() {
		f.cancels = append(f.cancels, cb)
	}
	return f
}

func (f *future) OnFail(cb func (context.Context, error)) Future {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if f.CallbacksCalled() && f.state == failedFuture {
		cb(f.finalResult.ctx, f.finalResult.err)
	} else if !f.CallbacksCalled() {
		f.fails = append(f.fails, cb)
	}
	return f
}

func (f *future) success(ctx context.Context, result interface{}) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if f.state != undefinedFuture {
		return NewAlreadyCompletedError("Cannot succeed a future that has already completed")
	}

	f.finalResult = newFutureResult(ctx, result)
	f.state = succeededFuture
	f.doCallbacks(ctx)
	f.result <- f.finalResult
	return nil
}

func (f *future) fail(ctx context.Context, err error) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if f.state != undefinedFuture {
		return NewAlreadyCompletedError("Cannot fail a future that has already completed")
	}

	f.finalResult = newFailedFuture(ctx, err)
	f.state = failedFuture
	f.doCallbacks(ctx)
	f.result <- f.finalResult
	return nil
}

func (f *future) Cancel(ctx context.Context) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if f.state != undefinedFuture {
		return NewAlreadyCompletedError("Cannot cancel future that has already completed")
	}

	f.finalResult = newCancelledFuture(ctx)
	f.state = canceledFuture
	f.doCallbacks(ctx)
	f.result <- f.finalResult
	return nil
}

func (f *future) close() {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	close(f.result)
}

func (f *future) doCallbacks(ctx context.Context) {
	wg := &sync.WaitGroup{}

	numCallbacks := 0
	switch f.state {
	case succeededFuture:
		numCallbacks = len(f.successes)
	case failedFuture:
		numCallbacks = len(f.fails)
	case canceledFuture:
		numCallbacks = len(f.cancels)
	}

	wg.Add(numCallbacks)

	defer func() {
		f.callbacksCalled = true
	}()

	go func() {
		wg.Wait()
		f.callbacksCompleted = true
	}()

	switch f.state {
	case succeededFuture:
		for _, cb := range f.successes {
			go func(callback func(context.Context, interface{})) {
				callback(ctx, f.finalResult.value)
				wg.Done()
			}(cb)
		}
	case failedFuture:
		for _, cb := range f.fails {
			go func(callback func(context.Context, error)) {
				callback(ctx, f.finalResult.err)
				wg.Done()
			}(cb)
		}
	case canceledFuture:
		for _, cb := range f.cancels {
			go func(callback func(context.Context)) {
				callback(ctx)
				wg.Done()
			}(cb)
		}
	default:
		return
	}
}
