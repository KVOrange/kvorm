package models

import (
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

func (qb *QueryBuilder) join(dataset *goqu.SelectDataset, joiner Joiner) *goqu.SelectDataset {
	dataset = dataset.LeftJoin(joiner.Table, joiner.On)
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

//func (qb *QueryBuilder) Select(fields ...string) *SelectDataset {
//	dataset := goqu.From(qb.Model.TableName)
//	if len(fields) == 0 { // Выполнение полного Select при условии, что пользователь не указал ограничений по полям
//		dataset = qb.joinAll(dataset, qb.Model)
//		selectFields := qb.getSelectFields(qb.Model)
//		dataset = dataset.Select(selectFields...)
//	} else { // В случае если указаны некоторые поля, то выполняем частичный селект только с необходимыми полями и JOIN
//		var selectors []interface{}
//		joinedPaths := make(map[string]bool)
//		for _, field := range fields {
//			fieldSelector, ok := qb.Model.FindField(field)
//			if !ok {
//				panic(fmt.Errorf("field %s can not be found in model %s", field, qb.Model.TableName))
//			}
//
//			// Проверяем, требуется ли JOIN для данного поля
//			parts := strings.Split(field, "__")
//			if len(parts) > 1 {
//				path := parts[0]
//				for i := 1; i < len(parts); i++ {
//					if !joinedPaths[path] {
//						joiner, found := qb.Model.FindJoiner(path)
//						if !found {
//							panic(fmt.Errorf("joiner for path %s cannot be found", path))
//						}
//						dataset = qb.join(dataset, *joiner)
//						joinedPaths[path] = true
//					}
//					path += "__" + parts[i]
//				}
//			}
//
//			selectors = append(selectors, fieldSelector)
//		}
//		dataset = dataset.Select(selectors...)
//	}
//	query, _, _ := dataset.ToSQL()
//	fmt.Println(query)
//
//	return &SelectDataset{
//		Builder: qb,
//		Dataset: dataset,
//	}
//}
