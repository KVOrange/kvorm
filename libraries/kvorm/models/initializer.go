package models

import "kvorm_lib/libraries/kvorm/database"

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

func InitFkTable(model interface{}, client *database.DbClient, prefix string) {
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
