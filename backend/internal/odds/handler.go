package odds

import (
	"context"
	"errors"
	"fmt"

	"github.com/gdszyy/bc-feedconstruct-docs/backend/internal/feed"
)

// HandleOddsChange applies an odds_change delivery. It is safe to call
// concurrently; the underlying Repo serializes upserts per market row.
func (h *Handler) HandleOddsChange(ctx context.Context, _ feed.MessageType, env feed.Envelope, rawID [16]byte) error {
	p, err := parseOddsChange(env.Payload)
	if err != nil {
		return fmt.Errorf("odds: parse odds_change: %w", err)
	}
	matchID, ok := p.matchID()
	if !ok {
		if env.MatchID != nil {
			matchID = *env.MatchID
		} else {
			return errors.New("odds: odds_change without matchId")
		}
	}
	exists, err := h.Repo.MatchExists(ctx, matchID)
	if err != nil {
		return fmt.Errorf("odds: match lookup: %w", err)
	}
	if !exists {
		// Catalog handler will create the match shortly; recovery fills
		// persistent gaps. Skip silently rather than FK-fail.
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
		incoming := normaliseMarketStatus(m.statusString())
		if incoming == StatusUnknown {
			incoming = StatusActive // odds_change without explicit status implies active
		}
		if err := h.applyMarketTransition(ctx, matchID, marketTypeID, m.Specifier, m.GroupID, incoming, rawID); err != nil {
			return err
		}
		for _, o := range m.outcomes() {
			if o.ID == nil {
				continue
			}
			if err := h.Repo.UpsertOutcome(ctx, Outcome{
				MatchID:      matchID,
				MarketTypeID: marketTypeID,
				Specifier:    m.Specifier,
				OutcomeID:    *o.ID,
				Odds:         o.Odds,
				IsActive:     o.active(),
			}); err != nil {
				return fmt.Errorf("odds: upsert outcome %d: %w", *o.ID, err)
			}
		}
	}
	return nil
}

// HandleBetStop applies a bet_stop delivery. It enumerates the targeted
// markets through Repo.MarketsForBetStop and transitions each one whose
// current rank allows the move (acceptance #6 + #12 market-level).
func (h *Handler) HandleBetStop(ctx context.Context, _ feed.MessageType, env feed.Envelope, rawID [16]byte) error {
	p, err := parseBetStop(env.Payload)
	if err != nil {
		return fmt.Errorf("odds: parse bet_stop: %w", err)
	}
	matchID, ok := p.matchID()
	if !ok {
		if env.MatchID != nil {
			matchID = *env.MatchID
		} else {
			return errors.New("odds: bet_stop without matchId")
		}
	}
	exists, err := h.Repo.MatchExists(ctx, matchID)
	if err != nil {
		return fmt.Errorf("odds: bet_stop match lookup: %w", err)
	}
	if !exists {
		return nil
	}

	target := normaliseMarketStatus(p.statusString())
	if target == StatusUnknown {
		target = StatusSuspended
	}

	scope := BetStopScope{
		MatchID:      matchID,
		MarketTypeID: p.marketTypeID(),
		Specifier:    p.Specifier,
		GroupID:      p.GroupID,
	}
	markets, err := h.Repo.MarketsForBetStop(ctx, scope)
	if err != nil {
		return fmt.Errorf("odds: scan markets for bet_stop: %w", err)
	}
	for _, m := range markets {
		if err := h.applyMarketTransition(ctx, m.MatchID, m.MarketTypeID, m.Specifier, m.GroupID, target, rawID); err != nil {
			return err
		}
	}
	return nil
}

// applyMarketTransition is the shared upsert path: look up the current
// status, decide whether to keep or replace it, write the row, and
// append a history entry on actual moves.
func (h *Handler) applyMarketTransition(ctx context.Context, matchID int64, marketTypeID int32, specifier string, groupID *int32, target MarketStatus, rawID [16]byte) error {
	cur, exists, err := h.Repo.GetMarket(ctx, matchID, marketTypeID, specifier)
	if err != nil {
		return fmt.Errorf("odds: get market: %w", err)
	}

	from := MarketStatus("")
	if exists {
		from = cur.Status
	}

	write := target
	if !allowsTransition(from, target) {
		h.emit(AntiRegressionEvent{
			MatchID:      matchID,
			MarketTypeID: marketTypeID,
			Specifier:    specifier,
			From:         from,
			To:           target,
		})
		write = from
	}

	merged := Market{
		MatchID:      matchID,
		MarketTypeID: marketTypeID,
		Specifier:    specifier,
		Status:       write,
		GroupID:      groupID,
	}
	// Preserve existing group_id when the incoming delivery omits it.
	if merged.GroupID == nil && exists {
		merged.GroupID = cur.GroupID
	}
	if err := h.Repo.UpsertMarket(ctx, merged); err != nil {
		return fmt.Errorf("odds: upsert market: %w", err)
	}

	if !exists || from != write {
		if err := h.Repo.InsertMarketStatusHistory(ctx, MarketStatusHistoryRow{
			MatchID:      matchID,
			MarketTypeID: marketTypeID,
			Specifier:    specifier,
			From:         from,
			To:           write,
			RawMessageID: rawID,
		}); err != nil {
			return fmt.Errorf("odds: history: %w", err)
		}
	}
	return nil
}
