//go:build integration

package recovery_test

import (
	"context"
	"errors"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/gdszyy/bc-feedconstruct-docs/backend/internal/recovery"
	"github.com/gdszyy/bc-feedconstruct-docs/backend/internal/storage"
	"github.com/gdszyy/bc-feedconstruct-docs/backend/internal/webapi"
	"github.com/gdszyy/bc-feedconstruct-docs/backend/migrations"
)

var (
	recPool *storage.Pool
	recOnce sync.Once
	recErr  error
)

func setupPool(t *testing.T) *storage.Pool {
	t.Helper()
	dsn := os.Getenv("INTEGRATION_DSN")
	if dsn == "" {
		t.Skip("INTEGRATION_DSN not set; skipping recovery integration tests")
	}
	recOnce.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		p, err := storage.NewPool(ctx, dsn)
		if err != nil {
			recErr = err
			return
		}
		if _, err := storage.MigrateFromFS(ctx, p, migrations.FS()); err != nil {
			recErr = err
			return
		}
		recPool = p
	})
	require.NoError(t, recErr)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := recPool.Exec(ctx, "TRUNCATE TABLE recovery_jobs RESTART IDENTITY CASCADE")
	require.NoError(t, err)
	return recPool
}

// stubAPI captures every call and returns canned bodies.
type stubAPI struct {
	mu              sync.Mutex
	snapshotCalls   []snapshotCall
	matchCalls      []int64
	dataSnapshotErr error
	getMatchErr     error
	bodyJSON        []byte
}

type snapshotCall struct {
	isLive      bool
	changesFrom *time.Time
}

func (s *stubAPI) DataSnapshot(_ context.Context, isLive bool, changesFrom *time.Time) (webapi.SnapshotResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.snapshotCalls = append(s.snapshotCalls, snapshotCall{isLive: isLive, changesFrom: changesFrom})
	if s.dataSnapshotErr != nil {
		return webapi.SnapshotResult{}, s.dataSnapshotErr
	}
	body := s.bodyJSON
	if body == nil {
		body = []byte(`{"matches":[]}`)
	}
	return webapi.SnapshotResult{IsLive: isLive, BodyJSON: body}, nil
}

func (s *stubAPI) GetMatchByID(_ context.Context, matchID int64) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.matchCalls = append(s.matchCalls, matchID)
	if s.getMatchErr != nil {
		return nil, s.getMatchErr
	}
	return s.bodyJSON, nil
}

type countingIngester struct {
	count    int32
	bodyByID map[int64][]byte
}

func (c *countingIngester) IngestSnapshot(_ context.Context, _ recovery.Scope, _ string, mid *int64, body []byte) (int, error) {
	atomic.AddInt32(&c.count, 1)
	if mid != nil {
		if c.bodyByID == nil {
			c.bodyByID = map[int64][]byte{}
		}
		c.bodyByID[*mid] = body
	}
	return 1, nil
}

// 验收 10 — 恢复（启动级）
//
// Given a fresh BFF
// When startup recovery is scheduled and RunOnce executes
// Then DataSnapshot is invoked for both isLive=true and isLive=false
//      AND a recovery_jobs row finalises with status=success
func TestGiven_FreshStart_When_StartupRecoveryRuns_Then_FullSnapshotAndJobSuccess(t *testing.T) {
	pool := setupPool(t)
	ctx := context.Background()
	api := &stubAPI{}
	ing := &countingIngester{}
	coord := recovery.New(pool, api, ing, recovery.Options{})

	id, err := coord.ScheduleStartup(ctx)
	require.NoError(t, err)
	require.Greater(t, id, int64(0))

	processed, err := coord.RunOnce(ctx)
	require.NoError(t, err)
	require.True(t, processed)

	require.Len(t, api.snapshotCalls, 2, "must snapshot both products on startup")
	require.True(t, api.snapshotCalls[0].isLive)
	require.False(t, api.snapshotCalls[1].isLive)
	require.Nil(t, api.snapshotCalls[0].changesFrom, "startup must omit getChangesFrom")
	require.Equal(t, int32(2), atomic.LoadInt32(&ing.count))

	var status string
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT status FROM recovery_jobs WHERE id=$1`, id).Scan(&status))
	require.Equal(t, "success", status)
}

// Given an outage of less than 1 hour (changesFrom within 60 minutes)
// When recovery runs
// Then DataSnapshot is invoked WITH getChangesFrom
func TestGiven_ShortOutage_When_RecoveryRuns_Then_GetChangesFromUsed(t *testing.T) {
	pool := setupPool(t)
	ctx := context.Background()
	api := &stubAPI{}
	coord := recovery.New(pool, api, nil, recovery.Options{})

	cf := time.Now().Add(-30 * time.Minute)
	_, err := coord.ScheduleProduct(ctx, "live", &cf)
	require.NoError(t, err)

	ok, err := coord.RunOnce(ctx)
	require.NoError(t, err)
	require.True(t, ok)

	require.Len(t, api.snapshotCalls, 1)
	require.NotNil(t, api.snapshotCalls[0].changesFrom)
	require.WithinDuration(t, cf, *api.snapshotCalls[0].changesFrom, time.Second)
}

// Given a single match with stale data
// When event-level recovery is requested
// Then GetMatchByID is invoked and the body is handed to the ingester
func TestGiven_StaleLiveMatch_When_EventRecovery_Then_GetMatchByIDInvokedAndStateRefreshed(t *testing.T) {
	pool := setupPool(t)
	ctx := context.Background()
	api := &stubAPI{bodyJSON: []byte(`{"matchId":42}`)}
	ing := &countingIngester{}
	coord := recovery.New(pool, api, ing, recovery.Options{})

	_, err := coord.ScheduleEvent(ctx, 42)
	require.NoError(t, err)
	ok, err := coord.RunOnce(ctx)
	require.NoError(t, err)
	require.True(t, ok)

	require.Equal(t, []int64{42}, api.matchCalls)
	require.Equal(t, []byte(`{"matchId":42}`), ing.bodyByID[42])
}

// Given the WebAPI returns HTTP 429 Too Many Requests
// When recovery encounters it
// Then the job is marked rate_limited and retried after the documented backoff
func TestGiven_429FromWebAPI_When_RecoveryRetries_Then_ExponentialBackoffApplied(t *testing.T) {
	pool := setupPool(t)
	ctx := context.Background()
	api := &stubAPI{dataSnapshotErr: &webapi.RateLimitedError{RetryAfter: 4 * time.Second}}
	coord := recovery.New(pool, api, nil, recovery.Options{})

	id, err := coord.ScheduleStartup(ctx)
	require.NoError(t, err)
	ok, err := coord.RunOnce(ctx)
	require.NoError(t, err)
	require.True(t, ok)

	var (
		status      string
		attempt     int16
		nextRetryAt *time.Time
	)
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT status, attempt, next_retry_at FROM recovery_jobs WHERE id=$1`, id,
	).Scan(&status, &attempt, &nextRetryAt))
	require.Equal(t, "rate_limited", status)
	require.Equal(t, int16(1), attempt)
	require.NotNil(t, nextRetryAt)
	require.WithinDuration(t, time.Now().Add(4*time.Second), *nextRetryAt, 3*time.Second)
}

// Given the API consistently fails
// When the same job retries past MaxAttempts
// Then the job is marked failed and won't be picked again
func TestGiven_PersistentFailure_When_BeyondMaxAttempts_Then_StatusFailed(t *testing.T) {
	pool := setupPool(t)
	ctx := context.Background()
	api := &stubAPI{dataSnapshotErr: errors.New("upstream broken")}
	coord := recovery.New(pool, api, nil, recovery.Options{
		MaxAttempts: 2,
		BackoffBase: time.Millisecond,
		BackoffMax:  time.Millisecond,
	})

	id, err := coord.ScheduleStartup(ctx)
	require.NoError(t, err)

	for i := 0; i < 3; i++ {
		_, err := coord.RunOnce(ctx)
		require.NoError(t, err)
		time.Sleep(2 * time.Millisecond) // let next_retry_at expire
	}

	var status string
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT status FROM recovery_jobs WHERE id=$1`, id).Scan(&status))
	require.Equal(t, "failed", status)

	// Now the queue is empty.
	ok, err := coord.RunOnce(ctx)
	require.NoError(t, err)
	require.False(t, ok)
}

// Given an empty queue
// When RunOnce is invoked
// Then it reports (false, nil) without errors
func TestGiven_EmptyQueue_When_RunOnce_Then_NoOp(t *testing.T) {
	pool := setupPool(t)
	coord := recovery.New(pool, &stubAPI{}, nil, recovery.Options{})
	ok, err := coord.RunOnce(context.Background())
	require.NoError(t, err)
	require.False(t, ok)
}
