package database

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DbClient struct {
	Ctx  context.Context
	Pool *pgxpool.Pool
}

func (cli *DbClient) Connect(ctx context.Context, cfg DbConfig) error {
	connectionStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Password,
		cfg.Name,
	)
	pgxConfig, err := pgxpool.ParseConfig(connectionStr)
	if err != nil {
		return err
	}
	pgxConfig.MaxConns = cfg.PollCount

	db, err := pgxpool.NewWithConfig(context.Background(), pgxConfig)
	if err != nil {
		return err
	}

	pingErr := db.Ping(context.Background())
	if pingErr != nil {
		return pingErr
	}
	if db == nil {
		return fmt.Errorf("Ошибка соединения с базой данных")
	}

	cli.Pool = db
	cli.Ctx = ctx

	return nil
}
