package models

import (
	"fmt"
	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
	"kvorm_lib/libraries/kvorm/database"
	"reflect"
	"strings"
)

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
	TableName string             `json:"-"`
	PK        string             `json:"-"`
	DbClient  *database.DbClient `json:"-"`

	FieldMap map[string]DbField `json:"-"`
	FkModels map[string]FkModel `json:"-"`

	PreparedJoins     map[string]Joiner                `json:"-"`
	PreparedSelectors map[string]exp.AliasedExpression `json:"-"`
}

func (m *Model) ExtractTableName(v interface{}) {
	val := reflect.ValueOf(v).Elem()
	typeOfT := val.Type()
	// Ищем поле с именем "Model" и извлекаем тег "table"
	for i := 0; i < val.NumField(); i++ {
		field := typeOfT.Field(i)
		if field.Name == "Model" && field.Type == reflect.TypeOf(Model{}) {
			m.TableName = field.Tag.Get("table")
			break
		}
	}
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

	rValue := reflect.ValueOf(model).Elem()
	rType := rValue.Type()
	for i := 0; i < rType.NumField(); i++ {
		isPkField := false
		field := rType.Field(i)
		if field.Name == "Model" {
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
		tag := field.Tag.Get("db_name")
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
			m.PK = asNameValue
		}
	}

	m.PrepareJoins("")
	m.PrepareSelectors("", "")
}

func (m *Model) InitConnection(client *database.DbClient) {
	m.DbClient = client
}

func (m *Model) PrepareJoins(modelAs string) {
	for _, fkModel := range m.FkModels {
		tableAsName := fmt.Sprintf("%s_%s", modelAs, fkModel.As)
		if modelAs == "" {
			tableAsName = fmt.Sprintf("%s_%s", m.TableName, fkModel.As)
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
		}
		for _, fkModel := range m.FkModels {
			tableAsName := fmt.Sprintf("%s_%s", m.TableName, fkModel.As)
			fkModel.Model.PrepareSelectors(fkModel.As, tableAsName)
		}
	} else {
		for _, field := range m.FieldMap {
			m.PreparedSelectors[field.AsValue] = goqu.I(fmt.Sprintf("%s.%s", tableName, field.StringValue)).As(goqu.S(fmt.Sprintf("%s.%s", prefix, field.AsValue)))
		}
		for _, fkModel := range m.FkModels {
			tableAsName := fmt.Sprintf("%s_%s", tableName, fkModel.As)
			fkModel.Model.PrepareSelectors(fmt.Sprintf("%s.%s", prefix, fkModel.As), tableAsName)
		}
	}

}

func (m *Model) FindField(input string) (exp.AliasedExpression, bool) {
	parts := strings.Split(input, "__")
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

func (m *Model) FindJoiner(path string) (*Joiner, bool) {
	parts := strings.Split(path, "__")

	if joiner, exists := m.PreparedJoins[parts[0]]; exists {
		if len(parts) == 1 {
			return &joiner, true
		}

		// Если у нас есть FKModel для этого имени, ищем далее
		if nextModel, hasModel := m.FkModels[parts[0]]; hasModel {
			return nextModel.Model.FindJoiner(strings.Join(parts[1:], "__"))
		}
	}
	return nil, false
}

func (m *Model) joinAll(dataset *goqu.SelectDataset) *goqu.SelectDataset {
	for _, joiner := range m.PreparedJoins {
		dataset = dataset.LeftJoin(joiner.Table, joiner.On)
	}
	for _, fkModel := range m.FkModels {
		dataset = fkModel.Model.joinAll(dataset)
	}
	return dataset
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

func (m *Model) Select(fields ...string) *SelectDataset {
	dataset := goqu.From(m.TableName)
	if len(fields) == 0 { // Выполнение полного Select при условии, что пользователь не указал ограничений по полям
		dataset = m.joinAll(dataset)
		selectFields := m.getSelectFields()
		dataset = dataset.Select(selectFields...)
	} else { // В случае если указаны некоторые поля, то выполняем частичный селект только с необходимыми полями и JOIN
		var selectors []interface{}
		joinedPaths := make(map[string]bool)
		for _, field := range fields {
			fieldSelector, ok := m.FindField(field)
			if !ok {
				panic(fmt.Errorf("field %s can not be found in model %s", field, m.TableName))
			}

			// Проверяем, требуется ли JOIN для данного поля
			parts := strings.Split(field, "__")
			if len(parts) > 1 {
				path := parts[0]
				for i := 1; i < len(parts); i++ {
					if !joinedPaths[path] {
						joiner, found := m.FindJoiner(path)
						if !found {
							panic(fmt.Errorf("joiner for path %s cannot be found", path))
						}
						dataset = m.join(dataset, *joiner)
						joinedPaths[path] = true
					}
					path += "__" + parts[i]
				}
			}

			selectors = append(selectors, fieldSelector)
		}
		dataset = dataset.Select(selectors...)
	}
	return &SelectDataset{
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
