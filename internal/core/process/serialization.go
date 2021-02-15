package process

import (
	"bytes"
	"context"
	"fmt"
	"github.com/golang/protobuf/jsonpb"
	gUuid "github.com/google/uuid"
	"github.com/kmgreen2/agglo/internal/common"
	"github.com/kmgreen2/agglo/pkg/state"
	"github.com/kmgreen2/agglo/pkg/util"
	"github.com/kmgreen2/agglo/internal/core"
	"github.com/kmgreen2/agglo/pkg/kvs"
	"github.com/kmgreen2/agglo/generated/proto"
	"github.com/kmgreen2/agglo/pkg/observability"
	"github.com/kmgreen2/agglo/pkg/streaming"
	"go.opentelemetry.io/otel/trace"
	"net/http"
	"reflect"
	"time"
)

func protoExistsOpToInternal(in api.ExistsOperator) core.ExistsOperator {
	switch in {
	case api.ExistsOperator_Exists:
		return core.Exists
	case api.ExistsOperator_NotExists:
		return core.NotExists
	default:
		// Should never reach this, but need to return something to satisfy the compiler
		return -1
	}
}

func protoAggregationTypeToInternal(in api.AggregationType) core.AggregationType {
	switch in {
	case api.AggregationType_AggAvg:
		return core.AggAvg
	case api.AggregationType_AggCount:
		return core.AggCount
	case api.AggregationType_AggMin:
		return core.AggMin
	case api.AggregationType_AggMax:
		return core.AggMax
	case api.AggregationType_AggDiscreteHistogram:
		return core.AggDiscreteHistogram
	case api.AggregationType_AggSum:
		return core.AggSum
	default:
		return -1
	}
}

func buildExpression(expression *api.Expression) (core.Expression, error) {
	var err error
	var lhs, rhs interface{}
	switch e := expression.Expression.(type) {
	case *api.Expression_Boolean:
	case *api.Expression_Comparator:
		switch lhsOperand := e.Comparator.Lhs.Operand.(type) {
		case *api.Operand_Literal:
			lhs = lhsOperand.Literal
		case *api.Operand_Numeric:
			lhs = lhsOperand.Numeric
		case *api.Operand_Variable:
			lhs = core.Variable(lhsOperand.Variable.Name)
		case *api.Operand_Expression:
			lhs, err = buildExpression(lhsOperand.Expression)
			if err != nil {
				return nil, err
			}
		}
		switch rhsOperand := e.Comparator.Rhs.Operand.(type) {
		case *api.Operand_Literal:
			rhs = rhsOperand.Literal
		case *api.Operand_Numeric:
			rhs = rhsOperand.Numeric
		case *api.Operand_Variable:
			rhs = core.Variable(rhsOperand.Variable.Name)
		case *api.Operand_Expression:
			rhs, err = buildExpression(rhsOperand.Expression)
			if err != nil {
				return nil, err
			}
		}
		switch e.Comparator.Op {
		case api.ComparatorOperator_Equal:
			return core.NewComparatorExpression(lhs, rhs, core.Equal), nil
		case api.ComparatorOperator_NotEqual:
			return core.NewComparatorExpression(lhs, rhs, core.NotEqual), nil
		case api.ComparatorOperator_GreaterThan:
			return core.NewComparatorExpression(lhs, rhs, core.GreaterThan), nil
		case api.ComparatorOperator_GreaterThanOrEqual:
			return core.NewComparatorExpression(lhs, rhs, core.GreaterThanOrEqual), nil
		case api.ComparatorOperator_LessThan:
			return core.NewComparatorExpression(lhs, rhs, core.LessThan), nil
		case api.ComparatorOperator_LessThanOrEqual:
			return core.NewComparatorExpression(lhs, rhs, core.LessThanOrEqual), nil
		case api.ComparatorOperator_RegexMatch:
			return core.NewComparatorExpression(lhs, rhs, core.RegexMatch), nil
		case api.ComparatorOperator_RegexNotMatch:
			return core.NewComparatorExpression(lhs, rhs, core.RegexNotMatch), nil
		}

	case *api.Expression_Logical:
		switch lhsOperand := e.Logical.Lhs.Operand.(type) {
		case *api.Operand_Expression:
			lhs, err = buildExpression(lhsOperand.Expression)
			if err != nil {
				return nil, err
			}
		default:
			msg := fmt.Sprintf("logical expression operands *must* be expressions, not literal, " +
				"variable or numeric.  Got %v", reflect.TypeOf(lhsOperand))
			return nil, util.NewInvalidError(msg)
		}
		switch rhsOperand := e.Logical.Rhs.Operand.(type) {
		case *api.Operand_Expression:
			rhs, err = buildExpression(rhsOperand.Expression)
			if err != nil {
				return nil, err
			}
		default:
			msg := fmt.Sprintf("logical expression operands *must* be expressions, not literal, " +
				"variable or numeric.  Got %v", reflect.TypeOf(rhsOperand))
			return nil, util.NewInvalidError(msg)
		}

		// ToDo(KMG): Could use type assertions for lhs and rhs, but skipping it because the assertions above
		// should guarantee that lhs and rhs are Expressions
		switch e.Logical.Op {
		case api.LogicalOperator_LogicalAnd:
			return core.NewLogicalExpression(lhs.(core.Expression), rhs.(core.Expression), core.LogicalAnd), nil
		case api.LogicalOperator_LogicalOr:
			return core.NewLogicalExpression(lhs.(core.Expression), rhs.(core.Expression), core.LogicalAnd), nil
		}
	case *api.Expression_Binary:
		switch lhsOperand := e.Binary.Lhs.Operand.(type) {
		case *api.Operand_Numeric:
			lhs = lhsOperand.Numeric
		case *api.Operand_Variable:
			lhs = core.Variable(lhsOperand.Variable.Name)
		default:
			msg := fmt.Sprintf("binary expression operands *must* be numeric or variable, not literal, " +
				"or expressions.  Got %v", reflect.TypeOf(lhsOperand))
			return nil, util.NewInvalidError(msg)
		}
		switch rhsOperand := e.Binary.Rhs.Operand.(type) {
		case *api.Operand_Numeric:
			rhs = rhsOperand.Numeric
		case *api.Operand_Variable:
			rhs = core.Variable(rhsOperand.Variable.Name)
		default:
			msg := fmt.Sprintf("binary expression operands *must* be numeric or variable, not literal, " +
				"or expressions.  Got %v", reflect.TypeOf(rhsOperand))
			return nil, util.NewInvalidError(msg)
		}

		switch e.Binary.Op {
		case api.BinaryOperator_Addition:
			return core.NewBinaryExpression(lhs, rhs, core.Addition), nil
		case api.BinaryOperator_Subtract:
			return core.NewBinaryExpression(lhs, rhs, core.Subtract), nil
		case api.BinaryOperator_Multiply:
			return core.NewBinaryExpression(lhs, rhs, core.Multiply), nil
		case api.BinaryOperator_Divide:
			return core.NewBinaryExpression(lhs, rhs, core.Divide), nil
		case api.BinaryOperator_Power:
			return core.NewBinaryExpression(lhs, rhs, core.Power), nil
		case api.BinaryOperator_Modulus:
			return core.NewBinaryExpression(lhs, rhs, core.Modulus), nil
		case api.BinaryOperator_RightShift:
			return core.NewBinaryExpression(lhs, rhs, core.RightShift), nil
		case api.BinaryOperator_LeftShift:
			return core.NewBinaryExpression(lhs, rhs, core.LeftShift), nil
		case api.BinaryOperator_Or:
			return core.NewBinaryExpression(lhs, rhs, core.Or), nil
		case api.BinaryOperator_And:
			return core.NewBinaryExpression(lhs, rhs, core.And), nil
		case api.BinaryOperator_Xor:
			return core.NewBinaryExpression(lhs, rhs, core.Xor), nil
		}

	case *api.Expression_Unary:
		switch rhsOperand := e.Unary.Rhs.Operand.(type) {
		case *api.Operand_Numeric:
			rhs = rhsOperand.Numeric
		case *api.Operand_Variable:
			rhs = core.Variable(rhsOperand.Variable.Name)
		default:
			msg := fmt.Sprintf("unary expression operands *must* be numeric or variable, not literal, " +
				"or expressions.  Got %v", reflect.TypeOf(rhsOperand))
			return nil, util.NewInvalidError(msg)
		}
		switch e.Unary.Op {
		case api.UnaryOperator_Negation:
			return core.NewUnaryExpression(rhs, core.Negation), nil
		case api.UnaryOperator_Inversion:
			return core.NewUnaryExpression(rhs, core.Inversion), nil
		case api.UnaryOperator_LogicalNot:
			return core.NewUnaryExpression(rhs, core.Not), nil
		}
	}
	return nil, nil
}

func buildCondition(condition *api.Condition) (*core.Condition, error) {
	// If no condition is specified, then assume True
	if condition == nil || condition.Condition == nil {
		return core.TrueCondition, nil
	}
	switch c := condition.Condition.(type) {
	case *api.Condition_Expression:
		expression, err := buildExpression(c.Expression)
		if err != nil {
			return nil, err
		}
		return core.NewCondition(expression)
	case *api.Condition_Exists:
		builder := core.NewExistsExpressionBuilder()
		for _, op := range c.Exists.Ops {
			builder.Add(op.Key, protoExistsOpToInternal(op.Op))
		}
		return core.NewCondition(builder.Get())
	default:
		msg := fmt.Sprintf("invalid condition type: %v", reflect.TypeOf(c))
		return nil, util.NewInvalidError(msg)
	}
}

func buildTransformer(transformerSpec *api.Transformer) (*Transformer, error) {
	transformerSpecs := transformerSpec.Specs
	transformer := NewTransformer(transformerSpec.Name, nil, ".", ".",
		transformerSpec.ForwardInputFields)
	for _, spec := range transformerSpecs {
		condition, err := buildCondition(spec.Transformation.Condition)
		if err != nil {
			return nil, err
		}
		builder := core.NewTransformationBuilder()
		builder.AddCondition(condition)
		switch spec.Transformation.TransformationType {
		case api.TransformationType_TransformCopy:
			builder.AddFieldTransformation(&core.CopyTransformation{})
		case api.TransformationType_TransformCount:
			builder.AddFieldTransformation(core.LeftFoldCountAll)
		case api.TransformationType_TransformSum:
			builder.AddFieldTransformation(&core.SumTransformation{})
		case api.TransformationType_TransformMapAdd:
			switch transformArgs := spec.Transformation.TransformArgs.(type) {
			case *api.Transformation_MapAddArgs:
				builder.AddFieldTransformation(core.MapAddConstant(transformArgs.MapAddArgs.Value))
			}
		case api.TransformationType_TransformMapMult:
			switch transformArgs := spec.Transformation.TransformArgs.(type) {
			case *api.Transformation_MapMultArgs:
				builder.AddFieldTransformation(core.MapMultConstant(transformArgs.MapMultArgs.Value))
			}
		case api.TransformationType_TransformMapRegex:
			switch transformArgs := spec.Transformation.TransformArgs.(type) {
			case *api.Transformation_MapRegexArgs:
				builder.AddFieldTransformation(core.MapApplyRegex(transformArgs.MapRegexArgs.Regex,
					transformArgs.MapRegexArgs.Replace))
			}
		case api.TransformationType_TransformMap:
			switch transformArgs := spec.Transformation.TransformArgs.(type) {
			case *api.Transformation_MapArgs:
				builder.AddFieldTransformation(core.NewExecMapTransformation(transformArgs.MapArgs.Path))
			}
		case api.TransformationType_TransformLeftFold:
			switch transformArgs := spec.Transformation.TransformArgs.(type) {
			case *api.Transformation_LeftFoldArgs:
				builder.AddFieldTransformation(core.NewExecMapTransformation(transformArgs.LeftFoldArgs.Path))
			}
		case api.TransformationType_TransformRightFold:
			switch transformArgs := spec.Transformation.TransformArgs.(type) {
			case *api.Transformation_RightFoldArgs:
				builder.AddFieldTransformation(core.NewExecMapTransformation(transformArgs.RightFoldArgs.Path))
			}
		}
		transformer.AddSpec(spec.SourceField, spec.TargetField, builder.Get())
	}
	return transformer, nil
}

func PipelinesFromJson(pipelineJson []byte) (*Pipelines, error) {
	var pipelinesPb api.Pipelines
	byteBuffer := bytes.NewBuffer(pipelineJson)
	err := jsonpb.Unmarshal(byteBuffer, &pipelinesPb)
	if err != nil {
		return nil, err
	}
	return PipelinesFromPb(&pipelinesPb)
}

func PipelinesFromPb(pipelinesPb *api.Pipelines)  (*Pipelines, error) {
	var builtPipelines  []*Pipeline
	var shutdownFns []func() error

	externalKVStores := make(map[string]kvs.KVStore)
	externalPublisher := make(map[string]streaming.Publisher)
	externalHttp := make(map[string]string)
	externalLocalFile := make(map[string]string)
	processes := make(map[string]PipelineProcess)

	// Get Uuid
	partitionUuid, err := gUuid.Parse(pipelinesPb.PartitionUuid)
	if err != nil {
		return nil, err
	}

	// Get external systems
	for _, externalSystem := range pipelinesPb.ExternalSystems {
		switch externalSystem.ExternalType {
		case api.ExternalType_ExternalKVStore:
			externalKVStores[externalSystem.Name] = kvs.NewMemKVStore(kvs.WithTracing())
			shutdownFns = append(shutdownFns, externalKVStores[externalSystem.Name].Close)
		case api.ExternalType_ExternalPubSub:
			externalPublisher[externalSystem.Name], err = streaming.NewMemPublisher(streaming.NewMemPubSub(),
				externalSystem.ConnectionString)
			if err != nil {
				return nil, err
			}
			shutdownFns = append(shutdownFns, externalPublisher[externalSystem.Name].Close)
		case api.ExternalType_ExternalHttp:
			externalHttp[externalSystem.Name] = externalSystem.ConnectionString
		case api.ExternalType_ExternalLocalFile:
			externalLocalFile[externalSystem.Name] = externalSystem.ConnectionString
		}
	}

	// Get processes
	for _, processDefinition := range pipelinesPb.ProcessDefinitions {
		switch procDef := processDefinition.ProcessDefinition.(type) {
		case *api.ProcessDefinition_Annotator:
			if _, ok := processes[procDef.Annotator.Name]; ok {
				msg := fmt.Sprintf("name conflict in process definitions: %s", procDef.Annotator.Name)
				return nil, util.NewInvalidError(msg)
			}
			annotatorBuilder := NewAnnotatorBuilder(procDef.Annotator.Name)
			for _, annotation := range procDef.Annotator.Annotations {
				condition, err := buildCondition(annotation.Condition)
				if err != nil {
					return nil, err
				}
				annotatorBuilder.Add(core.NewAnnotation(annotation.FieldKey, annotation.Value, condition))
			}
			processes[procDef.Annotator.Name] = annotatorBuilder.Build()
		case *api.ProcessDefinition_Aggregator:
			if _, ok := processes[procDef.Aggregator.Name]; ok {
				msg := fmt.Sprintf("name conflict in process definitions: %s", procDef.Aggregator.Name)
				return nil, util.NewInvalidError(msg)
			}
			if kvStore, ok := externalKVStores[procDef.Aggregator.StateStore]; ok {
				aggregation := core.NewAggregation(core.NewFieldAggregation(procDef.Aggregator.Aggregation.Key,
					protoAggregationTypeToInternal(procDef.Aggregator.Aggregation.AggregationType),
					procDef.Aggregator.Aggregation.GroupByKeys))
				condition, err := buildCondition(procDef.Aggregator.Condition)
				if err != nil {
					return nil, err
				}
				processes[procDef.Aggregator.Name] = NewAggregator(procDef.Aggregator.Name, aggregation, condition,
					state.NewKvStateStore(kvStore), procDef.Aggregator.AsyncCheckpoint, procDef.Aggregator.ForwardState)
			} else {
				msg := fmt.Sprintf("unknown kvStore for %s: %s", procDef.Aggregator.Name, procDef.Aggregator.StateStore)
				return nil, util.NewInvalidError(msg)
			}
		case *api.ProcessDefinition_Completer:
			if _, ok := processes[procDef.Completer.Name]; ok {
				msg := fmt.Sprintf("name conflict in process definitions: %s", procDef.Completer.Name)
				return nil, util.NewInvalidError(msg)
			}
			if kvStore, ok := externalKVStores[procDef.Completer.StateStore]; ok {
				completion := core.NewCompletion(procDef.Completer.Completion.JoinKeys,
					time.Millisecond*time.Duration(procDef.Completer.Completion.TimeoutMs))
				processes[procDef.Completer.Name] = NewCompleter(procDef.Completer.Name, completion, kvStore)
			} else {
				msg := fmt.Sprintf("unknown kvStore for %s: %s", procDef.Completer.Name, procDef.Completer.StateStore)
				return nil, util.NewInvalidError(msg)
			}

		case *api.ProcessDefinition_Filter:
			if _, ok := processes[procDef.Filter.Name]; ok {
				msg := fmt.Sprintf("name conflict in process definitions: %s", procDef.Filter.Name)
				return nil, util.NewInvalidError(msg)
			}
			filter, err := NewRegexKeyFilter(procDef.Filter.Name, procDef.Filter.Regex, procDef.Filter.KeepMatched)
			if err != nil {
				return nil, err
			}
			processes[procDef.Filter.Name] = filter
		case *api.ProcessDefinition_Transformer:
			if _, ok := processes[procDef.Transformer.Name]; ok {
				msg := fmt.Sprintf("name conflict in process definitions: %s", procDef.Transformer.Name)
				return nil, util.NewInvalidError(msg)
			}
			transformer, err := buildTransformer(procDef.Transformer)
			if err != nil {
				 return nil, err
			}
			processes[procDef.Transformer.Name] = transformer
		case *api.ProcessDefinition_Tee:
			if _, ok := processes[procDef.Tee.Name]; ok {
				msg := fmt.Sprintf("name conflict in process definitions: %s", procDef.Tee.Name)
				return nil, util.NewInvalidError(msg)
			}

			var transformer *Transformer = nil;

			if len(procDef.Tee.TransformerRef) > 0 {
				var ok bool
				if _, ok = processes[procDef.Tee.TransformerRef]; !ok {
					msg := fmt.Sprintf("%s is not a valid transformer ref", procDef.Tee.TransformerRef)
					return nil, util.NewInvalidError(msg)
				}
				if transformer, ok = processes[procDef.Tee.TransformerRef].(*Transformer); !ok {
					msg := fmt.Sprintf("%s is not a transformer process", procDef.Tee.TransformerRef)
					return nil, util.NewInvalidError(msg)
				}
			}

			condition, err := buildCondition(procDef.Tee.Condition)
			if err != nil {
				return nil, err
			}

			if external, ok := externalKVStores[procDef.Tee.OutputConnectorRef]; ok {
				processes[procDef.Tee.Name] = NewKVTee(procDef.Tee.Name, external, condition, transformer,
					procDef.Tee.AdditionalBody.AsMap())
			} else if external, ok := externalPublisher[procDef.Tee.OutputConnectorRef]; ok {
				processes[procDef.Tee.Name] = NewPubSubTee(procDef.Tee.Name, external, condition, transformer,
					procDef.Tee.AdditionalBody.AsMap())
			} else if external, ok := externalHttp[procDef.Tee.OutputConnectorRef]; ok {
				processes[procDef.Tee.Name] = NewHttpTee(procDef.Tee.Name, http.DefaultClient, external, condition,
					transformer,
					procDef.Tee.AdditionalBody.AsMap())
			} else if external, ok := externalLocalFile[procDef.Tee.OutputConnectorRef]; ok {
				if processes[procDef.Tee.Name], err = NewLocalFileTee(procDef.Tee.Name, external, condition,
					transformer, procDef.Tee.AdditionalBody.AsMap()); err != nil {
					return nil, err
				}
			} else {
				msg := fmt.Sprintf("%v is not a valid external reference", procDef.Tee.TransformerRef)
				return nil, util.NewInvalidError(msg)
			}

		case *api.ProcessDefinition_Spawner:
			if _, ok := processes[procDef.Spawner.Name]; ok {
				msg := fmt.Sprintf("name conflict in process definitions: %s", procDef.Spawner.Name)
				return nil, util.NewInvalidError(msg)
			}
			job := core.NewCmdJob(procDef.Spawner.Job.Runnable.PathToExec, procDef.Spawner.Job.Runnable.CmdArgs...)
			condition, err := buildCondition(procDef.Spawner.Condition)
			if err != nil {
				return nil, err
			}
			processes[procDef.Spawner.Name] = NewSpawner(procDef.Spawner.Name, job, condition,
				time.Duration(procDef.Spawner.DelayInMs) * time.Millisecond, procDef.Spawner.DoSync)
		case *api.ProcessDefinition_Continuation:
			if _, ok := processes[procDef.Continuation.Name]; ok {
				msg := fmt.Sprintf("name conflict in process definitions: %s", procDef.Continuation.Name)
				return nil, util.NewInvalidError(msg)
			}
			condition, err := buildCondition(procDef.Continuation.Condition)
			if err != nil {
				return nil, err
			}
			processes[procDef.Continuation.Name] = NewContinuation(procDef.Continuation.Name, condition)
		}
	}

	// Build pipelines
	for _, pipeline := range pipelinesPb.Pipelines {
		pipelineBuilder := NewPipelineBuilder()
		pipelineBuilder.SetName(pipeline.Name)
		if pipeline.EnableTracing {
			pipelineBuilder.EnableTracing()
		}
		if pipeline.EnableMetrics {
			pipelineBuilder.EnableMetrics()
		}
		// Each pipeline must set annotations to ensure there are no conflicts in the state stores
		annotatorBuilder := NewAnnotatorBuilder("internalAnnotator")
		annotatorBuilder.Add(core.NewAnnotation(string(common.PartitionIDKey), partitionUuid.String(),
			core.TrueCondition))
		annotatorBuilder.Add(core.NewAnnotation(string(common.ResourceNameKey), pipeline.Name, core.TrueCondition))

		pipelineBuilder.Add(annotatorBuilder.Build())

		for _, processDesc := range pipeline.Processes {
			if proc, ok := processes[processDesc.Name]; ok {
				lifecycle := buildLifecycle(pipeline.Name, processDesc.Name, processDesc.Instrumentation)
				pipelineBuilder.Add(proc, WithRetry(processDesc.RetryStrategy),
					WithLifecycle(lifecycle))

			} else {
				msg := fmt.Sprintf("cannot find process: %s", processDesc.Name)
				return nil, util.NewInvalidError(msg)
			}
		}

		if pipeline.Checkpoint != nil {
			var checkPointer *CheckPointer
			if external, ok := externalKVStores[pipeline.Checkpoint.OutputConnectorRef]; ok {
				checkPointer = NewKVCheckPointer(pipeline.Name, external)
			}  else if external, ok := externalLocalFile[pipeline.Checkpoint.OutputConnectorRef]; ok {
				checkPointer, err = NewLocalFileCheckPointer(pipeline.Name, external)
				if err != nil {
					return nil, util.NewInvalidError(err.Error())
				}
			} else {
				msg := fmt.Sprintf("%s is not a valid KVStore reference", pipeline.Checkpoint.OutputConnectorRef)
				return nil, util.NewInvalidError(msg)
			}
			pipelineBuilder.Checkpoint(checkPointer)
		}

		builtPipelines = append(builtPipelines, pipelineBuilder.Get())
	}
	return NewPipelines(builtPipelines, shutdownFns), nil
}

func processKey(pipelineName, processName string) string {
	return fmt.Sprintf("%s.%s", pipelineName, processName)
}

func buildLifecycle(pipelineName, processName string, instrumentation *api.ProcessInstrumentation) *Lifecycle {
	lifecycleBuilder := NewLifecycleBuilder()

	if instrumentation == nil {
		return lifecycleBuilder.Build()
	}

	emitter := observability.NewEmitter("agglo/process")

	if instrumentation.EnableTracing {
		lifecycleBuilder.AppendPrepareFn(func(ctx, prev context.Context) context.Context {
			ctx, _ = emitter.CreateSpan(ctx, processKey(pipelineName, processName))
			return ctx
		})
		lifecycleBuilder.AppendSuccessFn(func(ctx context.Context) {
			span := trace.SpanFromContext(ctx)
			if span != nil {
				span.End()
			}
		})
		lifecycleBuilder.AppendFailFn(func(ctx context.Context, err error) {
			span := trace.SpanFromContext(ctx)
			if span != nil {
				span.RecordError(err)
				span.End()
			}
		})
	}

	if instrumentation.Latency {
		emitter.AddMetric(processKey(pipelineName, processName) + ".latency", observability.Int64Recorder)
		lifecycleBuilder.AppendPrepareFn(func(ctx, prev context.Context) context.Context {
			startTime := time.Now()
			return common.InjectProcessStartTime(processKey(pipelineName, processName), startTime, ctx)
		})
		lifecycleBuilder.AppendSuccessFn(func(ctx context.Context) {
			startTime := common.ExtractProcessStartTime(processKey(pipelineName, processName), ctx)
			if startTime != common.InvalidTime {
				emitter.RecordInt64(processKey(pipelineName, processName) + ".latency",
					time.Now().Sub(startTime).Nanoseconds())
			}
		})
	}

	if instrumentation.Counter {
		emitter.AddMetric(processKey(pipelineName, processName) + ".success", observability.Int64Counter)
		emitter.AddMetric(processKey(pipelineName, processName) + ".failure", observability.Int64Counter)
		lifecycleBuilder.AppendSuccessFn(func(ctx context.Context) {
			emitter.AddInt64(processKey(pipelineName, processName) + ".success", 1)
		})
		lifecycleBuilder.AppendFailFn(func(ctx context.Context, err error) {
			emitter.AddInt64(processKey(pipelineName, processName) + ".failure", 1)
		})
	}

	return lifecycleBuilder.Build()
}
