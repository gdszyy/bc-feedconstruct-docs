package subscription

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gdszyy/bc-feedconstruct-docs/backend/internal/feed"
)

// HandleBook applies a FC Book delivery (acceptance 13-a). The
// subscription is upserted to status=subscribed and one event row
// records the transition from the previous status (or 'requested' when
// the row did not exist yet — the conceptual previous state in the
// frontend M11 FSM).
func (m *Manager) HandleBook(ctx context.Context, _ feed.MessageType, env feed.Envelope, _ [16]byte) error {
	p, err := parseBook(env.Payload)
	if err != nil {
		return fmt.Errorf("subscription: parse book: %w", err)
	}
	matchID, ok := p.matchID()
	if !ok {
		if env.MatchID != nil {
			matchID = *env.MatchID
		} else {
			return errors.New("subscription: book without matchId")
		}
	}
	now := m.now()
	product := p.product()

	cur, exists, err := m.Repo.GetSubscription(ctx, matchID)
	if err != nil {
		return fmt.Errorf("subscription: get: %w", err)
	}
	from := StatusRequested
	if exists {
		from = cur.Status
		// Preserve product if the row already exists and the new
		// delivery omits it.
		if product == ProductLive && cur.Product != "" {
			product = cur.Product
		}
	}
	if exists && cur.Status == StatusSubscribed {
		// Already subscribed; record nothing.
		return nil
	}

	reason := ReasonBooked
	rec := Subscription{
		MatchID:      matchID,
		Product:      product,
		Status:       StatusSubscribed,
		SubscribedAt: timePtr(now),
		ReleasedAt:   nil,
		Reason:       strPtr(reason),
	}
	if exists {
		rec.RequestedAt = cur.RequestedAt
		rec.LastEventID = cur.LastEventID
	} else {
		rec.RequestedAt = timePtr(now)
	}
	if err := m.Repo.UpsertSubscription(ctx, rec); err != nil {
		return fmt.Errorf("subscription: upsert book: %w", err)
	}
	if err := m.recordEvent(ctx, matchID, from, StatusSubscribed, reason, now); err != nil {
		return err
	}
	m.bookCount.Add(1)

	// Drop any pending release from a prior terminal status — Book wins.
	m.mu.Lock()
	delete(m.pendingReleases, matchID)
	m.mu.Unlock()
	return nil
}

// HandleUnbook applies a FC Unbook delivery. The subscription
// transitions to 'unsubscribed' with reason='unbooked'.
func (m *Manager) HandleUnbook(ctx context.Context, _ feed.MessageType, env feed.Envelope, _ [16]byte) error {
	p, err := parseBook(env.Payload)
	if err != nil {
		return fmt.Errorf("subscription: parse unbook: %w", err)
	}
	matchID, ok := p.matchID()
	if !ok {
		if env.MatchID != nil {
			matchID = *env.MatchID
		} else {
			return errors.New("subscription: unbook without matchId")
		}
	}
	now := m.now()

	cur, exists, err := m.Repo.GetSubscription(ctx, matchID)
	if err != nil {
		return fmt.Errorf("subscription: get: %w", err)
	}
	if !exists {
		// Unbook for a match we never tracked; record an event so the
		// audit log shows the attempt, but skip the upsert (no FK in
		// the schema requires a row).
		return m.recordEvent(ctx, matchID, "", StatusUnsubscribed, ReasonUnbooked, now)
	}
	if cur.Status == StatusUnsubscribed || cur.Status == StatusExpired {
		return nil
	}
	rec := cur
	rec.Status = StatusUnsubscribed
	rec.ReleasedAt = timePtr(now)
	rec.Reason = strPtr(ReasonUnbooked)
	if err := m.Repo.UpsertSubscription(ctx, rec); err != nil {
		return fmt.Errorf("subscription: upsert unbook: %w", err)
	}
	if err := m.recordEvent(ctx, matchID, cur.Status, StatusUnsubscribed, ReasonUnbooked, now); err != nil {
		return err
	}
	m.unbookCount.Add(1)
	m.mu.Lock()
	delete(m.pendingReleases, matchID)
	m.mu.Unlock()
	return nil
}

// HandleFixtureChange listens to match status changes and, when the
// match enters a terminal state, either releases the subscription
// immediately (Grace == 0) or schedules it for release via the pending
// queue (drained by ProcessDueReleases).
func (m *Manager) HandleFixtureChange(ctx context.Context, _ feed.MessageType, env feed.Envelope, _ [16]byte) error {
	p, err := parseFixture(env.Payload)
	if err != nil {
		return fmt.Errorf("subscription: parse fixture_change: %w", err)
	}
	matchID, ok := p.matchID()
	if !ok {
		if env.MatchID != nil {
			matchID = *env.MatchID
		} else {
			return nil // nothing actionable; the catalog handler will FK-fail
		}
	}
	return m.OnMatchStatusChange(ctx, matchID, p.Status)
}

// OnMatchStatusChange is the entrypoint exposed to other modules. It
// inspects the new status and either releases the subscription
// immediately or registers a pending release for the grace window.
func (m *Manager) OnMatchStatusChange(ctx context.Context, matchID int64, status string) error {
	if !isTerminal(status) {
		return nil
	}
	cur, exists, err := m.Repo.GetSubscription(ctx, matchID)
	if err != nil {
		return fmt.Errorf("subscription: get for status change: %w", err)
	}
	if !exists || cur.Status != StatusSubscribed {
		return nil
	}
	if m.Grace <= 0 {
		return m.releaseNow(ctx, cur, ReasonMatchEnded, m.now())
	}
	due := m.now().Add(m.Grace)
	m.mu.Lock()
	m.pendingReleases[matchID] = due
	m.mu.Unlock()
	return nil
}

// ProcessDueReleases drains the pending-release queue, releasing every
// subscription whose grace window has elapsed. Returns the number of
// rows released.
func (m *Manager) ProcessDueReleases(ctx context.Context) (int, error) {
	now := m.now()
	type due struct {
		matchID int64
	}
	m.mu.Lock()
	var ready []due
	for id, t := range m.pendingReleases {
		if !t.After(now) {
			ready = append(ready, due{id})
			delete(m.pendingReleases, id)
		}
	}
	m.mu.Unlock()

	count := 0
	for _, d := range ready {
		cur, exists, err := m.Repo.GetSubscription(ctx, d.matchID)
		if err != nil {
			return count, fmt.Errorf("subscription: get for release: %w", err)
		}
		if !exists || cur.Status != StatusSubscribed {
			continue
		}
		if err := m.releaseNow(ctx, cur, ReasonMatchEnded, now); err != nil {
			return count, err
		}
		count++
	}
	return count, nil
}

// CleanupStuckRequests transitions any subscription stuck in
// 'requested' beyond StuckAfter to 'failed'. Returns the number of rows
// transitioned.
func (m *Manager) CleanupStuckRequests(ctx context.Context) (int, error) {
	now := m.now()
	cutoff := now.Add(-m.stuckAfter())
	rows, err := m.Repo.ListByStatus(ctx, StatusRequested)
	if err != nil {
		return 0, fmt.Errorf("subscription: list requested: %w", err)
	}
	count := 0
	for _, s := range rows {
		if s.RequestedAt == nil || s.RequestedAt.After(cutoff) {
			continue
		}
		next := s
		next.Status = StatusFailed
		next.Reason = strPtr(ReasonStuckRequest)
		next.ReleasedAt = timePtr(now)
		if err := m.Repo.UpsertSubscription(ctx, next); err != nil {
			return count, fmt.Errorf("subscription: upsert stuck: %w", err)
		}
		if err := m.recordEvent(ctx, s.MatchID, StatusRequested, StatusFailed, ReasonStuckRequest, now); err != nil {
			return count, err
		}
		m.stuckFailed.Add(1)
		count++
	}
	return count, nil
}

// MarkRequested is the entrypoint REST handlers will call when a user
// initiates a subscription. It is included now so wave-7 tests can seed
// 'requested' rows without reaching into the Repo.
func (m *Manager) MarkRequested(ctx context.Context, matchID int64, product Product) error {
	now := m.now()
	cur, exists, err := m.Repo.GetSubscription(ctx, matchID)
	if err != nil {
		return fmt.Errorf("subscription: get for request: %w", err)
	}
	if exists && cur.Status == StatusRequested {
		return nil
	}
	rec := Subscription{
		MatchID:     matchID,
		Product:     product,
		Status:      StatusRequested,
		RequestedAt: timePtr(now),
	}
	if err := m.Repo.UpsertSubscription(ctx, rec); err != nil {
		return fmt.Errorf("subscription: upsert requested: %w", err)
	}
	from := Status("")
	if exists {
		from = cur.Status
	}
	return m.recordEvent(ctx, matchID, from, StatusRequested, "", now)
}

func (m *Manager) releaseNow(ctx context.Context, cur Subscription, reason string, now time.Time) error {
	next := cur
	next.Status = StatusUnsubscribed
	next.ReleasedAt = timePtr(now)
	next.Reason = strPtr(reason)
	if err := m.Repo.UpsertSubscription(ctx, next); err != nil {
		return fmt.Errorf("subscription: release upsert: %w", err)
	}
	if err := m.recordEvent(ctx, cur.MatchID, cur.Status, StatusUnsubscribed, reason, now); err != nil {
		return err
	}
	m.autoRelease.Add(1)
	m.mu.Lock()
	delete(m.pendingReleases, cur.MatchID)
	m.mu.Unlock()
	return nil
}

func (m *Manager) recordEvent(ctx context.Context, matchID int64, from Status, to Status, reason string, at time.Time) error {
	ev := Event{
		MatchID:    matchID,
		ToStatus:   to,
		OccurredAt: at,
	}
	if from != "" {
		f := from
		ev.FromStatus = &f
	}
	if reason != "" {
		r := reason
		ev.Reason = &r
	}
	if err := m.Repo.InsertEvent(ctx, ev); err != nil {
		return fmt.Errorf("subscription: insert event: %w", err)
	}
	return nil
}

func timePtr(t time.Time) *time.Time { return &t }
func strPtr(s string) *string        { return &s }
