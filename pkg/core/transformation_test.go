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

func TestFoldMin(t *testing.T) {
}

func TestFoldMax(t *testing.T) {
}

func TestFoldCount(t *testing.T) {
}

func TestSumTransformation(t *testing.T) {
}

func TestCopyTransformation(t *testing.T) {
}
