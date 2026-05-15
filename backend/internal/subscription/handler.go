package subscription

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gdszyy/bc-feedconstruct-docs/backend/internal/catalog"
	"github.com/gdszyy/bc-feedconstruct-docs/backend/internal/feed"
)

// HandleBook applies a BookObject / PartnerBooking delivery. It is the
// canonical Subscribed promotion path:
//
//   - if no row exists, the prior status is treated as Idle and a
//     transition Idle → Subscribed is recorded;
//   - if a Requested row exists, the transition Requested → Subscribed
//     is recorded;
//   - if IsSubscribed is explicitly false on the payload, the handler
//     short circuits to HandleUnbook semantics.
//
// In all cases the subscriptions row is upserted and a
// subscription_events row is written.
func (m *Manager) HandleBook(ctx context.Context, _ feed.MessageType, env feed.Envelope, _ [16]byte) error {
	p, err := parseBookPayload(env.Payload)
	if err != nil {
		return fmt.Errorf("subscription: parse book: %w", err)
	}
	if subscribed, present := p.isSubscribed(); present && !subscribed {
		// FeedConstruct sometimes routes Unbook through the Book topic
		// with IsSubscribed=false. Treat it as a release.
		return m.unbook(ctx, env, p, ReasonUnbookOK)
	}
	matchID, ok := p.matchID(env)
	if !ok {
		return errors.New("subscription: book without matchId")
	}
	return m.book(ctx, matchID, p, env)
}

// HandleUnbook applies an explicit Unbook delivery. Idempotent: a
// duplicate Unbook for a row already in unsubscribed is a no-op.
func (m *Manager) HandleUnbook(ctx context.Context, _ feed.MessageType, env feed.Envelope, _ [16]byte) error {
	p, err := parseBookPayload(env.Payload)
	if err != nil {
		return fmt.Errorf("subscription: parse unbook: %w", err)
	}
	return m.unbook(ctx, env, p, ReasonUnbookOK)
}

func (m *Manager) book(ctx context.Context, matchID int64, p bookPayload, env feed.Envelope) error {
	prev, hadPrev, err := m.Repo.GetSubscription(ctx, matchID)
	if err != nil {
		return fmt.Errorf("subscription: load: %w", err)
	}

	from := StatusUnknown
	if hadPrev {
		from = prev.Status
	}
	// A duplicate Book delivery for an already-subscribed row is a
	// no-op (idempotency). The lastEventId guard below also covers
	// replayed identical envelopes.
	if from == StatusSubscribed {
		return nil
	}
	eventID := p.eventID(env)
	if hadPrev && eventID != "" && prev.LastEventID == eventID && from == StatusSubscribed {
		return nil
	}

	now := m.now()
	next := Subscription{
		MatchID:      matchID,
		Product:      p.product(),
		Status:       StatusSubscribed,
		SubscribedAt: timePtr(now),
		LastEventID:  eventID,
	}
	if hadPrev {
		// Preserve the requested_at audit field so listeners can show
		// "subscribed after X ms".
		next.RequestedAt = prev.RequestedAt
		// If the prior row already carried a product hint we trust it
		// when the new payload omitted IsLive.
		if next.Product == "" && prev.Product != "" {
			next.Product = prev.Product
		}
	} else {
		// First sight: the requested phase happened upstream of us.
		// Stamp requested_at = subscribed_at so the row has a sensible
		// audit window.
		next.RequestedAt = timePtr(now)
	}

	if err := m.Repo.UpsertSubscription(ctx, next); err != nil {
		return fmt.Errorf("subscription: upsert: %w", err)
	}
	if err := m.Repo.InsertEvent(ctx, Event{
		MatchID:    matchID,
		From:       from,
		To:         StatusSubscribed,
		Reason:     ReasonBookOK,
		OccurredAt: now,
	}); err != nil {
		return fmt.Errorf("subscription: event: %w", err)
	}
	m.bookCount.Add(1)
	m.emit(matchID, from, StatusSubscribed, ReasonBookOK)
	return nil
}

func (m *Manager) unbook(ctx context.Context, env feed.Envelope, p bookPayload, reason string) error {
	matchID, ok := p.matchID(env)
	if !ok {
		return errors.New("subscription: unbook without matchId")
	}
	return m.releaseLocked(ctx, matchID, reason)
}

// releaseLocked transitions a subscription to unsubscribed.
// Idempotent on already-released rows.
func (m *Manager) releaseLocked(ctx context.Context, matchID int64, reason string) error {
	prev, hadPrev, err := m.Repo.GetSubscription(ctx, matchID)
	if err != nil {
		return fmt.Errorf("subscription: load: %w", err)
	}
	if !hadPrev {
		// Nothing to release. Record a synthetic event so the audit log
		// still surfaces the upstream signal — but no subscriptions row.
		return nil
	}
	if prev.Status == StatusUnsubscribed || prev.Status == StatusExpired || prev.Status == StatusFailed {
		return nil
	}

	now := m.now()
	next := prev
	next.Status = StatusUnsubscribed
	next.ReleasedAt = timePtr(now)
	next.Reason = reason

	if err := m.Repo.UpsertSubscription(ctx, next); err != nil {
		return fmt.Errorf("subscription: upsert release: %w", err)
	}
	if err := m.Repo.InsertEvent(ctx, Event{
		MatchID:    matchID,
		From:       prev.Status,
		To:         StatusUnsubscribed,
		Reason:     reason,
		OccurredAt: now,
	}); err != nil {
		return fmt.Errorf("subscription: event release: %w", err)
	}
	switch reason {
	case ReasonMatchEnded, ReasonMatchClosed, ReasonMatchCancelled, ReasonMatchAbandoned:
		m.autoReleaseCount.Add(1)
	default:
		m.unbookCount.Add(1)
	}
	m.emit(matchID, prev.Status, StatusUnsubscribed, reason)
	return nil
}

// OnMatchStatusChanged is the catalog observer. It releases any active
// subscription on the given match when the match transitions into a
// terminal lifecycle status.
func (m *Manager) OnMatchStatusChanged(ctx context.Context, matchID int64, _, to catalog.MatchStatus) {
	if !m.AutoUnbookOnTerminal {
		return
	}
	reason, terminal := terminalReason(to)
	if !terminal {
		return
	}
	// Best-effort: errors are logged but not propagated; the catalog
	// transaction has already committed.
	if err := m.releaseLocked(ctx, matchID, reason); err != nil && m.Logger != nil {
		m.Logger.TransitionApplied(matchID, StatusUnknown, StatusFailed,
			fmt.Sprintf("auto_release_failed: %v", err))
	}
}

// CleanupTick scans for subscriptions stuck in requested for longer
// than StuckRequestTimeout and transitions them to failed. Returns the
// number of rows transitioned (handy for tests and metrics).
//
// The caller is responsible for invoking CleanupTick on a cadence
// (cmd/bffd wires it to a time.Ticker).
func (m *Manager) CleanupTick(ctx context.Context) (int, error) {
	now := m.now()
	cutoff := now.Add(-m.timeout())
	stuck, err := m.Repo.ListStuckRequests(ctx, cutoff)
	if err != nil {
		return 0, fmt.Errorf("subscription: list stuck requests: %w", err)
	}
	moved := 0
	for _, s := range stuck {
		next := s
		next.Status = StatusFailed
		next.Reason = ReasonStuckRequest
		next.ReleasedAt = timePtr(now)
		if err := m.Repo.UpsertSubscription(ctx, next); err != nil {
			return moved, fmt.Errorf("subscription: upsert stuck: %w", err)
		}
		if err := m.Repo.InsertEvent(ctx, Event{
			MatchID:    s.MatchID,
			From:       StatusRequested,
			To:         StatusFailed,
			Reason:     ReasonStuckRequest,
			OccurredAt: now,
		}); err != nil {
			return moved, fmt.Errorf("subscription: insert stuck event: %w", err)
		}
		m.stuckExpired.Add(1)
		moved++
		m.emit(s.MatchID, StatusRequested, StatusFailed, ReasonStuckRequest)
		if m.Logger != nil {
			m.Logger.StuckRequestExpired(s.MatchID)
		}
	}
	return moved, nil
}

// RunCleanupLoop drives CleanupTick on the given interval until ctx is
// cancelled. Errors from CleanupTick are not fatal: the loop logs them
// (via Logger.TransitionApplied) and continues.
func (m *Manager) RunCleanupLoop(ctx context.Context, interval time.Duration) error {
	if interval <= 0 {
		interval = m.timeout()
	}
	tick := time.NewTicker(interval)
	defer tick.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-tick.C:
			if _, err := m.CleanupTick(ctx); err != nil && m.Logger != nil {
				m.Logger.TransitionApplied(0, StatusUnknown, StatusUnknown,
					fmt.Sprintf("cleanup_failed: %v", err))
			}
		}
	}
}

func terminalReason(to catalog.MatchStatus) (string, bool) {
	switch to {
	case catalog.StatusEnded:
		return ReasonMatchEnded, true
	case catalog.StatusClosed:
		return ReasonMatchClosed, true
	case catalog.StatusCancelled:
		return ReasonMatchCancelled, true
	}
	return "", false
}

func timePtr(t time.Time) *time.Time {
	v := t
	return &v
}
