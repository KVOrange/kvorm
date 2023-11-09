package kvorm

import (
	"fmt"
	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
)

type ConditionI interface {
	Condition(model *Model) (exp.Expression, []string)
}

type Identifiable interface {
	exp.Inable
	exp.Comparable
	exp.Likeable
	exp.Isable
}

// And Структура для формирования AND условия
// Field - строковое обозначение поля в модели
// Op -строковое указание необходимого оператора.
// Доступные значения - in, notIn, eq, notEq, like, notLike, regex, regexI, notRegex, notRegexI
// Value - значение для сравнения
type And struct {
	Field interface{}
	Op    string
	Value interface{}
}

type AndGroup []And

func (group AndGroup) Condition(model *Model) (exp.Expression, []string) {
	var conditions []exp.Expression
	var allFields []string

	for _, andCondition := range group {
		condExp, fields := andCondition.Condition(model)
		conditions = append(conditions, condExp)
		allFields = append(allFields, fields...)
	}

	return goqu.And(conditions...), allFields
}

type ConditionGroup []ConditionI

func (group ConditionGroup) Condition(model *Model) (exp.Expression, []string) {
	var conditions []exp.Expression
	var allFields []string

	for _, andCondition := range group {
		condExp, fields := andCondition.Condition(model)
		conditions = append(conditions, condExp)
		allFields = append(allFields, fields...)
	}

	return goqu.And(conditions...), allFields
}

func (el And) Condition(model *Model) (exp.Expression, []string) {
	var fields []string
	var condition exp.Expression
	var ident Identifiable

	switch el.Field.(type) {
	case string:
		field := el.Field.(string)
		fields = append(fields, field)
		_, ok := model.PreparedSelectors[field]
		if ok {
			ident = goqu.I(fmt.Sprintf("%s.%s", model.TableName, el.Field))
		} else {
			ident = goqu.I(fmt.Sprintf("%s%s%s", model.TableName, SEPARATOR, replaceLastSeparator(field, SEPARATOR)))
		}
	case SubExpression:
		subExp := el.Field.(SubExpression)
		fields = append(fields, subExp.Field1)
		fields = append(fields, subExp.Field2)

		var fieldOneIdent exp.IdentifierExpression
		var fieldTwoIdent exp.IdentifierExpression
		_, ok := model.PreparedSelectors[subExp.Field1]
		if ok {
			fieldOneIdent = goqu.I(fmt.Sprintf("%s.%s", model.TableName, subExp.Field1))
		} else {
			fieldOneIdent = goqu.I(fmt.Sprintf("%s%s%s", model.TableName, SEPARATOR, replaceLastSeparator(subExp.Field1, SEPARATOR)))
		}

		_, ok = model.PreparedSelectors[subExp.Field2]
		if ok {
			fieldTwoIdent = goqu.I(fmt.Sprintf("%s.%s", model.TableName, subExp.Field2))
		} else {
			fieldTwoIdent = goqu.I(fmt.Sprintf("%s%s%s", model.TableName, SEPARATOR, replaceLastSeparator(subExp.Field2, SEPARATOR)))
		}
		ident = goqu.L("? - ?", fieldOneIdent, fieldTwoIdent)
	case AvgExpression:
		avgExp := el.Field.(AvgExpression)
		fields = avgExp.Fields
		var fieldsIdents []Identifiable
		for _, field := range fields {
			var fieldIdent exp.IdentifierExpression
			_, ok := model.PreparedSelectors[field]
			if ok {
				fieldIdent = goqu.I(fmt.Sprintf("%s.%s", model.TableName, field))
			} else {
				fieldIdent = goqu.I(fmt.Sprintf("%s%s%s", model.TableName, SEPARATOR, replaceLastSeparator(field, SEPARATOR)))
			}
			fieldsIdents = append(fieldsIdents, fieldIdent)
		}
		ident = goqu.L(avgExp.Sql, fieldsIdents)
	default:
		panic(fmt.Sprintf("unsupported type %T", el.Field))
	}

	switch el.Op {
	case "in":
		condition = ident.In(el.Value)
	case "notIn":
		condition = ident.NotIn(el.Value)
	case "eq":
		condition = ident.Eq(el.Value)
	case "notEq":
		condition = ident.Neq(el.Value)
	case "like":
		condition = ident.Like(el.Value)
	case "notLike":
		condition = ident.NotLike(el.Value)
	case "regex":
		condition = ident.RegexpLike(el.Value)
	case "regexI":
		condition = ident.RegexpILike(el.Value)
	case "notRegex":
		condition = ident.RegexpNotLike(el.Value)
	case "notRegexI":
		condition = ident.RegexpNotILike(el.Value)
	case "lt":
		condition = ident.Lt(el.Value)
	case "lte":
		condition = ident.Lte(el.Value)
	case "gt":
		condition = ident.Gt(el.Value)
	case "gte":
		condition = ident.Gte(el.Value)
	case "isNotNull":
		condition = ident.IsNotNull()
	case "isNull":
		condition = ident.IsNull()
	default:
		panic(fmt.Sprintf("operator %s can not be found", el.Op))
	}
	return condition, fields
}

type OrCondition struct {
	Conditions []ConditionI
}

func Or(conditions ...ConditionI) OrCondition {
	return OrCondition{Conditions: conditions}
}

func (el OrCondition) Condition(model *Model) (exp.Expression, []string) {
	var fields []string
	var conditions []exp.Expression
	for _, and := range el.Conditions {
		condition, condFields := and.Condition(model)
		fields = append(fields, condFields...)
		conditions = append(conditions, condition)
	}
	return goqu.Or(conditions...), fields
}
