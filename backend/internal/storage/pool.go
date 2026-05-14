// Package storage owns the pgx connection pool, migration runner and
// repository contracts. Repositories are thin: they hold a *pgxpool.Pool
// and translate domain operations into SQL.
package storage

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Pool is the application-wide pgx pool.
type Pool = pgxpool.Pool

// NewPool dials Postgres and pings until ctx is done or the first
// successful round-trip completes.
func NewPool(ctx context.Context, dsn string) (*Pool, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("storage: parse dsn: %w", err)
	}
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("storage: connect: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("storage: ping: %w", err)
	}
	return pool, nil
}
