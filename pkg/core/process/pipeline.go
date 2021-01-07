package process

import (
	"fmt"
	"github.com/kmgreen2/agglo/pkg/common"
	"github.com/pkg/errors"
	"reflect"
)

type PipelineProcess interface {
	Process(in map[string]interface{}) (map[string]interface{}, error)
}

type RunnableStartProcess struct {
	process PipelineProcess
	in      map[string]interface{}
}

func (runnable *RunnableStartProcess) Run() (interface{}, error) {
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

func (runnable *RunnablePartialProcess) Run() (interface{}, error) {
	return runnable.process.Process(runnable.in)
}

func (runnable *RunnablePartialProcess) SetArgs(args ...interface{}) error {
	if len(args) > 1 {
		msg := fmt.Sprintf("RunnablePartialProcess should have 1 int arg, found %d args", len(args))
		return common.NewInvalidError(msg)
	}
	switch rv := args[0].(type) {
	case map[string]interface{}:
		runnable.in = rv
	default:
		msg := fmt.Sprintf("RunnablePartialProcess should have an map[string]interface{} arg, found %v",
			reflect.TypeOf(args[0]))
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

func (runnable *RunnablePartialFinalizer) Run() (interface{}, error) {
	return nil, runnable.finalizer.Finalize(runnable.in)
}

func (runnable *RunnablePartialFinalizer) SetArgs(args ...interface{}) error {
	if len(args) > 1 {
		msg := fmt.Sprintf("RunnablePartialFinalizer should have 1 int arg, found %d args", len(args))
		return common.NewInvalidError(msg)
	}
	switch rv := args[0].(type) {
	case map[string]interface{}:
		runnable.in = rv
	default:
		msg := fmt.Sprintf("RunnablePartialFinalizer should have an map[string]interface{} arg, found %v",
			reflect.TypeOf(args[0]))
		return common.NewInvalidError(msg)
	}
	return nil
}

func NewRunnablePartialFinalizer(finalizer Finalizer) *RunnablePartialFinalizer {
	return &RunnablePartialFinalizer{
		finalizer: finalizer,
	}
}

type Pipeline struct {
	processes []PipelineProcess
	checkPointer *CheckPointer
}

func (pipeline Pipeline) RunSync(in map[string]interface{}) (map[string]interface{}, error) {
	f := pipeline.RunAsync(in)
	result, err := f.Get()

	if err != nil {
		return nil, err
	}

	switch rv := result.(type) {
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
		_ = completable.Success(in)
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
				_ = completable.Fail(err)
				return completable.Future()
			}
		}
	}

	// Chain together the processes of the pipeline
	f := common.CreateFuture(NewRunnableStartProcess(pipeline.processes[startIndex], inMap))
	for i := int(startIndex + 1); i < len(pipeline.processes); i++ {
		f = f.Then(NewRunnablePartialProcess(pipeline.processes[i]))
		// If this pipeline has checkpointing enabled, then checkpoint after each process
		if pipeline.checkPointer != nil {
			f = f.Then(NewRunnablePartialProcess(pipeline.checkPointer))
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

func (builder *PipelineBuilder) Add(process PipelineProcess) *PipelineBuilder {
	builder.pipeline.processes = append(builder.pipeline.processes, process)
	return builder
}

func (builder *PipelineBuilder) Checkpoint(checkPointer *CheckPointer) *PipelineBuilder {
	builder.pipeline.checkPointer = checkPointer
	return builder
}

func (builder *PipelineBuilder) Get() *Pipeline {
	return builder.pipeline
}
