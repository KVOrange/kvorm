package kvorm

import (
	"kvorm_lib/libraries/kvorm/database"
	"kvorm_lib/libraries/kvorm/models"
)

type (
	Model = models.Model

	DbClient = database.DbClient
	DbConfig = database.DbConfig
)

func InitTable(model interface{}, client *DbClient) {
	models.InitTable(model, client)
}
