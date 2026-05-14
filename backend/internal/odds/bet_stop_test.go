package odds_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gdszyy/bc-feedconstruct-docs/backend/internal/feed"
	"github.com/gdszyy/bc-feedconstruct-docs/backend/internal/odds"
)

// 验收 6 — 停投
//
// Given a bet_stop for match=42 with marketStatus=suspended targeting all markets
// When BetStopHandler processes it
// Then every markets row of match 42 transitions to status=suspended
//      AND a market_status_history row is appended for each transition
//      with raw_message_id linked when non-zero
func TestGiven_BetStopAllMarkets_When_Handled_Then_MarketsSuspendedAndHistoryAppended(t *testing.T) {
	repo := newFakeRepo()
	repo.seedMatch(42)
	repo.seedMarket(odds.Market{MatchID: 42, MarketTypeID: 1, Specifier: "", Status: odds.StatusActive})
	repo.seedMarket(odds.Market{MatchID: 42, MarketTypeID: 2, Specifier: "hcap=-1", Status: odds.StatusActive})
	h := odds.New(repo)

	rawID := [16]byte{0xAB}
	body := `{"matchId":42,"marketStatus":"suspended"}`
	require.NoError(t, h.HandleBetStop(context.Background(), feed.MsgBetStop,
		envWith(body), rawID))

	markets, _, hist := repo.snapshot()
	require.Equal(t, odds.StatusSuspended, markets[marketKey{42, 1, ""}].Status)
	require.Equal(t, odds.StatusSuspended, markets[marketKey{42, 2, "hcap=-1"}].Status)
	require.Len(t, hist, 2, "one history row per transitioned market")
	for _, row := range hist {
		require.Equal(t, odds.StatusActive, row.From)
		require.Equal(t, odds.StatusSuspended, row.To)
		require.Equal(t, rawID, row.RawMessageID)
	}
}

// Given a bet_stop targeting a specific market_group
// When BetStopHandler processes it
// Then only markets with that group_id transition; others remain active
func TestGiven_BetStopByGroup_When_Handled_Then_OnlyTargetedMarketsTransition(t *testing.T) {
	repo := newFakeRepo()
	repo.seedMatch(42)
	g7 := int32(7)
	g8 := int32(8)
	repo.seedMarket(odds.Market{MatchID: 42, MarketTypeID: 1, Specifier: "", Status: odds.StatusActive, GroupID: &g7})
	repo.seedMarket(odds.Market{MatchID: 42, MarketTypeID: 2, Specifier: "", Status: odds.StatusActive, GroupID: &g8})
	h := odds.New(repo)

	body := `{"matchId":42,"groupId":7,"marketStatus":"suspended"}`
	require.NoError(t, h.HandleBetStop(context.Background(), feed.MsgBetStop,
		envWith(body), [16]byte{}))

	markets, _, _ := repo.snapshot()
	require.Equal(t, odds.StatusSuspended, markets[marketKey{42, 1, ""}].Status, "group 7 must be suspended")
	require.Equal(t, odds.StatusActive, markets[marketKey{42, 2, ""}].Status, "group 8 must stay active")
}

// Given a bet_stop targeting a specific marketTypeId + specifier
// When BetStopHandler processes it
// Then only that exact market transitions
func TestGiven_BetStopByMarketTypeID_When_Handled_Then_OnlyExactMarketTransitions(t *testing.T) {
	repo := newFakeRepo()
	repo.seedMatch(42)
	repo.seedMarket(odds.Market{MatchID: 42, MarketTypeID: 1, Specifier: "", Status: odds.StatusActive})
	repo.seedMarket(odds.Market{MatchID: 42, MarketTypeID: 2, Specifier: "", Status: odds.StatusActive})
	h := odds.New(repo)

	body := `{"matchId":42,"marketTypeId":1,"marketStatus":"suspended"}`
	require.NoError(t, h.HandleBetStop(context.Background(), feed.MsgBetStop,
		envWith(body), [16]byte{}))

	markets, _, _ := repo.snapshot()
	require.Equal(t, odds.StatusSuspended, markets[marketKey{42, 1, ""}].Status)
	require.Equal(t, odds.StatusActive, markets[marketKey{42, 2, ""}].Status)
}

// Given a bet_stop hitting a market currently in status=settled
// When BetStopHandler processes it
// Then the suspend is rejected (anti-regression), status stays settled,
//      AntiRegressionEvent is emitted, and no history row is appended for
//      the rejected transition
func TestGiven_BetStopOnSettledMarket_When_Handled_Then_NoRegression(t *testing.T) {
	repo := newFakeRepo()
	repo.seedMatch(42)
	repo.seedMarket(odds.Market{MatchID: 42, MarketTypeID: 1, Specifier: "", Status: odds.StatusSettled})
	log := &captureLogger{}
	h := odds.New(repo)
	h.Logger = log

	body := `{"matchId":42,"marketStatus":"suspended"}`
	require.NoError(t, h.HandleBetStop(context.Background(), feed.MsgBetStop,
		envWith(body), [16]byte{}))

	markets, _, hist := repo.snapshot()
	require.Equal(t, odds.StatusSettled, markets[marketKey{42, 1, ""}].Status,
		"settled must not regress to suspended")
	require.Empty(t, hist, "rejected transitions must not write history")

	events := log.snapshot()
	require.Len(t, events, 1)
	require.Equal(t, odds.StatusSettled, events[0].From)
	require.Equal(t, odds.StatusSuspended, events[0].To)
}

// Given a bet_stop targeting a match the catalog handler has not seen
// When processed
// Then the handler short-circuits silently (no upserts, no history)
func TestGiven_BetStopForUnknownMatch_When_Handled_Then_SkipsSilently(t *testing.T) {
	repo := newFakeRepo()
	// no seedMatch
	h := odds.New(repo)
	body := `{"matchId":999,"marketStatus":"suspended"}`
	require.NoError(t, h.HandleBetStop(context.Background(), feed.MsgBetStop,
		envWith(body), [16]byte{}))

	markets, _, hist := repo.snapshot()
	require.Empty(t, markets)
	require.Empty(t, hist)
}

// Given a bet_stop without any explicit marketStatus
// When processed
// Then the default target is suspended (FC's most common bet_stop signal)
func TestGiven_BetStopWithoutStatus_When_Handled_Then_DefaultsToSuspended(t *testing.T) {
	repo := newFakeRepo()
	repo.seedMatch(42)
	repo.seedMarket(odds.Market{MatchID: 42, MarketTypeID: 1, Specifier: "", Status: odds.StatusActive})
	h := odds.New(repo)

	require.NoError(t, h.HandleBetStop(context.Background(), feed.MsgBetStop,
		envWith(`{"matchId":42}`), [16]byte{}))
	markets, _, _ := repo.snapshot()
	require.Equal(t, odds.StatusSuspended, markets[marketKey{42, 1, ""}].Status)
}
