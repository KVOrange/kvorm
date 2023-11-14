package kvorm

import (
	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
)

type CountExpression struct {
	Field ModelField
}

type LiteralExpression struct {
	Fields []ModelField
	Format string
}

func (e LiteralExpression) Sql() exp.LiteralExpression {
	var aliasedFields []interface{}
	for _, field := range e.Fields {
		aliasedFields = append(aliasedFields, field.Field.Aliased())
	}
	return goqu.L(e.Format, aliasedFields...)
}

func (e LiteralExpression) ToCondition() (Identifiable, []ModelField) {
	return e.Sql(), e.Fields
}

type CalculationExpression interface {
	Sql() exp.LiteralExpression
	Fields() (ModelField, ModelField)
}

type SubExpression struct {
	Field1 ModelField
	Field2 ModelField
}

func (e SubExpression) Sql() exp.LiteralExpression {
	return goqu.L("? - ?", e.Field1.Field.Aliased(), e.Field2.Field.Aliased())
}

func (e SubExpression) Fields() (ModelField, ModelField) {
	return e.Field1, e.Field2
}

func (e SubExpression) ToCondition() (Identifiable, []ModelField) {
	return e.Sql(), []ModelField{e.Field1, e.Field2}
}

type AddExpression struct {
	Field1 ModelField
	Field2 ModelField
}

func (e AddExpression) Sql() exp.LiteralExpression {
	return goqu.L("? + ?", e.Field1.Field.Aliased(), e.Field2.Field.Aliased())
}

func (e AddExpression) Fields() (ModelField, ModelField) {
	return e.Field1, e.Field2
}

func (e AddExpression) ToCondition() (Identifiable, []ModelField) {
	return e.Sql(), []ModelField{e.Field1, e.Field2}
}

type MultiplicationExpression struct {
	Field1 ModelField
	Field2 ModelField
}

func (e MultiplicationExpression) Sql() exp.LiteralExpression {
	return goqu.L("? * ?", e.Field1.Field.Aliased(), e.Field2.Field.Aliased())
}

func (e MultiplicationExpression) Fields() (ModelField, ModelField) {
	return e.Field1, e.Field2
}

func (e MultiplicationExpression) ToCondition() (Identifiable, []ModelField) {
	return e.Sql(), []ModelField{e.Field1, e.Field2}
}

type DivisionExpression struct {
	Field1 ModelField
	Field2 ModelField
}

func (e DivisionExpression) Sql() exp.LiteralExpression {
	return goqu.L("? / ?", e.Field1.Field.Aliased(), e.Field2.Field.Aliased())
}

func (e DivisionExpression) Fields() (ModelField, ModelField) {
	return e.Field1, e.Field2
}

func (e DivisionExpression) ToCondition() (Identifiable, []ModelField) {
	return e.Sql(), []ModelField{e.Field1, e.Field2}
}
