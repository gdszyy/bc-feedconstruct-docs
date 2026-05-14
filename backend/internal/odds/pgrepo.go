package odds

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PgRepo persists markets / outcomes / market_status_history via pgx.
// Satisfies Repo.
type PgRepo struct{ pool *pgxpool.Pool }

// NewPgRepo returns a PgRepo bound to pool.
func NewPgRepo(pool *pgxpool.Pool) *PgRepo { return &PgRepo{pool: pool} }

func nullableRawID(id [16]byte) any {
	var zero [16]byte
	if id == zero {
		return nil
	}
	return id
}

func nullableMarketStatus(s MarketStatus) any {
	if s == StatusUnknown {
		return nil
	}
	return string(s)
}

// MatchExists reports whether a matches row already exists. Bet_stop and
// odds_change short-circuit when the catalog handler has not yet seen the
// match (the FK on markets.match_id would otherwise fail).
func (r *PgRepo) MatchExists(ctx context.Context, matchID int64) (bool, error) {
	var ok bool
	if err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM matches WHERE id = $1)`, matchID,
	).Scan(&ok); err != nil {
		return false, fmt.Errorf("storage: match exists: %w", err)
	}
	return ok, nil
}

func (r *PgRepo) GetMarket(ctx context.Context, matchID int64, marketTypeID int32, specifier string) (Market, bool, error) {
	const q = `
		SELECT match_id, market_type_id, specifier, status, group_id
		  FROM markets
		 WHERE match_id = $1 AND market_type_id = $2 AND specifier = $3`
	var (
		m      Market
		status string
	)
	err := r.pool.QueryRow(ctx, q, matchID, marketTypeID, specifier).Scan(
		&m.MatchID, &m.MarketTypeID, &m.Specifier, &status, &m.GroupID,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Market{}, false, nil
		}
		return Market{}, false, fmt.Errorf("storage: get market: %w", err)
	}
	m.Status = MarketStatus(status)
	return m, true, nil
}

func (r *PgRepo) UpsertMarket(ctx context.Context, m Market) error {
	const q = `
		INSERT INTO markets (match_id, market_type_id, specifier, status, group_id, updated_at)
		VALUES ($1, $2, $3, $4, $5, now())
		ON CONFLICT (match_id, market_type_id, specifier) DO UPDATE
		   SET status     = EXCLUDED.status,
		       group_id   = COALESCE(EXCLUDED.group_id, markets.group_id),
		       updated_at = now()`
	if _, err := r.pool.Exec(ctx, q,
		m.MatchID, m.MarketTypeID, m.Specifier, string(m.Status), m.GroupID,
	); err != nil {
		return fmt.Errorf("storage: upsert market: %w", err)
	}
	return nil
}

func (r *PgRepo) UpsertOutcome(ctx context.Context, o Outcome) error {
	const q = `
		INSERT INTO outcomes (match_id, market_type_id, specifier, outcome_id, odds, is_active, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, now())
		ON CONFLICT (match_id, market_type_id, specifier, outcome_id) DO UPDATE
		   SET odds       = COALESCE(EXCLUDED.odds, outcomes.odds),
		       is_active  = EXCLUDED.is_active,
		       updated_at = now()`
	if _, err := r.pool.Exec(ctx, q,
		o.MatchID, o.MarketTypeID, o.Specifier, o.OutcomeID, o.Odds, o.IsActive,
	); err != nil {
		return fmt.Errorf("storage: upsert outcome: %w", err)
	}
	return nil
}

func (r *PgRepo) InsertMarketStatusHistory(ctx context.Context, row MarketStatusHistoryRow) error {
	const q = `
		INSERT INTO market_status_history
		    (match_id, market_type_id, specifier, from_status, to_status, raw_message_id, changed_at)
		VALUES ($1, $2, $3, $4, $5, $6, now())`
	if _, err := r.pool.Exec(ctx, q,
		row.MatchID, row.MarketTypeID, row.Specifier,
		nullableMarketStatus(row.From), string(row.To),
		nullableRawID(row.RawMessageID),
	); err != nil {
		return fmt.Errorf("storage: insert market_status_history: %w", err)
	}
	return nil
}

func (r *PgRepo) MarketsForBetStop(ctx context.Context, scope BetStopScope) ([]Market, error) {
	args := []any{scope.MatchID}
	where := "WHERE match_id = $1"
	switch {
	case scope.MarketTypeID != nil:
		args = append(args, *scope.MarketTypeID)
		where += fmt.Sprintf(" AND market_type_id = $%d", len(args))
		if scope.Specifier != "" {
			args = append(args, scope.Specifier)
			where += fmt.Sprintf(" AND specifier = $%d", len(args))
		}
	case scope.GroupID != nil:
		args = append(args, *scope.GroupID)
		where += fmt.Sprintf(" AND group_id = $%d", len(args))
	}
	q := "SELECT match_id, market_type_id, specifier, status, group_id FROM markets " + where
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("storage: markets for bet_stop: %w", err)
	}
	defer rows.Close()
	var out []Market
	for rows.Next() {
		var (
			m      Market
			status string
		)
		if err := rows.Scan(&m.MatchID, &m.MarketTypeID, &m.Specifier, &status, &m.GroupID); err != nil {
			return nil, fmt.Errorf("storage: scan bet_stop row: %w", err)
		}
		m.Status = MarketStatus(status)
		out = append(out, m)
	}
	return out, rows.Err()
}
