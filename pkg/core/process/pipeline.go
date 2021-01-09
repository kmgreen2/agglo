package process

import (
	"context"
	"fmt"
	api "github.com/kmgreen2/agglo/generated/proto"
	"github.com/kmgreen2/agglo/pkg/common"
	"github.com/pkg/errors"
	"reflect"
	"time"
)

type PipelineProcess interface {
	Process(in map[string]interface{}) (map[string]interface{}, error)
}

type RunnableStartProcess struct {
	process PipelineProcess
	in      map[string]interface{}
}

func (runnable *RunnableStartProcess) Run(ctx context.Context) (interface{}, error) {
	return runnable.process.Process(runnable.in)
}

func NewRunnableStartProcess(process PipelineProcess, in map[string]interface{}) *RunnableStartProcess {
	return &RunnableStartProcess{
		process: process,
		in: in,
	}
}

type RunnablePartialProcess struct {
	process PipelineProcess
	in      map[string]interface{}
}

func (runnable *RunnablePartialProcess) Run(ctx context.Context) (interface{}, error) {
	return runnable.process.Process(runnable.in)
}

func (runnable *RunnablePartialProcess) SetInData(inData interface{}) error {
	switch rv := inData.(type) {
	case map[string]interface{}:
		runnable.in = rv
	default:
		msg := fmt.Sprintf("RunnablePartialProcess should have an map[string]interface{} arg, found %v",
			reflect.TypeOf(rv))
		return common.NewInvalidError(msg)
	}
	return nil
}

func NewRunnablePartialProcess(process PipelineProcess) *RunnablePartialProcess {
	return &RunnablePartialProcess{
		process: process,
	}
}

type RunnablePartialFinalizer struct {
	finalizer Finalizer
	in      map[string]interface{}
}

func (runnable *RunnablePartialFinalizer) Run(ctx context.Context) (interface{}, error) {
	return nil, runnable.finalizer.Finalize(runnable.in)
}

func (runnable *RunnablePartialFinalizer) SetInData(inData interface{}) error {
	switch rv := inData.(type) {
	case map[string]interface{}:
		runnable.in = rv
	default:
		msg := fmt.Sprintf("RunnablePartialFinalizer should have an map[string]interface{} arg, found %v",
			reflect.TypeOf(rv))
		return common.NewInvalidError(msg)
	}
	return nil
}

func NewRunnablePartialFinalizer(finalizer Finalizer) *RunnablePartialFinalizer {
	return &RunnablePartialFinalizer{
		finalizer: finalizer,
	}
}

type Lifecycle struct {
	prepareFns []func(ctx, prev context.Context) context.Context
	successFns []func(ctx context.Context)
	failFns    []func(ctx context.Context, err error)
}

type LifecycleBuilder struct {
	lifecycle *Lifecycle
}

func NewLifecycleBuilder() *LifecycleBuilder {
	return &LifecycleBuilder{
		&Lifecycle{},
	}
}

func (lifecycleBuilder *LifecycleBuilder) AppendPrepareFn(
	fn func(ctx, prev context.Context) context.Context) *LifecycleBuilder {

	lifecycleBuilder.lifecycle.prepareFns = append(lifecycleBuilder.lifecycle.prepareFns, fn)
	return lifecycleBuilder
}

func (lifecycleBuilder *LifecycleBuilder) AppendSuccessFn(fn func(ctx context.Context)) *LifecycleBuilder {
	lifecycleBuilder.lifecycle.successFns = append(lifecycleBuilder.lifecycle.successFns, fn)
	return lifecycleBuilder
}

func (lifecycleBuilder *LifecycleBuilder) AppendFailFn(fn func(ctx context.Context, err error)) *LifecycleBuilder {
	lifecycleBuilder.lifecycle.failFns = append(lifecycleBuilder.lifecycle.failFns, fn)
	return lifecycleBuilder
}

func (LifecycleBuilder *LifecycleBuilder) Build() *Lifecycle {
	return LifecycleBuilder.lifecycle
}

func (lifecycle *Lifecycle) Prepare(ctx, prev context.Context) context.Context {
	for _, prepareFn := range lifecycle.prepareFns {
		ctx = prepareFn(ctx, prev)
	}
	return ctx
}

func (lifecycle *Lifecycle) Success(ctx context.Context) {
	for _, successFn := range lifecycle.successFns {
		successFn(ctx)
	}
}

func (lifecycle *Lifecycle) Failure(ctx context.Context, err error) {
	for _, failureFn := range lifecycle.failFns {
		failureFn(ctx, err)
	}
}

type OptionBuilder func(process *Options)

func WithRetry(strategy *api.RetryStrategy) OptionBuilder {
	return func(options *Options) {
		options.retryStrategy = strategy
	}
}

func WithLifecycle(processLifecycle *Lifecycle) OptionBuilder {
	return func(options *Options) {
		options.processLifecycle = processLifecycle
	}
}

type Options struct {
	retryStrategy *api.RetryStrategy
	processLifecycle *Lifecycle
}

type Pipeline struct {
	processes []PipelineProcess
	processOptions []*Options
	processLifecycles []*Lifecycle
	checkPointer *CheckPointer
}

func (pipeline *Pipeline) createFutureHelper(pipelineIndex int, in map[string]interface{}) common.Future {
	var future common.Future

	process := pipeline.processes[pipelineIndex]
	processOptions := pipeline.processOptions[pipelineIndex]

	var prepareFn func(ctx, prev context.Context) context.Context
	if processOptions.processLifecycle != nil {
		prepareFn = processOptions.processLifecycle.Prepare
	}

	if processOptions.retryStrategy != nil {
		future = common.CreateFuture(NewRunnableStartProcess(process, in),
			common.WithRetry(int(processOptions.retryStrategy.NumRetries),
				time.Duration(processOptions.retryStrategy.InitialBackOffMs) * time.Millisecond),
			common.SetContext(context.Background()),
			common.WithPrepare(prepareFn))
	} else {
		future = common.CreateFuture(NewRunnableStartProcess(process, in), common.SetContext(context.Background()),
			common.WithPrepare(prepareFn))
	}

	if processOptions.processLifecycle != nil {
		future.OnSuccess(func(ctx context.Context, x interface{}) {
			processOptions.processLifecycle.Success(ctx)
		})

		future.OnFail(func(ctx context.Context, err error) {
			processOptions.processLifecycle.Failure(ctx, err)
		})
	}

	return future
}

func (pipeline *Pipeline) thenFutureHelper(pipelineIndex int, inFuture common.Future) common.Future {
	var future common.Future

	process := pipeline.processes[pipelineIndex]
	processOptions := pipeline.processOptions[pipelineIndex]

	var prepareFn func(ctx, prev context.Context) context.Context
	if processOptions.processLifecycle != nil {
		prepareFn = processOptions.processLifecycle.Prepare
	}

	if processOptions.retryStrategy != nil {
		future = inFuture.Then(NewRunnablePartialProcess(process),
			common.WithRetry(int(processOptions.retryStrategy.NumRetries),
				time.Duration(processOptions.retryStrategy.InitialBackOffMs) * time.Millisecond),
			common.SetContext(context.Background()), common.WithPrepare(prepareFn))
	} else {
		future = inFuture.Then(NewRunnablePartialProcess(process), common.SetContext(context.Background()),
			common.WithPrepare(prepareFn))
	}

	if processOptions.processLifecycle != nil {
		future.OnSuccess(func(ctx context.Context, x interface{}) {
			processOptions.processLifecycle.Success(ctx)
		})

		future.OnFail(func(ctx context.Context, err error) {
			processOptions.processLifecycle.Failure(ctx, err)
		})
	}
	return future
}

func (pipeline Pipeline) RunSync(in map[string]interface{}) (map[string]interface{}, error) {
	f := pipeline.RunAsync(in)
	result := f.Get()

	if result.Error() != nil {
		return nil, result.Error()
	}

	switch rv := result.Value().(type) {
	case map[string]interface{}:
		return rv, nil
	default:
		msg := fmt.Sprintf("invalid return type in process (expected map[string]interface{}: %v",
			reflect.TypeOf(rv))
		return nil, common.NewInvalidError(msg)
	}
}

func (pipeline Pipeline) RunAsync(in map[string]interface{}) common.Future {
	var startIndex int64 = 0
	var err error
	var inMap map[string]interface{} = in

	// If there are no processes in this pipeline, do nothing and succeed
	if len(pipeline.processes) == 0 {
		completable := common.NewCompletable()
		_ = completable.Success(context.Background(), in)
		return completable.Future()
	}

	// If checkpointing is enabled, try to fetch the checkpoint
	if pipeline.checkPointer != nil {
		inMap, startIndex, err = pipeline.checkPointer.GetCheckpointWithIndex(in)
		if err != nil {
			// This means the checkpoint does not exist, so process as usual
			if errors.Is(err, &common.NotFoundError{}) {
				inMap = in
				startIndex = 0
			} else {
				completable := common.NewCompletable()
				_ = completable.Fail(context.Background(), err)
				return completable.Future()
			}
		}
	}

	// Chain together the processes of the pipeline
	f := pipeline.createFutureHelper(int(startIndex), inMap)
	for i := int(startIndex + 1); i < len(pipeline.processes); i++ {
		f = pipeline.thenFutureHelper(i, f)
		// If this pipeline has checkpointing enabled, then checkpoint after each process
		if pipeline.checkPointer != nil {
			f = f.Then(NewRunnablePartialProcess(pipeline.checkPointer),
			// Must pass the span context from the previous process; otherwise, the checkpoint processes
			// will break the chain
			common.WithPrepare(func(ctx, prev context.Context) context.Context {
				if spanCtx := common.ExtractSpanContext(prev); spanCtx != common.EmptySpanContext {
					return common.InjectSpanContext(ctx, spanCtx)
				}
				return ctx
			}))
		}
	}

	// If checkpointing is enabled for this pipeline, set the last process as the checkpoint finalizer
	if pipeline.checkPointer != nil {
		// Call checkPointer.Finalize() to remove the checkpoint
		f = f.Then(NewRunnablePartialFinalizer(pipeline.checkPointer))
	}

	return f
}

type PipelineBuilder struct {
	pipeline *Pipeline
}

func NewPipelineBuilder() *PipelineBuilder {
	return &PipelineBuilder{
		&Pipeline{},
	}
}

func (builder *PipelineBuilder) Add(process PipelineProcess, options ...OptionBuilder) *PipelineBuilder {
	builder.pipeline.processes = append(builder.pipeline.processes, process)

	option := &Options{}
	for _, opt := range options {
		opt(option)
	}

	builder.pipeline.processOptions = append(builder.pipeline.processOptions, option)
	return builder
}

func (builder *PipelineBuilder) Checkpoint(checkPointer *CheckPointer) *PipelineBuilder {
	builder.pipeline.checkPointer = checkPointer
	return builder
}

func (builder *PipelineBuilder) Get() *Pipeline {
	return builder.pipeline
}