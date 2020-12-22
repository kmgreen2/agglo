package core

import (
	"fmt"
	"reflect"
)

func GetNumeric(x interface{}) (float64, error) {
	switch x := x.(type) {
	case uint8:
		return float64(x), nil
	case int8:
		return float64(x), nil
	case uint16:
		return float64(x), nil
	case int16:
		return float64(x), nil
	case uint32:
		return float64(x), nil
	case int32:
		return float64(x), nil
	case uint64:
		return float64(x), nil
	case int64:
		return float64(x), nil
	case int:
		return float64(x), nil
	case float32:
		return float64(x), nil
	case float64:
		return float64(x), nil
	}
	return 0, fmt.Errorf("Invalid numeric type: %v", reflect.TypeOf(x))
}

func flatten(in interface{}, out map[string]interface{}, currKey string) {
	var thisKey string
	switch inVal := in.(type) {
	case map[string]interface{}:
		for k, _ := range inVal {
			if len(currKey) == 0 {
				thisKey = k
			} else {
				thisKey = fmt.Sprintf("%s.%s", currKey, k)
			}
			flatten(inVal[k], out, thisKey)
		}
	case []interface{}:
		for i, v := range inVal {
			if len(currKey) == 0 {
				thisKey = fmt.Sprintf("%d", i)
			} else {
				thisKey = fmt.Sprintf("%s.%d", currKey, i)
			}
			flatten(v, out, thisKey)
		}
	default:
		out[currKey] = in
	}
}

func Flatten(in map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{})

	flatten(in, out, "")
	return out
}

