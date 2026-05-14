// Package main is the entrypoint for the BFF service.
//
// Wave 2: in addition to /healthz + /readyz, the binary now starts the
// feed ingest pipeline. FEED_MODE=replay reads JSON fixtures from
// REPLAY_DIR (default: backend/internal/feed/testdata/replay) while
// FEED_MODE=live consumes both FeedConstruct partner queues.
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
	"github.com/gdszyy/bc-feedconstruct-docs/backend/internal/feed"
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

	// Optional internal RabbitMQ publisher. If the broker is unreachable
	// at boot we fall back to NopPublisher rather than blocking startup;
	// the bind error is logged so operators see it.
	pub := startPublisher(cfg)
	defer func() { _ = pub.Close() }()

	repo := storage.NewRawMessageRepo(pool)
	disp := feed.NewDispatcher(nil)
	proc := feed.NewProcessor(repo, pub, disp)

	feedErrCh := make(chan error, 1)
	go func() { feedErrCh <- runFeed(rootCtx, cfg, proc) }()

	var ready atomic.Bool
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

	httpErrCh := make(chan error, 1)
	go func() {
		fmt.Fprintf(os.Stdout, "bffd: listening on :%s\n", cfg.Port)
		httpErrCh <- srv.ListenAndServe()
	}()

	select {
	case <-rootCtx.Done():
		fmt.Fprintln(os.Stdout, "bffd: shutdown signal received")
	case err := <-httpErrCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			fmt.Fprintf(os.Stderr, "bffd: serve: %v\n", err)
			cancel()
			<-feedErrCh
			return 1
		}
	case err := <-feedErrCh:
		// Replayer completes naturally; live consumer only returns on error
		// or ctx cancellation.
		if err != nil && !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
			fmt.Fprintf(os.Stderr, "bffd: feed: %v\n", err)
		} else {
			fmt.Fprintln(os.Stdout, "bffd: feed loop finished")
		}
	}

	shutCtx, shutCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutCancel()
	_ = srv.Shutdown(shutCtx)
	return 0
}

func startPublisher(cfg *config.Config) feed.Publisher {
	if cfg.RabbitMQURL == "" {
		fmt.Fprintln(os.Stdout, "bffd: RABBITMQ_URL empty; using NopPublisher")
		return feed.NopPublisher{}
	}
	pub, err := feed.NewAMQPPublisher(cfg.RabbitMQURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "bffd: internal rabbitmq unavailable, falling back to NopPublisher: %v\n", err)
		return feed.NopPublisher{}
	}
	fmt.Fprintln(os.Stdout, "bffd: internal RabbitMQ exchange ready (feed.events)")
	return pub
}

func runFeed(ctx context.Context, cfg *config.Config, proc *feed.Processor) error {
	switch cfg.Mode {
	case config.ModeReplay:
		dir := cfg.ReplayDir
		if dir == "" {
			dir = "internal/feed/testdata/replay"
		}
		rep := &feed.Replayer{Dir: dir, Processor: proc, Source: "replay"}
		fmt.Fprintf(os.Stdout, "bffd: replay mode, reading %s\n", dir)
		n, err := rep.Run(ctx)
		fmt.Fprintf(os.Stdout, "bffd: replay finished, %d deliveries processed\n", n)
		return err
	case config.ModeLive:
		lc := &feed.LiveConsumer{
			Cfg: feed.LiveConsumerConfig{
				Host:      cfg.FCRMQHost,
				User:      cfg.FCRMQUser,
				Pass:      cfg.FCRMQPass,
				PartnerID: cfg.FCPartnerID,
				UseTLS:    cfg.FCRMQTLS,
			},
			Processor: proc,
		}
		fmt.Fprintf(os.Stdout, "bffd: live mode, connecting to %s (partner=%s, tls=%t)\n",
			cfg.FCRMQHost, cfg.FCPartnerID, cfg.FCRMQTLS)
		return lc.Run(ctx)
	default:
		return fmt.Errorf("bffd: unknown FEED_MODE %q", cfg.Mode)
	}
}
