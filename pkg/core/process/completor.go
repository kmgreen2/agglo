package process

import (
	"context"
	"errors"
	"fmt"
	gUuid "github.com/google/uuid"
	"github.com/kmgreen2/agglo/pkg/common"
	"github.com/kmgreen2/agglo/pkg/core"
	"github.com/kmgreen2/agglo/pkg/kvs"
	"time"
)

type Completer struct {
	name string
	completion *core.Completion
	completionStateStore kvs.KVStore
}

func NewCompleter(name string, completion *core.Completion, kvStore kvs.KVStore) *Completer {
	return &Completer{
		name: name,
		completion: completion,
		completionStateStore: kvStore,
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
		if errors.Is(err, &common.NotFoundError{}) {
			return nil, nil
		}
		return nil, err
	}
	return stateBytes, nil
}

func (c Completer) Process(ctx context.Context, in map[string]interface{}) (map[string]interface{}, error) {
	var completionState *core.CompletionState

	out := core.CopyableMap(in).DeepCopy()

	partitionID, err := core.GetPartitionID(in)
	if err != nil {
		return out, err
	}
	/*name, err := core.GetName(in)
	if err != nil {
		return out, err
	}
	*/
	name := c.name

	matchedKey, matchedVal, err := c.completion.Match(in)
	if err != nil {
		if errors.Is(err, &common.NotFoundError{}) {
			return out, nil
		}
		return out, err
	}

	completionStateBytes, err := c.getCompletionState(ctx, partitionID, name, matchedVal)
	if err != nil {
		return out, err
	}

	if completionStateBytes != nil {
		completionState, err = core.NewCompletionStateFromBytes(completionStateBytes)
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
		completionState = core.NewCompletionState(matchedVal, resolved, completionDeadline)
	}

	completionState.Resolved[matchedKey] = true

	stateKey, err := core.CompletionStateKey(partitionID, name, matchedVal)
	if err != nil {
		return out, err
	}

	newCompletionBytes, err := completionState.Bytes()
	if err != nil {
		return out, err
	}

	if completionState.IsDone() {
		err = c.completionStateStore.AtomicDelete(ctx, stateKey, completionStateBytes)
		if err != nil {
			return out, err
		}
		out[fmt.Sprintf("agglo:completion:%s", name)] = "complete"
		stateMap, err := common.JsonToMap(newCompletionBytes)
		if err != nil {
			return out, err
		}
		out[fmt.Sprintf("agglo:completion:state:%s", name)] = stateMap
	} else if completionState.CompletionDeadline > 0 && time.Now().UnixNano() > completionState.CompletionDeadline {
			out[fmt.Sprintf("agglo:completion:%s", name)] = "timedout"
	} else {
		err = c.completionStateStore.AtomicPut(ctx, stateKey, completionStateBytes, newCompletionBytes)
		if err != nil {
			return out, err
		}
		out[fmt.Sprintf("agglo:completion:%s", name)] = "triggered"
	}

	return out, nil
}
