// Package recovery coordinates startup, product-level, event-level,
// stateful-message and fixture-change recoveries with rate-limit aware retry.
// Maps to upload-guideline 业务域 "恢复补偿" (M10).
package recovery

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/gdszyy/bc-feedconstruct-docs/backend/internal/webapi"
)

// WebAPI is the slice of webapi.Client used by the coordinator. Keeping
// it narrow keeps the unit tests free of HTTP plumbing.
type WebAPI interface {
	DataSnapshot(ctx context.Context, isLive bool, getChangesFrom int) ([]json.RawMessage, error)
	MatchByID(ctx context.Context, matchID int64) (json.RawMessage, error)
}

// Jobs persists the lifecycle of one recovery_jobs row.
type Jobs interface {
	Record(ctx context.Context, scope, product string, matchID int64) (int64, error)
	Finalize(ctx context.Context, id int64, status string, detail map[string]any) error
}

// LastMessageAtFn returns the wall-clock timestamp of the most recently
// persisted RMQ message. The boolean is false when raw_messages is
// empty (fresh deployment).
type LastMessageAtFn func(ctx context.Context) (time.Time, bool, error)

// Coordinator orchestrates the recovery flows defined in
// docs/01_data_feed/rmq-web-api/036_integrationnotes.md.
type Coordinator struct {
	API           WebAPI
	Jobs          Jobs
	LastMessageAt LastMessageAtFn
	Now           func() time.Time

	// SafetyWindow is added to the observed outage duration when
	// computing getChangesFrom, so an in-flight delivery at the cutoff
	// is reprocessed.
	SafetyWindow time.Duration

	// BackoffBase / BackoffMax / MaxAttempts govern the rate-limit
	// retry envelope. Backoff doubles each attempt, capped at
	// BackoffMax; after MaxAttempts the call is abandoned.
	BackoffBase time.Duration
	BackoffMax  time.Duration
	MaxAttempts int

	// Sleep is the unit of waiting; tests substitute a recorder.
	Sleep func(time.Duration)
}

// shortOutageThreshold matches the FeedConstruct guideline that
// outages under one hour use a delta snapshot via getChangesFrom.
const shortOutageThreshold = time.Hour

// StartupRecovery is invoked once on boot. It chooses between a full
// snapshot (cold start or long outage) and a delta snapshot.
func (c *Coordinator) StartupRecovery(ctx context.Context) error {
	getChangesFrom, err := c.startupGetChangesFrom(ctx)
	if err != nil {
		return err
	}
	for _, isLive := range []bool{true, false} {
		product := "live"
		if !isLive {
			product = "prematch"
		}
		if err := c.runJob(ctx, "startup", product, 0, func(ctx context.Context) error {
			_, err := c.API.DataSnapshot(ctx, isLive, getChangesFrom)
			return err
		}); err != nil {
			return err
		}
	}
	return nil
}

// EventRecovery refreshes a single match via /MatchById.
func (c *Coordinator) EventRecovery(ctx context.Context, matchID int64) error {
	return c.runJob(ctx, "event", "", matchID, func(ctx context.Context) error {
		_, err := c.API.MatchByID(ctx, matchID)
		return err
	})
}

func (c *Coordinator) startupGetChangesFrom(ctx context.Context) (int, error) {
	if c.LastMessageAt == nil {
		return 0, nil
	}
	last, ok, err := c.LastMessageAt(ctx)
	if err != nil {
		return 0, fmt.Errorf("recovery: read last_message_at: %w", err)
	}
	if !ok || last.IsZero() {
		return 0, nil
	}
	outage := c.now().Sub(last)
	if outage <= 0 || outage >= shortOutageThreshold {
		return 0, nil
	}
	// Round up so a 14m45s outage asks for 15 minutes, then add the
	// safety window to cover messages in flight.
	minutes := int((outage + 59*time.Second) / time.Minute)
	minutes += int(c.SafetyWindow / time.Minute)
	if minutes <= 0 {
		minutes = 1
	}
	return minutes, nil
}

// runJob is the unit of "record → attempt with retry → finalize".
// The attempt is retried only for webapi.RateLimitError; other errors
// finalize the job as failed and propagate.
func (c *Coordinator) runJob(ctx context.Context, scope, product string, matchID int64, attempt func(context.Context) error) error {
	id, err := c.Jobs.Record(ctx, scope, product, matchID)
	if err != nil {
		return fmt.Errorf("recovery: record job: %w", err)
	}

	maxAttempts := c.MaxAttempts
	if maxAttempts <= 0 {
		maxAttempts = 1
	}

	var lastErr error
	for i := 0; i < maxAttempts; i++ {
		lastErr = attempt(ctx)
		if lastErr == nil {
			_ = c.Jobs.Finalize(ctx, id, "success", nil)
			return nil
		}
		if !webapi.IsRateLimited(lastErr) {
			_ = c.Jobs.Finalize(ctx, id, "failed", map[string]any{"error": lastErr.Error()})
			return lastErr
		}
		// Mark the attempt as rate_limited; we may try again.
		_ = c.Jobs.Finalize(ctx, id, "rate_limited", map[string]any{"attempt": i + 1})
		if i == maxAttempts-1 {
			break
		}
		wait := c.backoff(i)
		if rl := rateLimitRetryAfter(lastErr); rl > wait {
			wait = rl
		}
		c.sleep(wait)
	}
	return lastErr
}

func (c *Coordinator) backoff(attempt int) time.Duration {
	base := c.BackoffBase
	if base <= 0 {
		base = 250 * time.Millisecond
	}
	max := c.BackoffMax
	if max <= 0 {
		max = 30 * time.Second
	}
	d := base
	for i := 0; i < attempt; i++ {
		d *= 2
		if d >= max {
			return max
		}
	}
	return d
}

func rateLimitRetryAfter(err error) time.Duration {
	var rl *webapi.RateLimitError
	if !errors.As(err, &rl) {
		return 0
	}
	if rl.RetryAfterSeconds <= 0 {
		return 0
	}
	return time.Duration(rl.RetryAfterSeconds) * time.Second
}

func (c *Coordinator) now() time.Time {
	if c.Now != nil {
		return c.Now()
	}
	return time.Now()
}

func (c *Coordinator) sleep(d time.Duration) {
	if c.Sleep != nil {
		c.Sleep(d)
		return
	}
	time.Sleep(d)
}
