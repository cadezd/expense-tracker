package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPostgresPool(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	pgxConf, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database url %q: %v", databaseURL, err)
	}

	pgxConf.MaxConns = 10
	pgxConf.MinConns = 1

	pool, err := pgxpool.NewWithConfig(ctx, pgxConf)
	if err != nil {
		return nil, fmt.Errorf("failed to crete connection pool: %v", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to establish a connection to %q: %v", databaseURL, err)
	}

	return pool, nil
}
