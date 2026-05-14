// Package storage owns the pgx connection pool, migration runner and
// repository contracts for raw_messages and metrics_counters.
//
// Integration tests live in *_test.go files behind the `integration` build
// tag and require INTEGRATION_DSN to be set to a reachable Postgres
// connection string. Multiple packages run migrations against the same
// schema, so run with -p 1 to avoid concurrent CREATE EXTENSION races:
//
//	INTEGRATION_DSN=postgres://... \
//	  go test -tags=integration -p 1 \
//	    ./internal/storage ./internal/feed ./internal/recovery
package storage
