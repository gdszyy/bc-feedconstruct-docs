package settlement

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PgRepo persists settlements / cancels / rollbacks via pgx and drives
// terminal status transitions on the markets table (plus the
// market_status_history append). Satisfies Repo.
type PgRepo struct{ pool *pgxpool.Pool }

// NewPgRepo returns a PgRepo bound to pool.
func NewPgRepo(pool *pgxpool.Pool) *PgRepo { return &PgRepo{pool: pool} }

func rawIDArg(id [16]byte) any {
	var zero [16]byte
	if id == zero {
		return nil
	}
	return id
}

func (r *PgRepo) MatchExists(ctx context.Context, matchID int64) (bool, error) {
	var ok bool
	if err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM matches WHERE id = $1)`, matchID,
	).Scan(&ok); err != nil {
		return false, fmt.Errorf("storage: match exists: %w", err)
	}
	return ok, nil
}

func (r *PgRepo) InsertSettlement(ctx context.Context, s Settlement) (int64, error) {
	const q = `
		INSERT INTO settlements (
			match_id, market_type_id, specifier, outcome_id,
			result, certainty, void_factor, dead_heat_factor,
			raw_message_id, settled_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		ON CONFLICT (match_id, market_type_id, specifier, outcome_id, settled_at)
		DO UPDATE SET
			result            = EXCLUDED.result,
			certainty         = GREATEST(settlements.certainty, EXCLUDED.certainty),
			void_factor       = COALESCE(EXCLUDED.void_factor, settlements.void_factor),
			dead_heat_factor  = COALESCE(EXCLUDED.dead_heat_factor, settlements.dead_heat_factor)
		RETURNING id`
	var id int64
	err := r.pool.QueryRow(ctx, q,
		s.MatchID, s.MarketTypeID, s.Specifier, s.OutcomeID,
		string(s.Result), s.Certainty, s.VoidFactor, s.DeadHeatFactor,
		rawIDArg(s.RawMessageID), s.SettledAt,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("storage: insert settlement: %w", err)
	}
	return id, nil
}

func (r *PgRepo) LatestSettlementForOutcome(ctx context.Context, matchID int64, marketTypeID int32, specifier string, outcomeID int32) (Settlement, bool, error) {
	const q = `
		SELECT id, match_id, market_type_id, specifier, outcome_id,
		       result, certainty, void_factor, dead_heat_factor,
		       raw_message_id, settled_at, rolled_back_at
		  FROM settlements
		 WHERE match_id = $1 AND market_type_id = $2 AND specifier = $3 AND outcome_id = $4
		 ORDER BY settled_at DESC, id DESC
		 LIMIT 1`
	var (
		s        Settlement
		resStr   string
		rawID    *[16]byte
		rolledBy *time.Time
	)
	err := r.pool.QueryRow(ctx, q, matchID, marketTypeID, specifier, outcomeID).Scan(
		&s.ID, &s.MatchID, &s.MarketTypeID, &s.Specifier, &s.OutcomeID,
		&resStr, &s.Certainty, &s.VoidFactor, &s.DeadHeatFactor,
		&rawID, &s.SettledAt, &rolledBy,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Settlement{}, false, nil
		}
		return Settlement{}, false, fmt.Errorf("storage: latest settlement: %w", err)
	}
	s.Result = Result(resStr)
	if rawID != nil {
		s.RawMessageID = *rawID
	}
	s.RolledBackAt = rolledBy
	return s, true, nil
}

func (r *PgRepo) MarkSettlementRolledBack(ctx context.Context, id int64, at time.Time) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE settlements SET rolled_back_at = $2 WHERE id = $1`, id, at)
	if err != nil {
		return fmt.Errorf("storage: mark settlement rolled_back: %w", err)
	}
	return nil
}

func (r *PgRepo) InsertCancel(ctx context.Context, c Cancel) (int64, error) {
	const q = `
		INSERT INTO cancels (
			match_id, market_type_id, specifier,
			void_reason, void_action, superceded_by,
			from_ts, to_ts, raw_message_id, cancelled_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		RETURNING id`
	var id int64
	err := r.pool.QueryRow(ctx, q,
		c.MatchID, c.MarketTypeID, c.Specifier,
		c.VoidReason, c.VoidAction, c.SupercededBy,
		c.FromTS, c.ToTS, rawIDArg(c.RawMessageID), c.CancelledAt,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("storage: insert cancel: %w", err)
	}
	return id, nil
}

func (r *PgRepo) LatestCancelForScope(ctx context.Context, matchID int64, marketTypeID *int32, specifier string) (Cancel, bool, error) {
	var (
		q    string
		args []any
	)
	if marketTypeID == nil {
		q = `
			SELECT id, match_id, market_type_id, specifier,
			       void_reason, void_action, superceded_by,
			       from_ts, to_ts, raw_message_id, cancelled_at, rolled_back_at
			  FROM cancels
			 WHERE match_id = $1 AND market_type_id IS NULL
			 ORDER BY cancelled_at DESC, id DESC LIMIT 1`
		args = []any{matchID}
	} else if specifier == "" {
		q = `
			SELECT id, match_id, market_type_id, specifier,
			       void_reason, void_action, superceded_by,
			       from_ts, to_ts, raw_message_id, cancelled_at, rolled_back_at
			  FROM cancels
			 WHERE match_id = $1 AND market_type_id = $2
			 ORDER BY cancelled_at DESC, id DESC LIMIT 1`
		args = []any{matchID, *marketTypeID}
	} else {
		q = `
			SELECT id, match_id, market_type_id, specifier,
			       void_reason, void_action, superceded_by,
			       from_ts, to_ts, raw_message_id, cancelled_at, rolled_back_at
			  FROM cancels
			 WHERE match_id = $1 AND market_type_id = $2 AND specifier = $3
			 ORDER BY cancelled_at DESC, id DESC LIMIT 1`
		args = []any{matchID, *marketTypeID, specifier}
	}
	var (
		c        Cancel
		rawID    *[16]byte
		rolledBy *time.Time
	)
	err := r.pool.QueryRow(ctx, q, args...).Scan(
		&c.ID, &c.MatchID, &c.MarketTypeID, &c.Specifier,
		&c.VoidReason, &c.VoidAction, &c.SupercededBy,
		&c.FromTS, &c.ToTS, &rawID, &c.CancelledAt, &rolledBy,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Cancel{}, false, nil
		}
		return Cancel{}, false, fmt.Errorf("storage: latest cancel: %w", err)
	}
	if rawID != nil {
		c.RawMessageID = *rawID
	}
	c.RolledBackAt = rolledBy
	return c, true, nil
}

func (r *PgRepo) MarkCancelRolledBack(ctx context.Context, id int64, at time.Time) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE cancels SET rolled_back_at = $2 WHERE id = $1`, id, at)
	if err != nil {
		return fmt.Errorf("storage: mark cancel rolled_back: %w", err)
	}
	return nil
}

func (r *PgRepo) HasRollback(ctx context.Context, target RollbackTarget, targetID int64, rawID [16]byte) (bool, error) {
	var zero [16]byte
	if rawID == zero {
		// Without a stable rawID we cannot dedup; rely on the unique
		// constraint on insert to surface duplicates. Report "not seen".
		return false, nil
	}
	var ok bool
	if err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM rollbacks WHERE target=$1 AND target_id=$2 AND raw_message_id=$3)`,
		string(target), targetID, rawID,
	).Scan(&ok); err != nil {
		return false, fmt.Errorf("storage: has rollback: %w", err)
	}
	return ok, nil
}

func (r *PgRepo) InsertRollback(ctx context.Context, rb Rollback) (int64, error) {
	const q = `
		INSERT INTO rollbacks (target, target_id, raw_message_id, applied_at)
		VALUES ($1,$2,$3,$4)
		ON CONFLICT (target, target_id, raw_message_id) DO UPDATE SET applied_at = rollbacks.applied_at
		RETURNING id`
	var id int64
	err := r.pool.QueryRow(ctx, q,
		string(rb.Target), rb.TargetID, rawIDArg(rb.RawMessageID), rb.AppliedAt,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("storage: insert rollback: %w", err)
	}
	return id, nil
}

func (r *PgRepo) GetMarket(ctx context.Context, matchID int64, marketTypeID int32, specifier string) (MarketRef, bool, error) {
	const q = `
		SELECT match_id, market_type_id, specifier, status
		  FROM markets
		 WHERE match_id = $1 AND market_type_id = $2 AND specifier = $3`
	var (
		m      MarketRef
		status string
	)
	err := r.pool.QueryRow(ctx, q, matchID, marketTypeID, specifier).Scan(
		&m.MatchID, &m.MarketTypeID, &m.Specifier, &status,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return MarketRef{}, false, nil
		}
		return MarketRef{}, false, fmt.Errorf("storage: get market: %w", err)
	}
	m.Status = MarketStatus(status)
	return m, true, nil
}

func (r *PgRepo) ListMarketsForMatch(ctx context.Context, matchID int64) ([]MarketRef, error) {
	const q = `
		SELECT match_id, market_type_id, specifier, status
		  FROM markets
		 WHERE match_id = $1
		 ORDER BY market_type_id, specifier`
	rows, err := r.pool.Query(ctx, q, matchID)
	if err != nil {
		return nil, fmt.Errorf("storage: list markets: %w", err)
	}
	defer rows.Close()
	var out []MarketRef
	for rows.Next() {
		var (
			m      MarketRef
			status string
		)
		if err := rows.Scan(&m.MatchID, &m.MarketTypeID, &m.Specifier, &status); err != nil {
			return nil, fmt.Errorf("storage: scan markets: %w", err)
		}
		m.Status = MarketStatus(status)
		out = append(out, m)
	}
	return out, rows.Err()
}

// SetMarketStatus performs the markets update plus a history append in
// one transaction so the prior status is captured atomically.
func (r *PgRepo) SetMarketStatus(ctx context.Context, matchID int64, marketTypeID int32, specifier string, to MarketStatus, rawID [16]byte) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("storage: begin set market status: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var fromStatus string
	err = tx.QueryRow(ctx,
		`SELECT status FROM markets WHERE match_id=$1 AND market_type_id=$2 AND specifier=$3 FOR UPDATE`,
		matchID, marketTypeID, specifier,
	).Scan(&fromStatus)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		fromStatus = ""
	case err != nil:
		return fmt.Errorf("storage: read market for transition: %w", err)
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO markets (match_id, market_type_id, specifier, status, updated_at)
		VALUES ($1,$2,$3,$4, now())
		ON CONFLICT (match_id, market_type_id, specifier) DO UPDATE
		   SET status = EXCLUDED.status, updated_at = now()`,
		matchID, marketTypeID, specifier, string(to),
	); err != nil {
		return fmt.Errorf("storage: upsert market status: %w", err)
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO market_status_history
		    (match_id, market_type_id, specifier, from_status, to_status, raw_message_id, changed_at)
		VALUES ($1,$2,$3,$4,$5,$6, now())`,
		matchID, marketTypeID, specifier,
		nullableStatus(fromStatus), string(to), rawIDArg(rawID),
	); err != nil {
		return fmt.Errorf("storage: append market_status_history: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("storage: commit set market status: %w", err)
	}
	return nil
}

// RevertMarketStatus rolls a market back to the most recent
// operational status (active / suspended / deactivated) recorded in
// market_status_history. Terminal statuses (settled / cancelled /
// handed_over) and the current row itself are excluded so the lookup
// returns the state that preceded the most recent terminal transition.
func (r *PgRepo) RevertMarketStatus(ctx context.Context, matchID int64, marketTypeID int32, specifier string, rawID [16]byte) (MarketStatus, bool, error) {
	const q = `
		SELECT from_status
		  FROM market_status_history
		 WHERE match_id = $1 AND market_type_id = $2 AND specifier = $3
		   AND from_status IS NOT NULL
		   AND from_status NOT IN ('settled','cancelled','handed_over')
		 ORDER BY changed_at DESC, id DESC
		 LIMIT 1`
	var prior string
	err := r.pool.QueryRow(ctx, q, matchID, marketTypeID, specifier).Scan(&prior)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return StatusUnknown, false, nil
		}
		return StatusUnknown, false, fmt.Errorf("storage: revert lookup: %w", err)
	}
	target := MarketStatus(prior)
	if err := r.SetMarketStatus(ctx, matchID, marketTypeID, specifier, target, rawID); err != nil {
		return StatusUnknown, false, err
	}
	return target, true, nil
}

func nullableStatus(s string) any {
	if s == "" {
		return nil
	}
	return s
}
