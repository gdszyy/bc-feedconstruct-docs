package settlement_test

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/gdszyy/bc-feedconstruct-docs/backend/internal/settlement"
)

// marketKey indexes the in-memory markets snapshot by primary key
// (match_id, market_type_id, specifier).
type marketKey struct {
	matchID      int64
	marketTypeID int32
	specifier    string
}

type fakeMarket struct {
	current settlement.MarketStatus
	// prior is the operational status the market held before its current
	// terminal status was assigned. RevertMarketStatus restores it.
	prior settlement.MarketStatus
}

// fakeRepo mirrors the settlements / cancels / rollbacks tables in
// memory, plus the markets fields the handler reads/writes. Shared by
// every *_test.go in this package.
type fakeRepo struct {
	mu sync.Mutex

	matches map[int64]bool
	markets map[marketKey]fakeMarket

	settlementSeq int64
	cancelSeq     int64
	rollbackSeq   int64

	settlements []settlement.Settlement
	cancels     []settlement.Cancel
	rollbacks   []settlement.Rollback

	statusChanges []statusChange
}

type statusChange struct {
	key      marketKey
	from, to settlement.MarketStatus
	rawID    [16]byte
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{
		matches: map[int64]bool{},
		markets: map[marketKey]fakeMarket{},
	}
}

func (r *fakeRepo) seedMatch(id int64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.matches[id] = true
}

// seedMarket installs a market with both its current status and the
// status to revert to (prior). priorStatus="" means the market has no
// recoverable predecessor.
func (r *fakeRepo) seedMarket(matchID int64, marketTypeID int32, specifier string, current, prior settlement.MarketStatus) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.markets[marketKey{matchID, marketTypeID, specifier}] = fakeMarket{current: current, prior: prior}
}

func (r *fakeRepo) MatchExists(_ context.Context, id int64) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.matches[id], nil
}

func (r *fakeRepo) InsertSettlement(_ context.Context, s settlement.Settlement) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.settlementSeq++
	s.ID = r.settlementSeq
	r.settlements = append(r.settlements, s)
	return s.ID, nil
}

// LatestSettlementForOutcome returns the most recent settlements row
// for the outcome (by settled_at) regardless of rolled-back status, so
// the handler can resolve the target row a rollback message refers to
// and then use HasRollback as the idempotency gate.
func (r *fakeRepo) LatestSettlementForOutcome(_ context.Context, matchID int64, marketTypeID int32, specifier string, outcomeID int32) (settlement.Settlement, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	matches := make([]settlement.Settlement, 0)
	for _, s := range r.settlements {
		if s.MatchID == matchID && s.MarketTypeID == marketTypeID && s.Specifier == specifier && s.OutcomeID == outcomeID {
			matches = append(matches, s)
		}
	}
	if len(matches) == 0 {
		return settlement.Settlement{}, false, nil
	}
	sort.SliceStable(matches, func(i, j int) bool { return matches[i].SettledAt.Before(matches[j].SettledAt) })
	return matches[len(matches)-1], true, nil
}

func (r *fakeRepo) MarkSettlementRolledBack(_ context.Context, id int64, at time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.settlements {
		if r.settlements[i].ID == id {
			t := at
			r.settlements[i].RolledBackAt = &t
			return nil
		}
	}
	return nil
}

func (r *fakeRepo) InsertCancel(_ context.Context, c settlement.Cancel) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.cancelSeq++
	c.ID = r.cancelSeq
	r.cancels = append(r.cancels, c)
	return c.ID, nil
}

// LatestCancelForScope returns the most recent cancel row for the
// scope, regardless of rolled-back status. Same rationale as
// LatestSettlementForOutcome.
func (r *fakeRepo) LatestCancelForScope(_ context.Context, matchID int64, marketTypeID *int32, specifier string) (settlement.Cancel, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	matches := make([]settlement.Cancel, 0)
	for _, c := range r.cancels {
		if c.MatchID != matchID {
			continue
		}
		// Match scope on (marketTypeID, specifier) when caller specifies one.
		if marketTypeID != nil {
			if c.MarketTypeID == nil || *c.MarketTypeID != *marketTypeID {
				continue
			}
			if specifier != "" && c.Specifier != specifier {
				continue
			}
		} else {
			// Match-level scope: prefer the same shape (match-level cancel).
			if c.MarketTypeID != nil {
				continue
			}
		}
		matches = append(matches, c)
	}
	if len(matches) == 0 {
		return settlement.Cancel{}, false, nil
	}
	sort.SliceStable(matches, func(i, j int) bool { return matches[i].CancelledAt.Before(matches[j].CancelledAt) })
	return matches[len(matches)-1], true, nil
}

func (r *fakeRepo) MarkCancelRolledBack(_ context.Context, id int64, at time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.cancels {
		if r.cancels[i].ID == id {
			t := at
			r.cancels[i].RolledBackAt = &t
			return nil
		}
	}
	return nil
}

func (r *fakeRepo) HasRollback(_ context.Context, target settlement.RollbackTarget, targetID int64, rawID [16]byte) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, rb := range r.rollbacks {
		if rb.Target == target && rb.TargetID == targetID && rb.RawMessageID == rawID {
			return true, nil
		}
	}
	return false, nil
}

func (r *fakeRepo) InsertRollback(_ context.Context, rb settlement.Rollback) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.rollbackSeq++
	rb.ID = r.rollbackSeq
	r.rollbacks = append(r.rollbacks, rb)
	return rb.ID, nil
}

func (r *fakeRepo) GetMarket(_ context.Context, matchID int64, marketTypeID int32, specifier string) (settlement.MarketRef, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	m, ok := r.markets[marketKey{matchID, marketTypeID, specifier}]
	if !ok {
		return settlement.MarketRef{}, false, nil
	}
	return settlement.MarketRef{MatchID: matchID, MarketTypeID: marketTypeID, Specifier: specifier, Status: m.current}, true, nil
}

func (r *fakeRepo) ListMarketsForMatch(_ context.Context, matchID int64) ([]settlement.MarketRef, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]settlement.MarketRef, 0)
	for k, m := range r.markets {
		if k.matchID != matchID {
			continue
		}
		out = append(out, settlement.MarketRef{
			MatchID: k.matchID, MarketTypeID: k.marketTypeID, Specifier: k.specifier, Status: m.current,
		})
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].MarketTypeID != out[j].MarketTypeID {
			return out[i].MarketTypeID < out[j].MarketTypeID
		}
		return out[i].Specifier < out[j].Specifier
	})
	return out, nil
}

func (r *fakeRepo) SetMarketStatus(_ context.Context, matchID int64, marketTypeID int32, specifier string, to settlement.MarketStatus, rawID [16]byte) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	k := marketKey{matchID, marketTypeID, specifier}
	cur, ok := r.markets[k]
	if !ok {
		// Allow tests to skip seeding markets and just verify "no transition".
		cur = fakeMarket{current: settlement.StatusUnknown}
	}
	// Remember the operational status the market is leaving so a later
	// RevertMarketStatus can restore it.
	prior := cur.prior
	if !isTerminal(cur.current) && cur.current != settlement.StatusUnknown {
		prior = cur.current
	}
	r.markets[k] = fakeMarket{current: to, prior: prior}
	r.statusChanges = append(r.statusChanges, statusChange{key: k, from: cur.current, to: to, rawID: rawID})
	return nil
}

func (r *fakeRepo) RevertMarketStatus(_ context.Context, matchID int64, marketTypeID int32, specifier string, rawID [16]byte) (settlement.MarketStatus, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	k := marketKey{matchID, marketTypeID, specifier}
	cur, ok := r.markets[k]
	if !ok || cur.prior == settlement.StatusUnknown {
		return settlement.StatusUnknown, false, nil
	}
	target := cur.prior
	r.statusChanges = append(r.statusChanges, statusChange{key: k, from: cur.current, to: target, rawID: rawID})
	r.markets[k] = fakeMarket{current: target, prior: settlement.StatusUnknown}
	return target, true, nil
}

func isTerminal(s settlement.MarketStatus) bool {
	switch s {
	case settlement.StatusSettled, settlement.StatusCancelled, settlement.StatusHandedOver:
		return true
	}
	return false
}

// snapshot returns deep copies of the persisted state for assertions.
func (r *fakeRepo) snapshot() ([]settlement.Settlement, []settlement.Cancel, []settlement.Rollback, map[marketKey]fakeMarket) {
	r.mu.Lock()
	defer r.mu.Unlock()
	settlements := append([]settlement.Settlement(nil), r.settlements...)
	cancels := append([]settlement.Cancel(nil), r.cancels...)
	rollbacks := append([]settlement.Rollback(nil), r.rollbacks...)
	markets := make(map[marketKey]fakeMarket, len(r.markets))
	for k, v := range r.markets {
		markets[k] = v
	}
	return settlements, cancels, rollbacks, markets
}

