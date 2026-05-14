package storage

import (
	"context"
	"fmt"
	"io/fs"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// MigrateFromFS applies *.sql files from fs in lexical order. Each file
// runs in its own transaction. Already-applied filenames are tracked in
// schema_migrations (created lazily). Safe to call repeatedly.
func MigrateFromFS(ctx context.Context, pool *pgxpool.Pool, sqlFS fs.FS) (applied []string, err error) {
	if _, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			name        text PRIMARY KEY,
			applied_at  timestamptz NOT NULL DEFAULT now()
		)`); err != nil {
		return nil, fmt.Errorf("migrate: ensure schema_migrations: %w", err)
	}

	files, err := listSQL(sqlFS)
	if err != nil {
		return nil, err
	}

	for _, name := range files {
		var exists bool
		err := pool.QueryRow(ctx,
			`SELECT exists(SELECT 1 FROM schema_migrations WHERE name = $1)`,
			name,
		).Scan(&exists)
		if err != nil {
			return applied, fmt.Errorf("migrate: lookup %s: %w", name, err)
		}
		if exists {
			continue
		}

		body, err := fs.ReadFile(sqlFS, name)
		if err != nil {
			return applied, fmt.Errorf("migrate: read %s: %w", name, err)
		}

		err = withTx(ctx, pool, func(tx pgx.Tx) error {
			if _, err := tx.Exec(ctx, string(body)); err != nil {
				return fmt.Errorf("exec %s: %w", name, err)
			}
			if _, err := tx.Exec(ctx,
				`INSERT INTO schema_migrations (name) VALUES ($1)`, name,
			); err != nil {
				return fmt.Errorf("record %s: %w", name, err)
			}
			return nil
		})
		if err != nil {
			return applied, fmt.Errorf("migrate: %w", err)
		}
		applied = append(applied, name)
	}
	return applied, nil
}

func listSQL(sqlFS fs.FS) ([]string, error) {
	var out []string
	err := fs.WalkDir(sqlFS, ".", func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".sql") {
			out = append(out, path)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("migrate: walk: %w", err)
	}
	sort.Strings(out)
	return out, nil
}

func withTx(ctx context.Context, pool *pgxpool.Pool, fn func(pgx.Tx) error) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
