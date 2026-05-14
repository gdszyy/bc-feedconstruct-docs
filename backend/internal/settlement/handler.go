package settlement

import (
	"context"
	"errors"
	"fmt"

	"github.com/gdszyy/bc-feedconstruct-docs/backend/internal/feed"
)

// HandleBetSettlement applies a bet_settlement delivery (acceptance 7).
// It inserts one settlements row per outcome and transitions the
// targeted markets row to status=settled. Per-outcome certainty is
// preserved verbatim; uncertain (0) and certain (1) deliveries for the
// same outcome are persisted as separate rows whose ordering by
// settled_at lets the snapshot prefer the later certainty.
func (h *Handler) HandleBetSettlement(ctx context.Context, _ feed.MessageType, env feed.Envelope, rawID [16]byte) error {
	p, err := parseBetSettlement(env.Payload)
	if err != nil {
		return fmt.Errorf("settlement: parse bet_settlement: %w", err)
	}
	matchID, ok := p.matchID()
	if !ok {
		if env.MatchID != nil {
			matchID = *env.MatchID
		} else {
			return errors.New("settlement: bet_settlement without matchId")
		}
	}
	exists, err := h.Repo.MatchExists(ctx, matchID)
	if err != nil {
		return fmt.Errorf("settlement: match lookup: %w", err)
	}
	if !exists {
		if h.Logger != nil {
			h.Logger.SkipUnknownMatch(matchID, "bet_settlement")
		}
		return nil
	}

	markets := p.flatten()
	if len(markets) == 0 {
		return nil
	}

	for _, m := range markets {
		marketTypeID, ok := m.marketTypeID()
		if !ok {
			continue
		}
		certainty := CertaintyConfirmed
		if m.Certainty != nil {
			certainty = int16(*m.Certainty)
		} else if p.Certainty != nil {
			certainty = int16(*p.Certainty)
		}

		anyPersisted := false
		for _, o := range m.selections() {
			outcomeID, ok := o.outcomeID()
			if !ok {
				continue
			}
			result := normaliseResult(o.Result, o.ResultCode)
			if result == "" {
				continue
			}
			cert := certainty
			if o.Certainty != nil {
				cert = int16(*o.Certainty)
			}
			rec := Settlement{
				MatchID:        matchID,
				MarketTypeID:   marketTypeID,
				Specifier:      m.Specifier,
				OutcomeID:      outcomeID,
				Result:         result,
				Certainty:      cert,
				VoidFactor:     o.VoidFactor,
				DeadHeatFactor: o.DeadHeatFactor,
				RawMessageID:   rawID,
				SettledAt:      h.now(),
			}
			if _, err := h.Repo.InsertSettlement(ctx, rec); err != nil {
				return fmt.Errorf("settlement: insert: %w", err)
			}
			h.settlementCount.Add(1)
			anyPersisted = true
		}

		if anyPersisted {
			if err := h.Repo.SetMarketStatus(ctx, matchID, marketTypeID, m.Specifier, StatusSettled, rawID); err != nil {
				return fmt.Errorf("settlement: set market settled: %w", err)
			}
		}
	}
	return nil
}

// HandleBetCancel applies a VoidNotification (acceptance 8).
// VoidAction=1 → cancel; VoidAction=2 (unvoid) is routed through
// HandleRollback. Match-level cancels (ObjectType=4 or absent
// marketTypeId) iterate every markets row of the match.
func (h *Handler) HandleBetCancel(ctx context.Context, _ feed.MessageType, env feed.Envelope, rawID [16]byte) error {
	p, err := parseVoidNotification(env.Payload)
	if err != nil {
		return fmt.Errorf("settlement: parse bet_cancel: %w", err)
	}

	// VoidAction=2 reuses this code path through the dispatcher mapping
	// in some FC variants; defer to rollback in that case.
	action := VoidActionVoid
	if p.VoidAction != nil {
		action = int16(*p.VoidAction)
	} else if env.VoidAction != nil {
		action = int16(*env.VoidAction)
	}
	if action == VoidActionUnvoid {
		return h.handleUnvoid(ctx, env, p, rawID)
	}

	matchID, ok := p.matchID()
	if !ok {
		if env.MatchID != nil {
			matchID = *env.MatchID
		} else {
			return errors.New("settlement: bet_cancel without matchId")
		}
	}
	exists, err := h.Repo.MatchExists(ctx, matchID)
	if err != nil {
		return fmt.Errorf("settlement: match lookup: %w", err)
	}
	if !exists {
		if h.Logger != nil {
			h.Logger.SkipUnknownMatch(matchID, "bet_cancel")
		}
		return nil
	}

	now := h.now()
	cancel := Cancel{
		MatchID:      matchID,
		MarketTypeID: p.marketTypeID(),
		Specifier:    p.Specifier,
		VoidReason:   p.reason(),
		VoidAction:   action,
		SupercededBy: p.SupercededBy,
		FromTS:       p.fromTime(),
		ToTS:         p.toTime(),
		RawMessageID: rawID,
		CancelledAt:  now,
	}
	if _, err := h.Repo.InsertCancel(ctx, cancel); err != nil {
		return fmt.Errorf("settlement: insert cancel: %w", err)
	}
	h.cancelCount.Add(1)

	if p.isMatchLevel() {
		markets, err := h.Repo.ListMarketsForMatch(ctx, matchID)
		if err != nil {
			return fmt.Errorf("settlement: list markets: %w", err)
		}
		for _, m := range markets {
			if err := h.Repo.SetMarketStatus(ctx, m.MatchID, m.MarketTypeID, m.Specifier, StatusCancelled, rawID); err != nil {
				return fmt.Errorf("settlement: cancel market: %w", err)
			}
		}
		return nil
	}

	if cancel.MarketTypeID != nil {
		if err := h.Repo.SetMarketStatus(ctx, matchID, *cancel.MarketTypeID, p.Specifier, StatusCancelled, rawID); err != nil {
			return fmt.Errorf("settlement: cancel market: %w", err)
		}
	}
	return nil
}

// HandleRollback applies a rollback (acceptance 9). The message type
// distinguishes settlement rollback from cancel rollback; when absent we
// fall back to the payload 'target' field. Idempotency is provided by
// the unique index (target, target_id, raw_message_id) on rollbacks.
func (h *Handler) HandleRollback(ctx context.Context, msgType feed.MessageType, env feed.Envelope, rawID [16]byte) error {
	p, err := parseRollback(env.Payload)
	if err != nil {
		return fmt.Errorf("settlement: parse rollback: %w", err)
	}
	target := rollbackTarget(msgType, p, env)
	if target == "" {
		return errors.New("settlement: rollback without resolvable target")
	}
	if target == TargetCancel {
		return h.rollbackCancel(ctx, env, p, rawID)
	}
	return h.rollbackSettlement(ctx, env, p, rawID)
}

// handleUnvoid is the route VoidNotification with VoidAction=2 takes.
// It always rolls back the most recent matching cancel.
func (h *Handler) handleUnvoid(ctx context.Context, env feed.Envelope, p voidNotificationPayload, rawID [16]byte) error {
	matchID, ok := p.matchID()
	if !ok {
		if env.MatchID != nil {
			matchID = *env.MatchID
		} else {
			return errors.New("settlement: unvoid without matchId")
		}
	}
	exists, err := h.Repo.MatchExists(ctx, matchID)
	if err != nil {
		return fmt.Errorf("settlement: match lookup: %w", err)
	}
	if !exists {
		return nil
	}
	cancel, ok, err := h.Repo.LatestCancelForScope(ctx, matchID, p.marketTypeID(), p.Specifier)
	if err != nil {
		return fmt.Errorf("settlement: find cancel for unvoid: %w", err)
	}
	if !ok {
		// Nothing to roll back; treat as no-op rather than fail.
		return nil
	}
	return h.applyRollback(ctx, TargetCancel, cancel.ID, matchID, cancel.MarketTypeID, cancel.Specifier, rawID)
}

func (h *Handler) rollbackSettlement(ctx context.Context, env feed.Envelope, p rollbackPayload, rawID [16]byte) error {
	matchID, ok := p.matchID()
	if !ok {
		if env.MatchID != nil {
			matchID = *env.MatchID
		} else {
			return errors.New("settlement: rollback without matchId")
		}
	}
	exists, err := h.Repo.MatchExists(ctx, matchID)
	if err != nil {
		return fmt.Errorf("settlement: match lookup: %w", err)
	}
	if !exists {
		return nil
	}
	mtID := p.marketTypeID()
	if mtID == nil || p.OutcomeID == nil {
		return errors.New("settlement: rollback_settlement requires marketTypeId and outcomeId")
	}
	s, ok, err := h.Repo.LatestSettlementForOutcome(ctx, matchID, *mtID, p.Specifier, *p.OutcomeID)
	if err != nil {
		return fmt.Errorf("settlement: find settlement for rollback: %w", err)
	}
	if !ok {
		return nil
	}
	mt := s.MarketTypeID
	return h.applyRollback(ctx, TargetSettlement, s.ID, matchID, &mt, s.Specifier, rawID)
}

func (h *Handler) rollbackCancel(ctx context.Context, env feed.Envelope, p rollbackPayload, rawID [16]byte) error {
	matchID, ok := p.matchID()
	if !ok {
		if env.MatchID != nil {
			matchID = *env.MatchID
		} else {
			return errors.New("settlement: rollback_cancel without matchId")
		}
	}
	exists, err := h.Repo.MatchExists(ctx, matchID)
	if err != nil {
		return fmt.Errorf("settlement: match lookup: %w", err)
	}
	if !exists {
		return nil
	}
	c, ok, err := h.Repo.LatestCancelForScope(ctx, matchID, p.marketTypeID(), p.Specifier)
	if err != nil {
		return fmt.Errorf("settlement: find cancel for rollback: %w", err)
	}
	if !ok {
		return nil
	}
	return h.applyRollback(ctx, TargetCancel, c.ID, matchID, c.MarketTypeID, c.Specifier, rawID)
}

// applyRollback enforces idempotency, writes the rollbacks row, marks
// the source row rolled_back_at, and reverts the market status. The
// revert is best-effort: when there is no prior history to fall back on
// (e.g. settlement arrived before any active state was recorded) the
// market is left in its current terminal status.
func (h *Handler) applyRollback(ctx context.Context, target RollbackTarget, targetID, matchID int64, marketTypeID *int32, specifier string, rawID [16]byte) error {
	dup, err := h.Repo.HasRollback(ctx, target, targetID, rawID)
	if err != nil {
		return fmt.Errorf("settlement: rollback dedup: %w", err)
	}
	if dup {
		h.duplicateCount.Add(1)
		return nil
	}
	now := h.now()
	if _, err := h.Repo.InsertRollback(ctx, Rollback{
		Target: target, TargetID: targetID, RawMessageID: rawID, AppliedAt: now,
	}); err != nil {
		return fmt.Errorf("settlement: insert rollback: %w", err)
	}
	h.rollbackCount.Add(1)

	switch target {
	case TargetSettlement:
		if err := h.Repo.MarkSettlementRolledBack(ctx, targetID, now); err != nil {
			return fmt.Errorf("settlement: mark settlement rolled_back: %w", err)
		}
	case TargetCancel:
		if err := h.Repo.MarkCancelRolledBack(ctx, targetID, now); err != nil {
			return fmt.Errorf("settlement: mark cancel rolled_back: %w", err)
		}
	}

	if marketTypeID != nil {
		if _, _, err := h.Repo.RevertMarketStatus(ctx, matchID, *marketTypeID, specifier, rawID); err != nil {
			return fmt.Errorf("settlement: revert market: %w", err)
		}
		return nil
	}
	// Match-level rollback: revert every market of the match.
	markets, err := h.Repo.ListMarketsForMatch(ctx, matchID)
	if err != nil {
		return fmt.Errorf("settlement: list markets for revert: %w", err)
	}
	for _, m := range markets {
		if _, _, err := h.Repo.RevertMarketStatus(ctx, m.MatchID, m.MarketTypeID, m.Specifier, rawID); err != nil {
			return fmt.Errorf("settlement: revert market: %w", err)
		}
	}
	return nil
}

// rollbackTarget picks between settlement / cancel using the dispatcher
// message type first, then the payload's explicit target field, then
// VoidAction=2 as a fallback signal for unvoid.
func rollbackTarget(msgType feed.MessageType, p rollbackPayload, env feed.Envelope) RollbackTarget {
	switch msgType {
	case feed.MsgRollbackCancel:
		return TargetCancel
	case feed.MsgRollback:
		// Honor an explicit target if the payload carries one; otherwise
		// default to settlement (FC's rolled_back_bet_settlement uses
		// MsgRollback in our routing key scheme).
		switch p.Target {
		case string(TargetCancel):
			return TargetCancel
		case string(TargetSettlement), "":
			return TargetSettlement
		}
	}
	if env.VoidAction != nil && *env.VoidAction == int(VoidActionUnvoid) {
		return TargetCancel
	}
	if p.VoidAction != nil && *p.VoidAction == int(VoidActionUnvoid) {
		return TargetCancel
	}
	return TargetSettlement
}
