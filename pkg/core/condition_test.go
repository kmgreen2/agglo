package core_test

import (
	"fmt"
	"github.com/kmgreen2/agglo/pkg/core"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewBinaryExpressions(t *testing.T) {
	lhsVar := core.Variable("foo")
	rhsVar := core.Variable("bar")
	lhsNumber := core.Numeric(1)
	rhsNumber := core.Numeric(2)
	ops := []string{"+", "-", "*", "/", "**", "%", ">>", "<<", "|", "&", "^"}
	opCodes := []core.BinaryOperator{core.Addition, core.Subtract, core.Multiply, core.Divide, core.Power,
		core.Modulus, core.RightShift, core.LeftShift, core.Or, core.And, core.Xor}

	for i, op := range ops {
		expr := core.NewBinaryExpression(lhsVar, rhsVar, opCodes[i])
		assert.Equal(t, fmt.Sprintf("([foo] %s [bar])", op), expr.String())
		expr = core.NewBinaryExpression(lhsNumber, rhsVar, opCodes[i])
		assert.Equal(t, fmt.Sprintf("(1.0000 %s [bar])", op), expr.String())
		expr = core.NewBinaryExpression(lhsNumber, rhsNumber, opCodes[i])
		assert.Equal(t, fmt.Sprintf("(1.0000 %s 2.0000)", op), expr.String())
		expr = core.NewBinaryExpression(lhsVar, rhsNumber, opCodes[i])
		assert.Equal(t, fmt.Sprintf("([foo] %s 2.0000)", op), expr.String())
	}
}

func TestNewUnaryExpressions(t *testing.T) {
	expr := core.NewUnaryExpression(core.Variable("foo"), core.Negation)
	assert.Equal(t, expr.String(), "(-[foo])")
	expr = core.NewUnaryExpression(core.Variable("foo"), core.Inversion)
	assert.Equal(t, expr.String(), "(![foo])")
	expr = core.NewUnaryExpression(core.Variable("foo"), core.Not)
	assert.Equal(t, expr.String(), "(~[foo])")
	expr = core.NewUnaryExpression(core.Numeric(1), core.Negation)
	assert.Equal(t, expr.String(), "(-1)")
	expr = core.NewUnaryExpression(core.Numeric(1), core.Inversion)
	assert.Equal(t, expr.String(), "(!1)")
	expr = core.NewUnaryExpression(core.Numeric(1), core.Not)
	assert.Equal(t, expr.String(), "(~1)")
}

func TestTrueExpression(t *testing.T) {
	expr := core.NewTrueExpression()
	assert.Equal(t, "true", expr.String())
	cond, err := core.NewCondition(expr)
	assert.Nil(t, err)
	result, err := cond.Evaluate(make(map[string]interface{}))
	assert.Nil(t, err)
	assert.True(t, result)
}

func TestNewComparatorExpressions(t *testing.T) {
	lhsVar := core.Variable("foo")
	rhsVar := core.Variable("bar")
	lhsNumber := core.Numeric(1)
	rhsNumber := core.Numeric(2)
	ops := []string{">", "<", ">=", "<=", "==", "!=", "=~", "!~"}
	opCodes := []core.ComparatorOperator{core.GreaterThan, core.LessThan, core.GreaterThanOrEqual, core.LessThanOrEqual,
		core.Equal, core.NotEqual, core.RegexMatch, core.RegexNotMatch}

	for i, op := range ops {
		expr := core.NewComparatorExpression(lhsVar, rhsVar, opCodes[i])
		assert.Equal(t, fmt.Sprintf("([foo] %s [bar])", op), expr.String())
		expr = core.NewComparatorExpression(lhsNumber, rhsVar, opCodes[i])
		assert.Equal(t, fmt.Sprintf("(1.0000 %s [bar])", op), expr.String())
		expr = core.NewComparatorExpression(lhsNumber, rhsNumber, opCodes[i])
		assert.Equal(t, fmt.Sprintf("(1.0000 %s 2.0000)", op), expr.String())
		expr = core.NewComparatorExpression(lhsVar, rhsNumber, opCodes[i])
		assert.Equal(t, fmt.Sprintf("([foo] %s 2.0000)", op), expr.String())
	}
}

func TestComplexExpressions(t *testing.T) {
	lhsExpr := core.NewComparatorExpression(core.Variable("foo"), core.Variable("bar"), core.GreaterThan)
	rhsExpr := core.NewUnaryExpression(core.Variable("baz"), core.Not)
	expr := core.NewLogicalExpression(lhsExpr, rhsExpr, core.LogicalAnd)
	assert.Equal(t, expr.String(), "(([foo] > [bar]) && (~[baz]))")
}

func TestComplexEvaluation(t *testing.T) {
	testMap := map[string]interface{}{
		"foo": 5,
		"bar": 7,
		"baz": 1,
	}
	lhsExpr := core.NewComparatorExpression(core.Variable("foo"), core.Variable("bar"), core.GreaterThan)
	rhsExpr := core.NewComparatorExpression(core.Variable("foo"), core.Variable("bar"), core.LessThan)
	expr := core.NewLogicalExpression(lhsExpr, rhsExpr, core.LogicalAnd)
	cond, err := core.NewCondition(expr)
	assert.Nil(t, err)
	result, err := cond.Evaluate(testMap)
	assert.Nil(t, err)
	assert.False(t, result)
}

