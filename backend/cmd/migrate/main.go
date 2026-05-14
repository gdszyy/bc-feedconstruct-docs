// Package main is the standalone migration runner. It reads DATABASE_URL,
// applies every embedded SQL file in order, then exits.
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/gdszyy/bc-feedconstruct-docs/backend/internal/config"
	"github.com/gdszyy/bc-feedconstruct-docs/backend/internal/storage"
	"github.com/gdszyy/bc-feedconstruct-docs/backend/migrations"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "migrate: config: %v\n", err)
		os.Exit(2)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	pool, err := storage.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "migrate: connect: %v\n", err)
		os.Exit(2)
	}
	defer pool.Close()

	applied, err := storage.MigrateFromFS(ctx, pool, migrations.FS())
	if err != nil {
		fmt.Fprintf(os.Stderr, "migrate: apply: %v\n", err)
		os.Exit(1)
	}
	if len(applied) == 0 {
		fmt.Fprintln(os.Stdout, "migrate: schema already up-to-date")
		return
	}
	for _, name := range applied {
		fmt.Fprintf(os.Stdout, "migrate: applied %s\n", name)
	}
}
