package process

import (
	"context"
	gUuid "github.com/google/uuid"
	"github.com/kmgreen2/agglo/internal/common"
	"github.com/kmgreen2/agglo/internal/core"
	"github.com/kmgreen2/agglo/pkg/state"
	"github.com/kmgreen2/agglo/pkg/util"
	"github.com/pkg/errors"
	"time"
)

type Aggregator struct {
	aggregation *core.Aggregation
	condition *core.Condition
	aggregatorStateStore state.StateStore
	asyncCheckpoint bool
	forwardState bool
}

func NewAggregator(aggregation *core.Aggregation, condition *core.Condition, stateStore state.StateStore,
	asyncCheckpoint, forwardState bool) *Aggregator {
	return &Aggregator{
		aggregation: aggregation,
		condition: condition,
		aggregatorStateStore: stateStore,
		asyncCheckpoint: asyncCheckpoint,
		forwardState: forwardState,
	}
}

func (a Aggregator) getAggregationState(ctx context.Context, partitionID gUuid.UUID, name string) ([]byte, error) {
	stateBytes, err := a.aggregatorStateStore.Get(ctx, core.AggregationStateKey(partitionID, name))
	if err != nil {
		if errors.Is(err, &util.NotFoundError{}) {
			return nil, nil
		}
		return nil, err
	}
	return stateBytes, nil
}

func (a Aggregator) checkpointMapFunc() (func(curr, val []byte) ([]byte, error)) {
	return func(curr, val []byte) ([]byte, error) {
		var aggregationState *core.AggregationState
		var err error
		if curr == nil {
			aggregationState = core.NewAggregationState(make(map[string]interface{}))
		} else {
			aggregationState, err = core.NewAggregationStateFromBytes(curr)
			if err != nil {
				return nil, err
			}
		}
		valMap, err := util.JsonToMap(val)
		if err != nil {
			return nil, err
		}

		_, _, err = a.aggregation.Update(valMap, aggregationState)
		if err != nil {
			return nil, err
		}
		return aggregationState.Bytes()
	}
}

func (a Aggregator) Process(ctx context.Context, in map[string]interface{}) (map[string]interface{}, error) {
	if shouldProcess, err := a.condition.Evaluate(in); !shouldProcess || err != nil {
		return in, PipelineProcessError(a, err, "evaluating condition")
	}

	out := util.CopyableMap(in).DeepCopy()

	partitionID, err := core.GetPartitionID(in)
	if err != nil {
		return out, PipelineProcessError(a, err, "get partition id")
	}
	name, err := core.GetName(in)
	if err != nil {
		return out, PipelineProcessError(a, err, "get pipeline name")
	}

	inBytes, err := util.MapToJson(in)
	if err != nil {
		return nil, PipelineProcessError(a, err, "converting map to JSON")
	}

	err = a.aggregatorStateStore.Append(ctx, core.AggregationStateKey(partitionID, name), inBytes)
	if err != nil {
		return nil, PipelineProcessError(a, err, "appending to state store")
	}

	if a.asyncCheckpoint {
		// ToDo(KMG): We need a standard way to control the checkpoint calls...
		go func() {
			<- time.NewTimer(100 * time.Millisecond).C
			// ToDo(KMG): Log and mark checkpoint failure
			_ = a.Checkpoint(ctx, in)
		}()
	} else {
		err = a.Checkpoint(ctx, in)
		if err != nil {
			return nil, PipelineProcessError(a, err, "checkpoint")
		}
	}

	if a.forwardState {
		var aggregationState *core.AggregationState

		aggregationStateBytes, err := a.getAggregationState(ctx, partitionID, name)
		if err != nil {
			return nil, PipelineProcessError(a, err, "get aggregation state")
		}

		if aggregationStateBytes == nil {
			aggregationState = core.NewAggregationState(make(map[string]interface{}))
			return out, nil
		} else {
			aggregationState, err = core.NewAggregationStateFromBytes(aggregationStateBytes)
			if err != nil {
				return nil, PipelineProcessError(a, err, "new aggregation state from bytes")
			}
		}
		err = common.SetUsingInternalPrefix(common.AggregationDataPrefix, name, aggregationState.Values,
			out, true)
		if err != nil {
			return nil, PipelineProcessError(a, err, "set new aggregation state")
		}
	}
	return out, nil
}

func (a Aggregator) Checkpoint(ctx context.Context, in map[string]interface{}) error {
	partitionID, err := core.GetPartitionID(in)
	if err != nil {
		return err
	}
	name, err := core.GetName(in)
	if err != nil {
		return err
	}
	return a.aggregatorStateStore.Checkpoint(ctx, core.AggregationStateKey(partitionID, name), a.checkpointMapFunc())
}
