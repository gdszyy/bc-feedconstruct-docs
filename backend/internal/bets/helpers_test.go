package bets_test

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"sync/atomic"

	"github.com/gdszyy/bc-feedconstruct-docs/backend/internal/bets"
)

// fakeRepo is an in-memory Repo for unit tests.
type fakeRepo struct {
	mu          sync.Mutex
	byID        map[string]*bets.Bet
	idemKey     map[string]string // userID|key -> betID
	transitions map[string][]bets.Transition
	autoTransID int64
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{
		byID:        map[string]*bets.Bet{},
		idemKey:     map[string]string{},
		transitions: map[string][]bets.Transition{},
	}
}

func idemRefKey(userID, key string) string { return userID + "|" + key }

func (r *fakeRepo) FindByIdempotencyKey(_ context.Context, userID, key string) (*bets.Bet, bool, error) {
	r.mu.Lock()
	id, ok := r.idemKey[idemRefKey(userID, key)]
	r.mu.Unlock()
	if !ok {
		return nil, false, nil
	}
	return r.GetByID(context.Background(), id)
}

func (r *fakeRepo) CreatePending(_ context.Context, b *bets.Bet, initial bets.Transition) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.idemKey[idemRefKey(b.UserID, b.IdempotencyKey)]; exists {
		return fmt.Errorf("fakeRepo: duplicate idempotency key %s/%s", b.UserID, b.IdempotencyKey)
	}
	r.autoTransID++
	initial.ID = r.autoTransID
	b.Transitions = []bets.Transition{initial}
	stored := *b
	r.byID[b.ID] = &stored
	r.idemKey[idemRefKey(b.UserID, b.IdempotencyKey)] = b.ID
	r.transitions[b.ID] = []bets.Transition{initial}
	return nil
}

func (r *fakeRepo) GetByID(_ context.Context, betID string) (*bets.Bet, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	b, ok := r.byID[betID]
	if !ok {
		return nil, false, nil
	}
	clone := *b
	clone.Selections = append([]bets.Selection(nil), b.Selections...)
	clone.Transitions = append([]bets.Transition(nil), r.transitions[betID]...)
	return &clone, true, nil
}

func (r *fakeRepo) List(_ context.Context, f bets.ListFilter) ([]*bets.Bet, error) {
	r.mu.Lock()
	want := map[bets.State]bool{}
	for _, s := range f.States {
		want[s] = true
	}
	var out []*bets.Bet
	for _, b := range r.byID {
		if b.UserID != f.UserID {
			continue
		}
		if len(want) > 0 && !want[b.State] {
			continue
		}
		clone := *b
		clone.Selections = append([]bets.Selection(nil), b.Selections...)
		clone.Transitions = append([]bets.Transition(nil), r.transitions[b.ID]...)
		out = append(out, &clone)
	}
	r.mu.Unlock()
	sort.Slice(out, func(i, j int) bool { return out[i].PlacedAt.After(out[j].PlacedAt) })
	return out, nil
}

func (r *fakeRepo) AppendTransition(_ context.Context, betID string, t bets.Transition, p *bets.Payout) (bets.Transition, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	b, ok := r.byID[betID]
	if !ok {
		return bets.Transition{}, false, fmt.Errorf("fakeRepo: bet %s not found", betID)
	}
	if t.EventID != "" {
		for _, ex := range r.transitions[betID] {
			if ex.EventID == t.EventID {
				return bets.Transition{}, false, nil // duplicate event
			}
		}
	}
	r.autoTransID++
	t.ID = r.autoTransID
	r.transitions[betID] = append(r.transitions[betID], t)
	b.State = t.To
	if p != nil {
		applyPayout(b, p)
	}
	return t, true, nil
}

func applyPayout(b *bets.Bet, p *bets.Payout) {
	if p.ClearPayout {
		b.PayoutGross = nil
		b.PayoutCurrency = ""
		b.VoidFactor = nil
		b.DeadHeatFactor = nil
		return
	}
	if p.Gross != nil {
		v := *p.Gross
		b.PayoutGross = &v
	}
	if p.Currency != "" {
		b.PayoutCurrency = p.Currency
	}
	if p.VoidFactor != nil {
		v := *p.VoidFactor
		b.VoidFactor = &v
	}
	if p.DeadHeatFactor != nil {
		v := *p.DeadHeatFactor
		b.DeadHeatFactor = &v
	}
}

// transitionCount exposes the raw transition history for a bet so tests
// can assert append-only behaviour.
func (r *fakeRepo) transitionCount(betID string) int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.transitions[betID])
}

// fakeOutcomes is an in-memory OutcomeStateLookup. Per outcome we keep
// market+outcome activeness and the current odds.
type fakeOutcomes struct {
	mu   sync.Mutex
	data map[string]bets.OutcomeView
}

func newFakeOutcomes() *fakeOutcomes {
	return &fakeOutcomes{data: map[string]bets.OutcomeView{}}
}

func outcomeKey(matchID, marketID, outcomeID string) string {
	return matchID + "|" + marketID + "|" + outcomeID
}

func (f *fakeOutcomes) set(matchID, marketID, outcomeID string, v bets.OutcomeView) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.data[outcomeKey(matchID, marketID, outcomeID)] = v
}

func (f *fakeOutcomes) OutcomeState(_ context.Context, matchID, marketID, outcomeID string) (bets.OutcomeView, bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	v, ok := f.data[outcomeKey(matchID, marketID, outcomeID)]
	return v, ok, nil
}

// counterIDs is a deterministic IDGenerator for assertions.
type counterIDs struct{ n atomic.Int64 }

func (c *counterIDs) NextID() string {
	v := c.n.Add(1)
	return fmt.Sprintf("bet_%04d", v)
}

// captureLogger collects every Manager callback.
type captureLogger struct {
	mu       sync.Mutex
	placed   []placeRecord
	applied  []applyRecord
	skipped  []skipRecord
	validate []bets.ValidateResponse
}

type placeRecord struct {
	BetID   string
	Deduped bool
	State   bets.State
}

type applyRecord struct {
	BetID string
	Event bets.EventKind
	From  bets.State
	To    bets.State
}

type skipRecord struct {
	BetID  string
	Event  bets.EventKind
	From   bets.State
	Reason string
}

func (l *captureLogger) Validated(_ bets.ValidateRequest, resp bets.ValidateResponse) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.validate = append(l.validate, resp)
}

func (l *captureLogger) Placed(b *bets.Bet, deduped bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.placed = append(l.placed, placeRecord{BetID: b.ID, Deduped: deduped, State: b.State})
}

func (l *captureLogger) Applied(betID string, ev bets.EventKind, from, to bets.State) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.applied = append(l.applied, applyRecord{BetID: betID, Event: ev, From: from, To: to})
}

func (l *captureLogger) Skipped(betID string, ev bets.EventKind, from bets.State, reason string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.skipped = append(l.skipped, skipRecord{BetID: betID, Event: ev, From: from, Reason: reason})
}
