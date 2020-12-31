package pipeline_test

import (
	"context"
	"fmt"
	gUuid "github.com/google/uuid"
	"github.com/kmgreen2/agglo/pkg/core"
	"github.com/kmgreen2/agglo/pkg/core/pipeline"
	"github.com/kmgreen2/agglo/pkg/kvs"
	"github.com/kmgreen2/agglo/test"
	"github.com/stretchr/testify/assert"
	"math"
	"reflect"
	"strings"
	"testing"
)

type EvaluateValAt func(val interface{}, index int) error

func doBasicAggregation(maps []map[string]interface{}, aggType core.AggregationType,
name string, partitionID gUuid.UUID, aggPath string, evalValAt EvaluateValAt) (interface{}, error) {
	kvStore := kvs.NewMemKVStore()

	fieldAggregation := core.NewFieldAggregation(aggPath, aggType, []string{})

	aggregation := core.NewAggregation(partitionID, name, fieldAggregation)

	aggregator := pipeline.NewAggregator(aggregation, core.TrueCondition, kvStore)

	for i, m := range maps {
		out, err := aggregator.Process(m)
		if err != nil {
			return nil, err
		}

		val, err := core.GetMap(out,
			[]string{
				fmt.Sprintf("agglo:aggregation:%s", name),
				fmt.Sprintf("%s:%s", fieldAggregation.Key, aggType.String()),
			})

		if err := evalValAt(val, i); err != nil {
			return nil, err
		}
	}

	stateBytes, err := kvStore.Get(context.Background(), core.AggregationStateKey(partitionID, name))
	if err != nil {
		return nil, err
	}

	state, err := core.NewAggregationStateFromBytes(stateBytes)
	if err != nil {
		return nil, err
	}

	val, err := core.GetMap(state.Values,
		[]string{
			fmt.Sprintf("%s:%s", fieldAggregation.Key, aggType.String()),
		})

	if err != nil {
		return nil, err
	}
	return val, nil
}

func TestBasicCount(t *testing.T) {
	numMaps := 16
	var paths [][]string = [][]string{
		{"foo", "bar", "baz"},
	}

	name := "foo"

	partitionID, err := gUuid.NewUUID()
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	maps, _ := test.GetAggMapsWithFloats(numMaps, paths, partitionID, name)
	evalFunc := func(val interface{}, index int) error {
		switch v := val.(type) {
		case int64:
			if int64(index+1) == v {
				return nil
			}
			return fmt.Errorf("%d != %d", v, index+1)
		default:
			return fmt.Errorf("invalid type for val: %v", reflect.TypeOf(v))
		}
	}


	val, err := doBasicAggregation(maps, core.AggCount, name, partitionID, strings.Join(paths[0], "."), evalFunc)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	countState, err := core.AggregationCountStateFromMap(val.(map[string]interface{}))
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	assert.Equal(t, int64(numMaps), countState.Value)
}

func TestBasicSum(t *testing.T) {
	numMaps := 16
	var paths [][]string = [][]string{
		{"foo", "bar", "baz"},
	}

	name := "foo"

	partitionID, err := gUuid.NewUUID()
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	maps, mapValues := test.GetAggMapsWithFloats(numMaps, paths, partitionID, name)
	evalFunc := func(val interface{}, index int) error {
		sum := float64(0)
		for i := 0; i <= index; i++ {
			sum += mapValues[i][0]
		}
		switch v := val.(type) {
		case float64:
			if sum == v {
				return nil
			}
			return fmt.Errorf("%f != %f", v, sum)
		default:
			return fmt.Errorf("invalid type for val: %v", reflect.TypeOf(v))
		}
	}


	val, err := doBasicAggregation(maps, core.AggSum, name, partitionID, strings.Join(paths[0], "."), evalFunc)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	sumState, err := core.AggregationSumStateFromMap(val.(map[string]interface{}))
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	sum := float64(0)
	for i := 0; i < len(mapValues); i++ {
		sum += mapValues[i][0]
	}


	assert.Equal(t, sum, sumState.Value)
}

func TestBasicMax(t *testing.T) {
	numMaps := 16
	var paths [][]string = [][]string{
		{"foo", "bar", "baz"},
	}

	name := "foo"

	partitionID, err := gUuid.NewUUID()
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	maps, mapValues := test.GetAggMapsWithFloats(numMaps, paths, partitionID, name)
	evalFunc := func(val interface{}, index int) error {
		switch v := val.(type) {
		case float64:
			if mapValues[index][0] <= v {
				return nil
			}
			return fmt.Errorf("%f !> %f", v, mapValues[index][0])
		default:
			return fmt.Errorf("invalid type for val: %v", reflect.TypeOf(v))
		}
	}


	val, err := doBasicAggregation(maps, core.AggMax, name, partitionID, strings.Join(paths[0], "."), evalFunc)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	maxState, err := core.AggregationMaxStateFromMap(val.(map[string]interface{}))
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	max := float64(-1)
	for i := 0; i < len(mapValues); i++ {
		if mapValues[i][0] > max {
			max = mapValues[i][0]
		}
	}


	assert.Equal(t, max, maxState.Value)
}

func TestBasicMin(t *testing.T) {
	numMaps := 16
	var paths [][]string = [][]string{
		{"foo", "bar", "baz"},
	}

	name := "foo"

	partitionID, err := gUuid.NewUUID()
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	maps, mapValues := test.GetAggMapsWithFloats(numMaps, paths, partitionID, name)
	evalFunc := func(val interface{}, index int) error {
		switch v := val.(type) {
		case float64:
			if mapValues[index][0] >= v {
				return nil
			}
			return fmt.Errorf("%f !< %f", v, mapValues[index][0])
		default:
			return fmt.Errorf("invalid type for val: %v", reflect.TypeOf(v))
		}
	}


	val, err := doBasicAggregation(maps, core.AggMin, name, partitionID, strings.Join(paths[0], "."), evalFunc)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	minState, err := core.AggregationMinStateFromMap(val.(map[string]interface{}))
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	min := math.MaxFloat64
	for i := 0; i < len(mapValues); i++ {
		if mapValues[i][0] < min {
			min = mapValues[i][0]
		}
	}


	assert.Equal(t, min, minState.Value)
}

func TestBasicAvg(t *testing.T) {
	numMaps := 16
	var paths [][]string = [][]string{
		{"foo", "bar", "baz"},
	}

	name := "foo"

	partitionID, err := gUuid.NewUUID()
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	maps, mapValues := test.GetAggMapsWithFloats(numMaps, paths, partitionID, name)

	evalFunc := func(val interface{}, index int) error {
		sum := float64(0)
		for i := 0; i <= index; i++ {
			sum += mapValues[i][0]
		}
		avg := sum / float64(index+1)
		switch v := val.(type) {
		case float64:
			if avg == v {
				return nil
			}
			return fmt.Errorf("%f != %f", v, avg)
		default:
			return fmt.Errorf("invalid type for val: %v", reflect.TypeOf(v))
		}
	}


	val, err := doBasicAggregation(maps, core.AggAvg, name, partitionID, strings.Join(paths[0], "."), evalFunc)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	avgState, err := core.AggregationAvgStateFromMap(val.(map[string]interface{}))
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	sum := float64(0)
	for i := 0; i < len(mapValues); i++ {
		sum += mapValues[i][0]
	}


	assert.Equal(t, sum / float64(len(mapValues)), avgState.Sum / avgState.Num)
}

func TestBasicDiscreteHistogram(t *testing.T) {
	numMaps := 16
	var paths [][]string = [][]string{
		{"foo", "bar", "baz"},
	}

	name := "foo"

	partitionID, err := gUuid.NewUUID()
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	maps, mapValues := test.GetAggMapsWithStrings(numMaps, paths, partitionID, name, 2)

	buckets := make(map[string]int)
	evalFunc := func(val interface{}, index int) error {
		buckets[mapValues[index][0]]++
		switch v := val.(type) {
		case map[string]int:
			if len(buckets) == len(v) {
				for k, _ := range buckets {
					if buckets[k] != v[k] {
						return fmt.Errorf("%v != %v", v, buckets)
					}
				}
			}
			return nil
		default:
			return fmt.Errorf("invalid type for val: %v", reflect.TypeOf(v))
		}
	}


	val, err := doBasicAggregation(maps, core.AggDiscreteHistogram, name, partitionID, strings.Join(paths[0], "."),
		evalFunc)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	histogramState, err := core.AggregationDiscreteHistogramStateFromMap(val.(map[string]interface{}))
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	assert.Equal(t, buckets, histogramState.Buckets)
}