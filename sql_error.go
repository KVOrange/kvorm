package kvorm

type SqlError struct {
	Err        error
	QueryValue string
}

func (e *SqlError) Error() string {
	return e.Err.Error()
}

func (e *SqlError) Query() string {
	return e.QueryValue
}
