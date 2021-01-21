package process_test

import (
	"context"
	"fmt"
	gUuid "github.com/google/uuid"
	"github.com/kmgreen2/agglo/pkg/common"
	"github.com/kmgreen2/agglo/pkg/core"
	"github.com/kmgreen2/agglo/pkg/core/process"
	"github.com/kmgreen2/agglo/pkg/kvs"
	"github.com/kmgreen2/agglo/test"
	"github.com/stretchr/testify/assert"
	"math"
	"reflect"
	"strings"
	"testing"
)



type EvaluateValAt func(val interface{}, index int) error

/**
 * NOTE: All of the aggregation tests run using doBasicAggregation intentionally test the aggregation
 * functions by serially calling Process() and Checkpoint().  This allows us to test the basic functionality.
 *
 * ToDo(KMG): Add doConcurrentAggregation, which will process maps in parallel, then call Checkpoint() and check the
 * final result.
 */
func doBasicAggregation(maps []map[string]interface{}, aggType core.AggregationType,
name string, partitionID gUuid.UUID, aggPath string, evalValAt EvaluateValAt) (interface{}, error) {
	stateStore := kvs.NewMemStateStore()

	fieldAggregation := core.NewFieldAggregation(aggPath, aggType, []string{})

	aggregation := core.NewAggregation(fieldAggregation)

	aggregator := process.NewAggregator(aggregation, core.TrueCondition, stateStore, false, true)

	for i, m := range maps {
		out, err := aggregator.Process(context.Background(), m)
		if err != nil {
			return nil, err
		}

		val, err := core.GetMap(out,
			[]string{
				common.InternalKeyFromPrefix(common.AggregationDataPrefix, name),
				fmt.Sprintf("%s:%s", fieldAggregation.Key, aggType.String()),
			})

		if err := evalValAt(val, i); err != nil {
			return nil, err
		}
	}

	stateBytes, err := stateStore.Get(context.Background(), core.AggregationStateKey(partitionID, name))
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

func getAggregationFromMap(key string, in interface{}) (interface{}, error) {
	if m, mapOk := in.(map[string]interface{}); !mapOk {
		return nil, fmt.Errorf("in is not a map")
	} else {
		if val, ok := m[key]; !ok {
			return nil, fmt.Errorf("cannot find key='%s' in map=%v", key, m)
		} else {
			return val, nil
		}
	}
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
		val, err := getAggregationFromMap("value", val)
		if err != nil {
			return err
		}
		switch v := val.(type) {
		case float64:
			if int64(index+1) == int64(v) {
				return nil
			}
			return fmt.Errorf("%d != %d", int64(v), index+1)
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
		val, err := getAggregationFromMap("value", val)
		if err != nil {
			return err
		}
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
		val, err := getAggregationFromMap("value", val)
		if err != nil {
			return err
		}
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
		val, err := getAggregationFromMap("value", val)
		if err != nil {
			return err
		}
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
		var aggNum, aggSum float64
		var ok bool
		sumVal, err := getAggregationFromMap("sum", val)
		if err != nil {
			return err
		}

		if aggSum, ok = sumVal.(float64); !ok {
			return fmt.Errorf("invalid type for val: %v", reflect.TypeOf(aggSum))
		}

		numVal, err := getAggregationFromMap("num", val)
		if err != nil {
			return err
		}

		if aggNum, ok = numVal.(float64); !ok {
			return fmt.Errorf("invalid type for val: %v", reflect.TypeOf(aggNum))
		}

		aggAvg := aggSum / aggNum

		sum := float64(0)
		for i := 0; i <= index; i++ {
			sum += mapValues[i][0]
		}
		avg := sum / float64(index+1)
		if avg == aggAvg {
			return nil
		}
		return fmt.Errorf("%f != %f", aggAvg, avg)
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
		aggState, err := core.AggregationDiscreteHistogramStateFromMap(val.(map[string]interface{}))
		if err != nil {
			return err
		}

		buckets[mapValues[index][0]]++
		if len(buckets) == len(aggState.Buckets) {
			for k, _ := range buckets {
				if buckets[k] != aggState.Buckets[k] {
					return fmt.Errorf("%v != %v", aggState.Buckets, buckets)
				}
			}
		} else {
			return fmt.Errorf("%v != %v", aggState.Buckets, buckets)
		}
		return nil
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