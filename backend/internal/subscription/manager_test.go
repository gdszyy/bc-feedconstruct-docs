package subscription_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/gdszyy/bc-feedconstruct-docs/backend/internal/feed"
	"github.com/gdszyy/bc-feedconstruct-docs/backend/internal/subscription"
)

// 验收 13 — 订阅生命周期（M11）
//
// Given a Book object delivered for match=42 (product=live)
// When subscription.Manager handles it
// Then a subscriptions row is upserted with status=subscribed
//      AND a subscription_events row records the transition requested→subscribed
func TestGiven_BookDelivered_When_Handled_Then_SubscriptionUpsertedAndEventRecorded(t *testing.T) {
	repo := newFakeRepo()
	m := subscription.New(repo)
	clock := time.Date(2026, 5, 14, 10, 0, 0, 0, time.UTC)
	m.Now = func() time.Time { return clock }

	body := `{"matchId":42,"product":"live","book":true}`
	require.NoError(t, m.HandleBook(context.Background(), feed.MsgSubscriptionBook,
		envWith(body), [16]byte{0x01}))

	subs := repo.snapshotSubs()
	require.Len(t, subs, 1)
	s := subs[42]
	require.EqualValues(t, 42, s.MatchID)
	require.Equal(t, subscription.ProductLive, s.Product)
	require.Equal(t, subscription.StatusSubscribed, s.Status,
		"acceptance 13-a: Book delivery must put the subscription in status=subscribed")
	require.NotNil(t, s.SubscribedAt)
	require.Equal(t, clock, *s.SubscribedAt)
	require.Nil(t, s.ReleasedAt)

	events := repo.snapshotEvents()
	require.Len(t, events, 1)
	ev := events[0]
	require.EqualValues(t, 42, ev.MatchID)
	require.NotNil(t, ev.FromStatus)
	require.Equal(t, subscription.StatusRequested, *ev.FromStatus,
		"acceptance 13-a: event must record requested→subscribed transition")
	require.Equal(t, subscription.StatusSubscribed, ev.ToStatus)
	require.EqualValues(t, 1, m.BookCount())
}

// Given a live match whose status transitions to ended
// When the catalog handler emits the status change
// Then subscription.Manager auto-unbooks within the configured grace period
//      AND subscriptions.released_at is set with reason="match_ended"
func TestGiven_LiveMatchEnds_When_StatusObserved_Then_AutoUnbookedWithReason(t *testing.T) {
	repo := newFakeRepo()
	m := subscription.New(repo)
	clock := time.Date(2026, 5, 14, 12, 0, 0, 0, time.UTC)
	m.Now = func() time.Time { return clock }
	m.Grace = 5 * time.Minute

	// Seed: match=42 is currently subscribed (Book happened earlier).
	requested := time.Date(2026, 5, 14, 11, 0, 0, 0, time.UTC)
	subscribed := time.Date(2026, 5, 14, 11, 0, 5, 0, time.UTC)
	repo.seed(subscription.Subscription{
		MatchID: 42, Product: subscription.ProductLive,
		Status: subscription.StatusSubscribed,
		RequestedAt: &requested, SubscribedAt: &subscribed,
	})

	// The catalog handler routes fixture_change frames through us.
	body := `{"matchId":42,"status":"ended"}`
	require.NoError(t, m.HandleFixtureChange(context.Background(), feed.MsgFixtureChange,
		envWith(body), [16]byte{}))

	// Within the grace window — not yet released.
	require.Equal(t, subscription.StatusSubscribed, repo.snapshotSubs()[42].Status,
		"release must not fire before the grace window elapses")

	// Advance past grace then drain the queue.
	clock = clock.Add(5*time.Minute + time.Second)
	released, err := m.ProcessDueReleases(context.Background())
	require.NoError(t, err)
	require.Equal(t, 1, released)

	s := repo.snapshotSubs()[42]
	require.Equal(t, subscription.StatusUnsubscribed, s.Status,
		"acceptance 13-b: subscription must be auto-released after grace")
	require.NotNil(t, s.ReleasedAt)
	require.Equal(t, clock, *s.ReleasedAt)
	require.NotNil(t, s.Reason)
	require.Equal(t, subscription.ReasonMatchEnded, *s.Reason)

	// The audit log shows the transition.
	events := repo.snapshotEvents()
	require.Len(t, events, 1)
	require.Equal(t, subscription.StatusUnsubscribed, events[0].ToStatus)
	require.NotNil(t, events[0].Reason)
	require.Equal(t, subscription.ReasonMatchEnded, *events[0].Reason)
	require.EqualValues(t, 1, m.AutoReleased())
}

// Given a subscription stuck in status=requested for >5 minutes
// When the cleanup tick runs
// Then status transitions to failed and a subscription_events row records reason="stuck_request"
func TestGiven_StuckRequest_When_CleanupTick_Then_TransitionsToFailed(t *testing.T) {
	repo := newFakeRepo()
	m := subscription.New(repo)
	now := time.Date(2026, 5, 14, 12, 0, 0, 0, time.UTC)
	m.Now = func() time.Time { return now }
	m.StuckAfter = 5 * time.Minute

	// One stuck row (requested 6 minutes ago) and one fresh row that
	// must NOT be touched.
	old := now.Add(-6 * time.Minute)
	fresh := now.Add(-1 * time.Minute)
	repo.seed(subscription.Subscription{
		MatchID: 42, Product: subscription.ProductLive,
		Status: subscription.StatusRequested, RequestedAt: &old,
	})
	repo.seed(subscription.Subscription{
		MatchID: 43, Product: subscription.ProductLive,
		Status: subscription.StatusRequested, RequestedAt: &fresh,
	})

	failed, err := m.CleanupStuckRequests(context.Background())
	require.NoError(t, err)
	require.Equal(t, 1, failed)

	subs := repo.snapshotSubs()
	require.Equal(t, subscription.StatusFailed, subs[42].Status,
		"acceptance 13-c: stuck request must transition to failed")
	require.NotNil(t, subs[42].Reason)
	require.Equal(t, subscription.ReasonStuckRequest, *subs[42].Reason)
	require.Equal(t, subscription.StatusRequested, subs[43].Status,
		"fresh requested row must not be touched")

	events := repo.snapshotEvents()
	require.Len(t, events, 1)
	require.EqualValues(t, 42, events[0].MatchID)
	require.NotNil(t, events[0].FromStatus)
	require.Equal(t, subscription.StatusRequested, *events[0].FromStatus)
	require.Equal(t, subscription.StatusFailed, events[0].ToStatus)
	require.NotNil(t, events[0].Reason)
	require.Equal(t, subscription.ReasonStuckRequest, *events[0].Reason)
	require.EqualValues(t, 1, m.StuckFailed())
}
