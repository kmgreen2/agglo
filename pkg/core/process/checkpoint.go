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

var CheckpointIdxKey string = "agglo:checkpoint:idx"
var CheckpointDataKey string = "agglo:checkpoint:data"

type Finalizer interface {
	Finalize(in map[string]interface{}) error
}

type CheckPointer struct {
	updateFunc func(key string, checkpointState, data map[string]interface{}) error
	fetchFunc func(key string) (map[string]interface{}, error)
	removeFunc func(key string) error
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
	updateFunc := func(key string, checkpointState, data map[string]interface{}) error {
		oldCheckpoint, newCheckpoint, err := updateSerializeCheckpoint(checkpointState, data)
		if err != nil {
			return err
		}
		return kvStore.AtomicPut(context.Background(), key, oldCheckpoint, newCheckpoint)
	}

	fetchFunc := func(key string) (map[string]interface{}, error) {
		var state map[string]interface{}
		stateBytes, err := kvStore.Get(context.Background(), key)
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

	removeFunc := func(key string) error {
		return kvStore.Delete(context.Background(), key)
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

	updateFunc := func(key string, checkpointState, data map[string]interface{}) error {
		_, newCheckpoint, err := updateSerializeCheckpoint(checkpointState, data)
		if err != nil {
			return err
		}
		return ioutil.WriteFile(fmt.Sprintf("%s/%s.json", path, key), newCheckpoint, 0644)
	}

	fetchFunc := func(key string) (map[string]interface{}, error) {
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

	removeFunc := func(key string) error {
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

func (c CheckPointer) Process(in map[string]interface{}) (map[string]interface{}, error) {
	if pipelineName, nameOk := in["agglo:internal:name"]; nameOk {
		if messageID, messageOk := in["agglo:messageID"]; messageOk {
			checkpointStateKey := fmt.Sprintf("%s:%s", pipelineName, messageID)
			checkpoint, err := c.fetchFunc(checkpointStateKey)
			if err != nil {
				return nil, err
			}
			err = c.updateFunc(checkpointStateKey, checkpoint, in)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, common.NewInternalError("could not find 'agglo:messageID' to checkpoint")
		}
	} else {
		return nil, common.NewInternalError("could not find 'agglo:internal:name' to checkpoint")
	}
	return in, nil
}

func (c CheckPointer) Finalize(in map[string]interface{}) error {
	if pipelineName, nameOk := in["agglo:internal:name"]; nameOk {
		if messageID, messageOk := in["agglo:messageID"]; messageOk {
			checkpointStateKey := fmt.Sprintf("%s:%s", pipelineName, messageID)
			return c.removeFunc(checkpointStateKey)
		} else {
			return common.NewInternalError("could not find 'agglo:messageID' to remove checkpoint")
		}
	} else {
		return common.NewInternalError("could not find 'agglo:internal:name' to remove checkpoint")
	}
}
