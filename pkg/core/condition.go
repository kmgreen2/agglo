package core

import (
	"github.com/Knetic/govaluate"
	"fmt"
)

type variable struct {
	name string
}
func Variable(name string) variable {
	return variable{name}
}

type numeric struct {
	value float64
}
func Numeric(value float64) numeric {
	return numeric{value}
}

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
	rhs interface{}
	operator UnaryOperator
}

func NewUnaryExpression(rhs interface{}, operator UnaryOperator) *UnaryExpression {
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

	switch rhs := expr.rhs.(type) {
	case variable:
		return fmt.Sprintf("(%s[%s])", opStr, rhs.name)
	case numeric:
		return fmt.Sprintf("(%s%d)", opStr, int(rhs.value))
	}
	return "<invalid>"
}
func (expr *UnaryExpression) OperatorType() OperatorType {
	return getOperatorType(expr.operator)
}

type BinaryExpression struct {
	lhs interface{}
	rhs interface{}
	operator BinaryOperator
}

func NewBinaryExpression(lhs, rhs interface{}, operator BinaryOperator) *BinaryExpression {
	return &BinaryExpression{
		lhs: lhs,
		rhs: rhs,
		operator: operator,
	}
}
func (expr *BinaryExpression) String() string {
	var opStr string
	formatString := "("
	var argList []interface{}
	switch lhs := expr.lhs.(type) {
	case variable:
		formatString += "[%s]"
		argList = append(argList, lhs.name)
	case numeric:
		formatString += "%.4f"
		argList = append(argList, lhs.value)
	}
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
	formatString += " %s"
	argList = append(argList, opStr)
	switch rhs := expr.rhs.(type) {
	case variable:
		formatString += " [%s]"
		argList = append(argList, rhs.name)
	case numeric:
		formatString += " %.4f"
		argList = append(argList, rhs.value)
	}
	formatString += ")"
	return fmt.Sprintf(formatString, argList...)
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
	var opStr string
	formatString := "("
	var argList []interface{}

	switch lhs := expr.lhs.(type) {
	case string:
		formatString += "%s"
		argList = append(argList, lhs)
	case variable:
		formatString += "[%s]"
		argList = append(argList, lhs.name)
	case numeric:
		formatString += "%.4f"
		argList = append(argList, lhs.value)
	case Expression:
		formatString += "%s"
		argList = append(argList, lhs.String())
	}

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

	formatString += " %s"
	argList = append(argList, opStr)

	switch rhs := expr.rhs.(type) {
	case string:
		formatString += " %s"
		argList = append(argList, rhs)
	case variable:
		formatString += " [%s]"
		argList = append(argList, rhs.name)
	case numeric:
		formatString += " %.4f"
		argList = append(argList, rhs.value)
	case Expression:
		formatString += " %s"
		argList = append(argList, rhs.String())
	}
	formatString += ")"
	return fmt.Sprintf(formatString, argList...)
}
func (expr *ComparatorExpression) OperatorType() OperatorType {
	return getOperatorType(expr.operator)
}

type TrueExpression struct {
}

func (expr *TrueExpression) String() string {
	return "true"
}

func (expr *TrueExpression) OperatorType() OperatorType {
	return UnaryType
}

func NewTrueExpression() *TrueExpression {
	return &TrueExpression{}
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
	case *TrueExpression: return cond, nil
	default: return nil, fmt.Errorf("")
	}
}

var TrueCondition = &Condition {
	NewTrueExpression(),
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