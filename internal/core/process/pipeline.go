package process

import (
	"context"
	"fmt"
	api "github.com/kmgreen2/agglo/generated/proto"
	"github.com/kmgreen2/agglo/pkg/util"
	"github.com/kmgreen2/agglo/pkg/observability"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"reflect"
	"time"
)

type PipelineProcess interface {
	Process(ctx context.Context, in map[string]interface{}) (map[string]interface{}, error)
	Name() string
}

func getType(myvar interface{}) string {
	if t := reflect.TypeOf(myvar); t.Kind() == reflect.Ptr {
		return "*" + t.Elem().Name()
	} else {
		return t.Name()
	}
}

func PipelineProcessError(p PipelineProcess, err error, msg string) error {
	return errors.Wrap(err, fmt.Sprintf("%s(%s) - %s", p.Name(), getType(p), msg))
}

type RunnableStartProcess struct {
	process PipelineProcess
	in      map[string]interface{}
}

func (runnable *RunnableStartProcess) Run(ctx context.Context) (interface{}, error) {
	return runnable.process.Process(ctx, runnable.in)
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
	return runnable.process.Process(ctx, runnable.in)
}

func (runnable *RunnablePartialProcess) SetInData(inData interface{}) error {
	switch rv := inData.(type) {
	case map[string]interface{}:
		runnable.in = rv
	default:
		msg := fmt.Sprintf("RunnablePartialProcess should have an map[string]interface{} arg, found %v",
			reflect.TypeOf(rv))
		return util.NewInvalidError(msg)
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
	return nil, runnable.finalizer.Finalize(ctx, runnable.in)
}

func (runnable *RunnablePartialFinalizer) SetInData(inData interface{}) error {
	switch rv := inData.(type) {
	case map[string]interface{}:
		runnable.in = rv
	default:
		msg := fmt.Sprintf("RunnablePartialFinalizer should have an map[string]interface{} arg, found %v",
			reflect.TypeOf(rv))
		return util.NewInvalidError(msg)
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

type RunOptions struct {
	logger *zap.Logger
}

type RunOptionBuilder func(runOptions *RunOptions)

func WithLogger(logger *zap.Logger) RunOptionBuilder {
	return func (options *RunOptions) {
		options.logger = logger
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
	name string
	processes []PipelineProcess
	processOptions []*Options
	processLifecycles []*Lifecycle
	checkPointer *CheckPointer
	emitter *observability.Emitter
	enableTracing bool
	enableMetrics bool
}

type Pipelines struct {
	pipelines []*Pipeline
	shutdownFns []func() error
}

func NewPipelines(pipelines []*Pipeline, shutdownFns []func() error) *Pipelines {
	return &Pipelines{
		pipelines,
		shutdownFns,
	}
}

func (pipelines Pipelines) Underlying() []*Pipeline {
	return pipelines.pipelines
}

func (pipelines Pipelines) Shutdown() error {
	errStr := ""
	for _, fn := range pipelines.shutdownFns {
		err := fn()
		if err != nil {
			errStr += err.Error() + "\n"
		}
	}
	if len(errStr) > 0 {
		return util.NewInternalError(errStr)
	}
	return nil
}

func (pipelines Pipelines) GetCheckpoints(messageID string) ([]map[string]interface{}, error) {
	var err error
	var errStr string
	numNotFound := 0
	checkpoints := make([]map[string]interface{}, len(pipelines.pipelines))
	for i, pipeline := range pipelines.pipelines {
		checkpoints[i], err = pipeline.checkPointer.GetCheckpoint(context.Background(), pipeline.name, messageID)
		if err != nil && !errors.Is(err, &util.NotFoundError{}){
			if !errors.Is(err, &util.NotFoundError{}){
				errStr += err.Error() + "\n"
			}
			numNotFound++
		}
	}
	if numNotFound == len(pipelines.pipelines) {
		return checkpoints, util.NewInternalError("no checkpoints found")
	}
	if len(errStr) > 0 {
		return checkpoints, util.NewInternalError(fmt.Sprintf("found errors getting checkpoints: %s", errStr))
	}
	return checkpoints, nil
}

func (pipeline *Pipeline) createFutureHelper(ctx context.Context, pipelineIndex int,
	in map[string]interface{}) util.Future {
	var future util.Future

	process := pipeline.processes[pipelineIndex]
	processOptions := pipeline.processOptions[pipelineIndex]

	var prepareFn func(ctx, prev context.Context) context.Context
	if processOptions.processLifecycle != nil {
		prepareFn = processOptions.processLifecycle.Prepare
	}

	if processOptions.retryStrategy != nil {
		future = util.CreateFuture(NewRunnableStartProcess(process, in),
			util.WithRetry(int(processOptions.retryStrategy.NumRetries),
				time.Duration(processOptions.retryStrategy.InitialBackOffMs) * time.Millisecond),
			util.SetContext(ctx),
			util.WithPrepare(prepareFn))
	} else {
		future = util.CreateFuture(NewRunnableStartProcess(process, in), util.SetContext(ctx),
			util.WithPrepare(prepareFn))
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

func (pipeline *Pipeline) thenFutureHelper(ctx context.Context, pipelineIndex int,
	inFuture util.Future) util.Future {

	var future util.Future

	process := pipeline.processes[pipelineIndex]
	processOptions := pipeline.processOptions[pipelineIndex]

	var prepareFn func(ctx, prev context.Context) context.Context
	if processOptions.processLifecycle != nil {
		prepareFn = processOptions.processLifecycle.Prepare
	}

	if processOptions.retryStrategy != nil {
		future = inFuture.Then(NewRunnablePartialProcess(process),
			util.WithRetry(int(processOptions.retryStrategy.NumRetries),
				time.Duration(processOptions.retryStrategy.InitialBackOffMs) * time.Millisecond),
			util.SetContext(ctx), util.WithPrepare(prepareFn))
	} else {
		future = inFuture.Then(NewRunnablePartialProcess(process), util.SetContext(ctx),
			util.WithPrepare(prepareFn))
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

func (pipeline Pipeline) RunSync(in map[string]interface{}, runOptions ...RunOptionBuilder) (map[string]interface{},
error) {

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
		return nil, util.NewInvalidError(msg)
	}
}

func onFailLogHelper(logger *zap.Logger, msg string) func(ctx context.Context, err error) {
	return func(ctx context.Context, err error) {
		// Do not log warnings
		// ToDo(KMG): Is there a better way to determine if an error is jsut a warning?
		// We currently use errors to halt pipeline progress.  This is mostly needed for
		// continuations or processes that are fail-able.  In those cases, we do not need
		// to log errors.
		if !util.IsWarning(err) {
			logger.Error(fmt.Sprintf("%s: %+v", msg, err))
		}
	}
}

func (pipeline Pipeline) RunAsync(in map[string]interface{}, options ...RunOptionBuilder) util.Future {
	var startIndex int64 = 0
	var err error
	var span trace.Span
	var ctx context.Context
	var startTime time.Time
	var inMap map[string]interface{} = in

	runOptions := &RunOptions{}
	for _, opt := range options {
		opt(runOptions)
	}

	ctx = context.Background()

	// If there are no processes in this pipeline, do nothing and succeed
	if len(pipeline.processes) == 0 {
		completable := util.NewCompletable()
		_ = completable.Success(ctx, in)
		return completable.Future()
	}

	if pipeline.enableTracing {
		ctx, span = pipeline.emitter.CreateSpan(ctx, "pipeline")
	}

	if pipeline.enableMetrics {
		startTime = time.Now()
	}

	// If checkpointing is enabled, try to fetch the checkpoint
	if pipeline.checkPointer != nil {
		inMap, startIndex, err = pipeline.checkPointer.GetCheckpointWithIndexFromMap(ctx, in)
		if err != nil {
			// This means the checkpoint does not exist, so process as usual
			if errors.Is(err, &util.NotFoundError{}) {
				inMap = in
				startIndex = 0
			} else {
				completable := util.NewCompletable()
				_ = completable.Fail(context.Background(), err)
				return completable.Future()
			}
		}
	}

	// Chain together the processes of the pipeline
	f := pipeline.createFutureHelper(ctx, int(startIndex), inMap)
	if runOptions.logger != nil {
		f.OnFail(func(ctx context.Context, err error) {
			msg := fmt.Sprintf("Process index: 0")
			onFailLogHelper(runOptions.logger, msg)(ctx, err)
		})
	}
	for i := int(startIndex + 1); i < len(pipeline.processes); i++ {
		f = pipeline.thenFutureHelper(ctx, i, f)
		if runOptions.logger != nil {
			f.OnFail(func(processIndex int) func(ctx context.Context, err error) {
				msg := fmt.Sprintf("Process index: %d", processIndex)
				return onFailLogHelper(runOptions.logger, msg)
			}(i))
		}
		// If this pipeline has checkpointing enabled, then checkpoint after each process
		if pipeline.checkPointer != nil {
			f = f.Then(NewRunnablePartialProcess(pipeline.checkPointer), util.SetContext(ctx),
			// Must pass the span context from the previous process; otherwise, the checkpoint processes
			// will break the chain
			util.WithPrepare(func(ctx, prev context.Context) context.Context {
				if pipeline.enableTracing {
					ctx, span = pipeline.emitter.CreateSpan(ctx, "checkpoint")
				}
				return ctx
			})).OnSuccess(func(ctx context.Context, x interface{}) {
				span := trace.SpanFromContext(ctx)
				if span != nil {
					span.End()
				}
			}).OnFail(func(ctx context.Context, err error) {
				span := trace.SpanFromContext(ctx)
				if span != nil {
					span.RecordError(err)
					span.End()
				}
			})
		}
	}

	// If checkpointing is enabled for this pipeline, set the last process as the checkpoint finalizer
	if pipeline.checkPointer != nil {
		// Call checkPointer.Finalize() to remove the checkpoint
		f = f.Then(NewRunnablePartialFinalizer(pipeline.checkPointer), util.SetContext(ctx),
			util.WithPrepare(func(ctx, prev context.Context) context.Context {
				if pipeline.enableTracing {
					ctx, span = pipeline.emitter.CreateSpan(ctx, "checkpoint-finalize")
				}
				return ctx
			})).OnSuccess(func(ctx context.Context, x interface{}) {
			span := trace.SpanFromContext(ctx)
			if span != nil {
				span.End()
			}
		}).OnFail(func(ctx context.Context, err error) {
			span := trace.SpanFromContext(ctx)
			if span != nil {
				span.RecordError(err)
				span.End()
			}
		})
	}

	f.OnSuccess(func(ctx context.Context, x interface{}) {
		if pipeline.enableMetrics {
			pipeline.emitter.RecordInt64(pipeline.name + ".latency", time.Now().Sub(startTime).Milliseconds())
			pipeline.emitter.AddInt64(pipeline.name + ".fuccess", 1)
		}
		if pipeline.enableTracing {
			span.End()
		}
	})
	f.OnFail(func(ctx context.Context, err error) {
		if pipeline.enableMetrics {
			pipeline.emitter.AddInt64(pipeline.name + ".failure", 1)
		}
		if pipeline.enableTracing {
			span.RecordError(err)
			span.End()
		}
	})

	return f
}

type PipelineBuilder struct {
	pipeline *Pipeline
}

func NewPipelineBuilder() *PipelineBuilder {
	return &PipelineBuilder{
		&Pipeline{
			emitter: observability.NewEmitter("agglo/pipeline"),
		},
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

func (builder *PipelineBuilder) EnableTracing() *PipelineBuilder {
	builder.pipeline.enableTracing = true
	return builder
}

func (builder *PipelineBuilder) EnableMetrics() *PipelineBuilder {
	builder.pipeline.emitter.AddMetric(builder.pipeline.name + ".latency", observability.Int64Recorder)
	builder.pipeline.emitter.AddMetric(builder.pipeline.name + ".success", observability.Int64Counter)
	builder.pipeline.emitter.AddMetric(builder.pipeline.name + ".failure", observability.Int64Counter)
	builder.pipeline.enableMetrics = true
	return builder
}

func (builder *PipelineBuilder) SetName(name string) *PipelineBuilder {
	builder.pipeline.name = name
	return builder
}

func (builder *PipelineBuilder) Get() *Pipeline {
	return builder.pipeline
}
