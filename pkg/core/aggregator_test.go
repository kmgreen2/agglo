package core_test

import (
	"context"
	"fmt"
	gUuid "github.com/google/uuid"
	"github.com/kmgreen2/agglo/pkg/core"
	"github.com/kmgreen2/agglo/pkg/kvs"
	"github.com/kmgreen2/agglo/test"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAggregationCount(t *testing.T) {
	var paths [][]string = [][]string{
		{"foo", "bar", "baz"},
		{"fizz", "buzz"},
	}

	numMaps := 16
	name := "foo"

	partitionID, err := gUuid.NewUUID()
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	maps, _ := test.GetAggMaps(numMaps, paths, partitionID, name)

	kvStore := kvs.NewMemKVStore()

	fieldAggregation := core.NewFieldAggregation("foo.bar.baz", core.AggCount, []string{})

	aggregation := core.NewAggregation(partitionID, name, []*core.FieldAggregation{fieldAggregation})

	aggregator := core.NewAggregator(aggregation, core.TrueCondition, kvStore)

	for i, m := range maps {
		out, err := aggregator.Process(m)
		if err != nil {
			assert.FailNow(t, err.Error())
		}

		val, err := core.GetMap(out,
			[]string{
				fmt.Sprintf("agglo:aggregation:%s", name),
				fmt.Sprintf("%s:%s", fieldAggregation.Key, core.AggCount.String()),
			})

		assert.Equal(t, int64(i + 1), val)
	}

	stateBytes, err := kvStore.Get(context.Background(), core.AggregationStateKey(partitionID, name))
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	state, err := core.NewAggregationStateFromBytes(stateBytes)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	val, err := core.GetMap(state.Values,
		[]string{
			fmt.Sprintf("%s:%s", fieldAggregation.Key, core.AggCount.String()),
		})

	if err != nil {
		assert.FailNow(t, err.Error())
	}

	countState, err := core.AggregationCountStateFromMap(val.(map[string]interface{}))
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	assert.Equal(t, int64(len(maps)), countState.Value)
}
