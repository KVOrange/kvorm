package kvorm

import (
	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
	"github.com/georgysavva/scany/v2/pgxscan"
)

type SelectDataset struct {
	model        *Model
	dataset      *goqu.SelectDataset
	joinedTables map[string]bool
}

func (sd *SelectDataset) Where(conditions ...Condition) *SelectDataset {
	var exps []exp.Expression
	for _, condition := range conditions {
		cond, fields := condition.Condition()
		exps = append(exps, cond)
		for _, field := range fields {
			dataset, joinedTables := sd.model.joinModel(field.Model, sd.dataset, sd.joinedTables)
			sd.dataset = dataset
			sd.joinedTables = joinedTables
		}
	}
	sd.dataset = sd.dataset.Where(exps...)
	return sd
}

func (sd *SelectDataset) Limit(limit uint) *SelectDataset {
	sd.dataset = sd.dataset.Limit(limit)
	return sd
}

func (sd *SelectDataset) Offset(offset uint) *SelectDataset {
	sd.dataset = sd.dataset.Offset(offset)
	return sd
}

func (sd *SelectDataset) OrderAsc(fields ...ModelField) *SelectDataset {
	for _, field := range fields {
		ident, _ := field.Field.Aliased().(exp.IdentifierExpression)
		sd.dataset = sd.dataset.OrderAppend(ident.Asc())
	}
	return sd
}

func (sd *SelectDataset) OrderDesc(fields ...ModelField) *SelectDataset {
	for _, field := range fields {
		ident, _ := field.Field.Aliased().(exp.IdentifierExpression)
		sd.dataset = sd.dataset.OrderAppend(ident.Desc())
	}
	return sd
}

func (sd *SelectDataset) Scan(dst interface{}) error {
	query, _, _ := sd.dataset.ToSQL()
	err := pgxscan.Select(sd.model.db.Ctx, sd.model.db.Pool, dst, query)
	if err != nil {
		return &SqlError{
			Err:        err,
			QueryValue: query,
		}
	}
	return nil
}

func (sd *SelectDataset) ScanOne(dst interface{}) error {
	query, _, _ := sd.dataset.ToSQL()
	err := pgxscan.Get(sd.model.db.Ctx, sd.model.db.Pool, dst, query)
	if err != nil {
		return &SqlError{
			Err:        err,
			QueryValue: query,
		}
	}
	return nil
}

func (sd *SelectDataset) MapScan() ([]map[string]interface{}, error) {
	var dst []map[string]interface{}
	query, _, _ := sd.dataset.ToSQL()
	err := pgxscan.Select(sd.model.db.Ctx, sd.model.db.Pool, &dst, query)
	if err != nil {
		return nil, &SqlError{
			Err:        err,
			QueryValue: query,
		}
	}
	result := nestedMapSlice(dst)
	return result, nil
}

func (sd *SelectDataset) Query() *goqu.SelectDataset {
	return sd.dataset
}

func (sd *SelectDataset) String() string {
	query, _, _ := sd.dataset.ToSQL()
	return query
}
