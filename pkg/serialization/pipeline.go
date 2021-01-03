package serialization

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	gUuid "github.com/google/uuid"
	"github.com/kmgreen2/agglo/pkg/common"
	"github.com/kmgreen2/agglo/pkg/core"
	"github.com/kmgreen2/agglo/pkg/core/process"
	"github.com/kmgreen2/agglo/pkg/core/proto"
	"github.com/kmgreen2/agglo/pkg/kvs"
	"github.com/kmgreen2/agglo/pkg/streaming"
	"net/http"
	"reflect"
	"time"
)

func protoExistsOpToInternal(in pipelineapi.ExistsOperator) core.ExistsOperator {
	switch in {
	case pipelineapi.ExistsOperator_Exists:
		return core.Exists
	case pipelineapi.ExistsOperator_NotExists:
		return core.NotExists
	default:
		// Should never reach this, but need to return something to satisfy the compiler
		return -1
	}
}

func protoAggregationTypeToInternal(in pipelineapi.AggregationType) core.AggregationType {
	switch in {
	case pipelineapi.AggregationType_AggAvg:
		return core.AggAvg
	case pipelineapi.AggregationType_AggCount:
		return core.AggCount
	case pipelineapi.AggregationType_AggMin:
		return core.AggMin
	case pipelineapi.AggregationType_AggMax:
		return core.AggMax
	case pipelineapi.AggregationType_AggDiscreteHistogram:
		return core.AggDiscreteHistogram
	case pipelineapi.AggregationType_AggSum:
		return core.AggSum
	default:
		return -1
	}
}

func buildExpression(expression *pipelineapi.Expression) (core.Expression, error) {
	var err error
	var lhs, rhs interface{}
	switch e := expression.Expression.(type) {
	case *pipelineapi.Expression_Boolean:
	case *pipelineapi.Expression_Comparator:
		switch lhsOperand := e.Comparator.Lhs.Operand.(type) {
		case *pipelineapi.Operand_Literal:
			lhs = lhsOperand.Literal
		case *pipelineapi.Operand_Numeric:
			lhs = lhsOperand.Numeric
		case *pipelineapi.Operand_Variable:
			lhs = core.Variable(lhsOperand.Variable.Name)
		case *pipelineapi.Operand_Expression:
			lhs, err = buildExpression(lhsOperand.Expression)
			if err != nil {
				return nil, err
			}
		}
		switch rhsOperand := e.Comparator.Rhs.Operand.(type) {
		case *pipelineapi.Operand_Literal:
			rhs = rhsOperand.Literal
		case *pipelineapi.Operand_Numeric:
			rhs = rhsOperand.Numeric
		case *pipelineapi.Operand_Variable:
			rhs = core.Variable(rhsOperand.Variable.Name)
		case *pipelineapi.Operand_Expression:
			rhs, err = buildExpression(rhsOperand.Expression)
			if err != nil {
				return nil, err
			}
		}
		switch e.Comparator.Op {
		case pipelineapi.ComparatorOperator_Equal:
			return core.NewComparatorExpression(lhs, rhs, core.Equal), nil
		case pipelineapi.ComparatorOperator_NotEqual:
			return core.NewComparatorExpression(lhs, rhs, core.NotEqual), nil
		case pipelineapi.ComparatorOperator_GreaterThan:
			return core.NewComparatorExpression(lhs, rhs, core.GreaterThan), nil
		case pipelineapi.ComparatorOperator_GreaterThanOrEqual:
			return core.NewComparatorExpression(lhs, rhs, core.GreaterThanOrEqual), nil
		case pipelineapi.ComparatorOperator_LessThan:
			return core.NewComparatorExpression(lhs, rhs, core.LessThan), nil
		case pipelineapi.ComparatorOperator_LessThanOrEqual:
			return core.NewComparatorExpression(lhs, rhs, core.LessThanOrEqual), nil
		case pipelineapi.ComparatorOperator_RegexMatch:
			return core.NewComparatorExpression(lhs, rhs, core.RegexMatch), nil
		case pipelineapi.ComparatorOperator_RegexNotMatch:
			return core.NewComparatorExpression(lhs, rhs, core.RegexNotMatch), nil
		}

	case *pipelineapi.Expression_Logical:
		switch lhsOperand := e.Logical.Lhs.Operand.(type) {
		case *pipelineapi.Operand_Expression:
			lhs, err = buildExpression(lhsOperand.Expression)
			if err != nil {
				return nil, err
			}
		default:
			msg := fmt.Sprintf("logical expression operands *must* be expressions, not literal, " +
				"variable or numeric.  Got %v", reflect.TypeOf(lhsOperand))
			return nil, common.NewInvalidError(msg)
		}
		switch rhsOperand := e.Logical.Rhs.Operand.(type) {
		case *pipelineapi.Operand_Expression:
			rhs, err = buildExpression(rhsOperand.Expression)
			if err != nil {
				return nil, err
			}
		default:
			msg := fmt.Sprintf("logical expression operands *must* be expressions, not literal, " +
				"variable or numeric.  Got %v", reflect.TypeOf(rhsOperand))
			return nil, common.NewInvalidError(msg)
		}

		// ToDo(KMG): Could use type assertions for lhs and rhs, but skipping it because the assertions above
		// should guarantee that lhs and rhs are Expressions
		switch e.Logical.Op {
		case pipelineapi.LogicalOperator_LogicalAnd:
			return core.NewLogicalExpression(lhs.(core.Expression), rhs.(core.Expression), core.LogicalAnd), nil
		case pipelineapi.LogicalOperator_LogicalOr:
			return core.NewLogicalExpression(lhs.(core.Expression), rhs.(core.Expression), core.LogicalAnd), nil
		}
	case *pipelineapi.Expression_Binary:
		switch lhsOperand := e.Binary.Lhs.Operand.(type) {
		case *pipelineapi.Operand_Numeric:
			lhs = lhsOperand.Numeric
		case *pipelineapi.Operand_Variable:
			lhs = core.Variable(lhsOperand.Variable.Name)
		default:
			msg := fmt.Sprintf("binary expression operands *must* be numeric or variable, not literal, " +
				"or expressions.  Got %v", reflect.TypeOf(lhsOperand))
			return nil, common.NewInvalidError(msg)
		}
		switch rhsOperand := e.Binary.Rhs.Operand.(type) {
		case *pipelineapi.Operand_Numeric:
			rhs = rhsOperand.Numeric
		case *pipelineapi.Operand_Variable:
			rhs = core.Variable(rhsOperand.Variable.Name)
		default:
			msg := fmt.Sprintf("binary expression operands *must* be numeric or variable, not literal, " +
				"or expressions.  Got %v", reflect.TypeOf(rhsOperand))
			return nil, common.NewInvalidError(msg)
		}

		switch e.Binary.Op {
		case pipelineapi.BinaryOperator_Addition:
			return core.NewBinaryExpression(lhs, rhs, core.Addition), nil
		case pipelineapi.BinaryOperator_Subtract:
			return core.NewBinaryExpression(lhs, rhs, core.Subtract), nil
		case pipelineapi.BinaryOperator_Multiply:
			return core.NewBinaryExpression(lhs, rhs, core.Multiply), nil
		case pipelineapi.BinaryOperator_Divide:
			return core.NewBinaryExpression(lhs, rhs, core.Divide), nil
		case pipelineapi.BinaryOperator_Power:
			return core.NewBinaryExpression(lhs, rhs, core.Power), nil
		case pipelineapi.BinaryOperator_Modulus:
			return core.NewBinaryExpression(lhs, rhs, core.Modulus), nil
		case pipelineapi.BinaryOperator_RightShift:
			return core.NewBinaryExpression(lhs, rhs, core.RightShift), nil
		case pipelineapi.BinaryOperator_LeftShift:
			return core.NewBinaryExpression(lhs, rhs, core.LeftShift), nil
		case pipelineapi.BinaryOperator_Or:
			return core.NewBinaryExpression(lhs, rhs, core.Or), nil
		case pipelineapi.BinaryOperator_And:
			return core.NewBinaryExpression(lhs, rhs, core.And), nil
		case pipelineapi.BinaryOperator_Xor:
			return core.NewBinaryExpression(lhs, rhs, core.Xor), nil
		}

	case *pipelineapi.Expression_Unary:
		switch rhsOperand := e.Unary.Rhs.Operand.(type) {
		case *pipelineapi.Operand_Numeric:
			rhs = rhsOperand.Numeric
		case *pipelineapi.Operand_Variable:
			rhs = core.Variable(rhsOperand.Variable.Name)
		default:
			msg := fmt.Sprintf("unary expression operands *must* be numeric or variable, not literal, " +
				"or expressions.  Got %v", reflect.TypeOf(rhsOperand))
			return nil, common.NewInvalidError(msg)
		}
		switch e.Unary.Op {
		case pipelineapi.UnaryOperator_Negation:
			return core.NewUnaryExpression(rhs, core.Negation), nil
		case pipelineapi.UnaryOperator_Inversion:
			return core.NewUnaryExpression(rhs, core.Inversion), nil
		case pipelineapi.UnaryOperator_LogicalNot:
			return core.NewUnaryExpression(rhs, core.Not), nil
		}
	}
	return nil, nil
}

func buildCondition(condition *pipelineapi.Condition) (*core.Condition, error) {
	if condition.Condition == nil {
		return nil, nil
	}
	switch c := condition.Condition.(type) {
	case *pipelineapi.Condition_Expression:
		expression, err := buildExpression(c.Expression)
		if err != nil {
			return nil, err
		}
		return core.NewCondition(expression)
	case *pipelineapi.Condition_Exists:
		builder := core.NewExistsExpressionBuilder()
		for _, op := range c.Exists.Ops {
			builder.Add(op.Key, protoExistsOpToInternal(op.Op))
		}
		return core.NewCondition(builder.Get())
	default:
		msg := fmt.Sprintf("invalid condition type: %v", reflect.TypeOf(c))
		return nil, common.NewInvalidError(msg)
	}
}

func buildTransformer(transformerSpecs []*pipelineapi.TransformerSpec) (*process.Transformer, error) {
	transformer := process.NewTransformer(nil, ".", ".")
	for _, spec := range transformerSpecs {
		condition, err := buildCondition(spec.Transformation.Condition)
		if err != nil {
			return nil, err
		}
		builder := core.NewTransformationBuilder()
		builder.AddCondition(condition)
		switch spec.Transformation.TransformationType {
		case pipelineapi.TransformationType_TransformCopy:
			builder.AddFieldTransformation(&core.CopyTransformation{})
		case pipelineapi.TransformationType_TransformCount:
			builder.AddFieldTransformation(core.LeftFoldCountAll)
		case pipelineapi.TransformationType_TransformSum:
			builder.AddFieldTransformation(&core.SumTransformation{})
		case pipelineapi.TransformationType_TransformMapAdd:
			switch transformArgs := spec.Transformation.TransformArgs.(type) {
			case *pipelineapi.Transformation_MapAddArgs:
				builder.AddFieldTransformation(core.MapAddConstant(transformArgs.MapAddArgs.Value))
			}
		case pipelineapi.TransformationType_TransformMapMult:
			switch transformArgs := spec.Transformation.TransformArgs.(type) {
			case *pipelineapi.Transformation_MapMultArgs:
				builder.AddFieldTransformation(core.MapMultConstant(transformArgs.MapMultArgs.Value))
			}
		case pipelineapi.TransformationType_TransformMapRegex:
			switch transformArgs := spec.Transformation.TransformArgs.(type) {
			case *pipelineapi.Transformation_MapRegexArgs:
				builder.AddFieldTransformation(core.MapApplyRegex(transformArgs.MapRegexArgs.Regex,
					transformArgs.MapRegexArgs.Replace))
			}
		case pipelineapi.TransformationType_TransformMap:
			switch transformArgs := spec.Transformation.TransformArgs.(type) {
			case *pipelineapi.Transformation_MapArgs:
				builder.AddFieldTransformation(core.NewExecMapTransformation(transformArgs.MapArgs.Path))
			}
		case pipelineapi.TransformationType_TransformLeftFold:
			switch transformArgs := spec.Transformation.TransformArgs.(type) {
			case *pipelineapi.Transformation_LeftFoldArgs:
				builder.AddFieldTransformation(core.NewExecMapTransformation(transformArgs.LeftFoldArgs.Path))
			}
		case pipelineapi.TransformationType_TransformRightFold:
			switch transformArgs := spec.Transformation.TransformArgs.(type) {
			case *pipelineapi.Transformation_RightFoldArgs:
				builder.AddFieldTransformation(core.NewExecMapTransformation(transformArgs.RightFoldArgs.Path))
			}
		}
		transformer.AddSpec(spec.SourceField, spec.TargetField, builder.Get())
	}
	return transformer, nil
}

func PipelinesFromBytes(pipelineBytes []byte)  ([]*process.Pipeline, error) {
	var builtPipelines  []*process.Pipeline
	pipelines := &pipelineapi.Pipelines{}

	externalKVStores := make(map[string]kvs.KVStore)
	externalPublisher := make(map[string]streaming.Publisher)
	externalHttp := make(map[string]string)
	processes := make(map[string]process.PipelineProcess)

	if err := proto.Unmarshal(pipelineBytes, pipelines); err != nil {
		return nil, err
	}

	// Get Uuid
	partitionUuid, err := gUuid.Parse(pipelines.PartitionUuid)
	if err != nil {
		return nil, err
	}

	// Get external systems
	for _, externalSystem := range pipelines.ExternalSystems {
		switch externalSystem.ExternalType {
		case pipelineapi.ExternalType_ExternalKVStore:
			externalKVStores[externalSystem.Name] = kvs.NewMemKVStore()
		case pipelineapi.ExternalType_ExternalPubSub:
			externalPublisher[externalSystem.Name], err = streaming.NewMemPublisher(streaming.NewMemPubSub(),
				externalSystem.ConnectionString)
			if err != nil {
				return nil, err
			}
		case pipelineapi.ExternalType_ExternalHttp:
			externalHttp[externalSystem.Name] = externalSystem.ConnectionString
		}
	}

	// Get processes
	for _, processDefinition := range pipelines.ProcessDefinitions {
		switch procDef := processDefinition.ProcessDefinition.(type) {
		case *pipelineapi.ProcessDefinition_Annotator:
			if _, ok := processes[procDef.Annotator.Name]; ok {
				msg := fmt.Sprintf("name conflict in process definitions: %s", procDef.Annotator.Name)
				return nil, common.NewInvalidError(msg)
			}
			annotatorBuilder := process.NewAnnotatorBuilder()
			for _, annotation := range procDef.Annotator.Annotations {
				condition, err := buildCondition(annotation.Condition)
				if err != nil {
					return nil, err
				}
				annotatorBuilder.Add(core.NewAnnotation(annotation.FieldKey, annotation.Value, condition))
			}
			processes[procDef.Annotator.Name] = annotatorBuilder.Build()
		case *pipelineapi.ProcessDefinition_Aggregator:
			if _, ok := processes[procDef.Aggregator.Name]; ok {
				msg := fmt.Sprintf("name conflict in process definitions: %s", procDef.Aggregator.Name)
				return nil, common.NewInvalidError(msg)
			}
			if kvStore, ok := externalKVStores[procDef.Aggregator.StateStore]; ok {
				aggregation := core.NewAggregation(core.NewFieldAggregation(procDef.Aggregator.Aggregation.Key,
					protoAggregationTypeToInternal(procDef.Aggregator.Aggregation.AggregationType),
					procDef.Aggregator.Aggregation.GroupByKeys))
				condition, err := buildCondition(procDef.Aggregator.Condition)
				if err != nil {
					return nil, err
				}
				processes[procDef.Aggregator.Name] = process.NewAggregator(aggregation, condition, kvStore)
			} else {
				msg := fmt.Sprintf("unknown kvStore for %s: %s", procDef.Aggregator.Name, procDef.Aggregator.StateStore)
				return nil, common.NewInvalidError(msg)
			}
		case *pipelineapi.ProcessDefinition_Completer:
			if _, ok := processes[procDef.Completer.Name]; ok {
				msg := fmt.Sprintf("name conflict in process definitions: %s", procDef.Completer.Name)
				return nil, common.NewInvalidError(msg)
			}
			if kvStore, ok := externalKVStores[procDef.Completer.StateStore]; ok {
				completion := core.NewCompletion(procDef.Completer.Completion.JoinKeys,
					time.Millisecond*time.Duration(procDef.Completer.Completion.TimeoutMs))
				processes[procDef.Completer.Name] = process.NewCompleter(completion, kvStore)
			} else {
				msg := fmt.Sprintf("unknown kvStore for %s: %s", procDef.Completer.Name, procDef.Completer.StateStore)
				return nil, common.NewInvalidError(msg)
			}

		case *pipelineapi.ProcessDefinition_Filter:
			if _, ok := processes[procDef.Filter.Name]; ok {
				msg := fmt.Sprintf("name conflict in process definitions: %s", procDef.Filter.Name)
				return nil, common.NewInvalidError(msg)
			}
			filter, err := process.NewRegexKeyFilter(procDef.Filter.Regex, procDef.Filter.KeepMatched)
			if err != nil {
				return nil, err
			}
			processes[procDef.Filter.Name] = filter
		case *pipelineapi.ProcessDefinition_Transformer:
			if _, ok := processes[procDef.Transformer.Name]; ok {
				msg := fmt.Sprintf("name conflict in process definitions: %s", procDef.Transformer.Name)
				return nil, common.NewInvalidError(msg)
			}
			transformer, err := buildTransformer(procDef.Transformer.Specs)
			if err != nil {
				 return nil, err
			}
			processes[procDef.Transformer.Name] = transformer
		case *pipelineapi.ProcessDefinition_Tee:
			if _, ok := processes[procDef.Tee.Name]; ok {
				msg := fmt.Sprintf("name conflict in process definitions: %s", procDef.Tee.Name)
				return nil, common.NewInvalidError(msg)
			}
			transformer, err := buildTransformer(procDef.Tee.TransformerSpecs)
			if err != nil {
				return nil, err
			}

			condition, err := buildCondition(procDef.Tee.Condition)
			if err != nil {
				return nil, err
			}

			if procDef.Tee.ExternalType == pipelineapi.ExternalType_ExternalKVStore {
				if external, ok := externalKVStores[procDef.Tee.OutputConnector]; ok {
					processes[procDef.Tee.Name] = process.NewKVTee(external, condition, transformer)
				} else {
					msg := fmt.Sprintf("%s is not a valid KVStore", procDef.Tee.OutputConnector)
					return nil, common.NewInvalidError(msg)
				}
			} else if procDef.Tee.ExternalType == pipelineapi.ExternalType_ExternalPubSub {
				if external, ok := externalPublisher[procDef.Tee.OutputConnector]; ok {
					processes[procDef.Tee.Name] = process.NewPubSubTee(external, condition, transformer)
				} else {
					msg := fmt.Sprintf("%s is not a valid Publisher", procDef.Tee.OutputConnector)
					return nil, common.NewInvalidError(msg)
				}
			} else if procDef.Tee.ExternalType == pipelineapi.ExternalType_ExternalHttp {
				if external, ok := externalHttp[procDef.Tee.OutputConnector]; ok {
					processes[procDef.Tee.Name] = process.NewHttpTee(http.DefaultClient, external, condition, transformer)
				} else {
					msg := fmt.Sprintf("%s is not a valid Publisher", procDef.Tee.OutputConnector)
					return nil, common.NewInvalidError(msg)
				}
			} else {
				msg := fmt.Sprintf("%v is not a valid external system", procDef.Tee.ExternalType)
				return nil, common.NewInvalidError(msg)
			}

		case *pipelineapi.ProcessDefinition_Spawner:
			if _, ok := processes[procDef.Spawner.Name]; ok {
				msg := fmt.Sprintf("name conflict in process definitions: %s", procDef.Spawner.Name)
				return nil, common.NewInvalidError(msg)
			}
			runnable := common.NewExecRunnable(procDef.Spawner.Job.Runnable.PathToExec)
			job := core.NewLocalJob(runnable)
			condition, err := buildCondition(procDef.Spawner.Condition)
			if err != nil {
				return nil, err
			}
			processes[procDef.Spawner.Name] = process.NewSpawner(job, condition,
				time.Duration(procDef.Spawner.DelayInMs) * time.Millisecond, procDef.Spawner.DoSync)
		}
	}

	// Build pipelines
	for _, pipeline := range pipelines.Pipelines {
		pipelineBuilder := process.NewPipelineBuilder()
		// Each pipeline must set annotations to ensure there are no conflicts in the state stores
		annotatorBuilder := process.NewAnnotatorBuilder()
		annotatorBuilder.Add(core.NewAnnotation("agglo:internal:partitionID", partitionUuid.String(), core.TrueCondition))
		annotatorBuilder.Add(core.NewAnnotation("agglo:internal:name", pipeline.Name, core.TrueCondition))

		pipelineBuilder.Add(annotatorBuilder.Build())

		for _, processDesc := range pipeline.Processes {
			if proc, ok := processes[processDesc.Name]; ok {
				pipelineBuilder.Add(proc)
			} else {
				msg := fmt.Sprintf("cannot find process: %s", processDesc.Name)
				return nil, common.NewInvalidError(msg)
			}
		}

		builtPipelines = append(builtPipelines, pipelineBuilder.Get())
	}
	return builtPipelines, nil
}

