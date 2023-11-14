# kvorm

`kvorm` is an ORM focused on working with Postresql

## Installation
```sh
go get github.com/KVOrange/kvorm@0.0.2
```

##  Quick Start
```go
package main

import (
	"github.com/KVOrange/kvorm"
	"github.com/jackc/pgx/v5/pgtype"
)

type JobTitle struct {
    kvorm.Model `table:"job_title"`
    
    Id      pgtype.Int8 `json:"id" type:"pk"`
    Name    pgtype.Text `json:"name"`
    NameAdd pgtype.Text `json:"name_add"`
}

type Person struct {
    kvorm.Model `table:"person"`
    
    Id   pgtype.Int8 `json:"id" type:"pk"`
    Code pgtype.Text `json:"code"`
    Fio  pgtype.Text `json:"fio"`
    
    JobTitle JobTitle `db:"job_title" fk:"job_title_id,id"`
}

func main() {
    dbClient := kvorm.DbClient{}
    err = dbClient.Connect(ctx, kvorm.DbConfig{
        Host:      "localhost",
        Port:      5432,
        User:      "user",
        Password:  "password",
        Name:      "db_name",
        PollCount: 10,
    })

    var personTable Person
    err = personTable.Init(&dbClient, &personTable)
    if err != nil {
        panic(err)
    }

    var person Person
    err = personTable.Select().Where(kvorm.And{
        Field: personTable.Field("id"),
        Op:    kvorm.Eq,
        Value: 1,
    }).ScanOne(&person)
    if err != nil {
    panic(err)
    }
    
    fmt.Println(person)
}
```

