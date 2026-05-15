// Package bets is the backend half of frontend modules M13 (Bet Slip)
// and M14 (My Bets). It owns:
//
//   - the Bet aggregate (id, selections, stake, currency, FSM state)
//   - the append-only BetTransition history that drives M14's timeline
//   - validate / place / list / get-by-id REST endpoints exposed through
//     internal/webapi
//   - the idempotency-key dedupe enforced at place() time
//   - the wiring from settlement/cancel/rollback events back into the
//     Bet FSM (Pending → Accepted → Settled → … → RolledBack)
//
// The package is purposely event-sourced: every state change inserts a
// new bet_transitions row, never UPDATEs an existing one. M14's
// timeline rendering depends on this.
package bets

import (
	"context"
	"errors"
	"time"
)

// State is the Bet FSM state. The string values are persisted, so any
// rename is a migration.
type State string

// Bet states. See docs/07_frontend_architecture/04_state_machines.md §4.
const (
	StatePending   State = "pending"
	StateAccepted  State = "accepted"
	StateRejected  State = "rejected"
	StateSettled   State = "settled"
	StateCancelled State = "cancelled"
)

// BetType is the slip composition.
type BetType string

// Recognised bet types per M13.
const (
	BetTypeSingle BetType = "single"
	BetTypeCombo  BetType = "combo"
	BetTypeSystem BetType = "system"
)

// Selection is one outcome in a Bet.
type Selection struct {
	Position   int
	MatchID    string
	MarketID   string
	OutcomeID  string
	LockedOdds float64
}

// Bet is the top-level aggregate.
type Bet struct {
	ID             string
	UserID         string
	PlacedAt       time.Time
	Stake          float64
	Currency       string
	BetType        BetType
	State          State
	IdempotencyKey string
	PayoutGross    *float64
	PayoutCurrency string
	VoidFactor     *float64
	DeadHeatFactor *float64
	Selections     []Selection
	Transitions    []Transition
}

// Transition records a single FSM step. event_id may be empty for
// transitions synthesised by the BFF itself (e.g. the initial Pending
// row inserted at place() time).
type Transition struct {
	ID            int64
	At            time.Time
	From          State
	To            State
	Reason        string
	EventID       string
	CorrelationID string
}

// PriceChange is one outcome's locked-vs-current odds delta. The
// validate endpoint surfaces this so M13's NeedsReview state can
// render "old → new".
type PriceChange struct {
	OutcomeID string  `json:"outcome_id"`
	From      float64 `json:"from"`
	To        float64 `json:"to"`
}

// ValidateRequest mirrors POST /api/v1/bet-slip/validate.
type ValidateRequest struct {
	Selections []Selection
	Stake      float64
	Currency   string
	BetType    BetType
}

// ValidateResponse mirrors the 200 body of /bet-slip/validate. Code
// is empty when OK is true.
type ValidateResponse struct {
	OK           bool          `json:"ok"`
	Code         string        `json:"code,omitempty"`
	Message      string        `json:"message,omitempty"`
	PriceChanges []PriceChange `json:"price_changes,omitempty"`
	Unavailable  []string      `json:"unavailable_outcomes,omitempty"`
}

// Validate response codes. The frontend M13 routes on these strings.
const (
	CodeOutcomeUnavailable = "OUTCOME_UNAVAILABLE"
	CodePriceChanged       = "PRICE_CHANGED"
	CodeStakeOverLimit     = "STAKE_EXCEEDS_LIMIT"
	CodeStakeUnderLimit    = "STAKE_BELOW_MIN"
	CodeIdempotencyMissing = "IDEMPOTENCY_KEY_REQUIRED"
)

// PlaceRequest mirrors POST /api/v1/bet-slip/place.
type PlaceRequest struct {
	UserID         string
	IdempotencyKey string
	Selections     []Selection
	Stake          float64
	Currency       string
	BetType        BetType
}

// PlaceResponse mirrors the 200/201 body of /bet-slip/place.
type PlaceResponse struct {
	BetID   string `json:"bet_id"`
	State   State  `json:"state"`
	Deduped bool   `json:"deduped,omitempty"`
}

// ListFilter is the slice of query parameters the manager understands
// for GET /api/v1/my-bets.
type ListFilter struct {
	UserID string
	States []State
	Limit  int
}

// Repo abstracts persistence for testing.
type Repo interface {
	// FindByIdempotencyKey returns an existing bet for the given user
	// and key, or (nil, false, nil) when no match exists.
	FindByIdempotencyKey(ctx context.Context, userID, key string) (*Bet, bool, error)

	// CreatePending inserts the Bet, its Selections, and the initial
	// Pending transition in one transaction. The Bet.ID and
	// Transitions[0].ID are populated on success.
	CreatePending(ctx context.Context, b *Bet, initial Transition) error

	// GetByID returns a Bet with selections and transitions populated.
	GetByID(ctx context.Context, betID string) (*Bet, bool, error)

	// List returns bets matching f ordered by placed_at desc.
	List(ctx context.Context, f ListFilter) ([]*Bet, error)

	// AppendTransition inserts a transition row and updates the live
	// state on the bets row in one transaction. The unique index on
	// (bet_id, event_id WHERE event_id<>'') makes replays a no-op:
	// AppendTransition returns (nil, false) without erroring when the
	// event was already applied.
	AppendTransition(ctx context.Context, betID string, t Transition, payout *Payout) (Transition, bool, error)
}

// Payout carries the post-settlement money fields written to the bets
// row. Pointer fields are nil-aware: nil means "leave untouched"
// (e.g. on Cancel transitions we don't blank a previous Settled
// payout — the rolled_back path resets payout to nil explicitly).
type Payout struct {
	Gross          *float64
	Currency       string
	VoidFactor     *float64
	DeadHeatFactor *float64
	ClearPayout    bool
}

// OutcomeStateLookup is the slice of M05/M06 the bets manager needs at
// validate() time. Pass any concrete type that satisfies it (the odds
// package's Repo can be wrapped trivially).
type OutcomeStateLookup interface {
	OutcomeState(ctx context.Context, matchID, marketID, outcomeID string) (OutcomeView, bool, error)
}

// OutcomeView is the minimum slice of M05/M06 state Validate consults.
type OutcomeView struct {
	MarketActive  bool
	OutcomeActive bool
	CurrentOdds   float64
}

// IDGenerator returns a fresh, unique Bet ID. Construct one via
// NewULIDGenerator in production; tests inject a deterministic counter.
type IDGenerator interface {
	NextID() string
}

// Common errors.
var (
	ErrIdempotencyKeyMissing = errors.New("bets: idempotency key required")
	ErrUnknownEvent          = errors.New("bets: unrecognised event type for FSM")
)
