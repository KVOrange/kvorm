package kvorm

import (
	"fmt"
	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
	"reflect"
	"strings"
)

type Model struct {
	db     *DbClient
	parent *Model

	tableName string
	asName    string
	pk        string
	dbFields  []string

	fkModels          map[string]fkModel
	preparedSelectors map[string]exp.AliasedExpression
	preparedJoins     map[string]joiner
}

func (m *Model) Init(db *DbClient, model interface{}) error {
	m.db = db
	m.preparedSelectors = make(map[string]exp.AliasedExpression)
	m.fkModels = make(map[string]fkModel)
	m.preparedJoins = make(map[string]joiner)

	rValue := reflect.ValueOf(model).Elem()
	rType := rValue.Type()

	for i := 0; i < rType.NumField(); i++ {
		field := rType.Field(i)

		// Проверка тега table
		if field.Name == "Model" && field.Type == reflect.TypeOf(Model{}) {
			m.tableName = field.Tag.Get("table")
			continue
		}

		// Проверка тега type
		isPkField := false
		typeTag := field.Tag.Get("type")
		if typeTag != "" {
			if typeTag == "virtual" {
				continue
			}
			if typeTag == "pk" {
				isPkField = true
			}
		}

		dbTag := field.Tag.Get("db")

		// Проверка тега fk
		fkTag := field.Tag.Get("fk")
		if fkTag != "" {
			if dbTag == "" {
				return fmt.Errorf("error in init table: not found db tag with fk %s", fkTag)
			}
			fkValues := strings.Split(fkTag, ",")
			if len(fkValues) != 2 {
				return fmt.Errorf("error in init table: uncorrect value in fk tag. Expected from_field,to_field. Goted: %s", fkTag)
			}

			fkModelInterface, ok := rValue.Field(i).Addr().Interface().(modelI)
			if !ok {
				return fmt.Errorf("error in init table: fk field %s is not a model", dbTag)
			}
			err := fkModelInterface.Init(db, fkModelInterface)
			if err != nil {
				return err
			}

			nestedModelField := reflect.ValueOf(fkModelInterface).Elem().FieldByName("Model")
			nestedModel, ok := nestedModelField.Addr().Interface().(*Model)
			if !ok {
				return fmt.Errorf("error in init table: fk field %s is not a model", dbTag)
			}
			nestedModel.asName = dbTag
			nestedModel.parent = m

			var fkM fkModel
			fkM.FromField = fkValues[0]
			fkM.ToField = fkValues[1]
			fkM.Model = *nestedModel
			fkM.As = dbTag
			m.fkModels[fkM.As] = fkM
			continue
		}

		// Проверка тега db
		fieldName := ""
		if dbTag != "" {
			fieldName = dbTag
		} else {
			fieldName = toSnakeCase(field.Name)
		}
		m.dbFields = append(m.dbFields, fieldName)
		if isPkField {
			m.pk = fieldName
		}
	}

	if m.tableName == "" {
		return fmt.Errorf("error in init table: table name not found")
	}
	if m.pk == "" {
		return fmt.Errorf("error in init %s table: field with pk type not found", m.tableName)
	}

	m.initJoins("")
	m.initSelectors("", "")
	return nil
}

func (m *Model) Self() SelfFields {
	return SelfFields{
		Model: m,
	}
}

func (m *Model) Field(name string) ModelField {
	field, ok := m.preparedSelectors[name]
	if !ok {
		panic(fmt.Sprintf("field %s not found in model %s", name, m.tableName))
	}
	return ModelField{
		Model: m,
		Field: field,
	}
}

func (m *Model) Select(fields ...interface{}) *SelectDataset {
	if len(fields) == 0 {
		return m.selectAll()
	}
	return m.selectPart(fields...)
}

func (m *Model) Delete() *DeleteDataset {
	dataset := goqu.Delete(m.tableName)
	return &DeleteDataset{
		model:   m,
		dataset: dataset,
		tx:      nil,
	}
}

func (m *Model) Update(values interface{}) *UpdateDataset {
	dataset := goqu.Update(m.tableName).Set(values)
	return &UpdateDataset{
		model:   m,
		dataset: dataset,
		tx:      nil,
	}
}

func (m *Model) Insert(rows ...interface{}) *InsertDataset {
	dataset := goqu.Insert(m.tableName).Rows(rows)
	return &InsertDataset{
		model:   m,
		dataset: dataset,
		tx:      nil,
	}
}

// -------------------------------------------------------------------------------------------------------- //

type modelI interface {
	Init(db *DbClient, model interface{}) error
}

type fkModel struct {
	As        string
	FromField string
	ToField   string
	Model     Model
}

type joiner struct {
	Table exp.AliasedExpression
	On    exp.JoinCondition
}

func (m *Model) initSelectors(prefix, tableName string) {
	if prefix == "" && tableName == "" {
		for _, field := range m.dbFields {
			m.preparedSelectors[field] = goqu.I(fmt.Sprintf("%s.%s", m.tableName, field)).As(field)
		}
		for _, fkModel := range m.fkModels {
			tableAsName := fmt.Sprintf("%s%s%s", m.tableName, SEPARATOR, fkModel.As)
			fkModel.Model.initSelectors(fkModel.As, tableAsName)
		}
	} else {
		for _, field := range m.dbFields {
			m.preparedSelectors[field] = goqu.I(fmt.Sprintf("%s.%s", tableName, field)).As(goqu.S(fmt.Sprintf("%s.%s", prefix, field)))
		}
		for _, fkModel := range m.fkModels {
			tableAsName := fmt.Sprintf("%s%s%s", tableName, SEPARATOR, fkModel.As)
			fkModel.Model.initSelectors(fmt.Sprintf("%s.%s", prefix, fkModel.As), tableAsName)
		}
	}
}

func (m *Model) initJoins(modelAs string) {
	for _, fkModel := range m.fkModels {
		tableAsName := fmt.Sprintf("%s%s%s", modelAs, SEPARATOR, fkModel.As)
		if modelAs == "" {
			tableAsName = fmt.Sprintf("%s%s%s", m.tableName, SEPARATOR, fkModel.As)
			modelAs = m.tableName
		}
		var joiner joiner
		joiner.Table = goqu.T(fkModel.Model.tableName).As(tableAsName)
		joiner.On = goqu.On(
			goqu.Ex{
				fmt.Sprintf("%s.%s", modelAs, fkModel.FromField): goqu.I(fmt.Sprintf("%s.%s", tableAsName, fkModel.ToField)),
			},
		)
		m.preparedJoins[fkModel.As] = joiner
		fkModel.Model.initJoins(tableAsName)
	}
}

func (m *Model) joinAll(dataset *goqu.SelectDataset, joinTables map[string]bool) (*goqu.SelectDataset, map[string]bool) {
	for _, joiner := range m.preparedJoins {
		dataset = dataset.LeftJoin(joiner.Table, joiner.On)
	}
	for _, fkModel := range m.fkModels {
		joinTables[fkModel.As] = true
		dataset, joinTables = fkModel.Model.joinAll(dataset, joinTables)
	}
	return dataset, joinTables
}

func (m *Model) getAllFields() []interface{} {
	var result []interface{}
	for _, field := range m.preparedSelectors {
		result = append(result, field)
	}
	for _, fkModel := range m.fkModels {
		result = append(result, fkModel.Model.getAllFields()...)
	}
	return result
}

func (m *Model) selectAll() *SelectDataset {
	dataset := goqu.From(m.tableName)
	joinedTables := make(map[string]bool)
	dataset, joinedTables = m.joinAll(dataset, joinedTables)
	selectFields := m.getAllFields()
	dataset = dataset.Select(selectFields...)
	return &SelectDataset{
		model:        m,
		dataset:      dataset,
		joinedTables: joinedTables,
	}
}

func (m *Model) selectPart(fields ...interface{}) *SelectDataset {
	dataset := goqu.From(m.tableName)
	joinedTables := make(map[string]bool)

	var selectFields []interface{}
	for _, field := range fields {
		switch field.(type) {
		case string:
			fmt.Println()
		case SelfFields:
			selfExp, _ := field.(SelfFields)
			dataset, joinedTables = m.joinModel(selfExp.Model, dataset, joinedTables)
			for _, selectors := range selfExp.Model.preparedSelectors {
				selectFields = append(selectFields, selectors.(interface{}))
			}
		case ModelField:
			fieldExp, _ := field.(ModelField)
			dataset, joinedTables = m.joinModel(fieldExp.Model, dataset, joinedTables)
			selectFields = append(selectFields, fieldExp.Field.(interface{}))
		case CountExpression:
			countExp, _ := field.(CountExpression)
			dataset, joinedTables = m.joinModel(countExp.Field.Model, dataset, joinedTables)
			selectFields = append(selectFields, goqu.COUNT(countExp.Field.Field.Aliased()))
		case CalculationExpression:
			calExp, _ := field.(CalculationExpression)
			field1, field2 := calExp.Fields()
			dataset, joinedTables = m.joinModel(field1.Model, dataset, joinedTables)
			dataset, joinedTables = m.joinModel(field2.Model, dataset, joinedTables)
			selectFields = append(selectFields, calExp.Sql())
		case LiteralExpression:
			lExp, _ := field.(LiteralExpression)
			for _, field := range lExp.Fields {
				dataset, joinedTables = m.joinModel(field.Model, dataset, joinedTables)
			}
			selectFields = append(selectFields, lExp.Sql())
		default:
			panic(fmt.Sprintf("unsupported type %T", field))
		}
	}

	dataset = dataset.Select(selectFields...)
	return &SelectDataset{
		model:        m,
		dataset:      dataset,
		joinedTables: joinedTables,
	}
}

func (m *Model) joinModel(subModel *Model, dataset *goqu.SelectDataset, joinedTables map[string]bool) (*goqu.SelectDataset, map[string]bool) {
	if subModel.asName != "" && !joinedTables[subModel.asName] {
		if subModel.parent == nil {
			panic(fmt.Sprintf("error in select. Model %s has no parent model", subModel.tableName))
		}
		joiner, ok := subModel.parent.preparedJoins[subModel.asName]
		if !ok {
			panic(fmt.Sprintf("error in select. Model %s has no joiner", subModel.tableName))
		}
		dataset = dataset.LeftJoin(joiner.Table, joiner.On)
	}
	return dataset, joinedTables
}
