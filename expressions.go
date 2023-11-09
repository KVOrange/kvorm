package kvorm

import (
	"fmt"
	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
)

type CountExpression struct {
	Field      string
	Expression exp.SQLFunctionExpression
}

func Count(col string) CountExpression {
	return CountExpression{col, goqu.COUNT(col)}
}

type SubExpression struct {
	Field1 string
	Field2 string
}

func Sub(field1, field2 string) SubExpression {
	return SubExpression{field1, field2}
}

type AvgExpression struct {
	Fields []string
	Sql    string
}

func Avg(expression interface{}) AvgExpression {
	switch expression.(type) {
	case string:
		return AvgExpression{[]string{expression.(string)}, "?"}
	case SubExpression:
		subExp := expression.(SubExpression)
		return AvgExpression{[]string{subExp.Field1, subExp.Field2}, "? - ?"}
	default:
		panic(fmt.Sprintf("unsupported type %T in kvorm.Avg", expression))
	}
}
