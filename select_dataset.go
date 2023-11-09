package kvorm

import (
	"fmt"
	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
	"github.com/georgysavva/scany/v2/pgxscan"
	"strings"
)

type SelectDataset struct {
	Model        *Model
	Dataset      *goqu.SelectDataset
	JoinedTables map[string]bool
}

func (sd *SelectDataset) Scan(dst interface{}) error {
	query, _, _ := sd.Dataset.ToSQL()
	err := pgxscan.Select(sd.Model.DbClient.Ctx, sd.Model.DbClient.Pool, dst, query)
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
	query, _, _ := sd.Dataset.ToSQL()
	err := pgxscan.Select(sd.Model.DbClient.Ctx, sd.Model.DbClient.Pool, &dst, query)
	if err != nil {
		return nil, &SqlError{
			Err:        err,
			QueryValue: query,
		}
	}
	result := nestedMapSlice(dst)
	return result, nil
}

func (sd *SelectDataset) ScanOne(dst interface{}) error {
	query, _, _ := sd.Dataset.ToSQL()
	err := pgxscan.Get(sd.Model.DbClient.Ctx, sd.Model.DbClient.Pool, dst, query)
	if err != nil {
		return &SqlError{
			Err:        err,
			QueryValue: query,
		}
	}
	return nil
}

func (sd *SelectDataset) Where(conditions ...ConditionI) *SelectDataset {
	var exps []exp.Expression
	for _, condition := range conditions {
		cond, fields := condition.Condition(sd.Model)
		exps = append(exps, cond)
		for _, field := range fields {
			currentModel := sd.Model
			parts := strings.Split(field, SEPARATOR)
			for _, part := range parts {
				if subModel, exists := currentModel.FkModels[part]; exists {
					joiner, exists := currentModel.PreparedJoins[part]
					if !exists {
						panic(fmt.Errorf("join for submodel %s not found", part))
					}
					if !sd.JoinedTables[subModel.As] {
						sd.Dataset = sd.Model.join(sd.Dataset, joiner)
						sd.JoinedTables[subModel.As] = true
					}
					currentModel = &subModel.Model
				}
			}
		}
	}
	sd.Dataset = sd.Dataset.Where(exps...)
	return sd
}

func (sd *SelectDataset) Limit(limit uint) *SelectDataset {
	sd.Dataset = sd.Dataset.Limit(limit)
	return sd
}

func (sd *SelectDataset) Offset(offset uint) *SelectDataset {
	sd.Dataset = sd.Dataset.Offset(offset)
	return sd
}

func (sd *SelectDataset) OrderBy(field string) *SelectDataset {
	if strings.HasPrefix(field, "-") {
		trimmedField := field[1:]
		sd.Dataset = sd.Dataset.Order(goqu.I(trimmedField).Desc())
	} else {
		sd.Dataset = sd.Dataset.Order(goqu.I(field).Asc())
	}
	return sd
}

func (sd *SelectDataset) Query() *goqu.SelectDataset {
	return sd.Dataset
}

func (sd *SelectDataset) String() string {
	query, _, _ := sd.Dataset.ToSQL()
	return query
}
