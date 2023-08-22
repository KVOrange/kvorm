package main

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgtype"
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

	type Models struct {
		Person database.Person
	}

	var models Models
	kvorm.InitAllModels(&models, &dbClient)

	query := models.Person.Select("id", "jobber__city__id").String()
	fmt.Println(query)

	var person database.Person
	person.Id = pgtype.Int8{Int64: 1, Valid: true}
	person.Fio = pgtype.Text{String: "ВОРОНКИН КИРИЛЛ", Valid: true}
	person.Code = pgtype.Text{String: "999", Valid: true}
	person.IsFired = pgtype.Bool{Bool: false, Valid: true}
	models.Person.Save(&person)
}
