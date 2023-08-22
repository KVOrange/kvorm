package database

import (
	"github.com/jackc/pgx/v5/pgtype"
	"kvorm_lib/libraries/kvorm"
)

type City struct {
	kvorm.Model `table:"dict_city"`

	Id      pgtype.Int8 `json:"id"`
	Code    pgtype.Text `json:"code"`
	Name    pgtype.Text `json:"name"`
	NameAdd pgtype.Text `json:"name_add"`
	Profile CallProfile `json:"profile" fk:"call_profile_id,id" db:"profile"`
}

type CallProfile struct {
	kvorm.Model `table:"dict_call_profile"`

	Id      pgtype.Int8 `json:"id"`
	Code    pgtype.Text `json:"code"`
	Name    pgtype.Text `json:"name"`
	NameAdd pgtype.Text `json:"name_add"`
}

type JobTitle struct {
	kvorm.Model `table:"job_title"`

	Id      pgtype.Int8 `json:"id"`
	Name    pgtype.Text `json:"name"`
	NameAdd pgtype.Text `json:"name_add"`
	City    City        `json:"city" fk:"city_id,id" db:"city"`
}

type Person struct {
	kvorm.Model `table:"person"`

	Id      pgtype.Int8 `json:"id"`
	Code    pgtype.Text `json:"code"`
	Fio     pgtype.Text `json:"fio"`
	IsFired pgtype.Bool `json:"is_fired"`

	JobTitle JobTitle `json:"job_title" fk:"job_title_id,id" db:"job_title"`
	Jobber   JobTitle `json:"jobber" fk:"jobber_id,id" db:"jobber"`
}
