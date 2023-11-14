package kvorm

const (
	In        = "in"
	NotIn     = "notIn"
	Eq        = "eq"
	NotEq     = "notEq"
	Like      = "like"
	NotLike   = "notLike"
	Regex     = "regex"
	RegexI    = "regexI"
	NotRegex  = "notRegex"
	NotRegexI = "notRegexI"
	Lt        = "lt"
	Lte       = "lte"
	Gt        = "gt"
	Gte       = "gte"
	IsNotNull = "isNotNull"
	IsNull    = "isNull"
)

func L(sql string, fields ...ModelField) LiteralExpression {
	return LiteralExpression{
		Fields: fields,
		Format: sql,
	}
}

func Or(conditions ...Condition) OrCondition {
	return OrCondition{Conditions: conditions}
}
