package recovery

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/gdszyy/bc-feedconstruct-docs/backend/internal/storage"
	"github.com/gdszyy/bc-feedconstruct-docs/backend/internal/webapi"
)

// Scope is the recovery_jobs.scope value.
type Scope string

const (
	ScopeStartup        Scope = "startup"
	ScopeProduct        Scope = "product"
	ScopeEvent          Scope = "event"
	ScopeStateful       Scope = "stateful"
	ScopeFixtureChange  Scope = "fixture_change"
)

// Status mirrors recovery_jobs.status.
type Status string

const (
	StatusQueued      Status = "queued"
	StatusRunning     Status = "running"
	StatusSuccess     Status = "success"
	StatusFailed      Status = "failed"
	StatusRateLimited Status = "rate_limited"
)

// SnapshotAPI is the slice of webapi.Client used by the coordinator.
// Tests substitute a stub; production wires *webapi.Client.
type SnapshotAPI interface {
	DataSnapshot(ctx context.Context, isLive bool, changesFrom *time.Time) (webapi.SnapshotResult, error)
	GetMatchByID(ctx context.Context, matchID int64) ([]byte, error)
}

// SnapshotIngester is the hook the coordinator calls after a successful
// API call to integrate snapshot bytes into the BFF (rows -> raw_messages
// via the same Processor path). The default is a no-op; later waves wire
// a concrete ingester that fans messages out to the feed pipeline.
type SnapshotIngester interface {
	IngestSnapshot(ctx context.Context, scope Scope, product string, matchID *int64, body []byte) (int, error)
}

// NopIngester accepts every snapshot and reports zero deliveries.
type NopIngester struct{}

func (NopIngester) IngestSnapshot(context.Context, Scope, string, *int64, []byte) (int, error) {
	return 0, nil
}

// Options tunes the Coordinator.
type Options struct {
	BackoffBase time.Duration // first retry delay, default 2s
	BackoffMax  time.Duration // cap, default 5m
	MaxAttempts int           // give up after this many failed attempts; default 6
	Now         func() time.Time
}

// Coordinator owns recovery_jobs scheduling and execution.
type Coordinator struct {
	pool *storage.Pool
	api  SnapshotAPI
	ing  SnapshotIngester
	opt  Options
}

// New returns a Coordinator. ing may be nil; defaults to NopIngester.
func New(pool *storage.Pool, api SnapshotAPI, ing SnapshotIngester, opt Options) *Coordinator {
	if ing == nil {
		ing = NopIngester{}
	}
	if opt.BackoffBase == 0 {
		opt.BackoffBase = 2 * time.Second
	}
	if opt.BackoffMax == 0 {
		opt.BackoffMax = 5 * time.Minute
	}
	if opt.MaxAttempts == 0 {
		opt.MaxAttempts = 6
	}
	if opt.Now == nil {
		opt.Now = time.Now
	}
	return &Coordinator{pool: pool, api: api, ing: ing, opt: opt}
}

// Request describes a recovery to enqueue.
type Request struct {
	Scope       Scope
	Product     string         // "live" / "prematch" / "" for event-scope
	MatchID     *int64
	ChangesFrom *time.Time     // optional; used by product scope
	Detail      map[string]any // free-form metadata persisted to recovery_jobs.detail
}

// Schedule inserts a queued job and returns its id.
func (c *Coordinator) Schedule(ctx context.Context, req Request) (int64, error) {
	if err := req.validate(); err != nil {
		return 0, err
	}
	detail := req.Detail
	if detail == nil {
		detail = map[string]any{}
	}
	if req.ChangesFrom != nil {
		detail["changes_from"] = req.ChangesFrom.UTC().Format(time.RFC3339)
	}
	dj, err := json.Marshal(detail)
	if err != nil {
		return 0, fmt.Errorf("recovery: marshal detail: %w", err)
	}

	var id int64
	err = c.pool.QueryRow(ctx, `
		INSERT INTO recovery_jobs (scope, product, match_id, status, detail)
		VALUES ($1, $2, $3, 'queued', $4::jsonb)
		RETURNING id`,
		string(req.Scope), nullableString(req.Product), req.MatchID, dj,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("recovery: insert job: %w", err)
	}
	return id, nil
}

// RunOnce picks one queued job whose next_retry_at is due (or NULL) and
// processes it inside a transaction so concurrent coordinators don't pick
// the same row. Returns (true, nil) if a job was processed; (false, nil)
// when the queue is empty.
func (c *Coordinator) RunOnce(ctx context.Context) (bool, error) {
	tx, err := c.pool.Begin(ctx)
	if err != nil {
		return false, fmt.Errorf("recovery: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var job jobRow
	row := tx.QueryRow(ctx, `
		SELECT id, scope, COALESCE(product, ''), match_id, attempt, COALESCE(detail, '{}'::jsonb)
		FROM recovery_jobs
		WHERE status = 'queued'
		  AND (next_retry_at IS NULL OR next_retry_at <= now())
		ORDER BY requested_at ASC
		FOR UPDATE SKIP LOCKED
		LIMIT 1`)
	err = row.Scan(&job.ID, &job.Scope, &job.Product, &job.MatchID, &job.Attempt, &job.Detail)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, tx.Commit(ctx)
	}
	if err != nil {
		return false, fmt.Errorf("recovery: scan next job: %w", err)
	}
	if _, err := tx.Exec(ctx,
		`UPDATE recovery_jobs SET status='running', started_at=now(), attempt=attempt+1 WHERE id=$1`,
		job.ID); err != nil {
		return false, fmt.Errorf("recovery: mark running: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return false, fmt.Errorf("recovery: commit running: %w", err)
	}

	ingested, runErr := c.execute(ctx, job)

	if runErr == nil {
		if err := c.markSuccess(ctx, job.ID, ingested); err != nil {
			return true, err
		}
		return true, nil
	}

	// Classify the error.
	var rlErr *webapi.RateLimitedError
	switch {
	case errors.As(runErr, &rlErr):
		return true, c.scheduleRetry(ctx, job.ID, job.Attempt+1, StatusRateLimited, rlErr.RetryAfter, runErr.Error())
	default:
		if int(job.Attempt+1) >= c.opt.MaxAttempts {
			return true, c.markFailed(ctx, job.ID, runErr.Error())
		}
		return true, c.scheduleRetry(ctx, job.ID, job.Attempt+1, StatusQueued, 0, runErr.Error())
	}
}

type jobRow struct {
	ID      int64
	Scope   string
	Product string
	MatchID *int64
	Attempt int16
	Detail  []byte
}

// execute dispatches by scope. Returns the count reported by the ingester.
func (c *Coordinator) execute(ctx context.Context, j jobRow) (int, error) {
	switch Scope(j.Scope) {
	case ScopeStartup:
		// Full snapshot for both products.
		ingested := 0
		for _, isLive := range []bool{true, false} {
			res, err := c.api.DataSnapshot(ctx, isLive, nil)
			if err != nil {
				return ingested, fmt.Errorf("startup snapshot (isLive=%t): %w", isLive, err)
			}
			n, err := c.ing.IngestSnapshot(ctx, ScopeStartup, productName(isLive), nil, res.BodyJSON)
			if err != nil {
				return ingested, fmt.Errorf("startup ingest (isLive=%t): %w", isLive, err)
			}
			ingested += n
		}
		return ingested, nil

	case ScopeProduct:
		isLive := j.Product == "live"
		var changesFrom *time.Time
		if t, ok := parseDetailTime(j.Detail, "changes_from"); ok {
			changesFrom = &t
		}
		res, err := c.api.DataSnapshot(ctx, isLive, changesFrom)
		if err != nil {
			return 0, err
		}
		return c.ing.IngestSnapshot(ctx, ScopeProduct, j.Product, nil, res.BodyJSON)

	case ScopeEvent:
		if j.MatchID == nil {
			return 0, errors.New("recovery: event scope requires match_id")
		}
		body, err := c.api.GetMatchByID(ctx, *j.MatchID)
		if err != nil {
			return 0, err
		}
		return c.ing.IngestSnapshot(ctx, ScopeEvent, j.Product, j.MatchID, body)

	case ScopeStateful:
		// Stateful messages are re-derived by combining a product-level
		// incremental snapshot with downstream handler replay; for this
		// wave we treat it as a product snapshot with a configurable window.
		isLive := j.Product != "prematch"
		var changesFrom *time.Time
		if t, ok := parseDetailTime(j.Detail, "changes_from"); ok {
			changesFrom = &t
		}
		res, err := c.api.DataSnapshot(ctx, isLive, changesFrom)
		if err != nil {
			return 0, err
		}
		return c.ing.IngestSnapshot(ctx, ScopeStateful, j.Product, nil, res.BodyJSON)

	case ScopeFixtureChange:
		if j.MatchID == nil {
			return 0, errors.New("recovery: fixture_change scope requires match_id")
		}
		body, err := c.api.GetMatchByID(ctx, *j.MatchID)
		if err != nil {
			return 0, err
		}
		return c.ing.IngestSnapshot(ctx, ScopeFixtureChange, "", j.MatchID, body)
	}
	return 0, fmt.Errorf("recovery: unknown scope %q", j.Scope)
}

func (c *Coordinator) markSuccess(ctx context.Context, id int64, ingested int) error {
	detail := map[string]any{"ingested": ingested, "finished_at": c.opt.Now().UTC().Format(time.RFC3339)}
	dj, _ := json.Marshal(detail)
	_, err := c.pool.Exec(ctx, `
		UPDATE recovery_jobs SET status='success', finished_at=now(),
		       detail = COALESCE(detail, '{}'::jsonb) || $2::jsonb
		WHERE id=$1`, id, dj)
	if err != nil {
		return fmt.Errorf("recovery: mark success: %w", err)
	}
	return nil
}

func (c *Coordinator) markFailed(ctx context.Context, id int64, reason string) error {
	dj, _ := json.Marshal(map[string]any{"error": reason})
	_, err := c.pool.Exec(ctx, `
		UPDATE recovery_jobs SET status='failed', finished_at=now(),
		       detail = COALESCE(detail, '{}'::jsonb) || $2::jsonb
		WHERE id=$1`, id, dj)
	if err != nil {
		return fmt.Errorf("recovery: mark failed: %w", err)
	}
	return nil
}

func (c *Coordinator) scheduleRetry(ctx context.Context, id int64, attempt int16, status Status, override time.Duration, reason string) error {
	delay := override
	if delay <= 0 {
		delay = c.backoffFor(attempt)
	}
	dj, _ := json.Marshal(map[string]any{"last_error": reason, "delay_seconds": int(delay.Seconds())})
	next := c.opt.Now().Add(delay)
	_, err := c.pool.Exec(ctx, `
		UPDATE recovery_jobs
		   SET status=$2, next_retry_at=$3,
		       detail = COALESCE(detail, '{}'::jsonb) || $4::jsonb
		 WHERE id=$1`, id, string(status), next, dj)
	if err != nil {
		return fmt.Errorf("recovery: schedule retry: %w", err)
	}
	return nil
}

func (c *Coordinator) backoffFor(attempt int16) time.Duration {
	d := c.opt.BackoffBase
	for i := int16(1); i < attempt && d < c.opt.BackoffMax; i++ {
		d *= 2
	}
	if d > c.opt.BackoffMax {
		d = c.opt.BackoffMax
	}
	return d
}

// Schedule helpers ----------------------------------------------------------

// ScheduleStartup enqueues a full snapshot for both products.
func (c *Coordinator) ScheduleStartup(ctx context.Context) (int64, error) {
	return c.Schedule(ctx, Request{Scope: ScopeStartup})
}

// ScheduleProduct enqueues a product-level snapshot, with an optional
// changesFrom window (FC docs: < 1h outage uses incremental fetch).
func (c *Coordinator) ScheduleProduct(ctx context.Context, product string, changesFrom *time.Time) (int64, error) {
	return c.Schedule(ctx, Request{Scope: ScopeProduct, Product: product, ChangesFrom: changesFrom})
}

// ScheduleEvent enqueues an event-level GetMatchByID re-fetch.
func (c *Coordinator) ScheduleEvent(ctx context.Context, matchID int64) (int64, error) {
	return c.Schedule(ctx, Request{Scope: ScopeEvent, MatchID: &matchID})
}

// ScheduleStateful enqueues a stateful-window recovery.
func (c *Coordinator) ScheduleStateful(ctx context.Context, product string, changesFrom time.Time) (int64, error) {
	cf := changesFrom
	return c.Schedule(ctx, Request{Scope: ScopeStateful, Product: product, ChangesFrom: &cf})
}

// ScheduleFixtureChange enqueues a fixture_change re-pull.
func (c *Coordinator) ScheduleFixtureChange(ctx context.Context, matchID int64) (int64, error) {
	return c.Schedule(ctx, Request{Scope: ScopeFixtureChange, MatchID: &matchID})
}

// Helpers ------------------------------------------------------------------

func (r *Request) validate() error {
	if r.Scope == "" {
		return errors.New("recovery: scope required")
	}
	switch r.Scope {
	case ScopeProduct, ScopeStateful:
		if r.Product != "live" && r.Product != "prematch" {
			return fmt.Errorf("recovery: product scope needs Product=live|prematch, got %q", r.Product)
		}
	case ScopeEvent, ScopeFixtureChange:
		if r.MatchID == nil {
			return errors.New("recovery: event/fixture_change scope needs MatchID")
		}
	}
	return nil
}

func productName(isLive bool) string {
	if isLive {
		return "live"
	}
	return "prematch"
}

func nullableString(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func parseDetailTime(detail []byte, key string) (time.Time, bool) {
	var m map[string]any
	if err := json.Unmarshal(detail, &m); err != nil {
		return time.Time{}, false
	}
	raw, ok := m[key].(string)
	if !ok || raw == "" {
		return time.Time{}, false
	}
	t, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return time.Time{}, false
	}
	return t, true
}
