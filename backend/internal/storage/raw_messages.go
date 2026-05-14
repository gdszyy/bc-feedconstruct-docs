package storage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// RawMessage is the audit row written before any handler runs.
// All fields except ID/ReceivedAt are populated by the feed consumer.
type RawMessage struct {
	ID          [16]byte
	ReceivedAt  time.Time
	Source      string
	RoutingKey  string
	Queue       string
	MessageType string
	EventID     string
	ProductID   *int16
	SportID     *int32
	TSProvider  *time.Time
	Payload     []byte // JSON, NOT NULL
	RawBlob     []byte // optional original GZIP bytes
}

// InsertResult reports whether the insert wrote a new row or collapsed
// onto an existing audit row (acceptance #11 — idempotency).
type InsertResult struct {
	ID         [16]byte
	Inserted   bool
	ReceivedAt time.Time
}

// RawMessageRepo persists raw_messages.
type RawMessageRepo struct{ pool *Pool }

func NewRawMessageRepo(pool *Pool) *RawMessageRepo { return &RawMessageRepo{pool: pool} }

// Insert writes the message; if a row with the same idempotency key
// already exists it returns Inserted=false with the existing ID.
func (r *RawMessageRepo) Insert(ctx context.Context, m RawMessage) (InsertResult, error) {
	if len(m.Payload) == 0 {
		return InsertResult{}, errors.New("storage: payload required")
	}
	if m.Source == "" || m.MessageType == "" {
		return InsertResult{}, errors.New("storage: source and message_type required")
	}

	const q = `
		INSERT INTO raw_messages (
			source, routing_key, queue, message_type,
			event_id, product_id, sport_id, ts_provider,
			payload, raw_blob
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		RETURNING id, received_at`

	var (
		id [16]byte
		at time.Time
	)
	err := r.pool.QueryRow(ctx, q,
		m.Source, nullable(m.RoutingKey), nullable(m.Queue), m.MessageType,
		nullable(m.EventID), m.ProductID, m.SportID, m.TSProvider,
		m.Payload, m.RawBlob,
	).Scan(&id, &at)
	if err == nil {
		return InsertResult{ID: id, Inserted: true, ReceivedAt: at}, nil
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		// Idempotent collapse: fetch the existing row.
		existing, ferr := r.lookupExisting(ctx, m)
		if ferr != nil {
			return InsertResult{}, fmt.Errorf("storage: lookup existing: %w", ferr)
		}
		return InsertResult{ID: existing.ID, Inserted: false, ReceivedAt: existing.ReceivedAt}, nil
	}
	return InsertResult{}, fmt.Errorf("storage: insert raw_message: %w", err)
}

func (r *RawMessageRepo) lookupExisting(ctx context.Context, m RawMessage) (InsertResult, error) {
	const q = `
		SELECT id, received_at FROM raw_messages
		WHERE source = $1
		  AND message_type = $2
		  AND COALESCE(event_id, '') = COALESCE($3, '')
		  AND COALESCE(ts_provider, 'epoch'::timestamptz) = COALESCE($4, 'epoch'::timestamptz)
		LIMIT 1`
	var out InsertResult
	err := r.pool.QueryRow(ctx, q, m.Source, m.MessageType, nullable(m.EventID), m.TSProvider).
		Scan(&out.ID, &out.ReceivedAt)
	if err != nil {
		return InsertResult{}, err
	}
	return out, nil
}

// CountSince returns the number of raw_messages received since t.
func (r *RawMessageRepo) CountSince(ctx context.Context, t time.Time) (int64, error) {
	var n int64
	err := r.pool.QueryRow(ctx,
		`SELECT count(*) FROM raw_messages WHERE received_at >= $1`, t,
	).Scan(&n)
	return n, err
}

// DeleteOlderThan removes rows with received_at < cutoff and increments
// metrics_counters.retention_deleted by the deleted count. Acceptance #16.
func (r *RawMessageRepo) DeleteOlderThan(ctx context.Context, cutoff time.Time) (int64, error) {
	var deleted int64
	err := withTx(ctx, r.pool, func(tx pgx.Tx) error {
		ct, err := tx.Exec(ctx,
			`DELETE FROM raw_messages WHERE received_at < $1`, cutoff)
		if err != nil {
			return err
		}
		deleted = ct.RowsAffected()
		if deleted > 0 {
			if _, err := tx.Exec(ctx, `
				INSERT INTO metrics_counters (name, value, updated_at)
				VALUES ('retention_deleted', $1, now())
				ON CONFLICT (name) DO UPDATE
				   SET value = metrics_counters.value + EXCLUDED.value,
				       updated_at = now()`, deleted); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return 0, fmt.Errorf("storage: retention delete: %w", err)
	}
	return deleted, nil
}

// Counter reads a metrics_counters row; returns 0 if missing.
func Counter(ctx context.Context, pool *Pool, name string) (int64, error) {
	var v int64
	err := pool.QueryRow(ctx,
		`SELECT COALESCE((SELECT value FROM metrics_counters WHERE name = $1), 0)`,
		name,
	).Scan(&v)
	return v, err
}

func nullable(s string) any {
	if s == "" {
		return nil
	}
	return s
}
