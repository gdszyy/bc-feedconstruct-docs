package subscription

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PgRepo is the production pgxpool-backed implementation of Repo.
type PgRepo struct{ pool *pgxpool.Pool }

// NewPgRepo returns a PgRepo bound to pool.
func NewPgRepo(pool *pgxpool.Pool) *PgRepo { return &PgRepo{pool: pool} }

func (r *PgRepo) GetSubscription(ctx context.Context, matchID int64) (Subscription, bool, error) {
	const q = `
		SELECT match_id, product, status,
		       requested_at, subscribed_at, released_at,
		       COALESCE(last_event_id, ''), COALESCE(reason, '')
		  FROM subscriptions
		 WHERE match_id = $1`
	var (
		s              Subscription
		productStr     string
		statusStr      string
		requestedAt    *time.Time
		subscribedAt   *time.Time
		releasedAt     *time.Time
	)
	err := r.pool.QueryRow(ctx, q, matchID).Scan(
		&s.MatchID, &productStr, &statusStr,
		&requestedAt, &subscribedAt, &releasedAt,
		&s.LastEventID, &s.Reason,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Subscription{}, false, nil
		}
		return Subscription{}, false, fmt.Errorf("storage: get subscription: %w", err)
	}
	s.Product = Product(productStr)
	s.Status = Status(statusStr)
	s.RequestedAt = requestedAt
	s.SubscribedAt = subscribedAt
	s.ReleasedAt = releasedAt
	return s, true, nil
}

func (r *PgRepo) UpsertSubscription(ctx context.Context, s Subscription) error {
	const q = `
		INSERT INTO subscriptions (
			match_id, product, status,
			requested_at, subscribed_at, released_at,
			last_event_id, reason)
		VALUES ($1,$2,$3,$4,$5,$6, NULLIF($7,''), NULLIF($8,''))
		ON CONFLICT (match_id) DO UPDATE SET
			product       = EXCLUDED.product,
			status        = EXCLUDED.status,
			requested_at  = COALESCE(EXCLUDED.requested_at, subscriptions.requested_at),
			subscribed_at = COALESCE(EXCLUDED.subscribed_at, subscriptions.subscribed_at),
			released_at   = COALESCE(EXCLUDED.released_at, subscriptions.released_at),
			last_event_id = COALESCE(EXCLUDED.last_event_id, subscriptions.last_event_id),
			reason        = COALESCE(EXCLUDED.reason, subscriptions.reason)`
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
		INSERT INTO subscription_events (
			match_id, from_status, to_status, reason, occurred_at)
		VALUES ($1, NULLIF($2,''), $3, NULLIF($4,''), $5)`
	if _, err := r.pool.Exec(ctx, q,
		e.MatchID, string(e.From), string(e.To), e.Reason, e.OccurredAt,
	); err != nil {
		return fmt.Errorf("storage: insert subscription event: %w", err)
	}
	return nil
}

func (r *PgRepo) ListStuckRequests(ctx context.Context, olderThan time.Time) ([]Subscription, error) {
	const q = `
		SELECT match_id, product, status,
		       requested_at, subscribed_at, released_at,
		       COALESCE(last_event_id, ''), COALESCE(reason, '')
		  FROM subscriptions
		 WHERE status = 'requested'
		   AND requested_at IS NOT NULL
		   AND requested_at <= $1`
	rows, err := r.pool.Query(ctx, q, olderThan)
	if err != nil {
		return nil, fmt.Errorf("storage: list stuck: %w", err)
	}
	defer rows.Close()
	var out []Subscription
	for rows.Next() {
		var (
			s            Subscription
			productStr   string
			statusStr    string
			requestedAt  *time.Time
			subscribedAt *time.Time
			releasedAt   *time.Time
		)
		if err := rows.Scan(
			&s.MatchID, &productStr, &statusStr,
			&requestedAt, &subscribedAt, &releasedAt,
			&s.LastEventID, &s.Reason,
		); err != nil {
			return nil, fmt.Errorf("storage: scan stuck: %w", err)
		}
		s.Product = Product(productStr)
		s.Status = Status(statusStr)
		s.RequestedAt = requestedAt
		s.SubscribedAt = subscribedAt
		s.ReleasedAt = releasedAt
		out = append(out, s)
	}
	return out, rows.Err()
}
