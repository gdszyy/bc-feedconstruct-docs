package bets

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"
)

// Limits is the per-bet validation envelope. Zero MinStake / MaxStake
// disable the corresponding check.
type Limits struct {
	MinStake float64
	MaxStake float64
	// PriceTolerance is the absolute float64 difference at which a
	// price move is considered material (default 0.0001).
	PriceTolerance float64
}

// Logger observes Manager lifecycle.
type Logger interface {
	Validated(req ValidateRequest, resp ValidateResponse)
	Placed(b *Bet, deduped bool)
	Applied(betID string, ev EventKind, from, to State)
	Skipped(betID string, ev EventKind, from State, reason string)
}

// Manager is the bets domain entry point.
type Manager struct {
	Repo     Repo
	Outcomes OutcomeStateLookup
	IDs      IDGenerator
	Limits   Limits
	Logger   Logger
	Now      func() time.Time
}

// New returns a Manager wired with safe defaults.
func New(repo Repo, outcomes OutcomeStateLookup, ids IDGenerator) *Manager {
	return &Manager{
		Repo:     repo,
		Outcomes: outcomes,
		IDs:      ids,
		Limits:   Limits{PriceTolerance: 0.0001},
	}
}

func (m *Manager) now() time.Time {
	if m.Now != nil {
		return m.Now()
	}
	return time.Now().UTC()
}

func (m *Manager) tolerance() float64 {
	if m.Limits.PriceTolerance > 0 {
		return m.Limits.PriceTolerance
	}
	return 0.0001
}

// Validate runs the pre-place checks: every selection's outcome must
// be Active and the locked odds must match the current odds within
// PriceTolerance. Stake limits are enforced when set.
func (m *Manager) Validate(ctx context.Context, req ValidateRequest) (ValidateResponse, error) {
	resp := ValidateResponse{OK: true}

	if m.Limits.MaxStake > 0 && req.Stake > m.Limits.MaxStake {
		resp.OK = false
		resp.Code = CodeStakeOverLimit
		resp.Message = fmt.Sprintf("stake %.4f exceeds max %.4f", req.Stake, m.Limits.MaxStake)
		m.logValidate(req, resp)
		return resp, nil
	}
	if m.Limits.MinStake > 0 && req.Stake < m.Limits.MinStake {
		resp.OK = false
		resp.Code = CodeStakeUnderLimit
		resp.Message = fmt.Sprintf("stake %.4f below min %.4f", req.Stake, m.Limits.MinStake)
		m.logValidate(req, resp)
		return resp, nil
	}

	if m.Outcomes != nil {
		for _, sel := range req.Selections {
			view, ok, err := m.Outcomes.OutcomeState(ctx, sel.MatchID, sel.MarketID, sel.OutcomeID)
			if err != nil {
				return ValidateResponse{}, fmt.Errorf("bets: outcome lookup: %w", err)
			}
			if !ok || !view.MarketActive || !view.OutcomeActive {
				resp.OK = false
				resp.Code = CodeOutcomeUnavailable
				resp.Unavailable = append(resp.Unavailable, sel.OutcomeID)
				continue
			}
			if math.Abs(view.CurrentOdds-sel.LockedOdds) > m.tolerance() {
				resp.OK = false
				if resp.Code == "" {
					resp.Code = CodePriceChanged
				}
				resp.PriceChanges = append(resp.PriceChanges, PriceChange{
					OutcomeID: sel.OutcomeID,
					From:      sel.LockedOdds,
					To:        view.CurrentOdds,
				})
			}
		}
	}
	if !resp.OK && resp.Message == "" {
		resp.Message = humanReason(resp.Code)
	}
	m.logValidate(req, resp)
	return resp, nil
}

func humanReason(code string) string {
	switch code {
	case CodeOutcomeUnavailable:
		return "one or more selections are not currently bettable"
	case CodePriceChanged:
		return "odds moved since the slip was built"
	case CodeStakeOverLimit:
		return "stake exceeds the per-bet maximum"
	case CodeStakeUnderLimit:
		return "stake is below the per-bet minimum"
	}
	return ""
}

func (m *Manager) logValidate(req ValidateRequest, resp ValidateResponse) {
	if m.Logger != nil {
		m.Logger.Validated(req, resp)
	}
}

// Place persists a new Pending bet, or returns the original bet when
// the (UserID, IdempotencyKey) pair already exists.
func (m *Manager) Place(ctx context.Context, req PlaceRequest) (PlaceResponse, error) {
	if strings.TrimSpace(req.IdempotencyKey) == "" {
		return PlaceResponse{}, ErrIdempotencyKeyMissing
	}
	if strings.TrimSpace(req.UserID) == "" {
		return PlaceResponse{}, errors.New("bets: user id required")
	}
	if len(req.Selections) == 0 {
		return PlaceResponse{}, errors.New("bets: at least one selection required")
	}
	if m.Limits.MaxStake > 0 && req.Stake > m.Limits.MaxStake {
		return PlaceResponse{}, &PlaceError{Code: CodeStakeOverLimit, Message: humanReason(CodeStakeOverLimit)}
	}
	if m.Limits.MinStake > 0 && req.Stake < m.Limits.MinStake {
		return PlaceResponse{}, &PlaceError{Code: CodeStakeUnderLimit, Message: humanReason(CodeStakeUnderLimit)}
	}

	existing, ok, err := m.Repo.FindByIdempotencyKey(ctx, req.UserID, req.IdempotencyKey)
	if err != nil {
		return PlaceResponse{}, fmt.Errorf("bets: idempotency lookup: %w", err)
	}
	if ok {
		if m.Logger != nil {
			m.Logger.Placed(existing, true)
		}
		return PlaceResponse{BetID: existing.ID, State: existing.State, Deduped: true}, nil
	}

	now := m.now()
	bet := &Bet{
		ID:             m.IDs.NextID(),
		UserID:         req.UserID,
		PlacedAt:       now,
		Stake:          req.Stake,
		Currency:       strings.ToUpper(strings.TrimSpace(req.Currency)),
		BetType:        req.BetType,
		State:          StatePending,
		IdempotencyKey: req.IdempotencyKey,
		Selections:     normaliseSelections(req.Selections),
	}
	initial := Transition{
		At:     now,
		From:   "",
		To:     StatePending,
		Reason: "place",
	}
	if err := m.Repo.CreatePending(ctx, bet, initial); err != nil {
		return PlaceResponse{}, fmt.Errorf("bets: create pending: %w", err)
	}
	if m.Logger != nil {
		m.Logger.Placed(bet, false)
	}
	return PlaceResponse{BetID: bet.ID, State: bet.State}, nil
}

func normaliseSelections(in []Selection) []Selection {
	out := make([]Selection, len(in))
	copy(out, in)
	for i := range out {
		if out[i].Position == 0 {
			out[i].Position = i + 1
		}
	}
	return out
}

// Get returns a single bet with full history.
func (m *Manager) Get(ctx context.Context, betID string) (*Bet, bool, error) {
	return m.Repo.GetByID(ctx, betID)
}

// List returns the user's bets matching the filter.
func (m *Manager) List(ctx context.Context, f ListFilter) ([]*Bet, error) {
	return m.Repo.List(ctx, f)
}

// EventInput is the slice of an upstream WS envelope the bets manager
// needs to advance the FSM.
type EventInput struct {
	BetID         string
	Kind          EventKind
	EventID       string
	CorrelationID string
	OccurredAt    time.Time
	Reason        string

	// Optional payout fields, applied only on EventSettlementApplied
	// (or cleared on rollback).
	PayoutGross    *float64
	PayoutCurrency string
	VoidFactor     *float64
	DeadHeatFactor *float64
}

// ApplyEvent advances the FSM for one bet. Idempotent on (bet_id,
// event_id): a duplicate event_id returns (zero, false, nil) and
// records nothing.
//
// Returns:
//   - the new transition (with ID set) on success
//   - applied=false, err=nil when the FSM rejected the event (logged)
//   - applied=false, err=nil when the event was a duplicate (logged)
func (m *Manager) ApplyEvent(ctx context.Context, in EventInput) (Transition, bool, error) {
	bet, ok, err := m.Repo.GetByID(ctx, in.BetID)
	if err != nil {
		return Transition{}, false, fmt.Errorf("bets: get bet: %w", err)
	}
	if !ok {
		// Unknown bet — could be replay before our snapshot. Drop
		// silently; this matches M02's "events for unknown entities
		// are dropped" rule.
		return Transition{}, false, nil
	}

	// Duplicate-event check runs BEFORE the FSM gate so a replayed
	// event whose target state already matches current is reported
	// as "duplicate_event", not "fsm_reject". The repo's unique index
	// is the safety net; this in-memory pre-check just makes the
	// reason field semantically correct.
	if in.EventID != "" {
		for _, prior := range bet.Transitions {
			if prior.EventID == in.EventID {
				if m.Logger != nil {
					m.Logger.Skipped(bet.ID, in.Kind, bet.State, "duplicate_event")
				}
				return Transition{}, false, nil
			}
		}
	}

	next, valid := Apply(bet.State, in.Kind)
	if !valid {
		if m.Logger != nil {
			m.Logger.Skipped(bet.ID, in.Kind, bet.State, "fsm_reject")
		}
		return Transition{}, false, nil
	}

	occurred := in.OccurredAt
	if occurred.IsZero() {
		occurred = m.now()
	}
	t := Transition{
		At:            occurred,
		From:          bet.State,
		To:            next,
		Reason:        in.Reason,
		EventID:       in.EventID,
		CorrelationID: in.CorrelationID,
	}

	payout := payoutForEvent(in)
	written, fresh, err := m.Repo.AppendTransition(ctx, bet.ID, t, payout)
	if err != nil {
		return Transition{}, false, fmt.Errorf("bets: append transition: %w", err)
	}
	if !fresh {
		// Repo-level safety net (race window between our pre-check
		// and the insert). Treat as duplicate.
		if m.Logger != nil {
			m.Logger.Skipped(bet.ID, in.Kind, bet.State, "duplicate_event")
		}
		return Transition{}, false, nil
	}
	if m.Logger != nil {
		m.Logger.Applied(bet.ID, in.Kind, bet.State, next)
	}
	return written, true, nil
}

func payoutForEvent(in EventInput) *Payout {
	switch in.Kind {
	case EventSettlementApplied:
		return &Payout{
			Gross:          in.PayoutGross,
			Currency:       in.PayoutCurrency,
			VoidFactor:     in.VoidFactor,
			DeadHeatFactor: in.DeadHeatFactor,
		}
	case EventSettlementRolledBack:
		return &Payout{ClearPayout: true}
	}
	return nil
}

// PlaceError is the typed rejection a Place call returns when the
// validation fails before the bet is persisted. The HTTP layer
// translates Code into the documented error envelope.
type PlaceError struct {
	Code    string
	Message string
}

// Error implements error.
func (e *PlaceError) Error() string { return e.Code + ": " + e.Message }
