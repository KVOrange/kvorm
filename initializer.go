package kvorm

import (
	"reflect"
)

func InitTable(model interface{}, client *DbClient) {
	if tabler, ok := model.(interface {
		InitFields(interface{})
		InitConnection(dbClient *DbClient)
	}); ok {
		tabler.InitFields(model)
		tabler.InitConnection(client)
	}
}

func InitAllModels(modelsStruct interface{}, client *DbClient) {
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
