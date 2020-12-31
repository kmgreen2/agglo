package process

import (
	"context"
	"errors"
	"fmt"
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

func (a Aggregator) getAggregationState(partitionID gUuid.UUID, name string) ([]byte, error) {
	stateBytes, err := a.aggregatorStateStore.Get(context.Background(), core.AggregationStateKey(partitionID, name))
	if err != nil {
		if errors.Is(err, &common.NotFoundError{}) {
			return nil, nil
		}
		return nil, err
	}
	return stateBytes, nil
}

func (a Aggregator) updateAggregationState(partitionID gUuid.UUID, name string, prev, newState []byte) error {
	return a.aggregatorStateStore.AtomicPut(context.Background(), core.AggregationStateKey(partitionID, name), prev,
		newState)
}

func (a Aggregator) Process(in map[string]interface{}) (map[string]interface{}, error) {
	var aggregationState *core.AggregationState
	out := core.CopyableMap(in).DeepCopy()

	partitionID, err := core.GetPartitionID(in)
	if err != nil {
		return out, err
	}
	name, err := core.GetName(in)
	if err != nil {
		return out, err
	}

	stateBytes, err := a.getAggregationState(partitionID, name)
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

	err = a.updateAggregationState(partitionID, name, stateBytes, newStateBytes)
	if err != nil {
		return out, err
	}

	updatedMap := make(map[string]interface{})
	for i, _ := range updatedKeys {
		updatedMap[updatedKeys[i]] = updatedValues[i]
	}

	out[fmt.Sprintf("agglo:aggregation:%s", name)] = updatedMap

	return out, nil
}
