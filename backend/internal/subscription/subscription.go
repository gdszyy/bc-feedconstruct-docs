// Package subscription owns the booking / unbook / release lifecycle of
// matches. Maps to upload-guideline 业务域 "订阅生命周期" and to the
// frontend module M11 in docs/07_frontend_architecture/modules/.
//
// The package owns two tables: subscriptions (one row per match_id) and
// subscription_events (append-only audit log of state transitions). See
// migrations/005_subscriptions.sql.
//
// Three lifecycle inputs feed the Manager:
//
//  1. Book / Unbook envelopes from the FeedConstruct partner queue
//     (MsgSubscriptionBook / MsgSubscriptionUnbk). These promote a
//     subscription from requested to subscribed (or to unsubscribed when
//     IsSubscribed is false).
//
//  2. Match status changes observed by the catalog handler. When a match
//     transitions to ended/closed/cancelled/abandoned the Manager auto
//     unbooks within the configured grace period and records the reason.
//
//  3. Periodic CleanupTick. Subscriptions stuck in status=requested for
//     longer than StuckRequestTimeout transition to failed.
package subscription

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/gdszyy/bc-feedconstruct-docs/backend/internal/catalog"
	"github.com/gdszyy/bc-feedconstruct-docs/backend/internal/feed"
)

// Status mirrors the values enforced by the CHECK constraint on
// subscriptions.status (see migrations/005_subscriptions.sql).
type Status string

const (
	StatusUnknown      Status = ""
	StatusRequested    Status = "requested"
	StatusSubscribed   Status = "subscribed"
	StatusUnsubscribed Status = "unsubscribed"
	StatusExpired      Status = "expired"
	StatusFailed       Status = "failed"
)

// Product mirrors the values enforced by the CHECK constraint on
// subscriptions.product.
type Product string

const (
	ProductLive     Product = "live"
	ProductPrematch Product = "prematch"
)

// Reason values used in subscription_events.reason and (when terminal)
// subscriptions.reason. The enumeration is open: callers may pass any
// short snake_case string, but these constants document the canonical
// values the Manager itself emits.
const (
	ReasonBookOK         = "book_ok"
	ReasonUnbookOK       = "unbook_ok"
	ReasonMatchEnded     = "match_ended"
	ReasonMatchClosed    = "match_closed"
	ReasonMatchCancelled = "match_cancelled"
	ReasonMatchAbandoned = "match_abandoned"
	ReasonStuckRequest   = "stuck_request"
)

// Subscription is one row of the subscriptions table.
type Subscription struct {
	MatchID       int64
	Product       Product
	Status        Status
	RequestedAt   *time.Time
	SubscribedAt  *time.Time
	ReleasedAt    *time.Time
	LastEventID   string
	Reason        string
}

// Event is one row of the subscription_events table. Events are
// append-only and ordered by occurred_at (then id) when reconstructing
// the lifecycle.
type Event struct {
	ID         int64
	MatchID    int64
	From       Status
	To         Status
	Reason     string
	OccurredAt time.Time
}

// Repo abstracts persistence. PgRepo implements it for the production
// stack; unit tests use an in-memory fake that mirrors the two tables.
type Repo interface {
	GetSubscription(ctx context.Context, matchID int64) (Subscription, bool, error)
	UpsertSubscription(ctx context.Context, s Subscription) error
	InsertEvent(ctx context.Context, e Event) error

	// ListStuckRequests returns subscriptions whose status=requested and
	// whose requested_at is at or before olderThan. Used by CleanupTick.
	ListStuckRequests(ctx context.Context, olderThan time.Time) ([]Subscription, error)
}

// Logger receives lifecycle observations the operator may want to see.
// Pass nil to silently drop.
type Logger interface {
	TransitionApplied(matchID int64, from, to Status, reason string)
	StuckRequestExpired(matchID int64)
}

// LoggerFunc adapts a function pair into a Logger. Nil sub-functions
// short circuit to a no-op.
type LoggerFunc struct {
	OnTransition func(matchID int64, from, to Status, reason string)
	OnStuck      func(matchID int64)
}

// TransitionApplied implements Logger.
func (f LoggerFunc) TransitionApplied(matchID int64, from, to Status, reason string) {
	if f.OnTransition != nil {
		f.OnTransition(matchID, from, to, reason)
	}
}

// StuckRequestExpired implements Logger.
func (f LoggerFunc) StuckRequestExpired(matchID int64) {
	if f.OnStuck != nil {
		f.OnStuck(matchID)
	}
}

// DefaultStuckRequestTimeout matches BDD scenario 3 in manager_test.go:
// a subscription stuck in status=requested for longer than five minutes
// transitions to failed when CleanupTick runs.
const DefaultStuckRequestTimeout = 5 * time.Minute

// Manager owns the subscription lifecycle. Construct via New, then call
// Register to bind the Book/Unbook handlers to the feed dispatcher and
// AttachToCatalog to receive auto-unbook signals.
type Manager struct {
	Repo   Repo
	Logger Logger

	// Now is injectable so tests get strictly increasing timestamps.
	Now func() time.Time

	// StuckRequestTimeout controls when CleanupTick promotes a stuck
	// requested row to failed. Defaults to DefaultStuckRequestTimeout.
	StuckRequestTimeout time.Duration

	// AutoUnbookOnTerminal opts into auto-release when a match
	// transitions to a terminal lifecycle status (ended / closed /
	// cancelled / abandoned). Defaults to true; tests can disable to
	// keep the catalog hook observable independently.
	AutoUnbookOnTerminal bool

	bookCount        atomic.Int64
	unbookCount      atomic.Int64
	autoReleaseCount atomic.Int64
	stuckExpired     atomic.Int64
}

// New returns a Manager bound to repo with safe defaults.
func New(repo Repo) *Manager {
	return &Manager{
		Repo:                 repo,
		StuckRequestTimeout:  DefaultStuckRequestTimeout,
		AutoUnbookOnTerminal: true,
	}
}

// BookCount returns how many Book deliveries promoted a subscription to
// status=subscribed since construction.
func (m *Manager) BookCount() int64 { return m.bookCount.Load() }

// UnbookCount returns how many Unbook deliveries marked a subscription
// as released (status=unsubscribed) since construction.
func (m *Manager) UnbookCount() int64 { return m.unbookCount.Load() }

// AutoReleaseCount returns how many auto-unbook releases were triggered
// by terminal match status transitions.
func (m *Manager) AutoReleaseCount() int64 { return m.autoReleaseCount.Load() }

// StuckExpiredCount returns how many requested rows CleanupTick
// transitioned to failed.
func (m *Manager) StuckExpiredCount() int64 { return m.stuckExpired.Load() }

// Register wires the subscription handler into a feed dispatcher.
func (m *Manager) Register(d *feed.Dispatcher) {
	d.Register(feed.MsgSubscriptionBook, feed.HandlerFunc(m.HandleBook))
	d.Register(feed.MsgSubscriptionUnbk, feed.HandlerFunc(m.HandleUnbook))
}

// AttachToCatalog installs Manager as a MatchObserver on the catalog
// handler. Whenever the catalog observes a real status transition the
// Manager evaluates whether to release the subscription.
func (m *Manager) AttachToCatalog(h *catalog.Handler) {
	h.Observer = catalog.MatchObserverFunc(m.OnMatchStatusChanged)
}

func (m *Manager) now() time.Time {
	if m.Now != nil {
		return m.Now()
	}
	return time.Now().UTC()
}

func (m *Manager) timeout() time.Duration {
	if m.StuckRequestTimeout > 0 {
		return m.StuckRequestTimeout
	}
	return DefaultStuckRequestTimeout
}

func (m *Manager) emit(matchID int64, from, to Status, reason string) {
	if m.Logger == nil {
		return
	}
	m.Logger.TransitionApplied(matchID, from, to, reason)
}
