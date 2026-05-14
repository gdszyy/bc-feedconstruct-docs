package subscription

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PgRepo persists subscriptions and subscription_events via pgx.
// Satisfies Repo.
type PgRepo struct{ pool *pgxpool.Pool }

// NewPgRepo returns a PgRepo bound to pool.
func NewPgRepo(pool *pgxpool.Pool) *PgRepo { return &PgRepo{pool: pool} }

func (r *PgRepo) GetSubscription(ctx context.Context, matchID int64) (Subscription, bool, error) {
	const q = `
		SELECT match_id, product, status,
		       requested_at, subscribed_at, released_at,
		       last_event_id, reason
		  FROM subscriptions
		 WHERE match_id = $1`
	var (
		s      Subscription
		prod   string
		status string
	)
	err := r.pool.QueryRow(ctx, q, matchID).Scan(
		&s.MatchID, &prod, &status,
		&s.RequestedAt, &s.SubscribedAt, &s.ReleasedAt,
		&s.LastEventID, &s.Reason,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Subscription{}, false, nil
		}
		return Subscription{}, false, fmt.Errorf("storage: get subscription: %w", err)
	}
	s.Product = Product(prod)
	s.Status = Status(status)
	return s, true, nil
}

func (r *PgRepo) UpsertSubscription(ctx context.Context, s Subscription) error {
	const q = `
		INSERT INTO subscriptions (
			match_id, product, status,
			requested_at, subscribed_at, released_at,
			last_event_id, reason)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		ON CONFLICT (match_id) DO UPDATE SET
			product       = EXCLUDED.product,
			status        = EXCLUDED.status,
			requested_at  = COALESCE(EXCLUDED.requested_at, subscriptions.requested_at),
			subscribed_at = COALESCE(EXCLUDED.subscribed_at, subscriptions.subscribed_at),
			released_at   = EXCLUDED.released_at,
			last_event_id = COALESCE(EXCLUDED.last_event_id, subscriptions.last_event_id),
			reason        = EXCLUDED.reason`
	if _, err := r.pool.Exec(ctx, q,
		s.MatchID, string(s.Product), string(s.Status),
		s.RequestedAt, s.SubscribedAt, s.ReleasedAt,
		s.LastEventID, s.Reason,
	); err != nil {
		return fmt.Errorf("storage: upsert subscription: %w", err)
	}
	return nil
}

func (r *PgRepo) InsertEvent(ctx context.Context, e Event) error {
	const q = `
		INSERT INTO subscription_events (match_id, from_status, to_status, reason, occurred_at)
		VALUES ($1,$2,$3,$4,$5)
		RETURNING id`
	var from any
	if e.FromStatus != nil {
		from = string(*e.FromStatus)
	}
	if err := r.pool.QueryRow(ctx, q,
		e.MatchID, from, string(e.ToStatus), e.Reason, e.OccurredAt,
	).Scan(&e.ID); err != nil {
		return fmt.Errorf("storage: insert subscription event: %w", err)
	}
	return nil
}

func (r *PgRepo) ListByStatus(ctx context.Context, status Status) ([]Subscription, error) {
	const q = `
		SELECT match_id, product, status,
		       requested_at, subscribed_at, released_at,
		       last_event_id, reason
		  FROM subscriptions
		 WHERE status = $1
		 ORDER BY match_id`
	rows, err := r.pool.Query(ctx, q, string(status))
	if err != nil {
		return nil, fmt.Errorf("storage: list subscriptions: %w", err)
	}
	defer rows.Close()
	var out []Subscription
	for rows.Next() {
		var (
			s      Subscription
			prod   string
			stat   string
		)
		if err := rows.Scan(
			&s.MatchID, &prod, &stat,
			&s.RequestedAt, &s.SubscribedAt, &s.ReleasedAt,
			&s.LastEventID, &s.Reason,
		); err != nil {
			return nil, fmt.Errorf("storage: scan subscription: %w", err)
		}
		s.Product = Product(prod)
		s.Status = Status(stat)
		out = append(out, s)
	}
	return out, rows.Err()
}
