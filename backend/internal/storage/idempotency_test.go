//go:build integration

package storage_test

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/gdszyy/bc-feedconstruct-docs/backend/internal/storage"
	"github.com/gdszyy/bc-feedconstruct-docs/backend/migrations"
)

var (
	sharedPool *storage.Pool
	sharedOnce sync.Once
	sharedErr  error
)

// getSharedPool migrates the schema exactly once per test process. Each
// test truncates the tables it touches via cleanTables.
func getSharedPool(t *testing.T) *storage.Pool {
	t.Helper()
	dsn := os.Getenv("INTEGRATION_DSN")
	if dsn == "" {
		t.Skip("INTEGRATION_DSN not set; skipping storage integration tests")
	}
	sharedOnce.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		p, err := storage.NewPool(ctx, dsn)
		if err != nil {
			sharedErr = err
			return
		}
		if _, err := storage.MigrateFromFS(ctx, p, migrations.FS()); err != nil {
			sharedErr = err
			return
		}
		sharedPool = p
	})
	require.NoError(t, sharedErr)
	return sharedPool
}

func cleanTables(t *testing.T, pool *storage.Pool, tables ...string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	for _, tbl := range tables {
		_, err := pool.Exec(ctx, "TRUNCATE TABLE "+tbl+" RESTART IDENTITY CASCADE")
		require.NoError(t, err)
	}
}

// 验收 11 — 幂等
//
// Given a raw_messages row with (source, message_type, event_id, ts_provider) already inserted
// When the same delivery is consumed again
// Then the unique constraint blocks the second insert and the count stays at 1
func TestGiven_DuplicateDelivery_When_InsertRawMessage_Then_UniqueConstraintHolds(t *testing.T) {
	pool := getSharedPool(t)
	cleanTables(t, pool, "raw_messages", "metrics_counters")
	repo := storage.NewRawMessageRepo(pool)
	ctx := context.Background()

	ts := time.Date(2026, 5, 14, 12, 0, 0, 0, time.UTC)
	msg := storage.RawMessage{
		Source:      "rmq.live",
		Queue:       "P123_live",
		RoutingKey:  "odds_change.42",
		MessageType: "odds_change",
		EventID:     "match-42",
		TSProvider:  &ts,
		Payload:     []byte(`{"odds":1.85}`),
	}

	r1, err := repo.Insert(ctx, msg)
	require.NoError(t, err)
	require.True(t, r1.Inserted)

	r2, err := repo.Insert(ctx, msg)
	require.NoError(t, err)
	require.False(t, r2.Inserted, "second insert must collapse onto existing row")
	require.Equal(t, r1.ID, r2.ID, "idempotent insert must surface the original ID")

	n, err := repo.CountSince(ctx, time.Time{})
	require.NoError(t, err)
	require.Equal(t, int64(1), n)
}

// Given two messages identical except for ts_provider
// When both are inserted
// Then both rows are written (different idempotency keys)
func TestGiven_DistinctTSProvider_When_Insert_Then_BothRowsWritten(t *testing.T) {
	pool := getSharedPool(t)
	cleanTables(t, pool, "raw_messages", "metrics_counters")
	repo := storage.NewRawMessageRepo(pool)
	ctx := context.Background()

	ts1 := time.Date(2026, 5, 14, 12, 0, 0, 0, time.UTC)
	ts2 := ts1.Add(time.Second)
	base := storage.RawMessage{
		Source:      "rmq.live",
		MessageType: "odds_change",
		EventID:     "match-42",
		Payload:     []byte(`{"odds":1.85}`),
	}
	base.TSProvider = &ts1
	_, err := repo.Insert(ctx, base)
	require.NoError(t, err)

	base.TSProvider = &ts2
	_, err = repo.Insert(ctx, base)
	require.NoError(t, err)

	n, err := repo.CountSince(ctx, time.Time{})
	require.NoError(t, err)
	require.Equal(t, int64(2), n)
}

// 验收 16 — 数据治理 / 保留窗口
//
// Given raw_messages older than the retention window (default 7d)
// When the retention job runs
// Then those rows are deleted and metrics_counters.retention_deleted increments
func TestGiven_RawMessagesPastRetention_When_RetentionJobRuns_Then_RowsDeletedAndCounterIncrements(t *testing.T) {
	pool := getSharedPool(t)
	cleanTables(t, pool, "raw_messages", "metrics_counters")
	repo := storage.NewRawMessageRepo(pool)
	ctx := context.Background()

	old1 := time.Now().Add(-10 * 24 * time.Hour)
	old2 := time.Now().Add(-8 * 24 * time.Hour)
	fresh := time.Now().Add(-1 * time.Hour)

	for i, ts := range []time.Time{old1, old2, fresh} {
		_, err := pool.Exec(ctx, `
			INSERT INTO raw_messages
				(source, message_type, event_id, ts_provider, payload, received_at)
			VALUES ($1,$2,$3,$4,$5,$6)`,
			"rmq.live", "odds_change",
			"match-"+itoa(i), ts, []byte(`{}`), ts,
		)
		require.NoError(t, err)
	}

	cutoff := time.Now().Add(-7 * 24 * time.Hour)
	deleted, err := repo.DeleteOlderThan(ctx, cutoff)
	require.NoError(t, err)
	require.Equal(t, int64(2), deleted)

	n, err := repo.CountSince(ctx, time.Time{})
	require.NoError(t, err)
	require.Equal(t, int64(1), n, "only the fresh row should remain")

	v, err := storage.Counter(ctx, pool, "retention_deleted")
	require.NoError(t, err)
	require.Equal(t, int64(2), v)
}

func itoa(i int) string {
	const digits = "0123456789"
	if i == 0 {
		return "0"
	}
	var b []byte
	for i > 0 {
		b = append([]byte{digits[i%10]}, b...)
		i /= 10
	}
	return string(b)
}
