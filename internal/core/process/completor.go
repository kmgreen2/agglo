package process

import (
	"context"
	"errors"
	gUuid "github.com/google/uuid"
	"github.com/kmgreen2/agglo/internal/common"
	"github.com/kmgreen2/agglo/pkg/state"
	"github.com/kmgreen2/agglo/pkg/util"
	"github.com/kmgreen2/agglo/internal/core"
	"github.com/kmgreen2/agglo/pkg/kvs"
	"time"
)

type Completer struct {
	name string
	completion *core.Completion
	completionStateStore state.StateStore
}

func NewCompleter(name string, completion *core.Completion, kvStore kvs.KVStore) *Completer {
	return &Completer{
		name: name,
		completion: completion,
		completionStateStore: state.NewKvStateStore(kvStore),
	}
}

func (c Completer) getCompletionState(ctx context.Context, partitionID gUuid.UUID, name string,
	value interface{}) ([]byte, error) {
	stateKey, err := core.CompletionStateKey(partitionID, name, value)
	if err != nil {
		return nil, err
	}

	stateBytes, err := c.completionStateStore.Get(ctx, stateKey)
	if err != nil {
		if errors.Is(err, &util.NotFoundError{}) {
			return nil, nil
		}
		return nil, err
	}
	return stateBytes, nil
}

func (c Completer) checkpointMapFunc() func(curr, val []byte) ([]byte, error) {
	return func(curr, val []byte) ([]byte, error) {
		var completionState *core.CompletionState
		var err error
		if curr != nil {
			completionState, err = core.NewCompletionStateFromBytes(curr)
			if err != nil {
				return nil, err
			}
		}

		valMap, err := util.JsonToMap(val)
		if err != nil {
			return nil, err
		}

		matchedKey, matchedVal, err := c.completion.Match(valMap)
		if err != nil && !errors.Is(err, &util.NotFoundError{}){
			return nil, err
		} else if err != nil && completionState == nil {
			return nil, nil
		} else if err != nil {
			return completionState.Bytes()
		}

		if curr == nil {
			var completionDeadline int64 = -1
			if c.completion.Timeout > 0 {
				completionDeadline = time.Now().Add(c.completion.Timeout).UnixNano()
			}
			resolved := make(map[string]bool)
			for _, key := range c.completion.JoinKeys {
				resolved[key] = false
			}
			completionState = core.NewCompletionState(matchedVal, resolved, completionDeadline)
		}

		completionState.Resolved[matchedKey] = true

		return completionState.Bytes()
	}
}

func (c Completer) Process(ctx context.Context, in map[string]interface{}) (map[string]interface{}, error) {
	var completionState *core.CompletionState

	out := util.CopyableMap(in).DeepCopy()

	partitionID, err := core.GetPartitionID(in)
	if err != nil {
		return out, PipelineProcessError(c, err, "getting parition ID")
	}
	name := c.name

	_, matchedVal, err := c.completion.Match(in)
	if err != nil {
		if errors.Is(err, &util.NotFoundError{}) {
			return out, nil
		}
		return out, PipelineProcessError(c, err, "matching completion keys")
	}

	stateKey, err := core.CompletionStateKey(partitionID, name, matchedVal)
	if err != nil {
		return out, PipelineProcessError(c, err, "getting state key keys")
	}

	mapBytes, err := util.MapToJson(in)
	if err != nil {
		return out, PipelineProcessError(c, err, "serializing in map to bytes")
	}

	err = c.completionStateStore.Append(ctx, stateKey, mapBytes)
	if err != nil {
		return out, PipelineProcessError(c, err, "appending state to state store")
	}

	err = c.Checkpoint(ctx, in, matchedVal)
	if err != nil {
		return nil, PipelineProcessError(c, err, "checkpointing state")
	}

	completionStateBytes, err := c.getCompletionState(ctx, partitionID, name, matchedVal)
	if err != nil {
		return out, PipelineProcessError(c, err, "getting completion state")
	}


	completionState, err = core.NewCompletionStateFromBytes(completionStateBytes)
	if err != nil {
		return nil, PipelineProcessError(c, err, "deserializing completion state")
	}

	if completionState.CompletionDeadline > 0 && time.Now().UnixNano() > completionState.CompletionDeadline {
		err = common.SetUsingInternalPrefix(common.CompletionStatusPrefix, c.name, "timedout", out,
			true)
		if err != nil {
			return out, PipelineProcessError(c, err, "setting state to timed-out")
		}
	} else if completionState.IsDone() {
		err = c.completionStateStore.AtomicDelete(ctx, stateKey, completionStateBytes)
		if err != nil {
			return out, PipelineProcessError(c, err, "deleting state after completed")
		}
		err = common.SetUsingInternalPrefix(common.CompletionStatusPrefix, c.name, "complete", out, true)
		if err != nil {
			return out, PipelineProcessError(c, err, "setting status to completed")
		}
		stateMap, err := util.JsonToMap(completionStateBytes)
		if err != nil {
			return out, PipelineProcessError(c, err, "deserializing state")
		}
		err = common.SetUsingInternalPrefix(common.CompletionStatePrefix, c.name, stateMap, out, true)
		if err != nil {
			return out, PipelineProcessError(c, err, "setting state to complete")
		}
	} else {
		err = common.SetUsingInternalPrefix(common.CompletionStatusPrefix, c.name, "triggered", out,
			true)
		if err != nil {
			return out, PipelineProcessError(c, err, "setting state to triggered")
		}
	}


	return out, nil
}

func (c Completer) Checkpoint(ctx context.Context, in map[string]interface{}, matchedVal interface{}) error {
	partitionID, err := core.GetPartitionID(in)
	if err != nil {
		return err
	}
	stateKey, err := core.CompletionStateKey(partitionID, c.name, matchedVal)
	if err != nil {
		return err
	}
	return c.completionStateStore.Checkpoint(ctx, stateKey, c.checkpointMapFunc())
}
