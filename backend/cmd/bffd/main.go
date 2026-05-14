// Package main is the entrypoint for the BFF service.
//
// At this stage (BDD wave 1) the binary loads config, opens a Postgres pool,
// runs migrations and serves /healthz + /readyz. Feed consumers, handlers and
// the WebSocket hub are wired in subsequent waves.
package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/gdszyy/bc-feedconstruct-docs/backend/internal/config"
	"github.com/gdszyy/bc-feedconstruct-docs/backend/internal/storage"
	"github.com/gdszyy/bc-feedconstruct-docs/backend/migrations"
)

func main() {
	if code := run(); code != 0 {
		os.Exit(code)
	}
}

func run() int {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "bffd: config: %v\n", err)
		return 2
	}
	fmt.Fprintf(os.Stdout, "bffd: starting %s\n", cfg.Redacted())

	rootCtx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	bootCtx, bootCancel := context.WithTimeout(rootCtx, 30*time.Second)
	defer bootCancel()

	pool, err := storage.NewPool(bootCtx, cfg.DatabaseURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "bffd: postgres: %v\n", err)
		return 2
	}
	defer pool.Close()

	applied, err := storage.MigrateFromFS(bootCtx, pool, migrations.FS())
	if err != nil {
		fmt.Fprintf(os.Stderr, "bffd: migrate: %v\n", err)
		return 2
	}
	for _, name := range applied {
		fmt.Fprintf(os.Stdout, "bffd: migrated %s\n", name)
	}

	var ready atomic.Bool
	// At this wave the only readiness gate is Postgres + migrations.
	ready.Store(true)

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintln(w, "ok")
	})
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		ctx, c := context.WithTimeout(r.Context(), 2*time.Second)
		defer c()
		if !ready.Load() {
			http.Error(w, "starting", http.StatusServiceUnavailable)
			return
		}
		if err := pool.Ping(ctx); err != nil {
			http.Error(w, "db down: "+err.Error(), http.StatusServiceUnavailable)
			return
		}
		_, _ = fmt.Fprintln(w, "ready")
	})

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		fmt.Fprintf(os.Stdout, "bffd: listening on :%s\n", cfg.Port)
		errCh <- srv.ListenAndServe()
	}()

	select {
	case <-rootCtx.Done():
		fmt.Fprintln(os.Stdout, "bffd: shutdown signal received")
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			fmt.Fprintf(os.Stderr, "bffd: serve: %v\n", err)
			return 1
		}
	}

	shutCtx, shutCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutCancel()
	_ = srv.Shutdown(shutCtx)
	return 0
}
