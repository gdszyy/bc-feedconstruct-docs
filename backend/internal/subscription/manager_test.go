package subscription_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/gdszyy/bc-feedconstruct-docs/backend/internal/catalog"
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
	mgr := subscription.New(repo)
	mgr.Now = func() time.Time { return time.Date(2026, 5, 14, 12, 0, 0, 0, time.UTC) }

	// Seed a Requested row to verify the explicit Requested→Subscribed transition.
	requestedAt := time.Date(2026, 5, 14, 11, 59, 30, 0, time.UTC)
	repo.seed(subscription.Subscription{
		MatchID:     42,
		Product:     subscription.ProductLive,
		Status:      subscription.StatusRequested,
		RequestedAt: &requestedAt,
	})

	payload := []byte(`{"objectId":42,"objectTypeId":4,"isLive":true,"isSubscribed":true}`)
	env := feed.Envelope{
		ObjectType: 4,
		MatchID:    ptrInt64(42),
		Book:       true,
		Payload:    payload,
	}

	require.NoError(t, mgr.HandleBook(context.Background(), feed.MsgSubscriptionBook, env, [16]byte{}))

	rows, events := repo.snapshot()
	require.Len(t, rows, 1)
	row, ok := rows[42]
	require.True(t, ok)
	require.Equal(t, subscription.ProductLive, row.Product)
	require.Equal(t, subscription.StatusSubscribed, row.Status)
	require.NotNil(t, row.RequestedAt)
	require.True(t, row.RequestedAt.Equal(requestedAt),
		"requested_at must be preserved across the transition")
	require.NotNil(t, row.SubscribedAt)
	require.True(t, row.SubscribedAt.Equal(time.Date(2026, 5, 14, 12, 0, 0, 0, time.UTC)))
	require.Nil(t, row.ReleasedAt)

	require.Len(t, events, 1)
	require.Equal(t, int64(42), events[0].MatchID)
	require.Equal(t, subscription.StatusRequested, events[0].From)
	require.Equal(t, subscription.StatusSubscribed, events[0].To)
	require.Equal(t, subscription.ReasonBookOK, events[0].Reason)

	require.Equal(t, int64(1), mgr.BookCount())
}

// Given an already-subscribed match
// When a duplicate Book delivery arrives
// Then the upsert is collapsed and no new subscription_events row is appended
func TestGiven_DuplicateBook_When_Handled_Then_NoNewEvent(t *testing.T) {
	repo := newFakeRepo()
	mgr := subscription.New(repo)
	mgr.Now = func() time.Time { return time.Date(2026, 5, 14, 12, 0, 5, 0, time.UTC) }

	subscribedAt := time.Date(2026, 5, 14, 12, 0, 0, 0, time.UTC)
	repo.seed(subscription.Subscription{
		MatchID:      42,
		Product:      subscription.ProductLive,
		Status:       subscription.StatusSubscribed,
		SubscribedAt: &subscribedAt,
		LastEventID:  "evt-1",
	})

	env := feed.Envelope{
		ObjectType: 4,
		MatchID:    ptrInt64(42),
		Book:       true,
		EventID:    "evt-1",
		Payload:    []byte(`{"objectId":42,"objectTypeId":4,"isLive":true,"isSubscribed":true,"eventId":"evt-1"}`),
	}

	require.NoError(t, mgr.HandleBook(context.Background(), feed.MsgSubscriptionBook, env, [16]byte{}))

	_, events := repo.snapshot()
	require.Empty(t, events, "duplicate book must not append an event")
	require.Equal(t, int64(0), mgr.BookCount())
}

// Given a live match whose status transitions to ended
// When the catalog handler emits the status change
// Then subscription.Manager auto-unbooks within the configured grace period
//      AND subscriptions.released_at is set with reason="match_ended"
func TestGiven_LiveMatchEnds_When_StatusObserved_Then_AutoUnbookedWithReason(t *testing.T) {
	repo := newFakeRepo()
	mgr := subscription.New(repo)
	now := time.Date(2026, 5, 14, 13, 0, 0, 0, time.UTC)
	mgr.Now = func() time.Time { return now }

	logger := &captureLogger{}
	mgr.Logger = logger

	subscribedAt := time.Date(2026, 5, 14, 12, 0, 0, 0, time.UTC)
	repo.seed(subscription.Subscription{
		MatchID:      42,
		Product:      subscription.ProductLive,
		Status:       subscription.StatusSubscribed,
		SubscribedAt: &subscribedAt,
	})

	// Simulate the catalog hand-off.
	mgr.OnMatchStatusChanged(context.Background(), 42, catalog.StatusLive, catalog.StatusEnded)

	rows, events := repo.snapshot()
	row, ok := rows[42]
	require.True(t, ok)
	require.Equal(t, subscription.StatusUnsubscribed, row.Status)
	require.NotNil(t, row.ReleasedAt)
	require.True(t, row.ReleasedAt.Equal(now))
	require.Equal(t, subscription.ReasonMatchEnded, row.Reason)

	require.Len(t, events, 1)
	require.Equal(t, subscription.StatusSubscribed, events[0].From)
	require.Equal(t, subscription.StatusUnsubscribed, events[0].To)
	require.Equal(t, subscription.ReasonMatchEnded, events[0].Reason)

	require.Equal(t, int64(1), mgr.AutoReleaseCount())

	tr := logger.snapshotTransitions()
	require.Len(t, tr, 1)
	require.Equal(t, subscription.StatusSubscribed, tr[0].From)
	require.Equal(t, subscription.StatusUnsubscribed, tr[0].To)
	require.Equal(t, subscription.ReasonMatchEnded, tr[0].Reason)
}

// Given the catalog Handler with an Observer attached
// When HandleMatch persists an effective Live→Ended transition
// Then the Observer is invoked exactly once with the right statuses,
//      so the subscription Manager actually unhooks via the catalog path.
func TestGiven_CatalogObserverAttached_When_MatchEnds_Then_ManagerReleases(t *testing.T) {
	repo := newFakeRepo()
	mgr := subscription.New(repo)
	mgr.Now = func() time.Time { return time.Date(2026, 5, 14, 13, 0, 0, 0, time.UTC) }

	subscribedAt := time.Date(2026, 5, 14, 12, 0, 0, 0, time.UTC)
	repo.seed(subscription.Subscription{
		MatchID:      42,
		Product:      subscription.ProductLive,
		Status:       subscription.StatusSubscribed,
		SubscribedAt: &subscribedAt,
	})

	catRepo := &catalogRepoStub{
		matches: map[int64]catalog.Match{
			42: {ID: 42, SportID: 1, Status: catalog.StatusLive, IsLive: true},
		},
		sports: map[int32]catalog.Sport{1: {ID: 1, IsActive: true}},
	}
	catHandler := catalog.New(catRepo)
	mgr.AttachToCatalog(catHandler)

	payload := []byte(`{"id":42,"sportId":1,"status":"ended","isLive":false}`)
	env := feed.Envelope{
		ObjectType:   4,
		MatchID:      ptrInt64(42),
		StatusChange: true,
		EventID:      "ended-1",
		Payload:      payload,
	}
	require.NoError(t, catHandler.HandleMatch(context.Background(), feed.MsgFixtureChange, env, [16]byte{}))

	rows, events := repo.snapshot()
	require.Equal(t, subscription.StatusUnsubscribed, rows[42].Status)
	require.Equal(t, subscription.ReasonMatchEnded, rows[42].Reason)
	require.Len(t, events, 1)
	require.Equal(t, subscription.ReasonMatchEnded, events[0].Reason)
}

// Given a subscription stuck in status=requested for >5 minutes
// When the cleanup tick runs
// Then status transitions to failed and a subscription_events row records reason="stuck_request"
func TestGiven_StuckRequest_When_CleanupTick_Then_TransitionsToFailed(t *testing.T) {
	repo := newFakeRepo()
	mgr := subscription.New(repo)
	now := time.Date(2026, 5, 14, 12, 6, 0, 0, time.UTC)
	mgr.Now = func() time.Time { return now }
	logger := &captureLogger{}
	mgr.Logger = logger

	// One subscription 6 minutes old (stuck) and one 1 minute old (fresh).
	repo.seed(subscription.Subscription{
		MatchID:     42,
		Product:     subscription.ProductLive,
		Status:      subscription.StatusRequested,
		RequestedAt: ptrTime(time.Date(2026, 5, 14, 12, 0, 0, 0, time.UTC)),
	})
	repo.seed(subscription.Subscription{
		MatchID:     99,
		Product:     subscription.ProductLive,
		Status:      subscription.StatusRequested,
		RequestedAt: ptrTime(time.Date(2026, 5, 14, 12, 5, 0, 0, time.UTC)),
	})

	moved, err := mgr.CleanupTick(context.Background())
	require.NoError(t, err)
	require.Equal(t, 1, moved)

	rows, events := repo.snapshot()
	require.Equal(t, subscription.StatusFailed, rows[42].Status)
	require.Equal(t, subscription.ReasonStuckRequest, rows[42].Reason)
	require.NotNil(t, rows[42].ReleasedAt)
	require.True(t, rows[42].ReleasedAt.Equal(now))

	require.Equal(t, subscription.StatusRequested, rows[99].Status,
		"fresh request must not be promoted to failed")

	require.Len(t, events, 1)
	require.Equal(t, int64(42), events[0].MatchID)
	require.Equal(t, subscription.StatusRequested, events[0].From)
	require.Equal(t, subscription.StatusFailed, events[0].To)
	require.Equal(t, subscription.ReasonStuckRequest, events[0].Reason)

	require.Equal(t, int64(1), mgr.StuckExpiredCount())
	tr := logger.snapshotTransitions()
	require.Len(t, tr, 1)
	require.Equal(t, subscription.StatusFailed, tr[0].To)
}

// Given a Book delivery whose payload sets isSubscribed=false
// When HandleBook routes it
// Then it is treated as an Unbook and the subscription is released
func TestGiven_BookWithIsSubscribedFalse_When_Handled_Then_TreatedAsUnbook(t *testing.T) {
	repo := newFakeRepo()
	mgr := subscription.New(repo)
	now := time.Date(2026, 5, 14, 12, 5, 0, 0, time.UTC)
	mgr.Now = func() time.Time { return now }

	repo.seed(subscription.Subscription{
		MatchID: 42,
		Product: subscription.ProductLive,
		Status:  subscription.StatusSubscribed,
	})
	env := feed.Envelope{
		ObjectType: 4,
		MatchID:    ptrInt64(42),
		Book:       true,
		Payload:    []byte(`{"objectId":42,"isLive":true,"isSubscribed":false}`),
	}

	require.NoError(t, mgr.HandleBook(context.Background(), feed.MsgSubscriptionBook, env, [16]byte{}))

	rows, events := repo.snapshot()
	require.Equal(t, subscription.StatusUnsubscribed, rows[42].Status)
	require.Equal(t, subscription.ReasonUnbookOK, rows[42].Reason)
	require.Len(t, events, 1)
	require.Equal(t, subscription.StatusSubscribed, events[0].From)
}

// catalogRepoStub is the minimum catalog.Repo needed to exercise the
// catalog→subscription wiring end-to-end. Only the methods touched by
// HandleMatch are populated; the others return zero values.
type catalogRepoStub struct {
	matches      map[int64]catalog.Match
	sports       map[int32]catalog.Sport
	regions      map[int32]catalog.Region
	competitions map[int32]catalog.Competition
	fixtures     []catalog.FixtureChangeRow
}

func (s *catalogRepoStub) UpsertSport(_ context.Context, sp catalog.Sport) error {
	if s.sports == nil {
		s.sports = map[int32]catalog.Sport{}
	}
	s.sports[sp.ID] = sp
	return nil
}
func (s *catalogRepoStub) SoftDeleteSport(_ context.Context, id int32) error {
	if s.sports == nil {
		s.sports = map[int32]catalog.Sport{}
	}
	sp := s.sports[id]
	sp.IsActive = false
	s.sports[id] = sp
	return nil
}
func (s *catalogRepoStub) GetSport(_ context.Context, id int32) (catalog.Sport, bool, error) {
	sp, ok := s.sports[id]
	return sp, ok, nil
}
func (s *catalogRepoStub) UpsertRegion(_ context.Context, r catalog.Region) error {
	if s.regions == nil {
		s.regions = map[int32]catalog.Region{}
	}
	s.regions[r.ID] = r
	return nil
}
func (s *catalogRepoStub) GetRegion(_ context.Context, id int32) (catalog.Region, bool, error) {
	r, ok := s.regions[id]
	return r, ok, nil
}
func (s *catalogRepoStub) UpsertCompetition(_ context.Context, c catalog.Competition) error {
	if s.competitions == nil {
		s.competitions = map[int32]catalog.Competition{}
	}
	s.competitions[c.ID] = c
	return nil
}
func (s *catalogRepoStub) UpsertMatch(_ context.Context, m catalog.Match) error {
	if s.matches == nil {
		s.matches = map[int64]catalog.Match{}
	}
	s.matches[m.ID] = m
	return nil
}
func (s *catalogRepoStub) GetMatch(_ context.Context, id int64) (catalog.Match, bool, error) {
	m, ok := s.matches[id]
	return m, ok, nil
}
func (s *catalogRepoStub) InsertFixtureChange(_ context.Context, row catalog.FixtureChangeRow) error {
	s.fixtures = append(s.fixtures, row)
	return nil
}
