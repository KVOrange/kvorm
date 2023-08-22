package models

import (
	"kvorm_lib/libraries/kvorm/database"
	"reflect"
)

func InitTable(model interface{}, client *database.DbClient) {
	if tabler, ok := model.(interface {
		ExtractTableName(interface{})
		InitFields(interface{})
		InitConnection(dbClient *database.DbClient)
	}); ok {
		tabler.ExtractTableName(model)
		tabler.InitFields(model)
		tabler.InitConnection(client)
	}
}

func InitAllModels(modelsStruct interface{}, client *database.DbClient) {
	rValue := reflect.ValueOf(modelsStruct).Elem()

	for i := 0; i < rValue.NumField(); i++ {
		modelField := rValue.Field(i)

		// Проверяем, является ли поле структурой, содержащей Model
		if modelField.Kind() == reflect.Struct {
			model := modelField.Addr().Interface()
			InitTable(model, client)
		}
	}
}
