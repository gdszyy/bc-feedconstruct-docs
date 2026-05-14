package odds_test

import (
	"context"
	"sync"

	"github.com/gdszyy/bc-feedconstruct-docs/backend/internal/odds"
)

// fakeRepo is the in-memory Repo implementation shared by all
// odds_test files. It mirrors the markets / outcomes / market_status_history
// tables and protects every map with a single mutex so concurrent
// scenarios stay race-clean.
type fakeRepo struct {
	mu sync.Mutex

	matches  map[int64]bool
	markets  map[marketKey]odds.Market
	outcomes map[outcomeKey]odds.Outcome
	history  []odds.MarketStatusHistoryRow

	marketUpsertCnt  int
	outcomeUpsertCnt int
	historyCnt       int
}

type marketKey struct {
	matchID      int64
	marketTypeID int32
	specifier    string
}

type outcomeKey struct {
	matchID      int64
	marketTypeID int32
	specifier    string
	outcomeID    int32
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{
		matches:  map[int64]bool{},
		markets:  map[marketKey]odds.Market{},
		outcomes: map[outcomeKey]odds.Outcome{},
	}
}

// seedMatch makes the fake report MatchExists=true for the given id.
func (r *fakeRepo) seedMatch(id int64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.matches[id] = true
}

// seedMarket inserts a market row at the given status without going
// through the Handler. Useful for setting up "settled / cancelled"
// preconditions in anti-regression tests.
func (r *fakeRepo) seedMarket(m odds.Market) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.markets[marketKey{m.MatchID, m.MarketTypeID, m.Specifier}] = m
}

func (r *fakeRepo) MatchExists(_ context.Context, id int64) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.matches[id], nil
}

func (r *fakeRepo) GetMarket(_ context.Context, matchID int64, marketTypeID int32, specifier string) (odds.Market, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	m, ok := r.markets[marketKey{matchID, marketTypeID, specifier}]
	return m, ok, nil
}

func (r *fakeRepo) UpsertMarket(_ context.Context, m odds.Market) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.marketUpsertCnt++
	r.markets[marketKey{m.MatchID, m.MarketTypeID, m.Specifier}] = m
	return nil
}

func (r *fakeRepo) UpsertOutcome(_ context.Context, o odds.Outcome) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.outcomeUpsertCnt++
	r.outcomes[outcomeKey{o.MatchID, o.MarketTypeID, o.Specifier, o.OutcomeID}] = o
	return nil
}

func (r *fakeRepo) InsertMarketStatusHistory(_ context.Context, row odds.MarketStatusHistoryRow) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.historyCnt++
	r.history = append(r.history, row)
	return nil
}

func (r *fakeRepo) MarketsForBetStop(_ context.Context, scope odds.BetStopScope) ([]odds.Market, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var out []odds.Market
	for k, m := range r.markets {
		if k.matchID != scope.MatchID {
			continue
		}
		if scope.MarketTypeID != nil && k.marketTypeID != *scope.MarketTypeID {
			continue
		}
		if scope.MarketTypeID != nil && scope.Specifier != "" && k.specifier != scope.Specifier {
			continue
		}
		if scope.MarketTypeID == nil && scope.GroupID != nil {
			if m.GroupID == nil || *m.GroupID != *scope.GroupID {
				continue
			}
		}
		out = append(out, m)
	}
	return out, nil
}

func (r *fakeRepo) snapshot() (markets map[marketKey]odds.Market, outcomes map[outcomeKey]odds.Outcome, hist []odds.MarketStatusHistoryRow) {
	r.mu.Lock()
	defer r.mu.Unlock()
	markets = make(map[marketKey]odds.Market, len(r.markets))
	for k, v := range r.markets {
		markets[k] = v
	}
	outcomes = make(map[outcomeKey]odds.Outcome, len(r.outcomes))
	for k, v := range r.outcomes {
		outcomes[k] = v
	}
	hist = append(hist, r.history...)
	return
}

type captureLogger struct {
	mu     sync.Mutex
	events []odds.AntiRegressionEvent
}

func (c *captureLogger) AntiRegressionBlocked(ev odds.AntiRegressionEvent) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.events = append(c.events, ev)
}

func (c *captureLogger) snapshot() []odds.AntiRegressionEvent {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]odds.AntiRegressionEvent, len(c.events))
	copy(out, c.events)
	return out
}

func ptrInt32(v int32) *int32 { return &v }
func ptrInt64(v int64) *int64 { return &v }
