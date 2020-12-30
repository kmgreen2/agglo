package core

import (
	"fmt"
	"github.com/kmgreen2/agglo/pkg/common"
	"github.com/pkg/errors"
	"math"
	"reflect"
	"strings"
)

func GetInteger(x interface{}) (int64, error) {
	switch x := x.(type) {
	case uint8:
		return int64(x), nil
	case int8:
		return int64(x), nil
	case uint16:
		return int64(x), nil
	case int16:
		return int64(x), nil
	case uint32:
		return int64(x), nil
	case int32:
		return int64(x), nil
	case uint64:
		if x > math.MaxInt64 {
			return 0, fmt.Errorf("Integer overflow: %d does not fit into int64", x)
		}
		return int64(x), nil
	case int64:
		return x, nil
	case int:
		return int64(x), nil
	}
	return 0, fmt.Errorf("Invalid integer type: %v", reflect.TypeOf(x))
}

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
		return x, nil
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

func updateMap(in interface{}, path []string, value interface{}) error {
	switch inVal := in.(type) {
	case map[string]interface{}:
		if len(path) == 1 {
			inVal[path[0]] = value
			return nil
		} else if len(path) > 1 {
			if v, ok := inVal[path[0]]; ok {
				return updateMap(v, path[1:], value)
			} else {
				return &common.NotFoundError{}
			}
		}
	}
	return &common.InvalidError{}
}

func UpdateMap(in map[string]interface{}, path []string, value interface{}) error {
	err := updateMap(in, path, value)

	if err != nil && errors.Is(err, &common.NotFoundError{}) {
		msg := fmt.Sprintf("path '%v' not found in map", path)
		return common.NewNotFoundError(msg)
	} else if err != nil {
		msg := fmt.Sprintf("intermediate value in path '%v' resolves to non-map value in map", path)
		return common.NewInvalidError(msg)
	}
	return nil
}

func getMap(in interface{}, path []string) (interface{}, error) {
	switch inVal := in.(type) {
	case map[string]interface{}:
		if len(path) == 1 {
			if v, ok := inVal[path[0]]; ok {
				return v, nil
			}
			return nil, &common.NotFoundError{}
		} else if len(path) > 1 {
			if v, ok := inVal[path[0]]; ok {
				return getMap(v, path[1:])
			} else {
				return nil, &common.NotFoundError{}
			}
		}
	}
	return nil, &common.InvalidError{}
}

func GetMap(in map[string]interface{}, path []string) (interface{}, error) {
	val, err := getMap(in, path)

	if err != nil && errors.Is(err, &common.NotFoundError{}) {
		msg := fmt.Sprintf("path '%v' not found in map", path)
		return nil, common.NewNotFoundError(msg)
	} else if err != nil {
		msg := fmt.Sprintf("intermediate value in path '%v' resolves to non-map value in map", path)
		return nil, common.NewInvalidError(msg)
	}
	return val, nil
}

func NumericEqual(lhs, rhs interface{}) bool {
	lhsInt, intLhsErr := GetInteger(lhs)
	rhsInt, intRhsErr := GetInteger(rhs)

	if intLhsErr != nil && intRhsErr != nil {
		lhsFloat, floatLhsErr := GetNumeric(lhs)
		rhsFloat, floatRhsErr := GetNumeric(rhs)
		if floatLhsErr != nil || floatRhsErr != nil {
			return false
		}
		if lhsFloat != rhsFloat {
			return false
		}
	} else if (intLhsErr != nil || intRhsErr != nil) {
		return false
	} else if lhsInt != rhsInt {
		return false
	}
	return true
}

func MapInterfaceToInt(in map[string]interface{}) (map[string]int, error) {
	newMap := make(map[string]int)
	for k, v := range in {
		if numericV, err := GetNumeric(v); err == nil {
			newMap[k] = int(numericV)
		} else {
			msg := fmt.Sprintf("cannot map value of type %v to int", reflect.TypeOf(v))
			return nil, common.NewInvalidError(msg)
		}
	}
	return newMap, nil
}

type CopyableMap map[string]interface{}
type CopyableSlice []interface{}

func (m CopyableMap) DeepCompare(in map[string]interface{}) bool {
	if len(m) != len(in) {
		return false
	}
	for k, v := range m {
		if _, ok := in[k]; !ok {
			return false
		}
		switch _v := v.(type) {
		case map[string]interface{}:
			if otherV, okKey := in[k]; !okKey {
				return false
			} else if vMap, okType := otherV.(map[string]interface{}); !okType {
				return false
			} else {
				if !CopyableMap(_v).DeepCompare(vMap) {
					return false
				}
			}
		case []interface{}:
			if vSlice, ok := in[k].([]interface{}); !ok {
				return false
			} else if len(vSlice) != len(_v) {
				return false
			}  else {
				if !CopyableSlice(_v).DeepCompare(vSlice) {
					return false
				}
			}
		case string:
			if vString, ok := in[k].(string); !ok {
				return false
			} else if strings.Compare(vString, _v) != 0 {
				return false
			}
		case bool:
			if _v != in[k] {
				return false
			}
		default:
			if !NumericEqual(_v, in[k]) {
				return false
			}
		}
	}
	return true
}

func (m CopyableMap) DeepCopy() map[string]interface{} {
	newMap := make(map[string]interface{})
	for k, v := range m {
		switch _v := v.(type) {
		case map[string]interface{}:
			newMap[k] = CopyableMap(_v).DeepCopy()
		case []interface{}:
			newMap[k] = CopyableSlice(_v).DeepCopy()
		default:
			newMap[k] = v
		}
	}
	return newMap
}

func (m CopyableSlice) DeepCompare(in []interface{}) bool {
	if len(m) != len(in) {
		return false
	}
	for i, v := range m {
		if i >= len(in) {
			return false
		}
		switch _v := v.(type) {
		case map[string]interface{}:
			if vMap, okType := in[i].(map[string]interface{}); !okType {
				return false
			} else {
				if !CopyableMap(_v).DeepCompare(vMap) {
					return false
				}
			}
		case []interface{}:
			if vSlice, ok := in[i].([]interface{}); !ok {
				return false
			} else if len(vSlice) != len(_v) {
				return false
			}  else {
				if !CopyableSlice(_v).DeepCompare(vSlice) {
					return false
				}
			}
		case string:
			if vString, ok := in[i].(string); !ok {
				return false
			} else if strings.Compare(vString, _v) != 0 {
				return false
			}
		case bool:
			if _v != in[i] {
				return false
			}
		default:
			if !NumericEqual(_v, in[i]) {
				return false
			}
		}
	}
	return true
}

func (m CopyableSlice) DeepCopy() []interface{} {
	newSlice := make([]interface{}, len(m))
	for i, v := range m {
		switch _v := v.(type) {
		case map[string]interface{}:
			newSlice[i] = CopyableMap(_v).DeepCopy()
		case []interface{}:
			newSlice[i] = CopyableSlice(_v).DeepCopy()
		default:
			newSlice[i] = v
		}
	}
	return newSlice
}


