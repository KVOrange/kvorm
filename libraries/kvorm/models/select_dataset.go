package models

import (
	"github.com/doug-martin/goqu/v9"
	"github.com/georgysavva/scany/v2/pgxscan"
)

type SelectDataset struct {
	Builder *QueryBuilder
	Dataset *goqu.SelectDataset
}

func (sd *SelectDataset) Scan(dst interface{}) error {
	query, _, _ := sd.Dataset.ToSQL()
	err := pgxscan.Select(sd.Builder.DbClient.Ctx, sd.Builder.DbClient.Pool, dst, query)
	return err
}

func (sd *SelectDataset) Where() {

}
