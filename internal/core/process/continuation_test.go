package process_test

import (
	"context"
	"github.com/kmgreen2/agglo/internal/core"
	"github.com/kmgreen2/agglo/internal/core/process"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestContinuation(t *testing.T) {
	m := map[string]interface{} {
		"foo": 1,
	}
	condition, err := core.NewCondition(core.NewTrueExpression())
	if err != nil {
		assert.Fail(t, err.Error())
	}
	continuation := process.NewContinuation("foo", condition)
	out, err := continuation.Process(context.Background(), m)
	assert.Equal(t, m, out)
	assert.Nil(t, err)
}

func TestNotContinuation(t *testing.T) {
	m := map[string]interface{} {
		"foo": 1,
	}
	condition, err := core.NewCondition(core.NewFalseExpression())
	if err != nil {
		assert.Fail(t, err.Error())
	}
	continuation := process.NewContinuation("foo", condition)
	out, err := continuation.Process(context.Background(), m)
	assert.Nil(t, out)
	assert.Error(t, err)
}