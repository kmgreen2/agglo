package core_test

import (
	"github.com/kmgreen2/agglo/pkg/core"
	"github.com/kmgreen2/agglo/test"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetInteger(t *testing.T) {
	vals := []interface{}{uint8(17), int8(17), uint16(17), int16(17), uint32(17), int32(17), uint64(17),
		int64(17), int(17)}

	for _, val := range vals {
		result, err := core.GetInteger(val)
		assert.Nil(t, err)
		assert.Equal(t, int64(17), result)
	}

	result, err := core.GetInteger(float64(5))
	assert.Error(t, err)
	assert.Equal(t, int64(0), result)

	result, err = core.GetInteger("hi")
	assert.Error(t, err)
	assert.Equal(t, int64(0), result)
}

func TestGetNumeric(t *testing.T) {
	vals := []interface{}{uint8(17), int8(17), uint16(17), int16(17), uint32(17), int32(17), uint64(17),
		int64(17), int(17), float32(17), float64(17)}

	for _, val := range vals {
		result, err := core.GetNumeric(val)
		assert.Nil(t, err)
		assert.Equal(t, float64(17), result)
	}

	result, err := core.GetInteger("hi")
	assert.Error(t, err)
	assert.Equal(t, int64(0), result)
}

func TestFlatten(t *testing.T) {
	jsonMap := test.TestJson()

	expectedFlatJson := map[string]interface{}{
		"a": float64(1),
		"b.c": "hello",
		"b.d.0": float64(3),
		"b.d.1": float64(4),
		"b.d.2": float64(5),
		"e.0": float64(6),
		"f.g.h": float64(7),
		"i.0.j.0": float64(8),
		"i.0.j.1": float64(9),
		"i.1": "k",
	}

	flatJson := core.Flatten(jsonMap)

	assert.Equal(t, len(expectedFlatJson), len(flatJson))

	for k, v := range expectedFlatJson {
		if _, ok := flatJson[k]; !ok {
			assert.True(t, ok)
		}
		switch _v := v.(type) {
		case string:
			if otherV, ok := flatJson[k].(string); ok {
				assert.Equal(t, _v, otherV)
			} else {
				assert.True(t, ok)
			}
		default:
			assert.Equal(t, _v, flatJson[k])
		}
	}
}


func TestNumericEqual(t *testing.T) {
	assert.True(t, core.NumericEqual(int8(5), int16(5)))
	assert.False(t, core.NumericEqual(int8(5), int16(7)))
	assert.True(t, core.NumericEqual(float32(5), float64(5)))
	assert.False(t, core.NumericEqual(int32(5), float64(5)))
	assert.False(t, core.NumericEqual(int64(5), ""))
}

func TestDeepCopy(t *testing.T) {
	jsonMap := test.TestJson()

	jsonMapCopy := core.CopyableMap(jsonMap).DeepCopy()
	assert.True(t, core.CopyableMap(jsonMap).DeepCompare(jsonMapCopy))
	jsonMapCopy["fizz"] = true
	assert.False(t, core.CopyableMap(jsonMap).DeepCompare(jsonMapCopy))
	jsonMapCopy = core.CopyableMap(jsonMap).DeepCopy()
	assert.True(t, core.CopyableMap(jsonMap).DeepCompare(jsonMapCopy))
	jsonMapCopy["b"].(map[string]interface{})["d"].([]interface{})[0] = 2
	assert.False(t, core.CopyableMap(jsonMap).DeepCompare(jsonMapCopy))
}

func TestUpdateMap(t *testing.T) {
	jsonMap := test.TestJson()

	err := core.UpdateMap(jsonMap, []string{"b", "c"}, "hi")
	assert.Nil(t, err)
	assert.Equal(t, "hi", jsonMap["b"].(map[string]interface{})["c"])

	err = core.UpdateMap(jsonMap, []string{"b", "d"}, []int{1,2})
	assert.Nil(t, err)
	assert.Equal(t, []int{1,2}, jsonMap["b"].(map[string]interface{})["d"])

	err = core.UpdateMap(jsonMap, []string{"b", "e", "f"}, []int{1,2})
	assert.Nil(t, err)
	assert.Equal(t, []int{1,2}, jsonMap["b"].(map[string]interface{})["e"].(map[string]interface{})["f"])

	err = core.UpdateMap(jsonMap, []string{"i", "0", "j"}, []int{1,2})
	assert.Error(t, err)
}

func TestGetMap(t *testing.T) {
	jsonMap := test.TestJson()

	val, err := core.GetMap(jsonMap, []string{"b", "c"})
	assert.Nil(t, err)
	assert.Equal(t, "hello", val)

	val, err = core.GetMap(jsonMap, []string{"b", "d"})
	assert.Nil(t, err)
	assert.Equal(t, []interface{}{float64(3),float64(4),float64(5)}, val)

	_, err = core.GetMap(jsonMap, []string{"b", "e", "f"})
	assert.Error(t, err)

	_, err = core.GetMap(jsonMap, []string{"i", "0", "j"})
	assert.Error(t, err)
}
