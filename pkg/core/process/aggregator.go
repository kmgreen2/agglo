package process

import (
	"context"
	"errors"
	gUuid "github.com/google/uuid"
	"github.com/kmgreen2/agglo/pkg/common"
	"github.com/kmgreen2/agglo/pkg/core"
	"github.com/kmgreen2/agglo/pkg/kvs"
)

type Aggregator struct {
	aggregation *core.Aggregation
	condition *core.Condition
	aggregatorStateStore kvs.KVStore
}

func NewAggregator(aggregation *core.Aggregation, condition *core.Condition, kvStore kvs.KVStore) *Aggregator {
	return &Aggregator{
		aggregation: aggregation,
		condition: condition,
		aggregatorStateStore: kvStore,
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

func (a Aggregator) updateAggregationState(ctx context.Context, partitionID gUuid.UUID, name string, prev,
	newState []byte) error {
	return a.aggregatorStateStore.AtomicPut(ctx, core.AggregationStateKey(partitionID, name), prev, newState)
}

func (a Aggregator) Process(ctx context.Context, in map[string]interface{}) (map[string]interface{}, error) {
	var aggregationState *core.AggregationState

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
	stateBytes, err := a.getAggregationState(ctx, partitionID, name)
	if err != nil {
		return out, err
	}

	if stateBytes == nil {
		aggregationState = core.NewAggregationState(make(map[string]interface{}))
	} else {
		aggregationState, err = core.NewAggregationStateFromBytes(stateBytes)
		if err != nil {
			return out, err
		}
	}

	updatedKeys, updatedValues, err := a.aggregation.Update(in, aggregationState)
	if err != nil {
		return out, err
	}

	newStateBytes, err := aggregationState.Bytes()
	if err != nil {
		return out, err
	}

	// Note: If there are any errors after this step, we should do a *hard* failure, since the process
	// cannot be retried.  Probably best to always call this function last
	err = a.updateAggregationState(ctx, partitionID, name, stateBytes, newStateBytes)
	if err != nil {
		return out, err
	}

	updatedMap := make(map[string]interface{})
	for i, _ := range updatedKeys {
		updatedMap[updatedKeys[i]] = updatedValues[i]
	}

	err = common.SetUsingInternalPrefix(common.AggregationDataPrefix, name, updatedMap, out, true)
	if err != nil {
		return nil, err
	}

	return out, nil
}
