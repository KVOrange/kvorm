package kvorm

import (
	"fmt"
	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"reflect"
	"strings"
)

const SEPARATOR = "__"

type DbField struct {
	StringValue string
	AsValue     string
}

type FkModel struct {
	As        string
	FromField string
	ToField   string
	Model     Model
}

type Joiner struct {
	Table exp.AliasedExpression
	On    exp.JoinCondition
}

type Model struct {
	TableName string    `json:"-"`
	PK        string    `json:"-"`
	DbClient  *DbClient `json:"-"`

	FieldMap map[string]DbField `json:"-"`
	FkModels map[string]FkModel `json:"-"`

	PreparedJoins              map[string]Joiner                   `json:"-"`
	PreparedSelectors          map[string]exp.AliasedExpression    `json:"-"`
	PreparedSelectorsWithoutAs map[string]exp.IdentifierExpression `json:"-"`
}

func (m *Model) InitFields(model interface{}) {
	if m.FieldMap == nil {
		m.FieldMap = make(map[string]DbField)
	}
	if m.FkModels == nil {
		m.FkModels = make(map[string]FkModel)
	}
	if m.PreparedJoins == nil {
		m.PreparedJoins = make(map[string]Joiner)
	}
	if m.PreparedSelectors == nil {
		m.PreparedSelectors = make(map[string]exp.AliasedExpression)
	}
	if m.PreparedSelectorsWithoutAs == nil {
		m.PreparedSelectorsWithoutAs = make(map[string]exp.IdentifierExpression)
	}

	rValue := reflect.ValueOf(model).Elem()
	rType := rValue.Type()
	for i := 0; i < rType.NumField(); i++ {
		isPkField := false
		field := rType.Field(i)
		if field.Name == "Model" && field.Type == reflect.TypeOf(Model{}) {
			m.TableName = field.Tag.Get("table")
			continue
		}

		typeTag := field.Tag.Get("type")
		if typeTag != "" {
			if typeTag == "virtual" {
				continue
			}
			if typeTag == "pk" {
				isPkField = true
			}
		}

		// Обработка FK
		dbFkTag := field.Tag.Get("fk")
		if dbFkTag != "" {
			asNameValue := field.Tag.Get("db")
			values := strings.Split(dbFkTag, ",")
			if len(values) == 2 {
				fkModelInterface := rValue.Field(i).Addr().Interface() // Получение интерфейса из reflect.Value
				InitTable(fkModelInterface, m.DbClient)                // Инициализация с интерфейсом

				nestedModelField := reflect.ValueOf(fkModelInterface).Elem().FieldByName("Model")
				if nestedModelField.IsValid() && nestedModelField.CanAddr() {
					nestedModel, ok := nestedModelField.Addr().Interface().(*Model) // Получение указателя на Model
					if ok {
						var fkModel FkModel
						fkModel.FromField = values[0]
						fkModel.ToField = values[1]
						fkModel.Model = *nestedModel
						if asNameValue != "" {
							fkModel.As = asNameValue
						} else {
							fkModel.As = fkModel.Model.TableName
						}

						m.FkModels[fkModel.As] = fkModel // Добавляем инициализированную модель
					}
				}
				continue
			}
		}

		// Проверка на тег "db_name"
		tag := field.Tag.Get("db")
		var dbField DbField
		fieldName := ""
		if tag != "" {
			fieldName = tag
		} else {
			fieldName = toSnakeCase(field.Name)
		}
		asNameValue := field.Tag.Get("db")
		dbField.StringValue = fieldName
		if asNameValue != "" {
			dbField.AsValue = asNameValue
		} else {
			dbField.AsValue = fieldName
		}
		m.FieldMap[field.Name] = dbField
		if isPkField {
			if asNameValue != "" {
				m.PK = asNameValue
			} else {
				m.PK = fieldName
			}
		}
	}

	if m.PK == "" {
		panic(fmt.Sprintf("Error in init %s table. Not found field with pk type.", m.TableName))
	}
	m.PrepareJoins("")
	m.PrepareSelectors("", "")
}

func (m *Model) InitConnection(client *DbClient) {
	m.DbClient = client
}

func (m *Model) PrepareJoins(modelAs string) {
	for _, fkModel := range m.FkModels {
		tableAsName := fmt.Sprintf("%s%s%s", modelAs, SEPARATOR, fkModel.As)
		if modelAs == "" {
			tableAsName = fmt.Sprintf("%s%s%s", m.TableName, SEPARATOR, fkModel.As)
			modelAs = m.TableName
		}
		var joiner Joiner
		joiner.Table = goqu.T(fkModel.Model.TableName).As(tableAsName)
		joiner.On = goqu.On(
			goqu.Ex{
				fmt.Sprintf("%s.%s", modelAs, fkModel.FromField): goqu.I(fmt.Sprintf("%s.%s", tableAsName, fkModel.ToField)),
			},
		)
		m.PreparedJoins[fkModel.As] = joiner
		fkModel.Model.PrepareJoins(tableAsName)
	}
}

func (m *Model) PrepareSelectors(prefix, tableName string) {
	if prefix == "" && tableName == "" {
		for _, field := range m.FieldMap {
			m.PreparedSelectors[field.AsValue] = goqu.I(fmt.Sprintf("%s.%s", m.TableName, field.StringValue)).As(field.AsValue)
			m.PreparedSelectorsWithoutAs[field.AsValue] = goqu.I(fmt.Sprintf("%s.%s", m.TableName, field.StringValue))
		}
		for _, fkModel := range m.FkModels {
			tableAsName := fmt.Sprintf("%s%s%s", m.TableName, SEPARATOR, fkModel.As)
			fkModel.Model.PrepareSelectors(fkModel.As, tableAsName)
		}
	} else {
		for _, field := range m.FieldMap {
			m.PreparedSelectors[field.AsValue] = goqu.I(fmt.Sprintf("%s.%s", tableName, field.StringValue)).As(goqu.S(fmt.Sprintf("%s.%s", prefix, field.AsValue)))
			m.PreparedSelectorsWithoutAs[field.AsValue] = goqu.I(fmt.Sprintf("%s.%s", tableName, field.StringValue))
		}
		for _, fkModel := range m.FkModels {
			tableAsName := fmt.Sprintf("%s%s%s", tableName, SEPARATOR, fkModel.As)
			fkModel.Model.PrepareSelectors(fmt.Sprintf("%s.%s", prefix, fkModel.As), tableAsName)
		}
	}

}

func (m *Model) FindField(input string) (exp.AliasedExpression, bool) {
	parts := strings.Split(input, SEPARATOR)
	return m.findFieldRecursive(parts)
}

func (m *Model) findFieldRecursive(parts []string) (exp.AliasedExpression, bool) {
	// Если у нас всего одна часть, это конечное поле.
	if len(parts) == 1 {
		field, exists := m.PreparedSelectors[parts[0]]
		return field, exists
	}

	// Иначе, ищем вложенную модель и рекурсивно ищем в ней.
	if nextModel, exists := m.FkModels[parts[0]]; exists {
		return nextModel.Model.findFieldRecursive(parts[1:])
	}

	// Если мы здесь, это означает, что вложенная модель не найдена.
	return nil, false
}

func (m *Model) FindFieldWithoutAs(input string) (exp.IdentifierExpression, bool) {
	parts := strings.Split(input, SEPARATOR)
	return m.findFieldWithoutAsRecursive(parts)
}

func (m *Model) findFieldWithoutAsRecursive(parts []string) (exp.IdentifierExpression, bool) {
	// Если у нас всего одна часть, это конечное поле.
	if len(parts) == 1 {
		field, exists := m.PreparedSelectorsWithoutAs[parts[0]]
		return field, exists
	}

	// Иначе, ищем вложенную модель и рекурсивно ищем в ней.
	if nextModel, exists := m.FkModels[parts[0]]; exists {
		return nextModel.Model.findFieldWithoutAsRecursive(parts[1:])
	}

	// Если мы здесь, это означает, что вложенная модель не найдена.
	return nil, false
}

func (m *Model) FindJoiner(path string) (*Joiner, bool) {
	parts := strings.Split(path, SEPARATOR)

	if joiner, exists := m.PreparedJoins[parts[0]]; exists {
		if len(parts) == 1 {
			return &joiner, true
		}

		// Если у нас есть FKModel для этого имени, ищем далее
		if nextModel, hasModel := m.FkModels[parts[0]]; hasModel {
			return nextModel.Model.FindJoiner(strings.Join(parts[1:], SEPARATOR))
		}
	}
	return nil, false
}

func (m *Model) joinAll(dataset *goqu.SelectDataset, joinTables map[string]bool) (*goqu.SelectDataset, map[string]bool) {
	for _, joiner := range m.PreparedJoins {
		dataset = dataset.LeftJoin(joiner.Table, joiner.On)
	}
	for _, fkModel := range m.FkModels {
		joinTables[fkModel.As] = true
		dataset, joinTables = fkModel.Model.joinAll(dataset, joinTables)
	}
	return dataset, joinTables
}

func (m *Model) join(dataset *goqu.SelectDataset, joiner Joiner) *goqu.SelectDataset {
	dataset = dataset.LeftJoin(joiner.Table, joiner.On)
	return dataset
}

func (m *Model) getSelectFields() []interface{} {
	var result []interface{}
	for _, field := range m.PreparedSelectors {
		result = append(result, field)
	}
	for _, fkModel := range m.FkModels {
		result = append(result, fkModel.Model.getSelectFields()...)
	}
	return result
}

func (m *Model) Select(fields ...interface{}) *SelectDataset {
	dataset := goqu.From(m.TableName)
	joinedTables := make(map[string]bool)
	if len(fields) == 0 {
		// Выполнение полного Select при условии, что пользователь не указал ограничений по полям
		dataset, joinedTables = m.joinAll(dataset, joinedTables)
		selectFields := m.getSelectFields()
		dataset = dataset.Select(selectFields...)
	} else {
		// В случае если указаны некоторые поля, то выполняем частичный селект только с необходимыми полями и JOIN
		var resultFields []interface{}
		for _, interfaceField := range fields {
			switch interfaceField.(type) {
			case []string:
				sliceFields := interfaceField.([]string)
				for _, strField := range sliceFields {
					resultFields = append(resultFields, strField)
				}
				continue
			default:
				resultFields = append(resultFields, interfaceField)
			}
		}
		var selectors []interface{}
		for _, interfaceField := range resultFields {
			field := ""
			isString := false
			isCount := false
			switch interfaceField.(type) {
			case string:
				isString = true
				field = interfaceField.(string)
			case CountExpression:
				isCount = true
				field = interfaceField.(CountExpression).Field
			case AvgExpression:
				var avgFieldSelectors []interface{}
				avgExp := interfaceField.(AvgExpression)
				avgFields := avgExp.Fields
				for _, avgField := range avgFields {
					currentModel := m
					parts := strings.Split(avgField, SEPARATOR)
					for idx, part := range parts {
						if subModel, exists := currentModel.FkModels[part]; exists {
							joiner, exists := currentModel.PreparedJoins[part]
							if !exists {
								panic(fmt.Errorf("join for submodel %s not found", part))
							}
							if !joinedTables[subModel.As] {
								dataset = m.join(dataset, joiner)
								joinedTables[subModel.As] = true
							}
							currentModel = &subModel.Model
						} else {
							if idx == len(parts)-1 {
								fieldSelector, ok := currentModel.FindFieldWithoutAs(part)
								if !ok {
									panic(fmt.Errorf("field %s not found in model %s", part, currentModel.TableName))
								}
								avgFieldSelectors = append(avgFieldSelectors, fieldSelector)
							} else {
								panic(fmt.Errorf("invalid part %s in field %s", part, field))
							}
						}
					}
				}
				selectors = append(selectors, goqu.AVG(goqu.L(avgExp.Sql, avgFieldSelectors...)))
				continue
			default:
				panic(fmt.Sprintf("unsupported type %T", interfaceField))
			}

			if field == "self" {
				if !isString {
					panic(fmt.Sprintf("unsupported type %T with value self", interfaceField))
				}
				for _, selector := range m.PreparedSelectors {
					selectors = append(selectors, selector.(interface{}))
				}
				continue
			}

			currentModel := m
			parts := strings.Split(field, SEPARATOR)
			isSubmodelPath := true // предположим, что это путь к подмодели
			var lastModel *Model
			for idx, part := range parts {
				// Проверим, является ли текущий элемент подмоделью
				if subModel, exists := currentModel.FkModels[part]; exists {
					lastModel = &subModel.Model
					// Добавить JOIN для подмодели
					joiner, exists := currentModel.PreparedJoins[part]
					if !exists {
						panic(fmt.Errorf("join for submodel %s not found", part))
					}
					if !joinedTables[subModel.As] {
						dataset = m.join(dataset, joiner)
						joinedTables[subModel.As] = true
					}
					currentModel = &subModel.Model
				} else {
					// Если это последний элемент, и он не является подмоделью
					if idx == len(parts)-1 {
						isSubmodelPath = false
						if isString {
							fieldSelector, ok := currentModel.FindField(part)
							if !ok {
								panic(fmt.Errorf("field %s not found in model %s", part, currentModel.TableName))
							}
							selectors = append(selectors, fieldSelector)

						} else {
							if isCount {
								fieldSelector, ok := currentModel.FindFieldWithoutAs(part)
								if !ok {
									panic(fmt.Errorf("field %s not found in model %s", part, currentModel.TableName))
								}
								selectors = append(selectors, goqu.COUNT(fieldSelector))
							}
						}
					} else {
						panic(fmt.Errorf("invalid part %s in field %s", part, field))
					}
				}
			}
			if isSubmodelPath && lastModel != nil && isString {
				// Добавляем все поля из этой подмодели
				for _, selector := range lastModel.PreparedSelectors {
					selectors = append(selectors, selector)
				}
			}
		}
		dataset = dataset.Select(selectors...)
	}
	return &SelectDataset{
		Model:        m,
		Dataset:      dataset,
		JoinedTables: joinedTables,
	}
}

func (m *Model) Delete() *DeleteDataset {
	dataset := goqu.Delete(m.TableName)
	return &DeleteDataset{
		Model:   m,
		Dataset: dataset,
	}
}

func (m *Model) Update(values interface{}) *UpdateDataset {
	dataset := goqu.Update(m.TableName).Set(values)
	return &UpdateDataset{
		Model:   m,
		Dataset: dataset,
	}
}

func (m *Model) Save(instance interface{}) error {
	rValue := reflect.ValueOf(instance).Elem()
	rType := rValue.Type()
	tableName := ""
	pkName := ""
	var pkValue interface{}
	fields := goqu.Record{}
	for i := 0; i < rType.NumField(); i++ {
		isPkField := false
		field := rType.Field(i)
		if field.Name == "Model" && field.Type == reflect.TypeOf(Model{}) {
			tableName = field.Tag.Get("table")
			continue
		}

		typeTag := field.Tag.Get("type")
		if typeTag != "" {
			if typeTag == "virtual" {
				continue
			}
			if typeTag == "pk" {
				isPkField = true
			}
		}

		dbFkTag := field.Tag.Get("fk")
		if dbFkTag != "" {
			continue
		}

		fieldName := ""
		nameTag := field.Tag.Get("db_name")
		if nameTag != "" {
			fieldName = nameTag
		} else {
			fieldName = toSnakeCase(field.Name)
		}
		if isPkField {
			pkName = fieldName
			pkValue = rValue.Field(i).Interface()
		} else {
			fields[fieldName] = rValue.Field(i).Interface()
		}
	}
	query, _, _ := goqu.Update(tableName).Set(fields).Where(goqu.Ex{pkName: pkValue}).ToSQL()
	_, err := m.DbClient.Pool.Exec(m.DbClient.Ctx, query)
	return err
}

func (m *Model) Add(instance interface{}) error {
	rValue := reflect.ValueOf(instance).Elem()
	rType := rValue.Type()
	tableName := ""
	fields := goqu.Record{}
	for i := 0; i < rType.NumField(); i++ {
		field := rType.Field(i)
		if field.Name == "Model" && field.Type == reflect.TypeOf(Model{}) {
			tableName = field.Tag.Get("table")
			continue
		}

		typeTag := field.Tag.Get("type")
		if typeTag != "" {
			if typeTag == "virtual" {
				continue
			}
			if typeTag == "pk" {
				continue
			}
		}

		dbFkTag := field.Tag.Get("fk")
		if dbFkTag != "" {
			continue
		}

		fieldName := ""
		nameTag := field.Tag.Get("db_name")
		if nameTag != "" {
			fieldName = nameTag
		} else {
			fieldName = toSnakeCase(field.Name)
		}
		fields[fieldName] = rValue.Field(i).Interface()
	}
	query, _, _ := goqu.Insert(tableName).Rows(fields).ToSQL()
	_, err := m.DbClient.Pool.Exec(m.DbClient.Ctx, query)
	return err
}

func (m *Model) Insert(rows ...interface{}) ([]int64, error) {
	var result []int64
	query, _, err := goqu.Insert(m.TableName).Rows(rows...).Returning(m.PK).ToSQL()
	if err != nil {
		return result, &SqlError{
			Err:        err,
			QueryValue: query,
		}
	}
	err = pgxscan.Select(m.DbClient.Ctx, m.DbClient.Pool, &result, query)
	if err != nil {
		return result, &SqlError{
			Err:        err,
			QueryValue: query,
		}
	}
	return result, nil
}

func (m *Model) StartTx() (pgx.Tx, error) {
	return m.DbClient.Pool.Begin(m.DbClient.Ctx)
}

func (m *Model) InsertTx(tx pgx.Tx, rows ...interface{}) ([]int64, error) {
	var result []int64
	query, _, err := goqu.Insert(m.TableName).Rows(rows...).Returning(m.PK).ToSQL()
	if err != nil {
		return result, err
	}
	err = pgxscan.Select(m.DbClient.Ctx, tx, &result, query)
	return result, err
}
