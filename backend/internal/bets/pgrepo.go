package bets

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PgRepo is the production pgxpool-backed Repo implementation.
type PgRepo struct{ pool *pgxpool.Pool }

// NewPgRepo returns a PgRepo bound to pool.
func NewPgRepo(pool *pgxpool.Pool) *PgRepo { return &PgRepo{pool: pool} }

// FindByIdempotencyKey implements Repo.
func (r *PgRepo) FindByIdempotencyKey(ctx context.Context, userID, key string) (*Bet, bool, error) {
	const q = `SELECT id FROM bets WHERE user_id = $1 AND idempotency_key = $2`
	var id string
	if err := r.pool.QueryRow(ctx, q, userID, key).Scan(&id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("storage: idempotency lookup: %w", err)
	}
	return r.GetByID(ctx, id)
}

// CreatePending implements Repo.
func (r *PgRepo) CreatePending(ctx context.Context, b *Bet, initial Transition) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("storage: begin: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	const insBet = `
		INSERT INTO bets (id, user_id, placed_at, stake, currency, bet_type, state, idempotency_key)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	if _, err := tx.Exec(ctx, insBet,
		b.ID, b.UserID, b.PlacedAt, b.Stake, b.Currency, string(b.BetType), string(b.State), b.IdempotencyKey,
	); err != nil {
		return fmt.Errorf("storage: insert bet: %w", err)
	}

	const insSel = `
		INSERT INTO bet_selections (bet_id, position, match_id, market_id, outcome_id, locked_odds)
		VALUES ($1, $2, $3, $4, $5, $6)`
	for _, s := range b.Selections {
		if _, err := tx.Exec(ctx, insSel,
			b.ID, s.Position, s.MatchID, s.MarketID, s.OutcomeID, s.LockedOdds,
		); err != nil {
			return fmt.Errorf("storage: insert selection %d: %w", s.Position, err)
		}
	}

	const insTrans = `
		INSERT INTO bet_transitions (bet_id, at, from_state, to_state, reason, event_id, correlation_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id`
	if err := tx.QueryRow(ctx, insTrans,
		b.ID, initial.At, string(initial.From), string(initial.To), initial.Reason, initial.EventID, initial.CorrelationID,
	).Scan(&initial.ID); err != nil {
		return fmt.Errorf("storage: insert initial transition: %w", err)
	}
	b.Transitions = []Transition{initial}

	return tx.Commit(ctx)
}

// GetByID implements Repo.
func (r *PgRepo) GetByID(ctx context.Context, betID string) (*Bet, bool, error) {
	const qBet = `
		SELECT id, user_id, placed_at, stake, currency, bet_type, state, idempotency_key,
		       payout_gross, payout_currency, void_factor, dead_heat_factor
		  FROM bets WHERE id = $1`
	var (
		bet         Bet
		payoutGross *float64
		payoutCcy   *string
		voidFactor  *float64
		dhFactor    *float64
		state       string
		bType       string
	)
	err := r.pool.QueryRow(ctx, qBet, betID).Scan(
		&bet.ID, &bet.UserID, &bet.PlacedAt, &bet.Stake, &bet.Currency, &bType, &state, &bet.IdempotencyKey,
		&payoutGross, &payoutCcy, &voidFactor, &dhFactor,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("storage: get bet: %w", err)
	}
	bet.State = State(state)
	bet.BetType = BetType(bType)
	bet.PayoutGross = payoutGross
	if payoutCcy != nil {
		bet.PayoutCurrency = *payoutCcy
	}
	bet.VoidFactor = voidFactor
	bet.DeadHeatFactor = dhFactor

	const qSel = `
		SELECT position, match_id, market_id, outcome_id, locked_odds
		  FROM bet_selections WHERE bet_id = $1 ORDER BY position`
	rows, err := r.pool.Query(ctx, qSel, betID)
	if err != nil {
		return nil, false, fmt.Errorf("storage: list selections: %w", err)
	}
	for rows.Next() {
		var s Selection
		if err := rows.Scan(&s.Position, &s.MatchID, &s.MarketID, &s.OutcomeID, &s.LockedOdds); err != nil {
			rows.Close()
			return nil, false, fmt.Errorf("storage: scan selection: %w", err)
		}
		bet.Selections = append(bet.Selections, s)
	}
	rows.Close()

	const qTrans = `
		SELECT id, at, from_state, to_state, reason, event_id, correlation_id
		  FROM bet_transitions WHERE bet_id = $1 ORDER BY at, id`
	tRows, err := r.pool.Query(ctx, qTrans, betID)
	if err != nil {
		return nil, false, fmt.Errorf("storage: list transitions: %w", err)
	}
	defer tRows.Close()
	for tRows.Next() {
		var (
			t       Transition
			fromStr string
			toStr   string
		)
		if err := tRows.Scan(&t.ID, &t.At, &fromStr, &toStr, &t.Reason, &t.EventID, &t.CorrelationID); err != nil {
			return nil, false, fmt.Errorf("storage: scan transition: %w", err)
		}
		t.From = State(fromStr)
		t.To = State(toStr)
		bet.Transitions = append(bet.Transitions, t)
	}
	return &bet, true, tRows.Err()
}

// List implements Repo.
func (r *PgRepo) List(ctx context.Context, f ListFilter) ([]*Bet, error) {
	limit := f.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}

	q := `SELECT id FROM bets WHERE user_id = $1`
	args := []any{f.UserID}
	if len(f.States) > 0 {
		states := make([]string, len(f.States))
		for i, s := range f.States {
			states[i] = string(s)
		}
		q += ` AND state = ANY($2)`
		args = append(args, states)
	}
	q += ` ORDER BY placed_at DESC LIMIT ` + fmt.Sprintf("%d", limit)

	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("storage: list bets: %w", err)
	}
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return nil, fmt.Errorf("storage: scan bet id: %w", err)
		}
		ids = append(ids, id)
	}
	rows.Close()

	out := make([]*Bet, 0, len(ids))
	for _, id := range ids {
		b, ok, err := r.GetByID(ctx, id)
		if err != nil {
			return nil, err
		}
		if ok {
			out = append(out, b)
		}
	}
	return out, nil
}

// AppendTransition implements Repo. The unique index on
// (bet_id, event_id WHERE event_id<>”) turns duplicate events into
// a no-op via ON CONFLICT DO NOTHING.
func (r *PgRepo) AppendTransition(ctx context.Context, betID string, t Transition, p *Payout) (Transition, bool, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return Transition{}, false, fmt.Errorf("storage: begin: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	const ins = `
		INSERT INTO bet_transitions (bet_id, at, from_state, to_state, reason, event_id, correlation_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (bet_id, event_id) WHERE event_id <> ''
		DO NOTHING
		RETURNING id`
	var id int64
	err = tx.QueryRow(ctx, ins,
		betID, t.At, string(t.From), string(t.To), t.Reason, t.EventID, t.CorrelationID,
	).Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Duplicate event_id — caller treats this as "already applied".
			_ = tx.Rollback(ctx)
			return Transition{}, false, nil
		}
		return Transition{}, false, fmt.Errorf("storage: append transition: %w", err)
	}
	t.ID = id

	const updBet = `UPDATE bets SET state = $2, updated_at = now() WHERE id = $1`
	if _, err := tx.Exec(ctx, updBet, betID, string(t.To)); err != nil {
		return Transition{}, false, fmt.Errorf("storage: update bet state: %w", err)
	}
	if p != nil {
		if err := updatePayout(ctx, tx, betID, p); err != nil {
			return Transition{}, false, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return Transition{}, false, fmt.Errorf("storage: commit: %w", err)
	}
	return t, true, nil
}

func updatePayout(ctx context.Context, tx pgx.Tx, betID string, p *Payout) error {
	if p.ClearPayout {
		const q = `
			UPDATE bets
			   SET payout_gross = NULL, payout_currency = NULL,
			       void_factor = NULL, dead_heat_factor = NULL
			 WHERE id = $1`
		if _, err := tx.Exec(ctx, q, betID); err != nil {
			return fmt.Errorf("storage: clear payout: %w", err)
		}
		return nil
	}
	const q = `
		UPDATE bets
		   SET payout_gross     = COALESCE($2, payout_gross),
		       payout_currency  = COALESCE(NULLIF($3, ''), payout_currency),
		       void_factor      = COALESCE($4, void_factor),
		       dead_heat_factor = COALESCE($5, dead_heat_factor)
		 WHERE id = $1`
	if _, err := tx.Exec(ctx, q, betID, p.Gross, p.Currency, p.VoidFactor, p.DeadHeatFactor); err != nil {
		return fmt.Errorf("storage: update payout: %w", err)
	}
	return nil
}

// RandomIDGenerator returns a new IDGenerator that produces 26-char
// uppercase hex IDs prefixed with the wall-clock time so IDs sort
// chronologically. ULID-style without a third-party dep.
func NewRandomIDGenerator() IDGenerator { return &randomID{} }

type randomID struct{}

// NextID implements IDGenerator. The format is
// <8 hex char timestamp>-<16 hex char random>; collisions over the
// 8-byte random tail are vanishingly rare.
func (r *randomID) NextID() string {
	now := uint64(time.Now().UnixMilli())
	var buf [8]byte
	_, _ = rand.Read(buf[:])
	return fmt.Sprintf("%012x-%s", now, hex.EncodeToString(buf[:]))
}
