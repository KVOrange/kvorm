package kvorm

import (
	"github.com/doug-martin/goqu/v9"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
)

type InsertDataset struct {
	model   *Model
	dataset *goqu.InsertDataset
	tx      pgx.Tx
}

func (d *InsertDataset) Returning(returning ...interface{}) *InsertDataset {
	d.dataset = d.dataset.Returning(returning...)
	return d
}

func (d *InsertDataset) Scan(dst interface{}) error {
	query, _, _ := d.dataset.ToSQL()
	var err error
	if d.tx != nil {
		err = pgxscan.Select(d.model.db.Ctx, d.tx, dst, query)
	} else {
		err = pgxscan.Select(d.model.db.Ctx, d.model.db.Pool, dst, query)
	}
	if err != nil {
		return &SqlError{
			Err:        err,
			QueryValue: query,
		}
	}
	return nil
}

func (d *InsertDataset) ScanOne(dst interface{}) error {
	query, _, _ := d.dataset.ToSQL()
	var err error
	if d.tx != nil {
		err = pgxscan.Get(d.model.db.Ctx, d.tx, dst, query)
	} else {
		err = pgxscan.Get(d.model.db.Ctx, d.model.db.Pool, dst, query)
	}
	if err != nil {
		return &SqlError{
			Err:        err,
			QueryValue: query,
		}
	}
	return nil
}

func (d *InsertDataset) String() string {
	query, _, _ := d.dataset.ToSQL()
	return query
}
