package core_test

import (
	"fmt"
	"github.com/kmgreen2/agglo/pkg/core"
	"github.com/stretchr/testify/assert"
	"reflect"
	"regexp"
	"testing"
)

func TestMapAddConstant(t *testing.T) {
	testMap := map[string]interface{} {
		"foo": 1,
		"bar": 5,
		"baz": 7,
	}

	testSlice := []interface{} {
		2,
		4,
		6,
	}

	testVal := 7

	builder := core.NewTransformationBuilder()

	builder.AddFieldTransformation(core.MapAddConstant(5))

	transformation := builder.Get()

	result, err := transformation.Transform(core.NewTransformable(testMap))
	assert.Nil(t, err)

	switch v := result.Value().(type) {
	case map[string]interface{}:
		if len(v) != len(testMap) {
			assert.FailNow(t, fmt.Sprintf("map length mismatch %d != %d", len(v), len(testMap)))
		}
		for k, _ := range testMap {
			x, y, err := core.NumericResolver(v[k], testMap[k])
			if err != nil {
				assert.FailNow(t, err.Error())
			}
			assert.Equal(t, y+5, x)
		}
	default:
		assert.FailNow(t, fmt.Sprintf("invalid type %v", reflect.TypeOf(v)))
	}

	result, err = transformation.Transform(core.NewTransformable(testSlice))
	assert.Nil(t, err)

	switch v := result.Value().(type) {
	case []interface{}:
		if len(v) != len(testSlice) {
			assert.FailNow(t, fmt.Sprintf("slice length mismatch %d != %d", len(v), len(testSlice)))
		}
		for i, _ := range testSlice {
			x, y, err := core.NumericResolver(v[i], testSlice[i])
			if err != nil {
				assert.FailNow(t, err.Error())
			}
			assert.Equal(t, y+5, x)
		}
	default:
		assert.FailNow(t, fmt.Sprintf("invalid type %v", reflect.TypeOf(v)))
	}

	result, err = transformation.Transform(core.NewTransformable(testVal))
	assert.Nil(t, err)

	switch v := result.Value().(type) {
	case float64:
		x, y, err := core.NumericResolver(v, testVal)
		if err != nil {
			assert.FailNow(t, err.Error())
		}
		assert.Equal(t, y+5, x)
	default:
		assert.FailNow(t, fmt.Sprintf("invalid type %v", reflect.TypeOf(v)))
	}
}

func TestMapApplyRegex(t *testing.T) {
	testMap := make(map[string]interface{})
	testMap["foo"] = "fizz"
	testMap["bar"] = "buzz"
	testMap["baz"] = "butt"

	builder := core.NewTransformationBuilder()

	builder.AddFieldTransformation(core.MapApplyRegex(`.*zz`, "********"))

	transformation := builder.Get()

	result, err := transformation.Transform(core.NewTransformable(testMap))
	assert.Nil(t, err)

	re, err := regexp.Compile(`.*zz`)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	switch v := result.Value().(type) {
	case map[string]interface{}:
		if len(v) != len(testMap) {
			assert.FailNow(t, fmt.Sprintf("map length mismatch %d != %d", len(v), len(testMap)))
		}
		for key, val := range testMap {
			if re.Match([]byte(val.(string))) {
				if vStr, ok := v[key].(string); ok {
					assert.Equal(t, "********", vStr)
				}
			}
		}
	default:
		assert.FailNow(t, fmt.Sprintf("invalid type %v", reflect.TypeOf(v)))
	}
}

func TestMapMultConstant(t *testing.T) {
	testMap := map[string]interface{} {
		"foo": 1,
		"bar": 5,
		"baz": 7,
	}

	builder := core.NewTransformationBuilder()

	builder.AddFieldTransformation(core.MapMultConstant(5))

	transformation := builder.Get()

	result, err := transformation.Transform(core.NewTransformable(testMap))
	assert.Nil(t, err)

	switch v := result.Value().(type) {
	case map[string]interface{}:
		if len(v) != len(testMap) {
			assert.FailNow(t, fmt.Sprintf("map length mismatch %d != %d", len(v), len(testMap)))
		}
		for k, _ := range testMap {
			x, y, err := core.NumericResolver(v[k], testMap[k])
			if err != nil {
				assert.FailNow(t, err.Error())
			}
			assert.Equal(t, y*5, x)
		}
	default:
		assert.FailNow(t, fmt.Sprintf("invalid type %v", reflect.TypeOf(v)))
	}
}

func doTestFolds(t *testing.T, testSlice []interface{}, expectedLeft, expectedRight interface{},
	left *core.LeftFoldTransformation, right *core.RightFoldTransformation) {

	leftBuilder := core.NewTransformationBuilder()
	leftBuilder.AddFieldTransformation(left)
	leftTransformation := leftBuilder.Get()
	leftResult, err := leftTransformation.Transform(core.NewTransformable(testSlice))
	assert.Nil(t, err)

	rightBuilder := core.NewTransformationBuilder()
	rightBuilder.AddFieldTransformation(right)
	rightTransformation := rightBuilder.Get()
	rightResult, err := rightTransformation.Transform(core.NewTransformable(testSlice))
	assert.Nil(t, err)

	assert.Equal(t, expectedLeft, leftResult.Value())
	assert.Equal(t, expectedRight, rightResult.Value())
}

func TestFoldCount(t *testing.T) {
	testSlice := []interface{} {
		1,
		5,
		7,
	}

	doTestFolds(t, testSlice, float64(3), float64(3), core.LeftFoldCountAll, core.RightFoldCountAll)
}

func TestFoldMin(t *testing.T) {
	testSlice := []interface{} {
		1,
		5,
		7,
	}

	doTestFolds(t, testSlice, float64(1), float64(1), core.LeftFoldMin, core.RightFoldMin)
}

func TestFoldMax(t *testing.T) {
	testSlice := []interface{} {
		1,
		5,
		7,
	}

	doTestFolds(t, testSlice, float64(7), float64(7), core.LeftFoldMax, core.RightFoldMax)
}

func TestSumTransformation(t *testing.T) {
	testSlice := []interface{} {
		10,
		5,
		7,
	}

	builder := core.NewTransformationBuilder()

	builder.AddFieldTransformation(core.SumTransformation{})

	transformation := builder.Get()

	result, err := transformation.Transform(core.NewTransformable(testSlice))
	assert.Nil(t, err)

	switch v := result.Value().(type) {
	case float64:
		assert.Equal(t, float64(22), v)
	}
}

func TestCopyTransformation(t *testing.T) {
	testMap := map[string]interface{} {
		"foo": 1,
		"bar": 5,
		"baz": 7,
	}

	testSlice := []interface{} {
		"hello",
		5,
		7,
	}

	builder := core.NewTransformationBuilder()
	builder.AddFieldTransformation(core.CopyTransformation{})
	transformation := builder.Get()

	result, err := transformation.Transform(core.NewTransformable(testMap))
	assert.Nil(t, err)

	switch v := result.Value().(type) {
	case map[string]interface{}:
		assert.Equal(t, testMap, v)
	}

	result, err = transformation.Transform(core.NewTransformable(testSlice))
	assert.Nil(t, err)

	switch v := result.Value().(type) {
	case []interface{}:
		assert.Equal(t, testSlice, v)
	}
}
