package settlement

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/gdszyy/bc-feedconstruct-docs/backend/internal/feed"
	"github.com/gdszyy/bc-feedconstruct-docs/backend/internal/storage"
)

// Handler dispatches bet_settlement, bet_cancel, rollback and rollback_cancel.
type Handler struct {
	pool *storage.Pool
}

// New returns a Handler bound to pool.
func New(pool *storage.Pool) *Handler { return &Handler{pool: pool} }

// Register binds the handler to its message types.
func (h *Handler) Register(d *feed.Dispatcher) {
	d.Register(feed.MsgBetSettlement, h)
	d.Register(feed.MsgBetCancel, h)
	d.Register(feed.MsgRollback, h)
	d.Register(feed.MsgRollbackCancel, h)
}

// Handle implements feed.Handler.
func (h *Handler) Handle(ctx context.Context, msgType feed.MessageType, env feed.Envelope, rawID [16]byte) error {
	switch msgType {
	case feed.MsgBetSettlement:
		return h.handleSettlement(ctx, env, rawID)
	case feed.MsgBetCancel:
		return h.handleCancel(ctx, env, rawID)
	case feed.MsgRollback:
		return h.handleRollback(ctx, env, rawID, "settlement")
	case feed.MsgRollbackCancel:
		return h.handleRollback(ctx, env, rawID, "cancel")
	}
	return fmt.Errorf("settlement: unsupported message type %q", msgType)
}

// ----- Settlement -----

func (h *Handler) handleSettlement(ctx context.Context, env feed.Envelope, rawID [16]byte) error {
	p, err := parseSettlement(env.Payload)
	if err != nil {
		return fmt.Errorf("settlement: parse: %w", err)
	}
	matchID, markets, ok := p.flatten()
	if !ok {
		return errors.New("settlement: missing matchId")
	}
	if !h.matchExists(ctx, matchID) {
		return nil
	}

	tx, err := h.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("settlement: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	for _, m := range markets {
		mt, ok := m.marketTypeID()
		if !ok {
			continue
		}
		for _, o := range m.outcomes() {
			oid, ok := o.outcomeID()
			if !ok {
				continue
			}
			result, ok := normaliseResult(o.Result)
			if !ok {
				return fmt.Errorf("settlement: unknown result %q (match=%d outcome=%d)", o.Result, matchID, oid)
			}
			if err := h.upsertSettlement(ctx, tx, settlementRow{
				MatchID:        matchID,
				MarketTypeID:   mt,
				Specifier:      m.Specifier,
				OutcomeID:      oid,
				Result:         result,
				Certainty:      o.certainty(),
				VoidFactor:     o.VoidFactor,
				DeadHeatFactor: o.DeadHeatFactor,
				RawID:          rawID,
			}); err != nil {
				return err
			}
		}
		// Transition the market to settled (no-regress: cancelled wins).
		if err := h.transitionMarket(ctx, tx, matchID, mt, m.Specifier, "settled", rawID); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

type settlementRow struct {
	MatchID        int64
	MarketTypeID   int64
	Specifier      string
	OutcomeID      int64
	Result         string
	Certainty      int
	VoidFactor     *float64
	DeadHeatFactor *float64
	RawID          [16]byte
}

func (h *Handler) upsertSettlement(ctx context.Context, tx pgx.Tx, r settlementRow) error {
	// Idempotency: same (match, marketType, specifier, outcome, settled_at)
	// collapses via UNIQUE INDEX. We set settled_at deterministically per
	// settlement run so duplicate deliveries don't insert twice. Use the
	// "current row" pattern: if a later certain row arrives for an existing
	// uncertain row, update in place rather than appending.
	var (
		existingID        *int64
		existingCertainty *int
	)
	err := tx.QueryRow(ctx, `
		SELECT id, certainty FROM settlements
		WHERE match_id=$1 AND market_type_id=$2 AND specifier=$3 AND outcome_id=$4
		ORDER BY settled_at DESC, id DESC LIMIT 1
		FOR UPDATE`,
		r.MatchID, r.MarketTypeID, r.Specifier, r.OutcomeID,
	).Scan(&existingID, &existingCertainty)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("settlement: lookup existing (%d,%d,%q,%d): %w",
			r.MatchID, r.MarketTypeID, r.Specifier, r.OutcomeID, err)
	}

	if existingID != nil {
		if existingCertainty != nil && *existingCertainty == 1 && r.Certainty == 1 {
			// Same certain row already exists — idempotent no-op.
			return nil
		}
		if existingCertainty != nil && *existingCertainty == 1 && r.Certainty == 0 {
			// Don't downgrade certainty.
			return nil
		}
		// Update the existing row in place. Certain supersedes uncertain.
		if _, err := tx.Exec(ctx, `
			UPDATE settlements
			   SET result=$1, certainty=$2, void_factor=$3, dead_heat_factor=$4,
			       raw_message_id = COALESCE($5, raw_message_id), settled_at = now(),
			       rolled_back_at = NULL
			 WHERE id = $6`,
			r.Result, r.Certainty, r.VoidFactor, r.DeadHeatFactor,
			nullableRawID(r.RawID), *existingID,
		); err != nil {
			return fmt.Errorf("settlement: update existing: %w", err)
		}
		return nil
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO settlements
		    (match_id, market_type_id, specifier, outcome_id, result, certainty,
		     void_factor, dead_heat_factor, raw_message_id, settled_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9, now())`,
		r.MatchID, r.MarketTypeID, r.Specifier, r.OutcomeID,
		r.Result, r.Certainty, r.VoidFactor, r.DeadHeatFactor,
		nullableRawID(r.RawID),
	); err != nil {
		return fmt.Errorf("settlement: insert: %w", err)
	}
	return nil
}

// ----- Cancel -----

func (h *Handler) handleCancel(ctx context.Context, env feed.Envelope, rawID [16]byte) error {
	p, err := parseCancel(env.Payload)
	if err != nil {
		return fmt.Errorf("cancel: parse: %w", err)
	}
	matchID, ok := pickID(p.MatchID, matchIDFromObject(p))
	if !ok {
		return errors.New("cancel: missing matchId")
	}
	if !h.matchExists(ctx, matchID) {
		return nil
	}

	tx, err := h.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("cancel: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var supercededBy *int64
	if p.SupercededBy != nil {
		// Best-effort link via raw id of the prior cancel that this one supersedes.
		supercededBy = p.SupercededBy
	}

	var (
		marketTypeID *int64
		specifier    = p.Specifier
	)
	if p.MarketTypeID != nil {
		marketTypeID = p.MarketTypeID
	} else if p.TypeID != nil {
		marketTypeID = p.TypeID
	}

	var cancelID int64
	err = tx.QueryRow(ctx, `
		INSERT INTO cancels
		    (match_id, market_type_id, specifier, void_reason, void_action,
		     superceded_by, from_ts, to_ts, raw_message_id, cancelled_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9, now())
		RETURNING id`,
		matchID, marketTypeID, specifier, nullableString(p.Reason),
		1, // VoidAction=1 (void); rollback_cancel uses a separate handler
		supercededBy, p.FromDate, p.ToDate, nullableRawID(rawID),
	).Scan(&cancelID)
	if err != nil {
		return fmt.Errorf("cancel: insert: %w", err)
	}

	// Propagate to market.status. ObjectType=4 (match) cancels every market.
	if p.ObjectType == 4 || marketTypeID == nil {
		rows, err := tx.Query(ctx,
			`SELECT market_type_id, specifier FROM markets WHERE match_id=$1 FOR UPDATE`, matchID)
		if err != nil {
			return fmt.Errorf("cancel: scan markets: %w", err)
		}
		var keys []struct {
			mt  int64
			spc string
		}
		for rows.Next() {
			var k struct {
				mt  int64
				spc string
			}
			if err := rows.Scan(&k.mt, &k.spc); err != nil {
				rows.Close()
				return fmt.Errorf("cancel: scan market row: %w", err)
			}
			keys = append(keys, k)
		}
		rows.Close()
		for _, k := range keys {
			if err := h.transitionMarket(ctx, tx, matchID, k.mt, k.spc, "cancelled", rawID); err != nil {
				return err
			}
		}
	} else {
		if err := h.transitionMarket(ctx, tx, matchID, *marketTypeID, specifier, "cancelled", rawID); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

// ----- Rollback -----

func (h *Handler) handleRollback(ctx context.Context, env feed.Envelope, rawID [16]byte, target string) error {
	p, err := parseRollback(env.Payload)
	if err != nil {
		return fmt.Errorf("rollback: parse: %w", err)
	}
	matchID, ok := pickID(p.MatchID)
	if !ok {
		return errors.New("rollback: missing matchId")
	}
	if !h.matchExists(ctx, matchID) {
		return nil
	}

	tx, err := h.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("rollback: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	mt := p.MarketTypeID
	if mt == nil {
		mt = p.TypeID
	}
	specifier := p.Specifier

	var targetIDs []int64
	switch target {
	case "settlement":
		q := `SELECT id FROM settlements WHERE match_id=$1 AND rolled_back_at IS NULL`
		args := []any{matchID}
		if mt != nil {
			q += ` AND market_type_id=$2`
			args = append(args, *mt)
			if specifier != "" {
				q += ` AND specifier=$3`
				args = append(args, specifier)
			}
			if p.OutcomeID != nil {
				q += ` AND outcome_id=$` + itoa(len(args)+1)
				args = append(args, *p.OutcomeID)
			}
		}
		q += ` FOR UPDATE`
		ids, err := scanInt64(ctx, tx, q, args...)
		if err != nil {
			return err
		}
		targetIDs = ids
		for _, id := range ids {
			if _, err := tx.Exec(ctx,
				`UPDATE settlements SET rolled_back_at=now() WHERE id=$1`, id,
			); err != nil {
				return fmt.Errorf("rollback: update settlement %d: %w", id, err)
			}
		}
	case "cancel":
		q := `SELECT id FROM cancels WHERE match_id=$1 AND rolled_back_at IS NULL`
		args := []any{matchID}
		if mt != nil {
			q += ` AND market_type_id=$2`
			args = append(args, *mt)
			if specifier != "" {
				q += ` AND specifier=$3`
				args = append(args, specifier)
			}
		}
		q += ` FOR UPDATE`
		ids, err := scanInt64(ctx, tx, q, args...)
		if err != nil {
			return err
		}
		targetIDs = ids
		for _, id := range ids {
			if _, err := tx.Exec(ctx,
				`UPDATE cancels SET rolled_back_at=now() WHERE id=$1`, id,
			); err != nil {
				return fmt.Errorf("rollback: update cancel %d: %w", id, err)
			}
		}
	default:
		return fmt.Errorf("rollback: unknown target %q", target)
	}

	// Write rollbacks rows (idempotent via UNIQUE (target, target_id, raw_message_id)).
	for _, id := range targetIDs {
		if _, err := tx.Exec(ctx, `
			INSERT INTO rollbacks (target, target_id, raw_message_id, applied_at)
			VALUES ($1, $2, $3, now())
			ON CONFLICT (target, target_id, raw_message_id) DO NOTHING`,
			target, id, nullableRawID(rawID),
		); err != nil {
			return fmt.Errorf("rollback: insert rollbacks row: %w", err)
		}
	}

	// Best-effort market.status recovery. For a settlement rollback we drop
	// the market back to "active" so it can accept new odds; for a cancel
	// rollback we recover the prior status by walking market_status_history.
	if mt != nil {
		previous := "active"
		if target == "cancel" {
			if recovered, ok := h.priorStatus(ctx, tx, matchID, *mt, specifier, "cancelled"); ok {
				previous = recovered
			}
		}
		if _, err := tx.Exec(ctx,
			`UPDATE markets SET status=$4, updated_at=now()
			   WHERE match_id=$1 AND market_type_id=$2 AND specifier=$3`,
			matchID, *mt, specifier, previous,
		); err != nil {
			return fmt.Errorf("rollback: revert market: %w", err)
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO market_status_history
			    (match_id, market_type_id, specifier, from_status, to_status, raw_message_id, changed_at)
			VALUES ($1,$2,$3, $4, $5, $6, now())`,
			matchID, *mt, specifier,
			func() any {
				if target == "settlement" {
					return "settled"
				}
				return "cancelled"
			}(),
			previous, nullableRawID(rawID),
		); err != nil {
			return fmt.Errorf("rollback: market history: %w", err)
		}
	}
	return tx.Commit(ctx)
}

// priorStatus returns the most recent from_status that preceded
// `currentStatus` in market_status_history.
func (h *Handler) priorStatus(ctx context.Context, tx pgx.Tx, matchID, marketTypeID int64, specifier, currentStatus string) (string, bool) {
	var prior *string
	err := tx.QueryRow(ctx, `
		SELECT from_status
		  FROM market_status_history
		 WHERE match_id=$1 AND market_type_id=$2 AND specifier=$3
		   AND to_status=$4
		 ORDER BY changed_at DESC, id DESC LIMIT 1`,
		matchID, marketTypeID, specifier, currentStatus,
	).Scan(&prior)
	if err != nil || prior == nil || *prior == "" {
		return "", false
	}
	return *prior, true
}

// ----- helpers -----

func (h *Handler) matchExists(ctx context.Context, matchID int64) bool {
	var ok bool
	if err := h.pool.QueryRow(ctx,
		`SELECT exists(SELECT 1 FROM matches WHERE id=$1)`, matchID,
	).Scan(&ok); err != nil {
		return false
	}
	return ok
}

// transitionMarket moves market.status to `to`, respecting status rank
// rules implemented inline (cancelled > settled > everything; settled
// cannot regress to active). Writes a history row when the status moves.
func (h *Handler) transitionMarket(ctx context.Context, tx pgx.Tx, matchID, marketTypeID int64, specifier, to string, rawID [16]byte) error {
	var cur string
	err := tx.QueryRow(ctx,
		`SELECT status FROM markets
		   WHERE match_id=$1 AND market_type_id=$2 AND specifier=$3 FOR UPDATE`,
		matchID, marketTypeID, specifier,
	).Scan(&cur)
	if errors.Is(err, pgx.ErrNoRows) {
		// Market row may not exist yet (cancel before any odds_change).
		// Insert directly in target status.
		_, err := tx.Exec(ctx, `
			INSERT INTO markets (match_id, market_type_id, specifier, status, updated_at)
			VALUES ($1,$2,$3,$4, now())
			ON CONFLICT (match_id, market_type_id, specifier) DO NOTHING`,
			matchID, marketTypeID, specifier, to)
		if err != nil {
			return fmt.Errorf("settlement: insert market (%d,%d,%q): %w", matchID, marketTypeID, specifier, err)
		}
		_, err = tx.Exec(ctx, `
			INSERT INTO market_status_history
			    (match_id, market_type_id, specifier, from_status, to_status, raw_message_id, changed_at)
			VALUES ($1,$2,$3, NULL, $4, $5, now())`,
			matchID, marketTypeID, specifier, to, nullableRawID(rawID))
		return err
	}
	if err != nil {
		return fmt.Errorf("settlement: lookup market: %w", err)
	}
	if !canTransitionTerminal(cur, to) {
		return nil // no-regress
	}
	if cur == to {
		return nil
	}
	if _, err := tx.Exec(ctx,
		`UPDATE markets SET status=$4, updated_at=now()
		   WHERE match_id=$1 AND market_type_id=$2 AND specifier=$3`,
		matchID, marketTypeID, specifier, to,
	); err != nil {
		return fmt.Errorf("settlement: update market: %w", err)
	}
	_, err = tx.Exec(ctx, `
		INSERT INTO market_status_history
		    (match_id, market_type_id, specifier, from_status, to_status, raw_message_id, changed_at)
		VALUES ($1,$2,$3,$4,$5,$6, now())`,
		matchID, marketTypeID, specifier, nullableString(cur), to, nullableRawID(rawID))
	return err
}

// canTransitionTerminal: cancelled > settled > handed_over > everything else.
// active/suspended/deactivated → settled allowed; settled → cancelled allowed;
// settled/cancelled/handed_over → active blocked.
func canTransitionTerminal(from, to string) bool {
	rank := map[string]int{
		"active":      1,
		"suspended":   2,
		"deactivated": 3,
		"settled":     10,
		"handed_over": 11,
		"cancelled":   20,
	}
	return rank[to] >= rank[from]
}

func matchIDFromObject(p cancelPayload) *int64 {
	if p.ObjectType == 4 && p.ObjectID != nil {
		return p.ObjectID
	}
	return nil
}

func scanInt64(ctx context.Context, tx pgx.Tx, q string, args ...any) ([]int64, error) {
	rows, err := tx.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("settlement: scan ids: %w", err)
	}
	defer rows.Close()
	var out []int64
	for rows.Next() {
		var v int64
		if err := rows.Scan(&v); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, nil
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

func itoa(i int) string {
	const digits = "0123456789"
	if i == 0 {
		return "0"
	}
	var b []byte
	for i > 0 {
		b = append([]byte{digits[i%10]}, b...)
		i /= 10
	}
	return string(b)
}
