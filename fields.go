package kvorm

import (
	"github.com/doug-martin/goqu/v9/exp"
)

type SelfFields struct {
	Model *Model
}

type ModelField struct {
	Model *Model
	Field exp.AliasedExpression
}

func (e ModelField) Count() CountExpression {
	return CountExpression{
		Field: e,
	}
}

func (e ModelField) Sub(field ModelField) SubExpression {
	return SubExpression{
		Field1: e,
		Field2: field,
	}
}

func (e ModelField) Add(field ModelField) AddExpression {
	return AddExpression{
		Field1: e,
		Field2: field,
	}
}

func (e ModelField) Mul(field ModelField) MultiplicationExpression {
	return MultiplicationExpression{
		Field1: e,
		Field2: field,
	}
}

func (e ModelField) Div(field ModelField) DivisionExpression {
	return DivisionExpression{
		Field1: e,
		Field2: field,
	}
}

func (e ModelField) ToCondition() (Identifiable, []ModelField) {
	i, _ := e.Field.Aliased().(exp.IdentifierExpression)
	return i, []ModelField{e}
}
