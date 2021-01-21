package process

import (
	"context"
	"errors"
	gUuid "github.com/google/uuid"
	"github.com/kmgreen2/agglo/pkg/common"
	"github.com/kmgreen2/agglo/pkg/core"
	"github.com/kmgreen2/agglo/pkg/kvs"
	"time"
)

type Aggregator struct {
	aggregation *core.Aggregation
	condition *core.Condition
	aggregatorStateStore kvs.StateStore
	asyncCheckpoint bool
	forwardState bool
}

func NewAggregator(aggregation *core.Aggregation, condition *core.Condition, stateStore kvs.StateStore) *Aggregator {
	return &Aggregator{
		aggregation: aggregation,
		condition: condition,
		aggregatorStateStore: stateStore,
	}
}

func (a Aggregator) getAggregationState(ctx context.Context, partitionID gUuid.UUID, name string) ([]byte, error) {
	stateBytes, err := a.aggregatorStateStore.Get(ctx, core.AggregationStateKey(partitionID, name))
	if err != nil {
		if errors.Is(err, &common.NotFoundError{}) {
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
		valMap, err := common.JsonToMap(val)
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
		return in, err
	}

	out := core.CopyableMap(in).DeepCopy()

	partitionID, err := core.GetPartitionID(in)
	if err != nil {
		return out, err
	}
	name, err := core.GetName(in)
	if err != nil {
		return out, err
	}

	inBytes, err := common.MapToJson(in)
	if err != nil {
		return nil, err
	}

	err = a.aggregatorStateStore.Append(ctx, core.AggregationStateKey(partitionID, name), inBytes)
	if err != nil {
		return nil, err
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
			return nil, err
		}
	}

	if !a.forwardState {
		var aggregationState *core.AggregationState

		aggregationStateBytes, err := a.getAggregationState(ctx, partitionID, name)
		if err != nil {
			return nil, err
		}

		if aggregationStateBytes == nil {
			aggregationState = core.NewAggregationState(make(map[string]interface{}))
			return out, nil
		} else {
			aggregationState, err = core.NewAggregationStateFromBytes(aggregationStateBytes)
			if err != nil {
				return nil, err
			}
		}
		err = common.SetUsingInternalPrefix(common.AggregationDataPrefix, name, aggregationState.Values,
			out, true)
		if err != nil {
			return nil, err
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
