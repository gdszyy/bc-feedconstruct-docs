// Package storage owns the pgx connection pool, migration runner and
// repository contracts for raw_messages and metrics_counters.
//
// Integration tests live in *_integration_test.go behind the
// `integration` build tag and require INTEGRATION_DSN to be set to a
// reachable Postgres connection string. Run them with:
//
//	INTEGRATION_DSN=postgres://... go test -tags=integration ./internal/storage
package storage
