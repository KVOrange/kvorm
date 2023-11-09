package kvorm

import (
	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
)

type (
	Ex     = goqu.Ex
	Record = goqu.Record
)

func I(ident string) exp.IdentifierExpression {
	return goqu.I(ident)
}

func L(sql string, args ...interface{}) exp.LiteralExpression {
	return goqu.L(sql, args...)
}
