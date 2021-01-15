package process

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/kmgreen2/agglo/pkg/common"
	"github.com/kmgreen2/agglo/pkg/core"
	"github.com/kmgreen2/agglo/pkg/kvs"
	"io/ioutil"
	"os"
	"reflect"
)

var CheckpointIdxKey string = string(common.CheckpointIndexKey)
var CheckpointDataKey string = string(common.CheckpointDataKey)

type Finalizer interface {
	Finalize(ctx context.Context, in map[string]interface{}) error
}

type CheckPointer struct {
	updateFunc func(ctx context.Context, key string, checkpointState, data map[string]interface{}) error
	fetchFunc func(ctx context.Context, key string) (map[string]interface{}, error)
	removeFunc func(ctx context.Context, key string) error
	outputType string
	connectionString string
}

func updateCheckpoint(checkpointState, data map[string]interface{}) (map[string]interface{}, error) {
	checkpointStateOut := core.CopyableMap(checkpointState).DeepCopy()
	if _, ok := checkpointStateOut[CheckpointIdxKey]; !ok {
		checkpointStateOut[CheckpointIdxKey] = 0;
	} else {
		if checkPointIdx, ok := checkpointStateOut[CheckpointIdxKey].(float64); ok {
			checkpointStateOut[CheckpointIdxKey] = checkPointIdx + 1
		} else {
			msg := fmt.Sprintf("checkpoint state is incorrect, expected int, got %v",
				reflect.TypeOf(checkpointStateOut[CheckpointIdxKey]))
			return nil, common.NewInvalidError(msg)
		}
	}
	checkpointStateOut[CheckpointDataKey] = data;
	return checkpointStateOut, nil
}

func updateSerializeCheckpoint(checkpointState, data map[string]interface{}) (old, new []byte, err error) {
	newCheckpointState, err := updateCheckpoint(checkpointState, data)
	if err != nil {
		return
	}
	if checkpointState != nil {
		old, err = common.MapToJson(checkpointState)
		if err != nil {
			return
		}
	}
	new, err = common.MapToJson(newCheckpointState)
	if err != nil {
		return
	}
	return
}

func NewKVCheckPointer(kvStore kvs.KVStore) *CheckPointer {
	updateFunc := func(ctx context.Context, key string, checkpointState, data map[string]interface{}) error {
		oldCheckpoint, newCheckpoint, err := updateSerializeCheckpoint(checkpointState, data)
		if err != nil {
			return err
		}
		return kvStore.AtomicPut(ctx, key, oldCheckpoint, newCheckpoint)
	}

	fetchFunc := func(ctx context.Context, key string) (map[string]interface{}, error) {
		var state map[string]interface{}
		stateBytes, err := kvStore.Get(ctx, key)
		if err != nil {
			// If not found, just return nil with no error
			if errors.Is(err, &common.NotFoundError{}) {
				return nil, nil
			}
			return nil, err
		}
		stateBuffer := bytes.NewBuffer(stateBytes)
		decoder := json.NewDecoder(stateBuffer)
		err = decoder.Decode(&state)
		if err != nil {
			return nil, err
		}
		return state, nil
	}

	removeFunc := func(ctx context.Context, key string) error {
		return kvStore.Delete(ctx, key)
	}

	return &CheckPointer{
		updateFunc: updateFunc,
		fetchFunc: fetchFunc,
		removeFunc: removeFunc,
		outputType: "kvstore",
		connectionString: "memKVStore",
	}
}

func NewLocalFileCheckPointer(path string) (*CheckPointer, error) {
	if d, err := os.Stat(path); err != nil || !d.IsDir() {
		msg := fmt.Sprintf("'%s is not a valid path", path)
		return nil, common.NewInvalidError(msg)
	}

	updateFunc := func(ctx context.Context, key string, checkpointState, data map[string]interface{}) error {
		_, newCheckpoint, err := updateSerializeCheckpoint(checkpointState, data)
		if err != nil {
			return err
		}
		return ioutil.WriteFile(fmt.Sprintf("%s/%s.json", path, key), newCheckpoint, 0644)
	}

	fetchFunc := func(ctx context.Context, key string) (map[string]interface{}, error) {
		var state map[string]interface{}
		stateBytes, err := ioutil.ReadFile(fmt.Sprintf("%s/%s.json", path, key))
		if err != nil {
			// If not found, return nil with no error
			switch err.(type) {
			case *os.PathError:
				return nil, nil
			default:
				return nil, err
			}
		}
		stateBuffer := bytes.NewBuffer(stateBytes)
		decoder := json.NewDecoder(stateBuffer)
		err = decoder.Decode(&state)
		if err != nil {
			return nil, err
		}
		return state, nil
	}

	removeFunc := func(ctx context.Context, key string) error {
		return os.Remove(fmt.Sprintf("%s/%s.json", path, key))
	}

	return &CheckPointer{
		updateFunc: updateFunc,
		fetchFunc: fetchFunc,
		removeFunc: removeFunc,
		outputType: "localfile",
		connectionString: path,
	}, nil
}

func getIndexFromCheckpoint(checkpoint map[string]interface{}) (int64, error) {
	if idx, ok := common.GetFromInternalKey(common.CheckpointIndexKey, checkpoint); ok {
		numericIdx, err := core.GetNumeric(idx)
		if err != nil {
			return -1, err
		}
		return int64(numericIdx), nil
	}
	return -1, common.NewNotFoundError("could not find checkpoint index")
}

func getDataFromCheckpoint(checkpoint map[string]interface{}) (map[string]interface{}, error) {
	if rawData, rawOk := common.GetFromInternalKey(common.CheckpointDataKey, checkpoint); rawOk {
		if data, ok := rawData.(map[string]interface{}); ok {
			return data, nil
		} else {
			msg := fmt.Sprintf("invalid checkpoint data, expected map[string]interface{}, got %v",
				reflect.TypeOf(rawData))
			return nil, common.NewNotFoundError(msg)
		}
	}
	return nil, common.NewNotFoundError("could not find checkpoint data")
}

func (c CheckPointer) GetCheckpoint(ctx context.Context, pipelineName, messageID string) (out map[string]interface{},
	err error) {
	var checkpoint map[string]interface{}
	checkpointStateKey := fmt.Sprintf("%s:%s", pipelineName, messageID)
	checkpoint, err = c.fetchFunc(ctx, checkpointStateKey)
	if err != nil {
				return
			}
	// Fetch returns nil.nil if no checkpoint is found
	if checkpoint == nil {
		out = nil
		err = common.NewNotFoundError("no checkpoint")
		return
	}
	out, err = getDataFromCheckpoint(checkpoint)
	if err != nil {
		return
	}
	return
}

func (c CheckPointer) GetCheckpointWithIndexFromMap(ctx context.Context, in map[string]interface{}) (out map[string]interface{},
	index int64, err error) {

	// If the internal fields are not found in the map, then we assume there
	// is no checkpoint
	err = common.NewNotFoundError("no checkpoint")

	if pipelineName, nameOk := common.GetFromInternalKey(common.ResourceNameKey, in); nameOk {
		if messageID, messageOk := common.GetFromInternalKey(common.MessageIDKey, in); messageOk {
			var checkpoint map[string]interface{}
			checkpointStateKey := fmt.Sprintf("%s:%s", pipelineName, messageID)
			checkpoint, err = c.fetchFunc(ctx, checkpointStateKey)
			if err != nil {
				return
			}
			// Fetch returns nil.nil if no checkpoint is found
			if checkpoint == nil {
				out = nil
				index = -1
				err = common.NewNotFoundError("no checkpoint")
				return
			}
			out, err = getDataFromCheckpoint(checkpoint)
			if err != nil {
				return
			}
			index, err = getIndexFromCheckpoint(checkpoint)
			if err != nil {
				return
			}
		}
	}
	return
}

func (c CheckPointer) Process(ctx context.Context, in map[string]interface{}) (map[string]interface{}, error) {
	if pipelineName, nameOk := common.GetFromInternalKey(common.ResourceNameKey, in); nameOk {
		if messageID, messageOk := common.GetFromInternalKey(common.MessageIDKey, in); messageOk {
			checkpointStateKey := fmt.Sprintf("%s:%s", pipelineName, messageID)
			checkpoint, err := c.fetchFunc(ctx, checkpointStateKey)
			if err != nil {
				return nil, err
			}
			err = c.updateFunc(ctx, checkpointStateKey, checkpoint, in)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, common.NewInternalError("could not find messageID to checkpoint")
		}
	} else {
		return nil, common.NewInternalError("could not find resource name to checkpoint")
	}
	return in, nil
}

func (c CheckPointer) Finalize(ctx context.Context, in map[string]interface{}) error {
	if pipelineName, nameOk := common.GetFromInternalKey(common.ResourceNameKey, in); nameOk {
		if messageID, messageOk := common.GetFromInternalKey(common.MessageIDKey, in); messageOk {
			checkpointStateKey := fmt.Sprintf("%s:%s", pipelineName, messageID)
			return c.removeFunc(ctx, checkpointStateKey)
		} else {
			return common.NewInternalError("could not find messageID to remove checkpoint")
		}
	} else {
		return common.NewInternalError("could not find resource name to remove checkpoint")
	}
}
