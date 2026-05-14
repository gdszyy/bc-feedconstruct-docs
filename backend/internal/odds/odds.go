// Package odds handles FeedConstruct odds_change and bet_stop deliveries.
// It maintains the markets / outcomes / market_status_history tables and
// enforces market-level anti-regression (settled / cancelled / handed_over
// cannot regress to active).
//
// Maps to M05 (Markets & Odds), M06 (Market Status Machine) and M07
// (Bet Stop) in docs/07_frontend_architecture/modules/.
package odds

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/gdszyy/bc-feedconstruct-docs/backend/internal/feed"
)

// MarketStatus mirrors the CHECK constraint in migrations/003_markets.sql.
type MarketStatus string

const (
	StatusUnknown    MarketStatus = ""
	StatusActive     MarketStatus = "active"
	StatusSuspended  MarketStatus = "suspended"
	StatusDeactivated MarketStatus = "deactivated"
	StatusSettled    MarketStatus = "settled"
	StatusCancelled  MarketStatus = "cancelled"
	StatusHandedOver MarketStatus = "handed_over"
)

// statusRank orders the market lifecycle. Anti-regression rejects any
// incoming status whose rank is lower than the persisted one. Terminal
// states (settled / cancelled / handed_over) sit above the operational
// ones so a stray odds_change cannot reopen a settled book.
var statusRank = map[MarketStatus]int{
	StatusActive:      1,
	StatusSuspended:   2,
	StatusDeactivated: 3,
	StatusSettled:     10,
	StatusHandedOver:  11,
	StatusCancelled:   20,
}

func rankOf(s MarketStatus) (int, bool) {
	r, ok := statusRank[s]
	return r, ok
}

// allowsTransition reports whether `to` may overwrite `from`. Empty from
// always passes (first sight of a market). Empty to is rejected because
// callers always normalise unknown strings to "".
func allowsTransition(from, to MarketStatus) bool {
	if from == StatusUnknown {
		return true
	}
	if to == StatusUnknown {
		return false
	}
	return statusRank[to] >= statusRank[from]
}

// Market is one row of the markets table.
type Market struct {
	MatchID      int64
	MarketTypeID int32
	Specifier    string
	Status       MarketStatus
	GroupID      *int32
}

// Outcome is one row of the outcomes table.
type Outcome struct {
	MatchID      int64
	MarketTypeID int32
	Specifier    string
	OutcomeID    int32
	Odds         *float64
	IsActive     bool
}

// MarketStatusHistoryRow is one append into market_status_history.
type MarketStatusHistoryRow struct {
	MatchID      int64
	MarketTypeID int32
	Specifier    string
	From         MarketStatus // may be StatusUnknown for first transition
	To           MarketStatus
	RawMessageID [16]byte // zero -> NULL
}

// BetStopScope narrows which markets a bet_stop targets.
type BetStopScope struct {
	MatchID      int64
	MarketTypeID *int32 // nil = every market
	Specifier    string // empty + MarketTypeID nil = every market
	GroupID      *int32 // mutually exclusive with MarketTypeID
}

// Repo abstracts persistence. PgRepo (pgrepo.go) implements it for the
// production stack; unit tests use an in-memory fake.
type Repo interface {
	UpsertMarket(ctx context.Context, m Market) error
	GetMarket(ctx context.Context, matchID int64, marketTypeID int32, specifier string) (Market, bool, error)
	UpsertOutcome(ctx context.Context, o Outcome) error
	InsertMarketStatusHistory(ctx context.Context, row MarketStatusHistoryRow) error

	// MarketsForBetStop returns the markets matching scope (FOR UPDATE in
	// the pgx implementation so a single bet_stop snapshots a consistent
	// view).
	MarketsForBetStop(ctx context.Context, scope BetStopScope) ([]Market, error)

	// MatchExists is consulted before any insert to short-circuit
	// deliveries that arrive before the catalog handler has seen the match.
	MatchExists(ctx context.Context, matchID int64) (bool, error)
}

// AntiRegressionEvent is emitted whenever an incoming status would
// regress a persisted market.
type AntiRegressionEvent struct {
	MatchID      int64
	MarketTypeID int32
	Specifier    string
	From         MarketStatus
	To           MarketStatus
}

// Logger receives anti-regression notices. Pass nil to silently drop.
type Logger interface {
	AntiRegressionBlocked(ev AntiRegressionEvent)
}

// LoggerFunc adapts a plain function into a Logger.
type LoggerFunc func(ev AntiRegressionEvent)

// AntiRegressionBlocked implements Logger.
func (f LoggerFunc) AntiRegressionBlocked(ev AntiRegressionEvent) { f(ev) }

// Handler is the odds facade. Construct via New, then call Register
// to bind it to the feed dispatcher.
type Handler struct {
	Repo   Repo
	Logger Logger
	Now    func() time.Time

	regressionCount atomic.Int64
}

// New returns a Handler bound to repo.
func New(repo Repo) *Handler { return &Handler{Repo: repo} }

// RegressionCount returns the number of anti-regression events blocked
// since construction.
func (h *Handler) RegressionCount() int64 { return h.regressionCount.Load() }

// Register wires the odds handler into a feed dispatcher.
func (h *Handler) Register(d *feed.Dispatcher) {
	d.Register(feed.MsgOddsChange, feed.HandlerFunc(h.HandleOddsChange))
	d.Register(feed.MsgBetStop, feed.HandlerFunc(h.HandleBetStop))
}

func (h *Handler) now() time.Time {
	if h.Now != nil {
		return h.Now()
	}
	return time.Now()
}

func (h *Handler) emit(ev AntiRegressionEvent) {
	h.regressionCount.Add(1)
	if h.Logger != nil {
		h.Logger.AntiRegressionBlocked(ev)
	}
}
