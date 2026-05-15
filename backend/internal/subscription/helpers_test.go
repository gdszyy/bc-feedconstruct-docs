package subscription_test

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/gdszyy/bc-feedconstruct-docs/backend/internal/subscription"
)

// fakeRepo is the in-memory Repo used by every subscription_test file.
// It mirrors the subscriptions and subscription_events tables 1:1 and
// protects them with a single mutex so concurrent scenarios remain race
// clean.
type fakeRepo struct {
	mu     sync.Mutex
	rows   map[int64]subscription.Subscription
	events []subscription.Event
	nextID int64
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{rows: map[int64]subscription.Subscription{}}
}

func (r *fakeRepo) GetSubscription(_ context.Context, matchID int64) (subscription.Subscription, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	s, ok := r.rows[matchID]
	if !ok {
		return subscription.Subscription{}, false, nil
	}
	// Return a copy so callers cannot mutate the store through the
	// pointer aliases on *time.Time.
	return cloneSubscription(s), true, nil
}

func (r *fakeRepo) UpsertSubscription(_ context.Context, s subscription.Subscription) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	// Preserve audit timestamps when the caller omitted them (mirrors
	// the PgRepo COALESCE-on-conflict semantics).
	if cur, ok := r.rows[s.MatchID]; ok {
		if s.RequestedAt == nil {
			s.RequestedAt = cur.RequestedAt
		}
		if s.SubscribedAt == nil {
			s.SubscribedAt = cur.SubscribedAt
		}
		if s.ReleasedAt == nil {
			s.ReleasedAt = cur.ReleasedAt
		}
		if s.LastEventID == "" {
			s.LastEventID = cur.LastEventID
		}
		if s.Reason == "" {
			s.Reason = cur.Reason
		}
		if s.Product == "" {
			s.Product = cur.Product
		}
	}
	r.rows[s.MatchID] = cloneSubscription(s)
	return nil
}

func (r *fakeRepo) InsertEvent(_ context.Context, e subscription.Event) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.nextID++
	e.ID = r.nextID
	r.events = append(r.events, e)
	return nil
}

func (r *fakeRepo) ListStuckRequests(_ context.Context, olderThan time.Time) ([]subscription.Subscription, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var out []subscription.Subscription
	ids := make([]int64, 0, len(r.rows))
	for id := range r.rows {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	for _, id := range ids {
		s := r.rows[id]
		if s.Status != subscription.StatusRequested {
			continue
		}
		if s.RequestedAt == nil {
			continue
		}
		if s.RequestedAt.After(olderThan) {
			continue
		}
		out = append(out, cloneSubscription(s))
	}
	return out, nil
}

func (r *fakeRepo) snapshot() (rows map[int64]subscription.Subscription, events []subscription.Event) {
	r.mu.Lock()
	defer r.mu.Unlock()
	rows = make(map[int64]subscription.Subscription, len(r.rows))
	for k, v := range r.rows {
		rows[k] = cloneSubscription(v)
	}
	events = append(events, r.events...)
	return rows, events
}

func (r *fakeRepo) seed(s subscription.Subscription) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.rows[s.MatchID] = cloneSubscription(s)
}

func cloneSubscription(s subscription.Subscription) subscription.Subscription {
	out := s
	if s.RequestedAt != nil {
		v := *s.RequestedAt
		out.RequestedAt = &v
	}
	if s.SubscribedAt != nil {
		v := *s.SubscribedAt
		out.SubscribedAt = &v
	}
	if s.ReleasedAt != nil {
		v := *s.ReleasedAt
		out.ReleasedAt = &v
	}
	return out
}

type captureLogger struct {
	mu          sync.Mutex
	transitions []loggedTransition
	stuck       []int64
}

type loggedTransition struct {
	MatchID int64
	From    subscription.Status
	To      subscription.Status
	Reason  string
}

func (c *captureLogger) TransitionApplied(matchID int64, from, to subscription.Status, reason string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.transitions = append(c.transitions, loggedTransition{matchID, from, to, reason})
}

func (c *captureLogger) StuckRequestExpired(matchID int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.stuck = append(c.stuck, matchID)
}

func (c *captureLogger) snapshotTransitions() []loggedTransition {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]loggedTransition, len(c.transitions))
	copy(out, c.transitions)
	return out
}

func ptrTime(t time.Time) *time.Time { return &t }

func ptrInt64(v int64) *int64 { return &v }
