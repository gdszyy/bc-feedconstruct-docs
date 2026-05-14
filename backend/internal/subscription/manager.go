package subscription

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/gdszyy/bc-feedconstruct-docs/backend/internal/feed"
	"github.com/gdszyy/bc-feedconstruct-docs/backend/internal/storage"
)

// Manager owns the subscriptions / subscription_events tables and exposes
// hooks for the feed dispatcher (book/unbook) and the catalog observer
// (auto-unbook on terminal status), plus a janitor for stuck rows.
type Manager struct {
	pool         *storage.Pool
	stuckTimeout time.Duration
}

// Options tunes the Manager. StuckTimeout defaults to 5 minutes.
type Options struct {
	StuckTimeout time.Duration
}

// New returns a Manager bound to pool.
func New(pool *storage.Pool, opt Options) *Manager {
	if opt.StuckTimeout == 0 {
		opt.StuckTimeout = 5 * time.Minute
	}
	return &Manager{pool: pool, stuckTimeout: opt.StuckTimeout}
}

// Register binds the Manager to the feed dispatcher for book/unbook events.
func (m *Manager) Register(d *feed.Dispatcher) {
	d.Register(feed.MsgSubscriptionBook, feed.HandlerFunc(m.handleBook))
	d.Register(feed.MsgSubscriptionUnbk, feed.HandlerFunc(m.handleUnbook))
}

func (m *Manager) handleBook(ctx context.Context, _ feed.MessageType, env feed.Envelope, _ [16]byte) error {
	p, err := parseBooking(env.Payload)
	if err != nil {
		return fmt.Errorf("subscription: parse book: %w", err)
	}
	matchID, ok := p.matchID()
	if !ok {
		return errors.New("subscription: book without matchId")
	}
	product := p.product()
	if product == "" {
		product = "live" // sensible default; book without product is rare
	}
	success := true
	if p.Success != nil {
		success = *p.Success
	}
	to := "subscribed"
	if !success {
		to = "failed"
	}
	return m.transition(ctx, matchID, product, to, p.Reason)
}

func (m *Manager) handleUnbook(ctx context.Context, _ feed.MessageType, env feed.Envelope, _ [16]byte) error {
	p, err := parseBooking(env.Payload)
	if err != nil {
		return fmt.Errorf("subscription: parse unbook: %w", err)
	}
	matchID, ok := p.matchID()
	if !ok {
		return errors.New("subscription: unbook without matchId")
	}
	product := p.product()
	return m.transition(ctx, matchID, product, "unsubscribed", p.Reason)
}

// OnMatchTerminal implements catalog.MatchObserver. When a match reaches a
// terminal status, any non-released subscription is released with the
// match's status as the reason.
func (m *Manager) OnMatchTerminal(ctx context.Context, matchID int64, status string) error {
	tx, err := m.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("subscription: begin terminal tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var (
		cur    string
		product string
	)
	err = tx.QueryRow(ctx,
		`SELECT status, product FROM subscriptions WHERE match_id=$1 FOR UPDATE`,
		matchID,
	).Scan(&cur, &product)
	if errors.Is(err, pgx.ErrNoRows) {
		return tx.Commit(ctx) // not subscribed; nothing to do
	}
	if err != nil {
		return fmt.Errorf("subscription: lookup terminal: %w", err)
	}
	if cur == "unsubscribed" || cur == "expired" || cur == "failed" {
		return tx.Commit(ctx) // already released
	}
	if _, err := tx.Exec(ctx, `
		UPDATE subscriptions
		   SET status = 'unsubscribed', released_at = now(), reason = $2
		 WHERE match_id = $1`,
		matchID, fmt.Sprintf("match_%s", status),
	); err != nil {
		return fmt.Errorf("subscription: terminal update: %w", err)
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO subscription_events (match_id, from_status, to_status, reason)
		VALUES ($1, $2, 'unsubscribed', $3)`,
		matchID, cur, fmt.Sprintf("match_%s", status),
	); err != nil {
		return fmt.Errorf("subscription: terminal event: %w", err)
	}
	return tx.Commit(ctx)
}

// CleanupStuck marks subscriptions stuck in `requested` for longer than the
// configured StuckTimeout as `failed`. Returns the number of rows touched.
func (m *Manager) CleanupStuck(ctx context.Context) (int64, error) {
	cutoff := time.Now().Add(-m.stuckTimeout)
	tx, err := m.pool.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("subscription: cleanup begin: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	rows, err := tx.Query(ctx, `
		SELECT match_id FROM subscriptions
		 WHERE status='requested' AND requested_at < $1
		 FOR UPDATE`, cutoff)
	if err != nil {
		return 0, fmt.Errorf("subscription: cleanup scan: %w", err)
	}
	defer rows.Close()
	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return 0, fmt.Errorf("subscription: cleanup row: %w", err)
		}
		ids = append(ids, id)
	}
	rows.Close()

	for _, id := range ids {
		if _, err := tx.Exec(ctx, `
			UPDATE subscriptions
			   SET status='failed', released_at=now(), reason='stuck_request'
			 WHERE match_id=$1`, id); err != nil {
			return 0, fmt.Errorf("subscription: cleanup update %d: %w", id, err)
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO subscription_events (match_id, from_status, to_status, reason)
			VALUES ($1, 'requested', 'failed', 'stuck_request')`, id); err != nil {
			return 0, fmt.Errorf("subscription: cleanup event %d: %w", id, err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("subscription: cleanup commit: %w", err)
	}
	return int64(len(ids)), nil
}

// transition is the shared upsert path for book / unbook.
func (m *Manager) transition(ctx context.Context, matchID int64, product string, to, reason string) error {
	tx, err := m.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("subscription: begin transition: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var (
		cur          string
		curProduct   string
		hadRow       bool
		requestedAt  *time.Time
	)
	err = tx.QueryRow(ctx, `
		SELECT status, product, requested_at FROM subscriptions
		 WHERE match_id=$1 FOR UPDATE`, matchID,
	).Scan(&cur, &curProduct, &requestedAt)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("subscription: lookup transition: %w", err)
	}
	hadRow = err == nil

	if !hadRow {
		// First sight of this match. If it's an unbook, record it directly.
		if to == "unsubscribed" {
			if _, err := tx.Exec(ctx, `
				INSERT INTO subscriptions
				    (match_id, product, status, requested_at, released_at, reason)
				VALUES ($1, $2, 'unsubscribed', now(), now(), $3)`,
				matchID, fallback(product, "live"), nullableString(reason),
			); err != nil {
				return fmt.Errorf("subscription: insert unsubscribed: %w", err)
			}
		} else {
			if _, err := tx.Exec(ctx, `
				INSERT INTO subscriptions
				    (match_id, product, status, requested_at, subscribed_at, reason)
				VALUES ($1, $2, $3, now(), CASE WHEN $3 = 'subscribed' THEN now() ELSE NULL END, $4)`,
				matchID, fallback(product, "live"), to, nullableString(reason),
			); err != nil {
				return fmt.Errorf("subscription: insert: %w", err)
			}
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO subscription_events (match_id, from_status, to_status, reason)
			VALUES ($1, 'requested', $2, $3)`,
			matchID, to, nullableString(reason),
		); err != nil {
			return fmt.Errorf("subscription: event: %w", err)
		}
		return tx.Commit(ctx)
	}

	// Existing row.
	switch to {
	case "subscribed":
		if _, err := tx.Exec(ctx, `
			UPDATE subscriptions
			   SET status='subscribed', subscribed_at=COALESCE(subscribed_at, now()),
			       released_at=NULL, reason=COALESCE($2, reason)
			 WHERE match_id=$1`, matchID, nullableString(reason),
		); err != nil {
			return fmt.Errorf("subscription: mark subscribed: %w", err)
		}
	case "unsubscribed":
		if _, err := tx.Exec(ctx, `
			UPDATE subscriptions
			   SET status='unsubscribed', released_at=now(), reason=COALESCE($2, reason)
			 WHERE match_id=$1`, matchID, nullableString(reason),
		); err != nil {
			return fmt.Errorf("subscription: mark unsubscribed: %w", err)
		}
	case "failed":
		if _, err := tx.Exec(ctx, `
			UPDATE subscriptions
			   SET status='failed', released_at=now(), reason=COALESCE($2, reason)
			 WHERE match_id=$1`, matchID, nullableString(reason),
		); err != nil {
			return fmt.Errorf("subscription: mark failed: %w", err)
		}
	default:
		return fmt.Errorf("subscription: unsupported transition target %q", to)
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO subscription_events (match_id, from_status, to_status, reason)
		VALUES ($1, $2, $3, $4)`,
		matchID, cur, to, nullableString(reason),
	); err != nil {
		return fmt.Errorf("subscription: event: %w", err)
	}
	return tx.Commit(ctx)
}

func fallback(v, d string) string {
	if v != "" {
		return v
	}
	return d
}

func nullableString(s string) any {
	if s == "" {
		return nil
	}
	return s
}
