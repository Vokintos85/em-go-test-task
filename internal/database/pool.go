package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
)

// NewPool creates a new pgx connection pool with retries.
func NewPool(ctx context.Context, connString string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("parse pool config: %w", err)
	}

	var pool *pgxpool.Pool
	for attempt := 0; attempt < 5; attempt++ {
		pool, err = pgxpool.ConnectConfig(ctx, cfg)
		if err == nil {
			return pool, nil
		}

		select {
		case <-time.After(time.Duration(attempt+1) * time.Second):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	if err != nil {
		return nil, fmt.Errorf("connect to database: %w", err)
	}
	return pool, nil
}
