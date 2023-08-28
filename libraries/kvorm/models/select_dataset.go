package models

import (
	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
	"github.com/georgysavva/scany/v2/pgxscan"
)

type SelectDataset struct {
	Model   *Model
	Dataset *goqu.SelectDataset
}

func (sd *SelectDataset) Scan(dst interface{}) error {
	query, _, _ := sd.Dataset.ToSQL()
	err := pgxscan.Select(sd.Model.DbClient.Ctx, sd.Model.DbClient.Pool, dst, query)
	return err
}

func (sd *SelectDataset) ScanOne(dst interface{}) error {
	query, _, _ := sd.Dataset.ToSQL()
	err := pgxscan.Get(sd.Model.DbClient.Ctx, sd.Model.DbClient.Pool, dst, query)
	return err
}

func (sd *SelectDataset) Where(expressions ...exp.Expression) *SelectDataset {
	sd.Dataset = sd.Dataset.Where(expressions...)
	return sd
}

func (sd *SelectDataset) Query() *goqu.SelectDataset {
	return sd.Dataset
}

func (sd *SelectDataset) String() string {
	query, _, _ := sd.Dataset.ToSQL()
	return query
}
