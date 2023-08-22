package models

import (
	"fmt"
	"github.com/doug-martin/goqu/v9"
	"kvorm_lib/libraries/kvorm/database"
)

type QueryBuilder struct {
	Model    *Model
	DbClient *database.DbClient
}

func (qb *QueryBuilder) joinAll(dataset *goqu.SelectDataset, model *Model) *goqu.SelectDataset {
	for _, joiner := range model.PreparedJoins {
		dataset = dataset.LeftJoin(joiner.Table, joiner.On)
	}
	for _, fkModel := range model.FkModels {
		dataset = qb.joinAll(dataset, &fkModel.Model)
	}
	return dataset
}

func (qb *QueryBuilder) getSelectFields(model *Model) []interface{} {
	var result []interface{}
	for _, field := range model.PreparedSelectors {
		result = append(result, field)
	}
	for _, fkModel := range model.FkModels {
		result = append(result, qb.getSelectFields(&fkModel.Model)...)
	}
	return result
}

func (qb *QueryBuilder) Select() *SelectDataset {
	dataset := goqu.From(qb.Model.TableName)
	dataset = qb.joinAll(dataset, qb.Model)

	selectFields := qb.getSelectFields(qb.Model)
	dataset = dataset.Select(selectFields...)
	query, _, _ := dataset.ToSQL()
	fmt.Println(query)

	return &SelectDataset{
		Builder: qb,
		Dataset: dataset,
	}
}
