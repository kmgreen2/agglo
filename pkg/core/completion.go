package core

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	gUuid "github.com/google/uuid"
	"github.com/kmgreen2/agglo/pkg/common"
	"github.com/kmgreen2/agglo/pkg/kvs"
	"reflect"
	"time"
)

type Completion struct {
	// Unique name for the completion
	// Source map must have field completion: "<name>" to be considered
	// This can be added by having an annotation process prior to the completion
	Name string				`json:"name"`
	PartitionID gUuid.UUID  `json:"partitionID"`
	JoinKeys []string		`json:"joinKeys"`
	Timeout time.Duration	`json:"timeout"`
}

func (c Completion) Bytes() ([]byte, error) {
	byteBuffer := bytes.NewBuffer([]byte{})
	encoder := json.NewEncoder(byteBuffer)
	err := encoder.Encode(c)
	if err != nil {
		return nil, err
	}
	return byteBuffer.Bytes(), nil
}

func NewCompletionFromBytes(completionBytes []byte) (*Completion, error) {
	completion := &Completion{}
	byteBuffer := bytes.NewBuffer(completionBytes)
	decoder := json.NewDecoder(byteBuffer)
	err := decoder.Decode(completion)
	if err != nil {
		return nil, err
	}
	return completion, nil
}

func NewCompletion(name string, partitionID gUuid.UUID, joinKeys []string, timeout time.Duration) *Completion {
	return &Completion{
		Name: name,
		PartitionID: partitionID,
		JoinKeys: joinKeys,
		Timeout: timeout,
	}
}

func (c Completion) Match(in map[string]interface{}) (string, interface{}, error) {
	flattened := Flatten(in)

	for _, key := range c.JoinKeys {
		if val, ok := flattened[key]; ok {
			return key, val, nil
		}
	}
	return "", nil, common.NewNotFoundError(fmt.Sprintf("completion key not found"))
}

type CompletionState struct {
	Value              interface{}     `json:"value"`
	Resolved           map[string]bool `json:"resolved"`
	CompletionDeadline int64           `json:"deadline"`
}

func (s CompletionState) IsDone() bool {
	for _, v := range s.Resolved {
		if v == false {
			return false
		}
	}
	return true
}

func (s CompletionState) Bytes() ([]byte, error) {
	byteBuffer := bytes.NewBuffer([]byte{})
	encoder := json.NewEncoder(byteBuffer)
	err := encoder.Encode(s)
	if err != nil {
		return nil, err
	}
	return byteBuffer.Bytes(), nil
}

func NewCompletionStateFromBytes(stateBytes []byte) (*CompletionState, error) {
	completionState := &CompletionState{}
	byteBuffer := bytes.NewBuffer(stateBytes)
	decoder := json.NewDecoder(byteBuffer)
	err := decoder.Decode(completionState)
	if err != nil {
		return nil, err
	}
	return completionState, nil
}

func NewCompletionState(value interface{}, resolved map[string]bool, completionDeadline int64) *CompletionState {
	return &CompletionState{
		Value:              value,
		Resolved:           resolved,
		CompletionDeadline: completionDeadline,
	}
}

// <UUID>:<name>:c
var completionKeyFormat string = "%s:%s:c"

func completionKey(partitionID gUuid.UUID, name string) string {
	return fmt.Sprintf(completionKeyFormat, partitionID.String(), name)
}

// <UUID>:<name>:c:matchingVal.String()
var completionStateKeyFormat string = "%s:%s:c:"

func completionStateKey(partitionID gUuid.UUID, name string, value interface{}) (string, error) {
	if stringValue, ok := value.(string); ok {
		return fmt.Sprintf(completionStateKeyFormat + "%s", partitionID.String(), name, stringValue), nil
	}
	if boolValue, ok := value.(bool); ok {
		return fmt.Sprintf(completionStateKeyFormat + "%v", partitionID.String(), name, boolValue), nil
	}
	if intValue, err := GetInteger(value); err == nil {
		return fmt.Sprintf(completionStateKeyFormat + "%d", partitionID.String(), name, intValue), nil
	}
	if numericValue, err := GetNumeric(value); err == nil {
		return fmt.Sprintf(completionStateKeyFormat + "%f", partitionID.String(), name, numericValue), nil
	}

	msg := fmt.Sprintf("completionStateKey must be a string, float, int or bool.  found: %v",
		reflect.TypeOf(value))
	return "", common.NewInvalidError(msg)
}

type Completer struct {
	completion *Completion
	completionStateStore kvs.KVStore
}

func NewCompleter(completion *Completion, kvStore kvs.KVStore) *Completer {
	return &Completer{
		completion: completion,
		completionStateStore: kvStore,
	}
}

func (c Completer) getCompletionState(partitionID gUuid.UUID, name string, value interface{}) ([]byte, error) {
	stateKey, err := completionStateKey(partitionID, name, value)
	if err != nil {
		return nil, err
	}

	stateBytes, err := c.completionStateStore.Get(context.Background(), stateKey)
	if err != nil {
		if errors.Is(err, &common.NotFoundError{}) {
			return nil, nil
		}
		return nil, err
	}
	return stateBytes, nil
}

func (c Completer) Process(in map[string]interface{}) (map[string]interface{}, error) {
	var completionState *CompletionState

	out := CopyableMap(in).DeepCopy()

	partitionID, err := GetPartitionID(in)
	if err != nil {
		return out, err
	}
	name, err := GetName(in)
	if err != nil {
		return out, err
	}

	matchedKey, matchedVal, err := c.completion.Match(in)
	if err != nil {
		if errors.Is(err, &common.NotFoundError{}) {
			return out, nil
		}
		return out, err
	}

	completionStateBytes, err := c.getCompletionState(partitionID, name, matchedVal)
	if err != nil {
		return out, err
	}

	if completionStateBytes != nil {
		completionState, err = NewCompletionStateFromBytes(completionStateBytes)
		if err != nil {
			return out, err
		}
	}

	if completionState == nil {
		var completionDeadline int64 = -1
		if c.completion.Timeout > 0 {
			completionDeadline = time.Now().Add(c.completion.Timeout).UnixNano()
		}
		resolved := make(map[string]bool)
		for _, key := range c.completion.JoinKeys {
			resolved[key] = false
		}
		completionState = NewCompletionState(matchedVal, resolved, completionDeadline)
	}

	completionState.Resolved[matchedKey] = true

	stateKey, err := completionStateKey(partitionID, name, matchedVal)
	if err != nil {
		return out, err
	}

	newCompletionBytes, err := completionState.Bytes()
	if err != nil {
		return out, err
	}

	if completionState.IsDone() {
		err = c.completionStateStore.AtomicDelete(context.Background(), stateKey, completionStateBytes)
		if err != nil {
			return out, err
		}
		out[fmt.Sprintf("agglo:completion:%s", name)] = "complete"
	} else if completionState.CompletionDeadline > 0 && time.Now().UnixNano() > completionState.CompletionDeadline {
			out[fmt.Sprintf("agglo:completion:%s", name)] = "timedout"
	} else {
		err = c.completionStateStore.AtomicPut(context.Background(), stateKey, completionStateBytes, newCompletionBytes)
		if err != nil {
			return out, err
		}
		out[fmt.Sprintf("agglo:completion:%s", name)] = "triggered"
	}

	return out, nil
}
