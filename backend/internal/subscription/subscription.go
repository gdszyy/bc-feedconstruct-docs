// Package subscription owns the booking / unbook / release lifecycle of
// matches. Maps to upload-guideline 业务域 "订阅生命周期" (M11).
//
// The package owns two tables: subscriptions (one row per match_id) and
// subscription_events (append-only history of status transitions). See
// migrations/005_subscriptions.sql.
//
// Inputs:
//   - feed.MsgSubscriptionBook   — FC Book object → status=subscribed
//   - feed.MsgSubscriptionUnbk   — FC Unbook object → status=unsubscribed
//   - feed.MsgFixtureChange      — when the match enters a terminal
//                                  state, the manager schedules a
//                                  release after the configured grace
//                                  period (acceptance 13-b)
//
// Periodic tick:
//   - ProcessDueReleases       — drains the grace-period queue
//   - CleanupStuckRequests     — fails subscriptions stuck in 'requested'
package subscription

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gdszyy/bc-feedconstruct-docs/backend/internal/feed"
)

// Status mirrors the CHECK constraint on subscriptions.status.
type Status string

const (
	StatusRequested    Status = "requested"
	StatusSubscribed   Status = "subscribed"
	StatusUnsubscribed Status = "unsubscribed"
	StatusExpired      Status = "expired"
	StatusFailed       Status = "failed"
)

// Product mirrors the CHECK constraint on subscriptions.product.
type Product string

const (
	ProductLive     Product = "live"
	ProductPrematch Product = "prematch"
)

// Reasons surfaced on subscription_events and the subscriptions.reason
// column. Frontend M11 displays these verbatim.
const (
	ReasonBooked       = "booked"
	ReasonUnbooked     = "unbooked"
	ReasonMatchEnded   = "match_ended"
	ReasonStuckRequest = "stuck_request"
)

// Subscription is one row of the subscriptions table.
type Subscription struct {
	MatchID      int64
	Product      Product
	Status       Status
	RequestedAt  *time.Time
	SubscribedAt *time.Time
	ReleasedAt   *time.Time
	LastEventID  *string
	Reason       *string
}

// Event is one row of the subscription_events table.
type Event struct {
	ID         int64
	MatchID    int64
	FromStatus *Status
	ToStatus   Status
	Reason     *string
	OccurredAt time.Time
}

// Repo abstracts persistence. PgRepo (pgrepo.go) implements it for the
// production stack; unit tests use an in-memory fake.
type Repo interface {
	GetSubscription(ctx context.Context, matchID int64) (Subscription, bool, error)
	UpsertSubscription(ctx context.Context, s Subscription) error
	InsertEvent(ctx context.Context, e Event) error
	// ListByStatus returns every subscription currently in the given
	// status. Used by CleanupStuckRequests to find rows that have
	// outlived the configured grace window.
	ListByStatus(ctx context.Context, status Status) ([]Subscription, error)
}

// Manager orchestrates subscriptions. Construct via New, then call
// Register to bind it to the feed dispatcher.
type Manager struct {
	Repo Repo
	// Now is injectable so tests can produce deterministic timestamps.
	Now func() time.Time
	// Grace is the delay between observing a terminal match status and
	// auto-releasing the subscription. Zero releases immediately.
	Grace time.Duration
	// StuckAfter is the threshold beyond which a 'requested' row is
	// considered stuck and gets transitioned to 'failed'. Defaults to
	// 5 minutes when zero.
	StuckAfter time.Duration

	mu sync.Mutex
	// pendingReleases maps match_id → earliest time at which the
	// auto-release may apply. Only populated when Grace > 0.
	pendingReleases map[int64]time.Time

	bookCount    atomic.Int64
	unbookCount  atomic.Int64
	autoRelease  atomic.Int64
	stuckFailed  atomic.Int64
}

// New returns a Manager bound to repo with sensible defaults.
func New(repo Repo) *Manager {
	return &Manager{Repo: repo, pendingReleases: make(map[int64]time.Time)}
}

// BookCount returns the number of Book deliveries persisted.
func (m *Manager) BookCount() int64 { return m.bookCount.Load() }

// UnbookCount returns the number of Unbook deliveries persisted.
func (m *Manager) UnbookCount() int64 { return m.unbookCount.Load() }

// AutoReleased returns the number of subscriptions released by
// OnMatchStatusChange / ProcessDueReleases.
func (m *Manager) AutoReleased() int64 { return m.autoRelease.Load() }

// StuckFailed returns the number of subscriptions failed by
// CleanupStuckRequests.
func (m *Manager) StuckFailed() int64 { return m.stuckFailed.Load() }

// Register wires the Book / Unbook handlers into a feed dispatcher.
// MsgFixtureChange is deliberately NOT registered here because the
// catalog handler already owns that route. Wire HandleFixtureChange in
// alongside the catalog handler at the cmd/bffd composition layer
// (e.g. via a composite handler).
func (m *Manager) Register(d *feed.Dispatcher) {
	d.Register(feed.MsgSubscriptionBook, feed.HandlerFunc(m.HandleBook))
	d.Register(feed.MsgSubscriptionUnbk, feed.HandlerFunc(m.HandleUnbook))
}

func (m *Manager) now() time.Time {
	if m.Now != nil {
		return m.Now()
	}
	return time.Now()
}

func (m *Manager) stuckAfter() time.Duration {
	if m.StuckAfter > 0 {
		return m.StuckAfter
	}
	return 5 * time.Minute
}
