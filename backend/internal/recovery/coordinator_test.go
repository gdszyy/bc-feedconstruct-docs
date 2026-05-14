package recovery_test

import (
	"context"
	"encoding/json"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/gdszyy/bc-feedconstruct-docs/backend/internal/recovery"
	"github.com/gdszyy/bc-feedconstruct-docs/backend/internal/webapi"
)

// 验收 10 — 恢复（启动级）
//
// Given the BFF starts cleanly with an empty raw_messages table
// When the recovery coordinator runs the startup scope
// Then DataSnapshot is invoked for both isLive=true and isLive=false
//      AND a recovery_jobs row is finalized with status=success
func TestGiven_FreshStart_When_StartupRecoveryRuns_Then_FullSnapshotAndJobSuccess(t *testing.T) {
	api := &fakeAPI{}
	jobs := newFakeJobs()
	now := time.Date(2026, 5, 14, 12, 0, 0, 0, time.UTC)
	c := newCoord(api, jobs, now, nilLastMsg)

	require.NoError(t, c.StartupRecovery(context.Background()))

	require.Equal(t, []snapshotCall{
		{IsLive: true, GetChangesFrom: 0},
		{IsLive: false, GetChangesFrom: 0},
	}, api.snapshots, "fresh start must request full snapshot for both products")

	require.Len(t, jobs.finalized, 2)
	for _, f := range jobs.finalized {
		require.Equal(t, "success", f.status)
		require.Equal(t, "startup", jobs.rows[f.id].Scope)
	}
}

// Given an outage of less than 1 hour (last_message_at within 60 minutes)
// When recovery runs
// Then DataSnapshot is invoked WITH getChangesFrom = last_message_at - safetyWindow
//
// We measure outage in whole minutes and add the safety window so any
// in-flight message at the cutoff is reprocessed.
func TestGiven_ShortOutage_When_RecoveryRuns_Then_GetChangesFromUsed(t *testing.T) {
	api := &fakeAPI{}
	jobs := newFakeJobs()
	now := time.Date(2026, 5, 14, 12, 0, 0, 0, time.UTC)
	lastMsg := now.Add(-15 * time.Minute)

	c := newCoord(api, jobs, now, func(ctx context.Context) (time.Time, bool, error) {
		return lastMsg, true, nil
	})
	c.SafetyWindow = 2 * time.Minute

	require.NoError(t, c.StartupRecovery(context.Background()))

	require.Equal(t, []snapshotCall{
		{IsLive: true, GetChangesFrom: 17},
		{IsLive: false, GetChangesFrom: 17},
	}, api.snapshots, "short outage must request only the changed window plus safety")
}

// Given a single match with stale data (no events for >5 minutes while live)
// When event-level recovery is requested
// Then GetMatchByID is invoked and that match's markets/outcomes are refreshed
func TestGiven_StaleLiveMatch_When_EventRecovery_Then_GetMatchByIDInvokedAndStateRefreshed(t *testing.T) {
	api := &fakeAPI{}
	jobs := newFakeJobs()
	now := time.Date(2026, 5, 14, 12, 0, 0, 0, time.UTC)
	c := newCoord(api, jobs, now, nilLastMsg)

	require.NoError(t, c.EventRecovery(context.Background(), 4242))

	require.Equal(t, []int64{4242}, api.matchIDs)
	require.Len(t, jobs.finalized, 1)
	require.Equal(t, "success", jobs.finalized[0].status)
	require.Equal(t, "event", jobs.rows[jobs.finalized[0].id].Scope)
	require.Equal(t, int64(4242), jobs.rows[jobs.finalized[0].id].MatchID)
}

// Given the WebAPI returns HTTP 429 Too Many Requests
// When recovery encounters it
// Then the job is marked rate_limited and retried with exponential backoff
//      capped at the documented max
func TestGiven_429FromWebAPI_When_RecoveryRetries_Then_ExponentialBackoffApplied(t *testing.T) {
	api := &fakeAPI{matchErrs: map[int64][]error{
		7: {
			&webapi.RateLimitError{Path: "/MatchById"},
			&webapi.RateLimitError{Path: "/MatchById"},
			&webapi.RateLimitError{Path: "/MatchById"},
			nil,
		},
	}}
	jobs := newFakeJobs()
	now := time.Date(2026, 5, 14, 12, 0, 0, 0, time.UTC)

	var sleeps []time.Duration
	var mu sync.Mutex
	c := newCoord(api, jobs, now, nilLastMsg)
	c.BackoffBase = 100 * time.Millisecond
	c.BackoffMax = 250 * time.Millisecond
	c.MaxAttempts = 5
	c.Sleep = func(d time.Duration) {
		mu.Lock()
		defer mu.Unlock()
		sleeps = append(sleeps, d)
	}

	require.NoError(t, c.EventRecovery(context.Background(), 7))

	require.Equal(t, []time.Duration{
		100 * time.Millisecond, // attempt 1 fail → wait 100ms
		200 * time.Millisecond, // attempt 2 fail → wait 200ms
		250 * time.Millisecond, // attempt 3 fail → would be 400ms, capped at 250ms
	}, sleeps)

	require.Equal(t, []string{"rate_limited", "rate_limited", "rate_limited", "success"}, jobs.statusHistory)
}

// Given exhausted retry attempts (still rate-limited after MaxAttempts)
// When recovery gives up
// Then the job is finalized with status=rate_limited and a typed error is returned
func TestGiven_RetriesExhausted_When_StillRateLimited_Then_FailsWithRateLimited(t *testing.T) {
	api := &fakeAPI{matchErrs: map[int64][]error{
		9: {
			&webapi.RateLimitError{Path: "/MatchById"},
			&webapi.RateLimitError{Path: "/MatchById"},
			&webapi.RateLimitError{Path: "/MatchById"},
		},
	}}
	jobs := newFakeJobs()
	now := time.Date(2026, 5, 14, 12, 0, 0, 0, time.UTC)
	c := newCoord(api, jobs, now, nilLastMsg)
	c.MaxAttempts = 3
	c.BackoffBase = time.Millisecond
	c.BackoffMax = 5 * time.Millisecond
	c.Sleep = func(time.Duration) {}

	err := c.EventRecovery(context.Background(), 9)
	require.Error(t, err)
	require.True(t, webapi.IsRateLimited(err))

	require.Equal(t, "rate_limited", jobs.finalized[len(jobs.finalized)-1].status)
}

// ---------------------------------------------------------------------------
// Test doubles
// ---------------------------------------------------------------------------

type snapshotCall struct {
	IsLive         bool
	GetChangesFrom int
}

type fakeAPI struct {
	mu          sync.Mutex
	snapshots   []snapshotCall
	matchIDs    []int64
	matchErrs   map[int64][]error
	snapshotErr error
}

func (f *fakeAPI) DataSnapshot(_ context.Context, isLive bool, getChangesFrom int) ([]json.RawMessage, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.snapshots = append(f.snapshots, snapshotCall{IsLive: isLive, GetChangesFrom: getChangesFrom})
	return nil, f.snapshotErr
}

func (f *fakeAPI) MatchByID(_ context.Context, id int64) (json.RawMessage, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.matchIDs = append(f.matchIDs, id)
	if errs, ok := f.matchErrs[id]; ok && len(errs) > 0 {
		err := errs[0]
		f.matchErrs[id] = errs[1:]
		if err != nil {
			return nil, err
		}
	}
	return json.RawMessage(`{"Id":` + strconv.FormatInt(id, 10) + `}`), nil
}

type fakeJob struct {
	ID      int64
	Scope   string
	Product string
	MatchID int64
}

type finalEntry struct {
	id     int64
	status string
}

type fakeJobs struct {
	mu            sync.Mutex
	seq           int64
	rows          map[int64]fakeJob
	finalized     []finalEntry
	statusHistory []string
}

func newFakeJobs() *fakeJobs {
	return &fakeJobs{rows: map[int64]fakeJob{}}
}

func (j *fakeJobs) Record(_ context.Context, scope, product string, matchID int64) (int64, error) {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.seq++
	j.rows[j.seq] = fakeJob{ID: j.seq, Scope: scope, Product: product, MatchID: matchID}
	return j.seq, nil
}

func (j *fakeJobs) Finalize(_ context.Context, id int64, status string, _ map[string]any) error {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.finalized = append(j.finalized, finalEntry{id: id, status: status})
	j.statusHistory = append(j.statusHistory, status)
	return nil
}

func newCoord(api recovery.WebAPI, jobs recovery.Jobs, now time.Time, last recovery.LastMessageAtFn) *recovery.Coordinator {
	return &recovery.Coordinator{
		API:           api,
		Jobs:          jobs,
		LastMessageAt: last,
		Now:           func() time.Time { return now },
		SafetyWindow:  time.Minute,
		BackoffBase:   time.Millisecond,
		BackoffMax:    10 * time.Millisecond,
		MaxAttempts:   3,
		Sleep:         func(time.Duration) {},
	}
}

func nilLastMsg(_ context.Context) (time.Time, bool, error) {
	return time.Time{}, false, nil
}
