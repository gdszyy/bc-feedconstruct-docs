package catalog_test

import (
	"context"
	"errors"
	"sync"

	"github.com/gdszyy/bc-feedconstruct-docs/backend/internal/catalog"
)

// fakeRepo is the in-memory Repo implementation shared by all
// catalog_test files. It mirrors the catalog tables 1:1 and protects
// every map with a single mutex so concurrent scenarios stay race-clean.
type fakeRepo struct {
	mu              sync.Mutex
	sports          map[int32]catalog.Sport
	regions         map[int32]catalog.Region
	competitions    map[int32]catalog.Competition
	matches         map[int64]catalog.Match
	fixtureChanges  []catalog.FixtureChangeRow
	sportUpsertCnt  int
	regionUpsertCnt int
	compUpsertCnt   int
	matchUpsertCnt  int
	failGetSport    bool
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{
		sports:       map[int32]catalog.Sport{},
		regions:      map[int32]catalog.Region{},
		competitions: map[int32]catalog.Competition{},
		matches:      map[int64]catalog.Match{},
	}
}

func (r *fakeRepo) UpsertSport(_ context.Context, s catalog.Sport) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sportUpsertCnt++
	if cur, ok := r.sports[s.ID]; ok && s.Name == "" {
		// Stub upsert (auto-created from region/match) must not blank
		// an existing name.
		s.Name = cur.Name
	}
	r.sports[s.ID] = s
	return nil
}

func (r *fakeRepo) SoftDeleteSport(_ context.Context, id int32) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	cur, ok := r.sports[id]
	if !ok {
		r.sports[id] = catalog.Sport{ID: id, IsActive: false}
		return nil
	}
	cur.IsActive = false
	r.sports[id] = cur
	return nil
}

func (r *fakeRepo) GetSport(_ context.Context, id int32) (catalog.Sport, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.failGetSport {
		return catalog.Sport{}, false, errors.New("simulated GetSport failure")
	}
	s, ok := r.sports[id]
	return s, ok, nil
}

func (r *fakeRepo) UpsertRegion(_ context.Context, rg catalog.Region) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.regionUpsertCnt++
	if cur, ok := r.regions[rg.ID]; ok && rg.Name == "" {
		rg.Name = cur.Name
	}
	r.regions[rg.ID] = rg
	return nil
}

func (r *fakeRepo) GetRegion(_ context.Context, id int32) (catalog.Region, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	rg, ok := r.regions[id]
	return rg, ok, nil
}

func (r *fakeRepo) UpsertCompetition(_ context.Context, c catalog.Competition) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.compUpsertCnt++
	if cur, ok := r.competitions[c.ID]; ok && c.Name == "" {
		c.Name = cur.Name
	}
	r.competitions[c.ID] = c
	return nil
}

func (r *fakeRepo) UpsertMatch(_ context.Context, m catalog.Match) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.matchUpsertCnt++
	r.matches[m.ID] = m
	return nil
}

func (r *fakeRepo) GetMatch(_ context.Context, id int64) (catalog.Match, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	m, ok := r.matches[id]
	return m, ok, nil
}

func (r *fakeRepo) InsertFixtureChange(_ context.Context, row catalog.FixtureChangeRow) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.fixtureChanges = append(r.fixtureChanges, row)
	return nil
}

func (r *fakeRepo) snapshot() (sports map[int32]catalog.Sport, regions map[int32]catalog.Region, comps map[int32]catalog.Competition, matches map[int64]catalog.Match, fcs []catalog.FixtureChangeRow) {
	r.mu.Lock()
	defer r.mu.Unlock()
	sports = make(map[int32]catalog.Sport, len(r.sports))
	for k, v := range r.sports {
		sports[k] = v
	}
	regions = make(map[int32]catalog.Region, len(r.regions))
	for k, v := range r.regions {
		regions[k] = v
	}
	comps = make(map[int32]catalog.Competition, len(r.competitions))
	for k, v := range r.competitions {
		comps[k] = v
	}
	matches = make(map[int64]catalog.Match, len(r.matches))
	for k, v := range r.matches {
		matches[k] = v
	}
	fcs = append(fcs, r.fixtureChanges...)
	return
}

type captureLogger struct {
	mu     sync.Mutex
	events []catalog.AntiRegressionEvent
}

func (c *captureLogger) AntiRegressionBlocked(ev catalog.AntiRegressionEvent) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.events = append(c.events, ev)
}

func (c *captureLogger) snapshot() []catalog.AntiRegressionEvent {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]catalog.AntiRegressionEvent, len(c.events))
	copy(out, c.events)
	return out
}
