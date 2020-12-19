package core

import (
	"github.com/Knetic/govaluate"
	_ "github.com/Knetic/govaluate"
	"fmt"
)

type OperatorType int
const (
	UnknownType OperatorType = iota
	UnaryType
	BinaryType
	LogicalType
	ComparatorType
)

func getOperatorType(operator interface{}) OperatorType {
	switch operator.(type) {
	case UnaryOperator: return UnaryType
	case BinaryOperator: return BinaryType
	case LogicalOperator: return LogicalType
	case ComparatorOperator: return ComparatorType
	}
	return UnknownType
}

type UnaryOperator int
const (
	Negation UnaryOperator = iota
	Inversion
	Not
)

type BinaryOperator int
const (
	Addition BinaryOperator = iota
	Subtract
	Multiply
	Divide
	Power
	Modulus
	RightShift
	LeftShift
	Or
	And
	Xor
)

type LogicalOperator int
const (
	LogicalAnd LogicalOperator = iota
	LogicalOr
)

type ComparatorOperator int
const (
	GreaterThan ComparatorOperator = iota
	LessThan
	GreaterThanOrEqual
	LessThanOrEqual
	Equal
	NotEqual
	RegexMatch
	RegexNotMatch
)

type UnaryExpression struct {
	rhs string
	operator UnaryOperator
}

func NewUnaryExpression(rhs string, operator UnaryOperator) *UnaryExpression {
	return &UnaryExpression{
		rhs: rhs,
		operator: operator,
	}
}
func (expr *UnaryExpression) String() string {
	var opStr string
	switch expr.operator {
	case Negation: opStr = "-"
	case Inversion: opStr = "!"
	case Not: opStr = "~"
	}
	return fmt.Sprintf("(%s[%s])", opStr, expr.rhs)
}
func (expr *UnaryExpression) OperatorType() OperatorType {
	return getOperatorType(expr.operator)
}

type BinaryExpression struct {
	lhs string
	rhs string
	operator BinaryOperator
}

func NewBinaryExpression(lhs, rhs string, operator BinaryOperator) *BinaryExpression {
	return &BinaryExpression{
		lhs: lhs,
		rhs: rhs,
		operator: operator,
	}
}
func (expr *BinaryExpression) String() string {
	var opStr string
	switch expr.operator {
	case Addition: opStr ="+"
	case Subtract: opStr ="-"
	case Multiply: opStr ="*"
	case Divide: opStr ="/"
	case Power: opStr ="**"
	case Modulus: opStr ="%"
	case RightShift: opStr =">>"
	case LeftShift: opStr ="<<"
	case Or: opStr ="|"
	case And: opStr ="&"
	case Xor: opStr ="^"
	}
	return fmt.Sprintf("([%s] %s [%s])", expr.lhs, opStr, expr.rhs)
}
func (expr *BinaryExpression) OperatorType() OperatorType {
	return getOperatorType(expr.operator)
}

type LogicalExpression struct {
	lhs Expression
	rhs Expression
	operator LogicalOperator
}

func NewLogicalExpression(lhs, rhs Expression, operator LogicalOperator) *LogicalExpression {
	return &LogicalExpression{
		lhs: lhs,
		rhs: rhs,
		operator: operator,
	}
}
func (expr *LogicalExpression) String() string {
	var opStr string
	switch expr.operator {
	case LogicalAnd: opStr = "&&"
	case LogicalOr: opStr = "||"
	}
	return fmt.Sprintf("(%s %s %s)", expr.lhs.String(), opStr, expr.rhs.String())
}
func (expr *LogicalExpression) OperatorType() OperatorType {
	return getOperatorType(expr.operator)
}

type ComparatorExpression struct {
	lhs interface{}
	rhs interface{}
	operator ComparatorOperator
}

func NewComparatorExpression(lhs, rhs interface{}, operator ComparatorOperator) *ComparatorExpression {
	return &ComparatorExpression{
		lhs: lhs,
		rhs: rhs,
		operator: operator,
	}
}
func (expr *ComparatorExpression) String() string {
	var opStr, lhsStr, rhsStr string
	switch expr.operator {
    case GreaterThan: opStr = ">"
    case LessThan: opStr = "<"
    case GreaterThanOrEqual: opStr = ">="
    case LessThanOrEqual: opStr = "<="
    case Equal: opStr = "=="
    case NotEqual: opStr = "!="
    case RegexMatch: opStr = "=~"
    case RegexNotMatch: opStr = "!~"
	}

	switch v := expr.lhs.(type) {
	case string: lhsStr = fmt.Sprintf("[%s]", v)
	case Expression: lhsStr = v.String()
	default: lhsStr = "<nil>"
	}

	switch v := expr.rhs.(type) {
	case string: rhsStr = fmt.Sprintf("[%s]", v)
	case Expression: rhsStr = v.String()
	default: rhsStr = "<nil>"
	}

	return fmt.Sprintf("(%s %s %s)", lhsStr, opStr, rhsStr)
}
func (expr *ComparatorExpression) OperatorType() OperatorType {
	return getOperatorType(expr.operator)
}

// ToDo(KMG): Create builder

type Expression interface {
	String() string
	OperatorType() OperatorType
	//Validate() -> Account for inconvenience of returning error from the builder functions
}

type Condition struct {
	expr Expression
}

func NewCondition(expr Expression) (*Condition, error) {
	cond := &Condition{
		expr,
	}
	switch expr.(type) {
	case *ComparatorExpression: return cond, nil
	case *LogicalExpression: return cond, nil
	default: return nil, fmt.Errorf("")
	}
}

func (c *Condition) Evaluate(in map[string]interface{}) (bool, error) {
	expression, err := govaluate.NewEvaluableExpression(c.expr.String())
	if err != nil {
		return false, err
	}
	result, err := expression.Evaluate(in)
	if err != nil {
		return false, err
	}
	if boolResult, ok := result.(bool); ok {
		return boolResult, nil
	}

	return false, fmt.Errorf("")
}