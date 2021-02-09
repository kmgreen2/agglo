package core

import (
	"github.com/Knetic/govaluate"
	"fmt"
	"github.com/kmgreen2/agglo/pkg/util"
	"strings"
)

var invalidExpression string = "<INVALID>"

type variable struct {
	name string
}
func Variable(name string) variable {
	return variable{name}
}

type OperatorType int
const (
	UnknownType OperatorType = iota
	UnaryType
	BinaryType
	LogicalType
	ComparatorType
	ExistsType
)

func getOperatorType(operator interface{}) OperatorType {
	switch operator.(type) {
	case UnaryOperator: return UnaryType
	case BinaryOperator: return BinaryType
	case LogicalOperator: return LogicalType
	case ComparatorOperator: return ComparatorType
	case ExistsOperator: return ExistsType
	}
	return UnknownType
}

type ExistsOperator int
const (
	NotExists ExistsOperator = iota
	Exists
)

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

type ExistsExpression struct {
	keys []string
	operators []ExistsOperator
}

type ExistsExpressionBuilder struct {
	expression *ExistsExpression
}

func NewExistsExpressionBuilder() *ExistsExpressionBuilder {
	return &ExistsExpressionBuilder{
		&ExistsExpression{},
	}
}

func (builder *ExistsExpressionBuilder) Add(key string, operator ExistsOperator) *ExistsExpressionBuilder {
	builder.expression.keys = append(builder.expression.keys, key)
	builder.expression.operators = append(builder.expression.operators, operator)
	return builder
}

func (builder *ExistsExpressionBuilder) Get() *ExistsExpression {
	return builder.expression
}

func (expr *ExistsExpression) String() string {
	var components []string

	for i, key := range expr.keys {
		if expr.operators[i] == NotExists {
			components = append(components, fmt.Sprintf("!%s", key))
		} else {
			components = append(components, fmt.Sprintf("%s", key))
		}
	}
	return strings.Join(components, ",")
}

func (expr *ExistsExpression) VariablesExist(in map[string]interface{}) bool {
	flattened := util.Flatten(in)

	components := strings.Split(expr.String(), ",")
	for _, component := range components {
		notExistOp := false
		if component[0] == '!' {
			component = component[1:]
			notExistOp = true
		}

		/*
		 * ToDo(KMG): This will only resolve to true if the key is equal to a flattened key.  This
		 * means that it currently does not work for paths in the map.  For example:
		 *
		 * - map[string]interface{} {
		 *     "foo" : { "bar": { "baz": 1 } }
		 *   }
		 *
		 * - Exists("foo.bar") and Exists("foo") would be false, while Exists("foo.bar.baz") would be true
		 *
		 * This can be fixed by doing a tree-based search.  We cannot rely on simple prefix matching, because
		 * flattened keys, such as "foo.ba" would evaluate to true
		 */
		if _, ok := flattened[component]; ok {
			if notExistOp {
				return false
			}
		} else {
			if !notExistOp {
				return false
			}
		}
	}
	return true
}

func (expr *ExistsExpression) OperatorType() OperatorType {
	return ExistsType
}

type UnaryExpression struct {
	rhs interface{}
	operator UnaryOperator
	valid bool
}

func NewUnaryExpression(rhs interface{}, operator UnaryOperator) *UnaryExpression {
	return &UnaryExpression{
		rhs: rhs,
		operator: operator,
		valid: true,
	}
}

func (expr *UnaryExpression) VariablesExist(in map[string]interface{}) bool {
	flattened := util.Flatten(in)
	switch rhs := expr.rhs.(type) {
	case variable:
		if _, ok := flattened[rhs.name]; !ok {
			return false
		}
	}
	return true
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
	default:
		numericValue, err := util.GetInteger(rhs)
		if err != nil {
			return invalidExpression
		}
		return fmt.Sprintf("(%s%d)", opStr, numericValue)
	}
}
func (expr *UnaryExpression) OperatorType() OperatorType {
	return getOperatorType(expr.operator)
}

type BinaryExpression struct {
	lhs interface{}
	rhs interface{}
	operator BinaryOperator
	valid bool
}

func NewBinaryExpression(lhs, rhs interface{}, operator BinaryOperator) *BinaryExpression {
	return &BinaryExpression{
		lhs: lhs,
		rhs: rhs,
		operator: operator,
		valid: true,
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
	default:
		integerValue, err := util.GetInteger(lhs)
		if err == nil {
			formatString += "%d"
			argList = append(argList, integerValue)
		} else {
			numericValue, err := util.GetNumeric(lhs)
			if err != nil {
				return invalidExpression
			}
			formatString += "%.4f"
			argList = append(argList, numericValue)
		}
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
	default:
		integerValue, err := util.GetInteger(rhs)
		if err == nil {
			formatString += " %d"
			argList = append(argList, integerValue)
		} else {
			numericValue, err := util.GetNumeric(rhs)
			if err != nil {
				return invalidExpression
			}
			formatString += " %.4f"
			argList = append(argList, numericValue)
		}
	}
	formatString += ")"
	return fmt.Sprintf(formatString, argList...)
}

func (expr *BinaryExpression) VariablesExist(in map[string]interface{}) bool {
	flattened := util.Flatten(in)
	switch rhs := expr.rhs.(type) {
	case variable:
		if _, ok := flattened[rhs.name]; !ok {
			return false
		}
	}
	switch lhs := expr.lhs.(type) {
	case variable:
		if _, ok := flattened[lhs.name]; !ok {
			return false
		}
	}
	return true
}

func (expr *BinaryExpression) OperatorType() OperatorType {
	return getOperatorType(expr.operator)
}

type LogicalExpression struct {
	lhs Expression
	rhs Expression
	operator LogicalOperator
	valid bool
}

func NewLogicalExpression(lhs, rhs Expression, operator LogicalOperator) *LogicalExpression {
	return &LogicalExpression{
		lhs: lhs,
		rhs: rhs,
		operator: operator,
		valid: true,
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

func (expr *LogicalExpression) VariablesExist(in map[string]interface{}) bool {
	return expr.lhs.VariablesExist(in) && expr.rhs.VariablesExist(in)
}

func (expr *LogicalExpression) OperatorType() OperatorType {
	return getOperatorType(expr.operator)
}

type ComparatorExpression struct {
	lhs interface{}
	rhs interface{}
	operator ComparatorOperator
	valid bool
}

func NewComparatorExpression(lhs, rhs interface{}, operator ComparatorOperator) *ComparatorExpression {
	return &ComparatorExpression{
		lhs: lhs,
		rhs: rhs,
		operator: operator,
		valid: true,
	}
}
func (expr *ComparatorExpression) String() string {
	var opStr string
	formatString := "("
	var argList []interface{}

	switch lhs := expr.lhs.(type) {
	case string:
		formatString += "\"%s\""
		argList = append(argList, lhs)
	case variable:
		formatString += "[%s]"
		argList = append(argList, lhs.name)
	case Expression:
		formatString += "%s"
		argList = append(argList, lhs.String())
	default:
		integerValue, err := util.GetInteger(lhs)
		if err == nil {
			formatString += "%d"
			argList = append(argList, integerValue)
		} else {
			numericValue, err := util.GetNumeric(lhs)
			if err != nil {
				return invalidExpression
			}
			formatString += "%.4f"
			argList = append(argList, numericValue)
		}
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
		formatString += " \"%s\""
		argList = append(argList, rhs)
	case variable:
		formatString += " [%s]"
		argList = append(argList, rhs.name)
	case Expression:
		formatString += " %s"
		argList = append(argList, rhs.String())
	default:
		integerValue, err := util.GetInteger(rhs)
		if err == nil {
			formatString += " %d"
			argList = append(argList, integerValue)
		} else {
			numericValue, err := util.GetNumeric(rhs)
			if err != nil {
				return invalidExpression
			}
			formatString += " %.4f"
			argList = append(argList, numericValue)
		}
	}
	formatString += ")"
	return fmt.Sprintf(formatString, argList...)
}

func (expr *ComparatorExpression) VariablesExist(in map[string]interface{}) bool {
	flattened := util.Flatten(in)
	switch rhs := expr.rhs.(type) {
	case variable:
		if _, ok := flattened[rhs.name]; !ok {
			return false
		}
	case Expression:
		if !rhs.VariablesExist(in) {
			return false
		}
	}
	switch lhs := expr.lhs.(type) {
	case variable:
		if _, ok := flattened[lhs.name]; !ok {
			return false
		}
	case Expression:
		if !lhs.VariablesExist(in) {
			return false
		}
	}
	return true
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

func (expr *TrueExpression) VariablesExist(map[string]interface{}) bool {
	return true
}

type FalseExpression struct {
}

func (expr *FalseExpression) String() string {
	return "false"
}

func (expr *FalseExpression) OperatorType() OperatorType {
	return UnaryType
}

func (expr *FalseExpression) VariablesExist(map[string]interface{}) bool {
	return true
}

func NewFalseExpression() *FalseExpression {
	return &FalseExpression{}
}

type Expression interface {
	String() string
	OperatorType() OperatorType
	VariablesExist(map[string]interface{}) bool
	//Validate() bool -> Account for inconvenience of returning error from the builder functions
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
	case *FalseExpression: return cond, nil
	case *ExistsExpression: return cond, nil
	default: return nil, fmt.Errorf("")
	}
}

var TrueCondition = &Condition {
	NewTrueExpression(),
}

var FalseCondition = &Condition {
	NewFalseExpression(),
}


func (c *Condition) Evaluate(in map[string]interface{}) (bool, error) {
	// govaluate does not support checking the existence of fields in,
	// so we had to do it ourselves
	// Note that Exists operations cannot be used in conjunction with
	// other types of expressions, so they have to be different processes
	// in a pipeline
	variablesExist := c.expr.VariablesExist(in)
	if c.expr.OperatorType() == ExistsType {
		return variablesExist, nil
	}

	// ToDo(KMG): Should we return an error here?  It seems that non-existence
	// of a field implies the condition is false, so not returning error for now
	//
	// This check is needed because govaluate will return an error is a variable
	// in an expression is not found in the provided map
	if !variablesExist {
		return false, nil
	}

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