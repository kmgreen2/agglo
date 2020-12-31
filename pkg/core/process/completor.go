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
	completion *core.Completion
	completionStateStore kvs.KVStore
}

func NewCompleter(completion *core.Completion, kvStore kvs.KVStore) *Completer {
	return &Completer{
		completion: completion,
		completionStateStore: kvStore,
	}
}

func (c Completer) getCompletionState(partitionID gUuid.UUID, name string, value interface{}) ([]byte, error) {
	stateKey, err := core.CompletionStateKey(partitionID, name, value)
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
	var completionState *core.CompletionState

	out := core.CopyableMap(in).DeepCopy()

	partitionID, err := core.GetPartitionID(in)
	if err != nil {
		return out, err
	}
	name, err := core.GetName(in)
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
