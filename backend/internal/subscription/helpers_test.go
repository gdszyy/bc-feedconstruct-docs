package subscription_test

import (
	"context"
	"sort"
	"sync"

	"github.com/gdszyy/bc-feedconstruct-docs/backend/internal/feed"
	"github.com/gdszyy/bc-feedconstruct-docs/backend/internal/subscription"
)

// fakeRepo mirrors the subscriptions / subscription_events tables in
// memory. Shared by every *_test.go in this package.
type fakeRepo struct {
	mu sync.Mutex

	subs    map[int64]subscription.Subscription
	events  []subscription.Event
	eventID int64
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{subs: map[int64]subscription.Subscription{}}
}

func (r *fakeRepo) seed(s subscription.Subscription) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.subs[s.MatchID] = s
}

func (r *fakeRepo) GetSubscription(_ context.Context, matchID int64) (subscription.Subscription, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	s, ok := r.subs[matchID]
	return s, ok, nil
}

func (r *fakeRepo) UpsertSubscription(_ context.Context, s subscription.Subscription) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.subs[s.MatchID] = s
	return nil
}

func (r *fakeRepo) InsertEvent(_ context.Context, e subscription.Event) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.eventID++
	e.ID = r.eventID
	r.events = append(r.events, e)
	return nil
}

func (r *fakeRepo) ListByStatus(_ context.Context, status subscription.Status) ([]subscription.Subscription, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]subscription.Subscription, 0)
	for _, s := range r.subs {
		if s.Status == status {
			out = append(out, s)
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].MatchID < out[j].MatchID })
	return out, nil
}

func (r *fakeRepo) snapshotSubs() map[int64]subscription.Subscription {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make(map[int64]subscription.Subscription, len(r.subs))
	for k, v := range r.subs {
		out[k] = v
	}
	return out
}

func (r *fakeRepo) snapshotEvents() []subscription.Event {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]subscription.Event(nil), r.events...)
}

func envWith(payload string) feed.Envelope {
	e, err := feed.DecodeEnvelope([]byte(payload))
	if err != nil {
		e = feed.Envelope{Payload: []byte(payload)}
	}
	e.Payload = []byte(payload)
	return e
}
