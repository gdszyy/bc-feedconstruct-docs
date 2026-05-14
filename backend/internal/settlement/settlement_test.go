package settlement_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/gdszyy/bc-feedconstruct-docs/backend/internal/feed"
	"github.com/gdszyy/bc-feedconstruct-docs/backend/internal/settlement"
)

// envWith decodes the JSON body into a feed.Envelope, preserving the raw
// payload so the handler can re-parse it.
func envWith(payload string) feed.Envelope {
	e, err := feed.DecodeEnvelope([]byte(payload))
	if err != nil {
		e = feed.Envelope{Payload: []byte(payload)}
	}
	e.Payload = []byte(payload)
	return e
}

// 验收 7 — 结算
//
// Given a bet_settlement for outcome (42, 1, "", 1) with result=win, certainty=1
// When SettlementHandler processes it
// Then a settlements row is inserted with result=win, certainty=1
//      AND markets row (42,1,"") transitions to status=settled
func TestGiven_BetSettlementWin_When_Handled_Then_SettlementRowAndMarketSettled(t *testing.T) {
	repo := newFakeRepo()
	repo.seedMatch(42)
	repo.seedMarket(42, 1, "", settlement.StatusActive, settlement.StatusUnknown)
	h := settlement.New(repo)

	body := `{"matchId":42,"marketTypeId":1,"specifier":"","certainty":1,
		"outcomes":[{"id":1,"result":"win"}]}`
	require.NoError(t, h.HandleBetSettlement(context.Background(), feed.MsgBetSettlement,
		envWith(body), [16]byte{0x07}))

	settlements, _, _, markets := repo.snapshot()
	require.Len(t, settlements, 1)
	s := settlements[0]
	require.EqualValues(t, 42, s.MatchID)
	require.EqualValues(t, 1, s.MarketTypeID)
	require.Equal(t, "", s.Specifier)
	require.EqualValues(t, 1, s.OutcomeID)
	require.Equal(t, settlement.ResultWin, s.Result)
	require.EqualValues(t, settlement.CertaintyConfirmed, s.Certainty)
	require.Equal(t, [16]byte{0x07}, s.RawMessageID)

	require.Equal(t, settlement.StatusSettled, markets[marketKey{42, 1, ""}].current,
		"acceptance 7: market must transition to settled when bet_settlement applied")
	require.EqualValues(t, 1, h.SettlementCount())
}

// Given a bet_settlement carrying void_factor=0.5 and dead_heat_factor=0.25
// When SettlementHandler processes it
// Then settlements row stores both factors verbatim
func TestGiven_VoidAndDeadHeatFactors_When_Handled_Then_FactorsPersistedExactly(t *testing.T) {
	repo := newFakeRepo()
	repo.seedMatch(42)
	repo.seedMarket(42, 1, "", settlement.StatusActive, settlement.StatusUnknown)
	h := settlement.New(repo)

	body := `{"matchId":42,"marketTypeId":1,"certainty":1,
		"outcomes":[{"id":1,"result":"half_win","voidFactor":0.5,"deadHeatFactor":0.25}]}`
	require.NoError(t, h.HandleBetSettlement(context.Background(), feed.MsgBetSettlement,
		envWith(body), [16]byte{}))

	settlements, _, _, _ := repo.snapshot()
	require.Len(t, settlements, 1)
	s := settlements[0]
	require.Equal(t, settlement.ResultHalfWin, s.Result)
	require.NotNil(t, s.VoidFactor)
	require.NotNil(t, s.DeadHeatFactor)
	require.EqualValues(t, 0.5, *s.VoidFactor)
	require.EqualValues(t, 0.25, *s.DeadHeatFactor)
}

// Given a bet_settlement with certainty=0 (uncertain) followed by a later
//       bet_settlement with certainty=1 for the same outcome
// When both are processed in order
// Then the later certain settlement supersedes the uncertain one
//      (history preserved; latest row reflects certainty=1)
func TestGiven_UncertainThenCertain_When_Handled_Then_CertainSupersedes(t *testing.T) {
	repo := newFakeRepo()
	repo.seedMatch(42)
	repo.seedMarket(42, 1, "", settlement.StatusActive, settlement.StatusUnknown)
	h := settlement.New(repo)
	// Strictly increasing time so the two settlements_uniq columns differ.
	clock := time.Date(2026, 5, 14, 12, 0, 0, 0, time.UTC)
	h.Now = func() time.Time {
		clock = clock.Add(time.Second)
		return clock
	}

	uncertain := `{"matchId":42,"marketTypeId":1,"certainty":0,
		"outcomes":[{"id":1,"result":"win","certainty":0}]}`
	require.NoError(t, h.HandleBetSettlement(context.Background(), feed.MsgBetSettlement,
		envWith(uncertain), [16]byte{0x01}))

	certain := `{"matchId":42,"marketTypeId":1,"certainty":1,
		"outcomes":[{"id":1,"result":"win","certainty":1}]}`
	require.NoError(t, h.HandleBetSettlement(context.Background(), feed.MsgBetSettlement,
		envWith(certain), [16]byte{0x02}))

	settlements, _, _, _ := repo.snapshot()
	require.Len(t, settlements, 2, "history must be preserved")
	require.EqualValues(t, settlement.CertaintyUncertain, settlements[0].Certainty)
	require.EqualValues(t, settlement.CertaintyConfirmed, settlements[1].Certainty)
	require.True(t, settlements[1].SettledAt.After(settlements[0].SettledAt),
		"certain settlement must be timestamped after uncertain")

	latest, ok, err := repo.LatestSettlementForOutcome(context.Background(), 42, 1, "", 1)
	require.NoError(t, err)
	require.True(t, ok)
	require.EqualValues(t, settlement.CertaintyConfirmed, latest.Certainty,
		"latest non-rolled settlement must be the certain one")
}
