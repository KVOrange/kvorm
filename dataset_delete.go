package kvorm

import (
	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
)

type DeleteDataset struct {
	model   *Model
	dataset *goqu.DeleteDataset
	tx      pgx.Tx
}

func (d *DeleteDataset) Where(conditions ...Condition) *DeleteDataset {
	var exps []exp.Expression
	for _, condition := range conditions {
		cond, _ := condition.Condition()
		exps = append(exps, cond)
	}
	d.dataset = d.dataset.Where(exps...)
	return d
}

func (d *DeleteDataset) Exec() error {
	query, _, _ := d.dataset.ToSQL()
	var err error
	if d.tx != nil {
		_, err = d.tx.Exec(d.model.db.Ctx, query)
	} else {
		_, err = d.model.db.Pool.Exec(d.model.db.Ctx, query)
	}

	if err != nil {
		return &SqlError{
			Err:        err,
			QueryValue: query,
		}
	}
	return nil
}

func (d *DeleteDataset) WithTx(tx pgx.Tx) *DeleteDataset {
	d.tx = tx
	return d
}

func (d *DeleteDataset) Returning(returning ...interface{}) *DeleteDataset {
	d.dataset = d.dataset.Returning(returning...)
	return d
}

func (d *DeleteDataset) Scan(dst interface{}) error {
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

func (d *DeleteDataset) ScanOne(dst interface{}) error {
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

func (d *DeleteDataset) String() string {
	query, _, _ := d.dataset.ToSQL()
	return query
}
