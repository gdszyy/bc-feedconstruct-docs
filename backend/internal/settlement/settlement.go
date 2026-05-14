// Package settlement processes bet_settlement, bet_cancel and rollback
// messages. Maps to upload-guideline 业务域 "结算状态" + "取消状态" + "回滚纠错"
// (frontend modules M08 and M09).
//
// The package owns three tables: settlements, cancels, rollbacks (see
// migrations/004_settlement.sql). It also drives terminal status
// transitions on the markets table (status=settled / cancelled) and
// reverses them when a rollback arrives, restoring the prior operational
// status recorded by the odds handler.
package settlement

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/gdszyy/bc-feedconstruct-docs/backend/internal/feed"
)

// MarketStatus mirrors the values enforced by the CHECK constraint on
// markets.status (see migrations/003_markets.sql). Kept as plain strings
// so we can interoperate with the odds package without importing it.
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

// Result mirrors the CHECK constraint on settlements.result.
type Result string

const (
	ResultWin      Result = "win"
	ResultLose     Result = "lose"
	ResultVoid     Result = "void"
	ResultHalfWin  Result = "half_win"
	ResultHalfLose Result = "half_lose"
)

// VoidAction values from FeedConstruct's VoidNotification.
const (
	VoidActionVoid   int16 = 1
	VoidActionUnvoid int16 = 2
)

// CertaintyConfirmed is the default certainty for bet_settlement frames;
// FC sends certainty=0 when the outcome is live-scouted and certainty=1
// when confirmed by the official source.
const (
	CertaintyUncertain int16 = 0
	CertaintyConfirmed int16 = 1
)

// RollbackTarget enumerates the rolled-back object kinds the rollbacks
// table can carry. Mirrors CHECK (target IN ('settlement','cancel')).
type RollbackTarget string

const (
	TargetSettlement RollbackTarget = "settlement"
	TargetCancel     RollbackTarget = "cancel"
)

// Settlement is one row of the settlements table.
type Settlement struct {
	ID             int64
	MatchID        int64
	MarketTypeID   int32
	Specifier      string
	OutcomeID      int32
	Result         Result
	Certainty      int16
	VoidFactor     *float64
	DeadHeatFactor *float64
	RawMessageID   [16]byte
	SettledAt      time.Time
	RolledBackAt   *time.Time
}

// Cancel is one row of the cancels table. MarketTypeID is nil for
// match-level cancels (ObjectType=4 in FC).
type Cancel struct {
	ID            int64
	MatchID       int64
	MarketTypeID  *int32
	Specifier     string
	VoidReason    *string
	VoidAction    int16
	SupercededBy  *int64
	FromTS        *time.Time
	ToTS          *time.Time
	RawMessageID  [16]byte
	CancelledAt   time.Time
	RolledBackAt  *time.Time
}

// Rollback is one row of the rollbacks table.
type Rollback struct {
	ID           int64
	Target       RollbackTarget
	TargetID     int64
	RawMessageID [16]byte
	AppliedAt    time.Time
}

// MarketRef identifies one markets row plus its current status. Returned
// by ListMarketsForMatch (used by match-level cancels) and GetMarket.
type MarketRef struct {
	MatchID      int64
	MarketTypeID int32
	Specifier    string
	Status       MarketStatus
}

// Repo abstracts persistence. PgRepo (pgrepo.go) implements it for the
// production stack; unit tests use an in-memory fake.
type Repo interface {
	// Match precondition guard. Mirrors odds.Repo.MatchExists so a
	// bet_settlement arriving before catalog has seen the match short
	// circuits rather than FK-failing.
	MatchExists(ctx context.Context, matchID int64) (bool, error)

	// Settlements.
	InsertSettlement(ctx context.Context, s Settlement) (int64, error)
	LatestSettlementForOutcome(ctx context.Context, matchID int64, marketTypeID int32, specifier string, outcomeID int32) (Settlement, bool, error)
	MarkSettlementRolledBack(ctx context.Context, settlementID int64, at time.Time) error

	// Cancels.
	InsertCancel(ctx context.Context, c Cancel) (int64, error)
	LatestCancelForScope(ctx context.Context, matchID int64, marketTypeID *int32, specifier string) (Cancel, bool, error)
	MarkCancelRolledBack(ctx context.Context, cancelID int64, at time.Time) error

	// Rollbacks. HasRollback enforces the unique index
	// (target, target_id, raw_message_id) so duplicate deliveries collapse.
	HasRollback(ctx context.Context, target RollbackTarget, targetID int64, rawID [16]byte) (bool, error)
	InsertRollback(ctx context.Context, r Rollback) (int64, error)

	// Market state transitions. The settlement handler does not append
	// to market_status_history directly; the Repo implementation is free
	// to do so (PgRepo does, the fake just tracks current+prior status).
	GetMarket(ctx context.Context, matchID int64, marketTypeID int32, specifier string) (MarketRef, bool, error)
	ListMarketsForMatch(ctx context.Context, matchID int64) ([]MarketRef, error)
	SetMarketStatus(ctx context.Context, matchID int64, marketTypeID int32, specifier string, to MarketStatus, rawID [16]byte) error
	// RevertMarketStatus restores the most recent non-terminal status
	// recorded for the market (used when a settlement or cancel is
	// rolled back). Returns (StatusUnknown, false) when no prior state
	// exists.
	RevertMarketStatus(ctx context.Context, matchID int64, marketTypeID int32, specifier string, rawID [16]byte) (MarketStatus, bool, error)
}

// Logger observes anti-regression and skip events. Pass nil to drop.
type Logger interface {
	SkipUnknownMatch(matchID int64, kind string)
}

// LoggerFunc adapts a function into a Logger.
type LoggerFunc func(matchID int64, kind string)

// SkipUnknownMatch implements Logger.
func (f LoggerFunc) SkipUnknownMatch(matchID int64, kind string) { f(matchID, kind) }

// Handler is the settlement facade. Construct via New, then call Register
// to bind it to the feed dispatcher.
type Handler struct {
	Repo   Repo
	Logger Logger
	// Now is injectable so tests can produce strictly increasing
	// settled_at values (acceptance 7 — uncertain then certain).
	Now func() time.Time

	settlementCount atomic.Int64
	cancelCount     atomic.Int64
	rollbackCount   atomic.Int64
	duplicateCount  atomic.Int64
}

// New returns a Handler bound to repo.
func New(repo Repo) *Handler { return &Handler{Repo: repo} }

// SettlementCount returns the number of settlements persisted.
func (h *Handler) SettlementCount() int64 { return h.settlementCount.Load() }

// CancelCount returns the number of cancels persisted.
func (h *Handler) CancelCount() int64 { return h.cancelCount.Load() }

// RollbackCount returns the number of rollbacks persisted.
func (h *Handler) RollbackCount() int64 { return h.rollbackCount.Load() }

// DuplicateRollbacks returns how many rollback deliveries were collapsed
// by the idempotency guard.
func (h *Handler) DuplicateRollbacks() int64 { return h.duplicateCount.Load() }

// Register wires the settlement handlers into a feed dispatcher.
func (h *Handler) Register(d *feed.Dispatcher) {
	d.Register(feed.MsgBetSettlement, feed.HandlerFunc(h.HandleBetSettlement))
	d.Register(feed.MsgBetCancel, feed.HandlerFunc(h.HandleBetCancel))
	d.Register(feed.MsgRollback, feed.HandlerFunc(h.HandleRollback))
	d.Register(feed.MsgRollbackCancel, feed.HandlerFunc(h.HandleRollback))
}

func (h *Handler) now() time.Time {
	if h.Now != nil {
		return h.Now()
	}
	return time.Now()
}
