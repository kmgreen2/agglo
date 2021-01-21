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

	// ToDo(KMG): Aggregations are maintained as follows:
	//
	// 1. Get the current aggregation state (stored in a KVStore), create if does not exist
	// 2. Conditionally update the aggregation
	// 3. Serialize and do an atomic put to the KVStore
	// 4. Annotate the output map with the updated aggregation
	//
	// In high-contention scenarios, step 3 could fail (racing updates).  There are a few solutions:
	//
	// 1. Block concurrent updates to aggregations: hard to do when multiple processes are updating the same aggregation
	// 2. Retry: this could lead to starvation
	// 3. Maintain log of deltas and opportunistically apply: this could limit certain aggregations
	//
	// For now, we will do #2, because it is easy
	//
	// The retry logic is abstracted away in CreateRetryableFuture, which can is to build a retryable process in
	// a pipeline.
	//
	// === Future work ===
	//
	// The ultimate solution is to use the underlying KVStore to maintain a log of deltas and ensure an Aggregation
	// implements a `Delta(other Aggregation) Aggregation` function, which would ensure that any update could be
	// applied in any order after an initial attempt.
	//
	// NOTE: The delta value *must* be computed against the current aggregation state.  All we are doing is deriving the
	// individual deltas for the overall aggregation to ensure they can be applied against a different aggregation
	// state.  From a software engineering perspective, it seems cleaner to let the underlying Aggregation objects
	// decide what the deltas are, especially since the underlying values are treated differently by each aggregation.
	//
	// For example,
	//
	// Min -> x.Delta(y) = min(x,y) : the delta is the max of the two aggregations
	// Max -> x.Delta(y) = max(x,y) : the delta is the min of the two aggregations
	// Sum -> x.Delta(y) = y - x : the delta is the additive difference between the two sums
	// Avg -> x.Delta(y) = {num: -1, y.sum-x.sum}: the delta is the difference between the num and sum fields
	// Count -> x.Delta(y) = 1
	// Histogram -> x.Delta(y) = {"<bucket>": -1} : the difference should be
	//
	// Notice that the differences should only 1 for the count-based aggregations.  Without this, the delta cannot
	// be replayed with any other future aggregation state.
	//

	inBytes, err := common.MapToJson(in)
	if err != nil {
		return nil, err
	}

	err = a.aggregatorStateStore.Append(ctx, core.AggregationStateKey(partitionID, name), inBytes)
	if err != nil {
		return nil, err
	}

	// ToDo(KMG): We need a standard way to control the checkpoint calls...
	go func() {
		<- time.NewTimer(100 * time.Millisecond).C
		// ToDo(KMG): Log and mark checkpoint failure
		_ = a.Checkpoint(ctx, in)
	}()

	updatedMap := make(map[string]interface{})

	aggregationStateBytes, err := a.getAggregationState(ctx, partitionID, name)
	if err != nil {
		return nil, err
	}

	var aggregationState *core.AggregationState

	if aggregationStateBytes == nil {
		aggregationState = core.NewAggregationState(make(map[string]interface{}))
	} else {
		aggregationState, err = core.NewAggregationStateFromBytes(aggregationStateBytes)
		if err != nil {
			return nil, err
		}
	}
	updatedKeys, updatedValues, err := a.aggregation.Update(in, aggregationState)
	if err != nil {
		return nil, err
	}
	for i, _ := range updatedKeys {
		updatedMap[updatedKeys[i]] = updatedValues[i]
	}

	err = common.SetUsingInternalPrefix(common.AggregationDataPrefix, name, updatedMap, out, true)
	if err != nil {
		return nil, err
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
