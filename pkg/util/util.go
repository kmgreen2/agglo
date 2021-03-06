package util

import (
	"bytes"
	"context"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/json"
	api "github.com/kmgreen2/agglo/generated/proto"
	"github.com/pkg/errors"
	"hash"
	"strings"
	"sync"
	"time"
	"fmt"
	"reflect"
	"math"
)

// Used to distinguish between different Digest algorithms
type DigestType int

const (
	SHA1 DigestType = iota
	SHA256
	MD5
)

// Construct a hash object using a supported Digest
// type.  If the Digest type is not supported, return
// nil.
func InitHash(digestType DigestType) hash.Hash {
	switch digestType {
	case SHA1:
		return sha1.New()
	case SHA256:
		return sha256.New()
	case MD5:
		return md5.New()
	default:
		return nil
	}
}

func DigestTypeToPb(digestType DigestType) api.DigestType {
	switch digestType {
	case SHA1:
		return api.DigestType_SHA1
	case SHA256:
		return api.DigestType_SHA256
	case MD5:
		return api.DigestType_MD5
	default:
		// Shouldn't ever reach this
		return 0
	}
}

func DigestTypeFromPb(digestType api.DigestType) DigestType {
	switch digestType {
	case api.DigestType_SHA1:
		return SHA1
	case api.DigestType_MD5:
		return MD5
	case api.DigestType_SHA256:
		return SHA256
	default:
		// Shouldn't ever reach this
		return 0
	}
}

// ToDo(KMG): Re-visit this function.  I could not think of a way to
// use WaitGroups without leaking a go routine when the Wait() call
// hangs forever when we set a timeout.  The best I could think of was
// to track the number of waiters and decrement the count when we timeout.
//
// This can probably done with atomic incr/decr and channels
func WaitAll(futures []Future, timeout time.Duration) {
	lock := &sync.Mutex{}
	numFutures := len(futures)
	done := make(chan bool, 1)
	wg := &sync.WaitGroup{}
	wg.Add(numFutures)

	go func() {
		var ctx context.Context
		var cancel context.CancelFunc

		ctx = context.Background()
		if timeout > 0 {
			ctx, cancel = context.WithTimeout(ctx, timeout)
			defer cancel()
		}

		select {
		case <-ctx.Done():
			lock.Lock()
			defer lock.Unlock()
			if numFutures == 0 {
				return
			}
			wg.Add(-numFutures)
		case <-done:
			return
		}
	}()

	for _, future := range futures {
		future.OnSuccess(func(ctx context.Context, x interface{}) {
			lock.Lock()
			defer lock.Unlock()
			numFutures--
			wg.Done()
		}).OnFail(func(ctx context.Context, err error) {
			lock.Lock()
			defer lock.Unlock()
			numFutures--
			wg.Done()
		}).OnCancel(func(ctx context.Context) {
			lock.Lock()
			defer lock.Unlock()
			numFutures--
			wg.Done()
		})
	}

	wg.Wait()
	done <- true
}

func MapToJson(in map[string]interface{}) ([]byte, error) {
	byteBuffer := bytes.NewBuffer([]byte{})
	encoder := json.NewEncoder(byteBuffer)
	err := encoder.Encode(in)
	if err != nil {
		return nil, err
	}
	return byteBuffer.Bytes(), nil
}

func JsonToMap(in []byte) (map[string]interface{}, error) {
	var out map[string]interface{}
	byteBuffer := bytes.NewBuffer(in)
	decoder := json.NewDecoder(byteBuffer)
	err := decoder.Decode(&out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

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

func NumericResolver(x, y interface{}) (float64, float64, error) {
	if xVal, err := GetNumeric(x); err != nil {
		return 0, 0, err
	} else if yVal, err := GetNumeric(y); err != nil {
		return 0, 0, err
	} else {
		return xVal, yVal, nil
	}
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
			if _, ok := inVal[path[0]]; !ok {
				inVal[path[0]] = make(map[string]interface{})
			}
			return updateMap(inVal[path[0]], path[1:], value)
		}
	}
	return &InvalidError{}
}

func UpdateMap(in map[string]interface{}, path []string, value interface{}) error {
	err := updateMap(in, path, value)

	if err != nil && errors.Is(err, &NotFoundError{}) {
		msg := fmt.Sprintf("path '%v' not found in map", path)
		return NewNotFoundError(msg)
	} else if err != nil {
		msg := fmt.Sprintf("intermediate value in path '%v' resolves to non-map value in map", path)
		return NewInvalidError(msg)
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
			return nil, &NotFoundError{}
		} else if len(path) > 1 {
			if v, ok := inVal[path[0]]; ok {
				return getMap(v, path[1:])
			} else {
				return nil, &NotFoundError{}
			}
		}
	}
	return nil, &InvalidError{}
}

func GetMap(in map[string]interface{}, path []string) (interface{}, error) {
	val, err := getMap(in, path)

	if err != nil && errors.Is(err, &NotFoundError{}) {
		msg := fmt.Sprintf("path '%v' not found in map", path)
		return nil, NewNotFoundError(msg)
	} else if err != nil {
		msg := fmt.Sprintf("intermediate value in path '%v' resolves to non-map value in map", path)
		return nil, NewInvalidError(msg)
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
			return nil, NewInvalidError(msg)
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

func MergeMaps(lhs, rhs map[string]interface{}) (map[string]interface{}, error) {
	var err error
	newMap := CopyableMap(lhs).DeepCopy()
	for k, v := range rhs {
		if _, ok := newMap[k]; !ok {
			newMap[k] = v
			continue
		}
		switch _v := v.(type) {
		case map[string]interface{}:
			if _newMap, ok := newMap[k].(map[string]interface{}); ok {
				newMap[k], err = MergeMaps(_newMap, _v)
				if err != nil {
					return nil, err
				}
			} else {
				msg := fmt.Sprintf("key conflict (%s): %v != %v", k, reflect.TypeOf(_newMap), reflect.TypeOf(_v))
				return nil, NewConflictError(msg)
			}
		case []interface{}:
			if _newSlice, ok := newMap[k].([]interface{}); ok {
				newMap[k] = append(_newSlice, _v...)
			} else {
				msg := fmt.Sprintf("key conflict (%s): %v != %v", k, reflect.TypeOf(_newSlice), reflect.TypeOf(_v))
				return nil, NewConflictError(msg)
			}
		default:
			msg := fmt.Sprintf("key conflict (%s)", k)
			return nil, NewConflictError(msg)
		}
	}
	return newMap, nil
}

