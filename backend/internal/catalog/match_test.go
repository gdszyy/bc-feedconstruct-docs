package catalog_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/gdszyy/bc-feedconstruct-docs/backend/internal/catalog"
	"github.com/gdszyy/bc-feedconstruct-docs/backend/internal/feed"
)

func mustTime(t *testing.T, s string) time.Time {
	t.Helper()
	v, err := time.Parse(time.RFC3339, s)
	require.NoError(t, err)
	return v
}

func ptrInt32(v int32) *int32 { return &v }

// 验收 4 — 主数据（M03/M04）— Match 上行处理
//
// Given a Match (objectType=4) delivery carrying sport / region / competition
//       references and home / away / start_at fields
// When the catalog handler processes it
// Then sports, regions, competitions and matches rows are upserted with
//      name / start_at / home / away / is_live populated AND the upsert is
//      a single transaction so partial-write is impossible.
func TestGiven_MatchObject_When_Handled_Then_HierarchyUpserted(t *testing.T) {
	repo := newFakeRepo()
	h := catalog.New(repo)
	startAt := mustTime(t, "2026-05-14T18:00:00Z")

	payload := []byte(`{
        "id": 42, "sportId": 1, "regionId": 11, "competitionId": 111,
        "name": "Real Madrid vs Barcelona", "home": "Real Madrid",
        "away": "Barcelona", "startAt": "2026-05-14T18:00:00Z",
        "isLive": false, "status": "not_started",
        "sport": {"id":1,"name":"Football"},
        "region": {"id":11,"sportId":1,"name":"Spain"},
        "competition": {"id":111,"regionId":11,"sportId":1,"name":"La Liga"}
    }`)

	require.NoError(t, h.HandleMatch(context.Background(), feed.MsgFixture, feed.Envelope{
		ObjectType: 4, MatchID: ptrInt64(42), SportID: ptrInt32(1), Payload: payload,
	}, [16]byte{}))

	sports, regions, comps, matches, fcs := repo.snapshot()
	require.Contains(t, sports, int32(1))
	require.Contains(t, regions, int32(11))
	require.Equal(t, int32(1), regions[11].SportID)
	require.Contains(t, comps, int32(111))
	require.Equal(t, int32(11), comps[111].RegionID)
	require.Equal(t, int32(1), comps[111].SportID)

	require.Contains(t, matches, int64(42))
	m := matches[42]
	require.Equal(t, int32(1), m.SportID)
	require.NotNil(t, m.CompetitionID)
	require.Equal(t, int32(111), *m.CompetitionID)
	require.Equal(t, "Real Madrid", m.Home)
	require.Equal(t, "Barcelona", m.Away)
	require.NotNil(t, m.StartAt)
	require.True(t, m.StartAt.Equal(startAt))
	require.False(t, m.IsLive)
	require.Equal(t, catalog.StatusNotStarted, m.Status)

	require.Empty(t, fcs, "initial fixture upsert without StatusChange must not write fixture_changes")
}

// Given the same Match delivery replayed (identical eventId)
// When the catalog handler processes it twice
// Then the matches row is unchanged after the second call AND no duplicate
//      fixture_changes history row is written.
func TestGiven_SameMatchDeliveredTwice_When_Handled_Then_Idempotent(t *testing.T) {
	repo := newFakeRepo()
	h := catalog.New(repo)
	ctx := context.Background()

	payload := []byte(`{
        "id": 42, "sportId": 1, "competitionId": 111, "regionId": 11,
        "eventId": "evt-1",
        "home": "A", "away": "B", "status": "not_started",
        "region":{"id":11,"sportId":1,"name":"X"},
        "competition":{"id":111,"regionId":11,"sportId":1,"name":"Y"}
    }`)
	env := feed.Envelope{
		ObjectType: 4, MatchID: ptrInt64(42), SportID: ptrInt32(1),
		EventID: "evt-1", StatusChange: true, Payload: payload,
	}

	require.NoError(t, h.HandleMatch(ctx, feed.MsgFixtureChange, env, [16]byte{1}))
	firstSnap, _, _, firstMatches, firstFCs := repo.snapshot()

	require.NoError(t, h.HandleMatch(ctx, feed.MsgFixtureChange, env, [16]byte{2}))
	secondSnap, _, _, secondMatches, secondFCs := repo.snapshot()

	require.Equal(t, firstSnap, secondSnap)
	require.Equal(t, firstMatches, secondMatches)
	require.Equal(t, firstFCs, secondFCs, "duplicate event must not append history rows")
	require.Equal(t, 1, repo.matchUpsertCnt, "duplicate event must short-circuit before upserting match")
}

// Given a fixture_change (statusChange=true) altering start_at AND status
// When the catalog handler processes it
// Then matches.start_at / matches.status are updated AND a fixture_changes
//      history row is inserted with change_type / old / new JSON diff and
//      raw_message_id linked to the source raw_messages row.
func TestGiven_FixtureChange_When_Handled_Then_MatchUpdatedAndHistoryRecorded(t *testing.T) {
	repo := newFakeRepo()
	h := catalog.New(repo)
	ctx := context.Background()

	seed := []byte(`{
        "id":42,"sportId":1,"regionId":11,"competitionId":111,
        "home":"A","away":"B","startAt":"2026-05-14T18:00:00Z",
        "status":"not_started","eventId":"evt-seed",
        "region":{"id":11,"sportId":1,"name":"X"},
        "competition":{"id":111,"regionId":11,"sportId":1,"name":"Y"}
    }`)
	require.NoError(t, h.HandleMatch(ctx, feed.MsgFixture, feed.Envelope{
		ObjectType: 4, MatchID: ptrInt64(42), SportID: ptrInt32(1),
		EventID: "evt-seed", Payload: seed,
	}, [16]byte{0xAA}))

	change := []byte(`{
        "id":42,"sportId":1,"eventId":"evt-2",
        "startAt":"2026-05-14T19:30:00Z","status":"live"
    }`)
	rawID := [16]byte{0xBB, 0xCC}
	require.NoError(t, h.HandleMatch(ctx, feed.MsgFixtureChange, feed.Envelope{
		ObjectType: 4, MatchID: ptrInt64(42), SportID: ptrInt32(1),
		EventID: "evt-2", StatusChange: true, Payload: change,
	}, rawID))

	_, _, _, matches, fcs := repo.snapshot()
	m := matches[42]
	require.Equal(t, catalog.StatusLive, m.Status)
	require.True(t, m.IsLive, "isLive must follow status=live when isLive is absent")
	require.NotNil(t, m.StartAt)
	require.True(t, m.StartAt.Equal(mustTime(t, "2026-05-14T19:30:00Z")))

	require.Len(t, fcs, 1)
	require.Equal(t, int64(42), fcs[0].MatchID)
	require.Equal(t, rawID, fcs[0].RawMessageID)
	require.Contains(t, fcs[0].ChangeType, "start_at")
	require.Contains(t, fcs[0].ChangeType, "status")
	require.Equal(t, "2026-05-14T18:00:00Z", fcs[0].Old["start_at"])
	require.Equal(t, "2026-05-14T19:30:00Z", fcs[0].New["start_at"])
	require.Equal(t, "not_started", fcs[0].Old["status"])
	require.Equal(t, "live", fcs[0].New["status"])
}

// Given a fixture_change that touches ONLY start_at (status unchanged)
// When the catalog handler processes it
// Then only the start_at diff is recorded in fixture_changes.new / .old
//      AND matches.status is left untouched.
func TestGiven_FixtureChangeStartAtOnly_When_Handled_Then_DiffIsMinimal(t *testing.T) {
	repo := newFakeRepo()
	h := catalog.New(repo)
	ctx := context.Background()

	require.NoError(t, h.HandleMatch(ctx, feed.MsgFixture, feed.Envelope{
		ObjectType: 4, MatchID: ptrInt64(42), SportID: ptrInt32(1),
		EventID: "evt-seed",
		Payload: []byte(`{"id":42,"sportId":1,"home":"A","away":"B",
            "startAt":"2026-05-14T18:00:00Z","status":"not_started",
            "eventId":"evt-seed"}`),
	}, [16]byte{}))

	require.NoError(t, h.HandleMatch(ctx, feed.MsgFixtureChange, feed.Envelope{
		ObjectType: 4, MatchID: ptrInt64(42), SportID: ptrInt32(1),
		EventID: "evt-shift", StatusChange: true,
		Payload: []byte(`{"id":42,"sportId":1,"eventId":"evt-shift",
            "startAt":"2026-05-14T20:00:00Z"}`),
	}, [16]byte{})) // no status field

	_, _, _, matches, fcs := repo.snapshot()
	require.Equal(t, catalog.StatusNotStarted, matches[42].Status, "status must be left untouched")
	require.Len(t, fcs, 1)
	require.Equal(t, "start_at", fcs[0].ChangeType)
	_, hasStatus := fcs[0].New["status"]
	require.False(t, hasStatus, "status must not appear in the diff when it didn't change")
}

// 验收 12 — 防回退（赛事级）
//
// Given matches.status currently = ended
// When a delivery arrives with status = live for the same match
// Then matches.status is NOT regressed AND a "status.regress.blocked"
//      structured log/event is emitted with match_id / from / to fields.
func TestGiven_EndedMatch_When_LiveStatusArrives_Then_NoRegressionLogged(t *testing.T) {
	repo := newFakeRepo()
	logger := &captureLogger{}
	h := catalog.New(repo)
	h.Logger = logger
	ctx := context.Background()

	require.NoError(t, h.HandleMatch(ctx, feed.MsgFixture, feed.Envelope{
		ObjectType: 4, MatchID: ptrInt64(42), SportID: ptrInt32(1),
		Payload: []byte(`{"id":42,"sportId":1,"status":"ended","eventId":"evt-a"}`),
	}, [16]byte{}))

	require.NoError(t, h.HandleMatch(ctx, feed.MsgFixtureChange, feed.Envelope{
		ObjectType: 4, MatchID: ptrInt64(42), SportID: ptrInt32(1),
		StatusChange: true,
		Payload:      []byte(`{"id":42,"sportId":1,"status":"live","eventId":"evt-b"}`),
	}, [16]byte{}))

	_, _, _, matches, _ := repo.snapshot()
	require.Equal(t, catalog.StatusEnded, matches[42].Status, "status must not regress from ended → live")
	require.EqualValues(t, 1, h.RegressionCount())
	events := logger.snapshot()
	require.Len(t, events, 1)
	require.Equal(t, catalog.AntiRegressionEvent{MatchID: 42, From: catalog.StatusEnded, To: catalog.StatusLive}, events[0])
}

// Given matches.status currently = closed
// When a delivery arrives with status = ended (lower-ranked terminal)
// Then matches.status is NOT regressed AND the anti-regression counter
//      increments by 1.
func TestGiven_ClosedMatch_When_EndedStatusArrives_Then_NoRegression(t *testing.T) {
	repo := newFakeRepo()
	h := catalog.New(repo)
	ctx := context.Background()

	require.NoError(t, h.HandleMatch(ctx, feed.MsgFixture, feed.Envelope{
		ObjectType: 4, MatchID: ptrInt64(42), SportID: ptrInt32(1),
		Payload: []byte(`{"id":42,"sportId":1,"status":"closed","eventId":"e1"}`),
	}, [16]byte{}))

	require.NoError(t, h.HandleMatch(ctx, feed.MsgFixtureChange, feed.Envelope{
		ObjectType: 4, MatchID: ptrInt64(42), SportID: ptrInt32(1),
		StatusChange: true,
		Payload:      []byte(`{"id":42,"sportId":1,"status":"ended","eventId":"e2"}`),
	}, [16]byte{}))

	_, _, _, matches, _ := repo.snapshot()
	require.Equal(t, catalog.StatusClosed, matches[42].Status)
	require.EqualValues(t, 1, h.RegressionCount())
}

func ptrInt64(v int64) *int64 { return &v }
