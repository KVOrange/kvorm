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
	TableName string
	DbClient  *database.DbClient

	FieldMap map[string]DbField
	FkModels []FkModel

	PreparedJoins     map[string]Joiner
	PreparedSelectors map[string]exp.AliasedExpression
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
	if m.PreparedJoins == nil {
		m.PreparedJoins = make(map[string]Joiner)
	}
	if m.PreparedSelectors == nil {
		m.PreparedSelectors = make(map[string]exp.AliasedExpression)
	}

	rValue := reflect.ValueOf(model).Elem()
	rType := rValue.Type()

	for i := 0; i < rType.NumField(); i++ {
		field := rType.Field(i)
		if field.Name == "Model" {
			continue
		}

		// Обработка FK
		dbTypeTag := field.Tag.Get("fk")
		if dbTypeTag != "" {
			asNameValue := field.Tag.Get("db")
			values := strings.Split(dbTypeTag, ",")
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

						m.FkModels = append(m.FkModels, fkModel) // Добавляем инициализированную модель
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

func (m *Model) Query() *QueryBuilder {
	return &QueryBuilder{
		Model:    m,
		DbClient: m.DbClient,
	}
}
