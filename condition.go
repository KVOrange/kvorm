package kvorm

import (
	"fmt"
	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
)

type Identifiable interface {
	exp.Inable
	exp.Comparable
	exp.Likeable
	exp.Isable
}

type Conditional interface {
	ToCondition() (Identifiable, []ModelField)
}

type Condition interface {
	Condition() (exp.Expression, []ModelField)
}

type And struct {
	Field Conditional
	Op    string
	Value interface{}
}

func (a And) Condition() (exp.Expression, []ModelField) {
	var condition exp.Expression
	ident, fields := a.Field.ToCondition()
	switch a.Op {
	case In:
		condition = ident.In(a.Value)
	case NotIn:
		condition = ident.NotIn(a.Value)
	case Eq:
		condition = ident.Eq(a.Value)
	case NotEq:
		condition = ident.Neq(a.Value)
	case Like:
		condition = ident.Like(a.Value)
	case NotLike:
		condition = ident.NotLike(a.Value)
	case Regex:
		condition = ident.RegexpLike(a.Value)
	case RegexI:
		condition = ident.RegexpILike(a.Value)
	case NotRegex:
		condition = ident.RegexpNotLike(a.Value)
	case NotRegexI:
		condition = ident.RegexpNotILike(a.Value)
	case Lt:
		condition = ident.Lt(a.Value)
	case Lte:
		condition = ident.Lte(a.Value)
	case Gt:
		condition = ident.Gt(a.Value)
	case Gte:
		condition = ident.Gte(a.Value)
	case IsNotNull:
		condition = ident.IsNotNull()
	case IsNull:
		condition = ident.IsNull()
	default:
		panic(fmt.Sprintf("operator %s can not be found", a.Op))
	}
	return condition, fields
}

type OrCondition struct {
	Conditions []Condition
}

func (c OrCondition) Condition() (exp.Expression, []ModelField) {
	var fields []ModelField
	var resultCondition []exp.Expression
	for _, condition := range c.Conditions {
		subCondition, subFields := condition.Condition()
		fields = append(fields, subFields...)
		resultCondition = append(resultCondition, subCondition)
	}
	return goqu.Or(resultCondition...), fields
}

type ConditionGroup []Condition

func (c ConditionGroup) Condition() (exp.Expression, []ModelField) {
	var fields []ModelField
	var resultCondition []exp.Expression
	for _, condition := range c {
		subCondition, subFields := condition.Condition()
		fields = append(fields, subFields...)
		resultCondition = append(resultCondition, subCondition)
	}
	return goqu.And(resultCondition...), fields
}
