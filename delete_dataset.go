package kvorm

import (
	"fmt"
	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
)

type DeleteDataset struct {
	Model   *Model
	Dataset *goqu.DeleteDataset
}

func (ds *DeleteDataset) Where(conditions ...ConditionI) *DeleteDataset {
	var exps []exp.Expression
	for _, condition := range conditions {
		cond, fields := condition.Condition(ds.Model)
		exps = append(exps, cond)
		for _, field := range fields {
			_, ok := ds.Model.PreparedSelectors[field]
			if !ok {
				panic(fmt.Sprintf("field %s not found in model %s", field, ds.Model.TableName))
			}
		}
	}
	ds.Dataset = ds.Dataset.Where(exps...)
	return ds
}

func (ds *DeleteDataset) String() string {
	query, _, _ := ds.Dataset.ToSQL()
	return query
}

func (ds *DeleteDataset) Exec() error {
	query, _, _ := ds.Dataset.ToSQL()
	_, err := ds.Model.DbClient.Pool.Exec(ds.Model.DbClient.Ctx, query)
	if err != nil {
		return &SqlError{
			Err:        err,
			QueryValue: query,
		}
	}
	return nil
}
