package process_test

import (
	"context"
	"github.com/kmgreen2/agglo/internal/core"
	"github.com/kmgreen2/agglo/internal/core/process"
	"github.com/kmgreen2/agglo/test"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAnnotateHappyPath(t *testing.T) {
	jsonMap := test.TestJson()

	builder := process.NewAnnotatorBuilder("foo")

	cond, err := core.NewCondition(core.NewComparatorExpression(core.Variable("b.d.0"), 3, core.Equal))
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	builder.Add(core.NewAnnotation("foo", "bar", cond))

	cond, err = core.NewCondition(core.NewComparatorExpression(core.Variable("b.d.1"), 2, core.Equal))
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	builder.Add(core.NewAnnotation("fizz", "buzz", cond))

	annotate := builder.Build()

	result, err := annotate.Process(context.Background(), jsonMap)
	assert.Nil(t, err)
	assert.Equal(t, "bar", result["foo"])
	_, ok := result["fizz"]
	assert.False(t, ok)
}

func TestAnnotateDuplicateField(t *testing.T) {
	jsonMap := test.TestJson()

	builder := process.NewAnnotatorBuilder("foo")

	cond, err := core.NewCondition(core.NewComparatorExpression(core.Variable("b.d.0"), 3, core.Equal))
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	builder.Add(core.NewAnnotation("a", "bar", cond))

	annotate := builder.Build()

	_, err = annotate.Process(context.Background(), jsonMap)
	assert.Error(t, err)
}
