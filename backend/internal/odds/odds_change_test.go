package odds_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gdszyy/bc-feedconstruct-docs/backend/internal/feed"
	"github.com/gdszyy/bc-feedconstruct-docs/backend/internal/odds"
)

func envWith(payload string) feed.Envelope {
	e, err := feed.DecodeEnvelope([]byte(payload))
	if err != nil {
		e = feed.Envelope{Payload: []byte(payload)}
	}
	e.Payload = []byte(payload)
	return e
}

// 验收 5 — 赔率（M05）
//
// Given an odds_change for match=42 carrying market_type_id=1, specifier="",
//       outcomes [{id:1, odds:1.85, active:true}, {id:2, odds:2.10, active:true}]
// When the OddsHandler processes it
// Then markets row (42,1,"") is upserted with status=active
//      and outcomes rows are upserted with the exact odds and is_active flags
func TestGiven_OddsChange_When_Handled_Then_MarketAndOutcomesUpserted(t *testing.T) {
	repo := newFakeRepo()
	repo.seedMatch(42)
	h := odds.New(repo)

	body := `{"matchId":42,"marketTypeId":1,"specifier":"","outcomes":[
		{"id":1,"odds":1.85,"isActive":true},
		{"id":2,"odds":2.10,"isActive":true}
	]}`
	require.NoError(t, h.HandleOddsChange(context.Background(), feed.MsgOddsChange,
		envWith(body), [16]byte{}))

	markets, outcomes, hist := repo.snapshot()
	require.Len(t, markets, 1)
	m := markets[marketKey{42, 1, ""}]
	require.Equal(t, odds.StatusActive, m.Status)

	require.Len(t, outcomes, 2)
	require.EqualValues(t, 1.85, *outcomes[outcomeKey{42, 1, "", 1}].Odds)
	require.True(t, outcomes[outcomeKey{42, 1, "", 1}].IsActive)
	require.EqualValues(t, 2.10, *outcomes[outcomeKey{42, 1, "", 2}].Odds)

	// market_status_history records the initial NULL -> active.
	require.Len(t, hist, 1)
	require.Equal(t, odds.StatusUnknown, hist[0].From)
	require.Equal(t, odds.StatusActive, hist[0].To)
}

// Given an odds_change with the same payload received twice
// When both deliveries are processed
// Then the outcomes row is updated to the same values exactly once
//      (no second history row beyond the initial transition)
func TestGiven_DuplicateOddsChange_When_Handled_Then_NoDuplicateRow(t *testing.T) {
	repo := newFakeRepo()
	repo.seedMatch(42)
	h := odds.New(repo)
	ctx := context.Background()

	body := `{"matchId":42,"marketTypeId":1,"outcomes":[{"id":1,"odds":1.85}]}`
	require.NoError(t, h.HandleOddsChange(ctx, feed.MsgOddsChange, envWith(body), [16]byte{}))
	require.NoError(t, h.HandleOddsChange(ctx, feed.MsgOddsChange, envWith(body), [16]byte{}))

	_, outcomes, hist := repo.snapshot()
	require.Len(t, outcomes, 1, "duplicate odds_change must not produce a second outcome row")
	require.Len(t, hist, 1, "no status change -> no new history row")
}

// Given an odds_change targeting a market that catalog has not yet seen
// When the handler runs
// Then it short-circuits silently (MatchExists=false) and writes nothing
func TestGiven_OddsChangeForUnknownMatch_When_Handled_Then_SkipsSilently(t *testing.T) {
	repo := newFakeRepo()
	// note: NOT seeding match
	h := odds.New(repo)

	body := `{"matchId":999,"marketTypeId":1,"outcomes":[{"id":1,"odds":1.85}]}`
	require.NoError(t, h.HandleOddsChange(context.Background(), feed.MsgOddsChange,
		envWith(body), [16]byte{}))

	markets, outcomes, hist := repo.snapshot()
	require.Empty(t, markets)
	require.Empty(t, outcomes)
	require.Empty(t, hist)
}

// 验收 12 (market level) — anti-regression: settled market cannot regress
//
// Given a markets row currently in status=settled
// When an odds_change arrives that would set it to active
// Then the transition is rejected, status stays settled, AntiRegressionEvent
//      is emitted, regressionCount is incremented, and no history row is appended
func TestGiven_SettledMarket_When_OddsChangeWouldActivate_Then_NoRegression(t *testing.T) {
	repo := newFakeRepo()
	repo.seedMatch(42)
	repo.seedMarket(odds.Market{
		MatchID: 42, MarketTypeID: 1, Specifier: "",
		Status: odds.StatusSettled,
	})
	log := &captureLogger{}
	h := odds.New(repo)
	h.Logger = log

	body := `{"matchId":42,"marketTypeId":1,"specifier":"","status":"active",
		"outcomes":[{"id":1,"odds":1.85}]}`
	require.NoError(t, h.HandleOddsChange(context.Background(), feed.MsgOddsChange,
		envWith(body), [16]byte{}))

	markets, _, hist := repo.snapshot()
	require.Equal(t, odds.StatusSettled, markets[marketKey{42, 1, ""}].Status)
	require.Empty(t, hist, "rejected transition must NOT be logged in history")

	events := log.snapshot()
	require.Len(t, events, 1)
	require.Equal(t, odds.StatusSettled, events[0].From)
	require.Equal(t, odds.StatusActive, events[0].To)
	require.Equal(t, int64(1), h.RegressionCount())
}

// Given an odds_change with a group_id that the persisted row lacked
// When processed
// Then the group_id is recorded so subsequent group-scoped bet_stop can target it
func TestGiven_OddsChangeWithGroupID_When_Handled_Then_GroupRecorded(t *testing.T) {
	repo := newFakeRepo()
	repo.seedMatch(42)
	h := odds.New(repo)

	body := `{"matchId":42,"marketTypeId":1,"groupId":7,"outcomes":[{"id":1,"odds":1.85}]}`
	require.NoError(t, h.HandleOddsChange(context.Background(), feed.MsgOddsChange,
		envWith(body), [16]byte{}))

	m, ok, _ := repo.GetMarket(context.Background(), 42, 1, "")
	require.True(t, ok)
	require.NotNil(t, m.GroupID)
	require.EqualValues(t, 7, *m.GroupID)
}

// Given an odds_change with no marketTypeId nor markets[]
// When processed
// Then the handler is a no-op (matchExists may be irrelevant here)
func TestGiven_OddsChangeWithoutMarketTypeID_When_Handled_Then_NoOp(t *testing.T) {
	repo := newFakeRepo()
	repo.seedMatch(42)
	h := odds.New(repo)

	body := `{"matchId":42,"outcomes":[{"id":1,"odds":1.85}]}`
	require.NoError(t, h.HandleOddsChange(context.Background(), feed.MsgOddsChange,
		envWith(body), [16]byte{}))

	markets, outcomes, _ := repo.snapshot()
	require.Empty(t, markets)
	require.Empty(t, outcomes)
}

// Given a flat odds_change without matchId at the payload root but with
// envelope.MatchID populated
// When the handler runs
// Then it falls back to envelope.MatchID
func TestGiven_OddsChangeMatchIDFromEnvelope_When_Handled_Then_FallsBack(t *testing.T) {
	repo := newFakeRepo()
	repo.seedMatch(42)
	h := odds.New(repo)

	body := `{"marketTypeId":1,"outcomes":[{"id":1,"odds":1.85}]}`
	env := envWith(body)
	env.MatchID = ptrInt64(42)

	require.NoError(t, h.HandleOddsChange(context.Background(), feed.MsgOddsChange,
		env, [16]byte{}))

	_, outcomes, _ := repo.snapshot()
	require.Len(t, outcomes, 1)
}
