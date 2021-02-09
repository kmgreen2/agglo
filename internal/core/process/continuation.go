package process

import (
	"context"
	"github.com/kmgreen2/agglo/internal/core"
	"github.com/kmgreen2/agglo/pkg/util"
)

type Continuation struct {
	condition *core.Condition
}

func NewContinuation(condition *core.Condition) *Continuation {
	return &Continuation{
		condition: condition,
	}
}

func (continuation *Continuation) Process(ctx context.Context, in map[string]interface{}) (map[string]interface{},
	error) {
	out := util.CopyableMap(in).DeepCopy()
	if ok, err := continuation.condition.Evaluate(in); ok && err == nil {
		return out, nil
	}
	return nil, util.NewContinuationNotSatisfied("continuation stopping processing")
}