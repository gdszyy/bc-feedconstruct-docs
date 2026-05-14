package odds

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/gdszyy/bc-feedconstruct-docs/backend/internal/feed"
	"github.com/gdszyy/bc-feedconstruct-docs/backend/internal/storage"
)

// Handler routes odds_change + bet_stop deliveries.
type Handler struct {
	pool *storage.Pool
}

// New returns a Handler bound to pool.
func New(pool *storage.Pool) *Handler { return &Handler{pool: pool} }

// Register binds the handler to its message types.
func (h *Handler) Register(d *feed.Dispatcher) {
	d.Register(feed.MsgOddsChange, h)
	d.Register(feed.MsgBetStop, h)
}

// Handle implements feed.Handler.
func (h *Handler) Handle(ctx context.Context, msgType feed.MessageType, env feed.Envelope, rawID [16]byte) error {
	switch msgType {
	case feed.MsgOddsChange:
		return h.handleOddsChange(ctx, env, rawID)
	case feed.MsgBetStop:
		return h.handleBetStop(ctx, env, rawID)
	}
	return fmt.Errorf("odds: unsupported message type %q", msgType)
}

func (h *Handler) handleOddsChange(ctx context.Context, env feed.Envelope, rawID [16]byte) error {
	p, err := parseOddsChange(env.Payload)
	if err != nil {
		return fmt.Errorf("odds: parse odds_change: %w", err)
	}
	matchID, markets, ok := p.flatten()
	if !ok {
		return errors.New("odds: odds_change without matchId")
	}
	if !h.matchExists(ctx, matchID) {
		// Catalog handler creates the match; if it hasn't yet, we'd FK-fail.
		// Skip silently — the catalog delivery is expected to arrive shortly
		// and a recovery snapshot will fill in any persistent gap.
		return nil
	}
	if len(markets) == 0 {
		return nil
	}

	tx, err := h.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("odds: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	for _, m := range markets {
		mt, ok := m.marketTypeID()
		if !ok {
			continue
		}
		if err := h.upsertMarket(ctx, tx, matchID, mt, m, rawID); err != nil {
			return err
		}
		for _, o := range m.outcomes() {
			if o.ID == nil {
				continue
			}
			if err := h.upsertOutcome(ctx, tx, matchID, mt, m.Specifier, o); err != nil {
				return err
			}
		}
	}
	return tx.Commit(ctx)
}

func (h *Handler) handleBetStop(ctx context.Context, env feed.Envelope, rawID [16]byte) error {
	p, err := parseBetStop(env.Payload)
	if err != nil {
		return fmt.Errorf("odds: parse bet_stop: %w", err)
	}
	matchID, ok := pickID(p.MatchID, p.ID, p.ObjectID)
	if !ok {
		return errors.New("odds: bet_stop without matchId")
	}
	if !h.matchExists(ctx, matchID) {
		return nil
	}
	target := normaliseMarketStatus(firstNonEmpty(p.Status, p.MarketStatus))
	if target == "" {
		target = "suspended"
	}

	tx, err := h.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("odds: begin bet_stop tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// Resolve targeted markets.
	type marketKey struct {
		MarketTypeID int64
		Specifier    string
		Status       string
	}
	rows, err := tx.Query(ctx, marketQuery(p), marketArgs(matchID, p)...)
	if err != nil {
		return fmt.Errorf("odds: scan markets for bet_stop: %w", err)
	}
	var keys []marketKey
	for rows.Next() {
		var k marketKey
		if err := rows.Scan(&k.MarketTypeID, &k.Specifier, &k.Status); err != nil {
			rows.Close()
			return fmt.Errorf("odds: scan bet_stop row: %w", err)
		}
		keys = append(keys, k)
	}
	rows.Close()

	for _, k := range keys {
		if !allowsTransition(k.Status, target) {
			// Acceptance #12 market-level: terminal states like settled
			// must not regress to suspended.
			continue
		}
		if k.Status == target {
			continue
		}
		if err := h.transitionMarket(ctx, tx, matchID, k.MarketTypeID, k.Specifier, k.Status, target, rawID); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (h *Handler) matchExists(ctx context.Context, matchID int64) bool {
	var ok bool
	if err := h.pool.QueryRow(ctx,
		`SELECT exists(SELECT 1 FROM matches WHERE id = $1)`, matchID,
	).Scan(&ok); err != nil {
		return false
	}
	return ok
}

func (h *Handler) upsertMarket(ctx context.Context, tx pgx.Tx, matchID, marketTypeID int64, m marketPayload, rawID [16]byte) error {
	target := normaliseMarketStatus(m.statusString())
	if target == "" {
		target = "active"
	}

	var curStatus string
	err := tx.QueryRow(ctx,
		`SELECT status FROM markets
		   WHERE match_id=$1 AND market_type_id=$2 AND specifier=$3
		   FOR UPDATE`,
		matchID, marketTypeID, m.Specifier,
	).Scan(&curStatus)
	exists := err == nil
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("odds: lookup market (%d,%d,%q): %w", matchID, marketTypeID, m.Specifier, err)
	}

	writeStatus := target
	if exists && !allowsTransition(curStatus, target) {
		writeStatus = curStatus
		fmt.Printf("market.status.regress.blocked match=%d type=%d spec=%q from=%s to=%s\n",
			matchID, marketTypeID, m.Specifier, curStatus, target)
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO markets (match_id, market_type_id, specifier, status, group_id, updated_at)
		VALUES ($1, $2, $3, $4, $5, now())
		ON CONFLICT (match_id, market_type_id, specifier) DO UPDATE
		   SET status = EXCLUDED.status,
		       group_id = COALESCE(EXCLUDED.group_id, markets.group_id),
		       updated_at = now()`,
		matchID, marketTypeID, m.Specifier, writeStatus, m.GroupID,
	)
	if err != nil {
		return fmt.Errorf("odds: upsert market (%d,%d,%q): %w", matchID, marketTypeID, m.Specifier, err)
	}

	if !exists || curStatus != writeStatus {
		_, err = tx.Exec(ctx, `
			INSERT INTO market_status_history
			    (match_id, market_type_id, specifier, from_status, to_status, raw_message_id, changed_at)
			VALUES ($1, $2, $3, $4, $5, $6, now())`,
			matchID, marketTypeID, m.Specifier, nullableString(curStatus), writeStatus, nullableRawID(rawID),
		)
		if err != nil {
			return fmt.Errorf("odds: history (%d,%d,%q): %w", matchID, marketTypeID, m.Specifier, err)
		}
	}
	return nil
}

func (h *Handler) upsertOutcome(ctx context.Context, tx pgx.Tx, matchID, marketTypeID int64, specifier string, o outcomePayload) error {
	if o.ID == nil {
		return nil
	}
	_, err := tx.Exec(ctx, `
		INSERT INTO outcomes (match_id, market_type_id, specifier, outcome_id, odds, is_active, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, now())
		ON CONFLICT (match_id, market_type_id, specifier, outcome_id) DO UPDATE
		   SET odds = COALESCE(EXCLUDED.odds, outcomes.odds),
		       is_active = EXCLUDED.is_active,
		       updated_at = now()`,
		matchID, marketTypeID, specifier, *o.ID, o.Odds, o.active(),
	)
	if err != nil {
		return fmt.Errorf("odds: upsert outcome (%d,%d,%q,%d): %w", matchID, marketTypeID, specifier, *o.ID, err)
	}
	return nil
}

func (h *Handler) transitionMarket(ctx context.Context, tx pgx.Tx, matchID, marketTypeID int64, specifier, from, to string, rawID [16]byte) error {
	_, err := tx.Exec(ctx,
		`UPDATE markets SET status=$4, updated_at=now()
		   WHERE match_id=$1 AND market_type_id=$2 AND specifier=$3`,
		matchID, marketTypeID, specifier, to,
	)
	if err != nil {
		return fmt.Errorf("odds: transition market (%d,%d,%q): %w", matchID, marketTypeID, specifier, err)
	}
	_, err = tx.Exec(ctx, `
		INSERT INTO market_status_history
		    (match_id, market_type_id, specifier, from_status, to_status, raw_message_id, changed_at)
		VALUES ($1, $2, $3, $4, $5, $6, now())`,
		matchID, marketTypeID, specifier, nullableString(from), to, nullableRawID(rawID),
	)
	if err != nil {
		return fmt.Errorf("odds: transition history (%d,%d,%q): %w", matchID, marketTypeID, specifier, err)
	}
	return nil
}

// marketQuery builds the SELECT to enumerate the markets a bet_stop targets.
// FOR UPDATE must follow every WHERE clause.
func marketQuery(p betStopPayload) string {
	where := `WHERE match_id = $1`
	if p.MarketTypeID != nil || p.TypeID != nil {
		where += ` AND market_type_id = $2`
		if p.Specifier != "" {
			where += ` AND specifier = $3`
		}
	} else if p.GroupID != nil {
		where += ` AND group_id = $2`
	}
	return `SELECT market_type_id, specifier, status FROM markets ` + where + ` FOR UPDATE`
}

func marketArgs(matchID int64, p betStopPayload) []any {
	args := []any{matchID}
	if p.MarketTypeID != nil || p.TypeID != nil {
		mt := p.MarketTypeID
		if mt == nil {
			mt = p.TypeID
		}
		args = append(args, *mt)
		if p.Specifier != "" {
			args = append(args, p.Specifier)
		}
		return args
	}
	if p.GroupID != nil {
		args = append(args, *p.GroupID)
	}
	return args
}

func nullableString(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func nullableRawID(id [16]byte) any {
	var zero [16]byte
	if id == zero {
		return nil
	}
	return id
}
