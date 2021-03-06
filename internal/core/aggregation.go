package core

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	gUuid "github.com/google/uuid"
	"github.com/kmgreen2/agglo/pkg/util"
	"math"
	"reflect"
	"strings"
)

type AggregationType int

const (
	AggSum AggregationType = iota
	AggMax
	AggMin
	AggAvg
	AggCount
	AggDiscreteHistogram
)

func (t AggregationType) String() string {
	switch t {
	case AggSum:
		return "AggSum"
	case AggMin:
		return "AggMin"
	case AggMax:
		return "AggMax"
	case AggAvg:
		return "AggAvg"
	case AggCount:
		return "AggCount"
	case AggDiscreteHistogram:
		return "AggDiscreteHistogram"
	}
	return "Unknown"
}

type FieldAggregationState interface {
	Update(val interface{})	error
	Get() interface{}
	ToMap() map[string]interface{}
}

type aggregationSumState struct {
	Value float64 `json:"value"`
}

func (s *aggregationSumState) Update(val interface{}) error {
	switch v := val.(type) {
	case float64:
		s.Value += v
		return nil
	}
	msg := fmt.Sprintf("expected float64 for AggregationSumState, got '%v'", reflect.TypeOf(val))
	return util.NewInvalidError(msg)
}

func (s *aggregationSumState) Get() interface{} {
	return s.Value
}

func (s *aggregationSumState) ToMap() map[string]interface{} {
	m := make(map[string]interface{})
	m["value"] = s.Value
	return m
}

func AggregationSumStateFromMap(in map[string]interface{}) (*aggregationSumState, error) {
	if value, ok := in["value"]; ok {
		if floatValue, floatOk := value.(float64); floatOk {
			return &aggregationSumState{floatValue}, nil
		} else {
			msg := fmt.Sprintf("invalid type for aggregation sum in map: %v", reflect.TypeOf(value))
			return nil, util.NewInvalidError(msg)
		}
	}
	return nil, util.NewInvalidError("could not find value for aggregation sum in map")
}

type aggregationMaxState struct {
	Value float64 `json:"value"`
}

func (s *aggregationMaxState) Update(val interface{}) error {
	switch v := val.(type) {
	case float64:
		if v > s.Value {
			s.Value = v
		}
		return nil
	}
	msg := fmt.Sprintf("expected float64 for AggregationMaxState, got '%v'", reflect.TypeOf(val))
	return util.NewInvalidError(msg)
}

func (s *aggregationMaxState) Get() interface{} {
	return s.Value
}

func (s *aggregationMaxState) ToMap() map[string]interface{} {
	m := make(map[string]interface{})
	m["value"] = s.Value
	return m
}

func AggregationMaxStateFromMap(in map[string]interface{}) (*aggregationMaxState, error) {
	if value, ok := in["value"]; ok {
		if floatValue, floatOk := value.(float64); floatOk {
			return &aggregationMaxState{floatValue}, nil
		} else {
			msg := fmt.Sprintf("invalid type for aggregation max in map: %v", reflect.TypeOf(value))
			return nil, util.NewInvalidError(msg)
		}
	}
	return nil, util.NewInvalidError("could not find value for aggregation max in map")
}

type aggregationMinState struct {
	Value float64 `json:"value"`
}

func (s *aggregationMinState) Update(val interface{}) error {
	switch v := val.(type) {
	case float64:
		if v < s.Value {
			s.Value = v
		}
		return nil
	}
	msg := fmt.Sprintf("expected float64 for AggregationMinState, got '%v'", reflect.TypeOf(val))
	return util.NewInvalidError(msg)
}

func (s *aggregationMinState) Get() interface{} {
	return s.Value
}

func (s *aggregationMinState) ToMap() map[string]interface{} {
	m := make(map[string]interface{})
	m["value"] = s.Value
	return m
}

func AggregationMinStateFromMap(in map[string]interface{}) (*aggregationMinState, error) {
	if value, ok := in["value"]; ok {
		if floatValue, floatOk := value.(float64); floatOk {
			return &aggregationMinState{floatValue}, nil
		} else {
			msg := fmt.Sprintf("invalid type for aggregation min in map: %v", reflect.TypeOf(value))
			return nil, util.NewInvalidError(msg)
		}
	}
	return nil, util.NewInvalidError("could not find value for aggregation min in map")
}

type aggregationAvgState struct {
	Sum float64 `json:"sum"`
	Num float64	`json:"num"`
}

func (s *aggregationAvgState) Update(val interface{}) error {
	switch v := val.(type) {
	case float64:
		s.Num++
		s.Sum += v
		return nil
	}
	msg := fmt.Sprintf("expected float64 for AggregationAvgState, got '%v'", reflect.TypeOf(val))
	return util.NewInvalidError(msg)
}

func (s *aggregationAvgState) Get() interface{} {
	return s.Sum / s.Num
}

func (s *aggregationAvgState) ToMap() map[string]interface{} {
	m := make(map[string]interface{})
	m["sum"] = s.Sum
	m["num"] = s.Num
	return m
}

func AggregationAvgStateFromMap(in map[string]interface{}) (*aggregationAvgState, error) {
	var sum, num float64
	if value, ok := in["sum"]; ok {
		if floatValue, floatOk := value.(float64); floatOk {
			sum = floatValue
		} else {
			msg := fmt.Sprintf("invalid type for aggregation avg ('sum') in map: %v", reflect.TypeOf(value))
			return nil, util.NewInvalidError(msg)
		}
	} else {
		return nil, util.NewInvalidError("could not find value for aggregation avg ('sum') in map")
	}
	if value, ok := in["num"]; ok {
		if floatValue, floatOk := value.(float64); floatOk {
			num = floatValue
		} else {
			msg := fmt.Sprintf("invalid type for aggregation avg ('sum') in map: %v", reflect.TypeOf(value))
			return nil, util.NewInvalidError(msg)
		}
	} else {
		return nil, util.NewInvalidError("could not find value for aggregation avg ('num') in map")
	}

	return &aggregationAvgState{sum, num}, nil
}

type aggregationCountState struct {
	Value int64	`json:"value"`
}

func (s *aggregationCountState) Update(val interface{}) error {
	s.Value++
	return nil
}

func (s *aggregationCountState) Get() interface{} {
	return s.Value
}

func (s *aggregationCountState) ToMap() map[string]interface{} {
	m := make(map[string]interface{})
	m["value"] = float64(s.Value)
	return m
}

func AggregationCountStateFromMap(in map[string]interface{}) (*aggregationCountState, error) {
	if value, ok := in["value"]; ok {
		if floatValue, floatOk := value.(float64); floatOk {
			return &aggregationCountState{int64(floatValue)}, nil
		} else {
			msg := fmt.Sprintf("invalid type for aggregation count in map: %v", reflect.TypeOf(value))
			return nil, util.NewInvalidError(msg)
		}
	}
	return nil, util.NewInvalidError("could not find value for aggregation count in map")
}

type aggregationDiscreteHistogramState struct {
	Buckets map[string]int `json:"buckets"`
}

func (s *aggregationDiscreteHistogramState) Update(val interface{}) error {
	switch v := val.(type) {
	case string:
		s.Buckets[v]++
		return nil
	case bool:
		s.Buckets[fmt.Sprintf("%v", v)]++
		return nil
	default:
		intVal, err := util.GetInteger(v)
		if err == nil {
			s.Buckets[fmt.Sprintf("%d", intVal)]++
			return nil
		}
	}
	msg := fmt.Sprintf("expected string, integer or bool for AggregationDiscreteHistogramState, got '%v'",
		reflect.TypeOf(val))
	return util.NewInvalidError(msg)
}

func (s *aggregationDiscreteHistogramState) Get() interface{} {
	return s.Buckets
}

func (s *aggregationDiscreteHistogramState) ToMap() map[string]interface{} {
	m := make(map[string]interface{})
	m["buckets"] = s.Buckets
	return m
}

func AggregationDiscreteHistogramStateFromMap(in map[string]interface{}) (*aggregationDiscreteHistogramState, error) {
	if value, ok := in["buckets"]; ok {
		// When this is first created it will be a map[string]int
		if mapIntValue, mapIntOk := value.(map[string]int); mapIntOk {
			return &aggregationDiscreteHistogramState{mapIntValue}, nil
		} else if mapValue, mapOk := value.(map[string]interface{}); mapOk {
			// When this is deserialized via JSON, it will be a map[string]interface{}
			intMap, err := util.MapInterfaceToInt(mapValue)
			if err != nil {
				return nil, err
			}
			return &aggregationDiscreteHistogramState{intMap}, nil
		} else {
			msg := fmt.Sprintf("invalid type for aggregation histogram in map: %v", reflect.TypeOf(value))
			return nil, util.NewInvalidError(msg)
		}
	}
	return nil, util.NewInvalidError("could not find value for aggregation count in map")
}

type FieldAggregation struct {
	Key         string          `json:"key"`
	Type        AggregationType `json:"type"`
	GroupByKeys []string        `json:"groupByKeys"`
}

func NewFieldAggregation(path string, aggType AggregationType, groupByKeys []string) *FieldAggregation {
	return &FieldAggregation{
		Key: path,
		Type: aggType,
		GroupByKeys: groupByKeys,
	}
}

func (fieldAggregation *FieldAggregation) getGroupByPath(in map[string]interface{}) []string {
	groupByPath := []string {fmt.Sprintf("%s:%s", fieldAggregation.Key, fieldAggregation.Type.String())}
	for _, groupByKey := range fieldAggregation.GroupByKeys {
		if v, ok := in[groupByKey]; ok {
			groupByPath = append(groupByPath, fmt.Sprintf("%v", v))
		} else {
			groupByPath = append(groupByPath, "nil")
		}
	}
	return groupByPath
}

type Aggregation struct {
	FieldAggregation *FieldAggregation `json:"fieldAggregation"`
}

func NewAggregation(fieldAggregation *FieldAggregation) *Aggregation {
	return &Aggregation{
		FieldAggregation: fieldAggregation,
	}
}

func (a Aggregation) Update(in map[string]interface{}, state *AggregationState) ([]string, []interface{}, error) {
	var fieldAggregationState FieldAggregationState
	var err error
	var updatedPaths []string
	var updatedValues []interface{}
	flattened := util.Flatten(in)

	if val, ok := flattened[a.FieldAggregation.Key]; ok {
		var currVal map[string]interface{}
		groupByPath := a.FieldAggregation.getGroupByPath(in)
		currVal, err = state.Get(groupByPath)
		if err != nil && errors.Is(err, &util.NotFoundError{}){
			err = state.Create(groupByPath, a.FieldAggregation.Type)
			if err != nil {
				return nil, nil, err
			}
			currVal, err = state.Get(groupByPath)
			if err != nil {
				return nil, nil, err
			}
		} else if err != nil {
			return nil, nil, err
		}
		switch a.FieldAggregation.Type {
		case AggCount:
			fieldAggregationState, err = AggregationCountStateFromMap(currVal)
		case AggSum:
			fieldAggregationState, err = AggregationSumStateFromMap(currVal)
		case AggAvg:
			fieldAggregationState, err = AggregationAvgStateFromMap(currVal)
		case AggMax:
			fieldAggregationState, err = AggregationMaxStateFromMap(currVal)
		case AggMin:
			fieldAggregationState, err = AggregationMinStateFromMap(currVal)
		case AggDiscreteHistogram:
			fieldAggregationState, err = AggregationDiscreteHistogramStateFromMap(currVal)
		default:
			err = util.NewInternalError(fmt.Sprintf("invalid aggregation type: %v", a.FieldAggregation.Type))
		}
		if err != nil {
			return nil, nil, err
		}
		err = fieldAggregationState.Update(val)
		if err != nil {
			return nil, nil, err
		}
		err = state.Update(groupByPath, fieldAggregationState)
		if err != nil {
			return nil, nil, err
		}
		updatedPaths = append(updatedPaths, strings.Join(groupByPath, "."))
		updatedValues = append(updatedValues, fieldAggregationState.Get())
	}
	return updatedPaths, updatedValues, nil
}

type AggregationState struct {
	// Mapping of groupBy values to the underlying value
	// Special 'nil' value used for aggregations where a groupBy key is not found
	Values map[string]interface{}	`json:"values"`
}

func (s AggregationState) Bytes() ([]byte, error) {
	byteBuffer := bytes.NewBuffer([]byte{})
	encoder := json.NewEncoder(byteBuffer)
	err := encoder.Encode(s)
	if err != nil {
		return nil, err
	}
	return byteBuffer.Bytes(), nil
}

func NewAggregationStateFromBytes(stateBytes []byte) (*AggregationState, error) {
	aggregationState := &AggregationState{}
	byteBuffer := bytes.NewBuffer(stateBytes)
	decoder := json.NewDecoder(byteBuffer)
	err := decoder.Decode(aggregationState)
	if err != nil {
		return nil, err
	}
	return aggregationState, nil
}

func NewAggregationState(values map[string]interface{}) *AggregationState {
	return &AggregationState{
		Values: values,
	}
}

func (s *AggregationState) Update(path []string, value interface{}) error {
	return util.UpdateMap(s.Values, path, value)
}

func (s *AggregationState) Create(path []string, aggType AggregationType) error {
	var value FieldAggregationState

	switch aggType {
	case AggCount:
		value = &aggregationCountState{}
	case AggSum:
		value = &aggregationSumState{}
	case AggAvg:
		value = &aggregationAvgState{}
	case AggMax:
		value = &aggregationMaxState{}
	case AggMin:
		value = &aggregationMinState{math.MaxFloat64}
	case AggDiscreteHistogram:
		value = &aggregationDiscreteHistogramState{make(map[string]int)}
	default:
		return util.NewInternalError(fmt.Sprintf("invalid aggregation type: %v", aggType))
	}

	return util.UpdateMap(s.Values, path, value.ToMap())
}

func (s *AggregationState) Get(path []string) (map[string]interface{}, error) {
	val, err := util.GetMap(s.Values, path)
	if err != nil {
		return nil, err
	}
	if mapVal, ok := val.(map[string]interface{}); ok {
		return mapVal, nil
	}
	if stateVal, ok := val.(FieldAggregationState); ok {
		return stateVal.ToMap(), nil
	}
	msg := fmt.Sprintf("value at path '%v' should be a map[string]interface, got %v", path, reflect.TypeOf(val))
	return nil, util.NewInvalidError(msg)
}

// <UUID>:<name>:a
var aggregationKeyFormat string = "%s:%s:a"
func AggregationKey(partitionID gUuid.UUID, name string) string {
	return fmt.Sprintf(aggregationKeyFormat, partitionID.String(), name)
}

// <UUID>:<name>:as
var aggregationStateKeyFormat string = "%s:%s:as"
func AggregationStateKey(partitionID gUuid.UUID, name string) string {
	return fmt.Sprintf(aggregationStateKeyFormat, partitionID.String(), name)
}

