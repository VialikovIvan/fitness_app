package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func InitDB(connString string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, err
	}

	const maxConns = 25
	const minConns = 25
	const maxConnLifetime = 5 * time.Minute

	config.MaxConns = maxConns
	config.MinConns = minConns
	config.MaxConnLifetime = maxConnLifetime

	db, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(context.Background()); err != nil {
		return nil, err
	}

	return db, nil
}

func CloseDB(db *pgxpool.Pool) {
	if db != nil {
		db.Close()
	}
}
