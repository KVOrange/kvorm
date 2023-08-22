package main

import (
	"context"
	"fmt"
	"kvorm_lib/config"
	database "kvorm_lib/database/models"
	"kvorm_lib/libraries/kvorm"
)

func main() {
	ctx := context.Background()

	cfg, err := config.InitConfig()
	if err != nil {
		panic(fmt.Sprintf(`Config init with error: %s`, err))
	}

	dbClient := kvorm.DbClient{}
	err = dbClient.Connect(ctx, kvorm.DbConfig{
		Host:      cfg.Database.Host,
		Port:      cfg.Database.Port,
		User:      cfg.Database.User,
		Password:  cfg.Database.Password,
		Name:      cfg.Database.Name,
		PollCount: cfg.Database.PollCount,
	})
	if err != nil {
		panic(fmt.Sprintf(`Database init with error: %s`, err))
	}

	var person database.Person
	kvorm.InitTable(&person, &dbClient)

	person.Query().Select()

}
