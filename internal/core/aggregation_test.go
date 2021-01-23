package core_test

import (
	"github.com/kmgreen2/agglo/internal/core"
	"github.com/stretchr/testify/assert"
	"testing"
)

func genAggregationMaps() []map[string]interface{} {
	out := []map[string]interface{} {
		{
			"name": "kevin",
			"num": 2,
		},
		{
			"name": "bob",
			"num": 1,
		},
		{
			"name": "alice",
			"num": 10,
		},
		{
			"name": "bob",
			"num": 4,
		},
		{
			"name": "alice",
			"num": 2,
		},
		{
			"name": "kevin",
			"num": 12,
		},
		{
			"name": "alice",
			"num": 2,
		},
		{
			"name": "bob",
			"num": 1,
		},
	}

	// JSON will always decode float64, so ensure numbers are float64
	for i, _ := range out {
		out[i]["num"] = float64(out[i]["num"].(int))
	}

	return out
}

func doBasicAggGroupBy(t *testing.T, aggType core.AggregationType, expected map[string]float64) {
	var err error
	var valMap map[string]interface{}
	maps := genAggregationMaps()
	aggregation := core.NewAggregation(core.NewFieldAggregation("num", aggType, []string{"name"}))
	aggState := core.NewAggregationState(make(map[string]interface{}))


	for _, m := range maps {
		_, _, err := aggregation.Update(m, aggState)
		assert.Nil(t, err)
	}

	switch aggType {
	case core.AggMin:
		valMap, err = aggState.Get([]string{"num:AggMin"})
	case core.AggMax:
		valMap, err = aggState.Get([]string{"num:AggMax"})
	case core.AggCount:
		valMap, err = aggState.Get([]string{"num:AggCount"})
	case core.AggSum:
		valMap, err = aggState.Get([]string{"num:AggSum"})
	}

	if err != nil {
		assert.FailNow(t, err.Error())
	}

	if kevin, ok := valMap["kevin"].(core.FieldAggregationState); ok {
		assert.Equal(t, map[string]interface{}{"value": expected["kevin"]}, kevin.ToMap())
	}

	if alice, ok := valMap["alice"].(core.FieldAggregationState); ok {
		assert.Equal(t, map[string]interface{}{"value": expected["alice"]}, alice.ToMap())
	}

	if bob, ok := valMap["bob"].(core.FieldAggregationState); ok {
		assert.Equal(t, map[string]interface{}{"value": expected["bob"]}, bob.ToMap())
	}
}

func TestBasicAggregationGroupBy(t *testing.T) {
	doBasicAggGroupBy(t, core.AggMin, map[string]float64{
		"kevin": 2,
		"alice": 2,
		"bob": 1,
	})

	doBasicAggGroupBy(t, core.AggMax, map[string]float64{
		"kevin": 12,
		"alice": 10,
		"bob": 4,
	})

	doBasicAggGroupBy(t, core.AggCount, map[string]float64{
		"kevin": 2,
		"alice": 3,
		"bob": 3,
	})

	doBasicAggGroupBy(t, core.AggSum, map[string]float64{
		"kevin": 14,
		"alice": 14,
		"bob": 6,
	})
}
