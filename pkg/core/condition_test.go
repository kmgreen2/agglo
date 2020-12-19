package core_test

import (
	"github.com/kmgreen2/agglo/pkg/core"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewBinaryExpressions(t *testing.T) {
	expr := core.NewBinaryExpression("foo", "bar", core.Addition)
	assert.Equal(t, expr.String(), "([foo] + [bar])")
	expr = core.NewBinaryExpression("foo", "bar", core.Subtract)
	assert.Equal(t, expr.String(), "([foo] - [bar])")
	expr = core.NewBinaryExpression("foo", "bar", core.Multiply)
	assert.Equal(t, expr.String(), "([foo] * [bar])")
	expr = core.NewBinaryExpression("foo", "bar", core.Divide)
	assert.Equal(t, expr.String(), "([foo] / [bar])")
	expr = core.NewBinaryExpression("foo", "bar", core.Power)
	assert.Equal(t, expr.String(), "([foo] ** [bar])")
	expr = core.NewBinaryExpression("foo", "bar", core.Modulus)
	assert.Equal(t, expr.String(), "([foo] % [bar])")
	expr = core.NewBinaryExpression("foo", "bar", core.RightShift)
	assert.Equal(t, expr.String(), "([foo] >> [bar])")
	expr = core.NewBinaryExpression("foo", "bar", core.LeftShift)
	assert.Equal(t, expr.String(), "([foo] << [bar])")
	expr = core.NewBinaryExpression("foo", "bar", core.Or)
	assert.Equal(t, expr.String(), "([foo] | [bar])")
	expr = core.NewBinaryExpression("foo", "bar", core.And)
	assert.Equal(t, expr.String(), "([foo] & [bar])")
	expr = core.NewBinaryExpression("foo", "bar", core.Xor)
	assert.Equal(t, expr.String(), "([foo] ^ [bar])")
}

func TestNewUnaryExpressions(t *testing.T) {
	expr := core.NewUnaryExpression("foo", core.Negation)
	assert.Equal(t, expr.String(), "(-[foo])")
	expr = core.NewUnaryExpression("foo", core.Inversion)
	assert.Equal(t, expr.String(), "(![foo])")
	expr = core.NewUnaryExpression("foo", core.Not)
	assert.Equal(t, expr.String(), "(~[foo])")
}

func TestNewComparatorExpressions(t *testing.T) {
	expr := core.NewComparatorExpression("foo", "bar", core.GreaterThan)
	assert.Equal(t, expr.String(), "([foo] > [bar])")
	expr = core.NewComparatorExpression("foo", "bar", core.LessThan)
	assert.Equal(t, expr.String(), "([foo] < [bar])")
	expr = core.NewComparatorExpression("foo", "bar", core.GreaterThanOrEqual)
	assert.Equal(t, expr.String(), "([foo] >= [bar])")
	expr = core.NewComparatorExpression("foo", "bar", core.LessThanOrEqual)
	assert.Equal(t, expr.String(), "([foo] >= [bar])")
	expr = core.NewComparatorExpression("foo", "bar", core.Equal)
	assert.Equal(t, expr.String(), "([foo] == [bar])")
	expr = core.NewComparatorExpression("foo", "bar", core.NotEqual)
	assert.Equal(t, expr.String(), "([foo] != [bar])")
	expr = core.NewComparatorExpression("foo", "bar", core.RegexMatch)
	assert.Equal(t, expr.String(), "([foo] =~ [bar])")
	expr = core.NewComparatorExpression("foo", "bar", core.RegexNotMatch)
	assert.Equal(t, expr.String(), "([foo] !~ [bar])")
}

func TestComplexExpressions(t *testing.T) {
	lhsExpr := core.NewComparatorExpression("foo", "bar", core.GreaterThan)
	rhsExpr := core.NewUnaryExpression("baz", core.Not)
	expr := core.NewLogicalExpression(lhsExpr, rhsExpr, core.LogicalAnd)
	assert.Equal(t, expr.String(), "(([foo] > [bar]) && (~[baz]))")
}

func TestComplexEvaluation(t *testing.T) {
	testMap := map[string]interface{}{
		"foo": 5,
		"bar": 7,
		"baz": 1,
	}
	lhsExpr := core.NewComparatorExpression("foo", "bar", core.GreaterThan)
	rhsExpr := core.NewComparatorExpression("foo", "bar", core.LessThan)
	expr := core.NewLogicalExpression(lhsExpr, rhsExpr, core.LogicalAnd)
	cond, err := core.NewCondition(expr)
	assert.Nil(t, err)
	result, err := cond.Evaluate(testMap)
	assert.Nil(t, err)
	assert.False(t, result)
}
