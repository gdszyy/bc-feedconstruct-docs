//go:build integration

package subscription_test

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/gdszyy/bc-feedconstruct-docs/backend/internal/catalog"
	"github.com/gdszyy/bc-feedconstruct-docs/backend/internal/feed"
	"github.com/gdszyy/bc-feedconstruct-docs/backend/internal/storage"
	"github.com/gdszyy/bc-feedconstruct-docs/backend/internal/subscription"
	"github.com/gdszyy/bc-feedconstruct-docs/backend/migrations"
)

var (
	subPool *storage.Pool
	subOnce sync.Once
	subErr  error
)

func setup(t *testing.T) *storage.Pool {
	t.Helper()
	dsn := os.Getenv("INTEGRATION_DSN")
	if dsn == "" {
		t.Skip("INTEGRATION_DSN not set; skipping subscription integration tests")
	}
	subOnce.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		p, err := storage.NewPool(ctx, dsn)
		if err != nil {
			subErr = err
			return
		}
		if _, err := storage.MigrateFromFS(ctx, p, migrations.FS()); err != nil {
			subErr = err
			return
		}
		subPool = p
	})
	require.NoError(t, subErr)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := subPool.Exec(ctx, `
		TRUNCATE TABLE subscription_events, subscriptions,
			rollbacks, cancels, settlements, market_status_history,
			outcomes, markets, fixture_changes, matches, competitions,
			regions, sports, raw_messages, metrics_counters
		RESTART IDENTITY CASCADE`)
	require.NoError(t, err)
	return subPool
}

func env(body string) feed.Envelope {
	e, err := feed.DecodeEnvelope([]byte(body))
	if err != nil {
		e = feed.Envelope{Payload: []byte(body)}
	}
	e.Payload = []byte(body)
	return e
}

func seedMatch(t *testing.T, pool *storage.Pool, matchID int64) *catalog.Handler {
	t.Helper()
	ctx := context.Background()
	cat := catalog.New(pool)
	require.NoError(t, cat.Handle(ctx, feed.MsgCatalogSport,
		env(`{"id":1,"name":"Soccer"}`), [16]byte{}))
	require.NoError(t, cat.Handle(ctx, feed.MsgFixture,
		env(`{"matchId":42,"sportId":1,"status":"live"}`), [16]byte{}))
	return cat
}

// 验收 13 — 订阅生命周期（M11）
//
// Given a Book delivery for matchId=42 (product=live)
// When subscription.Manager handles it
// Then a subscriptions row is upserted with status=subscribed
//      AND a subscription_events row records the transition
func TestGiven_BookDelivered_When_Handled_Then_SubscriptionUpsertedAndEventRecorded(t *testing.T) {
	pool := setup(t)
	seedMatch(t, pool, 42)
	ctx := context.Background()
	d := feed.NewDispatcher(nil)
	mgr := subscription.New(pool, subscription.Options{})
	mgr.Register(d)

	require.NoError(t, d.Dispatch(ctx, feed.MsgSubscriptionBook,
		env(`{"matchId":42,"product":"live"}`), [16]byte{}))

	var (
		status         string
		product        string
		subscribedAt   *time.Time
	)
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT status, product, subscribed_at FROM subscriptions WHERE match_id=42`,
	).Scan(&status, &product, &subscribedAt))
	require.Equal(t, "subscribed", status)
	require.Equal(t, "live", product)
	require.NotNil(t, subscribedAt)

	var to string
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT to_status FROM subscription_events WHERE match_id=42 ORDER BY id DESC LIMIT 1`,
	).Scan(&to))
	require.Equal(t, "subscribed", to)
}

// Given a live match whose status transitions to ended via catalog
// When catalog fires OnMatchTerminal on the subscription manager
// Then subscriptions.released_at is set with reason="match_ended"
func TestGiven_LiveMatchEnds_When_StatusObserved_Then_AutoUnbookedWithReason(t *testing.T) {
	pool := setup(t)
	cat := seedMatch(t, pool, 42)
	ctx := context.Background()
	d := feed.NewDispatcher(nil)
	mgr := subscription.New(pool, subscription.Options{})
	mgr.Register(d)
	require.NoError(t, d.Dispatch(ctx, feed.MsgSubscriptionBook,
		env(`{"matchId":42,"product":"live"}`), [16]byte{}))

	// Now drive the match to ended via the catalog handler with the
	// subscription manager attached as observer.
	cat2 := cat.WithObserver(mgr)
	require.NoError(t, cat2.Handle(ctx, feed.MsgFixtureChange,
		env(`{"matchId":42,"sportId":1,"status":"ended","statusChange":true}`),
		[16]byte{}))

	var (
		status      string
		reason      string
		releasedAt  *time.Time
	)
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT status, reason, released_at FROM subscriptions WHERE match_id=42`,
	).Scan(&status, &reason, &releasedAt))
	require.Equal(t, "unsubscribed", status)
	require.Equal(t, "match_ended", reason)
	require.NotNil(t, releasedAt)
}

// Given an Unbook delivery for an already-subscribed match
// When the manager handles it
// Then subscriptions.status -> unsubscribed and released_at is set
func TestGiven_Unbook_When_Handled_Then_StatusUnsubscribed(t *testing.T) {
	pool := setup(t)
	seedMatch(t, pool, 42)
	ctx := context.Background()
	d := feed.NewDispatcher(nil)
	mgr := subscription.New(pool, subscription.Options{})
	mgr.Register(d)

	require.NoError(t, d.Dispatch(ctx, feed.MsgSubscriptionBook,
		env(`{"matchId":42,"product":"live"}`), [16]byte{}))
	require.NoError(t, d.Dispatch(ctx, feed.MsgSubscriptionUnbk,
		env(`{"matchId":42,"reason":"client_request"}`), [16]byte{}))

	var (
		status    string
		releasedAt *time.Time
		reason    string
	)
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT status, released_at, reason FROM subscriptions WHERE match_id=42`,
	).Scan(&status, &releasedAt, &reason))
	require.Equal(t, "unsubscribed", status)
	require.NotNil(t, releasedAt)
	require.Equal(t, "client_request", reason)
}

// Given a subscription stuck in requested for longer than StuckTimeout
// When CleanupStuck runs
// Then it transitions to failed with reason="stuck_request"
//      AND a subscription_events row records the transition
func TestGiven_StuckRequest_When_CleanupTick_Then_TransitionsToFailed(t *testing.T) {
	pool := setup(t)
	seedMatch(t, pool, 42)
	ctx := context.Background()

	// Insert a stuck "requested" row 10 minutes in the past.
	old := time.Now().Add(-10 * time.Minute)
	_, err := pool.Exec(ctx, `
		INSERT INTO subscriptions (match_id, product, status, requested_at)
		VALUES (42, 'live', 'requested', $1)`, old)
	require.NoError(t, err)

	mgr := subscription.New(pool, subscription.Options{StuckTimeout: time.Minute})
	n, err := mgr.CleanupStuck(ctx)
	require.NoError(t, err)
	require.Equal(t, int64(1), n)

	var (
		status string
		reason string
	)
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT status, reason FROM subscriptions WHERE match_id=42`,
	).Scan(&status, &reason))
	require.Equal(t, "failed", status)
	require.Equal(t, "stuck_request", reason)

	var to string
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT to_status FROM subscription_events WHERE match_id=42 ORDER BY id DESC LIMIT 1`,
	).Scan(&to))
	require.Equal(t, "failed", to)
}

// Given a previously-released subscription
// When the match terminates again via OnMatchTerminal
// Then the manager makes no further changes (idempotent)
func TestGiven_AlreadyReleased_When_OnMatchTerminal_Then_NoOp(t *testing.T) {
	pool := setup(t)
	seedMatch(t, pool, 42)
	ctx := context.Background()
	d := feed.NewDispatcher(nil)
	mgr := subscription.New(pool, subscription.Options{})
	mgr.Register(d)

	require.NoError(t, d.Dispatch(ctx, feed.MsgSubscriptionBook,
		env(`{"matchId":42,"product":"live"}`), [16]byte{}))
	require.NoError(t, d.Dispatch(ctx, feed.MsgSubscriptionUnbk,
		env(`{"matchId":42}`), [16]byte{}))

	var beforeEvents int
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT count(*) FROM subscription_events WHERE match_id=42`,
	).Scan(&beforeEvents))

	require.NoError(t, mgr.OnMatchTerminal(ctx, 42, "ended"))

	var afterEvents int
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT count(*) FROM subscription_events WHERE match_id=42`,
	).Scan(&afterEvents))
	require.Equal(t, beforeEvents, afterEvents, "no new events for already-released sub")
}
