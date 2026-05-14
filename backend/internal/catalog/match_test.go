package catalog_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/gdszyy/bc-feedconstruct-docs/backend/internal/catalog"
)

// 验收 4 — 主数据（M03/M04）
//
// Given a Match object delivered with sport / region / competition references
// When the catalog handler processes it
// Then sports, regions, competitions and matches rows are upserted with
//      name / start_at / home / away / is_live populated
func TestGiven_MatchObject_When_Handled_Then_HierarchyUpserted(t *testing.T) {
	repo := newFakeRepo()
	log := &captureLog{}
	h := catalog.NewHandler(repo, catalog.Options{
		Logger: log,
		Now:    func() time.Time { return mustParse(t, "2026-05-14T12:00:00Z") },
	})

	payload := []byte(`{
		"Id": 42,
		"SportId": 1,
		"RegionId": 7,
		"CompetitionId": 100,
		"Date": "2026-05-14T20:00:00Z",
		"IsLive": true,
		"MatchStatus": 1,
		"MatchMembers": [
			{"Type": 1, "Name": "Home FC"},
			{"Type": 2, "Name": "Away United"}
		]
	}`)

	require.NoError(t, h.HandleMatch(context.Background(), payload, rawID(1)))

	require.Equal(t, []int32{1}, repo.sportUpserts, "sport row must be upserted")
	require.Equal(t, []catalog.RegionRef{{ID: 7, SportID: 1}}, repo.regionUpserts)
	require.Equal(t, []catalog.CompetitionRef{{ID: 100, RegionID: 7, SportID: 1}}, repo.competitionUpserts)

	require.Len(t, repo.matchUpserts, 1)
	m := repo.matchUpserts[0]
	require.EqualValues(t, 42, m.ID)
	require.EqualValues(t, 1, m.SportID)
	require.NotNil(t, m.CompetitionID)
	require.EqualValues(t, 100, *m.CompetitionID)
	require.Equal(t, "Home FC", m.Home)
	require.Equal(t, "Away United", m.Away)
	require.True(t, m.IsLive)
	require.Equal(t, "live", m.Status)
	require.Equal(t, mustParse(t, "2026-05-14T20:00:00Z"), m.StartAt.UTC())
}

// Given a fixture_change altering start_at and status
// When the catalog handler processes it
// Then matches.start_at / matches.status are updated AND a fixture_changes
//      history row is inserted with old/new diff and raw_message_id link
func TestGiven_FixtureChange_When_Handled_Then_MatchUpdatedAndHistoryRecorded(t *testing.T) {
	repo := newFakeRepo()
	repo.existing[42] = catalog.MatchRecord{
		ID:      42,
		SportID: 1,
		Status:  "not_started",
		IsLive:  false,
		StartAt: mustParse(t, "2026-05-14T19:00:00Z"),
	}
	log := &captureLog{}
	h := catalog.NewHandler(repo, catalog.Options{Logger: log})

	payload := []byte(`{
		"Id": 42,
		"SportId": 1,
		"RegionId": 7,
		"CompetitionId": 100,
		"Date": "2026-05-14T20:30:00Z",
		"IsLive": true,
		"MatchStatus": 1,
		"MatchMembers": [
			{"Type": 1, "Name": "Home"},
			{"Type": 2, "Name": "Away"}
		]
	}`)

	require.NoError(t, h.HandleFixtureChange(context.Background(), payload, rawID(2)))

	require.Len(t, repo.matchUpserts, 1)
	updated := repo.matchUpserts[0]
	require.Equal(t, "live", updated.Status)
	require.Equal(t, mustParse(t, "2026-05-14T20:30:00Z"), updated.StartAt.UTC())

	require.Len(t, repo.fixtureChanges, 1)
	fc := repo.fixtureChanges[0]
	require.EqualValues(t, 42, fc.MatchID)
	require.Equal(t, "fixture_change", fc.ChangeType)
	require.Equal(t, rawID(2), fc.RawMessageID)
	require.Equal(t, "not_started", fc.Old["status"])
	require.Equal(t, "live", fc.New["status"])
	require.Equal(t, "2026-05-14T19:00:00Z", fc.Old["start_at"])
	require.Equal(t, "2026-05-14T20:30:00Z", fc.New["start_at"])
}

// 验收 12 — 防回退（赛事级）
//
// Given matches.status currently = ended
// When a delivery arrives with status = live for the same match
// Then matches.status is NOT regressed and a "status.regress.blocked" log is emitted
func TestGiven_EndedMatch_When_LiveStatusArrives_Then_NoRegressionLogged(t *testing.T) {
	repo := newFakeRepo()
	repo.existing[42] = catalog.MatchRecord{
		ID:      42,
		SportID: 1,
		Status:  "ended",
		IsLive:  false,
	}
	log := &captureLog{}
	h := catalog.NewHandler(repo, catalog.Options{Logger: log})

	payload := []byte(`{
		"Id": 42,
		"SportId": 1,
		"RegionId": 7,
		"CompetitionId": 100,
		"Date": "2026-05-14T20:30:00Z",
		"IsLive": true,
		"MatchStatus": 1
	}`)

	require.NoError(t, h.HandleMatch(context.Background(), payload, rawID(3)))

	require.Len(t, repo.matchUpserts, 1)
	persisted := repo.matchUpserts[0]
	require.Equal(t, "ended", persisted.Status, "terminal status must not regress")
	require.False(t, persisted.IsLive, "is_live must not flip back to true on a terminal match")

	require.True(t, log.has("status.regress.blocked"),
		"a status.regress.blocked log must be emitted (events: %+v)", log.events)
}

// Given a Match payload describing a cancellation (MatchStatus=3)
// When the handler processes it
// Then matches.status becomes cancelled even if the previous state was live
//      (cancellation is a recognised terminal transition, not a regression)
func TestGiven_LiveMatch_When_CancelStatusArrives_Then_StatusBecomesCancelled(t *testing.T) {
	repo := newFakeRepo()
	repo.existing[42] = catalog.MatchRecord{ID: 42, SportID: 1, Status: "live", IsLive: true}
	log := &captureLog{}
	h := catalog.NewHandler(repo, catalog.Options{Logger: log})

	payload := []byte(`{
		"Id": 42, "SportId": 1, "CompetitionId": 100, "RegionId": 7,
		"Date": "2026-05-14T20:00:00Z",
		"IsLive": false,
		"MatchStatus": 3
	}`)
	require.NoError(t, h.HandleMatch(context.Background(), payload, rawID(4)))

	require.Len(t, repo.matchUpserts, 1)
	require.Equal(t, "cancelled", repo.matchUpserts[0].Status)
	require.False(t, log.has("status.regress.blocked"))
}

// ---------------------------------------------------------------------------
// Test doubles
// ---------------------------------------------------------------------------

type fakeRepo struct {
	mu                 sync.Mutex
	sportUpserts       []int32
	regionUpserts      []catalog.RegionRef
	competitionUpserts []catalog.CompetitionRef
	matchUpserts       []catalog.MatchRecord
	fixtureChanges     []catalog.FixtureChange
	existing           map[int64]catalog.MatchRecord
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{existing: map[int64]catalog.MatchRecord{}}
}

func (f *fakeRepo) UpsertSport(_ context.Context, id int32, _ string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.sportUpserts = append(f.sportUpserts, id)
	return nil
}

func (f *fakeRepo) UpsertRegion(_ context.Context, r catalog.RegionRef) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.regionUpserts = append(f.regionUpserts, r)
	return nil
}

func (f *fakeRepo) UpsertCompetition(_ context.Context, c catalog.CompetitionRef) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.competitionUpserts = append(f.competitionUpserts, c)
	return nil
}

func (f *fakeRepo) LoadMatch(_ context.Context, id int64) (catalog.MatchRecord, bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	m, ok := f.existing[id]
	return m, ok, nil
}

func (f *fakeRepo) UpsertMatch(_ context.Context, m catalog.MatchRecord) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.matchUpserts = append(f.matchUpserts, m)
	f.existing[m.ID] = m
	return nil
}

func (f *fakeRepo) AppendFixtureChange(_ context.Context, fc catalog.FixtureChange) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.fixtureChanges = append(f.fixtureChanges, fc)
	return nil
}

type captureLog struct {
	mu     sync.Mutex
	events []map[string]any
}

func (l *captureLog) Warn(event string, fields map[string]any) {
	l.mu.Lock()
	defer l.mu.Unlock()
	entry := map[string]any{"event": event}
	for k, v := range fields {
		entry[k] = v
	}
	l.events = append(l.events, entry)
}

func (l *captureLog) has(event string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	for _, e := range l.events {
		if e["event"] == event {
			return true
		}
	}
	return false
}

func mustParse(t *testing.T, s string) time.Time {
	t.Helper()
	ts, err := time.Parse(time.RFC3339, s)
	require.NoError(t, err)
	return ts.UTC()
}

func rawID(b byte) [16]byte {
	var id [16]byte
	id[15] = b
	return id
}
