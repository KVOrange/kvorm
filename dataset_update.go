package kvorm

import (
	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
)

type UpdateDataset struct {
	model   *Model
	dataset *goqu.UpdateDataset
	tx      pgx.Tx
}

func (d *UpdateDataset) Where(conditions ...Condition) *UpdateDataset {
	var exps []exp.Expression
	for _, condition := range conditions {
		cond, _ := condition.Condition()
		exps = append(exps, cond)
	}
	d.dataset = d.dataset.Where(exps...)
	return d
}

func (d *UpdateDataset) Exec() error {
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

func (d *UpdateDataset) WithTx(tx pgx.Tx) *UpdateDataset {
	d.tx = tx
	return d
}

func (d *UpdateDataset) Returning(returning ...interface{}) *UpdateDataset {
	d.dataset = d.dataset.Returning(returning...)
	return d
}

func (d *UpdateDataset) Scan(dst interface{}) error {
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

func (d *UpdateDataset) ScanOne(dst interface{}) error {
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

func (d *UpdateDataset) String() string {
	query, _, _ := d.dataset.ToSQL()
	return query
}
