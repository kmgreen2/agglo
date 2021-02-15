package process

import (
	"context"
	"github.com/kmgreen2/agglo/internal/core"
	"github.com/kmgreen2/agglo/pkg/util"
)

type Continuation struct {
	name string
	condition *core.Condition
}

func NewContinuation(name string, condition *core.Condition) *Continuation {
	return &Continuation{
		name: name,
		condition: condition,
	}
}

func (c Continuation) Name() string {
	return c.name
}

func (c *Continuation) Process(ctx context.Context, in map[string]interface{}) (map[string]interface{},
	error) {
	out := util.CopyableMap(in).DeepCopy()
	if ok, err := c.condition.Evaluate(in); ok && err == nil {
		return out, nil
	}
	err := util.NewContinuationNotSatisfied("c stopping processing")
	return nil, PipelineProcessError(c, err, "")
}